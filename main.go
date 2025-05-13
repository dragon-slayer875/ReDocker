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
	case "child":
		child()
	default:
		panic("found wrong argument")
	}
}

func run() {
	fmt.Println("Running:", os.Args[2:], "as", os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
	}
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func child() {
	fmt.Println("Running child:", os.Args[2:], "as", os.Getpid())

	syscall.Sethostname([]byte("container"))
	syscall.Chroot("/")
	syscall.Chdir("/")
	syscall.Mount("proc", "proc", "proc", 0, "")

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	syscall.Unmount("proc", 0)
}
