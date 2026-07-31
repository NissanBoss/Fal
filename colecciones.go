package main

// La biblioteca: listas, diccionarios, conjuntos y las funciones que
// reciben otras funciones.

import (
	"sort"
	"strings"
)

// Listas, textos, diccionarios y conjuntos

func registrarColecciones() {
	integrada("largo", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		switch x := a[0].(type) {
		case string:
			return Entero(int64(len([]rune(x)))), nil
		case *Lista:
			return Entero(int64(len(x.Datos))), nil
		case *Dicc:
			return Entero(int64(len(x.Claves))), nil
		case *Conjunto:
			return Entero(int64(len(x.Ordena))), nil
		}
		return nil, errTipo("Esto no tiene largo: es "+nombreTipo(a[0])+".", ln,
			"Tienen largo los textos, las listas, los diccionarios y los conjuntos.")
	})

	integrada("contiene", 2, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		return contieneValor(a[0], a[1], ln)
	})

	primeroUltimo := func(nombre string, quieroPrimero bool) {
		integrada(nombre, 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
			datos, _, err := pideSecuencia(a[0], ln, nombre)
			if err != nil {
				return nil, err
			}
			if len(datos) == 0 {
				return nil, errValor("Esta vacio, no tiene "+nombre+" elemento.", ln, "")
			}
			if quieroPrimero {
				return datos[0], nil
			}
			return datos[len(datos)-1], nil
		})
	}
	primeroUltimo("primero", true)
	primeroUltimo("ultimo", false)

	// posicion devuelve donde esta algo empezando en 1, o 0 si no esta.
	integrada("posicion", 2, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		if s, ok := a[0].(string); ok {
			buscado, err := pideTexto(a[1], ln, "posicion")
			if err != nil {
				return nil, err
			}
			i := strings.Index(s, buscado)
			if i < 0 {
				return Entero(0), nil
			}
			return Entero(int64(len([]rune(s[:i])) + 1)), nil
		}
		datos, _, err := pideSecuencia(a[0], ln, "posicion")
		if err != nil {
			return nil, err
		}
		for i, v := range datos {
			if sonIguales(v, a[1]) {
				return Entero(int64(i + 1)), nil
			}
		}
		return Entero(0), nil
	})

	integrada("cuenta", 2, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		if s, ok := a[0].(string); ok {
			buscado, err := pideTexto(a[1], ln, "cuenta")
			if err != nil {
				return nil, err
			}
			if buscado == "" {
				return Entero(0), nil
			}
			return Entero(int64(strings.Count(s, buscado))), nil
		}
		datos, _, err := pideSecuencia(a[0], ln, "cuenta")
		if err != nil {
			return nil, err
		}
		total := int64(0)
		for _, v := range datos {
			if sonIguales(v, a[1]) {
				total++
			}
		}
		return Entero(total), nil
	})

	// trozo va del elemento A al B, ambos incluidos, contando desde 1.
	integrada("trozo", 2, 3, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		datos, esTexto, err := pideSecuencia(a[0], ln, "trozo")
		if err != nil {
			return nil, err
		}
		desde, err := pideEntero(a[1], ln, "trozo")
		if err != nil {
			return nil, err
		}
		hasta := len(datos)
		if len(a) > 2 {
			if hasta, err = pideEntero(a[2], ln, "trozo"); err != nil {
				return nil, err
			}
		}
		if desde < 1 {
			desde = 1
		}
		if hasta > len(datos) {
			hasta = len(datos)
		}
		var recorte []Valor
		if desde <= hasta {
			recorte = datos[desde-1 : hasta]
		}
		if esTexto {
			var b strings.Builder
			for _, v := range recorte {
				b.WriteString(textoDe(v))
			}
			return b.String(), nil
		}
		return lista(append([]Valor{}, recorte...)), nil
	})

	integrada("invierte", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		datos, esTexto, err := pideSecuencia(a[0], ln, "invierte")
		if err != nil {
			return nil, err
		}
		alReves := make([]Valor, len(datos))
		for i, v := range datos {
			alReves[len(datos)-1-i] = v
		}
		if esTexto {
			var b strings.Builder
			for _, v := range alReves {
				b.WriteString(textoDe(v))
			}
			return b.String(), nil
		}
		return lista(alReves), nil
	})

	integrada("ordena", 1, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		l, err := pideLista(a[0], ln, "ordena")
		if err != nil {
			return nil, err
		}
		copia := append([]Valor{}, l.Datos...)

		if len(a) > 1 {
			f, err := pideFuncion(a[1], ln, "ordena")
			if err != nil {
				return nil, err
			}
			claves := make([]Valor, len(copia))
			for i, v := range copia {
				k, err := in.llamar(f, []Valor{v}, nil, ln, nil)
				if err != nil {
					return nil, err
				}
				claves[i] = k
			}
			if err := ordenarCon(copia, claves, ln); err != nil {
				return nil, err
			}
			return lista(copia), nil
		}

		if err := ordenarCon(copia, append([]Valor{}, copia...), ln); err != nil {
			return nil, err
		}
		return lista(copia), nil
	})

	integrada("unicos", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		l, err := pideLista(a[0], ln, "unicos")
		if err != nil {
			return nil, err
		}
		var salida []Valor
		for _, v := range l.Datos {
			repetido := false
			for _, y := range salida {
				if sonIguales(v, y) {
					repetido = true
					break
				}
			}
			if !repetido {
				salida = append(salida, v)
			}
		}
		return lista(salida), nil
	})

	integrada("copia", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		switch x := a[0].(type) {
		case *Lista:
			return lista(append([]Valor{}, x.Datos...)), nil
		case *Dicc:
			return x.Copia(), nil
		case *Conjunto:
			c := nuevoConjunto()
			for _, v := range x.Elementos() {
				c.Pon(v)
			}
			return c, nil
		case *Objeto:
			o := &Objeto{Tipo: x.Tipo, Campos: map[string]Valor{}}
			for _, k := range x.Orden {
				o.Pon(k, x.Campos[k])
			}
			return o, nil
		}
		return a[0], nil
	})

	integrada("inserta", 3, 3, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		l, err := pideLista(a[0], ln, "inserta")
		if err != nil {
			return nil, err
		}
		pos, err := pideEntero(a[1], ln, "inserta")
		if err != nil {
			return nil, err
		}
		if pos < 1 {
			pos = 1
		}
		if pos > len(l.Datos)+1 {
			pos = len(l.Datos) + 1
		}
		l.Datos = append(l.Datos, nil)
		copy(l.Datos[pos:], l.Datos[pos-1:])
		l.Datos[pos-1] = a[2]
		return l, nil
	})

	integrada("elige", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		l, err := pideLista(a[0], ln, "elige")
		if err != nil {
			return nil, err
		}
		if len(l.Datos) == 0 {
			return nil, errValor("No puedo elegir de una lista vacia.", ln, "")
		}
		return l.Datos[azarActual.Intn(len(l.Datos))], nil
	})

	integrada("mezcla", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		l, err := pideLista(a[0], ln, "mezcla")
		if err != nil {
			return nil, err
		}
		copia := append([]Valor{}, l.Datos...)
		azarActual.Shuffle(len(copia), func(i, j int) { copia[i], copia[j] = copia[j], copia[i] })
		return lista(copia), nil
	})

	integrada("rango", 1, 3, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		if len(a) == 1 {
			n, err := pideEntero(a[0], ln, "rango")
			if err != nil {
				return nil, err
			}
			var salida []Valor
			for i := 1; i <= n; i++ {
				salida = append(salida, Entero(int64(i)))
			}
			return lista(salida), nil
		}
		desde, err := pideEntero(a[0], ln, "rango")
		if err != nil {
			return nil, err
		}
		hasta, err := pideEntero(a[1], ln, "rango")
		if err != nil {
			return nil, err
		}
		direccion := 1
		if hasta < desde {
			direccion = -1
		}
		paso := direccion
		if len(a) > 2 {
			p, err := pideEntero(a[2], ln, "rango")
			if err != nil {
				return nil, err
			}
			if p < 0 {
				p = -p
			}
			if p == 0 {
				return nil, errValor(`El paso de "rango" no puede ser 0.`, ln, "")
			}
			paso = p * direccion
		}
		var salida []Valor
		for v := desde; (direccion > 0 && v <= hasta) || (direccion < 0 && v >= hasta); v += paso {
			salida = append(salida, Entero(int64(v)))
		}
		return lista(salida), nil
	})

	integrada("suma", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		l, err := pideLista(a[0], ln, "suma")
		if err != nil {
			return nil, err
		}
		total := Entero(0)
		for _, v := range l.Datos {
			n, err := pideNum(v, ln, "suma")
			if err != nil {
				return nil, err
			}
			total = sumaNum(total, n)
		}
		return total, nil
	})

	integrada("promedio", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		l, err := pideLista(a[0], ln, "promedio")
		if err != nil {
			return nil, err
		}
		if len(l.Datos) == 0 {
			return nil, errValor("No se puede sacar el promedio de una lista vacia.", ln, "")
		}
		total := Entero(0)
		for _, v := range l.Datos {
			n, err := pideNum(v, ln, "promedio")
			if err != nil {
				return nil, err
			}
			total = sumaNum(total, n)
		}
		return divideNum(total, Entero(int64(len(l.Datos)))), nil
	})

	// -- diccionarios --
	integrada("claves", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		d, ok := a[0].(*Dicc)
		if !ok {
			return nil, errTipo(`"claves" necesita un diccionario, pero le llego `+
				nombreTipo(a[0])+".", ln, "")
		}
		var salida []Valor
		for _, k := range d.Claves {
			salida = append(salida, d.Crudas[k])
		}
		return lista(salida), nil
	})

	integrada("valores", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		d, ok := a[0].(*Dicc)
		if !ok {
			return nil, errTipo(`"valores" necesita un diccionario, pero le llego `+
				nombreTipo(a[0])+".", ln, "")
		}
		var salida []Valor
		for _, k := range d.Claves {
			salida = append(salida, d.Datos[k])
		}
		return lista(salida), nil
	})

	// -- conjuntos --
	integrada("conjunto", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		datos, _, err := pideSecuencia(a[0], ln, "conjunto")
		if err != nil {
			return nil, err
		}
		c := nuevoConjunto()
		for _, v := range datos {
			c.Pon(v)
		}
		return c, nil
	})

	operacionConjunto := func(nombre string, f func(a, b *Conjunto) *Conjunto) {
		integrada(nombre, 2, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
			x, err := pideConjunto(a[0], ln, nombre)
			if err != nil {
				return nil, err
			}
			y, err := pideConjunto(a[1], ln, nombre)
			if err != nil {
				return nil, err
			}
			return f(x, y), nil
		})
	}
	operacionConjunto("union", func(x, y *Conjunto) *Conjunto {
		c := nuevoConjunto()
		for _, v := range x.Elementos() {
			c.Pon(v)
		}
		for _, v := range y.Elementos() {
			c.Pon(v)
		}
		return c
	})
	operacionConjunto("interseccion", func(x, y *Conjunto) *Conjunto {
		c := nuevoConjunto()
		for _, v := range x.Elementos() {
			if y.Tiene(v) {
				c.Pon(v)
			}
		}
		return c
	})
	operacionConjunto("diferencia", func(x, y *Conjunto) *Conjunto {
		c := nuevoConjunto()
		for _, v := range x.Elementos() {
			if !y.Tiene(v) {
				c.Pon(v)
			}
		}
		return c
	})
}

// ordenarCon ordena "datos" segun "claves", que tienen que ser todas
// numeros o todas textos. Mantiene el orden original entre iguales.
func ordenarCon(datos, claves []Valor, ln int) *ErrorFal {
	todosNum, todosTexto := true, true
	for _, k := range claves {
		if _, ok := k.(Num); !ok {
			todosNum = false
		}
		if _, ok := k.(string); !ok {
			todosTexto = false
		}
	}
	if !todosNum && !todosTexto {
		return errTipo("Solo puedo ordenar listas de numeros o listas de textos.", ln,
			"Esta lista mezcla varios tipos. Si quieres ordenar objetos, di por que:"+
				"   ordena con lista y funcion edad")
	}
	indices := make([]int, len(datos))
	for i := range indices {
		indices[i] = i
	}
	sort.SliceStable(indices, func(x, y int) bool {
		i, j := indices[x], indices[y]
		if todosNum {
			return comparaNum(claves[i].(Num), claves[j].(Num)) < 0
		}
		return clave(claves[i].(string)) < clave(claves[j].(string))
	})
	original := append([]Valor{}, datos...)
	for nuevo, viejo := range indices {
		datos[nuevo] = original[viejo]
	}
	return nil
}

// Funciones que reciben funciones

func registrarOrdenSuperior() {
	integrada("mapa", 2, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		datos, _, err := pideSecuencia(a[0], ln, "mapa")
		if err != nil {
			return nil, err
		}
		f, err := pideFuncion(a[1], ln, "mapa")
		if err != nil {
			return nil, err
		}
		salida := make([]Valor, 0, len(datos))
		for _, v := range datos {
			r, err := in.llamar(f, []Valor{v}, nil, ln, nil)
			if err != nil {
				return nil, err
			}
			salida = append(salida, r)
		}
		return lista(salida), nil
	})

	integrada("filtra", 2, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		datos, _, err := pideSecuencia(a[0], ln, "filtra")
		if err != nil {
			return nil, err
		}
		f, err := pideFuncion(a[1], ln, "filtra")
		if err != nil {
			return nil, err
		}
		var salida []Valor
		for _, v := range datos {
			r, err := in.llamar(f, []Valor{v}, nil, ln, nil)
			if err != nil {
				return nil, err
			}
			if esVerdad(r) {
				salida = append(salida, v)
			}
		}
		return lista(salida), nil
	})

	integrada("reduce", 2, 3, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		datos, _, err := pideSecuencia(a[0], ln, "reduce")
		if err != nil {
			return nil, err
		}
		f, err := pideFuncion(a[1], ln, "reduce")
		if err != nil {
			return nil, err
		}
		var total Valor
		if len(a) > 2 {
			total = a[2]
		} else if len(datos) > 0 {
			total, datos = datos[0], datos[1:]
		} else {
			return nil, errValor(`"reduce" sobre una lista vacia necesita un valor de partida.`, ln, "")
		}
		for _, v := range datos {
			total, err = in.llamar(f, []Valor{total, v}, nil, ln, nil)
			if err != nil {
				return nil, err
			}
		}
		return total, nil
	})

	prueba := func(nombre string, quePasa func(cuantos, total int) Valor) {
		integrada(nombre, 2, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
			datos, _, err := pideSecuencia(a[0], ln, nombre)
			if err != nil {
				return nil, err
			}
			f, err := pideFuncion(a[1], ln, nombre)
			if err != nil {
				return nil, err
			}
			cuantos := 0
			for _, v := range datos {
				r, err := in.llamar(f, []Valor{v}, nil, ln, nil)
				if err != nil {
					return nil, err
				}
				if esVerdad(r) {
					cuantos++
				}
			}
			return quePasa(cuantos, len(datos)), nil
		})
	}
	prueba("cuenta_si", func(c, t int) Valor { return Entero(int64(c)) })
	prueba("alguno", func(c, t int) Valor { return c > 0 })
	prueba("todos", func(c, t int) Valor { return c == t })
}
