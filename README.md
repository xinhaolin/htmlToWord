# HTMLWord

[English](README.md) · [简体中文](README.zh-CN.md)

A Go SDK for converting **static HTML and CSS into editable, flowing DOCX documents**. Text, lists and tables remain native document content. Adding paragraphs in an office editor can reflow the document and create new pages.

HTMLWord uses Chromium to measure layout and a Go OOXML writer to build the document. It has **no third-party Go module dependencies**. Chrome or Chromium is a required runtime; LibreOffice and the validation frontend are optional development tools.

Repository: [xinhaolin/htmlToWord](https://github.com/xinhaolin/htmlToWord) · License: [Apache-2.0](LICENSE)

## Actual output

These are screenshots of DOCX files generated through the public SDK and rendered by LibreOffice. They are not browser screenshots or a claim of identical rendering in every Word/WPS version. Inputs and image provenance are included in the repository.

![Operating review: editable text, native chart and merged table](docs/assets/sdk-report.png)

[HTML source](testdata/html/sdk-report.html) · Native Office chart with an embedded XLSX workbook; flowing paragraphs and merged cells.

![Project checklist: native checkbox controls and editable action table](docs/assets/sdk-form.png)

[HTML source](testdata/html/sdk-form.html) · Four native checkbox content controls, an editable table and space for additional paragraphs. A static screenshot cannot demonstrate toggling; see the [validation scope](docs/validation.md).

## Requirements and installation

- Go **1.23 or later**.
- Chrome or Chromium installed separately, with the required fonts available to Chromium and the target office client.
- macOS or Linux. The current pipe transport explicitly rejects Windows. The local rendering baseline was exercised on macOS; the [CI workflow](https://github.com/xinhaolin/htmlToWord/actions/workflows/ci.yml) also checks Linux.

Chromium's sandbox remains enabled. Run as a normal user with a working sandbox; container users and kernel permissions must support that runtime requirement.

Install the SDK in your Go module:

```sh
go get github.com/xinhaolin/htmlToWord@latest
```

Import `github.com/xinhaolin/htmlToWord` and use its Go package name, `htmlword`.

For local SDK development, replace the module with your checkout:

```sh
go mod edit -replace=github.com/xinhaolin/htmlToWord=../htmlToWord
```

Adjust the final directory to your checkout and remove the replacement when consuming the remote module.

## Quick start

```go
package main

import (
    "bytes"
    "context"
    "log"
    "os"

    "github.com/xinhaolin/htmlToWord"
)

func main() {
    converter, err := htmlword.New(htmlword.Config{})
    if err != nil { log.Fatal(err) }
    defer converter.Close()

    var output bytes.Buffer
    report, err := converter.Convert(context.Background(), &output,
        htmlword.Input{
            HTML: `<h1>Project notes</h1><p>Editable, flowing text.</p>
                   <p><input type="checkbox" checked> Reviewed</p>`,
            Title: "Project notes",
        }, htmlword.Options{})
    if err != nil { log.Fatal(err) }
    if err := os.WriteFile("document.docx", output.Bytes(), 0644); err != nil {
        log.Fatal(err)
    }
    log.Printf("%d bytes, %d diagnostics", report.Bytes, len(report.Diagnostics))
}
```

[Runnable basic example](examples/basic/main.go) · [Local asset example](examples/assets/main.go)

`Config.BrowserPath` selects a specific Chrome/Chromium executable; otherwise common executable names and macOS application locations are searched. The SDK does not change process environment variables. `Config.TempDir` selects an existing directory for temporary browser profiles.

## Conversion contract

```go
converter, err := htmlword.New(htmlword.Config{
    BrowserPath: "/path/to/chromium", // optional if auto-discovered
    MaxConcurrency: 2,
    Timeout: 70 * time.Second,
})
// Handle err, defer converter.Close(), then reuse the converter.
report, err := converter.Convert(ctx, writer, input, htmlword.Options{
    Page: &htmlword.PageSetup{Size: htmlword.A4, MarginMM: 20},
    PageNumbers: true,
    Strict: false,
})
```

The snippet uses `time`, a caller-owned `ctx`, `writer` and `input`. `Input.HTML` is the source string, `Input.CSS` adds a stylesheet, and `Input.BaseURL` resolves relative assets without fetching the HTML at that URL.

- Zero `Options` means A4 portrait, 20 mm margins, Arial 11 pt and print media. A non-nil `PageSetup` makes `MarginMM: 0` an explicit zero margin. A4, A3 and Letter, landscape and screen media are supported.
- Source scripts, embedded pages and event handlers are removed. Use already-rendered static markup for client-rendered applications.
- HTML footer content stays in the body by default. `FooterMode: htmlword.FooterAuto` opts into promotion of a short footer to a repeated Word footer. Explicit `Header` and `Footer` options create native repeated content.
- Real HTML checkboxes become native controls. Literal `□`, `☐`, `☑`, `☒` remain text unless `SymbolCheckboxes: true` is set.
- `Options.Validate()` validates output settings without launching a browser.
- A converter is safe for concurrent calls. Each conversion starts an isolated browser process and profile; **there is no browser pool**. The concurrency limit queues excess calls. The configured timeout includes queue time.
- `Close()` cancels active and queued calls and waits for cleanup. Do not call it from a resolver callback. Caller-provided resolvers and output writers must obey cancellation and concurrency contracts; a blocking custom `Read`/`Write` cannot be forcibly interrupted by the SDK.
- Input, strict-mode and generation failures leave the writer untouched. A writer I/O error or cancellation during output may leave partial bytes; callers own rollback. The CLI writes through a temporary file before renaming.
- `Deterministic: true` fixes package creation time. `CreatedAt` overrides it when set. Reproducibility still depends on the same fonts, Chromium, OS and input; it is not a cross-platform pixel guarantee.

### Resources and limits

External resources are **disabled by default**. Inline CSS and data URIs work without a resolver. Supply a resolver to allow specific assets:

```go
converter, err := htmlword.New(htmlword.Config{
    Resolver: htmlword.FSResolver(os.DirFS("assets")),
})
// Handle err and close the converter. For href="style.css", use:
input := htmlword.Input{HTML: source, BaseURL: "https://assets.invalid/"}
```

`FSResolver` checks URL paths for traversal and reads only through the supplied `fs.FS`. **The caller controls filesystem permissions and symlinks**; `os.DirFS` is not an OS sandbox. Embedded filesystems are useful for packaged assets.

For public network resources, explicitly use `HTTPResolver{AllowedHosts: []string{"assets.example.com"}}`. Hosts match exactly. An empty allowlist allows any public host. The resolver checks redirects and DNS destinations, rejects private/loopback/link-local and other reserved ranges and nonstandard ports, and ignores proxy environment variables. For authenticated or private resources, provide a `ResourceResolver` with your own policy. Returned bodies are closed by the SDK, and custom implementations must honor `context.Context`.

Resources are cached once per URL per conversion, with at most four simultaneous resolver reads. The byte budget covers external resolver responses, including CSS, fonts and images. Data URIs count toward source bytes; their decoded representation is not part of the external-resource total.

| Limit | Default |
| --- | ---: |
| HTML + extra CSS | 8 MiB |
| One external resource | 8 MiB |
| All external resources | 32 MiB |
| DOM elements | 16,000 |
| Traversal depth | 80 |
| Images | 100 |
| Compressed DOCX output | 32 MiB |

Set `Config.Limits` to override defaults. Zero fields select defaults; negatives are invalid. Supported ceilings are 64 MiB source, 100,000 DOM elements, depth 80 and 100 images; other byte limits cannot exceed 1 GiB. These controls are **not OS memory or CPU quotas**. Apply process/container isolation separately when operating a public conversion service.

### Diagnostics and errors

`Report` contains `Diagnostics`, `Stats`, `Timings`, output bytes and runtime information. Diagnostics expose a stable `Code`, English `Message`, `Severity`, `Stage`, and a DOM `Location` or resource URL when available. Diagnostic locations are best-effort, not a complete CSS audit.

`Options.Strict` rejects warning diagnostics and returns `ErrStrict` before writing. Informational mappings such as `LAYOUT_AS_TABLE` and `COLUMNS_NATIVE` do not trigger strict rejection. Strict success means no reported warnings, not certified pixel parity.

```go
if errors.Is(err, htmlword.ErrStrict) {
    // Inspect report.Diagnostics and decide whether approximations are acceptable.
}
var detail *htmlword.Error
if errors.As(err, &detail) {
    log.Printf("stage=%s kind=%v cause=%v", detail.Stage, detail.Kind, detail.Cause)
}
```

The error sentinels include `ErrInvalidInput`, `ErrInvalidConfig`, `ErrRuntimeUnavailable`, `ErrUnsupportedPlatform`, `ErrClosed`, `ErrLimit`, `ErrResource`, `ErrConversion`, `ErrStrict` and `ErrOutput`. Context cancellation and deadline errors are preserved through `errors.Is`. A denied or missing resource normally produces a warning; use strict mode to reject it. Limit violations fail conversion regardless of strict mode.

## Fidelity and native objects

| HTML/CSS feature | DOCX representation / current boundary |
| --- | --- |
| Text, headings, emphasis, hyperlinks, lists | Native editable paragraphs and runs |
| Tables, rowspan/colspan, repeating headers | Native tables; pagination follows the office layout engine |
| Many flex/grid layouts | Editable layout tables; complex spanning and dense grids may be approximated |
| Long sidebar / native CSS columns | Flowing structures; long rail adaptation may change its decorative extent |
| Images / SVG | Embedded images; SVG is rasterized and its internal text is not editable |
| Annotated charts | Native Office charts plus embedded XLSX data |
| HTML checkbox | Native `w14:checkbox` content control, preserving initial checked/disabled state |
| Radio, text input, textarea, select | Static text/value representation; no full Word form model |
| Absolute/fixed position, transforms, clipping | Converted to flow or approximated, with diagnostics |
| Gradients, CSS background images, shadows | Incomplete; essential visual content should use explicit images |
| Canvas, animation, scripts, iframes | Not an execution/rendering target; provide static content |

Flowing editable DOCX and browser pixels have different layout models. Arbitrary HTML cannot be promised a pixel-identical result after editing, across fonts or across Word/WPS/LibreOffice. The SDK reports known losses and preserves document flow instead of fixing every glyph at an absolute position.

### Native charts

Provide the chart's semantic data on an element with an explicit size. A browser visual may be placed inside it for an HTML preview; the annotated element is replaced by an Office chart during export.

```html
<div style="width:600px;height:240px" data-office-chart='{
  "type":"column",
  "categories":["Q1","Q2","Q3","Q4"],
  "series":[{"name":"Delivered","values":[18,23,28,36],"color":"#365e72"}]
}'></div>
```

Supported types: `line`, `column`, `bar`, `pie`, `doughnut`. Up to 512 categories and 16 series are supported; pie/doughnut require one series and nonnegative values with a positive total. Optional fields include `title`, `background`, `max` and per-series `colors`. Arbitrary SVG or screenshots are **not reverse-engineered into chart data**. Invalid metadata emits a diagnostic and retains the original visual when possible. Office client support for chart data editing must be verified in the target client.

### Native checkboxes

```html
<p><label><input type="checkbox" name="reviewed" checked> Reviewed</label></p>
<p><label><input type="checkbox" name="approved"> Approved</label></p>
```

These export as selectable content controls, not screenshots. A disabled checkbox is exported locked. HTML submission and browser form event logic are not carried into Word.

## Development and optional tools

```sh
go test -race ./...                 # unit tests, no browser required
go vet ./...
go test -run '^$' -fuzz FuzzNativeTextXML -fuzztime 5s .
HTMLWORD_INTEGRATION=1 go test -race ./...  # requires Chrome/Chromium
go run ./examples/basic
go run ./cmd/htmlword -input testdata/html/sdk-report.html -output report.docx
```

The CLI accepts `-chrome`, `-css`, `-assets`, `-base-url`, `-allow-remote`, `-strict`, page settings and explicit footer/symbol policies. `go run ./cmd/htmlword -h` lists all flags.

The optional frontend lives in a **separate Go module**, and imports the public SDK:

```sh
cd tools/playground
go test -race ./...
go run ./cmd/playground -addr 127.0.0.1:8787
```

It is a local validation tool, not part of the SDK API or a hosted conversion service. It deliberately enables legacy footer/symbol recognition for its old samples. DOCX export requires Chrome; PDF comparison additionally needs separately installed LibreOffice and Poppler. [Tool instructions](tools/playground/README.md) · [Regression instructions](tools/validation/README.md)

```text
htmlword/
  *.go                   Public SDK, configuration, resolver, errors
  internal/              Chromium transport and OOXML conversion engine
  cmd/htmlword/          Optional file conversion CLI
  examples/              Public API examples
  testdata/html/         Synthetic static HTML fixtures
  docs/                  Screenshots, provenance and validation scope
  tools/playground/      Separate module: local validation frontend
  tools/validation/      Regression tooling and ignored local artifacts
```

See [validation evidence and remaining gaps](docs/validation.md), [contribution guidelines](CONTRIBUTING.md), [change log](CHANGELOG.md) and [third-party runtime notices](THIRD_PARTY_NOTICES.md). Hosted check results are available from [GitHub Actions](https://github.com/xinhaolin/htmlToWord/actions/workflows/ci.yml); local rendering evidence is documented separately.
