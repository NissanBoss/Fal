package main

// Los valores que existen en Fal y como se enseñan por pantalla.

import (
	"fmt"
	"sort"
	"strings"
)

// Errores

type Clase string

const (
	ClaseMatematica Clase = "matematica"
	ClaseValor      Clase = "valor"
	ClaseTipo       Clase = "tipo"
	ClaseNombre     Clase = "nombre"
	ClaseArchivo    Clase = "archivo"
	ClaseRed        Clase = "red"
	ClasePrograma   Clase = "programa"
	ClaseSintaxis   Clase = "sintaxis"
	ClaseLimite     Clase = "limite"
)

var clasesValidas = map[string]bool{
	"matematica": true, "valor": true, "tipo": true, "nombre": true,
	"archivo": true, "red": true, "programa": true, "sintaxis": true,
	"limite": true,
}

type Marco struct {
	Nombre string
	Linea  int
}

type ErrorFal struct {
	Mensaje string
	Linea   int
	Pista   string
	Clase   Clase
	Pila    []Marco // por donde paso el programa
}

func (e *ErrorFal) Error() string { return e.Mensaje }

func nuevoError(mensaje string, linea int, pista string, clase Clase) *ErrorFal {
	return &ErrorFal{Mensaje: mensaje, Linea: linea, Pista: pista, Clase: clase}
}

func errValor(mensaje string, linea int, pista string) *ErrorFal {
	return nuevoError(mensaje, linea, pista, ClaseValor)
}

func errTipo(mensaje string, linea int, pista string) *ErrorFal {
	return nuevoError(mensaje, linea, pista, ClaseTipo)
}

// Valores

// Un Valor puede ser: nil (nada), bool, string, Num, *Lista, *Dicc,
// *Conjunto, *Funcion, *Tipo, *Objeto o *VistaPadre.
type Valor interface{}

type Lista struct{ Datos []Valor }

// Dicc conserva el orden en que se metieron las claves, para que al
// enseñarlo salga siempre igual y no vaya cambiando.
type Dicc struct {
	Claves []string
	Datos  map[string]Valor
	Crudas map[string]Valor // la clave original, por si era un numero
}

func nuevoDicc() *Dicc {
	return &Dicc{Datos: map[string]Valor{}, Crudas: map[string]Valor{}}
}

func (d *Dicc) Pon(clave Valor, v Valor) {
	k := textoDe(clave)
	if _, hay := d.Datos[k]; !hay {
		d.Claves = append(d.Claves, k)
		d.Crudas[k] = clave
	}
	d.Datos[k] = v
}

func (d *Dicc) Dame(clave Valor) (Valor, bool) {
	v, hay := d.Datos[textoDe(clave)]
	return v, hay
}

func (d *Dicc) Tiene(clave Valor) bool {
	_, hay := d.Datos[textoDe(clave)]
	return hay
}

func (d *Dicc) Quita(clave Valor) bool {
	k := textoDe(clave)
	if _, hay := d.Datos[k]; !hay {
		return false
	}
	delete(d.Datos, k)
	delete(d.Crudas, k)
	for i, c := range d.Claves {
		if c == k {
			d.Claves = append(d.Claves[:i], d.Claves[i+1:]...)
			break
		}
	}
	return true
}

func (d *Dicc) Copia() *Dicc {
	n := nuevoDicc()
	for _, k := range d.Claves {
		n.Pon(d.Crudas[k], d.Datos[k])
	}
	return n
}

// Conjunto: una bolsa sin repetidos ni orden.
type Conjunto struct {
	Datos  map[string]Valor
	Ordena []string
}

func nuevoConjunto() *Conjunto {
	return &Conjunto{Datos: map[string]Valor{}}
}

func (c *Conjunto) Pon(v Valor) {
	k := textoDe(v)
	if _, hay := c.Datos[k]; !hay {
		c.Ordena = append(c.Ordena, k)
	}
	c.Datos[k] = v
}

func (c *Conjunto) Tiene(v Valor) bool {
	_, hay := c.Datos[textoDe(v)]
	return hay
}

func (c *Conjunto) Quita(v Valor) bool {
	k := textoDe(v)
	if _, hay := c.Datos[k]; !hay {
		return false
	}
	delete(c.Datos, k)
	for i, x := range c.Ordena {
		if x == k {
			c.Ordena = append(c.Ordena[:i], c.Ordena[i+1:]...)
			break
		}
	}
	return true
}

// Elementos devuelve el contenido siempre en el mismo orden, para que
// dos ejecuciones del mismo programa enseñen lo mismo.
func (c *Conjunto) Elementos() []Valor {
	claves := append([]string{}, c.Ordena...)
	sort.Strings(claves)
	salida := make([]Valor, 0, len(claves))
	for _, k := range claves {
		salida = append(salida, c.Datos[k])
	}
	return salida
}

// Funcion: en Fal tambien es un valor, se puede guardar y pasar.
type Funcion struct {
	Nombre     string
	Parametros []string
	Cuerpo     []Instruccion
	Entorno    *Memoria // donde nacio: eso la convierte en clausura
	Integrada  string   // si no esta vacio, es una del propio lenguaje
}

type Tipo struct {
	Nombre  string
	Propios []string
	Metodos map[string]*Funcion
	Madre   *Tipo
	EsError bool
}

func (t *Tipo) Campos() []string {
	var salida []string
	if t.Madre != nil {
		salida = append(salida, t.Madre.Campos()...)
	}
	for _, c := range t.Propios {
		if !contieneTexto(salida, c) {
			salida = append(salida, c)
		}
	}
	return salida
}

func (t *Tipo) BuscaMetodo(nombre string) (*Funcion, *Tipo) {
	for actual := t; actual != nil; actual = actual.Madre {
		if f, hay := actual.Metodos[nombre]; hay {
			return f, actual
		}
	}
	return nil, nil
}

func (t *Tipo) NombresMetodos() []string {
	var salida []string
	for actual := t; actual != nil; actual = actual.Madre {
		for n := range actual.Metodos {
			if !contieneTexto(salida, n) {
				salida = append(salida, n)
			}
		}
	}
	return salida
}

type Objeto struct {
	Tipo   *Tipo
	Campos map[string]Valor
	Orden  []string
}

func (o *Objeto) Pon(campo string, v Valor) {
	if _, hay := o.Campos[campo]; !hay {
		o.Orden = append(o.Orden, campo)
	}
	o.Campos[campo] = v
}

// VistaPadre es lo que devuelve "padre" dentro de un tipo que hereda.
type VistaPadre struct {
	Objeto *Objeto
	Tipo   *Tipo
}

func contieneTexto(lista []string, x string) bool {
	for _, v := range lista {
		if v == x {
			return true
		}
	}
	return false
}

// Como se ve cada valor

var interpreteActual *Interprete // solo para que "texto" propio funcione

func textoDe(v Valor) string {
	switch x := v.(type) {
	case nil:
		return "nada"
	case bool:
		if x {
			return "verdadero"
		}
		return "falso"
	case string:
		return x
	case Num:
		return x.Texto()
	case *Lista:
		partes := make([]string, len(x.Datos))
		for i, e := range x.Datos {
			partes[i] = textoDe(e)
		}
		return "[" + strings.Join(partes, ", ") + "]"
	case *Conjunto:
		elems := x.Elementos()
		partes := make([]string, len(elems))
		for i, e := range elems {
			partes[i] = textoDe(e)
		}
		return "{" + strings.Join(partes, ", ") + "}"
	case *Dicc:
		partes := make([]string, 0, len(x.Claves))
		for _, k := range x.Claves {
			partes = append(partes, textoDe(x.Crudas[k])+": "+textoDe(x.Datos[k]))
		}
		return "{" + strings.Join(partes, ", ") + "}"
	case *Objeto:
		if x.Tipo.EsError {
			return textoDe(x.Campos["mensaje"])
		}
		// Si el tipo trae su propia funcion "texto", mandan sus reglas.
		if f, propietario := x.Tipo.BuscaMetodo("texto"); f != nil && interpreteActual != nil {
			if r, err := interpreteActual.llamar(f, nil, x, 0, propietario); err == nil {
				return textoDe(r)
			}
		}
		partes := make([]string, 0, len(x.Orden))
		for _, k := range x.Orden {
			partes = append(partes, k+": "+textoDe(x.Campos[k]))
		}
		return fmt.Sprintf("%s(%s)", x.Tipo.Nombre, strings.Join(partes, ", "))
	case *Funcion:
		n := x.Nombre
		if n == "" {
			n = "sin nombre"
		}
		return "<funcion " + n + ">"
	case *Tipo:
		return "<tipo " + x.Nombre + ">"
	case *VistaPadre:
		return "<padre de " + x.Objeto.Tipo.Nombre + ">"
	}
	return fmt.Sprintf("%v", v)
}

func esVerdad(v Valor) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case string:
		return x != ""
	case Num:
		return !x.EsCero()
	case *Lista:
		return len(x.Datos) > 0
	case *Dicc:
		return len(x.Claves) > 0
	case *Conjunto:
		return len(x.Ordena) > 0
	}
	return true
}

func nombreTipo(v Valor) string {
	switch x := v.(type) {
	case nil:
		return "nada"
	case bool:
		return "un si/no"
	case string:
		return "un texto"
	case Num:
		return "un numero"
	case *Lista:
		return "una lista"
	case *Dicc:
		return "un diccionario"
	case *Conjunto:
		return "un conjunto"
	case *Objeto:
		return "un " + x.Tipo.Nombre
	case *Funcion:
		return "una funcion"
	case *Tipo:
		return "un tipo"
	}
	return "algo raro"
}

// sonIguales compara dos valores. Numeros con numeros, textos con textos.
func sonIguales(a, b Valor) bool {
	na, aEsNum := a.(Num)
	nb, bEsNum := b.(Num)
	if aEsNum && bEsNum {
		return comparaNum(na, nb) == 0
	}
	if aEsNum != bEsNum {
		return false
	}
	switch x := a.(type) {
	case nil:
		return b == nil
	case bool:
		y, ok := b.(bool)
		return ok && x == y
	case string:
		y, ok := b.(string)
		return ok && x == y
	case *Lista:
		y, ok := b.(*Lista)
		if !ok || len(x.Datos) != len(y.Datos) {
			return false
		}
		for i := range x.Datos {
			if !sonIguales(x.Datos[i], y.Datos[i]) {
				return false
			}
		}
		return true
	case *Dicc:
		y, ok := b.(*Dicc)
		if !ok || len(x.Claves) != len(y.Claves) {
			return false
		}
		for _, k := range x.Claves {
			otro, hay := y.Datos[k]
			if !hay || !sonIguales(x.Datos[k], otro) {
				return false
			}
		}
		return true
	case *Conjunto:
		y, ok := b.(*Conjunto)
		if !ok || len(x.Ordena) != len(y.Ordena) {
			return false
		}
		for k := range x.Datos {
			if _, hay := y.Datos[k]; !hay {
				return false
			}
		}
		return true
	}
	return a == b
}

// sugerir busca el nombre conocido mas parecido, para el "¿quisiste decir?".
func sugerir(nombre string, candidatos []string) string {
	mejor, mejorPuntos := "", 0.0
	for _, c := range candidatos {
		p := parecido(clave(nombre), clave(c))
		if p > mejorPuntos {
			mejor, mejorPuntos = c, p
		}
	}
	if mejorPuntos >= 0.6 {
		return `Quizas querias decir "` + mejor + `".`
	}
	return ""
}

// parecido da un numero entre 0 y 1 segun lo que se parezcan dos palabras.
func parecido(a, b string) float64 {
	if a == b {
		return 1
	}
	if a == "" || b == "" {
		return 0
	}
	d := distancia([]rune(a), []rune(b))
	mayor := len([]rune(a))
	if l := len([]rune(b)); l > mayor {
		mayor = l
	}
	return 1 - float64(d)/float64(mayor)
}

// distancia cuenta cuantos cambios hacen falta para pasar de una palabra
// a la otra (insertar, borrar o cambiar una letra).
func distancia(a, b []rune) int {
	fila := make([]int, len(b)+1)
	for j := range fila {
		fila[j] = j
	}
	for i := 1; i <= len(a); i++ {
		anterior := fila[0]
		fila[0] = i
		for j := 1; j <= len(b); j++ {
			guardado := fila[j]
			coste := 1
			if a[i-1] == b[j-1] {
				coste = 0
			}
			fila[j] = minimo3(fila[j]+1, fila[j-1]+1, anterior+coste)
			anterior = guardado
		}
	}
	return fila[len(b)]
}

func minimo3(a, b, c int) int {
	if b < a {
		a = b
	}
	if c < a {
		a = c
	}
	return a
}
