// +build !windows

package main

import (
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/rs/zerolog/log"
)

// Start assigns a pseudo-terminal tty os.File to c.Stdin, c.Stdout,
// and c.Stderr, calls c.Start, and returns the File of the tty's
// corresponding pty.
func PtyRun(c *exec.Cmd, tty *os.File) (err error) {
	defer tty.Close()
	c.Stdout = tty
	c.Stdin = tty
	c.Stderr = tty
	c.SysProcAttr = &syscall.SysProcAttr{
		Setctty: true,
		Setsid:  true,
	}
	return c.Start()
}

// SetWinSize sets the size of the given pty.
func SetWinSize(fd uintptr, w, h uint32) {
	log.Printf("window resize %dx%d", w, h)
	ws := &WinSize{Width: uint16(w), Height: uint16(h)}
	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
}
