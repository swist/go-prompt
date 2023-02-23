package prompt

import (
	"math"
	"regexp"

	runewidth "github.com/mattn/go-runewidth"
)

func (r *Render) renderCompletion(buf *Buffer, completions *CompletionManager) {
	suggestions := completions.GetSuggestions()
	if len(suggestions) == 0 {
		return
	}
	columns := int(r.col)
	prefix := r.getCurrentPrefix(buf, false)
	maxWidth := columns - runewidth.StringWidth(prefix) - scrollBarWidth
	// TODO: should completion height be returned by formatSuggestions?
	completionHeight := len(suggestions)
	if completionHeight > int(completions.max) {
		completionHeight = int(completions.max)
	}
	formatted, width, leftWidth, centerWidth, rightWidth := formatSuggestions(suggestions, maxWidth, completions.verticalScroll, completionHeight, r.stringCaches)
	width += scrollBarWidth

	showingLast := completions.verticalScroll+completionHeight == len(suggestions)
	selected := completions.selected - completions.verticalScroll
	if selected >= 0 && completions.expandDescriptions {
		selectedSuggest := suggestions[completions.selected]
		desc := selectedSuggest.Description
		if selectedSuggest.ExpandedDescription != "" {
			desc = selectedSuggest.ExpandedDescription
		}
		formatted = r.expandDescription(formatted, desc, int(completions.max), rightWidth, leftWidth, centerWidth, r.stringCaches)
		completionHeight = len(formatted)
	}

	r.allocateArea(completionHeight + statusBarHeight)

	prefixWidth := runewidth.StringWidth(prefix)
	d := buf.Document()
	lw := prefixWidth + runewidth.StringWidth(d.Text)
	_, y := r.toPos(lw)
	r.eraseArea(y)

	cursor := prefixWidth + runewidth.StringWidth(buf.Document().TextBeforeCursor())
	previewWidth := 0
	if suggest, ok := completions.GetSelectedSuggestion(); ok {
		previewWidth = runewidth.StringWidth(suggest.Text)
		// Account for the cursor offset into the previewWidth
		offset := runewidth.StringWidth(d.GetWordBeforeCursorUntilSeparator(completions.wordSeparator))
		previewWidth = previewWidth - offset
		cursor += previewWidth
	}
	x, h1 := r.toPos(cursor)
	if x+width >= columns {
		// If the line position plus completion width is greater than the row width, rewind the cursor by the difference
		adj := x + width - columns
		cursor = r.backward(cursor, adj)
		// The previewWidth is used later to rewind the cursor again for completion box alignment, so adjust it as well
		previewWidth -= adj
		_, h2 := r.toPos(cursor - previewWidth)
		if previewWidth < 0 || h2 < h1 {
			// If the preview width is negative or would cause the cursor to retreat to the previous line, disregard it
			previewWidth = 0
		}
	} else if h1 > 0 && x < prefixWidth+previewWidth {
		// If the cursor has wrapped to a newline but the x pos is less than the prefixWidth+previewWidth we need to advance
		// the cursor (by the complement of the x pos with the prefix width) instead of rewind it so that the completion
		// view does not wrap backward to the previous line and is aligned with the prefix on the current line
		previewWidth = x - prefixWidth
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

	cursor = r.backward(cursor, previewWidth)
	for i := 0; i < completionHeight; i++ {
		r.out.CursorDown(1)
		tfg, tbg := r.suggestionTextColor, r.suggestionBGColor
		if i == selected {
			tfg, tbg = r.selectedSuggestionTextColor, r.selectedSuggestionBGColor
		} else if formatted[i].Type != SuggestTypeDefault {
			tfg, tbg = r.getSuggestionTypeColor(formatted[i].Type)
		}
		r.renderMarkup(formatted[i].Text, tfg, tbg)
		r.renderMarkup(formatted[i].Note, tfg, tbg)

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
	_ = r.move(cursor, cursor+previewWidth)

	if x+width >= columns {
		r.out.CursorForward(x + width - columns)
	}

	r.out.CursorUp(completionHeight)
	r.out.SetColor(DefaultColor, DefaultColor, false)
}

var boldRe = regexp.MustCompile("\u200b([^\u200b]+)\u200b")

func (r *Render) renderMarkup(s string, fg Color, bg Color) {
	r.out.SetColor(fg, bg, false)
	if !r.enableMarkup {
		r.out.WriteStr(s)
		return
	}

	begin, end := 0, 0
	zw := len("\u200b")
	for _, match := range boldRe.FindAllStringIndex(s, -1) {
		end = match[0]
		if match[1] != 0 {
			r.out.WriteStr(s[begin:end])
		}
		begin = match[1]
		r.out.SetColor(fg, bg, true)
		out := s[match[0]+zw : match[1]-zw]
		r.out.WriteStr(out)
		r.out.SetColor(fg, bg, false)
	}

	if end != len(s) {
		r.out.WriteStr(s[begin:])
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
