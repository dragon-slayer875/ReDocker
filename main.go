package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

// Entry point for the container.
// This sets up the container environment and spawns the process that runs the command inside the container.
func run() {
	fmt.Println("Running:", os.Args[2:], "as", os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

// Entry point for the child process.
// This actually runs the command inside the container.
func child() {
	fmt.Println("Running child:", os.Args[2:], "as", os.Getpid())

	cg()

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

// Setup cgroups for the container.
func cg() {
	// Determine which cgroup version we're using
	cgroupV2 := "/sys/fs/cgroup/cgroup.controllers"
	_, err := os.Stat(cgroupV2)
	isV2 := err == nil

	if isV2 {
		// Cgroups v2 (unified hierarchy)
		cgroups := "/sys/fs/cgroup/rudr"
		err := os.MkdirAll(cgroups, 0755)
		if err != nil && !os.IsExist(err) {
			panic(err)
		}

		// Enable pids controller for this cgroup
		must(writeFile("/sys/fs/cgroup/cgroup.subtree_control", "+pids"))

		// Set max pids
		must(writeFile(filepath.Join(cgroups, "pids.max"), "20"))

		// Add current process to the cgroup
		must(writeFile(filepath.Join(cgroups, "cgroup.procs"), strconv.Itoa(os.Getpid())))

	} else {
		// Cgroups v1
		cgroups := "/sys/fs/cgroup/"
		pids := filepath.Join(cgroups, "pids")
		err := os.MkdirAll(filepath.Join(pids, "rudr"), 0755)
		if err != nil && !os.IsExist(err) {
			panic(err)
		}

		// Set max pids
		must(writeFile(filepath.Join(pids, "rudr/pids.max"), "20"))

		// The notify_on_release doesn't exist in many modern systems
		notifyPath := filepath.Join(pids, "rudr/notify_on_release")
		if _, err := os.Stat(notifyPath); err == nil {
			must(writeFile(notifyPath, "1"))
		}

		// Add current process to the cgroup
		must(writeFile(filepath.Join(pids, "rudr/cgroup.procs"), strconv.Itoa(os.Getpid())))
	}
}

// Helper function to write to cgroup files
func writeFile(path, content string) error {
	// Open the file for writing, don't use create flags
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
