package prompt

import (
	"github.com/c-bata/go-prompt/internal/debug"
	runewidth "github.com/mattn/go-runewidth"
)

const (
	scrollBarWidth  = 1
	statusBarHeight = 1
)

// Render to render prompt information from state of Buffer.
type Render struct {
	out                ConsoleWriter
	prefix             string
	livePrefixCallback func(doc *Document, breakline bool) (prefix string, useLivePrefix bool)
	breakLineCallback  func(doc *Document)
	title              string
	row                uint16
	col                uint16
	stringCaches       bool
	statusBar          StatusBar
	allocatedLines     int
	enableMarkup       bool

	previousCursor int

	// colors,
	prefixTextColor              Color
	prefixBGColor                Color
	inputTextColor               Color
	inputBGColor                 Color
	previewSuggestionTextColor   Color
	previewSuggestionBGColor     Color
	previewNextTextColor         Color
	previewNextBGColor           Color
	suggestionTextColor          Color
	suggestionBGColor            Color
	selectedSuggestionTextColor  Color
	selectedSuggestionBGColor    Color
	descriptionTextColor         Color
	descriptionBGColor           Color
	selectedDescriptionTextColor Color
	selectedDescriptionBGColor   Color
	scrollbarThumbColor          Color
	scrollbarBGColor             Color
	statusBarTextColor           Color
	statusBarBGColor             Color
	suggestTypeLabelTextColor    Color
	suggestTypeLabelBGColor      Color
}

// Setup to initialize console output.
func (r *Render) Setup() {
	if r.title != "" {
		r.out.SetTitle(r.title)
		debug.AssertNoError(r.out.Flush())
	}
}

// TearDown to clear title and erasing.
func (r *Render) TearDown() {
	r.out.ClearTitle()
	r.out.EraseDown()
	debug.AssertNoError(r.out.Flush())
}

// Render renders to the console.
func (r *Render) Render(buffer *Buffer, completion *CompletionManager, lexer *Lexer) {
	// In situations where a pseudo tty is allocated (e.g. within a docker container),
	// window size via TIOCGWINSZ is not immediately available and will result in 0,0 dimensions.
	if r.col == 0 {
		return
	}

	defer func() { debug.AssertNoError(r.out.Flush()) }()

	line := buffer.Text()
	lw := runewidth.StringWidth(line)
	rows := lw / int(r.col)
	r.allocateArea(rows + statusBarHeight)
	r.eraseArea(rows)
	r.move(r.previousCursor, 0)

	// TODO: it feels like the above should account for the prefix width; investigate
	prefix := r.getCurrentPrefix(buffer, false)
	pw := runewidth.StringWidth(prefix)
	cursor := pw + lw

	// Ensure area size is large enough to work with
	_, y := r.toPos(cursor)
	h := y + int(completion.max) + statusBarHeight
	if h > int(r.row) || completionMargin > int(r.col) {
		r.renderWindowTooSmall()
		return
	}

	// Rendering
	r.out.HideCursor()
	defer r.out.ShowCursor()

	// TODO: reset to start of line (without messing up cursor tracking) to allow
	//       for previous cursor to be overwritten if the prompt is restarted
	r.renderPrefix(buffer, false)
	r.eraseArea(0)

	// Render lexed input line if lexing is enabled
	if lexer.IsEnabled {
		r.renderLexable(line, lexer)
	} else {
		r.out.SetColor(r.inputTextColor, r.inputBGColor, false)
		r.out.WriteStr(line)
	}

	if y == 0 {
		r.out.EraseEndOfLine()
	}

	// Prepare for render
	r.out.SetColor(DefaultColor, DefaultColor, false)
	r.lineWrap(cursor)
	cursor = r.backward(cursor, runewidth.StringWidth(line)-buffer.DisplayCursorPosition())

	// Render completion components
	r.renderCompletion(buffer, completion)
	r.previousCursor = r.renderPreview(buffer, completion, lexer, cursor)
}

// UpdateWinSize called when window size is changed.
func (r *Render) UpdateWinSize(ws *WinSize) {
	r.row = ws.Row
	r.col = ws.Col
}

// BreakLine to break line.
func (r *Render) BreakLine(buffer *Buffer, lexer *Lexer) {
	d := buffer.Document()
	line := d.Text + "\n"

	// Erasing and Render
	prefix := r.getCurrentPrefix(buffer, true)
	cursor := runewidth.StringWidth(d.TextBeforeCursor()) + runewidth.StringWidth(prefix)
	r.clear(cursor)

	r.renderPrefix(buffer, true)

	if lexer.IsEnabled {
		r.renderLexable(line, lexer)
	} else {
		r.out.SetColor(r.inputTextColor, r.inputBGColor, false)
		r.out.WriteStr(line)
	}

	r.out.SetColor(DefaultColor, DefaultColor, false)

	debug.AssertNoError(r.out.Flush())
	if r.breakLineCallback != nil {
		r.breakLineCallback(d)
	}

	r.previousCursor = 0
}

func (r *Render) allocateArea(lines int) {
	for i := 0; i < lines; i++ {
		r.out.ScrollDown()
	}
	for i := 0; i < lines; i++ {
		r.out.ScrollUp()
	}
	if r.allocatedLines < lines {
		r.allocatedLines = lines
	}
}

func (r *Render) eraseArea(skip int) {
	r.out.SaveCursor()
	defer r.out.UnSaveCursor()
	lines := r.allocatedLines
	for i := 0; i < lines; i++ {
		r.out.CursorDown(1)
		if i >= skip {
			r.out.EraseLine()
		}
	}
}

// clear erases the screen from a beginning of input
// even if there is line break which means input length exceeds a window's width.
func (r *Render) clear(cursor int) {
	r.move(cursor, 0)
	r.out.EraseDown()
}

// backward moves cursor to backward from a current cursor position
// regardless there is a line break.
func (r *Render) backward(from, n int) int {
	return r.move(from, from-n)
}

// move moves cursor to specified position from the beginning of input
// even if there is a line break.
func (r *Render) move(from, to int) int {
	fromX, fromY := r.toPos(from)
	toX, toY := r.toPos(to)

	r.out.CursorUp(fromY - toY)
	r.out.CursorBackward(fromX - toX)
	return to
}

// toPos returns the relative position from the beginning of the string.
func (r *Render) toPos(cursor int) (x, y int) {
	col := int(r.col)
	return cursor % col, cursor / col
}

func (r *Render) lineWrap(cursor int) {
	if cursor > 0 && cursor%int(r.col) == 0 {
		r.out.WriteRaw([]byte{'\n'})
	}
}

func clamp(high, low, x float64) float64 {
	switch {
	case high < x:
		return high
	case x < low:
		return low
	default:
		return x
	}
}
