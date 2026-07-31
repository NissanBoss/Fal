package main

// Aritmetica exacta.
//
// Un entero normal mientras quepa en int64, y una fraccion en cuanto
// aparecen decimales. Solo se cae al float cuando no hay manera de ser
// exacto: raices, sobre todo.
//
// Es por lo que 0.1 mas 0.2 da 0.3 y no 0.30000000000000004.

import (
	"math"
	"math/big"
	"strconv"
	"strings"
)

type Num struct {
	i     int64    // valido si !esRat && !esFlt
	rat   *big.Rat // valido si esRat
	flt   float64  // valido si esFlt
	esRat bool
	esFlt bool
}

func Entero(v int64) Num     { return Num{i: v} }
func Flotante(v float64) Num { return Num{flt: v, esFlt: true} }

func Racional(r *big.Rat) Num {
	// Si la fraccion resulta ser un entero que cabe, se baja al caso rapido.
	if r.IsInt() {
		n := r.Num()
		if n.IsInt64() {
			return Num{i: n.Int64()}
		}
	}
	return Num{rat: r, esRat: true}
}

// DesdeTexto convierte "19.99" en una fraccion exacta (1999/100).
func NumDesdeTexto(s string) (Num, bool) {
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		return Num{}, false
	}
	return Racional(r), true
}

func (n Num) Rat() *big.Rat {
	switch {
	case n.esRat:
		return n.rat
	case n.esFlt:
		r := new(big.Rat)
		if math.IsInf(n.flt, 0) || math.IsNaN(n.flt) {
			return r
		}
		r.SetFloat64(n.flt)
		return r
	default:
		return new(big.Rat).SetInt64(n.i)
	}
}

func (n Num) Float() float64 {
	switch {
	case n.esFlt:
		return n.flt
	case n.esRat:
		f, _ := n.rat.Float64()
		return f
	default:
		return float64(n.i)
	}
}

// EsEnteroExacto dice si el numero no tiene parte decimal.
func (n Num) EsEnteroExacto() bool {
	switch {
	case n.esFlt:
		return n.flt == math.Trunc(n.flt) && !math.IsInf(n.flt, 0)
	case n.esRat:
		return n.rat.IsInt()
	default:
		return true
	}
}

// Int devuelve el numero como entero, truncando lo que sobre.
func (n Num) Int() int64 {
	switch {
	case n.esFlt:
		return int64(n.flt)
	case n.esRat:
		q := new(big.Int).Quo(n.rat.Num(), n.rat.Denom())
		if q.IsInt64() {
			return q.Int64()
		}
		if q.Sign() < 0 {
			return math.MinInt64
		}
		return math.MaxInt64
	default:
		return n.i
	}
}

func (n Num) EsCero() bool {
	switch {
	case n.esFlt:
		return n.flt == 0
	case n.esRat:
		return n.rat.Sign() == 0
	default:
		return n.i == 0
	}
}

func (n Num) Signo() int {
	switch {
	case n.esFlt:
		if n.flt > 0 {
			return 1
		} else if n.flt < 0 {
			return -1
		}
		return 0
	case n.esRat:
		return n.rat.Sign()
	default:
		if n.i > 0 {
			return 1
		} else if n.i < 0 {
			return -1
		}
		return 0
	}
}

// DIGITOS es cuantos decimales se enseñan de una fraccion que no termina,
// como 1 entre 3. Es el mismo numero que usan otros lenguajes serios.
const DIGITOS = 28

func (n Num) Texto() string {
	if n.esFlt {
		return textoFlotante(n.flt)
	}
	if !n.esRat {
		return formatearEntero(n.i)
	}
	if n.rat.IsInt() {
		return n.rat.Num().String()
	}
	s := n.rat.FloatString(DIGITOS)
	// Quitamos los ceros de sobra: "0.1000...0" se enseña como "0.1".
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimSuffix(s, ".")
	}
	if s == "" || s == "-" {
		s = "0"
	}
	return s
}

func formatearEntero(v int64) string {
	return new(big.Int).SetInt64(v).String()
}

func textoFlotante(f float64) string {
	if math.IsInf(f, 1) {
		return "infinito"
	}
	if math.IsInf(f, -1) {
		return "menos infinito"
	}
	if math.IsNaN(f) {
		return "no es un numero"
	}
	if f == math.Trunc(f) && math.Abs(f) < 1e15 {
		return formatearEntero(int64(f))
	}
	// 'g' con -1 da la forma mas corta que vuelve a leerse igual.
	return strconv.FormatFloat(f, 'g', -1, 64)
}

// Operaciones

func sumaNum(a, b Num) Num {
	if a.rapido() && b.rapido() {
		if s, ok := sumaSegura(a.i, b.i); ok {
			return Entero(s)
		}
	}
	if a.esFlt || b.esFlt {
		return Flotante(a.Float() + b.Float())
	}
	return Racional(new(big.Rat).Add(a.Rat(), b.Rat()))
}

func restaNum(a, b Num) Num {
	if a.rapido() && b.rapido() {
		if s, ok := restaSegura(a.i, b.i); ok {
			return Entero(s)
		}
	}
	if a.esFlt || b.esFlt {
		return Flotante(a.Float() - b.Float())
	}
	return Racional(new(big.Rat).Sub(a.Rat(), b.Rat()))
}

func multiplicaNum(a, b Num) Num {
	if a.rapido() && b.rapido() {
		if s, ok := multiplicaSegura(a.i, b.i); ok {
			return Entero(s)
		}
	}
	if a.esFlt || b.esFlt {
		return Flotante(a.Float() * b.Float())
	}
	return Racional(new(big.Rat).Mul(a.Rat(), b.Rat()))
}

func divideNum(a, b Num) Num {
	if a.esFlt || b.esFlt {
		return Flotante(a.Float() / b.Float())
	}
	if a.rapido() && b.rapido() && b.i != 0 && a.i%b.i == 0 {
		if !(a.i == math.MinInt64 && b.i == -1) {
			return Entero(a.i / b.i)
		}
	}
	return Racional(new(big.Rat).Quo(a.Rat(), b.Rat()))
}

// restoNum es lo que sobra de una division entera, como 17 resto 5 = 2.
func restoNum(a, b Num) Num {
	if a.rapido() && b.rapido() && b.i != 0 {
		if !(a.i == math.MinInt64 && b.i == -1) {
			return Entero(a.i % b.i)
		}
	}
	if a.esFlt || b.esFlt {
		return Flotante(math.Mod(a.Float(), b.Float()))
	}
	// Para fracciones: a - trunc(a/b)*b
	q := new(big.Rat).Quo(a.Rat(), b.Rat())
	ent := new(big.Int).Quo(q.Num(), q.Denom())
	prod := new(big.Rat).Mul(new(big.Rat).SetInt(ent), b.Rat())
	return Racional(new(big.Rat).Sub(a.Rat(), prod))
}

func negaNum(a Num) Num { return restaNum(Entero(0), a) }

// comparaNum devuelve -1, 0 o 1.
func comparaNum(a, b Num) int {
	if a.rapido() && b.rapido() {
		switch {
		case a.i < b.i:
			return -1
		case a.i > b.i:
			return 1
		}
		return 0
	}
	if a.esFlt || b.esFlt {
		x, y := a.Float(), b.Float()
		switch {
		case x < y:
			return -1
		case x > y:
			return 1
		}
		return 0
	}
	return a.Rat().Cmp(b.Rat())
}

func (n Num) rapido() bool { return !n.esRat && !n.esFlt }

// Sumas con aviso de desbordamiento: si no cabe en int64 pasamos a
// fracciones, que no tienen limite de tamaño.
func sumaSegura(a, b int64) (int64, bool) {
	s := a + b
	if (a > 0 && b > 0 && s < 0) || (a < 0 && b < 0 && s >= 0) {
		return 0, false
	}
	return s, true
}

func restaSegura(a, b int64) (int64, bool) {
	if b == math.MinInt64 {
		return 0, false
	}
	return sumaSegura(a, -b)
}

func multiplicaSegura(a, b int64) (int64, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	p := a * b
	if p/b != a || (a == -1 && b == math.MinInt64) || (b == -1 && a == math.MinInt64) {
		return 0, false
	}
	return p, true
}
