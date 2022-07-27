//go:build windows

package prompt

/*

Navigate
--------

* [x] Control + ->           Move to the start of the next word
* [x] Control + <-           Move to the start of the previous word

*/

func defaultPlatformKeyBindings() []KeyBind {
	return []KeyBind{
		// Navigate
		{
			Key: ControlRight,
			Fn:  KeyBindMoveWordNext,
		},
		{
			Key: ControlLeft,
			Fn:  KeyBindMoveWordPrev,
		},
	}
}
