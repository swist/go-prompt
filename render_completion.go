package prompt

import (
	"math"

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
		formatted = r.expandDescription(formatted, selectedSuggest.Description, int(completions.max), rightWidth, leftWidth)
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
		cursor = r.backward(cursor+width, width)
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
