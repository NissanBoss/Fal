package main

import (
	"bufio"
	"bytes"
	"math"
	"os"
	"regexp"
	"strings"
	"testing"
)

// dibuja corre un programa y devuelve la tortuga con lo que haya trazado.
func dibuja(t *testing.T, fuente string) *Tortuga {
	t.Helper()
	carpeta, _ := os.Getwd()
	in := nuevoInterprete(carpeta, nil)
	in.salida = bufio.NewWriter(&bytes.Buffer{})
	if err := correrFuente(in, fuente); err != nil {
		t.Fatalf("el programa fallo: %s", err.Mensaje)
	}
	return in.tortuga
}

func TestCuadradoDejaCuatroTrazosYVuelveAlSitio(t *testing.T) {
	tor := dibuja(t, "repite 4 veces\n camina de 80\n gira de 90\nfin")
	if len(tor.trazos) != 4 {
		t.Fatalf("esperaba 4 trazos, hay %d", len(tor.trazos))
	}
	// Cuatro lados y cuatro giros de 90 tienen que cerrar el cuadrado.
	if math.Abs(tor.x) > 0.001 || math.Abs(tor.y) > 0.001 {
		t.Errorf("el cuadrado no cierra: acabo en %.3f, %.3f", tor.x, tor.y)
	}
}

func TestLapizLevantadoNoDejaRaya(t *testing.T) {
	tor := dibuja(t, "levanta\ncamina de 50\napoya\ncamina de 50")
	if len(tor.trazos) != 1 {
		t.Fatalf("esperaba 1 trazo (el de despues de apoyar), hay %d", len(tor.trazos))
	}
	// El unico trazo tiene que empezar donde acabo el paseo sin pintar.
	if math.Abs(tor.trazos[0].Y1-(-50)) > 0.001 {
		t.Errorf("el trazo no empieza donde se apoyo el lapiz: %.3f", tor.trazos[0].Y1)
	}
}

func TestCaminarHaciaAtras(t *testing.T) {
	// Los parentesis hacen falta: detras de un "de" no cabe un "menos N"
	// suelto. Pasa con todas las funciones, no solo con esta.
	tor := dibuja(t, "camina de (menos 40)")
	if len(tor.trazos) != 1 {
		t.Fatalf("esperaba 1 trazo, hay %d", len(tor.trazos))
	}
	// Mirando hacia arriba, retroceder baja: la y crece.
	if tor.trazos[0].Y2 <= tor.trazos[0].Y1 {
		t.Error("caminar un numero negativo tendria que ir hacia atras")
	}
}

func TestColorDesconocidoSugiereElParecido(t *testing.T) {
	_, err := ejecutarEnMemoria(`color de "morao"`)
	if err == nil {
		t.Fatal("un color que no existe tendria que dar error")
	}
	if !strings.Contains(err.Pista, "morado") {
		t.Errorf(`la pista tendria que sugerir "morado", dice: %q`, err.Pista)
	}
}

func TestSinDibujarNoSaleSvg(t *testing.T) {
	tor := dibuja(t, `escribe "aqui no se dibuja nada"`)
	if tor.svg() != "" {
		t.Error("sin trazos no tendria que haber svg")
	}
}

func TestElSvgLlevaTodasLasLineasYElColor(t *testing.T) {
	tor := dibuja(t, "color de \"rojo\"\nrepite 36 veces\n camina de 100\n gira de 170\nfin")
	svg := tor.svg()
	if n := strings.Count(svg, "<line"); n != 36 {
		t.Errorf("esperaba 36 lineas en el svg, hay %d", n)
	}
	if !strings.Contains(svg, coloresTortuga["rojo"]) {
		t.Error("el color elegido no aparece en el svg")
	}
	if !strings.Contains(svg, "viewBox=") {
		t.Error("sin viewBox el dibujo no se ajusta al hueco que tenga")
	}
}

func TestLasLineasSinColorSiguenAlTema(t *testing.T) {
	tor := dibuja(t, "camina de 10")
	// currentColor hace que la raya salga del color del texto de alrededor,
	// que es lo que permite que el dibujo se vea en claro y en oscuro.
	if !strings.Contains(tor.svg(), "currentColor") {
		t.Error("una linea sin color propio tendria que usar currentColor")
	}
}

// El manual dice cuantas funciones trae el lenguaje. Es de las cosas que se
// quedan viejas sin que nadie se entere, asi que mejor que lo diga una
// prueba y no la buena voluntad.
func TestElManualCuentaBienLasFunciones(t *testing.T) {
	datos, err := os.ReadFile("MANUAL.md")
	if err != nil {
		t.Skip("no encuentro el manual")
	}
	m := regexp.MustCompile(`\((\d+) funciones\)`).FindStringSubmatch(string(datos))
	if m == nil {
		t.Skip("el manual ya no dice cuantas funciones hay")
	}
	if m[1] != itoa(len(integradas)) {
		t.Errorf("el manual dice %s funciones y hay %d", m[1], len(integradas))
	}
}
