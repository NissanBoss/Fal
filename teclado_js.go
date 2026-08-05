//go:build js

package main

// En el navegador las teclas no se leen: llegan. La pagina las va mandando y
// aqui se guardan en cola hasta que el programa las pide con "tecla".
//
// La cola solo se llena mientras el programa deja respirar al navegador, que
// es lo que hace "espera". Por eso un juego siempre lleva una espera dentro
// del bucle: ademas de marcar el ritmo, es el momento en que entran las
// teclas.

var teclasPendientes []string

func apuntarTecla(t string) {
	if t != "" {
		teclasPendientes = append(teclasPendientes, t)
	}
}

func olvidarTeclas() { teclasPendientes = nil }

func soltarTeclado() {}

func leerTecla() string {
	if len(teclasPendientes) == 0 {
		return ""
	}
	t := teclasPendientes[0]
	teclasPendientes = teclasPendientes[1:]
	return t
}
