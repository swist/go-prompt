package prompt

import (
	runewidth "github.com/mattn/go-runewidth"
)

func (r *Render) renderPreview(d *Document, completion *CompletionManager, cursor int) (int, string) {
	suggest, ok := completion.GetSelectedSuggestion()
	if !ok {
		return cursor, ""
	}

	sw := runewidth.StringWidth(d.GetWordBeforeCursorUntilSeparator(completion.wordSeparator))
	cursor = r.backward(cursor, sw)
	r.out.SetColor(r.previewSuggestionTextColor, r.previewSuggestionBGColor, false)
	r.out.WriteStr(suggest.Text)
	r.out.SetColor(DefaultColor, DefaultColor, false)
	cursor += runewidth.StringWidth(suggest.Text)
	return cursor, suggest.Text
}
