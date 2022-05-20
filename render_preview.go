package prompt

import runewidth "github.com/mattn/go-runewidth"

func (r *Render) renderPreview(buffer *Buffer, completion *CompletionManager, lexer *Lexer, cursor int) int {
	if suggest, ok := completion.GetSelectedSuggestion(); ok {
		// r.out.EraseEndOfLine()
		cursor = r.backward(cursor, runewidth.StringWidth(buffer.Document().GetWordBeforeCursorUntilSeparator(completion.wordSeparator)))

		r.out.SetColor(r.previewSuggestionTextColor, r.previewSuggestionBGColor, false)
		r.out.WriteStr(suggest.Text)
		r.out.SetColor(DefaultColor, DefaultColor, false)
		cursor += runewidth.StringWidth(suggest.Text)

		rest := buffer.Document().TextAfterCursor()

		if lexer.IsEnabled {
			r.renderLexable(rest, lexer)
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
