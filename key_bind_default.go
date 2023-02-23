package prompt

/*

Navigate
--------

* [x] Ctrl + f              Move to the next character (->)
* [x] Ctrl + b              Move to the previous character (<-)
* [x] Meta + f              Move to the start of the next word
* [x] Meta + b              Move to the start of the previous word
* [x] Ctrl + e              Move to the end of the line (End)
* [x] Ctrl + a              Move to the beginning of the line (Home)

Edit
----

* [x] Ctrl + d              Cut the character after the cursor (Delete)
* [x] Ctrl + h              Cut the character before the cursor (Backspace)
* [x] Meta + d              Cut the word after the cursor
* [x] Ctrl + w              Cut the word before the cursor
* [x] Ctrl + k              Cut the line after the cursor
* [x] Ctrl + u              Cut the line before the cursor

Utility
-------

* [x] Ctrl + l       				Clear the Screen, similar to the clear command
* [x] Ctrl + y              Paste the last thing to be cut

TBD
---

* [ ] Meta + backspace      Cut the word before the cursor
* [ ] Ctrl + t              Swap the last two characters before the cursor
* [ ] Meta + t              Swap the current word with the previous word
* [ ] Ctrl + _              Undo

*/

func defaultKeyBindings() []KeyBind {
	return append([]KeyBind{
		// Navigate
		{
			Key: ControlF,
			Fn:  KeyBindMoveCharNext,
		},
		{
			Key: ControlB,
			Fn:  KeyBindMoveCharPrev,
		},
		{
			Key: MetaF,
			Fn:  KeyBindMoveWordNext,
		},
		{
			Key: MetaB,
			Fn:  KeyBindMoveWordPrev,
		},
		{
			Key: ControlE,
			Fn:  KeyBindMoveLineEnd,
		},
		{
			Key: ControlA,
			Fn:  KeyBindMoveLineStart,
		},
		// Edit
		{
			Key: ControlD,
			Fn:  KeyBindCutCharAfter,
		},
		{
			Key: ControlH,
			Fn:  KeyBindCutCharBefore,
		},
		{
			Key: MetaD,
			Fn:  KeyBindCutWordAfter,
		},
		{
			Key: ControlW,
			Fn:  KeyBindCutWordBefore,
		},
		{
			Key: ControlK,
			Fn:  KeyBindCutLineAfter,
		},
		{
			Key: ControlU,
			Fn:  KeyBindCutLineBefore,
		},
		{
			Key: ControlY,
			Fn:  KeyBindInsertLastCutAfter,
		},
		// Util
		{
			Key: ControlL,
			Fn:  KeyBindUtilScreenClear,
		},
	}, defaultPlatformKeyBindings()...)
}
