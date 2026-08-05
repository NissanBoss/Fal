package main

import "testing"

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
