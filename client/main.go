package main

import (
	"github.com/desertbit/grumble"
	"os"
	"os/exec"
)

func main() {
	_, errLd := os.Stat("/ld")
	_, isRunning := os.Stat("/ld-is-running")
	shouldRunAndSleep := errLd == nil && isRunning != nil
	if shouldRunAndSleep {
		_, _ = os.Create("/ld-is-running")
		ldArgs := os.Getenv("LD_ARGS")
		cmnd := exec.Command("/ld", ldArgs)
		cmnd.Start()
	}
	grumble.Main(app)
	if shouldRunAndSleep {
		select {} //sleep indefinitely (/ld is running) this process is PID 1 so container must have it running!
	}
}
