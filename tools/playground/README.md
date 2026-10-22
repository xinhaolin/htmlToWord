# Local validation playground

This optional frontend is a separate Go module and imports the public HTMLWord SDK using the local `replace` in its `go.mod`. It is not required by SDK consumers.

From this directory:

```sh
go test -race ./...
go run ./cmd/playground -addr 127.0.0.1:8787
```

Chrome/Chromium is required for conversion. Optionally set `HTMLWORD_CHROME` to the executable. PDF comparisons additionally require LibreOffice (`HTMLWORD_SOFFICE`) and Poppler's pdftoppm (`HTMLWORD_PDFTOPPM`), or their executables on PATH. These environment variables belong to this tool; the SDK itself takes explicit `Config` values.

To serve generated validation artifacts, pass `-batch-dir /path/to/artifacts/batch`. The default is `../validation/artifacts/batch`, relative to this tool's working directory. A fresh source archive does not contain generated batch artifacts. Follow [regression instructions](../validation/README.md) to generate them.

Open `http://127.0.0.1:8787/`. The old samples intentionally opt into short-footer promotion and square-symbol checkbox recognition; public SDK defaults preserve those source semantics. The UI's “Word preview” uses a LibreOffice PDF render and cannot demonstrate interactive controls. Download a DOCX to test native client behavior.

Keep the listener on loopback. This is a development tool, not a production upload or conversion service. Fonts must be available separately to both Chromium and the office renderer. The UI remains Chinese while the SDK documentation defaults to English.
