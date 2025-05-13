package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// docker run 			<image> <command> <args>
// go run main.go 					<command> <args>

func main() {
	switch os.Args[1] {
	case "run":
		run()
	default:
		panic("found wrong argument")
	}
}

func run() {
	fmt.Println("Running:", os.Args[2:])

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS,
	}
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}
