# Contributing

This is an experimental v0.x SDK hosted at [xinhaolin/htmlToWord](https://github.com/xinhaolin/htmlToWord). Use [issues](https://github.com/xinhaolin/htmlToWord/issues) for reproducible problems and pull requests for proposed changes. The v0.x API is experimental and may change before v1.

## Set up

Use Go 1.23+, then run `go test -race ./...` and `go vet ./...`. Unit tests do not require a browser. Install Chrome/Chromium to run `HTMLWORD_INTEGRATION=1 go test -race ./...`. The optional playground is a separate module; run its tests from `tools/playground`.

Use `go run ./examples/basic` for a public API example. Keep browser path, temporary directories, resource policy and budgets configurable through the public API. The SDK must not start an HTTP server, require LibreOffice or change process environment variables.

## Changes and tests

- Describe a concrete input, observed output and intended behavior. Add a small, synthetic regression fixture for conversion changes. Include the full static HTML/CSS needed to reproduce a layout problem.
- Preserve ordinary Word flow and editability. Do not add branches keyed to a sample filename, CSS class, person or template identifier.
- When using native chart data, retain the actual categories, series and values. Do not infer numbers from a screenshot or label a raster image editable.
- Test cancellation, cleanup and concurrency when changing browser or resource lifecycle code. Custom resolvers must honor contexts and close their underlying resources.
- Keep public errors inspectable with `errors.Is`/`errors.As`. Add stable diagnostic codes when introducing a lossy conversion. Do not silently drop content to make a comparison look better.
- Run real DOCX renders for layout changes and inspect every page. Record renderer, fonts, page settings and input hashes. XML validity alone does not establish visual correctness or office-client editing behavior.
- Keep English and Chinese READMEs aligned. Screenshots must come from current output, with provenance recorded in `docs/assets/manifest.json`.

The fixtures are fictional demonstration material. Do not contribute personal resumes, access tokens, private URLs, proprietary fonts or unlicensed assets. External runtime binaries and fonts are not bundled with this source distribution.

## Before each release

Review API names and licensing, inspect the repository CI results, configure a private security reporting channel, and verify supported Word/WPS versions. Create an immutable version tag only after those checks and document any remaining client limitations in the release notes. Hosted results are available from the repository Actions page; local rendering evidence has its own documented scope.
