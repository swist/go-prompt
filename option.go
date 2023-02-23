package prompt

// Option is the type to replace default parameters.
// prompt.New accepts any number of options (this is functional option pattern).
type Option func(prompt *Prompt) error

// OptionParser to set a custom ConsoleParser object. An argument should implement ConsoleParser interface.
func OptionParser(x ConsoleParser) Option {
	return func(p *Prompt) error {
		p.in = x
		return nil
	}
}

// OptionWriter to set a custom ConsoleWriter object. An argument should implement ConsoleWriter interface.
func OptionWriter(x ConsoleWriter) Option {
	return func(p *Prompt) error {
		registerConsoleWriter(x)
		p.renderer.out = x
		return nil
	}
}

// OptionTitle to set title displayed at the header bar of terminal.
func OptionTitle(x string) Option {
	return func(p *Prompt) error {
		p.renderer.title = x
		return nil
	}
}

// OptionPrefix to set prefix string.
func OptionPrefix(x string) Option {
	return func(p *Prompt) error {
		p.renderer.prefix = x
		return nil
	}
}

// OptionInitialBufferText to set the initial buffer text.
func OptionInitialBufferText(x string) Option {
	return func(p *Prompt) error {
		p.buf.InsertText(x, false, true)
		return nil
	}
}

// OptionCompletionWordSeparator to set word separators. Enable only ' ' if empty.
func OptionCompletionWordSeparator(x string) Option {
	return func(p *Prompt) error {
		p.completion.wordSeparator = x
		return nil
	}
}

// OptionCompletionExpandDescriptions to enable suggestion description expansion.
func OptionCompletionExpandDescriptions(x bool) Option {
	return func(p *Prompt) error {
		p.completion.expandDescriptions = x
		return nil
	}
}

// OptionLivePrefix to change the prefix dynamically by callback function.
func OptionLivePrefix(f func(doc *Document, isBreak bool) (prefix string, useLivePrefix bool)) Option {
	return func(p *Prompt) error {
		p.renderer.livePrefixCallback = f
		return nil
	}
}

// OptionSanitizer sets a callback function to be called before executing the command. It may update the document
// before execution.
func OptionSanitizer(f func(d Document) Document) Option {
	return func(p *Prompt) error {
		p.sanitizer = f
		return nil
	}
}

// OptionPrefixTextColor change a text color of prefix string.
func OptionPrefixTextColor(x Color) Option {
	return func(p *Prompt) error {
		p.renderer.prefixTextColor = x
		return nil
	}
}

// OptionPrefixBackgroundColor to change a background color of prefix string.
func OptionPrefixBackgroundColor(x Color) Option {
	return func(p *Prompt) error {
		p.renderer.prefixBGColor = x
		return nil
	}
}

// OptionInputTextColor to change a color of text which is input by user.
func OptionInputTextColor(x Color) Option {
	return func(p *Prompt) error {
		p.renderer.inputTextColor = x
		return nil
	}
}

// OptionInputBGColor to change a color of background which is input by user.
func OptionInputBGColor(x Color) Option {
	return func(p *Prompt) error {
		p.renderer.inputBGColor = x
		return nil
	}
}

// OptionPreviewSuggestionTextColor to change a text color which is completed.
func OptionPreviewSuggestionTextColor(x Color) Option {
	return func(p *Prompt) error {
		p.renderer.previewSuggestionTextColor = x
		return nil
	}
}

// OptionPreviewSuggestionBGColor to change a background color which is completed.
func OptionPreviewSuggestionBGColor(x Color) Option {
	return func(p *Prompt) error {
		p.renderer.previewSuggestionBGColor = x
		return nil
	}
}

// OptionInlineTextColor to change a text color which is inlined at end of input.
func OptionInlineTextColor(x Color) Option {
	return func(p *Prompt) error {
		p.renderer.inlineTextColor = x
		return nil
	}
}

// OptionInlineBGColor to change a background color which is inlined at end of input.
func OptionInlineBGColor(x Color) Option {
	return func(p *Prompt) error {
		p.renderer.inlineBGColor = x
		return nil
	}
}

// OptionSuggestionTextColor to change a text color in drop down suggestions.
func OptionSuggestionTextColor(x Color) Option {
	return func(p *Prompt) error {
		p.renderer.suggestionTextColor = x
		return nil
	}
}

// OptionSuggestionBGColor change a background color in drop down suggestions.
func OptionSuggestionBGColor(x Color) Option {
	return func(p *Prompt) error {
		p.renderer.suggestionBGColor = x
		return nil
	}
}

// OptionSelectedSuggestionTextColor to change a text color for completed text which is selected inside suggestions drop down box.
func OptionSelectedSuggestionTextColor(x Color) Option {
	return func(p *Prompt) error {
		p.renderer.selectedSuggestionTextColor = x
		return nil
	}
}

// OptionSelectedSuggestionBGColor to change a background color for completed text which is selected inside suggestions drop down box.
func OptionSelectedSuggestionBGColor(x Color) Option {
	return func(p *Prompt) error {
		p.renderer.selectedSuggestionBGColor = x
		return nil
	}
}

// OptionDescriptionTextColor to change a background color of description text in drop down suggestions.
func OptionDescriptionTextColor(x Color) Option {
	return func(p *Prompt) error {
		p.renderer.descriptionTextColor = x
		return nil
	}
}

// OptionDescriptionBGColor to change a background color of description text in drop down suggestions.
func OptionDescriptionBGColor(x Color) Option {
	return func(p *Prompt) error {
		p.renderer.descriptionBGColor = x
		return nil
	}
}

// OptionSelectedDescriptionTextColor to change a text color of description which is selected inside suggestions drop down box.
func OptionSelectedDescriptionTextColor(x Color) Option {
	return func(p *Prompt) error {
		p.renderer.selectedDescriptionTextColor = x
		return nil
	}
}

// OptionSelectedDescriptionBGColor to change a background color of description which is selected inside suggestions drop down box.
func OptionSelectedDescriptionBGColor(x Color) Option {
	return func(p *Prompt) error {
		p.renderer.selectedDescriptionBGColor = x
		return nil
	}
}

// OptionSuggestTypeLabelTextColor OptionLabelTextColor to change a text color of description which is selected inside suggestions drop down box.
func OptionSuggestTypeLabelTextColor(x Color) Option {
	return func(p *Prompt) error {
		p.renderer.suggestTypeLabelTextColor = x
		return nil
	}
}

// OptionSuggestTypeLabelBGColor OptionLabelBGColor to change a background color of description which is selected inside suggestions drop down box.
func OptionSuggestTypeLabelBGColor(x Color) Option {
	return func(p *Prompt) error {
		p.renderer.suggestTypeLabelBGColor = x
		return nil
	}
}

// OptionScrollbarThumbColor to change a thumb color on scrollbar.
func OptionScrollbarThumbColor(x Color) Option {
	return func(p *Prompt) error {
		p.renderer.scrollbarThumbColor = x
		return nil
	}
}

// OptionScrollbarBGColor to change a background color of scrollbar.
func OptionScrollbarBGColor(x Color) Option {
	return func(p *Prompt) error {
		p.renderer.scrollbarBGColor = x
		return nil
	}
}

// OptionStatusBarTextColor sets the foreground color of the statusbar.
func OptionStatusBarTextColor(x Color) Option {
	return func(p *Prompt) error {
		p.renderer.statusBarTextColor = x
		return nil
	}
}

// OptionStatusBarBGColor OptionStatusBarTextColor sets the foreground color of the statusbar.
func OptionStatusBarBGColor(x Color) Option {
	return func(p *Prompt) error {
		p.renderer.statusBarBGColor = x
		return nil
	}
}

// OptionMaxSuggestion specify the max number of displayed suggestions.
func OptionMaxSuggestion(x uint16) Option {
	return func(p *Prompt) error {
		p.completion.max = x
		return nil
	}
}

// OptionHistory to set history expressed by string array.
func OptionHistory(x []string) Option {
	return func(p *Prompt) error {
		p.history.histories = x
		p.history.Clear()
		return nil
	}
}

// OptionSwitchKeyBindMode set a key bind mode.
func OptionSwitchKeyBindMode(m KeyBindMode) Option {
	return func(p *Prompt) error {
		p.keyBindMode = m
		return nil
	}
}

// OptionCompletionOnDown allows for Down arrow key to trigger completion.
func OptionCompletionOnDown() Option {
	return func(p *Prompt) error {
		p.completionOnDown = true
		return nil
	}
}

// SwitchKeyBindMode to set a key bind mode.
// Deprecated: Please use OptionSwitchKeyBindMode.
var SwitchKeyBindMode = OptionSwitchKeyBindMode

// OptionAddKeyBind to set a custom key bind.
func OptionAddKeyBind(b ...KeyBind) Option {
	return func(p *Prompt) error {
		p.keyBindings = append(p.keyBindings, b...)
		return nil
	}
}

// OptionAddASCIICodeBind to set a custom key bind.
func OptionAddASCIICodeBind(b ...ASCIICodeBind) Option {
	return func(p *Prompt) error {
		p.ASCIICodeBindings = append(p.ASCIICodeBindings, b...)
		return nil
	}
}

// OptionShowCompletionAtStart to set completion window is open at start.
func OptionShowCompletionAtStart() Option {
	return func(p *Prompt) error {
		p.completion.showAtStart = true
		return nil
	}
}

// OptionBreakLineCallback to run a callback at every break line.
func OptionBreakLineCallback(fn func(*Document)) Option {
	return func(p *Prompt) error {
		p.renderer.breakLineCallback = fn
		return nil
	}
}

// OptionCancelLineCallback to run a callback at every line cancellation (ctrl-c).
func OptionCancelLineCallback(fn func(*Document)) Option {
	return func(p *Prompt) error {
		p.cancelLineCallback = fn
		return nil
	}
}

// OptionSetExitCheckerOnInput set an exit function which checks if go-prompt exits its Run loop.
func OptionSetExitCheckerOnInput(fn ExitChecker) Option {
	return func(p *Prompt) error {
		p.exitChecker = fn
		return nil
	}
}

// OptionSetLexer set lexer function and enable it.
func OptionSetLexer(fn LexerFunc) Option {
	return func(p *Prompt) error {
		p.lexer = &Lexer{fn}
		return nil
	}
}

// OptionTtyFallbackErrors list of error strings to consider when falling back when opening tty fd fails.
func OptionTtyFallbackErrors(errors []string) Option {
	return func(p *Prompt) error {
		ttyFallbackErrors = errors
		return nil
	}
}

func OptionCaptureRefreshTimings(enable bool) Option {
	return func(p *Prompt) error {
		p.captureRefreshTimings = enable
		return nil
	}
}

func OptionEnableRenderCaches(enable bool) Option {
	return func(p *Prompt) error {
		p.renderer.stringCaches = enable
		return nil
	}
}

func OptionEnableMarkup(enable bool) Option {
	return func(p *Prompt) error {
		p.renderer.enableMarkup = enable
		return nil
	}
}

// New returns a Prompt with powerful auto-completion.
func New(executor Executor, completer Completer, opts ...Option) *Prompt {
	defaultWriter := NewStdoutWriter()
	registerConsoleWriter(defaultWriter)

	pt := &Prompt{
		renderer: &Render{
			prefix:                       "> ",
			out:                          defaultWriter,
			livePrefixCallback:           func(*Document, bool) (string, bool) { return "", false },
			prefixTextColor:              Blue,
			prefixBGColor:                DefaultColor,
			inlineTextColor:              DarkGray,
			inlineBGColor:                DefaultColor,
			inputTextColor:               DefaultColor,
			inputBGColor:                 DefaultColor,
			previewSuggestionTextColor:   Green,
			previewSuggestionBGColor:     DefaultColor,
			suggestionTextColor:          White,
			suggestionBGColor:            Cyan,
			selectedSuggestionTextColor:  Black,
			selectedSuggestionBGColor:    Turquoise,
			descriptionTextColor:         Black,
			descriptionBGColor:           Turquoise,
			selectedDescriptionTextColor: White,
			selectedDescriptionBGColor:   Cyan,
			scrollbarThumbColor:          DarkGray,
			scrollbarBGColor:             Cyan,
			suggestTypeLabelTextColor:    White,
			suggestTypeLabelBGColor:      DarkGray,
		},
		buf:            NewBuffer(),
		executor:       executor,
		history:        NewHistory(),
		completion:     NewCompletionManager(completer, 6),
		keyBindMode:    DefaultKeyBind,
		refreshCh:      make(chan Refresh, 1000),
		refreshTimings: []float64{},
	}

	for _, opt := range opts {
		if err := opt(pt); err != nil {
			panic(err)
		}
	}

	if pt.lexer == nil {
		pt.lexer = NewLexer(pt.renderer.inputTextColor, pt.renderer.inputBGColor)
	}

	pt.in = NewStandardInputParser()
	pt.r = NewStandardInputReader(pt.in)

	return pt
}
