package prompt

import (
	runewidth "github.com/mattn/go-runewidth"
)

func (r *Render) renderPreview(d *Document, completion *CompletionManager, cursor int) int {
	if suggest, ok := completion.GetSelectedSuggestion(); ok {
		sw := runewidth.StringWidth(d.GetWordBeforeCursorUntilSeparator(completion.wordSeparator))
		cursor = r.backward(cursor, sw)
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
	}
	return cursor
}
