package prompt

import (
	"reflect"
	"strings"
	"testing"
)

func TestTrivialLexer_splitLexedAtCursor(t *testing.T) {
	lexer := NewLexer(DefaultColor, DefaultColor)

	type args struct {
		lexed  []LexerElement
		cursor int
	}
	tests := []struct {
		name   string
		args   args
		lWant  []string
		rWant  []string
		llWant int
		rlWant int
	}{
		{
			name: "nil",
			args: args{
				lexed:  nil,
				cursor: 0,
			},
			lWant:  nil,
			rWant:  nil,
			llWant: 0,
			rlWant: 0,
		},
		{
			name: "empty",
			args: args{
				lexed:  lexer.Process(Document{Text: ""}),
				cursor: 0,
			},
			lWant:  []string{""},
			rWant:  nil,
			llWant: 0,
			rlWant: 0,
		},
		{
			name: "oobPosCursor",
			args: args{
				lexed:  lexer.Process(Document{Text: "one two"}),
				cursor: 8,
			},
			lWant:  []string{"one two"},
			rWant:  nil,
			llWant: 7,
			rlWant: 0,
		},
		{
			name: "oobNegCursor",
			args: args{
				lexed:  lexer.Process(Document{Text: "one two"}),
				cursor: -1,
			},
			lWant:  nil,
			rWant:  []string{"one two"},
			llWant: 0,
			rlWant: 7,
		},
		{
			name: "allLeft",
			args: args{
				lexed:  lexer.Process(Document{Text: "one two three"}),
				cursor: 13,
			},
			lWant:  []string{"one two three"},
			rWant:  nil,
			llWant: 13,
			rlWant: 0,
		},
		{
			name: "allRight",
			args: args{
				lexed:  lexer.Process(Document{Text: "one two three"}),
				cursor: 0,
			},
			lWant:  nil,
			rWant:  []string{"one two three"},
			llWant: 0,
			rlWant: 13,
		},
		{
			name: "splitOnBoundary",
			args: args{
				lexed:  lexer.Process(Document{Text: "one two three"}),
				cursor: 3,
			},
			lWant:  []string{"one"},
			rWant:  []string{" two three"},
			llWant: 3,
			rlWant: 10,
		},
		{
			name: "splitInWord",
			args: args{
				lexed:  lexer.Process(Document{Text: "one two three"}),
				cursor: 6,
			},
			lWant:  []string{"one tw"},
			rWant:  []string{"o three"},
			llWant: 6,
			rlWant: 7,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lGot, rGot, llGot, rlGot := splitLexedAtCursor(test.args.lexed, test.args.cursor)
			if !reflect.DeepEqual(reduceGot(lGot), test.lWant) {
				t.Errorf("splitLexedAtCursor() lGot = %v, want %v", lGot, test.lWant)
			}
			if !reflect.DeepEqual(reduceGot(rGot), test.rWant) {
				t.Errorf("splitLexedAtCursor() rGot = %v, want %v", rGot, test.rWant)
			}
			if llGot != test.llWant {
				t.Errorf("splitLexedAtCursor() llGot = %v, want %v", llGot, test.llWant)
			}
			if rlGot != test.rlWant {
				t.Errorf("splitLexedAtCursor() rlGot = %v, want %v", rlGot, test.rlWant)
			}
		})
	}
}

func TestLexer_splitLexedAtCursor(t *testing.T) {
	lexer := &Lexer{
		func(d Document) (lexed []LexerElement) {
			words := strings.Split(d.Text, " ")
			for i, word := range words {
				if i > 0 {
					lexed = append(lexed, LexerElement{Text: " "})
				}
				lexed = append(lexed, LexerElement{Text: word})
			}
			return
		},
	}

	type args struct {
		lexed  []LexerElement
		cursor int
	}
	tests := []struct {
		name   string
		args   args
		lWant  []string
		rWant  []string
		llWant int
		rlWant int
	}{
		{
			name: "nil",
			args: args{
				lexed:  nil,
				cursor: 0,
			},
			lWant:  nil,
			rWant:  nil,
			llWant: 0,
			rlWant: 0,
		},
		{
			name: "empty",
			args: args{
				lexed:  lexer.Process(Document{Text: ""}),
				cursor: 0,
			},
			lWant:  []string{""},
			rWant:  nil,
			llWant: 0,
			rlWant: 0,
		},
		{
			name: "oobPosCursor",
			args: args{
				lexed:  lexer.Process(Document{Text: "one two"}),
				cursor: 8,
			},
			lWant:  []string{"one", " ", "two"},
			rWant:  nil,
			llWant: 7,
			rlWant: 0,
		},
		{
			name: "oobNegCursor",
			args: args{
				lexed:  lexer.Process(Document{Text: "one two"}),
				cursor: -1,
			},
			lWant:  nil,
			rWant:  []string{"one", " ", "two"},
			llWant: 0,
			rlWant: 7,
		},
		{
			name: "allLeft",
			args: args{
				lexed:  lexer.Process(Document{Text: "one two three"}),
				cursor: 13,
			},
			lWant:  []string{"one", " ", "two", " ", "three"},
			rWant:  nil,
			llWant: 13,
			rlWant: 0,
		},
		{
			name: "allRight",
			args: args{
				lexed:  lexer.Process(Document{Text: "one two three"}),
				cursor: 0,
			},
			lWant:  nil,
			rWant:  []string{"one", " ", "two", " ", "three"},
			llWant: 0,
			rlWant: 13,
		},
		{
			name: "splitOnBoundary",
			args: args{
				lexed:  lexer.Process(Document{Text: "one two three"}),
				cursor: 3,
			},
			lWant:  []string{"one"},
			rWant:  []string{" ", "two", " ", "three"},
			llWant: 3,
			rlWant: 10,
		},
		{
			name: "splitInWord",
			args: args{
				lexed:  lexer.Process(Document{Text: "one two three"}),
				cursor: 6,
			},
			lWant:  []string{"one", " ", "tw"},
			rWant:  []string{"o", " ", "three"},
			llWant: 6,
			rlWant: 7,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lGot, rGot, llGot, rlGot := splitLexedAtCursor(test.args.lexed, test.args.cursor)
			if !reflect.DeepEqual(reduceGot(lGot), test.lWant) {
				t.Errorf("splitLexedAtCursor() lGot = %v, want %v", lGot, test.lWant)
			}
			if !reflect.DeepEqual(reduceGot(rGot), test.rWant) {
				t.Errorf("splitLexedAtCursor() rGot = %v, want %v", rGot, test.rWant)
			}
			if llGot != test.llWant {
				t.Errorf("splitLexedAtCursor() llGot = %v, want %v", llGot, test.llWant)
			}
			if rlGot != test.rlWant {
				t.Errorf("splitLexedAtCursor() rlGot = %v, want %v", rlGot, test.rlWant)
			}
		})
	}
}

func reduceGot(lexed []LexerElement) (reduced []string) {
	for _, e := range lexed {
		reduced = append(reduced, e.Text)
	}
	return
}
