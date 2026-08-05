package main

// La tortuga: un lapiz que se arrastra por la pantalla y deja raya.
//
// Aqui no se pinta nada. El programa va apuntando los trazos segun avanza y
// al final los pinta quien puede: un canvas en el navegador, un archivo .svg
// desde la terminal. Asi la misma palabra vale en los dos sitios.

import (
	"math"
	"strconv"
	"strings"
)

type Trazo struct {
	X1, Y1, X2, Y2 float64
	Color          string // vacio = el color normal del texto
}

type Tortuga struct {
	x, y    float64
	angulo  float64 // grados; 0 mira hacia arriba y crece girando a la derecha
	color   string
	apoyada bool
	trazos  []Trazo
}

func nuevaTortuga() *Tortuga {
	return &Tortuga{apoyada: true}
}

// Los colores se dicen por su nombre. Son pocos a proposito: con doce hay de
// sobra para dibujar y no hay que aprenderse ningun codigo raro.
var coloresTortuga = map[string]string{
	"negro":    "#1a1a1a",
	"blanco":   "#ffffff",
	"gris":     "#8a8a8a",
	"rojo":     "#d83a3a",
	"naranja":  "#e08b2a",
	"amarillo": "#e6c229",
	"verde":    "#3f9e4d",
	"azul":     "#2f6fd0",
	"cian":     "#2aa8b8",
	"morado":   "#7a4fbf",
	"rosa":     "#d95f9a",
	"marron":   "#8a5a3c",
}

var nombresColores = []string{"negro", "blanco", "gris", "rojo", "naranja",
	"amarillo", "verde", "azul", "cian", "morado", "rosa", "marron"}

// limites devuelve el rectangulo por el que ha pasado la tortuga.
func (t *Tortuga) limites() (minX, minY, maxX, maxY float64) {
	minX, minY = t.trazos[0].X1, t.trazos[0].Y1
	maxX, maxY = minX, minY
	mira := func(x, y float64) {
		minX, maxX = math.Min(minX, x), math.Max(maxX, x)
		minY, maxY = math.Min(minY, y), math.Max(maxY, y)
	}
	for _, r := range t.trazos {
		mira(r.X1, r.Y1)
		mira(r.X2, r.Y2)
	}
	return
}

func num(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

// svg compone el dibujo entero. El tamaño sale de por donde ha pasado la
// tortuga, no de un lienzo fijo: asi da igual que el programa use numeros
// grandes o pequeños, porque el dibujo siempre acaba entero y centrado.
func (t *Tortuga) svg() string {
	if len(t.trazos) == 0 {
		return ""
	}
	minX, minY, maxX, maxY := t.limites()
	const margen = 12.0
	ancho := maxX - minX + margen*2
	alto := maxY - minY + margen*2

	var b strings.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="` +
		num(minX-margen) + " " + num(minY-margen) + " " + num(ancho) + " " + num(alto) +
		`" width="` + num(ancho) + `" height="` + num(alto) + `">` + "\n")
	b.WriteString(`<g fill="none" stroke-width="2" stroke-linecap="round" ` +
		`stroke-linejoin="round">` + "\n")
	for _, r := range t.trazos {
		color := r.Color
		if color == "" {
			color = "currentColor"
		}
		b.WriteString(`<line x1="` + num(r.X1) + `" y1="` + num(r.Y1) +
			`" x2="` + num(r.X2) + `" y2="` + num(r.Y2) +
			`" stroke="` + color + `"/>` + "\n")
	}
	b.WriteString("</g>\n</svg>\n")
	return b.String()
}

func registrarTortuga() {
	// camina mueve la tortuga en la direccion a la que mira. Para ir hacia
	// atras se le da un numero negativo, entre parentesis:
	//
	//     camina de (menos 40)
	//
	// No hay una palabra aparte para retroceder, igual que tampoco la hay
	// para girar a la izquierda.
	//
	// Se llama "camina" y no "avanza" porque "avanza" ya es lo de las fechas.
	integrada("camina", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		n, err := pideNum(a[0], ln, "camina")
		if err != nil {
			return nil, err
		}
		t := in.tortuga
		rad := t.angulo * math.Pi / 180
		destinoX := t.x + math.Sin(rad)*n.Float()
		destinoY := t.y - math.Cos(rad)*n.Float()
		if t.apoyada {
			t.trazos = append(t.trazos, Trazo{t.x, t.y, destinoX, destinoY, t.color})
		}
		t.x, t.y = destinoX, destinoY
		return nil, nil
	})

	// gira tuerce a la derecha los grados que le digas. A la izquierda se
	// gira con un numero negativo.
	integrada("gira", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		n, err := pideNum(a[0], ln, "gira")
		if err != nil {
			return nil, err
		}
		in.tortuga.angulo = math.Mod(in.tortuga.angulo+n.Float(), 360)
		return nil, nil
	})

	integrada("levanta", 0, 0, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		in.tortuga.apoyada = false
		return nil, nil
	})

	integrada("apoya", 0, 0, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		in.tortuga.apoyada = true
		return nil, nil
	})

	integrada("color", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		nombre, err := pideTexto(a[0], ln, "color")
		if err != nil {
			return nil, err
		}
		buscado := clave(nombre)
		codigo, hay := coloresTortuga[buscado]
		if !hay {
			pista := sugerir(buscado, nombresColores)
			if pista == "" {
				pista = "Los colores son: " + strings.Join(nombresColores, ", ") + "."
			}
			return nil, errValor(`No conozco el color "`+nombre+`".`, ln, pista)
		}
		in.tortuga.color = codigo
		return nil, nil
	})
}
