package convert

import (
	"errors"
	"math"
	"strings"
)

type Options struct {
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

func DefaultOptions() Options {
	return Options{PageSize: "A4", MarginMM: 20, Font: "Arial", FontSize: 11, Media: "print"}
}
func (o *Options) Validate() error {
	o.PageSize = strings.ToUpper(o.PageSize)
	if o.PageSize == "" {
		o.PageSize = "A4"
	}
	if o.PageSize != "A4" && o.PageSize != "LETTER" && o.PageSize != "A3" {
		return errors.New("page_size 必须为 A4、A3 或 Letter")
	}
	if math.IsNaN(o.MarginMM) || math.IsInf(o.MarginMM, 0) || o.MarginMM < 0 || o.MarginMM > 60 {
		return errors.New("页边距必须在 0–60 mm 之间")
	}
	if o.Font == "" {
		o.Font = "Arial"
	}
	if len(o.Font) > 200 {
		return errors.New("字体名称过长")
	}
	if o.FontSize == 0 {
		o.FontSize = 11
	}
	if o.FontSize < 5 || o.FontSize > 72 {
		return errors.New("默认字号必须在 5–72 pt 之间")
	}
	if o.Media == "" {
		o.Media = "print"
	}
	if o.Media != "screen" && o.Media != "print" {
		return errors.New("media 必须为 screen 或 print")
	}
	if len(o.Header) > 2000 || len(o.Footer) > 2000 {
		return errors.New("页眉页脚最多 2000 字节")
	}
	return nil
}
func (o Options) PageMM() (float64, float64) {
	w, h := 210.0, 297.0
	if o.PageSize == "LETTER" {
		w, h = 215.9, 279.4
	}
	if o.PageSize == "A3" {
		w, h = 297, 420
	}
	if o.Landscape {
		return h, w
	}
	return w, h
}
func (o Options) ContentPX() float64 { w, _ := o.PageMM(); return (w - 2*o.MarginMM) * 96 / 25.4 }

type Border struct {
	Width float64 `json:"width"`
	Style string  `json:"style"`
	Color string  `json:"color"`
}
type Style struct {
	Font          string     `json:"font"`
	EastAsia      string     `json:"east_asia"`
	Size          float64    `json:"size"`
	Color         string     `json:"color"`
	Background    string     `json:"background"`
	Bold          bool       `json:"bold"`
	Italic        bool       `json:"italic"`
	Underline     bool       `json:"underline"`
	Strike        bool       `json:"strike"`
	Vertical      string     `json:"vertical"`
	Align         string     `json:"align"`
	Direction     string     `json:"direction"`
	WhiteSpace    string     `json:"white_space"`
	LineHeight    float64    `json:"line_height"`
	LetterSpacing float64    `json:"letter_spacing"`
	TextIndent    float64    `json:"text_indent"`
	Margin        [4]float64 `json:"margin"`
	Padding       [4]float64 `json:"padding"`
	Borders       [4]Border  `json:"borders"`
	Width         float64    `json:"width"`
	Height        float64    `json:"height"`
	BreakBefore   bool       `json:"break_before"`
	BreakAfter    bool       `json:"break_after"`
	Keep          bool       `json:"keep"`
	KeepNext      bool       `json:"keep_next,omitempty"`
	NoWrap        bool       `json:"no_wrap,omitempty"`
	InlineBox     bool       `json:"inline_box,omitempty"`
}
type Run struct {
	Text     string    `json:"text,omitempty"`
	Style    Style     `json:"style"`
	Link     string    `json:"link,omitempty"`
	Break    bool      `json:"break,omitempty"`
	Image    *Picture  `json:"image,omitempty"`
	Chart    *Chart    `json:"chart,omitempty"`
	Checkbox *Checkbox `json:"checkbox,omitempty"`
}

// Chart contains source data, rather than values guessed from rendered pixels.
type Chart struct {
	Type       string        `json:"type"`
	Title      string        `json:"title,omitempty"`
	Categories []string      `json:"categories"`
	Series     []ChartSeries `json:"series"`
	Width      float64       `json:"width"`
	Height     float64       `json:"height"`
	Background string        `json:"background,omitempty"`
	Legend     bool          `json:"legend,omitempty"`
	ShowAxes   bool          `json:"show_axes,omitempty"`
	Max        float64       `json:"max,omitempty"`
}
type ChartSeries struct {
	Name   string    `json:"name"`
	Values []float64 `json:"values"`
	Color  string    `json:"color,omitempty"`
	Colors []string  `json:"colors,omitempty"`
}
type Checkbox struct {
	Checked  bool   `json:"checked"`
	Disabled bool   `json:"disabled,omitempty"`
	Name     string `json:"name,omitempty"`
}
type Picture struct {
	Source  string  `json:"source"`
	Alt     string  `json:"alt"`
	Width   float64 `json:"width"`
	Height  float64 `json:"height"`
	Outline Border  `json:"outline"`
	Radius  float64 `json:"radius,omitempty"`
}
type TabStop struct {
	Position float64 `json:"position"`
	Align    string  `json:"align"`
}
type Block struct {
	Kind       string    `json:"kind"`
	Tag        string    `json:"tag,omitempty"`
	ID         string    `json:"id,omitempty"`
	Style      Style     `json:"style"`
	Runs       []Run     `json:"runs,omitempty"`
	Children   []Block   `json:"children,omitempty"`
	Rows       []Row     `json:"rows,omitempty"`
	Level      int       `json:"level,omitempty"`
	ListID     int       `json:"list_id,omitempty"`
	ListLevel  int       `json:"list_level,omitempty"`
	Ordered    bool      `json:"ordered,omitempty"`
	Start      int       `json:"start,omitempty"`
	ListFormat string    `json:"list_format,omitempty"`
	Columns    int       `json:"columns,omitempty"`
	Gap        float64   `json:"gap,omitempty"`
	Grid       []float64 `json:"grid,omitempty"`
	Tabs       []TabStop `json:"tabs,omitempty"`
}
type Row struct {
	Cells  []Cell  `json:"cells"`
	Header bool    `json:"header,omitempty"`
	Keep   bool    `json:"keep,omitempty"`
	Height float64 `json:"height,omitempty"`
	Exact  bool    `json:"exact,omitempty"`
}
type Cell struct {
	Blocks  []Block `json:"blocks"`
	Style   Style   `json:"style"`
	Span    int     `json:"span"`
	RowSpan int     `json:"row_span"`
	VMerge  string  `json:"v_merge,omitempty"`
	VAlign  string  `json:"v_align,omitempty"`
}
type Document struct {
	Title       string       `json:"title"`
	Blocks      []Block      `json:"blocks"`
	Footer      []Block      `json:"footer,omitempty"`
	Warnings    []string     `json:"warnings"`
	Stats       Stats        `json:"stats"`
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
}
type Diagnostic struct {
	Stage    string `json:"stage,omitempty"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	Location string `json:"location,omitempty"`
	Resource string `json:"resource,omitempty"`
}
type Stats struct {
	Paragraphs int `json:"paragraphs"`
	Tables     int `json:"tables"`
	Images     int `json:"images"`
	Characters int `json:"characters"`
	Lists      int `json:"lists"`
	Charts     int `json:"charts"`
	Checkboxes int `json:"checkboxes"`
}
type Request struct {
	HTML     string  `json:"html"`
	CSS      string  `json:"css"`
	Filename string  `json:"filename"`
	Options  Options `json:"options"`
	Preview  bool    `json:"preview"`
}
type Result struct {
	Data       []byte
	Document   Document
	DurationMS int64
}
