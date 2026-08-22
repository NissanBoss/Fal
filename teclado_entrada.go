//go:build !js

package main

// Si hay una terminal de verdad al otro lado o no.

import "os"

// hayTecladoDeVerdad dice si la entrada del programa es una terminal con
// alguien delante. Cuando no lo es (una tuberia, un archivo redirigido, o el
// banco de pruebas) no hay teclas que pulsar, y "tecla" pasa a leerlas del
// texto que le llega.
//
// Hay que preguntarlo antes de tocar el teclado: en Windows _kbhit contesta
// lo mismo si no hay nadie tecleando que si no hay consola, asi que por ahi
// no se distingue.
func hayTecladoDeVerdad() bool {
	info, err := os.Stdin.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
