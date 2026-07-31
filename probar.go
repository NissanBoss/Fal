//go:build !js

package main

// El banco de pruebas, dentro del propio ejecutable.
//
//   fal --probar             ejecuta todas las pruebas de ./pruebas
//   fal --probar carpeta     ejecuta las de otra carpeta
//   fal --probar . textos    solo las que llevan "textos" en el nombre
//
// Cada prueba es un archivo <nombre>.fal junto a un <nombre>.esperado
// con la salida exacta que debe producir. Si falta el .esperado, se crea
// con la salida actual y se avisa para que la revises.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func probar(carpeta, filtro string) int {
	entradas, e := os.ReadDir(carpeta)
	if e != nil {
		fmt.Fprintf(os.Stderr, "No pude abrir la carpeta %s\n", carpeta)
		return 1
	}

	var nombres []string
	for _, x := range entradas {
		n := x.Name()
		// Los que empiezan por "_" son modulos de apoyo, no pruebas.
		if !strings.HasSuffix(n, ".fal") || strings.HasPrefix(n, "_") {
			continue
		}
		base := strings.TrimSuffix(n, ".fal")
		if filtro == "" || strings.Contains(base, filtro) {
			nombres = append(nombres, base)
		}
	}
	sort.Strings(nombres)

	yo, _ := os.Executable()
	type fallo struct{ nombre, esperado, obtenido string }
	var fallos []fallo
	var nuevas []string

	for _, nombre := range nombres {
		obtenido := correrPrueba(yo, carpeta, nombre+".fal")
		rutaEsperado := filepath.Join(carpeta, nombre+".esperado")

		datos, e := os.ReadFile(rutaEsperado)
		if e != nil {
			os.WriteFile(rutaEsperado, []byte(obtenido+"\n"), 0644)
			nuevas = append(nuevas, nombre)
			continue
		}
		esperado := normalizar(string(datos))
		if obtenido == esperado {
			fmt.Println("  ok    " + nombre)
		} else {
			fmt.Println("  FALLA " + nombre)
			fallos = append(fallos, fallo{nombre, esperado, obtenido})
		}
	}

	for _, n := range nuevas {
		fmt.Println("  nueva " + n + "  (esperado creado, revisalo)")
	}

	for _, f := range fallos {
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("FALLA: " + f.nombre)
		esp := strings.Split(f.esperado, "\n")
		obt := strings.Split(f.obtenido, "\n")
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
				fmt.Printf("  linea %d\n    esperaba: %s\n    obtuve:   %s\n", i+1, a, b)
			}
		}
	}

	fmt.Printf("\n%d pruebas, %d fallos, %d nuevas\n", len(nombres), len(fallos), len(nuevas))
	if len(fallos) > 0 {
		return 1
	}
	return 0
}

// correrPrueba lanza otra copia de este mismo programa, para que una
// prueba que se cuelgue o que llame a "termina" no se lleve por delante
// al que esta contando los resultados.
func correrPrueba(yo, carpeta, archivo string) string {
	cmd := exec.Command(yo, archivo)
	cmd.Dir = carpeta
	cmd.Stdin = nil
	salida, _ := cmd.CombinedOutput()
	return normalizar(string(salida))
}

func normalizar(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n"))
}
