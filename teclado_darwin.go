//go:build darwin

package main

import "syscall"

const (
	tcLeer     = syscall.TIOCGETA
	tcEscribir = syscall.TIOCSETA
)
