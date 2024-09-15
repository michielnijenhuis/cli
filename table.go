package cli

import (
	"fmt"
	"iter"
	"log"
	"regexp"
	"slices"
	"strings"

	"github.com/michielnijenhuis/cli/helper"
)

const (
	SeparatorTop                 = 0
	SeparatorTopBottom           = 1
	SeparatorMid                 = 2
	SeparatorBottom              = 3
	BorderOutside                = 0
	BorderInside                 = 1
	DisplayOrientationDefault    = "default"
	DisplayOrientationHorizontal = "horizontal"
	DisplayOrientationVertical   = "vertical"
	TableCellAlignLeft           = "left"
	TableCellAlignRight          = "right"
	TableCellAlignCenter         = "center"
	DefaultTableCellAlign        = TableCellAlignLeft
)

type TableCellStyle struct {
	Foreground string
	Background string
	Options    []string
	// left, center or right
	Align      string
	CellFormat string
}

type TableCell struct {
	Value       string
	RowSpan     uint
	ColSpan     uint
	Style       *TableCellStyle
	IsSeparator bool
}

func (c *TableCell) String() string {
	return c.Value
}

type TableStyle struct {
	PaddingChar                 string
	HorizontalOutsideBorderChar string
	HorizontalInsideBorderChar  string
	VerticalOutsideBorderChar   string
	VerticalInsideBorderChar    string
	CrossingChar                string
	CrossingTopRightChar        string
	CrossingTopMidChar          string
	CrossingTopLeftChar         string
	CrossingMidRightChar        string
	CrossingBottomRightChar     string
	CrossingBottomMidChar       string
	CrossingBottomLeftChar      string
	CrossingMidLeftChar         string
	CrossingTopLeftBottomChar   string
	CrossingTopMidBottomChar    string
	CrossingTopRightBottomChar  string
	HeaderTitleFormat           string
	FooterTitleFormat           string
	CellHeaderFormat            string
	CellRowFormat               string
	CellRowContentFormat        string
	BorderFormat                string
	PadType                     string
}

func (ts *TableStyle) SetDefaultCrossingChar(char string) {
	ts.CrossingChar = char
	ts.CrossingTopLeftChar = char
	ts.CrossingTopMidChar = char
	ts.CrossingTopRightChar = char
	ts.CrossingMidRightChar = char
	ts.CrossingBottomRightChar = char
	ts.CrossingBottomMidChar = char
	ts.CrossingBottomLeftChar = char
	ts.CrossingMidLeftChar = char
	ts.CrossingTopLeftBottomChar = char
	ts.CrossingTopMidBottomChar = char
	ts.CrossingTopRightBottomChar = char
}

func NewTableCell(value string) *TableCell {
	return &TableCell{
		Value:   value,
		RowSpan: 1,
		ColSpan: 1,
		Style:   nil,
	}
}

var defaultTableStyles map[string]*TableStyle

const (
	crossingChar                = "┼"
	horizontalOutsideBorderChar = "─"
	verticalBorderChar          = "│"
)

func makeDefaultTableStyles() map[string]*TableStyle {
	borderless := NewTableStyle("")
	borderless.HorizontalOutsideBorderChar = "="
	borderless.HorizontalInsideBorderChar = "="
	borderless.VerticalOutsideBorderChar = " "
	borderless.VerticalInsideBorderChar = " "
	borderless.SetDefaultCrossingChar(" ")

	compact := NewTableStyle("")
	compact.HorizontalOutsideBorderChar = ""
	compact.HorizontalInsideBorderChar = ""
	compact.VerticalOutsideBorderChar = ""
	compact.VerticalInsideBorderChar = ""
	compact.CellRowFormat = "%s "

	styleGuide := NewTableStyle("")
	styleGuide.HorizontalOutsideBorderChar = "-"
	styleGuide.HorizontalInsideBorderChar = "-"
	styleGuide.VerticalOutsideBorderChar = " "
	styleGuide.VerticalInsideBorderChar = " "
	styleGuide.SetDefaultCrossingChar(" ")
	styleGuide.CellRowFormat = "%s"

	box := NewTableStyle("")
	box.HorizontalOutsideBorderChar = horizontalOutsideBorderChar
	box.HorizontalInsideBorderChar = horizontalOutsideBorderChar
	box.VerticalOutsideBorderChar = verticalBorderChar
	box.VerticalInsideBorderChar = verticalBorderChar
	box.CrossingChar = crossingChar
	box.CrossingTopLeftChar = "┌"
	box.CrossingTopMidChar = "┬"
	box.CrossingTopRightChar = "┐"
	box.CrossingMidRightChar = "┤"
	box.CrossingBottomRightChar = "┘"
	box.CrossingBottomMidChar = "┴"
	box.CrossingBottomLeftChar = "└"
	box.CrossingMidLeftChar = "├"
	box.CrossingTopLeftBottomChar = "├"
	box.CrossingTopMidBottomChar = crossingChar
	box.CrossingTopRightBottomChar = "┤"

	boxDouble := NewTableStyle("")
	boxDouble.HorizontalOutsideBorderChar = "═"
	boxDouble.HorizontalInsideBorderChar = horizontalOutsideBorderChar
	boxDouble.VerticalOutsideBorderChar = "║"
	boxDouble.VerticalInsideBorderChar = verticalBorderChar
	boxDouble.CrossingChar = crossingChar
	boxDouble.CrossingTopLeftChar = "╔"
	boxDouble.CrossingTopMidChar = "╤"
	boxDouble.CrossingTopRightChar = "╗"
	boxDouble.CrossingMidRightChar = "╢"
	boxDouble.CrossingBottomRightChar = "╝"
	boxDouble.CrossingBottomMidChar = "╧"
	boxDouble.CrossingBottomLeftChar = "╚"
	boxDouble.CrossingMidLeftChar = "╟"
	boxDouble.CrossingTopLeftBottomChar = "╠"
	boxDouble.CrossingTopMidBottomChar = "╪"
	boxDouble.CrossingTopRightBottomChar = "╣"

	return map[string]*TableStyle{
		"default":     NewTableStyle(""),
		"borderless":  borderless,
		"compact":     compact,
		"style-guide": styleGuide,
		"box":         box,
		"box-double":  boxDouble,
	}
}

func RegisterTableStyle(name string, style *TableStyle) {
	defaultTableStyles[name] = style
}

func init() {
	defaultTableStyles = makeDefaultTableStyles()
}

func NewTableStyle(name string) *TableStyle {
	if name != "" {
		style, ok := defaultTableStyles[name]
		if ok {
			clone := *style
			return &clone
		}
	}

	return &TableStyle{
		PaddingChar:                 " ",
		HorizontalOutsideBorderChar: "-",
		HorizontalInsideBorderChar:  "-",
		VerticalOutsideBorderChar:   "|",
		VerticalInsideBorderChar:    "|",
		CrossingChar:                "+",
		CrossingTopRightChar:        "+",
		CrossingTopMidChar:          "+",
		CrossingTopLeftChar:         "+",
		CrossingMidRightChar:        "+",
		CrossingBottomRightChar:     "+",
		CrossingBottomMidChar:       "+",
		CrossingBottomLeftChar:      "+",
		CrossingMidLeftChar:         "+",
		CrossingTopLeftBottomChar:   "+",
		CrossingTopMidBottomChar:    "+",
		CrossingTopRightBottomChar:  "+",
		HeaderTitleFormat:           "<fg=black;bg=white;options=bold> %s </>",
		FooterTitleFormat:           "<fg=black;bg=white;options=bold> %s </>",
		CellHeaderFormat:            "<primary>%s</primary>",
		CellRowFormat:               "%s",
		CellRowContentFormat:        " %s ",
		BorderFormat:                "%s",
		PadType:                     "left",
	}
}

func (ts *TableStyle) Clone() *TableStyle {
	clone := *ts
	return &clone
}

type Table struct {
	headerTitle           string
	footerTitle           string
	headers               []string
	rows                  [][]*TableCell
	effectiveColumnWidths []int
	numberOfColumns       int
	style                 *TableStyle
	columnStyles          []*TableStyle
	columnWidths          []int
	columnMaxWidths       []int
	rendered              bool
	displayOrientation    string
	output                *ConsoleSectionOutput
}

func NewTable(o *Output) *Table {
	t := &Table{
		output:             NewConsoleSectionOutput(o, nil),
		displayOrientation: DisplayOrientationDefault,
	}

	t.SetStyleByName("default")

	return t
}

func NewTableSeparator() *TableCell {
	c := NewTableCell("")
	c.IsSeparator = true
	return c
}

func (t *Table) SetStyle(style *TableStyle) {
	t.style = style
}

func (t *Table) SetStyleByName(style string) {
	t.style = t.resolveStyle(style)
}

func (t *Table) Style() *TableStyle {
	return t.style
}

func (t *Table) SetColumnStyle(i uint, style *TableStyle) {
	if t.columnStyles == nil {
		t.columnStyles = make([]*TableStyle, i+1)
	}

	t.columnStyles[i] = style
}

func (t *Table) SetColumnStyleByName(i uint, name string) {
	t.SetColumnStyle(i, t.resolveStyle(name))
}

func (t *Table) ColumnStyle(columnIndex int) *TableStyle {
	if columnIndex < len(t.columnStyles) && t.columnStyles[columnIndex] != nil {
		return t.columnStyles[columnIndex]
	}

	return t.Style()
}

func (t *Table) SetColumnWidth(columnIndex int, width int) {
	if t.columnWidths == nil {
		t.columnWidths = make([]int, columnIndex+1)
	}

	t.columnWidths[columnIndex] = width
}

func (t *Table) SetColumnWidths(widths []int) {
	t.columnWidths = make([]int, len(widths))
	copy(t.columnWidths, widths)
}

func (t *Table) SetColumnMaxWidth(columnIndex int, maxWidth int) {
	if t.columnMaxWidths == nil {
		t.columnMaxWidths = make([]int, columnIndex+1)
	}

	t.columnMaxWidths[columnIndex] = maxWidth
}

func (t *Table) SetHeaders(headers []string) {
	t.headers = headers
}

func (t *Table) SetRows(rows [][]*TableCell) {
	t.rows = make([][]*TableCell, 0, len(rows))
	t.AddRows(rows)
}

func (t *Table) AddRows(rows [][]*TableCell) {
	for _, row := range rows {
		t.AddRow(row)
	}
}

func (t *Table) AddRow(rows []*TableCell) {
	t.rows = append(t.rows, rows)
}

func (t *Table) AppendRow(row []*TableCell) {
	if t.rendered {
		t.output.Clear(t.calculateRowCount())
	}

	t.AddRow(row)
	t.Render()
}

func (t *Table) SetRow(columnIndex int, row []*TableCell) {
	if t.rows == nil {
		t.rows = make([][]*TableCell, columnIndex+1)
	}

	t.rows[columnIndex] = row
}

func (t *Table) SetHeaderTitle(title string) {
	t.headerTitle = title
}

func (t *Table) SetFooterTitle(title string) {
	t.footerTitle = title
}

func (t *Table) SetHorizontal(horizontal bool) {
	if horizontal {
		t.displayOrientation = DisplayOrientationHorizontal
	} else {
		t.displayOrientation = DisplayOrientationDefault
	}
}

func (t *Table) SetVertical(vertical bool) {
	if vertical {
		t.displayOrientation = DisplayOrientationVertical
	} else {
		t.displayOrientation = DisplayOrientationDefault
	}
}

func rowIsTableSeparator(row []*TableCell) bool {
	return len(row) == 0 || (len(row) == 1 && row[0].IsSeparator)
}

func getEol(s string) string {
	eol := Eol
	if strings.Contains(s, "\r\n") {
		eol = "\r\n"
	}
	return eol
}

func (t *Table) Render() {
	divider := NewTableSeparator()
	horizontal := t.displayOrientation == DisplayOrientationHorizontal
	vertical := t.displayOrientation == DisplayOrientationVertical

	var rowLen int
	if horizontal {
		rowLen = max(len(t.headers), len(t.rows))
	}

	var rows [][]*TableCell
	if horizontal {
		rows = make([][]*TableCell, rowLen)

		for i, header := range t.headers {
			rows[i] = []*TableCell{NewTableCell(header)}
			for _, row := range t.rows {
				if rowIsTableSeparator(row) {
					continue
				}

				if i < len(row) && row[i] != nil {
					rows[i] = append(rows[i], row[i])
				} else if i < len(rows) && len(rows[i]) > 0 && rows[i][0] != nil {
					// noop
				} else {
					rows[i] = append(rows[i], nil)
				}
			}
		}
	} else if vertical {
		rows = make([][]*TableCell, 0, rowLen)
		formatter := t.output.Formatter()

		var maxHeaderLength int
		for _, header := range t.headers {
			maxHeaderLength = max(maxHeaderLength, helper.Width(formatter.RemoveDecoration(header)))
		}

		for _, row := range t.rows {
			if rowIsTableSeparator(row) {
				continue
			}

			if len(rows) > 0 {
				rows = append(rows, []*TableCell{divider})
			}

			containsColSpan := false
			for _, cell := range row {
				containsColSpan = cell.ColSpan >= 2
				if containsColSpan {
					break
				}
			}

			headers := t.headers
			maxRows := max(len(headers), len(row))

			for i := 0; i < maxRows; i++ {
				cell := row[i]

				var cellValue string
				if cell != nil {
					cellValue = cell.Value
				}

				eol := getEol(cellValue)
				parts := strings.Split(cellValue, eol)
				for idx, part := range parts {
					if len(headers) > 0 && !containsColSpan {
						if idx == 0 {
							header := ""
							if i < len(headers) {
								header = headers[i]
							}

							val := fmt.Sprintf("<comment>%s%s</>: %s", strings.Repeat(" ", maxHeaderLength-helper.Width(formatter.RemoveDecoration(header))), header, part)
							rows = append(rows, []*TableCell{NewTableCell(val)})
						} else {
							val := fmt.Sprintf("%s  %s", PadStart("", maxHeaderLength, " "), part)
							rows = append(rows, []*TableCell{NewTableCell(val)})
						}
					} else if cellValue != "" {
						rows = append(rows, []*TableCell{NewTableCell(part)})
					}
				}
			}
		}
	} else {
		rows = make([][]*TableCell, 0, rowLen)

		headers := make([]*TableCell, 0, len(t.headers))
		for _, header := range t.headers {
			headers = append(headers, NewTableCell(header))
		}
		rows = append(rows, headers)
		rows = append(rows, []*TableCell{divider})
		rows = append(rows, t.rows...)
	}

	t.calculateNumberOfColumns(rows)

	rowGroups := t.buildTableRows(rows)
	t.calculateColumnsWidth(rowGroups)

	isHeader := !horizontal
	isFirstRow := horizontal
	hasTitle := t.headerTitle != ""

	for rowGroup := range rowGroups {
		isHeaderSeparatorRendered := false

		for _, row := range rowGroup {
			if len(row) == 1 && row[0] == divider {
				isHeader = false
				isFirstRow = true
				continue
			}

			if rowIsTableSeparator(row) {
				t.renderRowSeparator(SeparatorMid, "", "")
				continue
			}

			if row == nil {
				continue
			}

			if isHeader && !isHeaderSeparatorRendered {
				if hasTitle {
					t.renderRowSeparator(SeparatorTop, t.headerTitle, t.style.HeaderTitleFormat)
				} else {
					t.renderRowSeparator(SeparatorTop, "", "")
				}

				hasTitle = false
				isHeaderSeparatorRendered = true
			}

			if isFirstRow {
				var separator int
				if horizontal {
					separator = SeparatorTop
				} else {
					separator = SeparatorTopBottom
				}

				if hasTitle {
					t.renderRowSeparator(separator, t.headerTitle, t.style.HeaderTitleFormat)
				} else {
					t.renderRowSeparator(separator, "", "")
				}

				isFirstRow = false
				hasTitle = false
			}

			if vertical {
				isHeader = false
				isFirstRow = false
			}

			if horizontal {
				t.renderRow(row, t.style.CellRowFormat, t.style.CellHeaderFormat)
			} else {
				if isHeader {
					t.renderRow(row, t.style.CellHeaderFormat, "")
				} else {
					t.renderRow(row, t.style.CellRowFormat, "")
				}
			}
		}
	}

	t.renderRowSeparator(SeparatorBottom, t.footerTitle, t.style.FooterTitleFormat)

	t.cleanup()
	t.rendered = true
}

func (t *Table) renderRowSeparator(separatorType int, title string, titleFormat string) {
	count := t.numberOfColumns
	if count == 0 {
		return
	}

	horizontalOutsideBorderChar := t.style.HorizontalOutsideBorderChar
	horizontalInsideBorderChar := t.style.HorizontalInsideBorderChar

	if horizontalOutsideBorderChar == "" && horizontalInsideBorderChar == "" && t.style.CrossingChar == "" {
		return
	}

	crossings := []string{
		t.style.CrossingChar,
		t.style.CrossingTopLeftChar,
		t.style.CrossingTopMidChar,
		t.style.CrossingTopRightChar,
		t.style.CrossingMidRightChar,
		t.style.CrossingBottomRightChar,
		t.style.CrossingBottomMidChar,
		t.style.CrossingBottomLeftChar,
		t.style.CrossingMidLeftChar,
		t.style.CrossingTopLeftBottomChar,
		t.style.CrossingTopMidBottomChar,
		t.style.CrossingTopRightBottomChar,
	}

	var horizontal string
	var leftChar string
	var midChar string
	var rightChar string

	if separatorType == SeparatorMid {
		horizontal = horizontalInsideBorderChar
		leftChar = crossings[8]
		midChar = crossings[0]
		rightChar = crossings[4]
	} else if separatorType == SeparatorTop {
		horizontal = horizontalOutsideBorderChar
		leftChar = crossings[1]
		midChar = crossings[2]
		rightChar = crossings[3]
	} else if separatorType == SeparatorTopBottom {
		horizontal = horizontalOutsideBorderChar
		leftChar = crossings[9]
		midChar = crossings[10]
		rightChar = crossings[11]
	} else {
		horizontal = horizontalOutsideBorderChar
		leftChar = crossings[7]
		midChar = crossings[6]
		rightChar = crossings[5]
	}

	markup := leftChar
	for column := 0; column < count; column++ {
		width := 0
		if column < len(t.effectiveColumnWidths) {
			width = t.effectiveColumnWidths[column]
		}

		markup += strings.Repeat(horizontal, width)

		if column == count-1 {
			markup += rightChar
		} else {
			markup += midChar
		}
	}

	if title != "" {
		formatter := t.output.Formatter()
		formattedTitle := fmt.Sprintf(titleFormat, title)
		titleLength := helper.Width(formatter.RemoveDecoration(formattedTitle))
		markupLength := helper.Width(markup)
		limit := markupLength - 4
		if titleLength > limit {
			titleLength = limit
			formatLength := helper.Width(formatter.RemoveDecoration(fmt.Sprintf(titleFormat, "")))
			formattedTitle = fmt.Sprintf(titleFormat, title[:limit-formatLength-3]) + "..."
		}

		titleStart := (markupLength - titleLength) / 2
		parts := []string{
			MbSubstr(markup, 0, titleStart),
			formattedTitle,
			MbSubstr(markup, titleStart+titleLength, markupLength),
		}
		markup = strings.Join(parts, "")
	}

	t.output.Writeln(fmt.Sprintf(t.style.BorderFormat, markup), 0)
}

func (t *Table) renderColumnSeparator(separatorType int) string {
	style := t.Style()

	if separatorType == 0 {
		return fmt.Sprintf(style.BorderFormat, style.VerticalOutsideBorderChar)
	}

	return fmt.Sprintf(style.BorderFormat, style.VerticalInsideBorderChar)
}

func (t *Table) renderRow(row []*TableCell, cellFormat string, firstCellFormat string) {
	var rowContent strings.Builder
	rowContent.WriteString(t.renderColumnSeparator(BorderOutside))
	columns := t.getRowColumns(row)
	last := columns[len(columns)-1]

	for i, column := range columns {
		if firstCellFormat != "" && i == 0 {
			rowContent.WriteString(t.renderCell(row, column, firstCellFormat))
		} else {
			rowContent.WriteString(t.renderCell(row, column, cellFormat))
		}

		if i == last {
			rowContent.WriteString(t.renderColumnSeparator(BorderOutside))
		} else {
			rowContent.WriteString(t.renderColumnSeparator(BorderInside))
		}
	}

	t.output.Writeln(rowContent.String(), 0)
}

func (t *Table) renderCell(row []*TableCell, column int, cellFormat string) string {
	var cell *TableCell
	if column < len(row) {
		cell = row[column]
	}

	width := 0
	if column < len(t.effectiveColumnWidths) {
		width = t.effectiveColumnWidths[column]
	}

	if cell.ColSpan > 1 {
		for nextColumn := column + 1; nextColumn < column+int(cell.ColSpan); nextColumn++ {
			width += t.getColumnSeparatorWidth()
			if nextColumn < len(t.effectiveColumnWidths) {
				width += t.effectiveColumnWidths[nextColumn]
			}
		}
	}

	style := t.ColumnStyle(column)

	if cell.IsSeparator {
		return fmt.Sprintf(style.BorderFormat, strings.Repeat(style.HorizontalInsideBorderChar, width))
	}

	content := fmt.Sprintf(style.CellRowContentFormat, cell.Value)

	padType := style.PadType
	if cell.Style != nil {
		re := regexp.MustCompile(`^<(\w+|(\w+=[\w,]+;?)*)>.+<\/(\w+|(\w+=\w+;?)*)?>$`)
		isNotStyledByTag := !re.MatchString(cell.Value)

		if isNotStyledByTag {
			cellFormat = cell.Style.CellFormat

			if strings.Contains(content, "</>") {
				content = strings.Replace(content, "</>", "", 1)
				width -= 3
			}

			styleTag := "<fg=default;bg=default>"
			if strings.Contains(content, styleTag) {
				content = strings.Replace(content, styleTag, "", 1)
				width -= len(styleTag)
			}
		}

		padType = cell.Style.Align
	}

	content = fmt.Sprintf(cellFormat, content)

	if padType == TableCellAlignCenter {
		content = PadCenter(content, width, style.PaddingChar)
	} else if padType == TableCellAlignRight {
		content = PadStart(content, width, style.PaddingChar)
	} else if padType == TableCellAlignLeft {
		content = PadEnd(content, width, style.PaddingChar)
	}

	return content
}

func (t *Table) calculateNumberOfColumns(rows [][]*TableCell) {
	c := 0
	for _, row := range rows {
		c = max(c, t.getNumberOfColumns(row))
	}
	t.numberOfColumns = c
}

func (t *Table) buildTableRows(rows [][]*TableCell) iter.Seq[[][]*TableCell] {
	formatter := t.output.Formatter()
	unmergedRows := make([][][]*TableCell, 0)

	for rowKey := 0; rowKey < len(rows); rowKey++ {
		rows := t.fillNextRows(rows, rowKey)

		// Remove any new line breaks and replace it with a new line
		for column, cell := range rows[rowKey] {
			colSpan := max(cell.ColSpan, 1)

			if column < len(t.columnMaxWidths) && helper.Width(formatter.RemoveDecoration(cell.Value)) > t.columnMaxWidths[column] {
				cell.Value = formatter.FormatAndWrap(cell.Value, t.columnMaxWidths[column]*int(colSpan))
			}

			if !strings.Contains(cell.Value, Eol) {
				continue
			}

			eol := getEol(cell.Value)
			parts := strings.Split(cell.Value, eol)
			for i, p := range parts {
				parts[i] = EscapeTrailingBackslash(p)
			}
			escaped := strings.Join(parts, eol)
			cell = NewTableCell(escaped)
			cell.ColSpan = colSpan

			lines := strings.Split(strings.ReplaceAll(cell.Value, eol, "<fg=default;bg=default></>"+eol), eol)
			for lineKey, line := range lines {
				lineCell := NewTableCell(line)
				lineCell.ColSpan = colSpan

				if lineKey == 0 {
					rows[rowKey][column] = lineCell
				} else {
					if len(unmergedRows) >= rowKey || len(unmergedRows[rowKey]) >= lineKey {
						unmergedRows = helper.Grow(unmergedRows, rowKey+1)
						unmergedRows[rowKey] = helper.Grow(unmergedRows[rowKey], lineKey+1)
						unmergedRows[rowKey][lineKey] = t.copyRow(rows, rowKey)
					}

					unmergedRows[rowKey][lineKey] = helper.Grow(unmergedRows[rowKey][lineKey], column+1)
					unmergedRows[rowKey][lineKey][column] = NewTableCell(line)
				}
			}
		}
	}

	return func(yield func([][]*TableCell) bool) {
		for rowKey, row := range rows {
			var rowGroup [][]*TableCell
			if rowIsTableSeparator(row) {
				rowGroup = [][]*TableCell{row}
			} else {
				rowGroup = [][]*TableCell{t.fillCells(row)}
			}

			if rowKey < len(unmergedRows) {
				for _, row := range unmergedRows[rowKey] {
					if rowIsTableSeparator(row) {
						rowGroup = append(rowGroup, row)
					} else {
						rowGroup = append(rowGroup, t.fillCells(row))
					}
				}
			}

			if !yield(rowGroup) {
				return
			}
		}
	}
}

func (t *Table) calculateRowCount() int {
	merged := make([][]*TableCell, 0, len(t.rows)+2)
	headers := make([]*TableCell, 0, len(t.headers))
	for _, h := range t.headers {
		headers = append(headers, NewTableCell(h))
	}
	merged = append(merged, headers)
	merged = append(merged, []*TableCell{NewTableSeparator()})
	merged = append(merged, t.rows...)

	tableRowsIter := t.buildTableRows(merged)
	tableRows := make([][][]*TableCell, 0)
	for rowGroup := range tableRowsIter {
		tableRows = append(tableRows, rowGroup)
	}

	numberOfRows := len(tableRows)

	if len(t.headers) > 0 {
		numberOfRows++ // Add row for header separator
	}

	if len(t.rows) > 0 {
		numberOfRows++ // Add row for footer separator
	}

	return numberOfRows
}

func (t *Table) fillNextRows(rows [][]*TableCell, line int) [][]*TableCell {
	unmergedRows := make([][]*TableCell, 0)

	for column, cell := range rows[line] {
		if cell.RowSpan > 1 {
			nbLines := cell.RowSpan - 1
			lines := []*TableCell{cell}

			if strings.Contains(cell.Value, Eol) {
				eol := getEol(cell.Value)

				lineParts := strings.Split(strings.ReplaceAll(cell.Value, eol, "<fg=default;bg=default>"+eol+"</>"), eol)
				newCell := NewTableCell(lineParts[0])
				newCell.ColSpan = cell.ColSpan
				newCell.Style = cell.Style
				rows[line][column] = newCell
				newLines := make([]*TableCell, 0, len(lines)-1)
				for _, part := range lineParts {
					newLines = append(newLines, NewTableCell(part))
				}
				lines = newLines
			}

			if line+1 < len(unmergedRows) {
				for i := line + 1; i < line+1+int(nbLines); i++ {
					if i < len(unmergedRows) {
						unmergedRows[i] = make([]*TableCell, 0)
					} else {
						unmergedRows = append(unmergedRows, make([]*TableCell, 0))
					}
				}
			}

			for unmergedRowKey := range unmergedRows {
				var value string
				if unmergedRowKey-line < len(lines) && lines[unmergedRowKey-line] != nil {
					value = lines[unmergedRowKey-line].Value
				}

				newCell := NewTableCell(value)
				newCell.ColSpan = cell.ColSpan
				newCell.Style = cell.Style
				unmergedRows[unmergedRowKey][column] = newCell

				if unmergedRowKey-line >= 0 && nbLines == uint(unmergedRowKey-line) {
					break
				}
			}
		}
	}

	for unmergedRowKey, unmergedRow := range unmergedRows {
		if unmergedRowKey < len(rows) && (t.getNumberOfColumns(rows[unmergedRowKey])+t.getNumberOfColumns(unmergedRow) <= t.numberOfColumns) {
			for cellKey, cell := range unmergedRow {
				rows[unmergedRowKey] = insertAt(rows[unmergedRowKey], cellKey, cell)
			}
		} else {
			row := t.copyRow(rows, unmergedRowKey-1)
			for column, cell := range unmergedRow {
				if cell != nil {
					row[column] = cell
				}
			}

			rows = insertAt(rows, unmergedRowKey, row)
		}
	}

	return rows
}

func insertAt[T any](s []T, offset int, replacement T) []T {
	if offset < 0 || offset >= len(s) {
		return append(s, replacement)
	}

	s = append(s, replacement)
	copy(s[offset+1:], s[offset:])
	s[offset] = replacement

	return s
}

func (t *Table) fillCells(row []*TableCell) []*TableCell {
	newRow := make([]*TableCell, 0, len(row))

	for column, cell := range row {
		newRow = append(newRow, cell)
		if cell.ColSpan > 1 {
			for i := column + 1; i < column+int(cell.ColSpan)-1; i++ {
				newRow = append(newRow, NewTableCell(""))
			}
		}
	}

	if len(newRow) == 0 {
		return row
	}

	return newRow
}

func (t *Table) copyRow(rows [][]*TableCell, line int) []*TableCell {
	row := rows[line]
	for cellKey, cellValue := range row {
		cell := NewTableCell("")
		cell.ColSpan = cellValue.ColSpan
		row[cellKey] = cell
	}
	return nil
}

func (t *Table) getNumberOfColumns(row []*TableCell) int {
	columns := len(row)
	for _, column := range row {
		columns += int(column.ColSpan) - 1
	}
	return columns
}

func (t *Table) getRowColumns(row []*TableCell) []int {
	columns := make([]int, 0, t.numberOfColumns-1)
	for i := 0; i < t.numberOfColumns; i++ {
		columns = append(columns, i)
	}

	for cellKey, cell := range row {
		if cell.ColSpan > 1 {
			start := cellKey + 1
			end := cellKey + int(cell.ColSpan)
			for i := start; i < end; i++ {
				for _, v := range columns {
					if v == i {
						columns = append(columns[:i], columns[i+1:]...)
						break
					}
				}
			}
		}
	}

	return columns
}

func (t *Table) calculateColumnsWidth(groups iter.Seq[[][]*TableCell]) {
	formatter := t.output.Formatter()

	for column := 0; column < t.numberOfColumns; column++ {
		lengths := make([]int, 0)

		for group := range groups {
			for _, row := range group {
				if rowIsTableSeparator(row) {
					continue
				}

				for i, cell := range row {
					textContent := formatter.RemoveDecoration(StripEscapeSequences(cell.Value))
					textLength := helper.Width(textContent)

					if len(textContent) > 0 {
						contentColumns := MbSplit(textContent, textLength/int(cell.ColSpan))

						for position, content := range contentColumns {
							idx := i + position

							if idx >= len(row) {
								for j := 0; j <= idx; j++ {
									row = append(row, nil)
								}
							}

							row[idx] = NewTableCell(content)
						}
					}
				}

				lengths = append(lengths, t.getCellWidth(row, column))
			}
		}

		t.effectiveColumnWidths = helper.Grow(t.effectiveColumnWidths, column+1)
		t.effectiveColumnWidths[column] = slices.Max(lengths) + helper.Width(t.style.CellRowContentFormat) - 2
	}
}

func (t *Table) getColumnSeparatorWidth() int {
	return helper.Width(fmt.Sprintf(t.style.BorderFormat, t.style.VerticalInsideBorderChar))
}

func (t *Table) getCellWidth(row []*TableCell, column int) int {
	cellWidth := 0

	if column < len(row) {
		cell := row[column]
		cellWidth = helper.Width(t.output.Formatter().RemoveDecoration(StripEscapeSequences(cell.Value)))
	}

	columnWidth := 0
	if column < len(t.columnWidths) {
		columnWidth = t.columnWidths[column]
	}

	cellWidth = max(cellWidth, columnWidth)

	if column < len(t.columnMaxWidths) {
		return min(t.columnMaxWidths[column], cellWidth)
	}

	return cellWidth
}

func (t *Table) cleanup() {
	t.effectiveColumnWidths = make([]int, 0)
	t.numberOfColumns = -1
}

func (t *Table) resolveStyle(name string) *TableStyle {
	style, ok := defaultTableStyles[name]
	if !ok {
		log.Fatalf("unknown table style: \"%s\"", name)
	}

	return style
}
