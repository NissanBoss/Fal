package main

import "strings"

var nombresOperacion = map[string]string{
	"mas": "sumar", "menos": "restar", "por": "multiplicar",
	"entre": "dividir", "resto": "sacar el resto de",
}

func aritmetica(op string, a, b Valor, ln int) (Valor, *ErrorFal) {
	if op == "mas" {
		// "mas" tambien junta textos, listas y diccionarios.
		_, aTexto := a.(string)
		_, bTexto := b.(string)
		if aTexto || bTexto {
			return textoDe(a) + textoDe(b), nil
		}
		la, aLista := a.(*Lista)
		lb, bLista := b.(*Lista)
		if aLista && bLista {
			juntas := append(append([]Valor{}, la.Datos...), lb.Datos...)
			return &Lista{juntas}, nil
		}
		da, aDicc := a.(*Dicc)
		db, bDicc := b.(*Dicc)
		if aDicc && bDicc {
			nuevo := da.Copia()
			for _, k := range db.Claves {
				nuevo.Pon(db.Crudas[k], db.Datos[k])
			}
			return nuevo, nil
		}
		ca, aConj := a.(*Conjunto)
		cb, bConj := b.(*Conjunto)
		if aConj && bConj {
			nuevo := nuevoConjunto()
			for _, v := range ca.Elementos() {
				nuevo.Pon(v)
			}
			for _, v := range cb.Elementos() {
				nuevo.Pon(v)
			}
			return nuevo, nil
		}
	}

	na, aNum := a.(Num)
	nb, bNum := b.(Num)
	if !aNum || !bNum {
		malo := a
		if aNum {
			malo = b
		}
		pista := ""
		if op == "mas" {
			pista = `Para unir textos, alguno de los dos tiene que ser un texto:   "Hola " mas nombre`
		} else if _, esTexto := malo.(string); esTexto {
			pista = `Si ese texto contiene un numero, conviertelo antes con "numero de".`
		}
		return nil, errTipo("No se puede "+nombresOperacion[op]+" "+nombreTipo(malo)+".", ln, pista)
	}

	switch op {
	case "mas":
		return sumaNum(na, nb), nil
	case "menos":
		return restaNum(na, nb), nil
	case "por":
		return multiplicaNum(na, nb), nil
	}

	if nb.EsCero() {
		return nil, nuevoError("No se puede dividir entre cero.", ln,
			"Comprueba el valor antes de dividir.", ClaseMatematica)
	}
	if op == "resto" {
		return restoNum(na, nb), nil
	}
	return divideNum(na, nb), nil
}

func comparar(op string, a, b Valor, ln int) (Valor, *ErrorFal) {
	if op == "==" {
		return sonIguales(a, b), nil
	}

	na, aNum := a.(Num)
	nb, bNum := b.(Num)
	if aNum && bNum {
		return resultadoOrden(op, comparaNum(na, nb)), nil
	}

	sa, aTexto := a.(string)
	sb, bTexto := b.(string)
	if aTexto && bTexto {
		// Se comparan sin tildes ni mayusculas, que es como se ordena
		// una lista de palabras en español.
		return resultadoOrden(op, strings.Compare(clave(sa), clave(sb))), nil
	}

	return nil, errTipo("No puedo comparar "+nombreTipo(a)+" con "+nombreTipo(b)+".", ln,
		"Solo se comparan numeros con numeros y textos con textos.")
}

func resultadoOrden(op string, c int) bool {
	switch op {
	case ">":
		return c > 0
	case "<":
		return c < 0
	case ">=":
		return c >= 0
	}
	return c <= 0
}

func contieneValor(donde, que Valor, ln int) (Valor, *ErrorFal) {
	switch d := donde.(type) {
	case string:
		s, ok := que.(string)
		if !ok {
			return nil, errTipo("Dentro de un texto solo se puede buscar otro texto, y esto es "+
				nombreTipo(que)+".", ln, "")
		}
		return strings.Contains(d, s), nil
	case *Lista:
		for _, v := range d.Datos {
			if sonIguales(v, que) {
				return true, nil
			}
		}
		return false, nil
	case *Dicc:
		return d.Tiene(que), nil
	case *Conjunto:
		return d.Tiene(que), nil
	}
	return nil, errTipo(`No se puede mirar dentro de `+nombreTipo(donde)+".", ln, "")
}
