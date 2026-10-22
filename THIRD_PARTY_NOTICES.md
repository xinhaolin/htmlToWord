# Third-party runtimes and assets

HTMLWord's root Go module currently imports only the Go standard library and its own internal packages. It does not embed TurboDocx/html-to-docx, a commercial DOCX SDK, Chromium, office applications or font binaries.

| Component | Relationship | Upstream information |
| --- | --- | --- |
| Go standard library/toolchain | Build/runtime libraries | [Go license](https://go.dev/LICENSE) |
| Chrome / Chromium | Separately installed layout runtime | [Chromium licenses](https://www.chromium.org/developers/how-tos/get-the-code/), [Chrome terms](https://www.google.com/chrome/terms/) |
| LibreOffice | Optional local DOCX/PDF verification | [LibreOffice licenses](https://www.libreoffice.org/about-us/licenses/) |
| Poppler | Optional local PDF page rendering | [Poppler project](https://poppler.freedesktop.org/) |
| actions/checkout, actions/setup-go | CI tools fetched by GitHub Actions | [checkout](https://github.com/actions/checkout), [setup-go](https://github.com/actions/setup-go) |

These components are not relicensed by this project's Apache-2.0 license. Their own terms apply when installed or redistributed. Users are responsible for font availability and any font embedding/distribution rights. Screenshot images show rendered output; no proprietary font files or third-party runtime executables are included in the SDK source archive.

HTML fixtures and demonstration values in `testdata/html` are synthetic material supplied with this project. Local generated reports, historical UI captures and client probes are kept under ignored validation artifacts and are not included in the SDK source archive. This notice is an inventory of the current source distribution, not a complete license audit of a future deployment image.
