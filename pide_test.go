package main

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"testing"
)

// Un % suelto en la direccion hacia que NewRequest no devolviera peticion,
// y pide la usaba igual: el interprete reventaba con un panic en vez de
// decir que la direccion estaba mal. No sale a la red, porque falla antes.
func TestPideConDireccionRotaAvisaSinReventar(t *testing.T) {
	carpeta, _ := os.Getwd()
	in := nuevoInterprete(carpeta, nil)
	in.salida = bufio.NewWriter(&bytes.Buffer{})
	err := correrFuente(in, `x es pide de "https://%zz"`)
	if err == nil {
		t.Fatal("esperaba un error con una direccion mal escrita")
	}
	if !strings.Contains(err.Mensaje, "no esta bien escrita") {
		t.Errorf("el aviso no dice que la direccion esta mal: %q", err.Mensaje)
	}
}
