package main

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBancoDePruebas corre las mismas pruebas que "fal --probar", pero
// dentro del propio proceso. Sirve para poder comprobarlas aunque el
// sistema no deje lanzar el ejecutable.
func TestBancoDePruebas(t *testing.T) {
	entradas, err := os.ReadDir("pruebas")
	if err != nil {
		t.Skip("no hay carpeta de pruebas")
	}
	corridas := 0
	for _, e := range entradas {
		nombre := e.Name()
		if !strings.HasSuffix(nombre, ".fal") || strings.HasPrefix(nombre, "_") {
			continue
		}
		base := strings.TrimSuffix(nombre, ".fal")
		esperadoBytes, err := os.ReadFile(filepath.Join("pruebas", base+".esperado"))
		if err != nil {
			continue
		}
		corridas++
		t.Run(base, func(t *testing.T) {
			obtenido := correrArchivoEnMemoria(filepath.Join("pruebas", nombre))
			esperado := normalizar(string(esperadoBytes))
			if obtenido != esperado {
				esp := strings.Split(esperado, "\n")
				obt := strings.Split(obtenido, "\n")
				total := len(esp)
				if len(obt) > total {
					total = len(obt)
				}
				for i := 0; i < total; i++ {
					a, b := "<falta>", "<falta>"
					if i < len(esp) {
						a = esp[i]
					}
					if i < len(obt) {
						b = obt[i]
					}
					if a != b {
						t.Errorf("linea %d\n  esperaba: %s\n  obtuve:   %s", i+1, a, b)
					}
				}
			}
		})
	}
	if corridas == 0 {
		t.Fatal("no se ejecuto ninguna prueba")
	}
	t.Logf("%d pruebas del lenguaje", corridas)
}

func correrArchivoEnMemoria(ruta string) string {
	datos, e := os.ReadFile(ruta)
	if e != nil {
		return "no pude abrir " + ruta
	}
	fuente := quitarBOM(string(datos))
	completa, _ := filepath.Abs(ruta)

	var buf bytes.Buffer
	in := nuevoInterprete(filepath.Dir(completa), nil)
	in.salida = bufio.NewWriter(&buf)
	in.entrada = bufio.NewReader(strings.NewReader(""))

	if err := correrFuente(in, fuente); err != nil {
		in.salida.Flush()
		escribirError(&buf, err, fuente)
		return normalizar(buf.String())
	}
	in.salida.Flush()
	return normalizar(buf.String())
}
