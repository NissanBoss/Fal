package main

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"testing"
)

// Sin terminal delante, "tecla" saca las pulsaciones del texto que le llega.
// Es lo que permite escribir una partida antes y repetirla igual, y sin ello
// no habria forma de probar un juego.
func TestTeclaLeeDelTextoQueLlega(t *testing.T) {
	var buf bytes.Buffer
	carpeta, _ := os.Getwd()
	in := nuevoInterprete(carpeta, nil)
	in.salida = bufio.NewWriter(&buf)
	in.conEntrada(strings.NewReader("wa\n z"))

	fuente := `repite 6 veces
    escribe "[" mas tecla mas "]"
fin`
	if err := correrFuente(in, fuente); err != nil {
		t.Fatalf("fallo: %s", err.Mensaje)
	}
	in.salida.Flush()

	// Las letras tal cual, el salto de linea es "intro", el espacio es
	// "espacio", y cuando se acaba el guion vuelve a contestar vacio.
	esperado := "[w]\n[a]\n[intro]\n[espacio]\n[z]\n[]"
	if obtenido := strings.TrimRight(buf.String(), "\n"); obtenido != esperado {
		t.Errorf("esperaba\n%s\ny salio\n%s", esperado, obtenido)
	}
}

func TestTeclaNoSeQuedaEsperando(t *testing.T) {
	// Sin nadie tecleando, "tecla" contesta vacio y sigue. Si se quedara
	// esperando, esta prueba no terminaria nunca.
	salida, err := ejecutarEnMemoria(`escribe "[" mas tecla mas "]"`)
	if err != nil {
		t.Fatalf("fallo: %s", err.Mensaje)
	}
	if salida != "[]" {
		t.Errorf("esperaba [] y salio %q", salida)
	}
}
