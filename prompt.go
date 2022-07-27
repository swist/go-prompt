package prompt

import (
	"bufio"
	"bytes"
	"os"
	"regexp"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/c-bata/go-prompt/internal/debug"
)

// Sanitizer is called when user input something text, but before calling breakline/executor.  May be used to
// modify the input text before committing to history, etc.
type Sanitizer func(d Document) Document

// Executor is called when user input something text.
type Executor func(string)

// ExitChecker is called after user input to check if prompt must stop and exit go-prompt Run loop.
// User input means: selecting/typing an entry, then, if said entry content matches the ExitChecker function criteria:
// - immediate exit (if breakline is false) without executor called
// - exit after typing <return> (meaning breakline is true), and the executor is called first, before exit.
// Exit means exit go-prompt (not the overall Go program)
type ExitChecker func(in string, breakline bool) bool

type Refresh struct {
	Options   RefreshOptions
	StatusBar StatusBar
}

// RefreshChecker is called to determine whether or not to refresh the prompt.
type RefreshChecker func(d *Document) bool

// Completer should return the suggest item from Document.
type Completer func(Document) []Suggest

// Prompt is core struct of go-prompt.
type Prompt struct {
	in                    ConsoleParser
	r                     *bufio.Reader
	buf                   *Buffer
	renderer              *Render
	sanitizer             Sanitizer
	executor              Executor
	history               *History
	lexer                 *Lexer
	completion            *CompletionManager
	keyBindings           []KeyBind
	ASCIICodeBindings     []ASCIICodeBind
	keyBindMode           KeyBindMode
	completionOnDown      bool
	exitChecker           ExitChecker
	skipTearDown          bool
	lastBytes             []byte
	cancelLineCallback    func(*Document)
	captureRefreshTimings bool
	refreshCh             chan Refresh
	refreshTimings        []float64
}

// Exec is the struct contains user input context.
type Exec struct {
	input string
}

func (p *Prompt) MustDestroy() {
	if err := p.Destroy(); err != nil {
		panic(err)
	}
}

func (p *Prompt) Destroy() error {
	return p.in.Destroy()
}

func (p *Prompt) Buffer(line string) {
	p.buf = NewBufferWithLine(line)
}

type RefreshOptions uint8

const (
	RefreshUpdate RefreshOptions = 1 << iota
	RefreshRender
	RefreshStatusBar
	RefreshAll  RefreshOptions = 0xFF
	RefreshNone RefreshOptions = 0x00
)

func (p *Prompt) refresh(options RefreshOptions) {
	if p.captureRefreshTimings {
		start := time.Now()
		defer func() { p.refreshTimings = append(p.refreshTimings, float64(time.Since(start).Nanoseconds())) }()
	}
	if options == RefreshAll {
		p.renderer.out.EraseDown()
	}
	if options&RefreshUpdate == RefreshUpdate {
		p.completion.Update(*p.buf.Document())
		p.renderer.Render(p.buf, p.completion, p.lexer)
	} else if options&RefreshRender == RefreshRender {
		p.renderer.Render(p.buf, p.completion, p.lexer)
	}
	if options&RefreshStatusBar == RefreshStatusBar {
		p.renderer.RenderStatusBar()
	}
}

func (p *Prompt) RefreshCh() chan<- Refresh {
	return p.refreshCh
}

// RefreshTimings returns the timings of each refresh.
func (p *Prompt) RefreshTimings() []float64 {
	return p.refreshTimings
}

// Run starts prompt.
func (p *Prompt) Run() {
	p.skipTearDown = false
	debug.Setup()
	defer debug.Teardown()
	debug.Log("Run start prompt")
	p.setUp()
	defer p.tearDown()

	opts := RefreshRender | RefreshStatusBar
	if p.completion.showAtStart {
		opts |= RefreshUpdate
	}
	p.refresh(opts)

	bufCh := make(chan []byte)
	go p.readLine(bufCh)

	exitCh := make(chan int)
	winSizeCh := make(chan *WinSize)
	stopHandleSignalCh := make(chan struct{})
	go p.handleSignals(exitCh, winSizeCh, stopHandleSignalCh)

	for {
		select {
		case b := <-bufCh:
			for {
				shouldExit, e, bufLeft, ended := p.feed(b)
				if shouldExit {
					p.renderer.BreakLine(p.buf, p.lexer)
					stopHandleSignalCh <- struct{}{}
					return
				} else if e != nil {
					stopHandleSignalCh <- struct{}{}

					// Unset raw mode
					// Reset to Blocking mode because returned EAGAIN when still set non-blocking mode.
					debug.AssertNoError(p.in.TearDown())

					p.executor(e.input)

					if p.exitChecker != nil && p.exitChecker(e.input, true) {
						p.skipTearDown = true
						return
					}
					// Set raw mode
					debug.AssertNoError(p.in.Setup())
					go p.handleSignals(exitCh, winSizeCh, stopHandleSignalCh)
				} else {
					p.refresh(RefreshAll)
				}
				if ended {
					go p.readLine(bufCh)
				}
				if len(bufLeft) == 0 {
					break
				}
				b = bufLeft
			}
		case w := <-winSizeCh:
			p.renderer.UpdateWinSize(w)
			p.refresh(RefreshAll)
		case code := <-exitCh:
			p.renderer.BreakLine(p.buf, p.lexer)
			p.tearDown()
			os.Exit(code)
		case refresh := <-p.refreshCh:
			if refresh.Options&RefreshStatusBar == RefreshStatusBar {
				p.renderer.statusBar = refresh.StatusBar
			}
			p.refresh(refresh.Options)
		}
	}
}

func (p *Prompt) feed(buf []byte) (shouldExit bool, exec *Exec, bufLeft []byte, ended bool) {
	idx := bytes.IndexAny(buf, string([]byte{0xa, 0xd}))
	var b []byte
	if idx == 0 {
		b = []byte{buf[0]}
		if len(buf) > 1 {
			bufLeft = buf[1:]
		}
	} else if idx != -1 {
		b = buf[0:idx]
		bufLeft = buf[idx:]
	} else {
		b = buf
	}

	defer func() { p.lastBytes = b }()

	key := GetKey(b)
	p.buf.lastKeyStroke = key
	// completion
	completing := p.completion.Completing()
	p.handleCompletionKeyBinding(key, completing)

	switch key {
	case Enter, ControlJ, ControlM:
		// TODO: update to match on odd numbers of trailing backslashes: `(?:^|[^\\])\\(\\\\)*$`
		match := regexp.MustCompile(`(\\+)$`).FindStringSubmatch(p.buf.Text())
		if len(match) == 2 {
			p.buf.DeleteBeforeCursor(len(match[1]))
		} else {
			if p.sanitizer != nil {
				d := p.sanitizer(*p.buf.Document())
				p.buf.setDocument(&d)
			}
			p.renderer.BreakLine(p.buf, p.lexer)
			exec = &Exec{input: p.buf.Text()}
			p.buf = NewBuffer()
			if exec.input != "" {
				p.history.Add(exec.input)
			}
		}
	case ControlC:
		if p.cancelLineCallback != nil {
			p.cancelLineCallback(p.buf.Document())
		}
		p.renderer.BreakLine(p.buf, p.lexer)
		p.buf = NewBuffer()
		p.history.Clear()
	case Up, ControlP:
		if !completing { // Don't use p.completion.Completing() because it takes double operation when switch to selected=-1.
			if newBuf, changed := p.history.Older(p.buf); changed {
				p.buf = newBuf
			}
		}
	case Down, ControlN:
		if !completing { // Don't use p.completion.Completing() because it takes double operation when switch to selected=-1.
			if newBuf, changed := p.history.Newer(p.buf); changed {
				p.buf = newBuf
			}
			return
		}
	case ControlD:
		if p.buf.Text() == "" {
			shouldExit = true
			return
		}
	case ControlS:
		ended = true
		return
	case NotDefined:
		if p.handleASCIICodeBinding(b) {
			return
		}
		// Only insert printable characters
		r, _ := utf8.DecodeRune(b)
		if strconv.IsPrint(r) {
			p.buf.InsertText(string(b), false, true)
		}
	}

	shouldExit = p.handleKeyBinding(key)
	return
}

func (p *Prompt) handleCompletionKeyBinding(key Key, completing bool) {
	switch key {
	case Down:
		if completing || p.completionOnDown {
			p.completion.Next()
		}
	case Tab, ControlI:
		p.completion.Next()
	case Up:
		if completing {
			p.completion.Previous()
		}
	case BackTab:
		p.completion.Previous()
	default:
		if s, ok := p.completion.GetSelectedSuggestion(); ok {
			w := p.buf.Document().GetWordBeforeCursorUntilSeparator(p.completion.wordSeparator)
			if w != "" {
				p.buf.DeleteBeforeCursor(len([]rune(w)))
			}
			s = s.committed()
			p.buf.InsertText(s.Text, false, true)
		}
		p.completion.Reset()
	}
}

func (p *Prompt) handleKeyBinding(key Key) bool {
	shouldExit := false
	for i := range commonKeyBindings {
		kb := commonKeyBindings[i]
		if kb.Key == key {
			kb.Fn(p.buf)
		}
	}

	if p.keyBindMode == DefaultKeyBind {
		for _, kb := range defaultKeyBindings() {
			if kb.Key == key {
				kb.Fn(p.buf)
			}
		}
	}

	// Custom key bindings
	for i := range p.keyBindings {
		kb := p.keyBindings[i]
		if kb.Key == key {
			kb.Fn(p.buf)
		}
	}
	if p.exitChecker != nil && p.exitChecker(p.buf.Text(), false) {
		shouldExit = true
	}
	return shouldExit
}

func (p *Prompt) handleASCIICodeBinding(b []byte) bool {
	checked := false
	for _, kb := range p.ASCIICodeBindings {
		if bytes.Equal(kb.ASCIICode, b) {
			kb.Fn(p.buf)
			checked = true
		}
	}
	return checked
}

// Input just returns user input text.
func (p *Prompt) Input() string {
	debug.Setup()
	defer debug.Teardown()
	debug.Log("Input start prompt")
	p.setUp()
	defer p.tearDown()

	opts := RefreshRender | RefreshStatusBar
	if p.completion.showAtStart {
		opts |= RefreshUpdate
	}
	p.refresh(opts)

	bufCh := make(chan []byte)
	go p.readLine(bufCh)

	result := ""
	for b := range bufCh {
		if shouldExit, e, _, _ := p.feed(b); shouldExit {
			p.renderer.BreakLine(p.buf, p.lexer)
			break
		} else if e != nil {
			result = e.input
		} else {
			p.refresh(RefreshAll)
		}
	}
	return result
}

func (p *Prompt) readChunk() ([]byte, error) {
	b := make([]byte, maxReadBytes)
	n, err := p.r.Read(b)
	return b[:n], err
}

func (p *Prompt) readLine(bufCh chan []byte) {
	debug.Log("start reading input until end of line")

	var last byte
	for read := true; read; {
		if chunk, err := p.readChunk(); err == nil && !(len(chunk) == 1 && chunk[0] == 0) {
			start := 0
			for i, b := range chunk {
				key := GetKey([]byte{b})
				isBreak := key == Enter || key == ControlJ || key == ControlM
				if isBreak && last != '\\' {
					bufCh <- chunk[start:i]
					bufCh <- []byte{b}
					start = i + 1
					read = false
				}
				last = b
			}
			bufCh <- chunk[start:]
		}
	}

	bufCh <- GetCode(ControlS)
	debug.Log("stopped reading input after reading end of line")
}

func (p *Prompt) setUp() {
	debug.AssertNoError(p.in.Setup())
	p.renderer.Setup()
	p.renderer.UpdateWinSize(p.in.GetWinSize())
}

func (p *Prompt) tearDown() {
	if !p.skipTearDown {
		debug.AssertNoError(p.in.TearDown())
	}
	p.renderer.TearDown()
}
