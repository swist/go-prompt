package prompt

import "strings"

func (r *Render) renderLexable(line string, lexer *Lexer) {
	els := lexer.Process(line)
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
