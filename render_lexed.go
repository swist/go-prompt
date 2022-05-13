package prompt

import "strings"

func (r *Render) renderLexed(line string, lexer *Lexer) {
	processed := lexer.Process(line)
	s := line

	for _, v := range processed {
		if v.Text == "" {
			continue
		}
		a := strings.SplitAfter(s, v.Text)
		if len(a) == 0 {
			continue
		}
		s = strings.TrimPrefix(s, a[0])

		r.out.SetColor(v.TextColor, v.BGColor, false)
		r.out.WriteStr(a[0])
	}
}
