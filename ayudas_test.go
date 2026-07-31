package main

import (
	"bufio"
	"bytes"
	"os"
	"strings"
)

// ejecutarEnMemoria corre un programa y devuelve lo que escribio, sin
// tocar disco ni lanzar procesos. Se usa desde las pruebas de Go.
func ejecutarEnMemoria(fuente string) (string, *ErrorFal) {
	var buf bytes.Buffer
	carpeta, _ := os.Getwd()
	in := nuevoInterprete(carpeta, nil)
	in.salida = bufio.NewWriter(&buf)
	if err := correrFuente(in, fuente); err != nil {
		return "", err
	}
	in.salida.Flush()
	return strings.TrimRight(buf.String(), "\n"), nil
}
