package main

// El armador: junta las piezas sueltas en un arbol de instrucciones.

import (
	"sort"
	"strings"
)

// Palabras del lenguaje: no sirven como nombres de variables.
// Las 82 funciones integradas NO estan aqui a proposito, para que puedas
// tener una variable llamada "numero", "lista" o "suma" sin problema.
var reservadas = map[string]bool{}

// La "y" no esta aqui a proposito, aunque sea operador y separador. Como
// palabra del lenguaje va siempre DETRAS de algo ("si a y b"), y como
// nombre va donde se espera un valor ("y es 5"), asi que no chocan. Los
// casos raros estan en y_test.go.
func init() {
	for _, p := range strings.Fields(`escribe es no si sino fin esta mientras repite
		veces para cada en desde hasta funcion devuelve retorna con de o mas menos
		por entre resto mayor menor igual que verdadero falso nada pregunta agrega
		quita detente continua intenta falla usa tipo nuevo mi comparte finalmente
		relanza hereda padre como`) {
		reservadas[p] = true
	}
}

// Lo que la gente escribe por costumbre de otros sitios o por poner el
// infinitivo. Son los tropiezos de siempre al empezar, asi que vale la
// pena reconocerlos por su nombre en vez de fiarlo al parecido.
var seEquivocaCon = map[string]string{
	"imprime": "escribe", "imprimir": "escribe", "print": "escribe",
	"muestra": "escribe", "mostrar": "escribe", "escribir": "escribe",
	"decir": "escribe", "di": "escribe", "repetir": "repite",
	"mientrasque": "mientras", "sino_si": "si no", "elif": "si no si",
	"else": "si no", "if": "si", "while": "mientras", "for": "para cada",
	"return": "devuelve", "true": "verdadero", "false": "falso",
}

// listaReservadas sirve para el "quizas querias decir" cuando alguien
// escribe una palabra del lenguaje casi bien.
func listaReservadas() []string {
	var salida []string
	for p := range reservadas {
		salida = append(salida, p)
	}
	sort.Strings(salida)
	return salida
}

// El arbol

type Instruccion interface{ linea() int }
type Expresion interface{ linea() int }

type base struct{ Ln int }

func (b base) linea() int { return b.Ln }

// -- instrucciones --
type InsEscribe struct {
	base
	Valor Expresion
	Salto bool
}
type InsGuarda struct {
	base
	Objetivo Expresion
	Valor    Expresion
}
type InsSi struct {
	base
	Condicion Expresion
	Cuerpo    []Instruccion
	Sino      []Instruccion
}
type InsMientras struct {
	base
	Condicion Expresion
	Cuerpo    []Instruccion
}
type InsRepite struct {
	base
	Veces  Expresion
	Cuerpo []Instruccion
}
type InsCuenta struct {
	base
	Nombre             string
	Desde, Hasta, Paso Expresion
	Cuerpo             []Instruccion
}
type InsRecorre struct {
	base
	Nombre    string
	Coleccion Expresion
	Cuerpo    []Instruccion
}
type InsFuncion struct {
	base
	Nombre     string
	Parametros []string
	Cuerpo     []Instruccion
}
type InsTipo struct {
	base
	Nombre  string
	Escrito string
	Madre   string
	Campos  []string
	Metodos []*InsFuncion
}
type InsDevuelve struct {
	base
	Valor Expresion
}
type InsDetente struct{ base }
type InsContinua struct{ base }
type InsFalla struct {
	base
	Mensaje Expresion
}
type InsRelanza struct{ base }
type InsComparte struct {
	base
	Nombres []string
}
type Rescate struct {
	Clases []string
	Cuerpo []Instruccion
}
type InsIntenta struct {
	base
	Cuerpo     []Instruccion
	Rescates   []Rescate
	Finalmente []Instruccion
}
type InsUsa struct {
	base
	Ruta  Expresion
	Apodo string
}
type InsAgrega struct {
	base
	Valor   Expresion
	Destino Expresion
}
type InsQuita struct {
	base
	Posicion Expresion
	Destino  Expresion
}
type InsSuelta struct {
	base
	Valor Expresion
}

// -- expresiones --
type ExValor struct {
	base
	V Valor
}
type ExNombre struct {
	base
	Nombre string
}
type ExMi struct{ base }
type ExPadre struct{ base }
type ExLista struct {
	base
	Elementos []Expresion
}
type ExDicc struct {
	base
	Elementos []Expresion
}
type ExConjunto struct {
	base
	Elementos []Expresion
}
type ExPregunta struct {
	base
	Mensaje Expresion
}
type ExElemento struct {
	base
	Posicion  Expresion
	Coleccion Expresion
}
type ExNuevo struct {
	base
	Tipo string
	Args []Expresion
}
type ExLlama struct {
	base
	Nombre string
	Args   []Expresion
}
type ExDe struct {
	base
	Nombre string
	Objeto Expresion
	Extras []Expresion
}
type ExFuncionValor struct {
	base
	Nombre     string // referencia a una que ya existe
	Parametros []string
	Cuerpo     []Instruccion // funcion sin nombre
	Anonima    bool
}
type ExBinaria struct {
	base
	Op       string
	Izq, Der Expresion
}
type ExCompara struct {
	base
	Op       string
	Izq, Der Expresion
}
type ExDentro struct {
	base
	Que, Donde Expresion
}
type ExNego struct {
	base
	Valor Expresion
}

// El armador

type Armador struct {
	p           []Pieza
	i           int
	deReservado int // mientras sea >0 no se consume ningun "de"
}

func armar(piezas []Pieza) ([]Instruccion, *ErrorFal) {
	a := &Armador{p: piezas}
	return a.programa()
}

func (a *Armador) actual() Pieza { return a.p[a.i] }
func (a *Armador) mirar(n int) Pieza {
	j := a.i + n
	if j >= len(a.p) {
		j = len(a.p) - 1
	}
	return a.p[j]
}
func (a *Armador) tipo() TipoPieza { return a.p[a.i].Tipo }
func (a *Armador) ln() int         { return a.p[a.i].Linea }

func (a *Armador) avanzar() Pieza {
	p := a.p[a.i]
	if a.i < len(a.p)-1 {
		a.i++
	}
	return p
}

func (a *Armador) esN(n int, palabras ...string) bool {
	p := a.mirar(n)
	if p.Tipo != PPalabra {
		return false
	}
	for _, w := range palabras {
		if p.Clave == w {
			return true
		}
	}
	return false
}

func (a *Armador) es(palabras ...string) bool { return a.esN(0, palabras...) }

func (a *Armador) come(palabras ...string) bool {
	if a.es(palabras...) {
		a.avanzar()
		return true
	}
	return false
}

func (a *Armador) visto() string {
	switch a.actual().Tipo {
	case PFinLinea:
		return "el final de la linea"
	case PFinArchivo:
		return "el final del programa"
	}
	return `"` + a.actual().Texto + `"`
}

func (a *Armador) exige(palabra, contexto string) *ErrorFal {
	if a.come(palabra) {
		return nil
	}
	return nuevoError(`Esperaba la palabra "`+palabra+`" `+contexto+
		", pero encontre "+a.visto()+".", a.ln(), "", ClaseSintaxis)
}

func (a *Armador) exigeNombre(contexto, ejemplo string) (string, *ErrorFal) {
	p := a.actual()
	if p.Tipo != PPalabra || reservadas[p.Clave] {
		return "", nuevoError("Hace falta un nombre "+contexto+".", p.Linea, ejemplo, ClaseSintaxis)
	}
	a.avanzar()
	return strings.ToLower(p.Texto), nil
}

func (a *Armador) saltarLineas() {
	for a.tipo() == PFinLinea {
		a.avanzar()
	}
}

func (a *Armador) finDeLinea() *ErrorFal {
	if a.tipo() == PFinLinea || a.tipo() == PFinArchivo {
		return nil
	}
	return nuevoError(`Sobra "`+a.actual().Texto+`" al final de la linea.`, a.ln(),
		"En Fal cada linea es una sola instruccion.", ClaseSintaxis)
}

func (a *Armador) programa() ([]Instruccion, *ErrorFal) {
	var salida []Instruccion
	for {
		a.saltarLineas()
		if a.tipo() == PFinArchivo {
			return salida, nil
		}
		if a.es("fin") {
			return nil, nuevoError(`Encontre un "fin" que no cierra nada.`, a.ln(),
				`Cada "fin" cierra un bloque que empezo antes.`, ClaseSintaxis)
		}
		ins, err := a.instruccion()
		if err != nil {
			return nil, err
		}
		salida = append(salida, ins)
	}
}

func (a *Armador) bloque(esFinal func() bool, queAbre string, lineaApertura int) ([]Instruccion, *ErrorFal) {
	var cuerpo []Instruccion
	for {
		a.saltarLineas()
		if a.tipo() == PFinArchivo {
			return nil, nuevoError(
				`Falta un "fin" para cerrar el "`+queAbre+`" de la linea `+itoa(lineaApertura)+".",
				lineaApertura,
				`Todo bloque termina con la palabra "fin" en su propia linea.`, ClaseSintaxis)
		}
		if esFinal() {
			return cuerpo, nil
		}
		ins, err := a.instruccion()
		if err != nil {
			return nil, err
		}
		cuerpo = append(cuerpo, ins)
	}
}

func (a *Armador) bloqueSimple(queAbre string, ln int) ([]Instruccion, *ErrorFal) {
	cuerpo, err := a.bloque(func() bool { return a.es("fin") }, queAbre, ln)
	if err != nil {
		return nil, err
	}
	return cuerpo, a.exige("fin", "para cerrar el "+queAbre)
}

func (a *Armador) listaDeNombres() ([]string, *ErrorFal) {
	var nombres []string
	if a.come("con") {
		for {
			n, err := a.exigeNombre("aqui", "Por ejemplo:   con a y b")
			if err != nil {
				return nil, err
			}
			nombres = append(nombres, n)
			if !a.come("y") {
				break
			}
		}
	}
	return nombres, nil
}
