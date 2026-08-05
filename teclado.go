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
		return leerTecla(), nil
	})
}
