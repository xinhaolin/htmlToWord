package convert

import (
	"archive/zip"
	"bytes"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

const chartNS = "http://schemas.openxmlformats.org/drawingml/2006/chart"
const sheetNS = "http://schemas.openxmlformats.org/spreadsheetml/2006/main"

func (w *writer) checkbox(box Checkbox, style Style, width float64) string {
	w.controlID++
	name := fallback(box.Name, "Checkbox "+strconv.Itoa(w.controlID))
	checked, glyph := "0", "☐"
	if box.Checked {
		checked, glyph = "1", "☒"
	}
	lock := ""
	if box.Disabled {
		lock = `<w:lock w:val="contentLocked"/>`
	}
	style.Font = "Arial"
	style.EastAsia = "Arial"
	style.Background = ""
	style.InlineBox = false
	style.Borders = [4]Border{}
	return `<w:sdt><w:sdtPr><w:alias w:val="` + esc(name) + `"/><w:tag w:val="` + esc(name) + `"/><w:id w:val="` + strconv.Itoa(100000+w.controlID) + `"/>` + lock + `<w14:checkbox><w14:checked w14:val="` + checked + `"/><w14:checkedState w14:val="2612" w14:font="Arial"/><w14:uncheckedState w14:val="2610" w14:font="Arial"/></w14:checkbox></w:sdtPr><w:sdtContent>` + w.run(Run{Text: glyph, Style: style}, width) + `</w:sdtContent></w:sdt>`
}

func (c Chart) validate() error {
	switch c.Type {
	case "line", "bar", "column", "pie", "doughnut":
	default:
		return fmt.Errorf("不支持原生图表类型 %q", c.Type)
	}
	if len(c.Categories) == 0 || len(c.Categories) > 512 || len(c.Series) == 0 || len(c.Series) > 16 {
		return fmt.Errorf("原生图表需要 1–512 个类别及 1–16 个系列")
	}
	if (c.Type == "pie" || c.Type == "doughnut") && len(c.Series) != 1 {
		return fmt.Errorf("饼图与圆环图仅支持一个系列")
	}
	if math.IsNaN(c.Max) || math.IsInf(c.Max, 0) || c.Max < 0 {
		return fmt.Errorf("原生图表坐标范围无效")
	}
	for _, series := range c.Series {
		if len(series.Values) != len(c.Categories) {
			return fmt.Errorf("原生图表系列和类别的数量不一致")
		}
		total := 0.0
		for _, value := range series.Values {
			if math.IsNaN(value) || math.IsInf(value, 0) || math.Abs(value) > 1e100 {
				return fmt.Errorf("原生图表包含无效数值")
			}
			if (c.Type == "pie" || c.Type == "doughnut") && value < 0 {
				return fmt.Errorf("饼图数值不能为负数")
			}
			total += value
		}
		if (c.Type == "pie" || c.Type == "doughnut") && total <= 0 {
			return fmt.Errorf("饼图数值合计必须大于零")
		}
	}
	return nil
}
func chartColor(value, other string) string {
	value = strings.ToUpper(strings.TrimPrefix(value, "#"))
	if len(value) == 6 && strings.Trim(value, "0123456789ABCDEF") == "" {
		return value
	}
	return other
}
func number(value float64) string { return strconv.FormatFloat(value, 'g', -1, 64) }

func (w *writer) chart(c Chart, width float64) string {
	if err := c.validate(); err != nil {
		w.err = err
		return ""
	}
	w.charts++
	n := strconv.Itoa(w.charts)
	workbook, err := chartWorkbook(c)
	if err != nil {
		w.err = err
		return ""
	}
	w.extraParts["word/embeddings/chart"+n+".xlsx"] = workbook
	w.extraParts["word/charts/chart"+n+".xml"] = []byte(chartXML(c))
	w.extraParts["word/charts/_rels/chart"+n+".xml.rels"] = []byte(xmlHead + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="` + relBase + `package" Target="../embeddings/chart` + n + `.xlsx"/></Relationships>`)
	id := w.addRel("chart", "charts/chart"+n+".xml", false)
	x, y := c.Width, c.Height
	if x <= 0 || math.IsNaN(x) || math.IsInf(x, 0) {
		x = width
	}
	if y <= 0 || math.IsNaN(y) || math.IsInf(y, 0) {
		y = x * .45
	}
	if x > width {
		y *= width / x
		x = width
	}
	_, pageHeight := w.options.PageMM()
	maxHeight := (pageHeight - 2*w.options.MarginMM) * 96 / 25.4 * .9
	if y > maxHeight {
		x *= maxHeight / y
		y = maxHeight
	}
	w.drawingID++
	return `<w:r><w:drawing><wp:inline distT="0" distB="0" distL="0" distR="0"><wp:extent cx="` + integer(x*9525) + `" cy="` + integer(y*9525) + `"/><wp:docPr id="` + strconv.Itoa(w.drawingID) + `" name="Editable chart ` + n + `" descr="` + esc(c.Title) + `"/><wp:cNvGraphicFramePr><a:graphicFrameLocks noChangeAspect="1"/></wp:cNvGraphicFramePr><a:graphic><a:graphicData uri="` + chartNS + `"><c:chart xmlns:c="` + chartNS + `" r:id="` + id + `"/></a:graphicData></a:graphic></wp:inline></w:drawing></w:r>`
}

func chartXML(c Chart) string {
	var b strings.Builder
	b.WriteString(xmlHead + `<c:chartSpace xmlns:c="` + chartNS + `" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="` + strings.TrimSuffix(relBase, "/") + `"><c:date1904 val="0"/><c:lang val="zh-CN"/><c:roundedCorners val="0"/><c:chart>`)
	if c.Title != "" {
		b.WriteString(`<c:title><c:tx><c:rich><a:bodyPr/><a:lstStyle/><a:p><a:r><a:rPr lang="zh-CN" sz="1100"/><a:t>` + esc(c.Title) + `</a:t></a:r></a:p></c:rich></c:tx><c:overlay val="0"/></c:title>`)
	} else {
		b.WriteString(`<c:autoTitleDeleted val="1"/>`)
	}
	b.WriteString(`<c:plotArea><c:layout/>`)
	tag := "lineChart"
	switch c.Type {
	case "bar", "column":
		tag = "barChart"
	case "pie":
		tag = "pieChart"
	case "doughnut":
		tag = "doughnutChart"
	}
	b.WriteString(`<c:` + tag + `>`)
	if tag == "barChart" {
		dir := "col"
		if c.Type == "bar" {
			dir = "bar"
		}
		b.WriteString(`<c:barDir val="` + dir + `"/><c:grouping val="clustered"/>`)
	}
	if tag == "lineChart" {
		b.WriteString(`<c:grouping val="standard"/>`)
	}
	circular := c.Type == "pie" || c.Type == "doughnut"
	b.WriteString(`<c:varyColors val="` + strconv.FormatBool(circular) + `"/>`)
	for i, series := range c.Series {
		col := string(rune('B' + i))
		name := fallback(series.Name, "Series "+strconv.Itoa(i+1))
		color := chartColor(series.Color, "397384")
		b.WriteString(`<c:ser><c:idx val="` + strconv.Itoa(i) + `"/><c:order val="` + strconv.Itoa(i) + `"/><c:tx><c:strRef><c:f>Sheet1!$` + col + `$1</c:f><c:strCache><c:ptCount val="1"/><c:pt idx="0"><c:v>` + esc(name) + `</c:v></c:pt></c:strCache></c:strRef></c:tx>`)
		if c.Type == "line" {
			b.WriteString(`<c:spPr><a:ln w="28575"><a:solidFill><a:srgbClr val="` + color + `"/></a:solidFill></a:ln></c:spPr><c:marker><c:symbol val="circle"/><c:size val="4"/><c:spPr><a:solidFill><a:srgbClr val="` + color + `"/></a:solidFill><a:ln><a:noFill/></a:ln></c:spPr></c:marker>`)
		} else {
			b.WriteString(`<c:spPr><a:solidFill><a:srgbClr val="` + color + `"/></a:solidFill><a:ln><a:noFill/></a:ln></c:spPr>`)
		}
		if circular {
			palette := []string{color, "D7E3E9", "8FB4BF", "7391AE", "BBCBA1", "D8B58A"}
			for j := range c.Categories {
				pointColor := palette[j%len(palette)]
				if j < len(series.Colors) {
					pointColor = chartColor(series.Colors[j], pointColor)
				}
				b.WriteString(`<c:dPt><c:idx val="` + strconv.Itoa(j) + `"/><c:spPr><a:solidFill><a:srgbClr val="` + pointColor + `"/></a:solidFill><a:ln><a:noFill/></a:ln></c:spPr></c:dPt>`)
			}
		}
		end := strconv.Itoa(len(c.Categories) + 1)
		b.WriteString(`<c:cat><c:strRef><c:f>Sheet1!$A$2:$A$` + end + `</c:f><c:strCache><c:ptCount val="` + strconv.Itoa(len(c.Categories)) + `"/>`)
		for j, v := range c.Categories {
			b.WriteString(`<c:pt idx="` + strconv.Itoa(j) + `"><c:v>` + esc(v) + `</c:v></c:pt>`)
		}
		b.WriteString(`</c:strCache></c:strRef></c:cat><c:val><c:numRef><c:f>Sheet1!$` + col + `$2:$` + col + `$` + end + `</c:f><c:numCache><c:formatCode>General</c:formatCode><c:ptCount val="` + strconv.Itoa(len(series.Values)) + `"/>`)
		for j, v := range series.Values {
			b.WriteString(`<c:pt idx="` + strconv.Itoa(j) + `"><c:v>` + number(v) + `</c:v></c:pt>`)
		}
		b.WriteString(`</c:numCache></c:numRef></c:val>`)
		if c.Type == "line" {
			b.WriteString(`<c:smooth val="0"/>`)
		}
		b.WriteString(`</c:ser>`)
	}
	if tag == "barChart" {
		b.WriteString(`<c:gapWidth val="65"/><c:overlap val="0"/>`)
	}
	if circular {
		b.WriteString(`<c:firstSliceAng val="270"/>`)
		if c.Type == "doughnut" {
			b.WriteString(`<c:holeSize val="61"/>`)
		}
	} else {
		b.WriteString(`<c:axId val="101"/><c:axId val="102"/>`)
	}
	b.WriteString(`</c:` + tag + `>`)
	if !circular {
		deleted, label := "1", "none"
		if c.ShowAxes {
			deleted, label = "0", "nextTo"
		}
		catPos, valPos := "b", "l"
		if c.Type == "bar" {
			catPos, valPos = "l", "b"
		}
		b.WriteString(`<c:catAx><c:axId val="101"/><c:scaling><c:orientation val="minMax"/></c:scaling><c:delete val="` + deleted + `"/><c:axPos val="` + catPos + `"/><c:majorTickMark val="none"/><c:minorTickMark val="none"/><c:tickLblPos val="` + label + `"/><c:crossAx val="102"/><c:crosses val="autoZero"/><c:auto val="1"/><c:lblAlgn val="ctr"/><c:lblOffset val="100"/></c:catAx>`)
		b.WriteString(`<c:valAx><c:axId val="102"/><c:scaling><c:orientation val="minMax"/>`)
		if c.Max > 0 {
			b.WriteString(`<c:max val="` + number(c.Max) + `"/>`)
		}
		negative := false
		for _, series := range c.Series {
			for _, v := range series.Values {
				negative = negative || v < 0
			}
		}
		if !negative {
			b.WriteString(`<c:min val="0"/>`)
		}
		b.WriteString(`</c:scaling><c:delete val="` + deleted + `"/><c:axPos val="` + valPos + `"/><c:majorGridlines><c:spPr><a:ln w="6350"><a:solidFill><a:srgbClr val="DBE4E8"/></a:solidFill></a:ln></c:spPr></c:majorGridlines><c:numFmt formatCode="General" sourceLinked="1"/><c:majorTickMark val="none"/><c:minorTickMark val="none"/><c:tickLblPos val="` + label + `"/><c:crossAx val="101"/><c:crosses val="autoZero"/><c:crossBetween val="between"/></c:valAx>`)
	}
	b.WriteString(`</c:plotArea>`)
	if c.Legend {
		b.WriteString(`<c:legend><c:legendPos val="b"/><c:layout/><c:overlay val="0"/></c:legend>`)
	}
	b.WriteString(`<c:plotVisOnly val="1"/><c:dispBlanksAs val="gap"/><c:showDLblsOverMax val="0"/></c:chart><c:spPr>`)
	if c.Background != "" {
		b.WriteString(`<a:solidFill><a:srgbClr val="` + chartColor(c.Background, "FFFFFF") + `"/></a:solidFill>`)
	} else {
		b.WriteString(`<a:noFill/>`)
	}
	b.WriteString(`<a:ln><a:noFill/></a:ln></c:spPr><c:txPr><a:bodyPr/><a:lstStyle/><a:p><a:pPr><a:defRPr sz="800"><a:solidFill><a:srgbClr val="64748B"/></a:solidFill><a:latin typeface="Arial"/><a:ea typeface="PingFang SC"/></a:defRPr></a:pPr><a:endParaRPr lang="zh-CN"/></a:p></c:txPr><c:externalData r:id="rId1"><c:autoUpdate val="0"/></c:externalData></c:chartSpace>`)
	return b.String()
}

// A real embedded workbook lets Office's Edit Data command change the chart.
// Text uses inlineStr, so untrusted labels are never spreadsheet formulas.
func chartWorkbook(c Chart) ([]byte, error) {
	var sheet strings.Builder
	sheet.WriteString(xmlHead + `<worksheet xmlns="` + sheetNS + `"><dimension ref="A1:` + string(rune('A'+len(c.Series))) + strconv.Itoa(len(c.Categories)+1) + `"/><sheetViews><sheetView workbookViewId="0"/></sheetViews><sheetFormatPr defaultRowHeight="15"/><sheetData><row r="1">`)
	textCell := func(ref, text string) string {
		return `<c r="` + ref + `" t="inlineStr"><is><t xml:space="preserve">` + esc(text) + `</t></is></c>`
	}
	sheet.WriteString(textCell("A1", "Category"))
	for i, series := range c.Series {
		sheet.WriteString(textCell(string(rune('B'+i))+"1", fallback(series.Name, "Series "+strconv.Itoa(i+1))))
	}
	sheet.WriteString(`</row>`)
	for i, category := range c.Categories {
		row := strconv.Itoa(i + 2)
		sheet.WriteString(`<row r="` + row + `">` + textCell("A"+row, category))
		for j, series := range c.Series {
			sheet.WriteString(`<c r="` + string(rune('B'+j)) + row + `"><v>` + number(series.Values[i]) + `</v></c>`)
		}
		sheet.WriteString(`</row>`)
	}
	sheet.WriteString(`</sheetData></worksheet>`)
	files := map[string]string{
		"[Content_Types].xml":        xmlHead + `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/></Types>`,
		"_rels/.rels":                xmlHead + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="` + relBase + `officeDocument" Target="xl/workbook.xml"/></Relationships>`,
		"xl/workbook.xml":            xmlHead + `<workbook xmlns="` + sheetNS + `" xmlns:r="` + strings.TrimSuffix(relBase, "/") + `"><sheets><sheet name="Sheet1" sheetId="1" r:id="rId1"/></sheets></workbook>`,
		"xl/_rels/workbook.xml.rels": xmlHead + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="` + relBase + `worksheet" Target="worksheets/sheet1.xml"/></Relationships>`,
		"xl/worksheets/sheet1.xml":   sheet.String(),
	}
	files["[Content_Types].xml"] = strings.Replace(files["[Content_Types].xml"], `</Types>`, `<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/></Types>`, 1)
	files["xl/_rels/workbook.xml.rels"] = strings.Replace(files["xl/_rels/workbook.xml.rels"], `</Relationships>`, `<Relationship Id="rId2" Type="`+relBase+`styles" Target="styles.xml"/></Relationships>`, 1)
	files["xl/styles.xml"] = xmlHead + `<styleSheet xmlns="` + sheetNS + `"><fonts count="1"><font><sz val="11"/><name val="Arial"/></font></fonts><fills count="2"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill></fills><borders count="1"><border><left/><right/><top/><bottom/><diagonal/></border></borders><cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs><cellXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/></cellXfs><cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles></styleSheet>`
	var out bytes.Buffer
	z := zip.NewWriter(&out)
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		entry, err := z.Create(name)
		if err != nil {
			return nil, err
		}
		if _, err = entry.Write([]byte(files[name])); err != nil {
			return nil, err
		}
	}
	if err := z.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
