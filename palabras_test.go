package main

import "strings"
import "testing"

// Los tropiezos de siempre al empezar tienen que explicarse solos.
func TestPalabrasCasiBien(t *testing.T) {
	casos := map[string]string{
		`escribir "Hola"`: "escribe",
		`imprime "Hola"`:  "escribe",
		`print "Hola"`:    "escribe",
		`muestra "Hola"`:  "escribe",
		`escrive "Hola"`:  "escribe",
		`repetir 3 veces`: "repite",
		`return 5`:        "devuelve",
		`while x`:         "mientras",
	}
	for programa, espera := range casos {
		_, err := ejecutarEnMemoria(programa)
		if err == nil {
			t.Errorf("%q no dio error", programa)
			continue
		}
		if !strings.Contains(err.Pista, espera) {
			t.Errorf("%q\n  esperaba que sugiriera %q\n  dijo: %s", programa, espera, err.Pista)
		}
	}
}
