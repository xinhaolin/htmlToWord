package htmlword

import "github.com/xinhaolin/htmlToWord/internal/convert"

var descriptions = map[string]string{
	"POSITION_TO_FLOW":         "Absolute/fixed positioning was converted to normal document flow.",
	"TRANSFORM_UNSUPPORTED":    "CSS transforms are not represented in editable content.",
	"DECORATION_APPROXIMATED":  "Shadows, filters, clipping or rounded corners are approximated.",
	"BACKGROUND_UNSUPPORTED":   "CSS background images or gradients are not represented; use img for essential content.",
	"OPACITY_UNSUPPORTED":      "Group opacity is not fully represented.",
	"WRITING_MODE_UNSUPPORTED": "Vertical text was converted to horizontal writing.",
	"SVG_RASTERIZED":           "SVG was embedded as a raster image; internal text and shapes are not individually editable.",
	"CHART_DATA_INVALID":       "Invalid chart metadata was ignored; the original visual content was retained.",
	"CANVAS_UNSUPPORTED":       "Canvas content was not converted into editable document content.",
	"RADIO_AS_TEXT":            "Radio state was preserved as text rather than a mutually exclusive control group.",
	"FOOTER_PROMOTED":          "A short document footer was promoted to a repeated native Word footer.",
	"GRID_SPAN_APPROXIMATED":   "Spanning CSS grid items were approximated with flowing tables.",
	"LAYOUT_AS_TABLE":          "Flex/grid layout was represented using editable layout tables.",
	"DENSE_GRID_FLATTENED":     "A dense layout grid was flattened into flowing content.",
	"COLUMNS_NATIVE":           "CSS columns were represented as native continuous document sections.",
	"CONTENT_SCALED":           "Wide content was scaled to the available page width.",
	"IMAGE_UNAVAILABLE":        "An image could not be loaded or converted; alternative text may be used.",
	"STYLESHEET_UNAVAILABLE":   "An external stylesheet could not be loaded.",
	"ACTIVE_CONTENT_REMOVED":   "Scripts, embedded pages or active content were removed; only static HTML is converted.",
	"RESOURCE_UNAVAILABLE":     "A resource could not be loaded under the configured resolver policy.",
	"NESTED_COLUMNS_FLATTENED": "Columns inside a table were flattened into sequential paragraphs.",
	"IMAGE_LIMIT":              "Additional images exceeded the supported image count.",
	"SIDEBAR_ADAPTED":          "A long sidebar layout was adapted: the rail ends after its content and the main column continues at its original width.",
	"LAYOUT_APPROXIMATED":      "A layout feature was approximated.",
}

func diagnostics(native []convert.Diagnostic) []Diagnostic {
	result := make([]Diagnostic, 0, len(native))
	seen := map[string]bool{}
	for _, d := range native {
		key := d.Code + "\x00" + d.Location + "\x00" + d.Resource
		if seen[key] {
			continue
		}
		seen[key] = true
		message := descriptions[d.Code]
		if message == "" {
			message = "Conversion produced a diagnostic."
		}
		stage := StageAnalysis
		if d.Stage == "docx" {
			stage = StageWrite
		}
		if d.Code == "RESOURCE_UNAVAILABLE" || d.Code == "IMAGE_UNAVAILABLE" || d.Code == "STYLESHEET_UNAVAILABLE" {
			stage = StageResources
		}
		severity := Warning
		if d.Code == "LAYOUT_AS_TABLE" || d.Code == "COLUMNS_NATIVE" {
			severity = Info
		}
		result = append(result, Diagnostic{Code: d.Code, Severity: severity, Stage: stage, Message: message, Location: d.Location, Resource: d.Resource})
	}
	return result
}
