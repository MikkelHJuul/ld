package main

import (
	"github.com/desertbit/grumble"
	"os"
	"os/exec"
)

func main() {
	_, errFindingLd := os.Stat("/ld")
	_, isRunning := os.Stat("/ld-is-running")
	var ldCommand *exec.Cmd
	if errFindingLd == nil && isRunning != nil {
		_, err := os.Create("/ld-is-running")
		if err != nil {
			panic("could not create run-file /ld-is-running")
		}
		ldCommand = exec.Command("/ld")
		err = ldCommand.Start()
		if err != nil {
			panic(err)
		}
	}
	grumble.Main(app)
	if ldCommand != nil {
		err := ldCommand.Wait()
		if err != nil {
			panic(err)
		}
	}
}
