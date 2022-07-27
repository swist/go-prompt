//go:build windows

package prompt

import (
	"errors"
	"fmt"
	"syscall"
	"unicode/utf8"
	"unsafe"

	"github.com/c-bata/go-prompt/internal/debug"
	tty "github.com/mattn/go-tty"
)

const maxReadBytes = 4096

var kernel32 = syscall.NewLazyDLL("kernel32.dll")

var procGetNumberOfConsoleInputEvents = kernel32.NewProc("GetNumberOfConsoleInputEvents")

// WindowsParser is a ConsoleParser implementation for Win32 console.
type WindowsParser struct {
	tty *tty.TTY
}

// Setup should be called before starting input
func (p *WindowsParser) Setup() error {
	t, err := tty.Open()
	if err != nil {
		return err
	}
	p.tty = t
	return nil
}

// TearDown should be called after stopping input
func (p *WindowsParser) TearDown() error {
	return nil
}

func (p *WindowsParser) Destroy() error {
	// TODO: investigate the root cause behind these panics
	defer func() {
		if r := recover(); r != nil {
			debug.Log(fmt.Sprintf("recovered panic closing go-tty channel: %v", r))
		}
	}()
	return p.tty.Close()
}

// Read returns byte array.
func (p *WindowsParser) Read() ([]byte, error) {
	var ev uint32
	r0, _, err := procGetNumberOfConsoleInputEvents.Call(p.tty.Input().Fd(), uintptr(unsafe.Pointer(&ev)))
	if r0 == 0 {
		return nil, err
	}
	if ev == 0 {
		return nil, errors.New("EAGAIN")
	}

	r, err := p.tty.ReadRune()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, maxReadBytes)
	n := utf8.EncodeRune(buf[:], r)
	for p.tty.Buffered() && n < maxReadBytes {
		r, err := p.tty.ReadRune()
		if err != nil {
			break
		}
		n += utf8.EncodeRune(buf[n:], r)
	}
	return buf[:n], nil
}

// GetWinSize returns WinSize object to represent width and height of terminal.
func (p *WindowsParser) GetWinSize() *WinSize {
	w, h, err := p.tty.Size()
	if err != nil {
		// If this errors, we simply return the default window size as
		// it's our best guess.
		return &WinSize{
			Row: 25,
			Col: 80,
		}
	}
	return &WinSize{
		Row: uint16(h),
		Col: uint16(w),
	}
}

// NewStandardInputParser returns ConsoleParser object to read from stdin.
func NewStandardInputParser() *WindowsParser {
	return &WindowsParser{}
}
