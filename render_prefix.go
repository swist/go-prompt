package prompt

func (r *Render) renderPrefix(buffer *Buffer, breakline bool) {
	r.out.SetColor(r.prefixTextColor, r.prefixBGColor, false)
	r.out.WriteStr(r.getCurrentPrefix(buffer, breakline))
	r.out.SetColor(DefaultColor, DefaultColor, false)
	r.out.Flush()
}

// getCurrentPrefix to get current prefix.
// If live-prefix is enabled, return live-prefix.
func (r *Render) getCurrentPrefix(buffer *Buffer, breakline bool) string {
	if prefix, ok := r.livePrefixCallback(buffer.Document(), breakline); ok {
		return prefix
	}
	return r.prefix
}
