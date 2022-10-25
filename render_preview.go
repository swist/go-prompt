package prompt

import (
	runewidth "github.com/mattn/go-runewidth"
)

func (r *Render) renderPreview(d *Document, completion *CompletionManager, cursor int, lexed []LexerElement) int {
	if suggest, ok := completion.GetSelectedSuggestion(); ok {
		r.out.EraseEndOfLine()
		cursor = r.backward(cursor, runewidth.StringWidth(d.GetWordBeforeCursorUntilSeparator(completion.wordSeparator)))

		r.out.SetColor(r.previewSuggestionTextColor, r.previewSuggestionBGColor, false)
		r.out.WriteStr(suggest.Text)
		if suggest.Next != "" {
			r.out.SetColor(DefaultColor, DefaultColor, false)
			r.out.WriteStr(" ")
			r.out.SetColor(r.previewNextTextColor, r.previewNextBGColor, false)
			r.out.WriteStr(suggest.Next)
		}
		r.out.SetColor(DefaultColor, DefaultColor, false)
		cursor += runewidth.StringWidth(suggest.textWithNext())

		rest := 0
		if len(lexed) > 0 {
			rest = r.renderLexedAfterCursor(lexed, d.cursorPosition)
		} else {
			s := d.TextAfterCursor()
			rest = runewidth.StringWidth(s)
			r.out.WriteStr(s)
		}

		r.out.SetColor(DefaultColor, DefaultColor, false)

		cursor += rest
		r.lineWrap(cursor)

		cursor = r.backward(cursor, rest)
	}
	return cursor
}
