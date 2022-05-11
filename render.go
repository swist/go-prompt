package prompt

import (
	"bytes"
	"math"
	"os"
	"runtime"
	"strings"
	"unicode"

	"github.com/c-bata/go-prompt/internal/debug"
	runewidth "github.com/mattn/go-runewidth"
)

const (
	scrollBarWidth  = 1
	statusBarHeight = 1
)

// Render to render prompt information from state of Buffer.
type Render struct {
	out                 ConsoleWriter
	prefix              string
	livePrefixCallback  func(doc *Document, breakline bool) (prefix string, useLivePrefix bool)
	breakLineCallback   func(doc *Document)
	title               string
	row                 uint16
	col                 uint16
	stringCaches        bool
	statusBar           StatusBar
	renderPrefixAtStart bool

	previousCursor int

	// colors,
	prefixTextColor              Color
	prefixBGColor                Color
	inputTextColor               Color
	inputBGColor                 Color
	previewSuggestionTextColor   Color
	previewSuggestionBGColor     Color
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

// getCurrentPrefix to get current prefix.
// If live-prefix is enabled, return live-prefix.
func (r *Render) getCurrentPrefix(buffer *Buffer, breakline bool) string {
	if prefix, ok := r.livePrefixCallback(buffer.Document(), breakline); ok {
		return prefix
	}
	return r.prefix
}

func (r *Render) renderPrefix(buffer *Buffer, breakline bool) {
	r.out.SetColor(r.prefixTextColor, r.prefixBGColor, false)
	r.out.WriteStr(r.getCurrentPrefix(buffer, breakline))
	r.out.SetColor(DefaultColor, DefaultColor, false)
	r.out.Flush()
}

// TearDown to clear title and erasing.
func (r *Render) TearDown() {
	r.out.ClearTitle()
	r.out.EraseDown()
	debug.AssertNoError(r.out.Flush())
}

func (r *Render) prepareArea(lines int) {
	for i := 0; i < lines; i++ {
		r.out.ScrollDown()
	}
	for i := 0; i < lines; i++ {
		r.out.ScrollUp()
	}
}

// UpdateWinSize called when window size is changed.
func (r *Render) UpdateWinSize(ws *WinSize) {
	r.row = ws.Row
	r.col = ws.Col
}

func (r *Render) renderWindowTooSmall() {
	r.out.CursorGoTo(0, 0)
	r.out.EraseScreen()
	r.out.SetColor(DarkRed, White, false)
	r.out.WriteStr("Your console window is too small...")
}

func (r *Render) renderCompletion(buf *Buffer, completions *CompletionManager) {
	suggestions := completions.GetSuggestions()
	if len(suggestions) == 0 {
		return
	}
	prefix := r.getCurrentPrefix(buf, false)
	maxWidth := int(r.col) - runewidth.StringWidth(prefix) - scrollBarWidth
	// TODO: should completion height be returned by formatSuggestions?
	completionHeight := len(suggestions)
	if completionHeight > int(completions.max) {
		completionHeight = int(completions.max)
	}
	formatted, width, leftWidth, rightWidth := formatSuggestions(suggestions, maxWidth, completions.verticalScroll, completionHeight, r.stringCaches)
	width += scrollBarWidth

	showingLast := completions.verticalScroll+completionHeight == len(suggestions)
	selected := completions.selected - completions.verticalScroll
	if selected >= 0 && completions.expandDescriptions {
		selectedSuggest := suggestions[completions.selected]
		formatted = r.expandDescription(formatted, selectedSuggest.Description, int(completions.max), rightWidth, leftWidth)
		completionHeight = len(formatted)
	}

	r.prepareArea(completionHeight + statusBarHeight)

	cursor := runewidth.StringWidth(prefix) + runewidth.StringWidth(buf.Document().TextBeforeCursor())
	x, _ := r.toPos(cursor)
	if x+width >= int(r.col) {
		cursor = r.backward(cursor, x+width-int(r.col))
	}

	contentHeight := len(completions.tmp)

	fractionVisible := float64(completionHeight) / float64(contentHeight)
	fractionAbove := float64(completions.verticalScroll) / float64(contentHeight)

	scrollbarHeight := int(clamp(float64(completionHeight), 1, float64(completionHeight)*fractionVisible))
	scrollbarTopFloat := float64(completionHeight) * fractionAbove
	var scrollbarTop int
	if showingLast {
		scrollbarTop = int(math.Ceil(scrollbarTopFloat))
	} else {
		scrollbarTop = int(math.Floor(scrollbarTopFloat))
	}

	isScrollThumb := func(row int) bool {
		return scrollbarTop <= row && row < scrollbarTop+scrollbarHeight
	}

	r.out.SetColor(White, Cyan, false)
	for i := 0; i < completionHeight; i++ {
		r.out.CursorDown(1)
		if i == selected {
			r.out.SetColor(r.selectedSuggestionTextColor, r.selectedSuggestionBGColor, true)
		} else if formatted[i].Type != SuggestTypeDefault {
			textColor, bgColor := r.getSuggestionTypeColor(formatted[i].Type)
			r.out.SetColor(textColor, bgColor, false)
		} else {
			r.out.SetColor(r.suggestionTextColor, r.suggestionBGColor, false)
		}
		r.out.WriteStr(formatted[i].Text)

		if i == selected && !completions.expandDescriptions {
			r.out.SetColor(r.selectedDescriptionTextColor, r.selectedDescriptionBGColor, false)
		} else if formatted[i].Type != SuggestTypeDefault && (!completions.expandDescriptions || selected < 0) {
			textColor, bgColor := r.getSuggestionTypeColor(formatted[i].Type)
			r.out.SetColor(textColor, bgColor, false)
		} else {
			r.out.SetColor(r.descriptionTextColor, r.descriptionBGColor, false)
		}
		r.out.WriteStr(formatted[i].Description)

		if isScrollThumb(i) {
			r.out.SetColor(DefaultColor, r.scrollbarThumbColor, false)
		} else {
			r.out.SetColor(DefaultColor, r.scrollbarBGColor, false)
		}
		r.out.WriteStr(" ")
		r.out.SetColor(DefaultColor, DefaultColor, false)

		r.lineWrap(cursor + width)
		r.backward(cursor+width, width)
	}

	if x+width >= int(r.col) {
		r.out.CursorForward(x + width - int(r.col))
	}

	r.out.CursorUp(completionHeight)
	r.out.SetColor(DefaultColor, DefaultColor, false)
}

func (r *Render) getSuggestionTypeColor(typ SuggestType) (textColor, bgColor Color) {
	switch typ {
	case SuggestTypeLabel:
		return r.suggestTypeLabelTextColor, r.suggestTypeLabelBGColor
	default:
		return r.suggestionTextColor, r.suggestionBGColor
	}
}

func (r *Render) expandDescription(formatted []Suggest, expand string, mh int, mw int, leftWidth int) []Suggest {
	if mh <= 0 || mw <= 2 {
		return formatted
	}

	mw = mw - 2
	reformatted := make([]Suggest, 0, len(formatted))
	wrapped := strings.Split(wrap(expand, uint(mw)), "\n")
	lf := len(formatted)
	lw := len(wrapped)
	l := lf
	if lw > lf {
		l = lw
	}
	if l > mh {
		l = mh
	}
	for i := 0; i < l; i++ {
		var desc string
		if i < lw {
			if i == l-1 && lw > l {
				desc = "..."
			} else {
				desc = wrapped[i]
				if len(desc) > mw {
					desc = desc[:mw-3] + "..."
				}
			}
			w := mw - len(desc)
			if w < 0 {
				w = 0
			}
			pad := strings.Repeat(" ", w)
			desc = " " + desc + pad + " "
		} else {
			desc = strings.Repeat(" ", mw+2)
		}
		text := strings.Repeat(" ", leftWidth)
		var typ SuggestType = SuggestTypeDefault
		if i < lf {
			text = formatted[i].Text
			typ = formatted[i].Type
		}
		reformatted = append(reformatted, Suggest{
			Text:        text,
			Description: desc,
			Type:        typ,
		})
	}

	return reformatted
}

func (r *Render) renderSelection(buffer *Buffer, completion *CompletionManager, lexer *Lexer, cursor int) int {
	if suggest, ok := completion.GetSelectedSuggestion(); ok {
		// r.out.EraseEndOfLine()
		cursor = r.backward(cursor, runewidth.StringWidth(buffer.Document().GetWordBeforeCursorUntilSeparator(completion.wordSeparator)))

		r.out.SetColor(r.previewSuggestionTextColor, r.previewSuggestionBGColor, false)
		r.out.WriteStr(suggest.Text)
		r.out.SetColor(DefaultColor, DefaultColor, false)
		cursor += runewidth.StringWidth(suggest.Text)

		rest := buffer.Document().TextAfterCursor()

		if lexer.IsEnabled {
			r.renderLexed(rest, lexer)
		} else {
			r.out.WriteStr(rest)
		}

		r.out.SetColor(DefaultColor, DefaultColor, false)

		cursor += runewidth.StringWidth(rest)
		r.lineWrap(cursor)

		cursor = r.backward(cursor, runewidth.StringWidth(rest))
	}
	return cursor
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
	r.prepareArea((lw / int(r.col)) + statusBarHeight)
	r.move(r.previousCursor, 0)

	prefix := r.getCurrentPrefix(buffer, false)
	pw := runewidth.StringWidth(prefix)
	cursor := lw
	if r.renderPrefixAtStart {
		cursor += pw
	}

	// Ensure area size is large enough to work with
	_, y := r.toPos(cursor)
	h := y + 1 + int(completion.max)
	if h > int(r.row) || completionMargin > int(r.col) {
		r.renderWindowTooSmall()
		return
	}

	// Rendering
	r.out.HideCursor()
	defer r.out.ShowCursor()

	// Whether or not to emit the prefix -- typically disabled when the prompt is restarted
	// in the middle of an existing session
	if r.renderPrefixAtStart {
		r.renderPrefix(buffer, false)
	} else {
		r.renderPrefixAtStart = true
	}

	// Render lexed input line if lexing is enabled
	if lexer.IsEnabled {
		r.renderLexed(line, lexer)
	} else {
		r.out.SetColor(r.inputTextColor, r.inputBGColor, false)
		r.out.WriteStr(line)
	}

	// Prepare for render
	r.out.SetColor(DefaultColor, DefaultColor, false)
	r.lineWrap(cursor)

	r.out.EraseDown()

	cursor = r.backward(cursor, runewidth.StringWidth(line)-buffer.DisplayCursorPosition())

	// Render components
	r.renderCompletion(buffer, completion)
	r.renderStatusBar()
	r.previousCursor = r.renderSelection(buffer, completion, lexer, cursor)
}

func (r *Render) renderStatusBar() {
	if r.statusBar == nil {
		return
	}

	r.out.SaveCursor()
	r.out.HideCursor()
	defer func() {
		r.out.UnSaveCursor()
		r.out.CursorUp(0)
		r.out.ShowCursor()
	}()

	if r.statusBar != nil {
		r.out.CursorDown(int(r.row))
		r.out.CursorBackward(int(r.col))
		for _, el := range r.statusBar.fit(int(r.col)) {
			r.renderStatusElement(el)
		}
		r.out.SetColor(DefaultColor, DefaultColor, false)
	}
	r.out.SetColor(DefaultColor, DefaultColor, false)
}

func (r *Render) renderStatusElement(el StatusElement) {
	fgColor := r.statusBarTextColor
	if el.TextColor != nil {
		fgColor = *el.TextColor
	}
	bgColor := r.statusBarBGColor
	if el.BGColor != nil {
		bgColor = *el.BGColor
	}
	r.out.SetColor(fgColor, bgColor, el.Bold)
	r.out.EraseEndOfLine()
	r.out.WriteStr(el.Text)
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
		r.renderLexed(line, lexer)
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

func (r *Render) renderLexed(line string, lexer *Lexer) {
	processed := lexer.Process(line)
	s := line

	for _, v := range processed {
		if v.Text == "" {
			continue
		}
		a := strings.SplitAfter(s, v.Text)
		if len(a) == 0 {
			continue
		}
		s = strings.TrimPrefix(s, a[0])

		r.out.SetColor(v.TextColor, v.BGColor, false)
		r.out.WriteStr(a[0])
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
	if runtime.GOOS == "windows" {
		// WT_SESSION indicates Windows Terminal, which is more rational than the older cmd or ps terminals
		if _, ok := os.LookupEnv("WT_SESSION"); !ok {
			return
		}
	}
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

func wrap(s string, lim uint) string {
	esc := false

	init := make([]byte, 0, len(s))
	buf := bytes.NewBuffer(init)

	var current uint
	var wordBuf, spaceBuf bytes.Buffer
	var wordBufLen, spaceBufLen uint

	for _, char := range s {
		if char == '\x1b' {
			esc = true
			wordBuf.WriteRune(char)
			continue
		}
		if esc {
			if char == 'm' {
				esc = false
			}
			wordBuf.WriteRune(char)
			continue
		}
		if char == '\n' {
			if wordBuf.Len() == 0 {
				if current+spaceBufLen > lim {
					current = 0
				} else {
					current += spaceBufLen
					spaceBuf.WriteTo(buf)
				}
				spaceBuf.Reset()
				spaceBufLen = 0
			} else {
				current += spaceBufLen + wordBufLen
				spaceBuf.WriteTo(buf)
				spaceBuf.Reset()
				spaceBufLen = 0
				wordBuf.WriteTo(buf)
				wordBuf.Reset()
				wordBufLen = 0
			}
			buf.WriteRune(char)
			current = 0
		} else if unicode.IsSpace(char) {
			if spaceBuf.Len() == 0 || wordBuf.Len() > 0 {
				current += spaceBufLen + wordBufLen
				spaceBuf.WriteTo(buf)
				spaceBuf.Reset()
				spaceBufLen = 0
				wordBuf.WriteTo(buf)
				wordBuf.Reset()
				wordBufLen = 0
			}

			spaceBuf.WriteRune(char)
			spaceBufLen++
		} else {
			wordBuf.WriteRune(char)
			wordBufLen++

			if current+wordBufLen+spaceBufLen > lim && wordBufLen < lim {
				buf.WriteRune('\n')
				current = 0
				spaceBuf.Reset()
				spaceBufLen = 0
			}
		}
	}

	if wordBuf.Len() == 0 {
		if current+spaceBufLen <= lim {
			spaceBuf.WriteTo(buf)
		}
	} else {
		spaceBuf.WriteTo(buf)
		wordBuf.WriteTo(buf)
	}

	return buf.String()
}
