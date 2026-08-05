//go:build linux

package main

import "syscall"

// Cada sistema llama de una forma a "leeme como esta la terminal" y
// "dejala asi". Solo cambian estos dos numeros.
const (
	tcLeer     = syscall.TCGETS
	tcEscribir = syscall.TCSETS
)
