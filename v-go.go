// This is a dead simple wrapper that can have setuid set on it so that
// the sudo helper is run as root. It is expected to run in the same CWD
// as the actual Ruby sudo helper. Any arguments to this script are forwarded
// to the Ruby sudo helper script.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

const sudoHelperCommand = "vagrant_vmware_desktop_sudo_helper"

func main() {
	// We need to be running as root to properly setuid and execute
	// the sudo helper as root.
	if os.Geteuid() != 0 {
		fmt.Fprintf(os.Stderr, "sudo helper setuid-wrapper must run as root.\n")
		os.Exit(1)
	}

	// We have to setuid here because suid bits only change the EUID,
	// and when Ruby sees the EUID != RUID, it resets the EUID back to RUID,
	// nullifying the change we actually want. This forces the script to
	// run as root.
	if err := syscall.Setuid(0); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	// Put together the complete path to the helper program.
	var helperPath = fmt.Sprintf("%s/%s", filepath.Dir(os.Args[0]), sudoHelperCommand)

	// Setup the argv array so that the arguments are preserved for the
	// helper, while also adding our arguments so the helper is properly called.
	var helperArgs = make([]string, len(os.Args)+1)
	helperArgs[0] = "ruby"
	helperArgs[1] = helperPath
	copy(helperArgs[2:], os.Args[1:])

	// Replace ourselves with the actual helper. This should terminate
	// execution of this program, but if not we output and exit.
	if err := syscall.Exec("ruby", helperArgs, os.Environ()); err != nil {
		fmt.Fprintf(os.Stderr, "Exec error: %s\n", err.Error())
		os.Exit(1)
	}

	panic("Reached unreachable code!")
}
Raw
