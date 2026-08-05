package main

// La biblioteca: textos y numeros.
//
// Una funcion nueva es una entrada mas en la tabla. Ni el lector ni el
// armador se enteran.

import (
	"math"
	"math/big"
	"math/rand"
	"strings"
	"time"
)

type funcionIntegrada func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal)

type infoIntegrada struct {
	min, max int
	fn       funcionIntegrada
}

var integradas = map[string]infoIntegrada{}

// azarActual es de donde sale todo el azar del lenguaje. Se puede fijar
// con "semilla de N" para que un programa haga siempre lo mismo, que es
// lo que permite probar un juego o repetir un experimento.
var azarActual = rand.New(rand.NewSource(time.Now().UnixNano()))

func integrada(nombre string, min, max int, fn funcionIntegrada) {
	// Registrar dos veces el mismo nombre no daria ningun error: la segunda
	// se comeria a la primera y la funcion perdida solo se echaria de menos
	// mucho despues. Mejor no arrancar.
	if _, repetida := integradas[nombre]; repetida {
		panic("la funcion integrada \"" + nombre + "\" esta registrada dos veces")
	}
	integradas[nombre] = infoIntegrada{min, max, fn}
}

// Comprobaciones de tipo, con mensajes en castellano

func pideTexto(v Valor, ln int, quien string) (string, *ErrorFal) {
	s, ok := v.(string)
	if !ok {
		return "", errTipo(`"`+quien+`" necesita un texto, pero le llego `+nombreTipo(v)+".",
			ln, `Puedes convertirlo con "texto de".`)
	}
	return s, nil
}

func pideNum(v Valor, ln int, quien string) (Num, *ErrorFal) {
	n, ok := v.(Num)
	if !ok {
		return Num{}, errTipo(`"`+quien+`" necesita un numero, pero le llego `+
			nombreTipo(v)+".", ln, "")
	}
	return n, nil
}

func pideEntero(v Valor, ln int, quien string) (int, *ErrorFal) {
	n, err := pideNum(v, ln, quien)
	if err != nil {
		return 0, err
	}
	return int(n.Int()), nil
}

func pideLista(v Valor, ln int, quien string) (*Lista, *ErrorFal) {
	l, ok := v.(*Lista)
	if !ok {
		return nil, errTipo(`"`+quien+`" necesita una lista, pero le llego `+
			nombreTipo(v)+".", ln, "")
	}
	return l, nil
}

// pideSecuencia acepta lo que se puede recorrer en orden: listas y textos.
func pideSecuencia(v Valor, ln int, quien string) ([]Valor, bool, *ErrorFal) {
	switch x := v.(type) {
	case *Lista:
		return x.Datos, false, nil
	case string:
		var salida []Valor
		for _, r := range x {
			salida = append(salida, string(r))
		}
		return salida, true, nil
	}
	return nil, false, errTipo(`"`+quien+`" necesita una lista o un texto, pero le llego `+
		nombreTipo(v)+".", ln, "")
}

func pideFuncion(v Valor, ln int, quien string) (*Funcion, *ErrorFal) {
	f, ok := v.(*Funcion)
	if !ok {
		return nil, errTipo(`"`+quien+`" necesita una funcion, pero le llego `+
			nombreTipo(v)+".", ln,
			"Se pasa asi:   "+quien+" con lista y funcion doble")
	}
	return f, nil
}

func pideConjunto(v Valor, ln int, quien string) (*Conjunto, *ErrorFal) {
	switch x := v.(type) {
	case *Conjunto:
		return x, nil
	case *Lista:
		c := nuevoConjunto()
		for _, e := range x.Datos {
			c.Pon(e)
		}
		return c, nil
	}
	return nil, errTipo(`"`+quien+`" necesita un conjunto, pero le llego `+
		nombreTipo(v)+".", ln, "")
}

func lista(vs []Valor) *Lista { return &Lista{vs} }

func listaDeTextos(xs []string) *Lista {
	salida := make([]Valor, len(xs))
	for i, x := range xs {
		salida[i] = x
	}
	return &Lista{salida}
}

func init() {
	registrarTextos()
	registrarNumeros()
	registrarColecciones()
	registrarOrdenSuperior()
	registrarArchivos()
	registrarDatos()
	registrarFechas()
	registrarSistema()
}

// Textos

func registrarTextos() {
	unTexto := func(nombre string, f func(string) string) {
		integrada(nombre, 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
			s, err := pideTexto(a[0], ln, nombre)
			if err != nil {
				return nil, err
			}
			return f(s), nil
		})
	}
	unTexto("mayusculas", strings.ToUpper)
	unTexto("minusculas", strings.ToLower)
	unTexto("recorta", strings.TrimSpace)
	unTexto("capitaliza", func(s string) string {
		if s == "" {
			return s
		}
		r := []rune(strings.ToLower(s))
		return strings.ToUpper(string(r[0])) + string(r[1:])
	})

	integrada("parte", 2, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		texto, err := pideTexto(a[0], ln, "parte")
		if err != nil {
			return nil, err
		}
		sep, err := pideTexto(a[1], ln, "parte")
		if err != nil {
			return nil, err
		}
		if sep == "" {
			var trozos []string
			for _, r := range texto {
				trozos = append(trozos, string(r))
			}
			return listaDeTextos(trozos), nil
		}
		return listaDeTextos(strings.Split(texto, sep)), nil
	})

	integrada("une", 1, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		l, err := pideLista(a[0], ln, "une")
		if err != nil {
			return nil, err
		}
		sep := ""
		if len(a) > 1 {
			if sep, err = pideTexto(a[1], ln, "une"); err != nil {
				return nil, err
			}
		}
		partes := make([]string, len(l.Datos))
		for i, v := range l.Datos {
			partes[i] = textoDe(v)
		}
		return strings.Join(partes, sep), nil
	})

	integrada("reemplaza", 3, 3, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		s, err := pideTexto(a[0], ln, "reemplaza")
		if err != nil {
			return nil, err
		}
		viejo, err := pideTexto(a[1], ln, "reemplaza")
		if err != nil {
			return nil, err
		}
		nuevo, err := pideTexto(a[2], ln, "reemplaza")
		if err != nil {
			return nil, err
		}
		return strings.ReplaceAll(s, viejo, nuevo), nil
	})

	dosTextos := func(nombre string, f func(a, b string) bool) {
		integrada(nombre, 2, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
			x, err := pideTexto(a[0], ln, nombre)
			if err != nil {
				return nil, err
			}
			y, err := pideTexto(a[1], ln, nombre)
			if err != nil {
				return nil, err
			}
			return f(x, y), nil
		})
	}
	dosTextos("empieza", strings.HasPrefix)
	// Se llamaba "termina", pero ese nombre ya era el de cortar el programa
	// y esta se quedaba sin registrar. Ahora hace pareja con "empieza".
	dosTextos("acaba", strings.HasSuffix)

	integrada("repetido", 2, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		s, err := pideTexto(a[0], ln, "repetido")
		if err != nil {
			return nil, err
		}
		n, err := pideEntero(a[1], ln, "repetido")
		if err != nil {
			return nil, err
		}
		if n < 0 {
			n = 0
		}
		return strings.Repeat(s, n), nil
	})

	integrada("numerico", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		if _, ok := a[0].(Num); ok {
			return true, nil
		}
		s, ok := a[0].(string)
		if !ok {
			return false, nil
		}
		_, bien := new(big.Rat).SetString(strings.TrimSpace(strings.ReplaceAll(s, ",", ".")))
		return bien, nil
	})

	// formato de "Hola {}, tienes {} anios" con nombre y edad
	integrada("formato", 1, 9, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		plantilla, err := pideTexto(a[0], ln, "formato")
		if err != nil {
			return nil, err
		}
		trozos := strings.Split(plantilla, "{}")
		huecos := len(trozos) - 1
		datos := a[1:]
		if huecos != len(datos) {
			return nil, errValor("La plantilla tiene "+itoa(huecos)+" hueco(s) {} pero le diste "+
				itoa(len(datos))+" dato(s).", ln, "")
		}
		var b strings.Builder
		b.WriteString(trozos[0])
		for i, d := range datos {
			b.WriteString(textoDe(d))
			b.WriteString(trozos[i+1])
		}
		return b.String(), nil
	})
}

// Numeros

func registrarNumeros() {
	integrada("numero", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		switch x := a[0].(type) {
		case Num:
			return x, nil
		case bool:
			if x {
				return Entero(1), nil
			}
			return Entero(0), nil
		case string:
			limpio := strings.TrimSpace(strings.ReplaceAll(x, ",", "."))
			if n, ok := NumDesdeTexto(limpio); ok {
				return n, nil
			}
			return nil, errValor(`"`+x+`" no se puede convertir en numero.`, ln,
				`Comprueba lo que se escribio. Puedes usar "numerico de" para saber si vale.`)
		}
		return nil, errTipo("No se puede convertir "+nombreTipo(a[0])+" en numero.", ln, "")
	})

	integrada("texto", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		return textoDe(a[0]), nil
	})

	integrada("redondea", 1, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		n, err := pideNum(a[0], ln, "redondea")
		if err != nil {
			return nil, err
		}
		decimales := 0
		if len(a) > 1 {
			if decimales, err = pideEntero(a[1], ln, "redondea"); err != nil {
				return nil, err
			}
		}
		return redondearNum(n, decimales), nil
	})

	integrada("arriba", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		n, err := pideNum(a[0], ln, "arriba")
		if err != nil {
			return nil, err
		}
		return Entero(int64(math.Ceil(n.Float()))), nil
	})

	integrada("abajo", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		n, err := pideNum(a[0], ln, "abajo")
		if err != nil {
			return nil, err
		}
		return Entero(int64(math.Floor(n.Float()))), nil
	})

	integrada("absoluto", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		n, err := pideNum(a[0], ln, "absoluto")
		if err != nil {
			return nil, err
		}
		if n.Signo() < 0 {
			return negaNum(n), nil
		}
		return n, nil
	})

	integrada("raiz", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		n, err := pideNum(a[0], ln, "raiz")
		if err != nil {
			return nil, err
		}
		if n.Signo() < 0 {
			return nil, nuevoError("No existe la raiz de un numero negativo.", ln, "", ClaseMatematica)
		}
		// Si la raiz es exacta la damos exacta; si no, aproximada.
		f := math.Sqrt(n.Float())
		if r := math.Round(f); r*r == n.Float() && n.EsEnteroExacto() {
			return Entero(int64(r)), nil
		}
		return Flotante(f), nil
	})

	integrada("potencia", 2, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		base, err := pideNum(a[0], ln, "potencia")
		if err != nil {
			return nil, err
		}
		exp, err := pideNum(a[1], ln, "potencia")
		if err != nil {
			return nil, err
		}
		// Con exponente entero se puede hacer exacto.
		if exp.EsEnteroExacto() {
			e := exp.Int()
			if e >= 0 && e < 4096 {
				resultado := Entero(1)
				for i := int64(0); i < e; i++ {
					resultado = multiplicaNum(resultado, base)
				}
				return resultado, nil
			}
		}
		return Flotante(math.Pow(base.Float(), exp.Float())), nil
	})

	extremo := func(nombre string, quieroMayor bool) {
		integrada(nombre, 1, 9, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
			valores := a
			if len(a) == 1 {
				if l, ok := a[0].(*Lista); ok {
					valores = l.Datos
				}
			}
			if len(valores) == 0 {
				return nil, errValor(`"`+nombre+`" necesita al menos un numero.`, ln, "")
			}
			mejor, err := pideNum(valores[0], ln, nombre)
			if err != nil {
				return nil, err
			}
			for _, v := range valores[1:] {
				n, err := pideNum(v, ln, nombre)
				if err != nil {
					return nil, err
				}
				c := comparaNum(n, mejor)
				if (quieroMayor && c > 0) || (!quieroMayor && c < 0) {
					mejor = n
				}
			}
			return mejor, nil
		})
	}
	integrada("semilla", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		n, err := pideNum(a[0], ln, "semilla")
		if err != nil {
			return nil, err
		}
		azarActual = rand.New(rand.NewSource(n.Int()))
		return nil, nil
	})

	extremo("minimo", false)
	extremo("maximo", true)

	integrada("azar", 0, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		switch len(a) {
		case 0:
			r := new(big.Rat).SetFloat64(azarActual.Float64())
			return Racional(r), nil
		case 1:
			n, err := pideEntero(a[0], ln, "azar")
			if err != nil {
				return nil, err
			}
			if n < 1 {
				return nil, errValor(`"azar de" necesita un numero mayor que 0.`, ln, "")
			}
			return Entero(int64(azarActual.Intn(n) + 1)), nil
		}
		x, err := pideEntero(a[0], ln, "azar")
		if err != nil {
			return nil, err
		}
		y, err := pideEntero(a[1], ln, "azar")
		if err != nil {
			return nil, err
		}
		if x > y {
			x, y = y, x
		}
		return Entero(int64(azarActual.Intn(y-x+1) + x)), nil
	})
}

func redondearNum(n Num, decimales int) Num {
	if decimales < 0 {
		decimales = 0
	}
	r := n.Rat()
	// Multiplicamos, redondeamos al entero mas cercano, y dividimos.
	escala := new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimales)), nil))
	escalado := new(big.Rat).Mul(r, escala)
	mitad := big.NewRat(1, 2)
	if escalado.Sign() < 0 {
		mitad = big.NewRat(-1, 2)
	}
	desplazado := new(big.Rat).Add(escalado, mitad)
	entero := new(big.Int).Quo(desplazado.Num(), desplazado.Denom())
	final := new(big.Rat).Quo(new(big.Rat).SetInt(entero), escala)
	return Racional(final)
}
