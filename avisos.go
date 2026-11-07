package main

// Fal no distingue mayusculas, y eso tiene un precio: marco y MARCO son la
// misma variable. Basta escribir marco dentro de una funcion para tapar la
// constante, y nada falla hasta que alguien la lee, lejos de donde se piso.
// A FalOS le ha pasado con MARCO, con HISTORIAL y con SUBIDAS. No es un
// error porque el programa puede ser correcto: se avisa y se sigue.

import (
	"fmt"
	"io"
	"strings"
)

type Choque struct {
	Nombre    string // como se escribio el que pisa
	Linea     int
	Constante string
	LineaDe   int // donde se puso la constante
}

// Una sola letra no se cuenta: X e x suelen ser dos formas de escribir una
// incognita, no una constante y una variable.
func pareceConstante(nombre string) bool {
	return len([]rune(nombre)) > 1 && nombre == strings.ToUpper(nombre) &&
		nombre != strings.ToLower(nombre)
}

// nombresPuestos recorre las piezas y no lo ya armado porque armar pasa los
// nombres a minusculas, y aqui lo unico que importa es como se escribieron.
func nombresPuestos(piezas []Pieza) (asignados, ligados []Pieza) {
	esNombre := func(i int) bool {
		return i < len(piezas) && piezas[i].Tipo == PPalabra && !reservadas[piezas[i].Clave]
	}
	palabra := func(i int, clave string) bool {
		return i < len(piezas) && piezas[i].Tipo == PPalabra && piezas[i].Clave == clave
	}
	alEmpezar := true
	for i, p := range piezas {
		if alEmpezar && esNombre(i) && palabra(i+1, "es") {
			asignados = append(asignados, p)
		}
		alEmpezar = p.Tipo == PFinLinea
		switch {
		case palabra(i, "para"):
			n := i + 1
			if palabra(n, "cada") {
				n++
			}
			if esNombre(n) {
				ligados = append(ligados, piezas[n])
			}
		case palabra(i, "funcion"):
			n := i + 1
			if esNombre(n) {
				n++
			}
			if !palabra(n, "con") {
				continue
			}
			for n++; esNombre(n); n += 2 {
				ligados = append(ligados, piezas[n])
				if !palabra(n+1, "y") {
					break
				}
			}
		}
	}
	return asignados, ligados
}

// choquesDeNombres da un choque por nombre, el primero: con arreglar ese y
// volver a compilar se ve si queda otro, y un nombre repetido treinta veces
// en una funcion no entierra a los demas.
func choquesDeNombres(piezas []Pieza) []Choque {
	asignados, ligados := nombresPuestos(piezas)
	constantes := map[string]Pieza{}
	for _, p := range asignados {
		minusculas := strings.ToLower(p.Texto)
		if _, hay := constantes[minusculas]; !hay && pareceConstante(p.Texto) {
			constantes[minusculas] = p
		}
	}

	var choques []Choque
	visto := map[string]bool{}
	apunta := func(p Pieza) {
		minusculas := strings.ToLower(p.Texto)
		constante, hay := constantes[minusculas]
		if !hay || p.Texto == constante.Texto || visto[minusculas] {
			return
		}
		visto[minusculas] = true
		choques = append(choques, Choque{p.Texto, p.Linea, constante.Texto, constante.Linea})
	}
	for _, p := range append(asignados, ligados...) {
		apunta(p)
	}
	return choques
}

// avisarChoques no dice nada si el programa ni siquiera se lee: de eso ya
// se queja quien lo vaya a ejecutar o compilar.
func avisarChoques(destino io.Writer, archivo, fuente string) {
	piezas, err := leer(fuente)
	if err != nil {
		return
	}
	lineas := strings.Split(fuente, "\n")
	for _, c := range choquesDeNombres(piezas) {
		fmt.Fprintf(destino, "\n  !  %s: %s y la constante %s (linea %d) son el mismo nombre.\n",
			archivo, c.Nombre, c.Constante, c.LineaDe)
		if c.Linea >= 1 && c.Linea <= len(lineas) {
			fmt.Fprintf(destino, "\n     linea %d |  %s\n", c.Linea, strings.TrimSpace(lineas[c.Linea-1]))
		}
		fmt.Fprintf(destino, "\n     Pista: En Fal las mayusculas no importan, asi que esto pisa la constante. "+
			"Cambia uno de los dos nombres.\n\n")
	}
}
