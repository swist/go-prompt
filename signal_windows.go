//go:build windows

package prompt

import (
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

			case os.Interrupt: // Ctrl+C
				debug.Log("Catch Ctrl+C")
				exitCh <- NativeInterrupt

			case syscall.SIGINT: // kill -SIGINT XXXX
				debug.Log("Catch SIGINT")
				exitCh <- 0

			case syscall.SIGTERM: // kill -SIGTERM XXXX
				debug.Log("Catch SIGTERM")
				exitCh <- 1

			case syscall.SIGQUIT: // kill -SIGQUIT XXXX
				debug.Log("Catch SIGQUIT")
				exitCh <- 0
			}
		}
	}
}
