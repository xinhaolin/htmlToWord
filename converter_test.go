package htmlword

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"io"
	"math"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/fstest"
	"time"

	"github.com/xinhaolin/htmlToWord/internal/cdp"
	"github.com/xinhaolin/htmlToWord/internal/convert"
)

func testConverter(t *testing.T, config Config) *Converter {
	t.Helper()
	exe, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	config.BrowserPath = exe
	c, e := New(config)
	if e != nil {
		t.Fatal(e)
	}
	c.analyze = func(ctx context.Context, html, css string, o convert.Options, r convert.Runtime) (convert.Document, error) {
		return convert.Document{Title: "SDK test", Blocks: []convert.Block{{Kind: "paragraph", Runs: []convert.Run{{Text: html}}}}, Stats: convert.Stats{Paragraphs: 1, Characters: len(html)}}, nil
	}
	t.Cleanup(func() { c.Close() })
	return c
}
func TestDefaultsAndDeterministicOutput(t *testing.T) {
	c := testConverter(t, Config{})
	var first, second bytes.Buffer
	for _, out := range []*bytes.Buffer{&first, &second} {
		report, err := c.Convert(context.Background(), out, Input{HTML: "Editable text"}, Options{Deterministic: true})
		if err != nil {
			t.Fatal(err)
		}
		if report.Bytes != int64(out.Len()) || report.Stats.Paragraphs != 1 {
			t.Fatalf("invalid report: %+v", report)
		}
	}
	if !bytes.Equal(first.Bytes(), second.Bytes()) {
		t.Fatal("deterministic output changed")
	}
	z, err := zip.NewReader(bytes.NewReader(first.Bytes()), int64(first.Len()))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range z.File {
		if f.Name == "word/document.xml" {
			r, _ := f.Open()
			raw, _ := io.ReadAll(r)
			r.Close()
			if !bytes.Contains(raw, []byte(`w:top="1134"`)) {
				t.Fatal("zero Options did not use 20mm margins")
			}
		}
	}
}
func TestZeroMarginsAndFiniteOptions(t *testing.T) {
	n, err := nativeOptions(Options{Page: &PageSetup{Size: Letter, MarginMM: 0}})
	if err != nil || n.MarginMM != 0 || n.PageSize != "LETTER" {
		t.Fatalf("zero margin lost: %+v %v", n, err)
	}
	for _, value := range []float64{math.NaN(), math.Inf(1)} {
		if _, err = nativeOptions(Options{FontSizePT: value}); err == nil {
			t.Fatal("non-finite font size accepted")
		}
	}
}
func TestErrorsDoNotWriteOutput(t *testing.T) {
	c := testConverter(t, Config{Limits: Limits{HTMLBytes: 12, OutputBytes: 100}})
	cases := []struct {
		input   Input
		options Options
		want    error
	}{{Input{}, Options{}, ErrInvalidInput}, {Input{HTML: strings.Repeat("x", 13)}, Options{}, ErrLimit}, {Input{HTML: "x", BaseURL: "file:///etc/passwd"}, Options{}, ErrInvalidInput}, {Input{HTML: "x"}, Options{}, ErrLimit}}
	for _, tc := range cases {
		var out bytes.Buffer
		_, err := c.Convert(context.Background(), &out, tc.input, tc.options)
		if !errors.Is(err, tc.want) || out.Len() != 0 {
			t.Fatalf("want %v with untouched output, got %v/%d", tc.want, err, out.Len())
		}
	}
	var detail *Error
	_, err := c.Convert(context.Background(), io.Discard, Input{}, Options{})
	if !errors.As(err, &detail) || detail.Stage != StageInput {
		t.Fatal("typed error missing")
	}
}
func TestStrictModeAndDiagnostics(t *testing.T) {
	c := testConverter(t, Config{})
	c.analyze = func(context.Context, string, string, convert.Options, convert.Runtime) (convert.Document, error) {
		return convert.Document{Diagnostics: []convert.Diagnostic{{Code: "TRANSFORM_UNSUPPORTED", Location: "body > div:nth-of-type(1)"}}}, nil
	}
	var out bytes.Buffer
	report, err := c.Convert(context.Background(), &out, Input{HTML: "x"}, Options{Strict: true})
	if !errors.Is(err, ErrStrict) || out.Len() != 0 || len(report.Diagnostics) != 1 || report.Diagnostics[0].Location == "" || report.Diagnostics[0].Message == "" {
		t.Fatalf("strict result: %+v %v", report, err)
	}
}

func TestNativeNestingFailsWithoutTruncation(t *testing.T) {
	c := testConverter(t, Config{})
	c.analyze = func(context.Context, string, string, convert.Options, convert.Runtime) (convert.Document, error) {
		b := convert.Block{Kind: "paragraph", Runs: []convert.Run{{Text: "must not disappear"}}}
		for i := 0; i < 82; i++ {
			b = convert.Block{Kind: "columns", Columns: 2, Children: []convert.Block{b}}
		}
		return convert.Document{Blocks: []convert.Block{b}}, nil
	}
	var out bytes.Buffer
	if _, err := c.Convert(context.Background(), &out, Input{HTML: "x"}, Options{}); !errors.Is(err, ErrLimit) || out.Len() != 0 {
		t.Fatalf("native nesting: %v / %d", err, out.Len())
	}
	items := diagnostics([]convert.Diagnostic{{Code: "NESTED_COLUMNS_FLATTENED", Stage: "docx"}})
	if len(items) != 1 || items[0].Stage != StageWrite {
		t.Fatalf("writer diagnostic stage: %+v", items)
	}
}

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return len(p) - 1, nil }
func TestOutputFailure(t *testing.T) {
	c := testConverter(t, Config{})
	_, err := c.Convert(context.Background(), shortWriter{}, Input{HTML: "x"}, Options{})
	if !errors.Is(err, ErrOutput) || !errors.Is(err, io.ErrShortWrite) {
		t.Fatal(err)
	}
}
func TestCloseCancelsActiveAndQueuedCalls(t *testing.T) {
	c := testConverter(t, Config{MaxConcurrency: 1})
	entered := make(chan struct{})
	var started atomic.Int32
	c.analyze = func(ctx context.Context, _ string, _ string, _ convert.Options, _ convert.Runtime) (convert.Document, error) {
		if started.Add(1) == 1 {
			close(entered)
		}
		<-ctx.Done()
		return convert.Document{}, ctx.Err()
	}
	results := make(chan error, 2)
	go func() {
		_, err := c.Convert(context.Background(), io.Discard, Input{HTML: "first"}, Options{})
		results <- err
	}()
	<-entered
	go func() {
		_, err := c.Convert(context.Background(), io.Discard, Input{HTML: "second"}, Options{})
		results <- err
	}()
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		select {
		case err := <-results:
			if !errors.Is(err, context.Canceled) && !errors.Is(err, ErrClosed) {
				t.Fatal(err)
			}
		case <-time.After(3 * time.Second):
			t.Fatal("Close did not release a call")
		}
	}
	if started.Load() != 1 {
		t.Fatal("queued call reached analyzer")
	}
	_, err := c.Convert(context.Background(), io.Discard, Input{HTML: "third"}, Options{})
	if !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
}
func TestCanceledContext(t *testing.T) {
	c := testConverter(t, Config{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.Convert(ctx, io.Discard, Input{HTML: "x"}, Options{})
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
func TestConcurrentConversions(t *testing.T) {
	c := testConverter(t, Config{MaxConcurrency: 2})
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var out bytes.Buffer
			_, e := c.Convert(context.Background(), &out, Input{HTML: "parallel"}, Options{})
			if e != nil {
				t.Error(e)
			}
		}()
	}
	wg.Wait()
}
func TestResourceBudgetAndSingleFlight(t *testing.T) {
	var calls atomic.Int32
	l := newResourceLoader(ResourceResolverFunc(func(context.Context, string) (Resource, error) {
		calls.Add(1)
		return Resource{Body: io.NopCloser(strings.NewReader("123456")), ContentType: "text/plain"}, nil
	}), Limits{ResourceBytes: 8, TotalResourceBytes: 10, DOMNodes: 30})
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, _, err := l.load(context.Background(), "https://assets.invalid/one")
			if err != nil || string(data) != "123456" {
				t.Errorf("load: %q %v", data, err)
			}
		}()
	}
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatal("duplicate resource reads")
	}
	if _, _, e := l.load(context.Background(), "https://assets.invalid/two"); !errors.Is(e, ErrLimit) || l.fatal() == nil {
		t.Fatal("total budget not enforced")
	}
}
func TestFSResolverTraversal(t *testing.T) {
	r := FSResolver(fstest.MapFS{"assets/site.css": &fstest.MapFile{Data: []byte("p{color:red}")}})
	resource, err := r.Resolve(context.Background(), "https://assets.invalid/assets/site.css?v=1")
	if err != nil {
		t.Fatal(err)
	}
	resource.Body.Close()
	for _, ref := range []string{"file:///etc/passwd", "https://assets.invalid/../secret", "https://assets.invalid/%2e%2e/secret", "https://assets.invalid/a%5cb"} {
		if result, e := r.Resolve(context.Background(), ref); e == nil {
			result.Body.Close()
			t.Fatalf("accepted %s", ref)
		}
	}
}
func TestRuntimeValidation(t *testing.T) {
	if _, err := New(Config{BrowserPath: filepath.Join(t.TempDir(), "missing")}); !errors.Is(err, ErrRuntimeUnavailable) {
		t.Fatal(err)
	}
	if _, err := New(Config{MaxConcurrency: -1}); !errors.Is(err, ErrInvalidConfig) {
		t.Fatal(err)
	}
}
func TestBrowserRuntimeIntegration(t *testing.T) {
	if os.Getenv("HTMLWORD_INTEGRATION") != "1" || testing.Short() {
		t.Skip("set HTMLWORD_INTEGRATION=1 to run Chromium tests")
	}
	path, err := cdp.FindChrome()
	if err != nil {
		t.Fatal(err)
	}
	temp := t.TempDir()
	assets := FSResolver(fstest.MapFS{"style.css": &fstest.MapFile{Data: []byte("h1{color:#397384}")}})
	c, err := New(Config{BrowserPath: path, TempDir: temp, Resolver: ResourceResolverFunc(func(ctx context.Context, ref string) (Resource, error) {
		// A real stylesheet need not arrive before the first layout evaluation.
		select {
		case <-time.After(200 * time.Millisecond):
			return assets.Resolve(ctx, ref)
		case <-ctx.Done():
			return Resource{}, ctx.Err()
		}
	})})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	html := `<title>SDK integration</title><link rel="stylesheet" href="style.css"><h1>Editable</h1><p><input type="checkbox" checked> Ready</p><div data-office-chart='{"type":"line","categories":["One","Two"],"series":[{"name":"Value","values":[12,23]}]}' style="width:400px;height:180px"></div>`
	var out bytes.Buffer
	report, err := c.Convert(context.Background(), &out, Input{HTML: html, BaseURL: "https://assets.invalid/"}, Options{Deterministic: true})
	if err != nil {
		t.Fatal(err)
	}
	if report.Stats.Charts != 1 || report.Stats.Checkboxes != 1 {
		t.Fatalf("native objects lost: %+v", report)
	}
	packageXML, err := zip.NewReader(bytes.NewReader(out.Bytes()), int64(out.Len()))
	if err != nil {
		t.Fatal(err)
	}
	stylesheetApplied := false
	for _, file := range packageXML.File {
		if file.Name != "word/document.xml" {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Fatal(err)
		}
		stylesheetApplied = bytes.Contains(bytes.ToLower(data), []byte(`w:color w:val="397384"`))
	}
	if !stylesheetApplied {
		t.Fatalf("resolved stylesheet was not applied before extraction: %+v", report.Diagnostics)
	}
	for _, d := range report.Diagnostics {
		if d.Stage == StageResources {
			t.Fatalf("relative stylesheet was not resolved: %+v", d)
		}
	}
	// Public defaults must not silently change literal symbols or repeat a footer.
	out.Reset()
	report, err = c.Convert(context.Background(), &out, Input{HTML: `<p>□ Keep this as text</p><footer>Body footer</footer>`}, Options{})
	if err != nil || report.Stats.Checkboxes != 0 {
		t.Fatalf("default symbols: %+v %v", report, err)
	}
	z, _ := zip.NewReader(bytes.NewReader(out.Bytes()), int64(out.Len()))
	for _, f := range z.File {
		if f.Name == "word/footer1.xml" {
			t.Fatal("body footer was promoted by default")
		}
	}
	out.Reset()
	report, err = c.Convert(context.Background(), &out, Input{HTML: `<p>□ Convert this symbol</p><footer>Repeat this footer</footer>`}, Options{SymbolCheckboxes: true, FooterMode: FooterAuto})
	if err != nil || report.Stats.Checkboxes != 1 {
		t.Fatalf("explicit semantics: %+v %v", report, err)
	}
	// Denied resources are observable and strict mode must leave output untouched.
	offline, err := New(Config{BrowserPath: path})
	if err != nil {
		t.Fatal(err)
	}
	defer offline.Close()
	out.Reset()
	report, err = offline.Convert(context.Background(), &out, Input{HTML: `<p>Offline</p><img src="https://assets.invalid/missing.png" alt="Missing">`}, Options{Strict: true})
	if !errors.Is(err, ErrStrict) || out.Len() != 0 || len(report.Diagnostics) == 0 {
		t.Fatalf("offline strict: %+v %v", report, err)
	}
	limited, err := New(Config{BrowserPath: path, Limits: Limits{DOMNodes: 3}})
	if err != nil {
		t.Fatal(err)
	}
	defer limited.Close()
	if _, err = limited.Convert(context.Background(), io.Discard, Input{HTML: `<p>one</p><p>two</p><p>three</p>`}, Options{}); !errors.Is(err, ErrLimit) {
		t.Fatalf("DOM limit: %v", err)
	}
	entries, err := os.ReadDir(temp)
	if err != nil || len(entries) != 0 {
		t.Fatalf("browser profile leaked: %v %v", entries, err)
	}
}
func FuzzNativeTextXML(f *testing.F) {
	for _, s := range []string{"", "plain text", "中文 & < > \"", "a\x00b\x01c", "🙂\nsecond line"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, text string) {
		if len(text) > 16384 {
			t.Skip()
		}
		c := testConverter(t, Config{})
		c.analyze = func(context.Context, string, string, convert.Options, convert.Runtime) (convert.Document, error) {
			return convert.Document{Title: text, Blocks: []convert.Block{{Kind: "paragraph", Runs: []convert.Run{{Text: text}}}}}, nil
		}
		var out bytes.Buffer
		if _, err := c.Convert(context.Background(), &out, Input{HTML: "<p>fuzz</p>"}, Options{Deterministic: true}); err != nil {
			t.Fatal(err)
		}
		z, err := zip.NewReader(bytes.NewReader(out.Bytes()), int64(out.Len()))
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range z.File {
			if !strings.HasSuffix(file.Name, ".xml") && !strings.HasSuffix(file.Name, ".rels") {
				continue
			}
			r, err := file.Open()
			if err != nil {
				t.Fatal(err)
			}
			decoder := xml.NewDecoder(r)
			for {
				_, err = decoder.Token()
				if err != nil {
					break
				}
			}
			r.Close()
			if !errors.Is(err, io.EOF) {
				t.Fatalf("invalid %s: %v", file.Name, err)
			}
		}
	})
}

func TestBrowserCancellationIntegration(t *testing.T) {
	if os.Getenv("HTMLWORD_INTEGRATION") != "1" || testing.Short() {
		t.Skip("set HTMLWORD_INTEGRATION=1 to run Chromium tests")
	}
	browser, err := cdp.FindChrome()
	if err != nil {
		t.Fatal(err)
	}
	profile := t.TempDir()
	entered := make(chan struct{})
	var once sync.Once
	c, err := New(Config{BrowserPath: browser, TempDir: profile, Resolver: ResourceResolverFunc(func(ctx context.Context, _ string) (Resource, error) {
		once.Do(func() { close(entered) })
		<-ctx.Done()
		return Resource{}, ctx.Err()
	})})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, e := c.Convert(ctx, io.Discard, Input{HTML: `<link rel="stylesheet" href="https://assets.invalid/block.css"><p>Cancel me</p>`}, Options{})
		done <- e
	}()
	select {
	case <-entered:
	case err := <-done:
		t.Fatalf("conversion ended before the blocking resource was requested: %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("browser did not request the blocking resource")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("browser cancellation did not complete")
	}
	entries, err := os.ReadDir(profile)
	if err != nil || len(entries) != 0 {
		t.Fatalf("canceled profile leaked: %v %v", entries, err)
	}
}

func TestBaseURLDoesNotChangeLayoutIntegration(t *testing.T) {
	if os.Getenv("HTMLWORD_INTEGRATION") != "1" || testing.Short() {
		t.Skip("set HTMLWORD_INTEGRATION=1 to run Chromium tests")
	}
	browser, err := cdp.FindChrome()
	if err != nil {
		t.Fatal(err)
	}
	c, err := New(Config{BrowserPath: browser})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	source := `<!doctype html><html><head><style>body{font:14px/1.55 Arial}table{border-collapse:collapse;width:100%;font-size:12px}td{padding:10px;border:1px solid #ddd}</style></head><body><p><label><input type="checkbox" checked> Ready</label></p><table><tr><td>Editable cell</td><td>Wrapped table text</td></tr></table></body></html>`
	var plain, withBase bytes.Buffer
	for _, item := range []struct {
		output *bytes.Buffer
		base   string
	}{{&plain, ""}, {&withBase, "https://assets.invalid/"}} {
		if _, err = c.Convert(context.Background(), item.output, Input{HTML: source, BaseURL: item.base}, Options{Deterministic: true}); err != nil {
			t.Fatal(err)
		}
	}
	if !bytes.Equal(plain.Bytes(), withBase.Bytes()) {
		t.Fatal("BaseURL changed DOCX layout for a document without external assets")
	}
}

func TestHTTPResourceDestinations(t *testing.T) {
	for _, address := range []string{"127.0.0.1", "::1", "10.0.0.1", "100.64.0.1", "169.254.169.254", "192.0.2.1", "2001:db8::1", "::ffff:127.0.0.1"} {
		if publicIP(netip.MustParseAddr(address)) {
			t.Fatalf("accepted non-public address %s", address)
		}
	}
	for _, address := range []string{"1.1.1.1", "2606:4700:4700::1111"} {
		if !publicIP(netip.MustParseAddr(address)) {
			t.Fatalf("rejected public address %s", address)
		}
	}
	r := HTTPResolver{AllowedHosts: []string{"assets.example.com"}}
	for _, ref := range []string{"file:///tmp/a", "https://other.example/a", "https://assets.example.com:8443/a", "https://user:pass@assets.example.com/a"} {
		if _, err := r.Resolve(context.Background(), ref); !errors.Is(err, ErrResource) {
			t.Fatalf("accepted invalid URL %s: %v", ref, err)
		}
	}
}

func TestConfigLimitsAndTimeout(t *testing.T) {
	exe, _ := os.Executable()
	if _, err := New(Config{BrowserPath: exe, Limits: Limits{ResourceBytes: math.MaxInt64}}); !errors.Is(err, ErrInvalidConfig) {
		t.Fatal(err)
	}
	c := testConverter(t, Config{Timeout: 10 * time.Millisecond})
	c.analyze = func(ctx context.Context, _ string, _ string, _ convert.Options, _ convert.Runtime) (convert.Document, error) {
		<-ctx.Done()
		return convert.Document{}, ctx.Err()
	}
	var out bytes.Buffer
	if _, err := c.Convert(context.Background(), &out, Input{HTML: "x"}, Options{}); !errors.Is(err, context.DeadlineExceeded) || out.Len() != 0 {
		t.Fatalf("timeout contract: %v/%d", err, out.Len())
	}
}
