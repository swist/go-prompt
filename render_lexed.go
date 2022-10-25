package prompt

import (
	"strings"
	runewidth "github.com/mattn/go-runewidth"
)

func (r *Render) renderLexable(d Document, lexer *Lexer) {
	els := lexer.Process(d)
	r.renderLexed(els)
}

func (r *Render) renderLexed(els []LexerElement) {
	for i, v := range els {
		if v.Text == "" {
			continue
		}
		r.out.SetColor(v.TextColor, v.BGColor, false)
		s := v.Text
		if i == len(els)-1 {
			s = strings.TrimRight(s, "\r\n")
		}
		r.out.WriteStr(s)
	}
}

func (r *Render) renderLexedAfterCursor(lexed []LexerElement, cursor int) int {
	rest := 0
	pos := 0
	var els []LexerElement
	for _, el := range lexed {
		tw := runewidth.StringWidth(el.Text)
		if pos+tw <= cursor {
			pos += tw
		} else if pos < cursor && pos+tw > cursor {
			w := cursor - pos
			el.Text = el.Text[w:]
			els = append(els, el)
			pos += tw
			rest += tw - w
		} else {
			els = append(els, el)
			rest += tw
		}
	}
	r.renderLexed(els)
	return rest
}
