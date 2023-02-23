package prompt

import (
	runewidth "github.com/mattn/go-runewidth"
)

// LexerFunc is a callback from render.
type LexerFunc = func(d Document) []LexerElement

// LexerElement is a element of lexer.
type LexerElement struct {
	Text      string
	TextColor Color
	BGColor   Color
}

// Lexer is a struct with lexer param and function.
type Lexer struct {
	fn LexerFunc
}

// NewLexer returns a new trivial Lexer.
func NewLexer(textColor Color, bgColor Color) *Lexer {
	return &Lexer{
		fn: func(d Document) []LexerElement {
			return []LexerElement{
				{TextColor: textColor, BGColor: bgColor, Text: d.Text},
			}
		},
	}
}

// Process line with a custom function.
func (l *Lexer) Process(d Document) []LexerElement {
	return l.fn(d)
}

func splitLexedAtCursor(lexed []LexerElement, cursor int) (left, right []LexerElement, ll, rl int) {
	for _, el := range lexed {
		tw := runewidth.StringWidth(el.Text)
		if ll+tw <= cursor {
			// Entirely before cursor
			ll += tw
			left = append(left, el)
		} else if ll < cursor && ll+tw > cursor {
			// Split by cursor
			w := cursor - ll
			lEl := el
			lEl.Text = lEl.Text[:w]
			left = append(left, lEl)
			rEl := el
			rEl.Text = rEl.Text[w:]
			right = append(right, rEl)
			ll += w
			rl += tw - w
		} else {
			// Entirely after cursor
			right = append(right, el)
			rl += tw
		}
	}
	return
}
