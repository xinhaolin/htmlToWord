package htmlword

import (
	"context"
	"io"
	"time"
)

// Config controls a converter instance. Zero fields select documented defaults.
// It is copied by New. Resolver implementations must be safe for concurrent use
// and must honor context cancellation. No process environment is changed.
type Config struct {
	BrowserPath    string
	TempDir        string
	MaxConcurrency int
	Timeout        time.Duration
	Limits         Limits
	Resolver       ResourceResolver
}

// Limits bounds inputs, resource reads and generated compressed DOCX bytes.
// It is not an operating-system process memory or CPU quota.
type Limits struct {
	HTMLBytes          int64
	ResourceBytes      int64
	TotalResourceBytes int64
	DOMNodes           int
	Depth              int
	Images             int
	OutputBytes        int64
}

// DefaultLimits returns fresh, conservative limits for one conversion.
func DefaultLimits() Limits {
	return Limits{HTMLBytes: 8 << 20, ResourceBytes: 8 << 20, TotalResourceBytes: 32 << 20, DOMNodes: 16000, Depth: 80, Images: 100, OutputBytes: 32 << 20}
}

// Input contains a static document. BaseURL resolves relative assets; it does
// not load the HTML at that URL. Source scripts are never executed.
type Input struct {
	HTML    string
	CSS     string
	BaseURL string
	Title   string
}

// PageSetup uses millimetres for margins. A nil Options.Page selects A4 portrait
// with 20 mm margins. A non-nil PageSetup explicitly permits MarginMM == 0.
type PageSetup struct {
	Size      PageSize
	Landscape bool
	MarginMM  float64
}
type PageSize string

const (
	A4     PageSize = "A4"
	A3     PageSize = "A3"
	Letter PageSize = "LETTER"
)

type Media string

const (
	Print  Media = "print"
	Screen Media = "screen"
)

type FooterMode string

const (
	FooterAuto     FooterMode = "auto"
	FooterPreserve FooterMode = "preserve"
)

// Options controls document output. Zero Options uses A4/20mm/Arial 11pt/print,
// preserves source footer content in the body and leaves literal square symbols
// as text. HTML input[type=checkbox] always produces native checkbox controls.
type Options struct {
	Page             *PageSetup
	Font             string
	FontSizePT       float64
	Media            Media
	Header           string
	Footer           string
	PageNumbers      bool
	Strict           bool
	FooterMode       FooterMode
	SymbolCheckboxes bool
	// Deterministic fixes package creation time. It does not normalize differences
	// between Chromium versions, fonts, operating systems or office renderers.
	Deterministic bool
	CreatedAt     time.Time
}

// Validate checks document options without starting Chromium.
func (o Options) Validate() error {
	_, err := nativeOptions(o)
	if err != nil {
		return failure(StageInput, ErrInvalidInput, err)
	}
	return nil
}

type Stage string

const (
	StageInput     Stage = "input"
	StageQueue     Stage = "queue"
	StageAnalysis  Stage = "analysis"
	StageResources Stage = "resources"
	StageWrite     Stage = "docx"
	StageOutput    Stage = "output"
)

type Severity string

const (
	Info    Severity = "info"
	Warning Severity = "warning"
)

// Diagnostic is a stable machine code with an English description. Location is
// an input DOM path when known. Resource is the absolute resource URL when known.
// Do not assume Location is present for package-level diagnostics.
type Diagnostic struct {
	Code     string
	Severity Severity
	Stage    Stage
	Message  string
	Location string
	Resource string
}
type Stats struct {
	Paragraphs int `json:"paragraphs"`
	Tables     int `json:"tables"`
	Images     int `json:"images"`
	Charts     int `json:"charts"`
	Checkboxes int `json:"checkboxes"`
	Characters int `json:"characters"`
	Lists      int `json:"lists"`
}
type Timings struct {
	Queue    time.Duration
	Analysis time.Duration
	DOCX     time.Duration
	Output   time.Duration
	Total    time.Duration
}
type Report struct {
	Diagnostics []Diagnostic
	Stats       Stats
	Timings     Timings
	Bytes       int64
	Runtime     RuntimeInfo
}
type RuntimeInfo struct {
	BrowserPath string
	Platform    string
}

// Resource bodies are closed by the SDK, including failed or oversized reads.
type Resource struct {
	Body        io.ReadCloser
	ContentType string
}
type ResourceResolver interface {
	Resolve(context.Context, string) (Resource, error)
}
type ResourceResolverFunc func(context.Context, string) (Resource, error)

func (f ResourceResolverFunc) Resolve(ctx context.Context, ref string) (Resource, error) {
	return f(ctx, ref)
}
