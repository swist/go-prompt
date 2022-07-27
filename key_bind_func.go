package prompt

import (
	"github.com/c-bata/go-prompt/internal/debug"
)

// Yank memory
var lastCut string

// Navigate

// Move to the next character
func KeyBindMoveCharNext(buf *Buffer) {
	buf.CursorRight(1)
}

// Move to the previous character
func KeyBindMoveCharPrev(buf *Buffer) {
	buf.CursorLeft(1)
}

// Move to the start of the next word
func KeyBindMoveWordNext(buf *Buffer) {
	buf.CursorRight(len([]rune(buf.Document().GetWordAfterCursorWithSpace())))
}

// Move to the start of the previous word
func KeyBindMoveWordPrev(buf *Buffer) {
	buf.CursorLeft(len([]rune(buf.Document().GetWordBeforeCursorWithSpace())))
}

// Move to the end of the line
func KeyBindMoveLineEnd(buf *Buffer) {
	x := []rune(buf.Document().TextAfterCursor())
	buf.CursorRight(len(x))
}

// Move to the beginning of the line
func KeyBindMoveLineStart(buf *Buffer) {
	x := []rune(buf.Document().TextBeforeCursor())
	buf.CursorLeft(len(x))
}

// Edit

// Cut the character after the cursor
func KeyBindCutCharAfter(buf *Buffer) {
	if buf.Text() != "" {
		lastCut = buf.Delete(1)
	}
}

// Cut the character before the cursor
func KeyBindCutCharBefore(buf *Buffer) {
	lastCut = buf.DeleteBeforeCursor(1)
}

// Cut the word after the cursor
func KeyBindCutWordAfter(buf *Buffer) {
	lastCut = buf.Delete(len([]rune(buf.Document().GetWordAfterCursorWithSpace())))
}

// Cut the word before the cursor
func KeyBindCutWordBefore(buf *Buffer) {
	lastCut = buf.DeleteBeforeCursor(len([]rune(buf.Document().GetWordBeforeCursorWithSpace())))
}

// Cut the line after the cursor
func KeyBindCutLineAfter(buf *Buffer) {
	x := []rune(buf.Document().TextAfterCursor())
	lastCut = buf.Delete(len(x))
}

// Cut the line before the cursor
func KeyBindCutLineBefore(buf *Buffer) {
	x := []rune(buf.Document().TextBeforeCursor())
	lastCut = buf.DeleteBeforeCursor(len(x))
}

// Insert the last thing to be cut after cursor
func KeyBindInsertLastCutAfter(buf *Buffer) {
	buf.InsertText(lastCut, false, true)
}

// Utility

// Clear the screen, similar to the clear command
func KeyBindUtilScreenClear(buf *Buffer) {
	consoleWriter.EraseScreen()
	consoleWriter.CursorGoTo(0, 0)
	debug.AssertNoError(consoleWriter.Flush())
}

// Deprecated

// Deprecated GoLineEnd Go to the End of the line
func GoLineEnd(buf *Buffer) {
	KeyBindMoveLineEnd(buf)
}

// Deprecated GoLineBeginning Go to the beginning of the line
func GoLineBeginning(buf *Buffer) {
	KeyBindMoveLineStart(buf)
}

// Deprecated DeleteChar Delete character under the cursor
func DeleteChar(buf *Buffer) {
	KeyBindCutCharAfter(buf)
}

// Deprecated DeleteBeforeChar Go to Backspace
func DeleteBeforeChar(buf *Buffer) {
	KeyBindCutCharBefore(buf)
}

// Deprecated GoRightChar Forward one character
func GoRightChar(buf *Buffer) {
	KeyBindMoveCharNext(buf)
}

// Deprecated GoLeftChar Backward one character
func GoLeftChar(buf *Buffer) {
	KeyBindMoveCharPrev(buf)
}

// Deprecated & unmapped

// Deprecated DeleteWord Delete word before the cursor
func DeleteWord(buf *Buffer) {
	buf.DeleteBeforeCursor(len([]rune(buf.Document().TextBeforeCursor())) - buf.Document().FindStartOfPreviousWordWithSpace())
}

// Deprecated GoRightWord Forward one word
func GoRightWord(buf *Buffer) {
	buf.CursorRight(buf.Document().FindEndOfCurrentWordWithSpace())
}

// Deprecated GoLeftWord Backward one word
func GoLeftWord(buf *Buffer) {
	buf.CursorLeft(len([]rune(buf.Document().TextBeforeCursor())) - buf.Document().FindStartOfPreviousWordWithSpace())
}
