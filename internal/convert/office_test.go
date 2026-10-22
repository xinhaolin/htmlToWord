package convert

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNativeChartsAndControls(t *testing.T) {
	for _, kind := range []string{"line", "column", "bar", "pie", "doughnut"} {
		t.Run(kind, func(t *testing.T) {
			chart := Chart{Type: kind, Categories: []string{"A & B", "第二期"}, Series: []ChartSeries{{Name: "预算", Values: []float64{12.5, 23}, Color: "397384"}}, Width: 500, Height: 190}
			doc := Document{Blocks: []Block{{Kind: "paragraph", Runs: []Run{{Chart: &chart}}}, {Kind: "paragraph", Runs: []Run{{Checkbox: &Checkbox{Name: "Ready", Checked: true}}, {Checkbox: &Checkbox{Name: "Pending"}}}}}}
			data, err := Build(context.Background(), &doc, DefaultOptions())
			if err != nil {
				t.Fatal(err)
			}
			f := unzip(t, data)
			if directory := os.Getenv("HTMLWORD_QA_OUT"); directory != "" {
				if err := os.MkdirAll(directory, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(directory, kind+".docx"), data, 0644); err != nil {
					t.Fatal(err)
				}
			}
			if !strings.Contains(f["word/document.xml"], "<c:chart ") || strings.Count(f["word/document.xml"], "<w14:checkbox>") != 2 {
				t.Fatal("missing native chart or checkbox controls")
			}
			if strings.Contains(f["word/document.xml"], "<pic:pic>") || strings.Contains(f["word/document.xml"], "documentProtection") {
				t.Fatal("native controls must not become screenshots or require protection")
			}
			if !strings.Contains(f["word/charts/chart1.xml"], "<c:externalData r:id=\"rId1\"") || !strings.Contains(f["word/charts/_rels/chart1.xml.rels"], "../embeddings/chart1.xlsx") {
				t.Fatal("chart editing workbook missing")
			}
			workbook := []byte(f["word/embeddings/chart1.xlsx"])
			reader, err := zip.NewReader(bytes.NewReader(workbook), int64(len(workbook)))
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, file := range reader.File {
				r, _ := file.Open()
				raw, _ := io.ReadAll(r)
				r.Close()
				if strings.HasSuffix(file.Name, ".xml") {
					decoder := xml.NewDecoder(bytes.NewReader(raw))
					for {
						_, e := decoder.Token()
						if e == io.EOF {
							break
						}
						if e != nil {
							t.Fatalf("invalid workbook XML: %v", e)
						}
					}
				}
				if file.Name == "xl/worksheets/sheet1.xml" {
					found = strings.Contains(string(raw), "12.5") && strings.Contains(string(raw), "23") && strings.Contains(string(raw), "A &amp; B")
				}
			}
			if !found {
				t.Fatal("editable workbook data changed")
			}
		})
	}
}

func TestInvalidChartDataRejected(t *testing.T) {
	chart := Chart{Type: "line", Categories: []string{"One"}, Series: []ChartSeries{{Values: []float64{math.NaN()}}}}
	doc := Document{Blocks: []Block{{Kind: "paragraph", Runs: []Run{{Chart: &chart}}}}}
	if _, err := Build(context.Background(), &doc, DefaultOptions()); err == nil {
		t.Fatal("non-finite chart data accepted")
	}
}

func TestOfficeSourceSemantics(t *testing.T) {
	if testing.Short() || os.Getenv("HTMLWORD_INTEGRATION") != "1" {
		t.Skip("browser integration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	html := `<style>body{font:12px Arial}</style><p><label><input type="checkbox" checked name="ready"> Ready</label></p><p>□ Pending</p><table data-office-chart="column"><thead><tr><th>Month</th><th>Sales</th></tr></thead><tbody><tr><td>Jan</td><td>12.5</td></tr><tr><td>Feb</td><td>23</td></tr></tbody></table><div data-office-chart='{"type":"line","categories":["Q1","Q2"],"series":[{"name":"Revenue","values":[17,29]}]}' style="width:300px;height:120px"></div>`
	doc, err := Analyze(ctx, html, "", DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if doc.Stats.Charts != 2 || doc.Stats.Checkboxes != 2 {
		t.Fatalf("source controls not recognized: %+v", doc.Stats)
	}
	data, err := Build(ctx, &doc, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	f := unzip(t, data)
	if !strings.Contains(f["word/document.xml"], "Ready") || !strings.Contains(f["word/document.xml"], "Pending") || !strings.Contains(f["word/document.xml"], "<w:tbl>") {
		t.Fatal("labels or source data table lost")
	}
	if strings.Count(f["word/document.xml"], "<c:chart ") != 2 || strings.Count(f["word/document.xml"], "<w14:checkbox>") != 2 {
		t.Fatal("source semantics not preserved in OOXML")
	}
}
