package prompt

import (
	runewidth "github.com/mattn/go-runewidth"
)

type StatusAlign uint8

const (
	StatusAlignLeft StatusAlign = iota
	StatusAlignRight
)

// StatusElement contains a status bar component.
type StatusElement struct {
	Text      string
	TextColor *Color
	BGColor   *Color
	Bold      bool
	Elastic   bool
	Align     StatusAlign
}

func (el StatusElement) Equals(o StatusElement) bool {
	if el.Text != o.Text || el.Bold != o.Bold || el.Elastic != o.Elastic || el.Align != o.Align {
		return false
	}
	sfg, ofg := el.TextColor, o.TextColor
	if sfg != ofg && (sfg == nil || ofg == nil || *sfg != *ofg) {
		return false
	}
	sbg, obg := el.BGColor, o.BGColor
	if sbg != obg && (sbg == nil || obg == nil || *sbg != *obg) {
		return false
	}
	return true
}

func (el StatusElement) isElastic(i, l, real int) bool {
	return el.Elastic || (real == 0 && i == l-1)
}

func (el StatusElement) fit(delta int) (StatusElement, int) {
	tl := runewidth.StringWidth(el.Text)
	l := tl + delta
	el.Text = el.fitText(el.Text, l)
	fl := runewidth.StringWidth(el.Text)
	remainder := fl - l
	if remainder < 0 {
		panic("remainder < 0")
	}
	return el, remainder
}

func (el StatusElement) fitText(t string, l int) string {
	tl := runewidth.StringWidth(t)
	if tl == l {
		return t
	} else if tl < l {
		return el.growText(t, tl, l)
	}
	return el.shrinkText(t, tl, l)
}

func (el StatusElement) growText(t string, tl, l int) string {
	if el.Align == StatusAlignRight {
		return runewidth.FillLeft(t, l)
	}
	return runewidth.FillRight(t, l)
}

func (el StatusElement) shrinkText(t string, tl, l int) string {
	if l == 0 || tl == 0 {
		return ""
	} else if l == tl {
		return t
	} else if l > tl {
		panic("tl < l")
	}
	s, sl := "...", 3
	t = runewidth.Truncate(t, l, s)
	tl = runewidth.StringWidth(t)
	if tl < sl {
		t = s[:tl]
	} else if tl > l && l >= 0 {
		t = t[:l]
		tl = l
	}
	if tl < l {
		t = runewidth.FillRight(t, l)
	}
	return t
}
