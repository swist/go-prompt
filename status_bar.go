package prompt

import (
	"strings"

	runewidth "github.com/mattn/go-runewidth"
)

type StatusBar []StatusElement

func (sb StatusBar) Equals(o StatusBar) bool {
	if len(sb) != len(o) {
		return false
	}
	for i := 0; i < len(sb); i++ {
		if !sb[i].Equals(o[i]) {
			return false
		}
	}
	return true
}

func (els StatusBar) fit(maxWidth int) StatusBar {
	if len(els) == 0 {
		return StatusBar{{Text: strings.Repeat(" ", maxWidth)}}
	}

	contentWidth := els.contentWidth()
	if contentWidth == maxWidth {
		return els
	}

	factor := 1
	delta := maxWidth - contentWidth
	if delta < 0 {
		factor = -1
		// delta = -delta
	}

	realElastic, adjustedElastic := els.countElastic()
	elastics, elasticsIndex := distribute(delta, adjustedElastic), 0
	excess := 0
	l := len(els)
	clones := make(StatusBar, l)
	for i, el := range els {
		clone := el
		clones[i] = clone
		if !el.isElastic(i, l, realElastic) {
			continue
		}
		remainder := 0
		clones[i], remainder = clone.fit(elastics[elasticsIndex])
		remainder *= factor
		if remainder > 0 {
			if i+1 < adjustedElastic {
				elastics[elasticsIndex+1] += remainder
			} else {
				excess += remainder
			}
		}
		elasticsIndex++
	}

	// TODO: distribute excess across non-elastic elements and truncate them as well
	// TODO: finally, any remaining excess should be distributed across all elements
	//       (applied in reverse) without adding more ellipses

	return clones
}

func (els StatusBar) contentWidth() (w int) {
	for _, el := range els {
		w += runewidth.StringWidth(el.Text)
	}
	return w
}

func (els StatusBar) countElastic() (real, adjusted int) {
	for _, el := range els {
		if el.Elastic {
			real++
		}
	}
	adjusted = real
	if real == 0 {
		adjusted = 1 // At least one element must be elastic, defaulting to the last element
	}
	return
}

func distribute(v, n int) []int {
	if n <= 0 {
		panic("n <= 0")
	}
	r := make([]int, n)
	if v == 0 {
		return r
	}
	f := 1
	if v < 0 {
		f = -1
		v = -v
	}
	d := v / n
	m := v % n
	for i := 0; i < n; i++ {
		r[i] = f * d
		if i < m {
			r[i] += f
		}
	}
	return r
}
