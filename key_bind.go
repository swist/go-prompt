package prompt

// KeyBindFunc receives buffer and processed it.
type KeyBindFunc func(*Buffer)

// KeyBind represents which key should do what operation.
type KeyBind struct {
	Key Key
	Fn  KeyBindFunc
}

// ASCIICodeBind represents which []byte should do what operation
type ASCIICodeBind struct {
	ASCIICode []byte
	Fn        KeyBindFunc
}

// KeyBindMode to switch a key binding flexibly.
type KeyBindMode string

const (
	// CommonKeyBind is a mode without any keyboard shortcut
	CommonKeyBind KeyBindMode = "common"
	// DefaultKeyBind is a mode to use bash-like keyboard shortcuts
	DefaultKeyBind KeyBindMode = "default"
)
