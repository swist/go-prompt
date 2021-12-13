package prompt

// StatusBar contains the status string and optional text and bg colors.
type StatusBar struct {
	Text      string
	TextColor *Color
	BGColor   *Color
	Bold      bool
}

func (s *StatusBar) Equal(o StatusBar) bool {
	if s == nil {
		return false
	}
	if s.Text != o.Text || s.Bold != o.Bold {
		return false
	}
	sfg, ofg := s.TextColor, o.TextColor
	if sfg != ofg && (sfg == nil || ofg == nil || *sfg != *ofg) {
		return false
	}
	sbg, obg := s.BGColor, o.BGColor
	if sbg != obg && (sbg == nil || obg == nil || *sbg != *obg) {
		return false
	}
	return true
}
