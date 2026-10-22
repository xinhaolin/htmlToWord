package htmlword

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/xinhaolin/htmlToWord/internal/cdp"
	"github.com/xinhaolin/htmlToWord/internal/convert"
)

// Converter is safe for concurrent Convert calls. Each call has an isolated
// Chromium process, profile and document model; MaxConcurrency limits active
// calls. Close cancels active/queued calls and waits for their cleanup.
// Chromium pooling is intentionally not implemented in this first SDK release.
type Converter struct {
	config   Config
	slots    chan struct{}
	lifetime context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
	closed   bool
	active   sync.WaitGroup
	analyze  func(context.Context, string, string, convert.Options, convert.Runtime) (convert.Document, error)
}

func New(config Config) (*Converter, error) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return nil, failure(StageInput, ErrUnsupportedPlatform, nil)
	}
	if config.MaxConcurrency < 0 || config.Timeout < 0 {
		return nil, failure(StageInput, ErrInvalidConfig, errors.New("negative concurrency or timeout"))
	}
	if config.MaxConcurrency == 0 {
		config.MaxConcurrency = 2
	}
	if config.MaxConcurrency > 128 {
		return nil, failure(StageInput, ErrInvalidConfig, errors.New("concurrency must not exceed 128"))
	}
	if config.Timeout == 0 {
		config.Timeout = 70 * time.Second
	}
	defaults := DefaultLimits()
	limits := &config.Limits
	for _, pair := range [][2]*int64{{&limits.HTMLBytes, &defaults.HTMLBytes}, {&limits.ResourceBytes, &defaults.ResourceBytes}, {&limits.TotalResourceBytes, &defaults.TotalResourceBytes}, {&limits.OutputBytes, &defaults.OutputBytes}} {
		if *pair[0] < 0 {
			return nil, failure(StageInput, ErrInvalidConfig, errors.New("negative byte limit"))
		}
		if *pair[0] == 0 {
			*pair[0] = *pair[1]
		}
		if *pair[0] > 1<<30 {
			return nil, failure(StageInput, ErrInvalidConfig, errors.New("byte limits must not exceed 1 GiB"))
		}
	}
	for _, pair := range [][2]*int{{&limits.DOMNodes, &defaults.DOMNodes}, {&limits.Depth, &defaults.Depth}, {&limits.Images, &defaults.Images}} {
		if *pair[0] < 0 {
			return nil, failure(StageInput, ErrInvalidConfig, errors.New("negative structural limit"))
		}
		if *pair[0] == 0 {
			*pair[0] = *pair[1]
		}
	}
	// The native image writer has an independent 100-image safety ceiling.
	if limits.Images > 100 || limits.Depth > 80 || limits.HTMLBytes > 64<<20 || limits.DOMNodes > 100000 {
		return nil, failure(StageInput, ErrInvalidConfig, errors.New("limits exceed supported ceilings (100 images, depth 80, 64 MiB HTML, 100000 nodes)"))
	}
	if config.TempDir != "" {
		info, err := os.Stat(config.TempDir)
		if err != nil || !info.IsDir() {
			return nil, failure(StageInput, ErrInvalidConfig, errors.New("TempDir must be an existing directory"))
		}
	}
	browser, err := cdp.FindChromePath(config.BrowserPath)
	if err != nil {
		return nil, failure(StageInput, ErrRuntimeUnavailable, err)
	}
	config.BrowserPath = browser
	lifetime, cancel := context.WithCancel(context.Background())
	return &Converter{config: config, slots: make(chan struct{}, config.MaxConcurrency), lifetime: lifetime, cancel: cancel, analyze: convert.AnalyzeWithRuntime}, nil
}

func (c *Converter) Runtime() RuntimeInfo {
	return RuntimeInfo{BrowserPath: c.config.BrowserPath, Platform: runtime.GOOS + "/" + runtime.GOARCH}
}
func (c *Converter) Close() error {
	c.mu.Lock()
	if !c.closed {
		c.closed = true
		c.cancel()
	}
	c.mu.Unlock()
	c.active.Wait()
	return nil
}

// Convert writes one DOCX after analysis, generation and strict checks succeed.
// Parse, limit and strict errors leave output untouched. An output I/O failure
// or cancellation while writing can leave partial bytes: callers own rollback.
// output and any shared ResourceResolver must support the caller's concurrency.
func (c *Converter) Convert(ctx context.Context, output io.Writer, input Input, options Options) (report Report, err error) {
	started := time.Now()
	report.Runtime = c.Runtime()
	defer func() { report.Timings.Total = time.Since(started) }()
	if ctx == nil || output == nil {
		return report, failure(StageInput, ErrInvalidInput, errors.New("context and output are required"))
	}
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return report, failure(StageInput, ErrClosed, nil)
	}
	c.active.Add(1)
	c.mu.Unlock()
	defer c.active.Done()
	job, cancel := context.WithTimeout(ctx, c.config.Timeout)
	stop := context.AfterFunc(c.lifetime, cancel)
	defer stop()
	defer cancel()
	if err = job.Err(); err != nil {
		return report, failure(StageInput, ErrConversion, err)
	}
	if strings.TrimSpace(input.HTML) == "" {
		return report, failure(StageInput, ErrInvalidInput, errors.New("HTML is empty"))
	}
	if int64(len(input.HTML))+int64(len(input.CSS)) > c.config.Limits.HTMLBytes {
		return report, failure(StageInput, ErrLimit, errors.New("HTML/CSS bytes"))
	}
	if input.BaseURL != "" {
		u, e := url.Parse(input.BaseURL)
		if e != nil || !u.IsAbs() || u.Host == "" || u.User != nil || (u.Scheme != "https" && u.Scheme != "http") {
			return report, failure(StageInput, ErrInvalidInput, errors.New("BaseURL must be an absolute HTTP(S) URL without credentials"))
		}
	}
	opts, e := nativeOptions(options)
	if e != nil {
		return report, failure(StageInput, ErrInvalidInput, e)
	}
	queue := time.Now()
	select {
	case c.slots <- struct{}{}:
		defer func() { <-c.slots }()
	case <-job.Done():
		return report, failure(StageQueue, ErrConversion, job.Err())
	}
	report.Timings.Queue = time.Since(queue)
	if e = job.Err(); e != nil {
		return report, failure(StageQueue, ErrConversion, e)
	}
	resources := newResourceLoader(c.config.Resolver, c.config.Limits)
	execution := convert.Runtime{Browser: cdp.Options{Path: c.config.BrowserPath, TempDir: c.config.TempDir}, BaseURL: input.BaseURL, LoadResource: resources.load, MaxNodes: c.config.Limits.DOMNodes, MaxDepth: c.config.Limits.Depth, MaxImages: c.config.Limits.Images, MaxOutputBytes: c.config.Limits.OutputBytes, ExplicitSemantics: true, FooterMode: string(options.FooterMode), SymbolCheckboxes: options.SymbolCheckboxes}
	if execution.FooterMode == "" {
		execution.FooterMode = "preserve"
	}
	if options.Deterministic {
		fixed := time.Unix(0, 0).UTC()
		execution.CreatedAt = &fixed
	}
	if !options.CreatedAt.IsZero() {
		fixed := options.CreatedAt
		execution.CreatedAt = &fixed
	}
	stage := time.Now()
	doc, e := c.analyze(job, input.HTML, input.CSS, opts, execution)
	report.Timings.Analysis = time.Since(stage)
	report.Diagnostics = diagnostics(doc.Diagnostics)
	report.Stats = statistics(doc.Stats)
	if limitErr := resources.fatal(); limitErr != nil {
		return report, failure(StageResources, ErrLimit, limitErr)
	}
	if e != nil {
		return report, conversionError(job, StageAnalysis, e)
	}
	if input.Title != "" {
		doc.Title = input.Title
	}
	if options.Strict && hasWarnings(report.Diagnostics) {
		return report, failure(StageAnalysis, ErrStrict, nil)
	}
	stage = time.Now()
	data, e := convert.BuildWithRuntime(job, &doc, opts, execution)
	report.Timings.DOCX = time.Since(stage)
	report.Diagnostics = diagnostics(doc.Diagnostics)
	if e != nil {
		return report, conversionError(job, StageWrite, e)
	}
	if limitErr := resources.fatal(); limitErr != nil {
		return report, failure(StageResources, ErrLimit, limitErr)
	}
	if options.Strict && hasWarnings(report.Diagnostics) {
		return report, failure(StageWrite, ErrStrict, nil)
	}
	stage = time.Now()
	reader := bytes.NewReader(data)
	buf := make([]byte, 32<<10)
	for reader.Len() > 0 {
		if e = job.Err(); e != nil {
			break
		}
		var n int
		n, e = reader.Read(buf)
		if e != nil {
			break
		}
		var written int
		written, e = output.Write(buf[:n])
		if written < 0 || written > n {
			e = io.ErrShortWrite
			written = 0
		}
		report.Bytes += int64(written)
		if e != nil {
			break
		}
		if written != n {
			e = io.ErrShortWrite
			break
		}
	}
	report.Timings.Output = time.Since(stage)
	if e != nil {
		return report, failure(StageOutput, ErrOutput, e)
	}
	return report, nil
}

func conversionError(ctx context.Context, stage Stage, err error) error {
	if ctx.Err() != nil {
		return failure(stage, ErrConversion, ctx.Err())
	}
	if errors.Is(err, convert.ErrOutputLimit) || strings.Contains(err.Error(), "HTMLWORD_LIMIT:") {
		return failure(stage, ErrLimit, err)
	}
	return failure(stage, ErrConversion, err)
}
func nativeOptions(o Options) (convert.Options, error) {
	n := convert.DefaultOptions()
	if o.Page != nil {
		n.PageSize = string(o.Page.Size)
		n.MarginMM = o.Page.MarginMM
		n.Landscape = o.Page.Landscape
	}
	if o.Font != "" {
		n.Font = o.Font
	}
	if o.FontSizePT != 0 {
		n.FontSize = o.FontSizePT
	}
	if o.Media != "" {
		n.Media = string(o.Media)
	}
	if math.IsNaN(o.FontSizePT) || math.IsInf(o.FontSizePT, 0) {
		return n, errors.New("font size must be finite")
	}
	if o.FooterMode != "" && o.FooterMode != FooterAuto && o.FooterMode != FooterPreserve {
		return n, fmt.Errorf("invalid footer mode %q", o.FooterMode)
	}
	n.Header = o.Header
	n.Footer = o.Footer
	n.PageNumbers = o.PageNumbers
	return n, n.Validate()
}
func statistics(s convert.Stats) Stats {
	return Stats{Paragraphs: s.Paragraphs, Tables: s.Tables, Images: s.Images, Charts: s.Charts, Checkboxes: s.Checkboxes, Characters: s.Characters, Lists: s.Lists}
}
func hasWarnings(ds []Diagnostic) bool {
	for _, d := range ds {
		if d.Severity == Warning {
			return true
		}
	}
	return false
}
