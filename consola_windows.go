//go:build windows

package main

// En Windows hay que pedirle permiso a la consola antes de poder mover el
// cursor o pintar en colores. Windows Terminal (el de Windows 11) ya viene
// preparado; las ventanas antiguas de cmd.exe hay que despertarlas.

import (
	"syscall"
	"unsafe"
)

// ENABLE_VIRTUAL_TERMINAL_PROCESSING
const habilitarSecuencias = 0x0004

func prepararConsola() bool {
	manejador, err := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	if err != nil {
		return false
	}
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	obtenerModo := kernel32.NewProc("GetConsoleMode")
	ponerModo := kernel32.NewProc("SetConsoleMode")

	var modo uint32
	if r, _, _ := obtenerModo.Call(uintptr(manejador),
		uintptr(unsafe.Pointer(&modo))); r == 0 {
		return false // no es una consola de verdad (una tuberia, un archivo)
	}
	if modo&habilitarSecuencias != 0 {
		return true // ya venia preparada
	}
	r, _, _ := ponerModo.Call(uintptr(manejador), uintptr(modo|habilitarSecuencias))
	return r != 0
}
