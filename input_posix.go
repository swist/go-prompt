//go:build !windows

package prompt

import (
	"os"
	"syscall"

	"github.com/c-bata/go-prompt/internal/strings"
	"github.com/c-bata/go-prompt/internal/term"
	"golang.org/x/sys/unix"
)

const maxReadBytes = 4096

// PosixParser is a ConsoleParser implementation for POSIX environment.
type PosixParser struct {
	fd int
}

// Setup should be called before starting input
func (t *PosixParser) Setup() error {
	// Set NonBlocking mode because if syscall.Read block this goroutine, it cannot receive data from stopCh.
	// Commented out as part of "<fix-sleeping-read> 667c9a8f0872726df999abd6c37cc9d6300ab630"
	//if err := syscall.SetNonblock(t.fd, true); err != nil {
	//	return err
	//}
	if err := term.SetRaw(t.fd); err != nil {
		return err
	}
	return nil
}

// TearDown should be called after stopping input
func (t *PosixParser) TearDown() error {
	if err := syscall.SetNonblock(t.fd, false); err != nil {
		return err
	}
	if err := term.Restore(); err != nil {
		return err
	}
	return nil
}

func (t *PosixParser) Destroy() error {
	return syscall.Close(t.fd)
}

// Read returns byte array.
func (t *PosixParser) Read() ([]byte, error) {
	buf := make([]byte, maxReadBytes)
	n, err := syscall.Read(t.fd, buf)
	if err != nil {
		return []byte{}, err
	}
	return buf[:n], nil
}

// GetWinSize returns WinSize object to represent width and height of terminal.
func (t *PosixParser) GetWinSize() *WinSize {
	ws, err := unix.IoctlGetWinsize(t.fd, unix.TIOCGWINSZ)
	if err != nil {
		// If this errors, we simply return the default window size as
		// it's our best guess.
		return &WinSize{
			Row: 25,
			Col: 80,
		}
	}
	return &WinSize{
		Row: ws.Row,
		Col: ws.Col,
	}
}

var _ ConsoleParser = &PosixParser{}

// NewStandardInputParser returns ConsoleParser object to read from stdin.
func NewStandardInputParser() *PosixParser {
	in, err := syscall.Open("/dev/tty", syscall.O_RDONLY, 0)
	if err != nil {
		if os.IsNotExist(err) || strings.Contains(ttyFallbackErrors, err.Error()) {
			in = syscall.Stdin
		} else if err != nil {
			panic(err)
		}
	}

	return &PosixParser{
		fd: in,
	}
}
