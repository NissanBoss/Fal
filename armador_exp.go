package main

import "strings"

// Las expresiones, de menos a mas fuerte:
//   o -> y -> no -> comparacion -> mas/menos -> por/entre/resto -> valor

func (a *Armador) expresion() (Expresion, *ErrorFal) { return a.expresionO() }

func (a *Armador) expresionO() (Expresion, *ErrorFal) {
	izq, err := a.expresionY()
	if err != nil {
		return nil, err
	}
	for a.es("o") {
		ln := a.avanzar().Linea
		der, err := a.expresionY()
		if err != nil {
			return nil, err
		}
		izq = &ExBinaria{base{ln}, "o", izq, der}
	}
	return izq, nil
}

func (a *Armador) expresionY() (Expresion, *ErrorFal) {
	izq, err := a.negacion()
	if err != nil {
		return nil, err
	}
	for a.es("y") {
		ln := a.avanzar().Linea
		der, err := a.negacion()
		if err != nil {
			return nil, err
		}
		izq = &ExBinaria{base{ln}, "y", izq, der}
	}
	return izq, nil
}

func (a *Armador) negacion() (Expresion, *ErrorFal) {
	if a.es("no") && !a.esN(1, "es") {
		ln := a.avanzar().Linea
		v, err := a.negacion()
		if err != nil {
			return nil, err
		}
		return &ExNego{base{ln}, v}, nil
	}
	return a.comparacion()
}

func (a *Armador) comparacion() (Expresion, *ErrorFal) {
	izq, err := a.suma()
	if err != nil {
		return nil, err
	}

	negado := false
	if a.es("no") && (a.esN(1, "es") || a.esN(1, "esta")) {
		a.avanzar()
		negado = true
	}

	if a.come("esta") {
		ln := a.ln()
		if err := a.exige("en", `despues de "esta"`); err != nil {
			return nil, err
		}
		der, err := a.suma()
		if err != nil {
			return nil, err
		}
		var nodo Expresion = &ExDentro{base{ln}, izq, der}
		if negado {
			nodo = &ExNego{base{ln}, nodo}
		}
		return nodo, nil
	}

	if !a.es("es") {
		return izq, nil
	}
	ln := a.avanzar().Linea

	op := "=="
	comparador := "" // "mayor" o "menor", para avisar si luego no viene nada
	switch {
	case a.come("mayor"):
		comparador = "mayor"
		if a.come("o") {
			if err := a.exige("igual", "en la comparacion"); err != nil {
				return nil, err
			}
			if err := a.exige("que", "en la comparacion"); err != nil {
				return nil, err
			}
			op = ">="
		} else {
			a.come("que")
			op = ">"
		}
	case a.come("menor"):
		comparador = "menor"
		if a.come("o") {
			if err := a.exige("igual", "en la comparacion"); err != nil {
				return nil, err
			}
			if err := a.exige("que", "en la comparacion"); err != nil {
				return nil, err
			}
			op = "<="
		} else {
			a.come("que")
			op = "<"
		}
	case a.come("igual"):
		a.come("a")
	}

	// "mayor" y "menor" tambien sirven como nombre de variable, asi que
	// "si x es mayor" se puede leer de dos formas. Gana el operador, que es
	// lo que quiere decir casi siempre, pero si detras no queda nada hay que
	// contar la otra manera en vez de soltar un "falta un valor" a secas.
	if comparador != "" && (a.tipo() == PFinLinea || a.tipo() == PFinArchivo) {
		return nil, nuevoError(`Falta decir con que comparar, despues de "`+comparador+`".`, a.ln(),
			`Si te referias a la variable "`+comparador+`", ponla entre parentesis:   `+
				`si x es (`+comparador+`)`, ClaseSintaxis)
	}

	der, err := a.suma()
	if err != nil {
		return nil, err
	}
	var nodo Expresion = &ExCompara{base{ln}, op, izq, der}
	if negado {
		nodo = &ExNego{base{ln}, nodo}
	}
	return nodo, nil
}

func (a *Armador) suma() (Expresion, *ErrorFal) {
	izq, err := a.producto()
	if err != nil {
		return nil, err
	}
	for a.es("mas") || a.es("menos") {
		p := a.avanzar()
		der, err := a.producto()
		if err != nil {
			return nil, err
		}
		izq = &ExBinaria{base{p.Linea}, p.Clave, izq, der}
	}
	return izq, nil
}

func (a *Armador) producto() (Expresion, *ErrorFal) {
	izq, err := a.unario()
	if err != nil {
		return nil, err
	}
	for a.es("por") || a.es("entre") || a.es("resto") {
		p := a.avanzar()
		a.come("de") // "resto de" tambien vale
		der, err := a.unario()
		if err != nil {
			return nil, err
		}
		izq = &ExBinaria{base{p.Linea}, p.Clave, izq, der}
	}
	return izq, nil
}

func (a *Armador) unario() (Expresion, *ErrorFal) {
	if a.es("menos") {
		ln := a.avanzar().Linea
		v, err := a.unario()
		if err != nil {
			return nil, err
		}
		return &ExBinaria{base{ln}, "menos", &ExValor{base{ln}, Entero(0)}, v}, nil
	}
	return a.valor(true)
}

// posicionHastaDe lee una expresion sin dejar que se coma el "de", porque
// en "elemento X de lista" ese "de" es del elemento, no de la posicion.
func (a *Armador) posicionHastaDe() (Expresion, *ErrorFal) {
	a.deReservado++
	e, err := a.suma()
	a.deReservado--
	return e, err
}

func (a *Armador) argumentos() ([]Expresion, *ErrorFal) {
	primero, err := a.comparacion()
	if err != nil {
		return nil, err
	}
	args := []Expresion{primero}
	for a.come("y") {
		siguiente, err := a.comparacion()
		if err != nil {
			return nil, err
		}
		args = append(args, siguiente)
	}
	return args, nil
}

// valor lee un valor suelto.
//
// permitirCon=false se usa para lo que va justo detras de un "de". Asi, en
// "parte de recorta de frase con \" \"", el "con" se lo queda siempre la
// llamada de fuera (parte), que es como se lee en voz alta. Si quieres lo
// contrario, usa parentesis.
func (a *Armador) valor(permitirCon bool) (Expresion, *ErrorFal) {
	p := a.actual()
	ln := p.Linea

	switch p.Tipo {
	case PNumero:
		a.avanzar()
		return &ExValor{base{ln}, p.Num}, nil
	case PTexto:
		a.avanzar()
		return &ExValor{base{ln}, p.Texto}, nil
	case PAbre:
		a.avanzar()
		guardado := a.deReservado
		a.deReservado = 0
		dentro, err := a.expresion()
		a.deReservado = guardado
		if err != nil {
			return nil, err
		}
		if a.tipo() != PCierra {
			return nil, nuevoError("Falta cerrar el parentesis.", ln, "", ClaseSintaxis)
		}
		a.avanzar()
		return dentro, nil
	case PPalabra:
		// sigue abajo
	default:
		return nil, nuevoError("Aqui falta un valor, pero encontre "+a.visto()+".", ln, "", ClaseSintaxis)
	}

	switch p.Clave {
	case "verdadero":
		a.avanzar()
		return &ExValor{base{ln}, true}, nil
	case "falso":
		a.avanzar()
		return &ExValor{base{ln}, false}, nil
	case "nada":
		a.avanzar()
		return &ExValor{base{ln}, nil}, nil
	case "mi":
		a.avanzar()
		return &ExMi{base{ln}}, nil
	case "padre":
		a.avanzar()
		return &ExPadre{base{ln}}, nil
	}

	// Una funcion tambien es un valor:
	//   f es funcion doble              (una que ya existe)
	//   f es funcion con a y b ... fin  (una sin nombre, ahi mismo)
	if p.Clave == "funcion" {
		a.avanzar()
		if a.es("con") || a.tipo() == PFinLinea {
			params, err := a.listaDeNombres()
			if err != nil {
				return nil, err
			}
			if err := a.finDeLinea(); err != nil {
				return nil, err
			}
			cuerpo, err := a.bloqueSimple("funcion", ln)
			if err != nil {
				return nil, err
			}
			return &ExFuncionValor{base{ln}, "", params, cuerpo, true}, nil
		}
		nombre, err := a.exigeNombre(`despues de "funcion"`, "Por ejemplo:   f es funcion doble")
		if err != nil {
			return nil, err
		}
		return &ExFuncionValor{base{ln}, nombre, nil, nil, false}, nil
	}

	if p.Clave == "pregunta" {
		a.avanzar()
		var mensaje Expresion
		if a.tipo() != PFinLinea && a.tipo() != PFinArchivo && a.tipo() != PCierra {
			var err *ErrorFal
			mensaje, err = a.valor(false)
			if err != nil {
				return nil, err
			}
		}
		return &ExPregunta{base{ln}, mensaje}, nil
	}

	if p.Clave == "nuevo" {
		a.avanzar()
		nombre, err := a.exigeNombre(`despues de "nuevo"`, `Por ejemplo:   nuevo Perro con "Toby"`)
		if err != nil {
			return nil, err
		}
		var args []Expresion
		if a.come("con") {
			args, err = a.argumentos()
			if err != nil {
				return nil, err
			}
		}
		return &ExNuevo{base{ln}, nombre, args}, nil
	}

	// Palabras especiales solo por contexto: puedes seguir usandolas como
	// nombres de variables.
	if p.Clave == "lista" && a.esN(1, "con", "vacia", "vacio") {
		a.avanzar()
		if a.come("vacia", "vacio") {
			return &ExLista{base{ln}, nil}, nil
		}
		a.avanzar()
		elems, err := a.argumentos()
		return &ExLista{base{ln}, elems}, err
	}

	if p.Clave == "diccionario" && a.esN(1, "con", "vacio", "vacia") {
		a.avanzar()
		if a.come("vacio", "vacia") {
			return &ExDicc{base{ln}, nil}, nil
		}
		a.avanzar()
		elems, err := a.argumentos()
		return &ExDicc{base{ln}, elems}, err
	}

	if p.Clave == "conjunto" && a.esN(1, "con", "vacio", "vacia") {
		a.avanzar()
		if a.come("vacio", "vacia") {
			return &ExConjunto{base{ln}, nil}, nil
		}
		a.avanzar()
		elems, err := a.argumentos()
		return &ExConjunto{base{ln}, elems}, err
	}

	if p.Clave == "elemento" && !a.esN(1, "es") {
		a.avanzar()
		pos, err := a.posicionHastaDe()
		if err != nil {
			return nil, err
		}
		if err := a.exige("de", "para decir de donde sacar el elemento"); err != nil {
			return nil, err
		}
		col, err := a.valor(false)
		return &ExElemento{base{ln}, pos, col}, err
	}

	if p.Clave == "azar" && a.esN(1, "entre") {
		a.avanzar()
		a.avanzar()
		desde, err := a.comparacion()
		if err != nil {
			return nil, err
		}
		if err := a.exige("y", "para decir hasta que numero"); err != nil {
			return nil, err
		}
		hasta, err := a.comparacion()
		if err != nil {
			return nil, err
		}
		return &ExLlama{base{ln}, "azar", []Expresion{desde, hasta}}, nil
	}

	if reservadas[p.Clave] {
		return nil, nuevoError(
			`Aqui falta un valor, pero encontre la palabra "`+p.Texto+`".`, ln,
			`"`+p.Texto+`" es una palabra del lenguaje y no puede ir aqui.`, ClaseSintaxis)
	}

	// Un nombre suelto, una llamada, o el campo de un objeto.
	nombre := strings.ToLower(a.avanzar().Texto)
	if a.deReservado == 0 && a.come("de") {
		objeto, err := a.valor(false)
		if err != nil {
			return nil, err
		}
		var extras []Expresion
		if permitirCon && a.come("con") {
			extras, err = a.argumentos()
			if err != nil {
				return nil, err
			}
		}
		return &ExDe{base{ln}, nombre, objeto, extras}, nil
	}
	if permitirCon && a.come("con") {
		args, err := a.argumentos()
		return &ExLlama{base{ln}, nombre, args}, err
	}
	return &ExNombre{base{ln}, nombre}, nil
}
