package main

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// El analizador de gastos es el ejemplo que enseña para que sirven los
// numeros exactos: los mismos 23 importes en coma flotante dan 1546.1900000000005
// y no cuadran con la suma por categorias.
func TestGastos(t *testing.T) {
	ruta := filepath.Join("ejemplos", "gastos.fal")
	datos, err := os.ReadFile(ruta)
	if err != nil {
		t.Fatal(err)
	}
	completa, _ := filepath.Abs(ruta)

	var buf bytes.Buffer
	in := nuevoInterprete(filepath.Dir(completa), nil)
	in.salida = bufio.NewWriter(&buf)
	in.entrada = bufio.NewReader(strings.NewReader(""))
	if e := correrFuente(in, quitarBOM(string(datos))); e != nil {
		in.salida.Flush()
		escribirError(&buf, e, string(datos))
		t.Fatal(buf.String())
	}
	in.salida.Flush()
	salida := buf.String()

	for _, espera := range []string{
		"1546.19 EUR", // el total, exacto
		"798.67 EUR",  // vivienda
		"133.10 EUR",  // con sus dos decimales, no 133.1
		"suman el total al centimo",
		"67.23 EUR", // la media
	} {
		if !strings.Contains(salida, espera) {
			t.Errorf("falta %q en el informe:\n%s", espera, salida)
		}
	}
}
