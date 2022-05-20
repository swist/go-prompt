package prompt

func (r *Render) renderLexable(line string, lexer *Lexer) {
	els := lexer.Process(line)
	r.renderLexed(els)
}

func (r *Render) renderLexed(els []LexerElement) {
	for _, v := range els {
		if v.Text == "" {
			continue
		}
		r.out.SetColor(v.TextColor, v.BGColor, false)
		r.out.WriteStr(v.Text)
	}
}
