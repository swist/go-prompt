package prompt

import "github.com/c-bata/go-prompt/internal/debug"

func (r *Render) RenderStatusBar() {
	if r.statusBar == nil {
		return
	}

	defer func() { debug.AssertNoError(r.out.Flush()) }()

	r.out.SaveCursor()
	r.out.HideCursor()
	defer func() {
		r.out.UnSaveCursor()
		r.out.CursorUp(0)
		r.out.ShowCursor()
	}()

	r.out.CursorGoTo(int(r.row), 1)
	for _, el := range r.statusBar.fit(int(r.col)) {
		r.renderStatusElement(el)
	}
	r.out.SetColor(DefaultColor, DefaultColor, false)
}

func (r *Render) renderStatusElement(el StatusElement) {
	fgColor := r.statusBarTextColor
	if el.TextColor != nil {
		fgColor = *el.TextColor
	}
	bgColor := r.statusBarBGColor
	if el.BGColor != nil {
		bgColor = *el.BGColor
	}
	r.out.SetColor(fgColor, bgColor, el.Bold)
	r.out.EraseEndOfLine()
	r.out.WriteStr(el.Text)
}
