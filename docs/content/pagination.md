# Editable DOCX pagination and layout fidelity

A browser lays out a page on a continuous canvas. Word lays out paragraphs, tables, sections and objects across physical pages. HTMLWord measures the first model and writes structures in the second, preserving document flow wherever possible.

## Keep content in normal flow

Use headings, paragraphs, lists and tables for content that must grow after export. Many flex and grid arrangements become editable layout tables. Adding text can then change line wrapping, row height and page count in the office editor.

Absolute positions, transforms and clipping do not have a universal equivalent in a flowing Word document. They may be approximated and reported through diagnostics. A layout that looks correct in one font environment can wrap differently in another.

## Choose page settings explicitly

```go
options := htmlword.Options{
    Page: &htmlword.PageSetup{
        Size: htmlword.A4,
        MarginMM: 20,
        Landscape: false,
    },
    PageNumbers: true,
}
```

A4, A3 and Letter are supported. Print media is the default; select screen media deliberately when the source page needs it. Include the required fonts in both the conversion environment and the target office environment.

## Footers and literal symbols are explicit choices

HTML footer content remains in the document body by default. FooterMode set to FooterAuto opts into promotion of a short footer to repeated Word footer content. Explicit Header and Footer options create native repeated content. These choices affect editing semantics, so they are not guessed silently by the public SDK.

## Read the validation correctly

The 12 complex regression fixtures contain 27 original pages, eight native charts and workbooks, eight checkbox controls and nine pictures. Three documents received 30 additional native paragraphs each and grew from two to four rendered pages. These edits did not rerun the source HTML.

All 39 rendered pages matched the earlier reviewed engine baseline in the same macOS/font/LibreOffice environment. This is regression evidence for the SDK refactor. It does not mean the HTML and Word versions have identical pixels, or that every office client renders them identically.

## Known boundaries

SVG is embedded as an image, so text inside SVG is not native editable text. Gradients, CSS background images, shadows, complex spanning grids and long-sidebar decoration have incomplete fidelity. Scripts and embedded pages are removed; canvas and animation are not executed. Supply static content and explicit chart data.

Use the [gallery](../examples/index.html) to inspect real output. The full [validation scope](../validation.md) records test environments and remaining client checks. Submit a minimal synthetic HTML fixture with the expected and actual result when reporting a conversion issue.
