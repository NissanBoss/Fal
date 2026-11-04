//go:build !js

package main

// "fal --nucleo" y "fal --arranca": compilar el sistema y, si se pide,
// arrancarlo en QEMU con el puerto serie conectado a esta terminal.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func compilarYArrancar(ruta string, arrancar bool) int {
	datos, e := os.ReadFile(ruta)
	if e != nil {
		fmt.Fprintf(os.Stderr, "No pude abrir el archivo %s\n", ruta)
		return 1
	}
	fuente := quitarBOM(string(datos))
	carpeta := filepath.Dir(ruta)
	imagen, err := compilarImagen(fuente, opcionesNucleo{leer: func(incluida string) (string, error) {
		datos, e := os.ReadFile(filepath.Join(carpeta, filepath.FromSlash(incluida)))
		return string(datos), e
	}})
	if err != nil {
		mostrarError(err, fuente)
		return 1
	}
	destino := strings.TrimSuffix(ruta, filepath.Ext(ruta)) + ".bin"
	if e := os.WriteFile(destino, imagen, 0644); e != nil {
		fmt.Fprintf(os.Stderr, "No pude guardar la imagen en %s\n", destino)
		return 1
	}
	if !arrancar {
		fmt.Printf("Listo: %s (%d bytes)\n", destino, len(imagen))
		fmt.Printf("Arrancalo con:   qemu-system-x86_64 -kernel %s -serial stdio\n", destino)
		return 0
	}

	qemu := buscarQemu()
	if qemu == "" {
		fmt.Fprintf(os.Stderr, "La imagen esta en %s, pero no encuentro QEMU para arrancarla.\n", destino)
		fmt.Fprintln(os.Stderr, "Instalalo con uno de estos:")
		fmt.Fprintln(os.Stderr, "  Windows:  winget install SoftwareFreedomConservancy.QEMU")
		fmt.Fprintln(os.Stderr, "  Linux:    sudo apt install qemu-system-x86")
		fmt.Fprintln(os.Stderr, "  macOS:    brew install qemu")
		return 1
	}
	// -no-reboot porque un nucleo que se cuelga del todo hace que el
	// procesador se reinicie, y sin esto QEMU lo vuelve a arrancar en bucle
	// sin que se vea que paso.
	orden := exec.Command(qemu, "-kernel", destino, "-serial", "stdio", "-no-reboot")
	orden.Stdin, orden.Stdout, orden.Stderr = os.Stdin, os.Stdout, os.Stderr
	if e := orden.Run(); e != nil {
		if _, salio := e.(*exec.ExitError); !salio {
			fmt.Fprintf(os.Stderr, "No pude arrancar QEMU: %v\n", e)
			return 1
		}
	}
	return 0
}

// buscarQemu mira el PATH y, en Windows, donde lo dejan sus instaladores,
// que no siempre lo añaden al PATH.
func buscarQemu() string {
	nombre := "qemu-system-x86_64"
	if ruta, e := exec.LookPath(nombre); e == nil {
		return ruta
	}
	if runtime.GOOS != "windows" {
		return ""
	}
	for _, carpeta := range []string{
		`C:\Program Files\qemu`,
		`C:\msys64\ucrt64\bin`,
		`C:\msys64\mingw64\bin`,
	} {
		ruta := filepath.Join(carpeta, nombre+".exe")
		if _, e := os.Stat(ruta); e == nil {
			return ruta
		}
	}
	return ""
}
