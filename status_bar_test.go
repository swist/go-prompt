package prompt

import (
	"fmt"
	"reflect"
	"testing"

	runewidth "github.com/mattn/go-runewidth"
)

func TestFormatStatusBar(t *testing.T) {
	tests := []struct {
		name  string
		sb    StatusBar
		want  StatusBar
		width int
	}{
		{
			name:  "empty",
			width: 10,
			sb:    StatusBar{},
			want:  StatusBar{{Text: "          "}},
		},
		{
			name:  "one element, fit, grow",
			width: 10,
			sb: StatusBar{
				{Text: "foo bar"},
			},
			want: StatusBar{
				{Text: "foo bar   "},
			},
		},
		{
			name:  "one element, fit, grow right aligned",
			width: 10,
			sb: StatusBar{
				{Text: "foo bar", Align: StatusAlignRight},
			},
			want: StatusBar{
				{Text: "   foo bar", Align: StatusAlignRight},
			},
		},
		{
			name:  "one element, no elastic, last too wide",
			width: 10,
			sb: StatusBar{
				{Text: "foo bar baz"},
			},
			want: StatusBar{
				{Text: "foo bar..."},
			},
		},
		{
			name:  "two elements, fit, grow last",
			width: 15,
			sb: StatusBar{
				{Text: "foo bar"},
				{Text: "baz qux"},
			},
			want: StatusBar{
				{Text: "foo bar"},
				{Text: "baz qux "},
			},
		},
		{
			name:  "two elements, last too wide, squashed ellipsis",
			width: 9,
			sb: StatusBar{
				{Text: "foo bar"},
				{Text: "baz qux"},
			},
			want: StatusBar{
				{Text: "foo bar"},
				{Text: ".."},
			},
		},
		{
			name:  "two elements, last too wide, ellipsis",
			width: 12,
			sb: StatusBar{
				{Text: "foo bar"},
				{Text: "baz qux"},
			},
			want: StatusBar{
				{Text: "foo bar"},
				{Text: "ba..."},
			},
		},
		{
			name:  "two elements, one too wide, elastic, only ellipsis",
			width: 10,
			sb: StatusBar{
				{Text: "foo bar", Elastic: true},
				{Text: "baz qux"},
			},
			want: StatusBar{
				{Text: "...", Elastic: true},
				{Text: "baz qux"},
			},
		},
		{
			name:  "two elements, one too wide, elastic, ellipsis",
			width: 12,
			sb: StatusBar{
				{Text: "foo bar", Elastic: true},
				{Text: "baz qux"},
			},
			want: StatusBar{
				{Text: "fo...", Elastic: true},
				{Text: "baz qux"},
			},
		},
		{
			name:  "five elements, two elastic, grow",
			width: 25,
			sb: StatusBar{
				{Text: "foo"},
				{Text: "bar baz", Elastic: true},
				{Text: "qux"},
				{Text: "quux corge", Elastic: true},
				{Text: "grault"},
			},
			want: StatusBar{
				{Text: "foo"},
				{Text: "ba...", Elastic: true},
				{Text: "qux"},
				{Text: "quux ...", Elastic: true},
				{Text: "grault"},
			},
		},
		{
			name:  "five elements, two too wide, two elastic",
			width: 25,
			sb: StatusBar{
				{Text: "foo"},
				{Text: "bar baz", Elastic: true},
				{Text: "qux"},
				{Text: "quux corge", Elastic: true},
				{Text: "grault"},
			},
			want: StatusBar{
				{Text: "foo"},
				{Text: "ba...", Elastic: true},
				{Text: "qux"},
				{Text: "quux ...", Elastic: true},
				{Text: "grault"},
			},
		},
		{
			name:  "five elements, two too wide, two elastic, too small, squashed ellipses",
			width: 15,
			sb: StatusBar{
				{Text: "foo"},
				{Text: "bar baz", Elastic: true},
				{Text: "qux"},
				{Text: "quux corge", Elastic: true},
				{Text: "grault"},
			},
			want: StatusBar{
				{Text: "foo"},
				{Text: "", Elastic: true},
				{Text: "qux"},
				{Text: "...", Elastic: true},
				{Text: "grault"},
			},
		},
		{
			name:  "five elements, two too wide, two elastic, too small, ellipsis",
			width: 23,
			sb: StatusBar{
				{Text: "foo"},
				{Text: "bar baz", Elastic: true},
				{Text: "qux"},
				{Text: "quux corge", Elastic: true},
				{Text: "grault"},
			},
			want: StatusBar{
				{Text: "foo"},
				{Text: "b...", Elastic: true},
				{Text: "qux"},
				{Text: "quux...", Elastic: true},
				{Text: "grault"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println("---", tt.name)
			got := tt.sb.fit(tt.width)
			if !equalStatusElementTexts(got, tt.want) {
				var g []string
				for _, v := range got {
					g = append(g, v.Text)
				}
				var w []string
				for _, v := range tt.want {
					w = append(w, v.Text)
				}
				t.Errorf("[%d] -> got %q, want %q", tt.width, g, w)
			}
			w := 0
			for _, el := range got {
				fmt.Printf("   %q, %t, %d\n", el.Text, el.Elastic, runewidth.StringWidth(el.Text))
				w += runewidth.StringWidth(el.Text)
			}
			if w != tt.width {
				t.Errorf("got width %d, want width %d", w, tt.width)
			}
		})
	}
}

func equalStatusElementTexts(a, b StatusBar) bool {
	if len(a) != len(b) {
		return false
	}
	for i, el := range a {
		if el.Text != b[i].Text {
			return false
		}
	}
	return true
}

func TestDistribute(t *testing.T) {
	tests := []struct {
		name string
		v    int
		n    int
		want []int
	}{
		{"0,1", 0, 1, []int{0}},
		{"5,1", 5, 1, []int{5}},
		{"1,5", 1, 5, []int{1, 0, 0, 0, 0}},
		{"4,5", 4, 5, []int{1, 1, 1, 1, 0}},
		{"9,5", 9, 5, []int{2, 2, 2, 2, 1}},
		{"8,4", 8, 4, []int{2, 2, 2, 2}},
		{"0,1", 0, 1, []int{0}},
		{"-5,1", -5, 1, []int{-5}},
		{"-1,5", -1, 5, []int{-1, 0, 0, 0, 0}},
		{"-4,5", -4, 5, []int{-1, -1, -1, -1, 0}},
		{"-9,5", -9, 5, []int{-2, -2, -2, -2, -1}},
		{"-8,4", -8, 4, []int{-2, -2, -2, -2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := distribute(tt.v, tt.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("[%d, %d] -> got %+v, want %+v", tt.v, tt.n, got, tt.want)
			}
		})
	}
}
