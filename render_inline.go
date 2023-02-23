package prompt

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

func (r *Render) renderInline(d *Document, completion *CompletionManager, cursor int) (int, int) {
	width := 0
	if completion.inline == "" {
		return cursor, width
	}
	if !strings.HasSuffix(d.Text, " ") {
		r.out.SetColor(DefaultColor, DefaultColor, false)
		r.out.WriteStr(" ")
		width = 1
	}
	r.out.SetColor(r.inlineTextColor, r.inlineBGColor, false)
	r.out.WriteStr(completion.inline)
	r.out.SetColor(DefaultColor, DefaultColor, false)
	width += runewidth.StringWidth(completion.inline)
	return cursor + width, width
}
