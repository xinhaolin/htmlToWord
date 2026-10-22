package convert

import "math"

// Word processors cannot reliably begin a tall, nested, single-row column
// layout partway down a page. Break that row into semantic bands. The shorter
// rail stays in its first cell; the primary reading column flows through rows.
// No source names, selectors, fixed page coordinates or client-sized pages occur.
func (w *writer) paginate(blocks []Block) []Block {
	return w.paginateAt(blocks, 0)
}
func (w *writer) paginateAt(blocks []Block, depth int) []Block {
	result := make([]Block, 0, len(blocks))
	_, pageMM := w.options.PageMM()
	pageHeight := (pageMM - 2*w.options.MarginMM) * 96 / 25.4
	for _, b := range blocks {
		b.Children = w.paginateAt(b.Children, depth+1)
		for i := range b.Rows {
			for j := range b.Rows[i].Cells {
				b.Rows[i].Cells[j].Blocks = w.paginateAt(b.Rows[i].Cells[j].Blocks, depth+1)
			}
		}
		if b.Kind == "table" && b.Tag == "layout" {
			if depth == 0 && len(b.Rows) == 1 {
				bands := segmentRow(b.Rows[0], pageHeight)
				if len(bands) > 1 {
					primary := 0
					for i, c := range b.Rows[0].Cells {
						if cellHeight(c) > cellHeight(b.Rows[0].Cells[primary]) {
							primary = i
						}
					}
					offset, col := b.Style.Margin[3], 0
					for _, c := range b.Rows[0].Cells[:primary] {
						for j := 0; j < c.Span && col < len(b.Grid); j++ {
							offset += b.Grid[col]
							col++
						}
					}
					first := b
					first.Rows = bands[:1]
					first.Style.Margin[2] = 0
					result = append(result, first)
					for _, band := range bands[1:] {
						for _, child := range band.Cells[primary].Blocks {
							child.Style.Margin[3] += offset
							result = append(result, child)
						}
					}
					w.warning("长侧栏布局的首段保留双栏，侧栏内容结束后主栏按原宽度连续排版，以支持原生表格跨页和重复表头。")
					continue
				}
			}
			var rows []Row
			for _, row := range b.Rows {
				rows = append(rows, segmentRow(row, pageHeight)...)
			}
			b.Rows = rows
		}
		result = append(result, b)
	}
	return result
}

func flowHeight(b Block) float64 {
	height := b.Style.Height
	if b.Kind == "table" && len(b.Rows) > 0 {
		height = 0
		for _, row := range b.Rows {
			rh := row.Height
			for _, cell := range row.Cells {
				rh = math.Max(rh, cellHeight(cell))
			}
			height += rh
		}
	}
	if height <= 0 {
		height = math.Max(1, b.Style.LineHeight)
		if height <= 1 {
			height = math.Max(16, b.Style.Size*1.4)
		}
	}
	return height + b.Style.Margin[0] + b.Style.Margin[2]
}
func cellHeight(cell Cell) float64 {
	h := cell.Style.Padding[0] + cell.Style.Padding[2]
	for _, b := range cell.Blocks {
		h += flowHeight(b)
	}
	return h
}

func segmentRow(row Row, pageHeight float64) []Row {
	if len(row.Cells) < 2 || row.Exact {
		return []Row{row}
	}
	primary := -1
	largest := 0.0
	for i, c := range row.Cells {
		if h := cellHeight(c); h > largest {
			primary, largest = i, h
		}
	}
	if largest < pageHeight*.85 {
		return []Row{row}
	}
	railHeight := 0.0
	for i, c := range row.Cells {
		if i != primary {
			railHeight = math.Max(railHeight, cellHeight(c))
		}
	}
	// Two equally tall prose columns need a different, interleaved reading model.
	// Leave them splittable rather than reorder their contents silently.
	if railHeight > pageHeight*.65 {
		return []Row{row}
	}
	budget := math.Max(pageHeight*.40, railHeight)
	source := row.Cells[primary].Blocks
	var groups [][]Block
	var group []Block
	used := 0.0
	for _, b := range source {
		keep := len(group) > 0 && (group[len(group)-1].Level > 0 || group[len(group)-1].Style.KeepNext)
		if len(group) > 0 && used+flowHeight(b) > budget && !keep {
			groups = append(groups, group)
			group = nil
			used = 0
		}
		group = append(group, b)
		used += flowHeight(b)
	}
	if len(group) > 0 {
		groups = append(groups, group)
	}
	if len(groups) < 2 {
		return []Row{row}
	}
	rows := make([]Row, 0, len(groups))
	for i, blocks := range groups {
		next := Row{Cells: make([]Cell, len(row.Cells))}
		for j, c := range row.Cells {
			c.Style.Borders[0] = Border{}
			c.Style.Borders[2] = Border{}
			if i == 0 {
				c.Style.Borders[0] = row.Cells[j].Style.Borders[0]
			} else {
				c.Style.Padding[0] = 0
				c.Blocks = nil
			}
			if i == len(groups)-1 {
				c.Style.Borders[2] = row.Cells[j].Style.Borders[2]
			} else {
				c.Style.Padding[2] = 0
			}
			if j == primary {
				c.Blocks = blocks
			}
			next.Cells[j] = c
		}
		height := 0.0
		for _, cell := range next.Cells {
			height = math.Max(height, cellHeight(cell))
		}
		next.Keep = height < pageHeight*.75
		rows = append(rows, next)
	}
	return rows
}
