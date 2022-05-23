package prompt

import (
	"bytes"
	"strings"
	"unicode"
)

func (r *Render) expandDescription(formatted []Suggest, expand string, mh, mw, leftWidth, centerWidth int, cache bool) []Suggest {
	if mh <= 0 || mw <= 2 {
		return formatted
	}

	mw = mw - 2
	reformatted := make([]Suggest, 0, len(formatted))
	wrapped := strings.Split(wrap(expand, uint(mw)), "\n")
	lf := len(formatted)
	lw := len(wrapped)
	l := lf
	if lw > lf {
		l = lw
	}
	if l > mh {
		l = mh
	}
	for i := 0; i < l; i++ {
		var desc string
		if i < lw {
			if i == l-1 && lw > l {
				desc = "..."
			} else {
				desc = wrapped[i]
				if runeWidth(desc, cache) > mw {
					desc = truncate(desc, mw, "", "", cache)
				}
			}
			w := mw - runeWidth(desc, cache)
			if w < 0 {
				w = 0
			}
			pad := strings.Repeat(" ", w)
			desc = " " + desc + pad + " "
		} else {
			desc = strings.Repeat(" ", mw+2)
		}
		text := strings.Repeat(" ", leftWidth)
		note := strings.Repeat(" ", centerWidth)
		var typ SuggestType = SuggestTypeDefault
		if i < lf {
			text = formatted[i].Text
			note = formatted[i].Note
			typ = formatted[i].Type
		}
		reformatted = append(reformatted, Suggest{
			Text:        text,
			Note:        note,
			Description: desc,
			Type:        typ,
		})
	}

	return reformatted
}

func wrap(s string, lim uint) string {
	esc := false
	bold := false

	init := make([]byte, 0, len(s))
	buf := bytes.NewBuffer(init)

	var current uint
	var wordBuf, spaceBuf bytes.Buffer
	var wordBufLen, spaceBufLen uint

	for _, char := range s {
		if char == '\x1b' {
			esc = true
			wordBuf.WriteRune(char)
			continue
		}
		if esc {
			if char == 'm' {
				esc = false
			}
			wordBuf.WriteRune(char)
			continue
		}
		if char == BOLD_MARKER {
			bold = !bold
			wordBuf.WriteRune(char)
			continue
		}
		if char == '\n' {
			if wordBuf.Len() == 0 {
				if current+spaceBufLen > lim {
					current = 0
				} else {
					current += spaceBufLen
					spaceBuf.WriteTo(buf)
				}
				spaceBuf.Reset()
				spaceBufLen = 0
			} else {
				current += spaceBufLen + wordBufLen
				spaceBuf.WriteTo(buf)
				spaceBuf.Reset()
				spaceBufLen = 0
				wordBuf.WriteTo(buf)
				wordBuf.Reset()
				wordBufLen = 0
			}
			buf.WriteRune(char)
			current = 0
		} else if unicode.IsSpace(char) {
			if spaceBuf.Len() == 0 || wordBuf.Len() > 0 {
				current += spaceBufLen + wordBufLen
				spaceBuf.WriteTo(buf)
				spaceBuf.Reset()
				spaceBufLen = 0
				wordBuf.WriteTo(buf)
				wordBuf.Reset()
				wordBufLen = 0
			}

			spaceBuf.WriteRune(char)
			spaceBufLen++
		} else {
			wordBuf.WriteRune(char)
			wordBufLen++

			if current+wordBufLen+spaceBufLen > lim && wordBufLen < lim {
				if bold {
					// preserve bold markers across line breaks
					// TODO: this doesn't yet seem to catch all cases... perhaps when multiple lines
					//       wrap within one bolded region?  haven't gotten tot he bottom of that yet
					buf.WriteRune(BOLD_MARKER)
					buf.WriteRune('\n')
					buf.WriteRune(BOLD_MARKER)
				} else {
					buf.WriteRune('\n')
				}
				current = 0
				spaceBuf.Reset()
				spaceBufLen = 0
			}
		}
	}

	if wordBuf.Len() == 0 {
		if current+spaceBufLen <= lim {
			spaceBuf.WriteTo(buf)
		}
	} else {
		spaceBuf.WriteTo(buf)
		wordBuf.WriteTo(buf)
	}

	return buf.String()
}
