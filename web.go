//go:build js

package main

// La version que corre dentro del navegador.
//
// Se compila a WebAssembly y expone una funcion a la pagina:
//
//   falEjecutar(programa, respuestas, alEscribir, alTerminar)
//
// "respuestas" son las lineas que el programa recibira cuando use
// "pregunta", una por linea, porque eso si se pide de antemano.
//
// "alEscribir" se llama con cada trozo que el programa saca por pantalla, en
// el momento en que lo saca. Antes se acumulaba todo y se devolvia al final,
// que es lo que hacia imposible cualquier animacion: cuando veias algo, el
// programa ya habia acabado.
//
// El programa arranca en una gorutina y falEjecutar vuelve enseguida; el
// final se avisa por "alTerminar". Tiene que ser asi: si el interprete
// corriera dentro de la propia llamada, el worker se quedaria ocupado y no
// llegaria a leer los mensajes que le manda la pagina, que es justo por
// donde entran las teclas.
//
// Esto corre dentro de un worker (ver docs/trabajador.js), asi que un bucle
// largo ya no congela la pagina y no hace falta cortar por tiempo.

import (
	"bufio"
	"bytes"
	"strings"
	"syscall/js"
)

// escritorPagina manda cada trozo segun llega, sin guardar nada.
type escritorPagina struct{ alEscribir js.Value }

func (e escritorPagina) WriteString(s string) (int, error) {
	if s != "" {
		e.alEscribir.Invoke(s)
	}
	return len(s), nil
}

func (e escritorPagina) Flush() error { return nil }

func main() {
	// La pagina si entiende la orden de borrar la pantalla: la reconoce en la
	// salida y vacia el panel de verdad. Antes esto era false y "limpia" se
	// apañaba soltando cincuenta saltos de linea.
	soportaSecuencias = true
	js.Global().Set("falEjecutar", js.FuncOf(ejecutarDesdeWeb))
	js.Global().Set("falTecla", js.FuncOf(teclaDesdeWeb))
	// El programa no puede terminar: si lo hace, las funciones dejan de existir.
	select {}
}

func teclaDesdeWeb(this js.Value, args []js.Value) any {
	if len(args) > 0 && args[0].Type() == js.TypeString {
		apuntarTecla(args[0].String())
	}
	return nil
}

func ejecutarDesdeWeb(this js.Value, args []js.Value) any {
	if len(args) < 4 {
		return nil
	}
	alTerminar := args[3]

	programa := args[0].String()
	respuestas := ""
	if len(args) > 1 && args[1].Type() == js.TypeString {
		respuestas = args[1].String()
	}
	if respuestas != "" && !strings.HasSuffix(respuestas, "\n") {
		respuestas += "\n"
	}

	in := nuevoInterprete(".", nil)
	in.salida = escritorPagina{args[2]}
	in.entrada = bufio.NewReader(strings.NewReader(respuestas))
	// Las teclas de la partida anterior no valen para esta.
	olvidarTeclas()

	go func() {
		resultado := map[string]any{"error": "", "dibujo": ""}
		err := correrFuente(in, programa)
		in.salida.Flush()
		resultado["dibujo"] = in.tortuga.svg()
		if err != nil {
			var texto bytes.Buffer
			escribirError(&texto, err, programa)
			resultado["error"] = strings.TrimRight(texto.String(), "\n")
		}
		alTerminar.Invoke(js.ValueOf(resultado))
	}()
	return nil
}
