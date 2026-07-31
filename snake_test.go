package main

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// juega corre snake.fal con las ordenes dadas y devuelve lo que pinta.
func juega(t *testing.T, ordenes string) string {
	t.Helper()
	return juegaConSemilla(t, ordenes, 0)
}

// juegaConSemilla corre snake.fal con las ordenes dadas. Con semilla > 0
// la partida es siempre identica.
func juegaConSemilla(t *testing.T, ordenes string, semilla int) string {
	t.Helper()
	ruta := filepath.Join("ejemplos", "snake.fal")
	datos, err := os.ReadFile(ruta)
	if err != nil {
		t.Fatal(err)
	}
	completa, _ := filepath.Abs(ruta)

	var buf bytes.Buffer
	in := nuevoInterprete(filepath.Dir(completa), nil)
	in.salida = bufio.NewWriter(&buf)
	in.entrada = bufio.NewReader(strings.NewReader(ordenes))

	fuente := strings.Replace(quitarBOM(string(datos)), "PAUSA es 0.06", "PAUSA es 0", 1)
	if semilla > 0 {
		fuente = "semilla de " + itoa(semilla) + "\n" + fuente
	}
	if e := correrFuente(in, fuente); e != nil {
		in.salida.Flush()
		escribirError(&buf, e, string(datos))
		t.Fatalf("snake fallo:\n%s", buf.String())
	}
	in.salida.Flush()
	return buf.String()
}

func TestSnakeChocaContraLaPared(t *testing.T) {
	// Empieza en la columna 8 mirando a la derecha; el tablero mide 22.
	salida := juega(t, strings.Repeat("d", 20)+"\nq\n")
	if !strings.Contains(salida, "Te has dado contra la pared") {
		t.Fatalf("esperaba choque contra la pared:\n%s", cola(salida))
	}
	t.Log(cola(salida))
}

func TestSnakeSeMuerdeLaCola(t *testing.T) {
	// Con cinco trozos, un giro cerrado (arriba, izquierda, abajo) hace
	// que la cabeza vuelva sobre el propio cuerpo.
	salida := juega(t, "was\n")
	if !strings.Contains(salida, "Te has mordido la cola") {
		t.Fatalf("esperaba mordisco en la cola:\n%s", cola(salida))
	}
	t.Log(cola(salida))
}

func TestSnakeSalirYTablero(t *testing.T) {
	salida := juega(t, "q\n")
	if !strings.Contains(salida, "Lo dejaste tu") {
		t.Fatal("no reconocio la orden de salir")
	}
	// El tablero tiene que dibujarse entero y con la serpiente dentro.
	if strings.Count(salida, "+----------------------+") < 2 {
		t.Error("faltan los bordes del tablero")
	}
	if !strings.Contains(salida, "@") {
		t.Error("no se ve la cabeza de la serpiente")
	}
	if !strings.Contains(salida, "*") {
		t.Error("no se ve la comida")
	}
	if !strings.Contains(salida, "Puntos: 0") {
		t.Error("no se ve el marcador")
	}
	t.Log(cola(salida))
}

func TestSnakeCreceAlComer(t *testing.T) {
	// Un barrido que recorre el tablero de lado a lado sin chocar: empieza
	// en la columna 8 de 22, asi que el primer tramo es mas corto.
	var b strings.Builder
	b.WriteString(strings.Repeat("d", 13))
	for fila := 0; fila < 5; fila++ {
		b.WriteString("s" + strings.Repeat("a", 20))
		b.WriteString("s" + strings.Repeat("d", 20))
	}
	barrido := b.String() + "\nq\n"

	// Con la semilla fijada la partida es siempre la misma, asi que esta
	// prueba no depende de la suerte.
	crecio := 0
	for semilla := 1; semilla <= 8; semilla++ {
		salida := juegaConSemilla(t, barrido, semilla)
		if strings.Contains(salida, "Puntos: 10") || strings.Contains(salida, "Puntos: 20") {
			crecio++
		}
	}
	if crecio == 0 {
		t.Error("con 8 semillas distintas no comio ni una vez")
	}
	t.Logf("comio en %d de 8 partidas", crecio)
}

func cola(s string) string {
	lineas := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lineas) > 18 {
		lineas = lineas[len(lineas)-18:]
	}
	return strings.Join(lineas, "\n")
}
