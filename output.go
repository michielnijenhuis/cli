package cli

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/michielnijenhuis/cli/helper"
	"github.com/michielnijenhuis/cli/terminal"
)

const maxLineLength = 120

type Output struct {
	Stream         *os.File
	Stderr         *Output
	verbosity      uint
	decorated      bool
	formatter      *OutputFormatter
	lineLength     int
	bufferedOutput *TrimmedBufferOutput
	input          *Input
	Logger
	// progressBar    *ProgressBar
}

const (
	VerbosityQuiet       uint = 16
	VerbosityNormal      uint = 32
	VerbosityVerbose     uint = 64
	VerbosityVeryVerbose uint = 128
	VerbosityDebug       uint = 256
)

const (
	OutputNormal uint = 1
	OutputRaw    uint = 2
	OutputPlain  uint = 4
)

func setupNewOutput(input *Input, stream *os.File, formatter *OutputFormatter) *Output {
	o := &Output{
		Stream:     stream,
		verbosity:  VerbosityNormal,
		decorated:  HasColorSupport(),
		formatter:  formatter,
		lineLength: maxLineLength,
		input:      input,
		bufferedOutput: &TrimmedBufferOutput{
			Output: &Output{
				Stream:     stream,
				verbosity:  VerbosityNormal,
				decorated:  HasColorSupport(),
				formatter:  formatter,
				lineLength: maxLineLength,
				input:      input,
			},
		},
	}

	o.Logger = NewLogger(o, -1)

	if formatter != nil {
		formatter.Decorated = o.decorated
	}

	width := terminal.Columns()
	if width == 0 {
		width = maxLineLength
	}

	o.lineLength = min(width, maxLineLength)

	return o
}

func NewOutput(input *Input) *Output {
	f := &OutputFormatter{}
	o := setupNewOutput(input, os.Stdout, f)
	o.Stderr = setupNewOutput(input, os.Stderr, f)
	return o
}

func (o *Output) Formatter() *OutputFormatter {
	return o.formatter
}

func (o *Output) IsDecorated() bool {
	return o.Formatter().Decorated
}

func (o *Output) Verbosity() uint {
	return o.verbosity
}

func (o *Output) IsQuiet() bool {
	return o.verbosity == VerbosityQuiet
}

func (o *Output) IsVerbose() bool {
	return o.verbosity == VerbosityVerbose
}

func (o *Output) IsVeryVerbose() bool {
	return o.verbosity == VerbosityVeryVerbose
}

func (o *Output) IsDebug() bool {
	return o.verbosity == VerbosityDebug
}

func (o *Output) Writeln(s string, options uint) {
	o.Writelns([]string{s}, options)
}

func (o *Output) Writelnf(f string, options uint, args ...any) {
	o.Writelns([]string{fmt.Sprintf(f, args...)}, options)
}

func (o *Output) Writelns(s []string, options uint) {
	o.WriteMany(s, true, options)
}

func (o *Output) Write(message string, newLine bool, options uint) {
	o.WriteMany([]string{message}, newLine, options)
}

func (o *Output) Format(message string) string {
	return o.Formatter().Format(message)
}

func (o *Output) Formatf(format string, args ...any) string {
	return o.Formatter().Format(fmt.Sprintf(format, args...))
}

func (o *Output) WriteMany(messages []string, newLine bool, options uint) {
	types := OutputNormal | OutputRaw | OutputPlain

	t := types & options
	if t == 0 {
		t = OutputNormal
	}

	verbosities := VerbosityQuiet | VerbosityNormal | VerbosityVerbose | VerbosityVeryVerbose | VerbosityDebug
	verbosity := verbosities & options
	if verbosity == 0 {
		verbosity = VerbosityNormal
	}

	if verbosity > o.Verbosity() {
		return
	}

	re := regexp.MustCompile("<[^>]+>")

	var message string
	for _, m := range messages {
		message = m
		switch t {
		case OutputNormal:
			message = o.Formatter().Format(message)
		case OutputPlain:
			message = re.ReplaceAllString(o.Formatter().Format(message), "")
		}

		o.DoWrite(message, newLine)
	}
}

func (o *Output) DoWrite(message string, newLine bool) error {
	if newLine {
		message += Eol
	}

	_, err := o.Stream.WriteString(message)
	return err
}

func (o *Output) SetDecorated(decorated bool) {
	doSetDecorated(o, decorated)
	doSetDecorated(o.Stderr, decorated)

	if o.Formatter() != nil {
		o.Formatter().Decorated = decorated
	}
}

func doSetDecorated(o *Output, decorated bool) {
	if o != nil {
		o.decorated = decorated
	}
}

func (o *Output) SetFormatter(formatter *OutputFormatter) {
	doSetFormatter(o, formatter)
	doSetFormatter(o.Stderr, formatter)
}

func doSetFormatter(o *Output, formatter *OutputFormatter) {
	if o != nil {
		o.formatter = formatter
	}
}

func (o *Output) SetVerbosity(verbose uint) {
	doSetVerbosity(o, verbose)
	doSetVerbosity(o.Stderr, verbose)
}

func doSetVerbosity(o *Output, verbose uint) {
	if o != nil {
		o.verbosity = verbose
	}

	if o.Logger != nil {
		var logLevel int
		switch verbose {
		case VerbosityQuiet:
			logLevel = LogLevelFatal
		case VerbosityNormal:
			logLevel = LogLevelError
		case VerbosityVerbose:
			logLevel = LogLevelWarn
		case VerbosityVeryVerbose:
			logLevel = LogLevelInfo
		case VerbosityDebug:
			logLevel = LogLevelDebug
		default:
			logLevel = LogLevelError
		}

		o.Logger.SetLogLevel(logLevel)
	}
}

func (o *Output) Title(message string) {
	o.autoPrependBlock()
	messages := []string{
		fmt.Sprintf("<comment>%s</>", EscapeTrailingBackslash(message)),
		fmt.Sprintf("<comment>%s</>", strings.Repeat("=", helper.Width(o.Formatter().RemoveDecoration(message)))),
	}
	o.Writelns(messages, 0)
	o.NewLine(1)
}

func (o *Output) Section(message string) {
	o.autoPrependBlock()
	messages := []string{
		fmt.Sprintf("<comment>%s</>", EscapeTrailingBackslash(message)),
		fmt.Sprintf("<comment>%s</>", strings.Repeat("-", helper.Width(o.Formatter().RemoveDecoration(message)))),
	}
	o.Writelns(messages, 0)
	o.NewLine(1)
}

func (o *Output) Listing(elements []string, prefix string) {
	o.autoPrependBlock()
	els := make([]string, len(elements))
	for i, element := range elements {
		els[i] = fmt.Sprintf(" %s %s", prefix, element)
	}
	o.Writelns(els, 0)
	o.NewLine(1)
}

func (o *Output) Block(messages []string, tag string, escape bool) {
	theme, _ := GetTheme(tag)

	if theme.Padding {
		o.autoPrependBlock()
	}

	o.Writelns(o.CreateBlock(messages, tag, theme, escape), 0)

	if theme.Padding {
		o.NewLine(1)
	}
}

func (o *Output) CreateBlock(messages []string, tag string, theme *Theme, escape bool) []string {
	prefix := ""
	if theme != nil && theme.Icon != "" {
		prefix = fmt.Sprintf("%s ", theme.Icon)
	}
	if theme != nil && theme.Label != "" {
		prefix += theme.Label
	}

	indentLength := 0
	prefixLength := helper.Width(o.Formatter().RemoveDecoration(prefix))
	lines := make([]string, 0, len(messages))
	lineIndentation := ""

	if prefix != "" {
		indentLength = helper.Width(prefix)
		lineIndentation = strings.Repeat(" ", indentLength)
	}

	for i, message := range messages {
		if escape {
			message = Escape(message)
		}

		wrapped := helper.Wrap(message, o.lineLength-prefixLength-indentLength, Eol, false)
		lines = append(lines, strings.Split(wrapped, Eol)...)

		if len(messages) > 1 && i < len(messages)-1 {
			lines = append(lines, "")
		}
	}

	firstLineIndex := 0
	if theme != nil && theme.Padding && o.IsDecorated() {
		firstLineIndex = 1
		helper.Unshift(&lines, "")
		lines = append(lines, "")
	}

	for i, line := range lines {
		if prefix != "" {
			if firstLineIndex == i {
				if tag != "" && theme != nil && !theme.FullyColored {
					prefix = fmt.Sprintf("<%s>%s</>", tag, prefix)
				}

				line = prefix + line
			} else {
				line = lineIndentation + line
			}
		}

		line += strings.Repeat(" ", max(o.lineLength-helper.Width(o.Formatter().RemoveDecoration(line)), 0))

		if tag != "" && theme.FullyColored {
			line = fmt.Sprintf("<%s>%s</>", tag, line)
		}

		lines[i] = line
	}

	return lines
}

func (o *Output) Text(messages ...string) {
	o.autoPrependBlock()
	for _, m := range messages {
		o.Writeln(" "+m, 0)
	}
}

func (o *Output) Ok(messages ...string) {
	o.Block(messages, "ok", true)
}

func (o *Output) Success(messages ...string) {
	o.Block(messages, "success", true)
}

func (o *Output) Error(messages ...string) {
	o.Block(messages, "error", true)
}

func (o *Output) Err(err error) {
	o.Error(err.Error())
}

func (o *Output) Warning(messages ...string) {
	o.Block(messages, "warning", true)
}

func (o *Output) Info(messages ...string) {
	o.Block(messages, "info", true)
}

func (o *Output) Note(messages ...string) {
	o.Block(messages, "note", false)
}

func (o *Output) Caution(messages ...string) {
	o.Block(messages, "caution", true)
}

func (o *Output) Comment(messages ...string) {
	o.Block(messages, "comment", false)
}

func (o *Output) TableFromSlices(headers []string, rows [][]any, options *TableOptions) {
	o.Table(headers, o.CreateTableRowsFromSlices(rows), options)
}

func (o *Output) TableFromMap(rows []map[string]any, headers []string, options *TableOptions) {
	tableRows := o.CreateTableRowsFromMaps(headers, rows)

	if headers == nil && len(rows) > 0 {
		headers = make([]string, 0, len(rows[0]))

		for k := range rows[0] {
			headers = append(headers, k)
		}
	}

	o.Table(headers, tableRows, options)
}

type TableOptions struct {
	DisplayOrientation string
	HeaderTitle        string
	FooterTitle        string
	Style              string
	Align              string
}

func (o *Output) Table(headers []string, rows [][]*TableCell, options *TableOptions) {
	t := o.CreateTable(headers, rows, options)
	t.Render()
	o.NewLine(1)
}

func (o *Output) CreateTableRowsFromMaps(headers []string, rows []map[string]any) [][]*TableCell {
	tableRows := make([][]*TableCell, 0, len(rows))
	for _, row := range rows {
		cells := make([]*TableCell, 0, len(row))
		for _, h := range headers {
			cells = append(cells, NewTableCell(fmt.Sprintf("%v", row[h])))
		}
		tableRows = append(tableRows, cells)
	}

	return tableRows
}

func (o *Output) CreateTableRowsFromSlices(rows [][]any) [][]*TableCell {
	tableRows, ok := any(rows).([][]*TableCell)
	if ok {
		return tableRows
	}

	tableRows = make([][]*TableCell, 0, len(rows))
	for _, row := range rows {
		cells, ok := any(row).([]*TableCell)
		if ok {
			tableRows = append(tableRows, cells)
			continue
		}

		cells = make([]*TableCell, 0, len(row))
		for _, v := range row {
			cells = append(cells, NewTableCell(fmt.Sprintf("%v", v)))
		}
		tableRows = append(tableRows, cells)
	}

	return tableRows
}

func (o *Output) CreateTable(headers []string, rows [][]*TableCell, options *TableOptions) *Table {
	t := NewTable(o)

	t.headers = headers
	t.rows = rows

	styleName := "box"
	if options != nil && options.Style != "" {
		styleName = options.Style
	}

	style := NewTableStyle(styleName)
	t.style = style

	if options != nil {
		if options.Align != "" {
			style.PadType = options.Align
		}

		if options.DisplayOrientation != "" {
			t.displayOrientation = options.DisplayOrientation
		}

		if options.HeaderTitle != "" {
			t.headerTitle = options.HeaderTitle
		}

		if options.FooterTitle != "" {
			t.footerTitle = options.FooterTitle
		}
	}

	return t
}

func (o *Output) Ask(question string, defaultValue string) (string, error) {
	textPrompt := NewTextPrompt(o.input, o, question, defaultValue)
	return textPrompt.Render()
}

func (o *Output) Confirm(q string, defaultValue bool) (bool, error) {
	confirmPrompt := NewConfirmPrompt(o.input, o, q, defaultValue)
	return confirmPrompt.Render()
}

func (o *Output) Choice(question string, choices map[string]string, defaultValue string) (string, error) {
	labels := make([]string, 0, len(choices))
	values := make([]string, 0, len(choices))
	for k, v := range choices {
		labels = append(labels, v)
		values = append(values, k)
	}
	choicePrompt := NewSelectPrompt(o.input, o, question, values, labels, defaultValue)
	return choicePrompt.Render()
}

func (o *Output) NewLine(count int) {
	for count > 0 {
		o.Writeln("", 0)
		count--
	}
}

// TODO: impl
func (o *Output) ProgressStart(max uint) {}

// TODO: impl
func (o *Output) ProgressAdvance(step uint) {}

// TODO: impl
func (o *Output) ProgressFinish() {}

func (o *Output) Box(title string, body string, footer string, color string, info string) {
	o.Writeln(Box(title, body, footer, color, info), 0)
}

func (o *Output) autoPrependBlock() {
	chars := o.bufferedOutput.Fetch()

	if len(chars) > 2 {
		chars = chars[:len(chars)-2]
	}

	if chars == "" {
		o.NewLine(1)
		return
	}

	var lineBreakCount int
	for i := 0; i < len(chars); i++ {
		char := chars[i]
		if char == '\n' {
			lineBreakCount++
		}
	}

	o.NewLine(2 - lineBreakCount)
}

func (o *Output) Search(label string, options func(string) SearchResult, placeholder string) (string, error) {
	p := NewSearchPrompt(o.input, o, label, options, placeholder)
	return p.Render()
}

func (o *Output) MultiSelect(label string, options MultiSelectOptions, defaultValues []string) ([]string, error) {
	p := NewMultiSelectPrompt(o.input, o, label, options, defaultValues)
	return p.Render()
}
