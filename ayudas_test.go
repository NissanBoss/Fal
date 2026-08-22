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
	// Sin entrada propia, una prueba que use "pregunta" o "tecla" se pondria
	// a leer del teclado de quien la lanza.
	in.conEntrada(strings.NewReader(""))
	if err := correrFuente(in, fuente); err != nil {
		return "", err
	}
	in.salida.Flush()
	return strings.TrimRight(buf.String(), "\n"), nil
}
