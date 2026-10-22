# Native Word charts and checkboxes

HTMLWord maps explicit chart data to Office chart parts and an embedded XLSX workbook. Real HTML checkboxes become Word checkbox content controls. These objects are stored in the DOCX package alongside ordinary paragraphs and tables.

## Export chart data

Add data-office-chart to an element with an explicit size. The annotation supplies the categories, series and numeric values; the SDK does not recover chart data from pixels.

```html
<div style="width:520px;height:240px"
  data-office-chart='{
    "type":"column",
    "title":"Quarterly delivery",
    "categories":["Q1","Q2","Q3","Q4"],
    "series":[
      {"name":"Planned","values":[18,24,28,34]},
      {"name":"Delivered","values":[16,23,27,36]}
    ]
  }'></div>
```

Supported types are line, column, bar, pie and doughnut. A chart supports up to 512 categories and 16 series. Pie and doughnut charts require one series, nonnegative values and a positive total.

The exported file contains word/charts parts and an XLSX workbook under word/embeddings. This distinguishes native chart output from a screenshot. If the annotation is invalid, diagnostics explain the failure and the original visual is retained when possible.

## Export selectable checkboxes

```html
<p><label><input type="checkbox" checked> Reviewed</label></p>
<p><label><input type="checkbox"> Approved</label></p>
<p><label><input type="checkbox" disabled> Locked item</label></p>
```

The output uses w14:checkbox content controls and preserves the initial checked state. Disabled inputs are locked. Browser event handlers and form submission logic are not exported.

Literal square symbols such as □ remain text by default. Set Options.SymbolCheckboxes to true only when you deliberately want leading square symbols converted into controls.

## What has been verified

The local structural suite checks chart categories, series values, chart caches and embedded workbook cells. An independent WPS macOS probe recognized a native chart, retained its data after saving, and changed a checkbox from unchecked to checked after a click and save.

The fresh launch demonstration records WPS text editing, page reflow and checkbox interaction. The complete chart Edit Data workflow is not yet verified. Microsoft Word was not available for the current local validation; client support should be tested on the exact versions you deploy.

Download the [operating report](https://github.com/xinhaolin/htmlToWord/releases/download/v0.1.0/sdk-report.docx) and [project checklist](https://github.com/xinhaolin/htmlToWord/releases/download/v0.1.0/sdk-form.docx). Their inputs are included in the [sample archive](https://github.com/xinhaolin/htmlToWord/releases/download/v0.1.0/htmlword-examples.zip).
