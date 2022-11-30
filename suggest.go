package prompt

type SuggestType string

const (
	SuggestTypeDefault = ""
	SuggestTypeLabel   = "label"
)

// Suggest is printed when completing.
type Suggest struct {
	Text                string
	Note                string
	Description         string
	ExpandedDescription string
	Type                SuggestType
	Context             map[string]interface{}  `json:"-"`
	OnSelected          func(s Suggest) Suggest `json:"-"`
	OnCommitted         func(s Suggest) Suggest `json:"-"`
}

func (s Suggest) selected() Suggest {
	if s.OnSelected != nil {
		return s.OnSelected(s)
	}
	return s
}

func (s Suggest) committed() Suggest {
	if s.OnCommitted != nil {
		return s.OnCommitted(s)
	}
	return s
}
