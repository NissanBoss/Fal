//go:build js

package main

// La version que corre dentro del navegador.
//
// Se compila a WebAssembly y expone una funcion a la pagina:
//
//   falEjecutar(programa, respuestas)  ->  { salida, error }
//
// "respuestas" son las lineas que el programa recibira cuando use
// "pregunta", una por linea. En el navegador no hay teclado bloqueante,
// asi que se le dan de antemano.

import (
	"bufio"
	"bytes"
	"strings"
	"syscall/js"
	"time"
)

// Tope de tiempo por ejecucion. Sin esto, un "mientras verdadero" colgaria
// la pestaña entera y habria que cerrarla a la fuerza.
const tiempoMaximo = 5 * time.Second

func main() {
	js.Global().Set("falEjecutar", js.FuncOf(ejecutarDesdeWeb))
	// El programa no puede terminar: si lo hace, la funcion deja de existir.
	select {}
}

func ejecutarDesdeWeb(this js.Value, args []js.Value) any {
	resultado := map[string]any{"salida": "", "error": ""}
	if len(args) == 0 {
		resultado["error"] = "No me llego ningun programa."
		return js.ValueOf(resultado)
	}

	programa := args[0].String()
	respuestas := ""
	if len(args) > 1 && args[1].Type() == js.TypeString {
		respuestas = args[1].String()
	}
	if respuestas != "" && !strings.HasSuffix(respuestas, "\n") {
		respuestas += "\n"
	}

	var salida bytes.Buffer
	in := nuevoInterprete(".", nil)
	in.salida = bufio.NewWriter(&salida)
	in.entrada = bufio.NewReader(strings.NewReader(respuestas))
	in.limite = time.Now().Add(tiempoMaximo)

	err := correrFuente(in, programa)
	in.salida.Flush()
	resultado["salida"] = salida.String()

	if err != nil {
		var texto bytes.Buffer
		escribirError(&texto, err, programa)
		resultado["error"] = strings.TrimRight(texto.String(), "\n")
	}
	return js.ValueOf(resultado)
}
