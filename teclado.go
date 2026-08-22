package main

// La palabra "tecla".
//
// Dice que tecla se acaba de pulsar, o texto vacio si ninguna. Lo importante
// es que NO espera: el programa sigue su camino aunque nadie toque nada, que
// es justo lo que hace falta para un juego.
//
// "pregunta" es lo contrario y sigue estando: se para hasta que escribes una
// linea entera y le das a Enter. Una sirve para pedir datos, la otra para
// jugar.

import "strings"

// Los nombres de las teclas que no son una letra. Van en castellano para
// poder escribir  si t es "arriba"  sin acordarse de ningun codigo.
const (
	teclaArriba    = "arriba"
	teclaAbajo     = "abajo"
	teclaIzquierda = "izquierda"
	teclaDerecha   = "derecha"
	teclaEscape    = "escape"
	teclaIntro     = "intro"
	teclaEspacio   = "espacio"
)

func registrarTeclado() {
	integrada("tecla", 0, 0, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		// Sin vaciar antes, lo que el programa acaba de pintar todavia no se
		// ve y estarias jugando a ciegas.
		in.salida.Flush()
		return in.unaTecla(), nil
	})
}

// unaTecla da la tecla recien pulsada, o texto vacio si ninguna.
//
// Con una terminal delante se le pregunta al teclado. Cuando no la hay (una
// tuberia, un archivo, o el banco de pruebas) se saca la siguiente letra del
// texto que le entra al programa, de forma que una partida se puede escribir
// antes y repetirse igual:
//
//	fal snake.fal < partida.txt
//
// Sin esto un juego no habria manera de probarlo, porque las teclas de
// verdad no se pueden teclear desde una prueba.
func (in *Interprete) unaTecla() string {
	if in.teclado {
		return leerTecla()
	}
	for {
		r, _, err := in.entrada.ReadRune()
		if err != nil {
			return "" // se acabo el guion: a partir de aqui nadie toca nada
		}
		switch r {
		case '\r':
			continue // el compañero del salto de linea en Windows, no es tecla
		case '\n':
			return teclaIntro
		case ' ':
			return teclaEspacio
		}
		return strings.ToLower(string(r))
	}
}
