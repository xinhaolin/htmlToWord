# Converting HTML to editable DOCX in Go

HTML export becomes difficult when the output must stay editable. A screenshot preserves appearance but cannot behave like a paragraph. Positioning every text fragment independently can preserve one layout while making later editing awkward. A useful Word export needs a document model that can accept more text and paginate again.

I built HTMLWord as a Go SDK for that workflow. It takes static HTML/CSS, uses Chromium to measure the layout, and writes DOCX with a Go OOXML engine. The current scope includes native text, lists, tables, annotated charts with XLSX data, and checkbox content controls. Chromium is required separately; the Go module itself has no third-party Go dependencies.

## The pipeline

The conversion has three broad stages. First, it parses static HTML and removes scripts, event handlers and embedded active content. Next, a dedicated Chromium process computes layout and styles under the selected page width. Finally, Go maps the measured content into Word paragraphs, runs, tables and Office objects and packages the result as DOCX.

The Chromium process uses an isolated temporary profile and communicates over pipes. It does not use the caller's browser profile. A converter limits concurrent jobs, applies a timeout that includes queue time, propagates cancellation and cleans up the temporary profile.

## Why document flow matters

A report does not end when it is exported. Someone adds an explanation, changes a table cell or inserts several paragraphs during review. The document needs to respond through normal Word layout.

HTMLWord therefore represents ordinary text as paragraphs and runs. Tables remain tables, including merged cells. Many multi-column browser layouts become editable layout tables. This makes the output useful for continued editing, while also introducing visible differences from browser layout.

There is a real tradeoff here. Word and the browser use different font metrics, line-breaking rules and layout structures. Arbitrary HTML cannot be promised pixel-identical after editing across Word, WPS and LibreOffice. The SDK exposes known approximations through structured diagnostics instead of making that promise.

## Charts need actual data

A chart screenshot has no categories or series for an office editor to work with. The SDK accepts explicit data-office-chart metadata and writes an Office chart plus an embedded XLSX workbook. It supports line, column, bar, pie and doughnut charts.

The annotation is deliberately explicit. SVG paths and raster images do not provide a reliable general way to recover the underlying dataset. Supplying categories, numeric values and series names preserves the information the document needs.

HTML checkbox inputs use another native representation: Word checkbox content controls. The initial checked state is retained. A disabled checkbox becomes locked. Browser form submission and JavaScript interactions are outside this static document conversion contract.

## Try it

```sh
go get github.com/xinhaolin/htmlToWord@v0.1.0
```

Create a converter with htmlword.New, call Converter.Convert with a context and an io.Writer, and close the converter when finished. The [getting-started guide](../guides/getting-started.html) contains a complete executable example and the CLI command.

External resources are disabled by default. Inline CSS and data URIs work immediately. Applications can opt into a local filesystem resolver, a public HTTP resolver with a host allowlist, or their own context-aware resolver for an authenticated asset service.

## What the tests establish

The repository includes 12 complex synthetic documents covering reports, specifications, agreements, catalogues and handbooks. The current regression set has 27 original pages, eight charts and embedded workbooks, eight checkboxes and nine pictures. The checks inspect content order, merged cells, chart values and workbook cells rather than only checking that a ZIP file exists.

Three files received 30 additional native paragraphs each and expanded from two to four rendered pages. The post-edit rendering exercises document flow without converting the HTML again. A separate WPS probe verified checkbox toggling and saving. The launch demo adds a fresh WPS 12.9 check: one checkbox toggled, 24 ordinary paragraphs appended, and the document growing from one to three pages. Saved XML confirms the control state and all added paragraphs.

These are bounded results. The complete chart Edit Data UI is not yet verified, and Microsoft Word was unavailable for local acceptance. The gallery contains actual DOCX renders and downloadable files so users can test the clients they care about.

## Where it needs work

There is room to improve CSS fidelity, broader independent fixture coverage and long-running service measurements. Windows pipe transport is currently unsupported. There is no browser pool, so each conversion starts a browser process; throughput depends on the deployment environment and deserves measurement before making performance claims.

Useful feedback includes a small static HTML fixture, the runtime and fonts, the office client/version, and the expected versus observed result. The [repository](https://github.com/xinhaolin/htmlToWord), [example gallery](../examples/index.html), and [release downloads](https://github.com/xinhaolin/htmlToWord/releases/tag/v0.1.0) contain the source and reproducible starting points.
