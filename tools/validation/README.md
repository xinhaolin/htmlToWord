# Regression tools

Run commands from the repository root. `testdata/html` contains the static fixture sources; generated documents, PDFs, screenshots and reports live in the ignored `tools/validation/artifacts` directory.

The twelve-case suite covers operating reports, specifications, quotations, handbooks, event guides, workbooks, plans, catalogues and agreements. Use the public-API playground for browser comparisons, or the SDK/CLI directly for consumers without a frontend.

```sh
python3 scripts/generate-complex-suite.py
# Start tools/playground in another terminal. For PDF comparisons install LibreOffice/Poppler.
python3 scripts/batch-qa.py \
  --base http://127.0.0.1:8787 \
  --out tools/validation/artifacts/qa \
  --renderer /path/to/render_docx.py \
  --fontconfig /path/to/fonts.conf
```

`batch-qa.py` needs Python and a DOCX renderer compatible with the documented `render_docx.py INPUT --output_dir DIR --emit_pdf` interface. That renderer is an external development dependency and is not bundled with the SDK. Use a controlled font configuration for comparisons. The report builder also requires manually reviewed page records and comparison screenshots; it does not manufacture visual-review evidence.

The script checks ZIP/XML validity, native text preservation and order, merged cells, links, chart caches and embedded XLSX data, checkbox initial states, and direct insertion of 30 native paragraphs into three samples. It records page counts and source hashes. These checks do not prove browser-to-Word pixel parity or replace office-client interaction/printing acceptance.

The historical batch report remains available in an existing local checkout when its artifacts are retained. It is not included as a freshly verified claim in a clean SDK source distribution. Current SDK scope and evidence are described in [docs/validation.md](../../docs/validation.md).
