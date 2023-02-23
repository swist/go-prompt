package prompt

import (
	"strings"

	runewidth "github.com/mattn/go-runewidth"
)

type stringWidthCache map[string]int

// We use the zero width space unicode char to designate bold region start/stop
const BOLD_MARKER = '\u200b'

func (sw stringWidthCache) get(s string) int {
	if s == "" {
		return 0
	}
	if s == " " {
		return 1
	}
	if w, ok := sw[s]; ok {
		return w
	}
	sw[s] = runewidth.StringWidth(s)
	return sw[s]
}

var swCache = stringWidthCache{}

type formattedStringCache map[string]map[int]string

func (f formattedStringCache) get(s string, max int) (string, bool) {
	if f[s] == nil {
		return "", false
	}
	v, ok := f[s][max]
	return v, ok
}

func (f formattedStringCache) set(s string, max int, v string) {
	if f[s] == nil {
		f[s] = make(map[int]string)
	}
	f[s][max] = v
}

var fsCacheLeft = formattedStringCache{}
var fsCacheRight = formattedStringCache{}

type dimensions struct {
	prefix  int
	suffix  int
	shorten int
	width   int
	widths  []int
}

func runeWidth(s string, cache bool) int {
	if cache {
		return swCache.get(s)
	}
	return runewidth.StringWidth(s)
}

func textsDimensions(texts []string, max int, prefix, suffix string, cache bool) dimensions {
	l := len(texts)
	d := dimensions{
		prefix:  runeWidth(prefix, cache),
		suffix:  runeWidth(suffix, cache),
		shorten: runeWidth(shortenSuffix, cache),
		widths:  make([]int, l),
	}

	min := d.prefix + d.suffix + d.shorten
	for i := 0; i < l; i++ {
		texts[i] = deleteBreakLineCharacters(texts[i])
		d.widths[i] = swCache.get(texts[i])
		if d.width < d.widths[i] {
			d.width = d.widths[i]
		}
	}
	if d.width == 0 || min >= max {
		d.width = 0
		return d
	}
	if d.prefix+d.width+d.suffix > max {
		d.width = max - d.prefix - d.suffix
	}
	return d
}

func formatTexts(texts []string, max int, prefix, suffix string, d dimensions, offset, limit int, rightAlign, cache bool) ([]string, int) {
	l := len(texts)
	r := l - offset
	if r < limit {
		limit = r
	}
	n := make([]string, limit)

	if d.width == 0 {
		return n, 0
	}
	fsCache := fsCacheLeft
	if rightAlign {
		fsCache = fsCacheRight
	}

	for x := 0; x < limit; x++ {
		i := offset + x
		if cache {
			if s, ok := fsCache.get(texts[i], d.width); ok {
				n[x] = s
				continue
			}
		}
		if d.widths[i] <= d.width {
			spaces := strings.Repeat(" ", d.width-d.widths[i])
			n[x] = prefix
			if rightAlign {
				n[x] += spaces + texts[i]
			} else {
				n[x] += texts[i] + spaces
			}
			n[x] += suffix
		} else {
			n[x] = truncate(texts[i], d.width, prefix, suffix, cache)
		}
		if cache {
			fsCache.set(texts[i], d.width, n[x])
		}
	}
	return n, d.prefix + d.width + d.suffix
}

func truncate(s string, w int, prefix, suffix string, cache bool) string {
	// If truncation interrupts a bold sequence, we need to close that sequence
	// before appending the shortenSuffix
	tail := shortenSuffix
	lt := runeWidth(tail, cache)
	bold := false
	x := 0
	for _, c := range s {
		if c == BOLD_MARKER {
			bold = !bold
			continue
		}
		if x+lt == w-1 {
			if bold {
				tail = string(BOLD_MARKER) + tail
			}
			break
		}
		x++
	}
	t := runewidth.Truncate(s, w, tail)
	// When calling runewidth.Truncate("您好xxx您好xxx", 11, "...") returns "您好xxx..."
	// But the length of this result is 10. So we need fill right using runewidth.FillRight.
	return prefix + runewidth.FillRight(t, w) + suffix
}

func formatSuggestions(suggests []Suggest, max, offset, limit int, cache bool) ([]Suggest, int, int, int, int) {
	num := len(suggests)

	left := make([]string, num)
	for i := 0; i < num; i++ {
		left[i] = suggests[i].Text
	}

	d := textsDimensions(left, max, leftPrefix, leftSuffix, cache)
	left, leftWidth := formatTexts(left, max, leftPrefix, leftSuffix, d, offset, limit, false, cache)
	if leftWidth == 0 {
		return []Suggest{}, 0, 0, 0, 0
	}

	center := make([]string, num)
	for i := 0; i < num; i++ {
		center[i] = suggests[i].Note
	}

	max = max - leftWidth
	d = textsDimensions(center, max, centerPrefix, centerSuffix, cache)
	center, centerWidth := formatTexts(center, max, centerPrefix, centerSuffix, d, offset, limit, true, cache)

	right := make([]string, num)
	for i := 0; i < num; i++ {
		right[i] = suggests[i].Description
	}

	max = max - centerWidth
	d = textsDimensions(right, max, rightPrefix, rightSuffix, cache)
	right, rightWidth := formatTexts(right, max, rightPrefix, rightSuffix, d, offset, limit, false, cache)

	l := len(left)
	list := make([]Suggest, l)
	for i := 0; i < l; i++ {
		list[i] = Suggest{Text: left[i], Note: center[i], Description: right[i], Type: suggests[i].Type}
	}

	return list, leftWidth + centerWidth + rightWidth, leftWidth, centerWidth, rightWidth
}

func deleteBreakLineCharacters(s string) string {
	// ~3x faster than invoking strings.Replace twice
	r := make([]byte, 0, len(s))
	for _, c := range []byte(s) {
		if c != '\n' && c != '\r' {
			r = append(r, c)
		}
	}
	return string(r)
}
