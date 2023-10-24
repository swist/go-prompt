//go:build windows

package prompt

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/c-bata/go-prompt/internal/debug"
)

func (p *Prompt) handleSignals(exitCh chan int, winSizeCh chan *WinSize, stop chan struct{}) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(
		sigCh,
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	for {
		select {
		case <-stop:
			debug.Log("stop handleSignals")
			return
		case s := <-sigCh:
			switch s {
			case os.Interrupt: // Ctrl+C, must handle first
				// Not needed when console raw mode is enabled, but will leave it in case raw mode setup fails
				debug.Log("Interrupt")
				exitCh <- int(ControlC)
			case syscall.SIGINT:
				debug.Log("SIGINT")
				exitCh <- 0
			case syscall.SIGTERM:
				debug.Log("SIGTERM")
				exitCh <- 1
			case syscall.SIGQUIT:
				debug.Log("SIGQUIT")
				exitCh <- 0
			default:
				debug.Log(fmt.Sprintf("Unhandled signal: %v", s))
			}
		}
	}
}
