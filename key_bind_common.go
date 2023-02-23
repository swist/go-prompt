package prompt

var commonKeyBindings = []KeyBind{
	// Go to the End of the line
	{
		Key: End,
		Fn:  KeyBindMoveLineEnd,
	},
	// Go to the beginning of the line
	{
		Key: Home,
		Fn:  KeyBindMoveLineStart,
	},
	// Delete character under the cursor
	{
		Key: Delete,
		Fn:  KeyBindCutCharAfter,
	},
	// Backspace
	{
		Key: Backspace,
		Fn:  KeyBindCutCharBefore,
	},
	// Right allow: Forward one character
	{
		Key: Right,
		Fn:  KeyBindMoveCharNext,
	},
	// Left allow: Backward one character
	{
		Key: Left,
		Fn:  KeyBindMoveCharPrev,
	},
}
