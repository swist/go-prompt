package prompt

import (
	"math"
	"regexp"
	"strings"

	runewidth "github.com/mattn/go-runewidth"
)

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
		desc := selectedSuggest.Description
		formatted = r.expandDescription(formatted, desc, int(completions.max), rightWidth, leftWidth)
		completionHeight = len(formatted)
	}

	r.allocateArea(completionHeight + statusBarHeight)

	pw := runewidth.StringWidth(prefix)
	lw := pw + runewidth.StringWidth(buf.Document().Text)
	_, y := r.toPos(lw)
	r.eraseArea(y)

	cursor := pw + runewidth.StringWidth(buf.Document().TextBeforeCursor())
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
		tfg, tbg := r.suggestionTextColor, r.suggestionBGColor
		if i == selected {
			tfg, tbg = r.selectedSuggestionTextColor, r.selectedSuggestionBGColor
		} else if formatted[i].Type != SuggestTypeDefault {
			tfg, tbg = r.getSuggestionTypeColor(formatted[i].Type)
		}
		r.renderMarkup(formatted[i].Text, tfg, tbg)

		dfg, dbg := r.descriptionTextColor, r.descriptionBGColor
		if i == selected && !completions.expandDescriptions {
			dfg, dbg = r.selectedDescriptionTextColor, r.selectedDescriptionBGColor
		} else if formatted[i].Type != SuggestTypeDefault && (!completions.expandDescriptions || selected < 0) {
			dfg, dbg = r.getSuggestionTypeColor(formatted[i].Type)
		}
		r.renderMarkup(formatted[i].Description, dfg, dbg)

		if isScrollThumb(i) {
			r.out.SetColor(DefaultColor, r.scrollbarThumbColor, false)
		} else {
			r.out.SetColor(DefaultColor, r.scrollbarBGColor, false)
		}
		r.out.WriteStr(" ")
		r.out.SetColor(DefaultColor, DefaultColor, false)

		r.lineWrap(cursor + width)
		cursor = r.backward(cursor+width, width)
	}

	if x+width >= int(r.col) {
		r.out.CursorForward(x + width - int(r.col))
	}

	r.out.CursorUp(completionHeight)
	r.out.SetColor(DefaultColor, DefaultColor, false)
}

var boldRe = regexp.MustCompile(`\x60((?:[^\x60\\]|\\.)*)\x60`)

func (r *Render) renderMarkup(s string, fg Color, bg Color) {
	r.out.SetColor(fg, bg, false)
	if !r.enableMarkup {
		r.out.WriteStr(s)
		return
	}

	// pad is a cheap trick to make up for the fact that text dimensions were computed before
	// parsing the markup; this could be done in a better way that would yield nicer results
	pad := 0

	begin, end := 0, 0
	for _, match := range boldRe.FindAllStringIndex(s, -1) {
		end = match[0]
		if match[1] != 0 {
			r.out.WriteStr(s[begin:end])
		}
		begin = match[1]
		r.out.SetColor(fg, bg, true)
		r.out.WriteStr(s[match[0]+1 : match[1]-1])
		pad += 2
		r.out.SetColor(fg, bg, false)
	}

	if end != len(s) {
		r.out.WriteStr(s[begin:])
	}

	if pad > 0 {
		r.out.WriteStr(strings.Repeat(" ", pad))
	}
}

func (r *Render) getSuggestionTypeColor(typ SuggestType) (textColor, bgColor Color) {
	switch typ {
	case SuggestTypeLabel:
		return r.suggestTypeLabelTextColor, r.suggestTypeLabelBGColor
	default:
		return r.suggestionTextColor, r.suggestionBGColor
	}
}
