package prompt

import (
	"strings"

	"github.com/c-bata/go-prompt/internal/debug"
	runewidth "github.com/mattn/go-runewidth"
)

const (
	shortenSuffix = "..."
	leftPrefix    = " "
	leftSuffix    = " "
	rightPrefix   = " "
	rightSuffix   = " "
)

type SuggestType string

const (
	SuggestTypeDefault = ""
	SuggestTypeLabel   = "label"
)

var (
	leftMargin       = runewidth.StringWidth(leftPrefix + leftSuffix)
	rightMargin      = runewidth.StringWidth(rightPrefix + rightSuffix)
	completionMargin = leftMargin + rightMargin
)

// Suggest is printed when completing.
type Suggest struct {
	Text        string
	Description string
	Type        SuggestType
	Context     interface{}             `json:"-"`
	OnSelected  func(s Suggest) Suggest `json:"-"`
	OnCommitted func(s Suggest) Suggest `json:"-"`
}

func (s *Suggest) selected() Suggest {
	if s.OnSelected != nil {
		return s.OnSelected(*s)
	}
	return *s
}

func (s *Suggest) committed() Suggest {
	if s.OnCommitted != nil {
		return s.OnCommitted(*s)
	}
	return *s
}

func (s *Suggest) CloneWithText(text string) Suggest {
	return Suggest{
		Text:        text,
		Description: s.Description,
		Context:     s.Context,
		OnSelected:  s.OnSelected,
		OnCommitted: s.OnCommitted,
	}
}

// CompletionManager manages which suggestion is now selected.
type CompletionManager struct {
	selected  int // -1 means nothing one is selected.
	tmp       []Suggest
	max       uint16
	completer Completer

	verticalScroll     int
	wordSeparator      string
	showAtStart        bool
	expandDescriptions bool
}

// GetSelectedSuggestion returns the selected item.
func (c *CompletionManager) GetSelectedSuggestion() (s Suggest, ok bool) {
	if c.selected == -1 {
		return Suggest{}, false
	} else if c.selected < -1 {
		debug.Assert(false, "must not reach here")
		c.selected = -1
		return Suggest{}, false
	} else if len(c.tmp) == 0 {
		return Suggest{}, false
	}
	s = c.tmp[c.selected].selected()
	return s, true
}

// GetSuggestions returns the list of suggestion.
func (c *CompletionManager) GetSuggestions() []Suggest {
	return c.tmp
}

// Reset to select nothing.
func (c *CompletionManager) Reset() {
	c.selected = -1
	c.verticalScroll = 0
	c.Update(*NewDocument())
}

// Update to update the suggestions.
func (c *CompletionManager) Update(in Document) {
	c.tmp = c.completer(in)
}

// Previous to select the previous suggestion item.
func (c *CompletionManager) Previous() {
	if c.verticalScroll == c.selected && c.selected > 0 {
		c.verticalScroll--
	}
	c.selected--
	c.update(c.Previous)
}

// Next to select the next suggestion item.
func (c *CompletionManager) Next() {
	if c.verticalScroll+int(c.max)-1 == c.selected {
		c.verticalScroll++
	}
	c.selected++
	c.update(c.Next)
}

// Completing returns whether the CompletionManager selects something one.
func (c *CompletionManager) Completing() bool {
	return c.selected != -1
}

func (c *CompletionManager) update(skip func()) {
	max := int(c.max)
	if len(c.tmp) < max {
		max = len(c.tmp)
	}

	if c.selected >= len(c.tmp) {
		c.Reset()
	} else if c.selected < -1 {
		c.selected = len(c.tmp) - 1
		c.verticalScroll = len(c.tmp) - max
	}

	if c.selected > -1 && c.tmp[c.selected].Type != SuggestTypeDefault {
		skip()
	}
}

func deleteBreakLineCharacters(s string) string {
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, "\r", "", -1)
	return s
}

func formatTexts(o []string, max int, prefix, suffix string) (new []string, width int) {
	l := len(o)
	n := make([]string, l)

	lenPrefix := runewidth.StringWidth(prefix)
	lenSuffix := runewidth.StringWidth(suffix)
	lenShorten := runewidth.StringWidth(shortenSuffix)
	min := lenPrefix + lenSuffix + lenShorten
	for i := 0; i < l; i++ {
		o[i] = deleteBreakLineCharacters(o[i])

		w := runewidth.StringWidth(o[i])
		if width < w {
			width = w
		}
	}

	if width == 0 {
		return n, 0
	}
	if min >= max {
		return n, 0
	}
	if lenPrefix+width+lenSuffix > max {
		width = max - lenPrefix - lenSuffix
	}

	for i := 0; i < l; i++ {
		x := runewidth.StringWidth(o[i])
		if x <= width {
			spaces := strings.Repeat(" ", width-x)
			n[i] = prefix + o[i] + spaces + suffix
		} else if x > width {
			x := runewidth.Truncate(o[i], width, shortenSuffix)
			// When calling runewidth.Truncate("您好xxx您好xxx", 11, "...") returns "您好xxx..."
			// But the length of this result is 10. So we need fill right using runewidth.FillRight.
			n[i] = prefix + runewidth.FillRight(x, width) + suffix
		}
	}
	return n, lenPrefix + width + lenSuffix
}

func formatSuggestions(suggests []Suggest, max int) ([]Suggest, int, int, int) {
	num := len(suggests)
	new := make([]Suggest, num)

	left := make([]string, num)
	for i := 0; i < num; i++ {
		left[i] = suggests[i].Text
	}
	right := make([]string, num)
	for i := 0; i < num; i++ {
		right[i] = suggests[i].Description
	}

	left, leftWidth := formatTexts(left, max, leftPrefix, leftSuffix)
	if leftWidth == 0 {
		return []Suggest{}, 0, 0, 0
	}
	right, rightWidth := formatTexts(right, max-leftWidth, rightPrefix, rightSuffix)

	for i := 0; i < num; i++ {
		new[i] = Suggest{Text: left[i], Description: right[i], Type: suggests[i].Type}
	}
	return new, leftWidth + rightWidth, leftWidth, rightWidth
}

// NewCompletionManager returns initialized CompletionManager object.
func NewCompletionManager(completer Completer, max uint16) *CompletionManager {
	return &CompletionManager{
		selected:  -1,
		max:       max,
		completer: completer,

		verticalScroll: 0,
	}
}
