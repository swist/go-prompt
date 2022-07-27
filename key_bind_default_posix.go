//go:build !windows

package prompt

/*

Navigate
--------

* [x] Meta + ->              Move to the start of the next word
* [x] Meta + <-              Move to the start of the previous word

*/

func defaultPlatformKeyBindings() []KeyBind {
	return []KeyBind{
		// Navigate
		{
			Key: MetaRight,
			Fn:  KeyBindMoveWordNext,
		},
		{
			Key: MetaLeft,
			Fn:  KeyBindMoveWordPrev,
		},
	}
}
