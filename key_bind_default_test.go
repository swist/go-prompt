package prompt

import "testing"

func TestDefaultKeyBindings(t *testing.T) {
	buf := NewBuffer()
	buf.InsertText("abcde", false, true)
	if buf.cursorPosition != len("abcde") {
		t.Errorf("Want %d, but got %d", len("abcde"), buf.cursorPosition)
	}

	// Go to the beginning of the line
	applyDefaultKeyBind(buf, ControlA)
	if buf.cursorPosition != 0 {
		t.Errorf("Want %d, but got %d", 0, buf.cursorPosition)
	}

	// Go to the end of the line
	applyDefaultKeyBind(buf, ControlE)
	if buf.cursorPosition != len("abcde") {
		t.Errorf("Want %d, but got %d", len("abcde"), buf.cursorPosition)
	}
}

func applyDefaultKeyBind(buf *Buffer, key Key) {
	for _, kb := range defaultKeyBindings() {
		if kb.Key == key {
			kb.Fn(buf)
		}
	}
}
