//go:build !js

package main

// Arranque: argumentos, consola interactiva y el formato de los errores.

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	soportaSecuencias = prepararConsola()
	args := os.Args[1:]

	if len(args) > 0 {
		switch args[0] {
		case "--ayuda", "-h", "--help", "ayuda":
			ayuda()
			return
		case "--editor":
			destino := "vscode-fal"
			if len(args) > 1 {
				destino = args[1]
			}
			os.Exit(generarEditor(destino))
		case "--probar":
			carpeta, filtro := "pruebas", ""
			if len(args) > 1 {
				carpeta = args[1]
			}
			if len(args) > 2 {
				filtro = args[2]
			}
			os.Exit(probar(carpeta, filtro))
		}
		os.Exit(ejecutarArchivo(args[0], args[1:]))
	}
	os.Exit(consola())
}

func ayuda() {
	fmt.Println(`Fal - el lenguaje de programacion mas facil del mundo.

  fal programa.fal        ejecuta un programa
  fal                     abre la consola interactiva
  fal --probar [carpeta]  ejecuta el banco de pruebas
  fal --editor [carpeta]  genera el coloreado para VS Code

Todo lo que necesitas esta dentro de este unico archivo.
No hace falta instalar nada mas.`)
}

func ejecutarArchivo(ruta string, args []string) int {
	datos, e := os.ReadFile(ruta)
	if e != nil {
		fmt.Fprintf(os.Stderr, "No pude abrir el archivo %s\n", ruta)
		return 1
	}
	fuente := quitarBOM(string(datos))

	completa, _ := filepath.Abs(ruta)
	in := nuevoInterprete(filepath.Dir(completa), args)
	defer in.salida.Flush()

	if err := correrFuente(in, fuente); err != nil {
		in.salida.Flush()
		mostrarError(err, fuente)
		return 1
	}
	return 0
}

func mostrarError(err *ErrorFal, fuente string) {
	escribirError(os.Stderr, err, fuente)
}

// La consola interactiva

var abrenBloque = map[string]bool{
	"si": true, "mientras": true, "repite": true, "para": true,
	"funcion": true, "tipo": true, "intenta": true,
}

func consola() int {
	fmt.Println("Fal - escribe tu programa linea a linea.")
	fmt.Println(`Escribe "adios" para salir.`)
	fmt.Println()

	carpeta, _ := os.Getwd()
	in := nuevoInterprete(carpeta, nil)
	lector := bufio.NewReader(os.Stdin)

	var acumulado []string
	profundidad := 0

	for {
		if len(acumulado) > 0 {
			fmt.Print("...  ")
		} else {
			fmt.Print(">>>  ")
		}
		linea, e := lector.ReadString('\n')
		if e != nil && linea == "" {
			fmt.Println()
			return 0
		}
		linea = strings.TrimRight(linea, "\r\n")

		if len(acumulado) == 0 {
			switch clave(strings.TrimSpace(linea)) {
			case "adios", "salir", "chao":
				return 0
			}
		}

		palabras := strings.Fields(strings.ReplaceAll(linea, "\t", " "))
		if len(palabras) > 0 {
			primera := clave(palabras[0])
			esSino := primera == "si" && len(palabras) > 1 &&
				(clave(palabras[1]) == "no" || clave(palabras[1]) == "falla")
			if abrenBloque[primera] && !esSino {
				profundidad++
			} else if primera == "fin" {
				profundidad--
				if profundidad < 0 {
					profundidad = 0
				}
			}
		}

		acumulado = append(acumulado, linea)
		if profundidad > 0 {
			continue
		}

		fuente := strings.Join(acumulado, "\n")
		acumulado = nil
		if strings.TrimSpace(fuente) == "" {
			continue
		}
		if err := correrFuente(in, fuente); err != nil {
			in.salida.Flush()
			mostrarError(err, fuente)
		}
		in.salida.Flush()
	}
}
