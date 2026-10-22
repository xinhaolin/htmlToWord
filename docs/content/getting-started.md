# Convert HTML to Word in Go

HTMLWord converts static HTML and CSS into editable DOCX. It uses Chromium to measure the browser layout and a Go OOXML writer to produce paragraphs, lists, tables, native charts and checkbox content controls. Text remains document content so an office editor can reflow it after an edit.

## Install

Install Go 1.23 or later and Chrome or Chromium, then add the SDK to your Go module:

```sh
go get github.com/xinhaolin/htmlToWord@v0.1.0
```

The supported runtime platforms are macOS and Linux. Chromium is a separate executable dependency; there are no third-party Go module dependencies. Fonts used by the source page should also be available to the target office client.

## A complete conversion

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/xinhaolin/htmlToWord"
)

func main() {
    converter, err := htmlword.New(htmlword.Config{})
    if err != nil { log.Fatal(err) }
    defer converter.Close()

    output, err := os.Create("document.docx")
    if err != nil { log.Fatal(err) }
    defer output.Close()

    report, err := converter.Convert(context.Background(), output,
        htmlword.Input{HTML: `<h1>Project notes</h1>
            <p>This paragraph remains editable.</p>
            <p><input type="checkbox"> Reviewed</p>`},
        htmlword.Options{})
    if err != nil { log.Fatal(err) }
    log.Printf("%d bytes, %d diagnostics", report.Bytes, len(report.Diagnostics))
}
```

Run `go run .`, then open `document.docx` in your target Word/WPS client. Adding a paragraph should remain an ordinary document operation. Check native control behavior in the actual client version you intend to support.

## Select Chromium and a page size

```go
converter, err := htmlword.New(htmlword.Config{
    BrowserPath: "/path/to/chromium",
    MaxConcurrency: 2,
})
// Handle err, defer converter.Close(), then reuse it.
options := htmlword.Options{
    Page: &htmlword.PageSetup{Size: htmlword.A4, MarginMM: 20},
    PageNumbers: true,
}
```

An empty BrowserPath searches common executable locations. The zero Options value uses A4 portrait, 20 mm margins and print media. A non-nil PageSetup with MarginMM set to zero explicitly requests zero margins.

## Load local CSS and images explicitly

External resources are disabled by default. Inline CSS and data URIs work without a resolver. To load assets from a controlled directory:

```go
converter, err := htmlword.New(htmlword.Config{
    Resolver: htmlword.FSResolver(os.DirFS("assets")),
})
// Handle err and close the converter when finished.
input := htmlword.Input{
    HTML: `<link rel="stylesheet" href="style.css"><h1>Report</h1>`,
    BaseURL: "https://assets.invalid/",
}
```

The directory belongs to the caller. os.DirFS is not an operating-system sandbox; symlink and filesystem access policy remain your responsibility. For public network assets, HTTPResolver supports an explicit host allowlist. Use a custom resolver for authenticated resources and honor context cancellation.

## Inspect approximations

Report.Diagnostics contains structured codes, severity, stage and best-effort DOM locations. Use Options.Strict to reject warning diagnostics before document bytes are written. A successful strict conversion means no reported warnings, not a guarantee of identical browser and Word pixels.

## Use the CLI

```sh
go install github.com/xinhaolin/htmlToWord/cmd/htmlword@v0.1.0
htmlword -input report.html -output report.docx
```

The CLI writes through a temporary file before renaming. The SDK accepts a caller-owned writer; a writer I/O error can leave partial bytes, so callers should implement their own rollback when necessary.

Continue with [native charts and checkboxes](native-objects.html), [pagination and fidelity](pagination.html), or the [complex example gallery](../examples/index.html).
