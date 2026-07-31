package main

// Lo que usan por igual el programa de escritorio, la version para el
// navegador y las pruebas.

import (
	"fmt"
	"io"
	"strings"
)

// soportaSecuencias dice si la consola entiende las ordenes para mover el
// cursor y borrar la pantalla. Se averigua una sola vez, al arrancar.
var soportaSecuencias bool

func correrFuente(in *Interprete, fuente string) *ErrorFal {
	piezas, err := leer(fuente)
	if err != nil {
		return err
	}
	instrucciones, err := armar(piezas)
	if err != nil {
		return err
	}
	return in.correr(instrucciones)
}

// escribirError compone el mensaje. Va aparte para que las pruebas puedan
// recogerlo sin tener que mirar la pantalla.
func escribirError(destino io.Writer, err *ErrorFal, fuente string) {
	lineas := strings.Split(fuente, "\n")
	textoLinea := func(n int) string {
		if n >= 1 && n <= len(lineas) {
			return strings.TrimSpace(lineas[n-1])
		}
		return ""
	}

	var b strings.Builder
	b.WriteString("\n  X  " + err.Mensaje + "\n")
	if t := textoLinea(err.Linea); t != "" {
		b.WriteString("\n     linea " + itoa(err.Linea) + " |  " + t + "\n")
	}
	// El camino que siguio el programa hasta reventar, de dentro hacia fuera.
	if len(err.Pila) > 0 {
		b.WriteString("\n     Se llego aqui asi:\n")
		for _, m := range err.Pila {
			detalle := ""
			if t := textoLinea(m.Linea); t != "" {
				detalle = "  |  " + t
			}
			b.WriteString("       dentro de " + m.Nombre + ", llamada en la linea " +
				itoa(m.Linea) + detalle + "\n")
		}
	}
	if err.Pista != "" {
		b.WriteString("\n     Pista: " + err.Pista + "\n")
	}
	b.WriteString("\n")
	fmt.Fprint(destino, b.String())
}
