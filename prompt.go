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

// Executor is called when user input something text.
type Executor func(string)

// ExitChecker is called after user input to check if prompt must stop and exit go-prompt Run loop.
// User input means: selecting/typing an entry, then, if said entry content matches the ExitChecker function criteria:
// - immediate exit (if breakline is false) without executor called
// - exit after typing <return> (meaning breakline is true), and the executor is called first, before exit.
// Exit means exit go-prompt (not the overall Go program)
type ExitChecker func(in string, breakline bool) bool

// RefreshChecker is called to determine whether or not to refresh the prompt.
type RefreshChecker func(d *Document) bool

// Completer should return the suggest item from Document.
type Completer func(Document) []Suggest

// Prompt is core struct of go-prompt.
type Prompt struct {
	in                 ConsoleParser
	r                  *bufio.Reader
	buf                *Buffer
	renderer           *Render
	executor           Executor
	history            *History
	lexer              *Lexer
	completion         *CompletionManager
	keyBindings        []KeyBind
	ASCIICodeBindings  []ASCIICodeBind
	keyBindMode        KeyBindMode
	completionOnDown   bool
	exitChecker        ExitChecker
	skipTearDown       bool
	lastBytes          []byte
	refreshTicker      *time.Ticker
	refreshChecker     RefreshChecker
	cancelLineCallback func(*Document)
}

// Exec is the struct contains user input context.
type Exec struct {
	input string
}

func (p *Prompt) Buffer(line string) {
	p.buf = NewBufferWithLine(line)
}

func (p *Prompt) Refresh(update bool) {
	if update {
		p.completion.Update(*p.buf.Document())
	}
	p.renderer.Render(p.buf, p.completion, p.lexer)
}

// Run starts prompt.
func (p *Prompt) Run() {
	p.skipTearDown = false
	defer debug.Teardown()
	debug.Log("Run start prompt")
	p.setUp()
	defer p.tearDown()

	p.Refresh(p.completion.showAtStart)

	bufCh := make(chan []byte)
	go p.readLine(bufCh)

	exitCh := make(chan int)
	winSizeCh := make(chan *WinSize)
	stopHandleSignalCh := make(chan struct{})
	go p.handleSignals(exitCh, winSizeCh, stopHandleSignalCh)

	var tickCh <-chan time.Time
	if p.refreshTicker != nil {
		tickCh = p.refreshTicker.C
	}

	nextRefresh := false

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
					p.Refresh(true)

					if p.exitChecker != nil && p.exitChecker(e.input, true) {
						p.skipTearDown = true
						return
					}
					// Set raw mode
					debug.AssertNoError(p.in.Setup())
					go p.handleSignals(exitCh, winSizeCh, stopHandleSignalCh)
				} else {
					p.Refresh(true)
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
			p.Refresh(false)
		case code := <-exitCh:
			p.renderer.BreakLine(p.buf, p.lexer)
			p.tearDown()
			os.Exit(code)
		case <-tickCh:
			if p.refreshChecker == nil {
				p.Refresh(true)
			} else {
				// Refreshes one additional time after p.refreshChecker returns false following a true check.
				prevRefresh := nextRefresh
				nextRefresh = p.refreshChecker(p.buf.Document())
				if nextRefresh || prevRefresh {
					p.Refresh(true)
				}
			}
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
		match := regexp.MustCompile(`(\\+)$`).FindStringSubmatch(p.buf.Text())
		if len(match) == 2 {
			p.buf.DeleteBeforeCursor(len(match[1]))
		} else {
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

	if p.keyBindMode == EmacsKeyBind {
		for i := range emacsKeyBindings {
			kb := emacsKeyBindings[i]
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
	defer debug.Teardown()
	debug.Log("Input start prompt")
	p.setUp()
	defer p.tearDown()

	p.Refresh(p.completion.showAtStart)

	bufCh := make(chan []byte)
	go p.readLine(bufCh)

	for {
		select {
		case b := <-bufCh:
			if shouldExit, e, _, _ := p.feed(b); shouldExit {
				p.renderer.BreakLine(p.buf, p.lexer)
				return ""
			} else if e != nil {
				return e.input
			} else {
				p.Refresh(true)
			}
		}
	}
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
	if p.refreshTicker != nil {
		p.refreshTicker.Stop()
	}
}
