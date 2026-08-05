//go:build linux || darwin

package main

// En Mac y en Linux la terminal viene en "modo linea": guarda lo que tecleas
// y no se lo da al programa hasta que pulsas Enter. Para leer teclas sueltas
// hay que quitarle esa costumbre, y devolversela al terminar.

import (
	"os"
	"strings"
	"syscall"
	"unsafe"
)

// Como estaba la terminal antes de que la tocaramos. Si es nil, es que
// todavia no la hemos tocado.
var estadoPrevio *syscall.Termios

func ioctl(fd int, orden uintptr, t *syscall.Termios) bool {
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), orden,
		uintptr(unsafe.Pointer(t)))
	return e == 0
}

func prepararTeclado() bool {
	if estadoPrevio != nil {
		return true
	}
	fd := int(os.Stdin.Fd())
	var t syscall.Termios
	if !ioctl(fd, tcLeer, &t) {
		return false // no es una terminal: una tuberia, o un archivo
	}
	previo := t
	nuevo := t
	// ICANON es lo de esperar a la linea entera; ECHO es que se vea lo que
	// tecleas, que en un juego solo ensucia la pantalla.
	nuevo.Lflag &^= syscall.ICANON | syscall.ECHO
	// VMIN 0 y VTIME 0 significan "dame lo que haya y no esperes".
	nuevo.Cc[syscall.VMIN] = 0
	nuevo.Cc[syscall.VTIME] = 0
	if !ioctl(fd, tcEscribir, &nuevo) {
		return false
	}
	estadoPrevio = &previo
	return true
}

func soltarTeclado() {
	if estadoPrevio == nil {
		return
	}
	ioctl(int(os.Stdin.Fd()), tcEscribir, estadoPrevio)
	estadoPrevio = nil
}

func leerTecla() string {
	if !prepararTeclado() {
		return ""
	}
	var buf [8]byte
	n, err := os.Stdin.Read(buf[:])
	if err != nil || n == 0 {
		return ""
	}
	return nombreTecla(buf[:n])
}

func nombreTecla(b []byte) string {
	// Las flechas llegan como tres bytes: escape, corchete y una letra.
	if len(b) >= 3 && b[0] == 27 && b[1] == '[' {
		switch b[2] {
		case 'A':
			return teclaArriba
		case 'B':
			return teclaAbajo
		case 'C':
			return teclaDerecha
		case 'D':
			return teclaIzquierda
		}
		return ""
	}
	switch b[0] {
	case 27:
		return teclaEscape
	case 13, 10:
		return teclaIntro
	case 32:
		return teclaEspacio
	}
	if b[0] < 32 {
		return ""
	}
	return strings.ToLower(string(rune(b[0])))
}
