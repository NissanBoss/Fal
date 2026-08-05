//go:build windows

package main

// En Windows la consola ya sabe dar teclas sueltas sin esperar al Enter:
// _kbhit dice si hay alguna esperando y _getch la recoge.

import (
	"strings"
	"syscall"
)

var (
	msvcrt        = syscall.NewLazyDLL("msvcrt.dll")
	hayTeclaProc  = msvcrt.NewProc("_kbhit")
	cogeTeclaProc = msvcrt.NewProc("_getch")
)

func soltarTeclado() {}

func leerTecla() string {
	hay, _, _ := hayTeclaProc.Call()
	if hay == 0 {
		return ""
	}
	c, _, _ := cogeTeclaProc.Call()
	b := byte(c)

	// Las flechas no son un caracter: llegan en dos golpes, primero un aviso
	// (0 o 224) y luego el codigo de verdad.
	if b == 0 || b == 224 {
		segunda, _, _ := cogeTeclaProc.Call()
		switch byte(segunda) {
		case 72:
			return teclaArriba
		case 80:
			return teclaAbajo
		case 75:
			return teclaIzquierda
		case 77:
			return teclaDerecha
		}
		return ""
	}

	switch b {
	case 27:
		return teclaEscape
	case 13, 10:
		return teclaIntro
	case 32:
		return teclaEspacio
	}
	if b < 32 {
		return ""
	}
	return strings.ToLower(string(rune(b)))
}
