package server

import (
	"context"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"bytes"
	"github.com/xinhaolin/htmlToWord"
)

//go:embed web/*
var assets embed.FS

type Server struct {
	slots           chan struct{}
	pdfSlot         chan struct{}
	soffice         string
	pdftoppm        string
	logger          *log.Logger
	converter       *htmlword.Converter
	remoteConverter *htmlword.Converter
	runtimeError    error
}

func New(logger *log.Logger) *Server {
	p := os.Getenv("HTMLWORD_SOFFICE")
	if p == "" {
		p, _ = exec.LookPath("soffice")
	}
	poppler := os.Getenv("HTMLWORD_PDFTOPPM")
	if poppler == "" {
		poppler, _ = exec.LookPath("pdftoppm")
	}
	converter, runtimeError := htmlword.New(htmlword.Config{BrowserPath: os.Getenv("HTMLWORD_CHROME")})
	remoteConverter, _ := htmlword.New(htmlword.Config{BrowserPath: os.Getenv("HTMLWORD_CHROME"), Resolver: htmlword.HTTPResolver{}})
	return &Server{slots: make(chan struct{}, 2), pdfSlot: make(chan struct{}, 1), soffice: p, pdftoppm: poppler, logger: logger, converter: converter, remoteConverter: remoteConverter, runtimeError: runtimeError}
}
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	sub, _ := fs.Sub(assets, "web")
	mux.Handle("GET /", http.FileServer(http.FS(sub)))
	// Local, fictional test artifacts are kept beside the executable so reports
	// and document downloads do not have to be compiled into the server binary.
	batchDir := os.Getenv("HTMLWORD_BATCH_DIR")
	if batchDir == "" {
		batchDir = "batch"
	}
	if stat, err := os.Stat(batchDir); err == nil && stat.IsDir() {
		mux.Handle("GET /batch/", http.StripPrefix("/batch/", http.FileServer(http.Dir(batchDir))))
	}
	mux.HandleFunc("GET /api/health", s.health)
	mux.HandleFunc("POST /api/convert", s.convert)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		if r.Method == "POST" {
			if origin := r.Header.Get("Origin"); origin != "" {
				u, e := url.Parse(origin)
				if e != nil || u.Host != r.Host {
					respond(w, http.StatusForbidden, map[string]string{"error": "请求来源不匹配"})
					return
				}
			}
		}
		defer func() {
			if err := recover(); err != nil {
				s.logger.Printf("handler panic: %v", err)
				http.Error(w, "转换失败", 500)
			}
		}()
		mux.ServeHTTP(w, r)
	})
}
func respond(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	respond(w, 200, map[string]any{"ok": s.runtimeError == nil, "engine": "Go SDK native OOXML", "chrome": s.converter != nil, "pdf_preview": s.soffice != "", "max_concurrency": cap(s.slots)})
}
func (s *Server) convert(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 12<<20)
	defer r.Body.Close()
	req := legacyRequest{Options: legacyOptions{PageSize: "A4", MarginMM: 20, Font: "Arial", FontSize: 11, Media: "print"}}
	decoder := json.NewDecoder(r.Body)
	if e := decoder.Decode(&req); e != nil {
		respond(w, 400, map[string]string{"error": "请求格式无效或超过 12 MiB：" + e.Error()})
		return
	}
	var extra any
	if e := decoder.Decode(&extra); e != io.EOF {
		respond(w, 400, map[string]string{"error": "请求只能包含一个 JSON 对象"})
		return
	}
	if e := req.Options.sdkOptions().Validate(); e != nil {
		respond(w, 400, map[string]string{"error": e.Error()})
		return
	}

	if strings.TrimSpace(req.HTML) == "" || len(req.HTML)+len(req.CSS) > 8<<20 {
		respond(w, 400, map[string]string{"error": "请输入不超过 8 MiB 的 HTML/CSS"})
		return
	}
	select {
	case s.slots <- struct{}{}:
		defer func() { <-s.slots }()
	default:
		respond(w, 429, map[string]string{"error": "转换任务已满，请稍后重试"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 70*time.Second)
	defer cancel()
	converter := s.converter
	if req.Options.AllowRemote {
		converter = s.remoteConverter
	}
	if converter == nil {
		respond(w, 503, map[string]string{"error": "Chromium runtime unavailable"})
		return
	}
	var output bytes.Buffer
	report, e := converter.Convert(ctx, &output, htmlword.Input{HTML: req.HTML, CSS: req.CSS}, req.Options.sdkOptions())
	result := output.Bytes()
	if e != nil {
		s.logger.Printf("convert error: %v", e)
		respond(w, 422, map[string]string{"error": e.Error()})
		return
	}
	filename := safeFilename(req.Filename)
	if r.URL.Query().Get("download") == "1" {
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
		w.Write(result)
		return
	}
	response := map[string]any{"filename": filename, "docx_base64": base64.StdEncoding.EncodeToString(result), "size": len(result), "stats": report.Stats, "warnings": legacyWarnings(report.Diagnostics), "duration_ms": report.Timings.Analysis.Milliseconds() + report.Timings.DOCX.Milliseconds(), "engine": "Go native OOXML"}
	if req.Preview && s.soffice != "" {
		select {
		case s.pdfSlot <- struct{}{}:
			pdf, pages, err := s.renderPDF(ctx, result)
			<-s.pdfSlot
			if err != nil {
				response["preview_error"] = err.Error()
			} else {
				response["pdf_base64"] = base64.StdEncoding.EncodeToString(pdf)
				response["preview_renderer"] = "LibreOffice"
				response["preview_pages"] = pages
			}
		case <-ctx.Done():
			response["preview_error"] = "预览等待超时"
		}
	}
	respond(w, 200, response)
}
func (s *Server) renderPDF(ctx context.Context, docx []byte) ([]byte, []string, error) {
	dir, e := os.MkdirTemp("", "htmlword-pdf-*")
	if e != nil {
		return nil, nil, e
	}
	defer os.RemoveAll(dir)
	if e = os.WriteFile(filepath.Join(dir, "document.docx"), docx, 0600); e != nil {
		return nil, nil, e
	}
	profile := (&url.URL{Scheme: "file", Path: filepath.Join(dir, "profile")}).String()
	child, cancel := context.WithTimeout(ctx, 35*time.Second)
	defer cancel()
	cmd := exec.CommandContext(child, s.soffice, "-env:UserInstallation="+profile, "--headless", "--convert-to", "pdf:writer_pdf_Export", "--outdir", dir, filepath.Join(dir, "document.docx"))
	previewEnv, envErr := previewFontEnv(dir)
	if envErr != nil {
		return nil, nil, envErr
	}
	cmd.Env = previewEnv
	output, e := cmd.CombinedOutput()
	if e != nil {
		return nil, nil, fmt.Errorf("DOCX 已生成，PDF 预览失败：%s %s", e, strings.TrimSpace(string(output)))
	}
	pdf, e := os.ReadFile(filepath.Join(dir, "document.pdf"))
	if e != nil {
		return nil, nil, fmt.Errorf("未能生成 PDF 预览：%s", strings.TrimSpace(string(output)))
	}
	pages := []string{}
	if s.pdftoppm != "" {
		cmd = exec.CommandContext(child, s.pdftoppm, "-f", "1", "-l", "20", "-scale-to", "1300", "-png", filepath.Join(dir, "document.pdf"), filepath.Join(dir, "page"))
		if output, e = cmd.CombinedOutput(); e != nil {
			return nil, nil, fmt.Errorf("分页预览渲染失败：%v %s", e, strings.TrimSpace(string(output)))
		}
		files, _ := filepath.Glob(filepath.Join(dir, "page-*.png"))
		pageNumber := func(p string) int {
			s := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(p), "page-"), ".png")
			n, _ := strconv.Atoi(s)
			return n
		}
		sort.Slice(files, func(i, j int) bool { return pageNumber(files[i]) < pageNumber(files[j]) })
		for _, file := range files {
			b, err := os.ReadFile(file)
			if err != nil {
				return nil, nil, err
			}
			pages = append(pages, base64.StdEncoding.EncodeToString(b))
		}
	}
	return pdf, pages, nil
}

// Headless macOS fontconfig builds may omit modern AssetsV2 font directories.
// A per-job configuration includes them without modifying user font settings.
func previewFontEnv(dir string) ([]string, error) {
	env := os.Environ()
	if runtime.GOOS != "darwin" || os.Getenv("FONTCONFIG_FILE") != "" {
		return env, nil
	}
	dirs := []string{"/System/Library/Fonts", "/Library/Fonts", "/System/Library/AssetsV2/com_apple_MobileAsset_Font7", "/System/Library/AssetsV2/com_apple_MobileAsset_Font6"}
	if extra := os.Getenv("HTMLWORD_FONT_DIRS"); extra != "" {
		dirs = append(dirs, filepath.SplitList(extra)...)
	}
	escape := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;")
	var config strings.Builder
	config.WriteString(`<?xml version="1.0"?><!DOCTYPE fontconfig SYSTEM "urn:fontconfig:fonts.dtd"><fontconfig>`)
	for _, p := range dirs {
		if info, e := os.Stat(p); e == nil && info.IsDir() {
			config.WriteString("<dir>" + escape.Replace(p) + "</dir>")
		}
	}
	cache := filepath.Join(dir, "fontcache")
	if e := os.MkdirAll(cache, 0700); e != nil {
		return nil, e
	}
	config.WriteString("<cachedir>" + escape.Replace(cache) + "</cachedir></fontconfig>")
	file := filepath.Join(dir, "fonts.conf")
	if e := os.WriteFile(file, []byte(config.String()), 0600); e != nil {
		return nil, e
	}
	return append(env, "FONTCONFIG_FILE="+file), nil
}

func (s *Server) Close() {
	if s.converter != nil {
		s.converter.Close()
	}
	if s.remoteConverter != nil {
		s.remoteConverter.Close()
	}
}

type legacyRequest struct {
	HTML     string        `json:"html"`
	CSS      string        `json:"css"`
	Filename string        `json:"filename"`
	Preview  bool          `json:"preview"`
	Options  legacyOptions `json:"options"`
}
type legacyOptions struct {
	PageSize    string  `json:"page_size"`
	Landscape   bool    `json:"landscape"`
	MarginMM    float64 `json:"margin_mm"`
	Font        string  `json:"font"`
	FontSize    float64 `json:"font_size"`
	Media       string  `json:"media"`
	Header      string  `json:"header"`
	Footer      string  `json:"footer"`
	PageNumbers bool    `json:"page_numbers"`
	AllowRemote bool    `json:"allow_remote"`
}

func (o legacyOptions) sdkOptions() htmlword.Options {
	return htmlword.Options{Page: &htmlword.PageSetup{Size: htmlword.PageSize(o.PageSize), Landscape: o.Landscape, MarginMM: o.MarginMM}, Font: o.Font, FontSizePT: o.FontSize, Media: htmlword.Media(o.Media), Header: o.Header, Footer: o.Footer, PageNumbers: o.PageNumbers, FooterMode: htmlword.FooterAuto, SymbolCheckboxes: true}
}

func safeFilename(value string) string {
	value = strings.Map(func(r rune) rune {
		if r < 32 || strings.ContainsRune("/\\:*?\"<>|", r) {
			return '-'
		}
		return r
	}, value)
	value = strings.TrimSpace(value)
	if value == "" {
		value = "document"
	}
	if !strings.HasSuffix(strings.ToLower(value), ".docx") {
		value += ".docx"
	}
	return value
}
func legacyWarnings(ds []htmlword.Diagnostic) []string {
	out := make([]string, 0, len(ds))
	for _, d := range ds {
		out = append(out, d.Message)
	}
	return out
}
