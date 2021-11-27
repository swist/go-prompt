package prompt

import (
	"github.com/c-bata/go-prompt/internal/debug"
)

/*

========
PROGRESS
========

Moving the cursor
-----------------

* [x] Ctrl + a   Go to the beginning of the line (Home)
* [x] Ctrl + e   Go to the End of the line (End)
* [x] Ctrl + p   Previous command (Up arrow)
* [x] Ctrl + n   Next command (Down arrow)
* [x] Ctrl + f   Forward one character
* [x] Ctrl + b   Backward one character
* [x] Ctrl + xx  Toggle between the start of line and current cursor position

Editing
-------

* [x] Ctrl + L   Clear the Screen, similar to the clear command
* [x] Ctrl + d   Delete character under the cursor
* [x] Ctrl + h   Delete character before the cursor (Backspace)

* [x] Ctrl + w   Cut the Word before the cursor to the clipboard.
* [x] Ctrl + k   Cut the Line after the cursor to the clipboard.
* [x] Ctrl + u   Cut/delete the Line before the cursor to the clipboard.

* [x] Meta + d   Delete from cursor to the end of the current word
* [x] Meta + ->  Move to the start of the next word
* [x] Meta + f   Move to the start of the next word
* [x] Meta + <-  Move to the start of the previous word
* [x] Meta + b   Move to the start of the previous word

* [ ] Ctrl + t   Swap the last two characters before the cursor (typo).
* [ ] Esc  + t   Swap the last two words before the cursor.

* [ ] Ctrl + y   Paste the last thing to be cut (yank)
* [ ] Ctrl + _   Undo

*/

var emacsKeyBindings = []KeyBind{
	// Go to the end of the line
	{
		Key: ControlE,
		Fn: func(buf *Buffer) {
			x := []rune(buf.Document().TextAfterCursor())
			buf.CursorRight(len(x))
		},
	},
	// Go to the beginning of the line
	{
		Key: ControlA,
		Fn: func(buf *Buffer) {
			x := []rune(buf.Document().TextBeforeCursor())
			buf.CursorLeft(len(x))
		},
	},
	// Cut the line after the cursor
	{
		Key: ControlK,
		Fn: func(buf *Buffer) {
			x := []rune(buf.Document().TextAfterCursor())
			buf.Delete(len(x))
		},
	},
	// Cut/delete the line before the cursor
	{
		Key: ControlU,
		Fn: func(buf *Buffer) {
			x := []rune(buf.Document().TextBeforeCursor())
			buf.DeleteBeforeCursor(len(x))
		},
	},
	// Delete character under the cursor
	{
		Key: ControlD,
		Fn: func(buf *Buffer) {
			if buf.Text() != "" {
				buf.Delete(1)
			}
		},
	},
	// Backspace
	{
		Key: ControlH,
		Fn: func(buf *Buffer) {
			buf.DeleteBeforeCursor(1)
		},
	},
	// Right allow: Forward one character
	{
		Key: ControlF,
		Fn: func(buf *Buffer) {
			buf.CursorRight(1)
		},
	},
	// Left allow: Backward one character
	{
		Key: ControlB,
		Fn: func(buf *Buffer) {
			buf.CursorLeft(1)
		},
	},
	// Cut the Word before the cursor.
	{
		Key: ControlW,
		Fn: func(buf *Buffer) {
			buf.DeleteBeforeCursor(len([]rune(buf.Document().GetWordBeforeCursorWithSpace())))
		},
	},
	// Clear the Screen, similar to the clear command
	{
		Key: ControlL,
		Fn: func(buf *Buffer) {
			consoleWriter.EraseScreen()
			consoleWriter.CursorGoTo(0, 0)
			debug.AssertNoError(consoleWriter.Flush())
		},
	},
	// Delete from cursor to the end of the current word
	{
		Key: MetaD,
		Fn: func(buf *Buffer) {
			buf.Delete(len([]rune(buf.Document().GetWordAfterCursorWithSpace())))
		},
	},
	// Move to the start of the previous word.
	{
		Key: MetaB,
		Fn: func(buf *Buffer) {
			buf.CursorLeft(len([]rune(buf.Document().GetWordBeforeCursorWithSpace())))
		},
	},
	// Move to the start of the previous word.
	{
		Key: MetaLeft,
		Fn: func(buf *Buffer) {
			buf.CursorLeft(len([]rune(buf.Document().GetWordBeforeCursorWithSpace())))
		},
	},
	// Move to the start of the next word.
	{
		Key: MetaF,
		Fn: func(buf *Buffer) {
			buf.CursorRight(len([]rune(buf.Document().GetWordAfterCursorWithSpace())))
		},
	},
	// Move to the start of the next word.
	{
		Key: MetaRight,
		Fn: func(buf *Buffer) {
			buf.CursorRight(len([]rune(buf.Document().GetWordAfterCursorWithSpace())))
		},
	},
}
