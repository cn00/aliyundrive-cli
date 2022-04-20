//go:build darwin
// +build darwin

package internal

import (
	"os"

	"golang.org/x/sys/unix"
)

type Termios struct {
	term *unix.Termios
}

var term *Termios

func GetTermios() *Termios {
	return term
}

func IoctlGetTermios() *Termios {
	termios, _ := unix.IoctlGetTermios(int(os.Stdin.Fd()), 0 /*unix.TCGETS*/)
	term = &Termios{term: termios}
	return term
}

func IoctlSetTermios(termios *Termios) {
	if termios != nil && termios.term != nil {
		unix.IoctlSetTermios(int(os.Stdin.Fd()), 0 /*unix.TCSETS*/, termios.term)
	}
}
