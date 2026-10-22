# Changelog

## Unreleased — SDK preview

- Set the repository/module path to `github.com/xinhaolin/htmlToWord` and document remote installation.

- Expose `New`, `Converter.Convert`, `Close`, configuration, typed errors, diagnostics, statistics and timing through the root Go package.
- Add cancellation, queue/concurrency control, temporary profile cleanup, source/resource/output limits and explicit resource resolvers.
- Make footer promotion and literal checkbox-symbol recognition explicit SDK options.
- Preserve HTML standards mode when installing an asset BaseURL, preventing table metric drift between SDK and CLI usage.
- Preserve the existing native paragraph/table/list engine, native Office charts with embedded XLSX data and checkbox content controls.
- Move the validation frontend to a separate module under `tools/playground` and provide a public-API-only CLI.
- Add English-default and Chinese documentation, runnable examples, rendered DOCX screenshots, regression instructions and CI definitions.

Repository and module path: `github.com/xinhaolin/htmlToWord`. A stable version tag has not been created.
