package main

func (a *Armador) instruccion() (Instruccion, *ErrorFal) {
	ln := a.ln()

	if a.come("escribe") {
		if a.tipo() == PFinLinea || a.tipo() == PFinArchivo {
			return &InsEscribe{base{ln}, &ExValor{base{ln}, ""}, true}, nil
		}
		valor, err := a.expresion()
		if err != nil {
			return nil, err
		}
		salto := true
		if a.come("sin") {
			if err := a.exige("salto", `despues de "sin"`); err != nil {
				return nil, err
			}
			salto = false
		}
		return &InsEscribe{base{ln}, valor, salto}, a.finDeLinea()
	}

	if a.es("si") {
		return a.instruccionSi()
	}

	if a.come("mientras") {
		cond, err := a.expresion()
		if err != nil {
			return nil, err
		}
		if err := a.finDeLinea(); err != nil {
			return nil, err
		}
		cuerpo, err := a.bloqueSimple("mientras", ln)
		return &InsMientras{base{ln}, cond, cuerpo}, err
	}

	if a.come("repite") {
		veces, err := a.expresion()
		if err != nil {
			return nil, err
		}
		a.come("veces")
		if err := a.finDeLinea(); err != nil {
			return nil, err
		}
		cuerpo, err := a.bloqueSimple("repite", ln)
		return &InsRepite{base{ln}, veces, cuerpo}, err
	}

	if a.es("para") {
		return a.instruccionPara()
	}

	// "funcion nombre <nl>" define; "funcion con ..." es una sin nombre,
	// y de eso se ocupa valor().
	if a.es("funcion") && !a.esN(1, "con") {
		return a.instruccionFuncion()
	}

	if a.es("tipo") {
		return a.instruccionTipo()
	}

	if a.es("intenta") {
		return a.instruccionIntenta()
	}

	if a.come("comparte") {
		n, err := a.exigeNombre(`despues de "comparte"`, "Por ejemplo:   comparte total")
		if err != nil {
			return nil, err
		}
		nombres := []string{n}
		for a.come("y") {
			n, err = a.exigeNombre(`despues de "y"`, "")
			if err != nil {
				return nil, err
			}
			nombres = append(nombres, n)
		}
		return &InsComparte{base{ln}, nombres}, a.finDeLinea()
	}

	if a.come("devuelve", "retorna") {
		if a.tipo() == PFinLinea || a.tipo() == PFinArchivo {
			return &InsDevuelve{base{ln}, nil}, nil
		}
		valor, err := a.expresion()
		if err != nil {
			return nil, err
		}
		return &InsDevuelve{base{ln}, valor}, a.finDeLinea()
	}

	if a.come("detente") {
		return &InsDetente{base{ln}}, a.finDeLinea()
	}
	if a.come("continua") {
		return &InsContinua{base{ln}}, a.finDeLinea()
	}
	if a.come("relanza") {
		return &InsRelanza{base{ln}}, a.finDeLinea()
	}

	if a.come("falla") {
		msg, err := a.expresion()
		if err != nil {
			return nil, err
		}
		return &InsFalla{base{ln}, msg}, a.finDeLinea()
	}

	if a.come("usa") {
		ruta, err := a.expresion()
		if err != nil {
			return nil, err
		}
		apodo := ""
		if a.come("como") {
			apodo, err = a.exigeNombre(`despues de "como"`, "")
			if err != nil {
				return nil, err
			}
		}
		return &InsUsa{base{ln}, ruta, apodo}, a.finDeLinea()
	}

	if a.es("agrega") || a.es("anade") {
		a.avanzar()
		valor, err := a.comparacion()
		if err != nil {
			return nil, err
		}
		if err := a.exige("a", "para decir a que lista agregar"); err != nil {
			return nil, err
		}
		destino, err := a.suma()
		if err != nil {
			return nil, err
		}
		return &InsAgrega{base{ln}, valor, destino}, a.finDeLinea()
	}

	if a.come("quita") {
		a.come("elemento")
		pos, err := a.posicionHastaDe()
		if err != nil {
			return nil, err
		}
		if err := a.exige("de", "para decir de donde quitar"); err != nil {
			return nil, err
		}
		destino, err := a.suma()
		if err != nil {
			return nil, err
		}
		return &InsQuita{base{ln}, pos, destino}, a.finDeLinea()
	}

	p := a.actual()
	if p.Tipo == PPalabra && reservadas[p.Clave] {
		return nil, nuevoError(
			`No se que hacer con la palabra "`+p.Texto+`" al empezar una linea.`, ln,
			`"`+p.Texto+`" es una palabra del lenguaje. Si querias una variable, usa otro nombre.`,
			ClaseSintaxis)
	}

	// O es "algo es valor", o es una llamada suelta. Se parte en suma()
	// a proposito, para no comerse el "es".
	objetivo, err := a.suma()
	if err != nil {
		return nil, err
	}
	if a.come("es") {
		valor, err := a.expresion()
		if err != nil {
			return nil, err
		}
		switch objetivo.(type) {
		case *ExNombre, *ExDe, *ExElemento:
		default:
			return nil, errValor("Esto no es un sitio donde se pueda guardar algo.", ln,
				"Se guarda asi:   nombre es valor")
		}
		return &InsGuarda{base{ln}, objetivo, valor}, a.finDeLinea()
	}
	if err := a.finDeLinea(); err != nil {
		// Escribir "escribir" en vez de "escribe" es el error mas comun de
		// todos al empezar. Sin esto el mensaje seria "sobra algo al final
		// de la linea", que no lleva a ninguna parte.
		if n, ok := objetivo.(*ExNombre); ok {
			pista := ""
			if bien, hay := seEquivocaCon[n.Nombre]; hay {
				pista = `En Fal se dice "` + bien + `".`
			} else {
				pista = sugerir(n.Nombre, listaReservadas())
			}
			if pista != "" {
				return nil, nuevoError(`No conozco la palabra "`+n.Nombre+`".`,
					ln, pista, ClaseSintaxis)
			}
		}
		return nil, err
	}
	return &InsSuelta{base{ln}, objetivo}, nil
}

// abreSino reconoce "sino" y tambien "si no" escrito en dos palabras.
func (a *Armador) abreSino() bool {
	if a.es("sino") {
		return true
	}
	return a.es("si") && a.esN(1, "no") &&
		(a.mirar(2).Tipo == PFinLinea || a.esN(2, "si"))
}

func (a *Armador) abreSiFalla() bool { return a.es("si") && a.esN(1, "falla") }

func (a *Armador) instruccionSi() (Instruccion, *ErrorFal) {
	ln := a.ln()
	a.avanzar()
	cond, err := a.expresion()
	if err != nil {
		return nil, err
	}
	if err := a.finDeLinea(); err != nil {
		return nil, err
	}
	cuerpo, err := a.bloque(func() bool { return a.es("fin") || a.abreSino() }, "si", ln)
	if err != nil {
		return nil, err
	}

	var sino []Instruccion
	if a.abreSino() {
		if !a.come("sino") {
			a.avanzar() // si
			a.avanzar() // no
		}
		if a.es("si") {
			// "si no si ..." encadenado: el si anidado consume su propio fin.
			anidado, err := a.instruccionSi()
			if err != nil {
				return nil, err
			}
			return &InsSi{base{ln}, cond, cuerpo, []Instruccion{anidado}}, nil
		}
		if err := a.finDeLinea(); err != nil {
			return nil, err
		}
		sino, err = a.bloque(func() bool { return a.es("fin") }, "si no", ln)
		if err != nil {
			return nil, err
		}
	}
	return &InsSi{base{ln}, cond, cuerpo, sino}, a.exige("fin", "para cerrar el si")
}

func (a *Armador) instruccionIntenta() (Instruccion, *ErrorFal) {
	ln := a.ln()
	a.avanzar()
	if err := a.finDeLinea(); err != nil {
		return nil, err
	}
	corta := func() bool { return a.es("fin") || a.abreSiFalla() || a.es("finalmente") }

	cuerpo, err := a.bloque(corta, "intenta", ln)
	if err != nil {
		return nil, err
	}

	var rescates []Rescate
	for a.abreSiFalla() {
		a.avanzar() // si
		a.avanzar() // falla
		var clases []string
		if a.come("de") {
			for {
				p := a.actual()
				nom := clave(p.Texto)
				if !clasesValidas[nom] {
					return nil, nuevoError(`No existe la clase de error "`+p.Texto+`".`, p.Linea,
						"Las clases son: archivo, limite, matematica, nombre, programa, red, sintaxis, tipo, valor",
						ClaseSintaxis)
				}
				a.avanzar()
				clases = append(clases, nom)
				if !a.come("y", "o") {
					break
				}
			}
		}
		if err := a.finDeLinea(); err != nil {
			return nil, err
		}
		bloque, err := a.bloque(corta, "si falla", ln)
		if err != nil {
			return nil, err
		}
		rescates = append(rescates, Rescate{clases, bloque})
	}

	var finalmente []Instruccion
	if a.come("finalmente") {
		if err := a.finDeLinea(); err != nil {
			return nil, err
		}
		finalmente, err = a.bloque(func() bool { return a.es("fin") }, "finalmente", ln)
		if err != nil {
			return nil, err
		}
	}
	return &InsIntenta{base{ln}, cuerpo, rescates, finalmente},
		a.exige("fin", "para cerrar el intenta")
}

func (a *Armador) instruccionPara() (Instruccion, *ErrorFal) {
	ln := a.ln()
	a.avanzar()
	a.come("cada")
	nombre, err := a.exigeNombre(`despues de "para cada"`,
		"Por ejemplo:   para cada fruta en frutas")
	if err != nil {
		return nil, err
	}

	if a.come("desde") {
		desde, err := a.comparacion()
		if err != nil {
			return nil, err
		}
		if err := a.exige("hasta", "para decir donde termina la cuenta"); err != nil {
			return nil, err
		}
		hasta, err := a.comparacion()
		if err != nil {
			return nil, err
		}
		var paso Expresion
		if a.come("de") {
			paso, err = a.comparacion()
			if err != nil {
				return nil, err
			}
		}
		if err := a.finDeLinea(); err != nil {
			return nil, err
		}
		cuerpo, err := a.bloqueSimple("para cada", ln)
		return &InsCuenta{base{ln}, nombre, desde, hasta, paso, cuerpo}, err
	}

	if err := a.exige("en", "para decir que recorrer"); err != nil {
		return nil, err
	}
	col, err := a.expresion()
	if err != nil {
		return nil, err
	}
	if err := a.finDeLinea(); err != nil {
		return nil, err
	}
	cuerpo, err := a.bloqueSimple("para cada", ln)
	return &InsRecorre{base{ln}, nombre, col, cuerpo}, err
}

func (a *Armador) instruccionFuncion() (*InsFuncion, *ErrorFal) {
	ln := a.ln()
	a.avanzar()
	nombre, err := a.exigeNombre(`despues de "funcion"`,
		"Por ejemplo:   funcion saludar con nombre")
	if err != nil {
		return nil, err
	}
	params, err := a.listaDeNombres()
	if err != nil {
		return nil, err
	}
	if err := a.finDeLinea(); err != nil {
		return nil, err
	}
	cuerpo, err := a.bloqueSimple("funcion", ln)
	return &InsFuncion{base{ln}, nombre, params, cuerpo}, err
}

func (a *Armador) instruccionTipo() (Instruccion, *ErrorFal) {
	ln := a.ln()
	a.avanzar()
	escrito := a.actual().Texto
	nombre, err := a.exigeNombre(`despues de "tipo"`,
		"Por ejemplo:   tipo Perro con nombre y edad")
	if err != nil {
		return nil, err
	}
	madre := ""
	if a.come("hereda") {
		a.come("de")
		madre, err = a.exigeNombre(`despues de "hereda de"`,
			"Por ejemplo:   tipo Perro hereda de Animal")
		if err != nil {
			return nil, err
		}
	}
	campos, err := a.listaDeNombres()
	if err != nil {
		return nil, err
	}
	if err := a.finDeLinea(); err != nil {
		return nil, err
	}

	var metodos []*InsFuncion
	for {
		a.saltarLineas()
		if a.tipo() == PFinArchivo {
			return nil, nuevoError(`Falta un "fin" para cerrar el tipo "`+escrito+`".`, ln, "", ClaseSintaxis)
		}
		if a.come("fin") {
			break
		}
		if !a.es("funcion") {
			return nil, nuevoError("Dentro de un tipo solo pueden ir funciones.", a.ln(),
				"Los campos van en la primera linea:   tipo "+escrito+" con a y b", ClaseSintaxis)
		}
		f, err := a.instruccionFuncion()
		if err != nil {
			return nil, err
		}
		metodos = append(metodos, f)
	}
	return &InsTipo{base{ln}, nombre, escrito, madre, campos, metodos}, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		return "-" + string(b)
	}
	return string(b)
}
