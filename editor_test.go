package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// Comprueba que el coloreado que se genera es JSON valido y que sus
// patrones compilan de verdad.
func TestGenerarEditor(t *testing.T) {
	destino := filepath.Join(t.TempDir(), "vscode-fal")
	if codigo := generarEditor(destino); codigo != 0 {
		t.Fatalf("generarEditor devolvio %d", codigo)
	}

	for _, nombre := range []string{"package.json", "language-configuration.json",
		filepath.Join("syntaxes", "fal.tmLanguage.json")} {
		datos, err := os.ReadFile(filepath.Join(destino, nombre))
		if err != nil {
			t.Fatalf("falta %s: %v", nombre, err)
		}
		var v interface{}
		if err := json.Unmarshal(datos, &v); err != nil {
			t.Fatalf("%s no es JSON valido: %v", nombre, err)
		}
	}

	datos, _ := os.ReadFile(filepath.Join(destino, "syntaxes", "fal.tmLanguage.json"))
	var g struct {
		Patterns []map[string]interface{} `json:"patterns"`
	}
	json.Unmarshal(datos, &g)
	if len(g.Patterns) < 10 {
		t.Fatalf("esperaba al menos 10 patrones, hay %d", len(g.Patterns))
	}
	comprobados := 0
	for _, p := range g.Patterns {
		for _, clave := range []string{"match", "begin", "end"} {
			s, ok := p[clave].(string)
			if !ok {
				continue
			}
			// Go usa RE2, que no admite lookbehind; VS Code usa Oniguruma,
			// que si. Aqui solo comprobamos que el patron no este vacio y
			// que las partes sin lookaround compilen.
			if s == "" {
				t.Fatalf("patron vacio en %v", p["name"])
			}
			sinMirada := regexp.MustCompile(`\(\?<!\[[^\]]*\]\)|\(\?![^)]*\)`).ReplaceAllString(s, "")
			if _, err := regexp.Compile(sinMirada); err != nil {
				t.Fatalf("patron invalido %q: %v", p["name"], err)
			}
			comprobados++
		}
	}
	t.Logf("%d patrones comprobados", comprobados)
}

// Comprueba el corazon del lenguaje sin lanzar ningun proceso aparte.
func TestLenguaje(t *testing.T) {
	casos := []struct{ programa, espera string }{
		{`escribe 0.1 mas 0.2`, "0.3"},
		{`escribe (0.1 mas 0.2) es 0.3`, "verdadero"},
		{`escribe 19.99 por 3`, "59.97"},
		{`escribe 1 entre 3`, "0.3333333333333333333333333333"},
		{`escribe 2 mas 3 por 4`, "14"},
		{`escribe largo de "hola"`, "4"},
		{`escribe elemento menos 1 de (lista con 1 y 2 y 3)`, "3"},
		{"funcion d con n\n devuelve n por 2\nfin\nescribe mapa con (lista con 1 y 2) y funcion d", "[2, 4]"},
		{"total es 9\nfuncion f\n total es 0\n devuelve total\nfin\nescribe f\nescribe total", "0\n9"},
		{`escribe formato de "{} y {}" con 1 y 2`, "1 y 2"},
		{"intenta\n escribe 1 entre 0\nsi falla de matematica\n escribe \"ok\"\nfin", "ok"},
	}
	for _, c := range casos {
		salida, err := ejecutarEnMemoria(c.programa)
		if err != nil {
			t.Errorf("%q fallo: %s", c.programa, err.Mensaje)
			continue
		}
		if salida != c.espera {
			t.Errorf("%q\n  esperaba: %q\n  obtuve:   %q", c.programa, c.espera, salida)
		}
	}
}
