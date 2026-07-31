package main

import "sort"

func (in *Interprete) evaluar(e Expresion, mem *Memoria) (Valor, *ErrorFal) {
	switch x := e.(type) {

	case *ExValor:
		return x.V, nil

	case *ExNombre:
		return in.resolverNombre(x.Nombre, mem, x.Ln)

	case *ExMi:
		v, hay := mem.leer("mi")
		if !hay {
			return nil, nuevoError(`"mi" solo se puede usar dentro de un tipo.`, x.Ln, "", ClaseSintaxis)
		}
		return v, nil

	case *ExPadre:
		v, hay := mem.leer("mi")
		if !hay {
			return nil, nuevoError(`"padre" solo se puede usar dentro de un tipo.`, x.Ln, "", ClaseSintaxis)
		}
		objeto := v.(*Objeto)
		tipoActual := objeto.Tipo
		if t, hay := mem.leer("_tipo_actual"); hay {
			tipoActual = t.(*Tipo)
		}
		if tipoActual.Madre == nil {
			return nil, nuevoError(`El tipo "`+tipoActual.Nombre+
				`" no hereda de nadie, asi que no tiene padre.`, x.Ln, "", ClaseNombre)
		}
		return &VistaPadre{objeto, tipoActual.Madre}, nil

	case *ExLista:
		datos := make([]Valor, 0, len(x.Elementos))
		for _, el := range x.Elementos {
			v, err := in.evaluar(el, mem)
			if err != nil {
				return nil, err
			}
			datos = append(datos, v)
		}
		return &Lista{datos}, nil

	case *ExConjunto:
		c := nuevoConjunto()
		for _, el := range x.Elementos {
			v, err := in.evaluar(el, mem)
			if err != nil {
				return nil, err
			}
			c.Pon(v)
		}
		return c, nil

	case *ExDicc:
		if len(x.Elementos)%2 != 0 {
			return nil, errValor("Un diccionario se crea por parejas: clave y valor.", x.Ln,
				`Por ejemplo:   diccionario con "ana" y 25`)
		}
		d := nuevoDicc()
		for i := 0; i < len(x.Elementos); i += 2 {
			k, err := in.evaluar(x.Elementos[i], mem)
			if err != nil {
				return nil, err
			}
			v, err := in.evaluar(x.Elementos[i+1], mem)
			if err != nil {
				return nil, err
			}
			d.Pon(k, v)
		}
		return d, nil

	case *ExFuncionValor:
		if x.Anonima {
			// Se lleva puesta la memoria de donde nacio: eso la convierte
			// en una clausura, capaz de recordar lo de su alrededor.
			return &Funcion{Parametros: x.Parametros, Cuerpo: x.Cuerpo, Entorno: mem}, nil
		}
		if v, hay := mem.leer(x.Nombre); hay {
			if f, ok := v.(*Funcion); ok {
				return f, nil
			}
		}
		if f, hay := in.funciones[x.Nombre]; hay {
			return f, nil
		}
		if _, hay := integradas[x.Nombre]; hay {
			return &Funcion{Nombre: x.Nombre, Integrada: x.Nombre}, nil
		}
		return nil, nuevoError(`No conozco ninguna funcion llamada "`+x.Nombre+`".`, x.Ln,
			sugerir(x.Nombre, in.nombresLlamables()), ClaseNombre)

	case *ExPregunta:
		mensaje := ""
		if x.Mensaje != nil {
			v, err := in.evaluar(x.Mensaje, mem)
			if err != nil {
				return nil, err
			}
			mensaje = textoDe(v)
			if mensaje != "" && mensaje[len(mensaje)-1] != ' ' {
				mensaje += " "
			}
		}
		respuesta, ok := in.preguntar(mensaje)
		if !ok {
			return nil, errValor("Nadie contesto a la pregunta.", x.Ln,
				"El programa se quedo sin respuestas que leer.")
		}
		return respuesta, nil

	case *ExElemento:
		col, err := in.evaluar(x.Coleccion, mem)
		if err != nil {
			return nil, err
		}
		pos, err := in.evaluar(x.Posicion, mem)
		if err != nil {
			return nil, err
		}
		return in.sacarElemento(col, pos, x.Ln)

	case *ExNuevo:
		args, err := in.evaluarTodos(x.Args, mem)
		if err != nil {
			return nil, err
		}
		return in.construir(x.Tipo, args, x.Ln)

	case *ExLlama:
		args, err := in.evaluarTodos(x.Args, mem)
		if err != nil {
			return nil, err
		}
		// Una variable que guarda una funcion se llama igual que todo.
		if v, hay := mem.leer(x.Nombre); hay {
			if f, ok := v.(*Funcion); ok {
				return in.llamar(f, args, nil, x.Ln, nil)
			}
		}
		return in.invocar(x.Nombre, args, x.Ln)

	case *ExDe:
		objeto, err := in.evaluar(x.Objeto, mem)
		if err != nil {
			return nil, err
		}
		extras, err := in.evaluarTodos(x.Extras, mem)
		if err != nil {
			return nil, err
		}
		return in.accesoDe(x.Nombre, objeto, extras, x.Ln, mem)

	case *ExNego:
		v, err := in.evaluar(x.Valor, mem)
		if err != nil {
			return nil, err
		}
		return !esVerdad(v), nil

	case *ExDentro:
		que, err := in.evaluar(x.Que, mem)
		if err != nil {
			return nil, err
		}
		donde, err := in.evaluar(x.Donde, mem)
		if err != nil {
			return nil, err
		}
		return contieneValor(donde, que, x.Ln)

	case *ExCompara:
		izq, err := in.evaluar(x.Izq, mem)
		if err != nil {
			return nil, err
		}
		der, err := in.evaluar(x.Der, mem)
		if err != nil {
			return nil, err
		}
		return comparar(x.Op, izq, der, x.Ln)

	case *ExBinaria:
		// "y" y "o" no evaluan el lado derecho si no hace falta.
		if x.Op == "y" || x.Op == "o" {
			izq, err := in.evaluar(x.Izq, mem)
			if err != nil {
				return nil, err
			}
			if x.Op == "y" && !esVerdad(izq) {
				return false, nil
			}
			if x.Op == "o" && esVerdad(izq) {
				return true, nil
			}
			der, err := in.evaluar(x.Der, mem)
			if err != nil {
				return nil, err
			}
			return esVerdad(der), nil
		}
		izq, err := in.evaluar(x.Izq, mem)
		if err != nil {
			return nil, err
		}
		der, err := in.evaluar(x.Der, mem)
		if err != nil {
			return nil, err
		}
		return aritmetica(x.Op, izq, der, x.Ln)
	}

	return nil, errValor("Expresion desconocida.", e.linea(), "")
}

func (in *Interprete) evaluarTodos(exprs []Expresion, mem *Memoria) ([]Valor, *ErrorFal) {
	if len(exprs) == 0 {
		return nil, nil
	}
	salida := make([]Valor, 0, len(exprs))
	for _, e := range exprs {
		v, err := in.evaluar(e, mem)
		if err != nil {
			return nil, err
		}
		salida = append(salida, v)
	}
	return salida, nil
}

func (in *Interprete) nombresLlamables() []string {
	var salida []string
	for n := range in.funciones {
		salida = append(salida, n)
	}
	for n := range integradas {
		salida = append(salida, n)
	}
	sort.Strings(salida)
	return salida
}

func (in *Interprete) resolverNombre(nombre string, mem *Memoria, ln int) (Valor, *ErrorFal) {
	if v, hay := mem.leer(nombre); hay {
		// Un nombre a secas que apunta a una funcion sin datos se entiende
		// como "llamala". Para quedarte con la funcion en si: funcion <nombre>.
		if f, ok := v.(*Funcion); ok && len(f.Parametros) == 0 && f.Integrada == "" {
			return in.llamar(f, nil, nil, ln, nil)
		}
		return v, nil
	}
	if f, hay := in.funciones[nombre]; hay {
		return in.llamar(f, nil, nil, ln, nil)
	}
	if info, hay := integradas[nombre]; hay && info.min == 0 {
		return in.invocar(nombre, nil, ln)
	}
	if t, hay := in.tipos[nombre]; hay {
		return t, nil
	}
	candidatos := append(mem.nombres(), in.nombresLlamables()...)
	pista := sugerir(nombre, candidatos)
	if pista == "" {
		pista = "Antes de usar algo hay que crearlo:   " + nombre + " es 0"
	}
	return nil, nuevoError(`No conozco nada llamado "`+nombre+`".`, ln, pista, ClaseNombre)
}

// accesoDe resuelve "nombre de objeto": campo, funcion del tipo o funcion normal.
func (in *Interprete) accesoDe(nombre string, objeto Valor, extras []Valor, ln int, mem *Memoria) (Valor, *ErrorFal) {
	switch o := objeto.(type) {

	case *VistaPadre:
		metodo, propietario := o.Tipo.BuscaMetodo(nombre)
		if metodo == nil {
			return nil, nuevoError("El padre ("+o.Tipo.Nombre+`) no tiene ninguna funcion "`+
				nombre+`".`, ln, sugerir(nombre, o.Tipo.NombresMetodos()), ClaseNombre)
		}
		return in.llamar(metodo, extras, o.Objeto, ln, propietario)

	case *Objeto:
		if metodo, propietario := o.Tipo.BuscaMetodo(nombre); metodo != nil {
			return in.llamar(metodo, extras, o, ln, propietario)
		}
		if v, hay := o.Campos[nombre]; hay {
			if len(extras) > 0 {
				return nil, errTipo(`"`+nombre+`" es un campo, no una funcion, asi que no lleva datos.`, ln, "")
			}
			return v, nil
		}
		candidatos := append(append([]string{}, o.Orden...), o.Tipo.NombresMetodos()...)
		return nil, nuevoError("Un "+o.Tipo.Nombre+` no tiene nada llamado "`+nombre+`".`,
			ln, sugerir(nombre, candidatos), ClaseNombre)
	}

	// Una funcion guardada en una variable: "doble de 5" si doble es un dato.
	if mem != nil {
		if v, hay := mem.leer(nombre); hay {
			if f, ok := v.(*Funcion); ok {
				return in.llamar(f, append([]Valor{objeto}, extras...), nil, ln, nil)
			}
		}
	}
	return in.invocar(nombre, append([]Valor{objeto}, extras...), ln)
}

func (in *Interprete) sacarElemento(col, pos Valor, ln int) (Valor, *ErrorFal) {
	switch c := col.(type) {
	case *Dicc:
		if v, hay := c.Dame(pos); hay {
			return v, nil
		}
		var claves []string
		for _, k := range c.Claves {
			claves = append(claves, k)
		}
		pista := sugerir(textoDe(pos), claves)
		if pista == "" {
			pista = `Puedes comprobarlo antes con "contiene".`
		}
		return nil, errValor(`El diccionario no tiene la clave "`+textoDe(pos)+`".`, ln, pista)
	case *Objeto:
		return in.accesoDe(textoDe(pos), c, nil, ln, nil)
	case *Lista:
		i, err := indice(pos, len(c.Datos), ln)
		if err != nil {
			return nil, err
		}
		return c.Datos[i], nil
	case string:
		runas := []rune(c)
		i, err := indice(pos, len(runas), ln)
		if err != nil {
			return nil, err
		}
		return string(runas[i]), nil
	}
	return nil, errTipo("Solo las listas, los textos y los diccionarios tienen elementos, "+
		"y esto es "+nombreTipo(col)+".", ln, "")
}

func (in *Interprete) construir(nombre string, args []Valor, ln int) (Valor, *ErrorFal) {
	tipo, hay := in.tipos[nombre]
	if !hay {
		var claves []string
		for k := range in.tipos {
			claves = append(claves, k)
		}
		sort.Strings(claves)
		pista := sugerir(nombre, claves)
		if pista == "" {
			pista = "Los tipos se crean asi:   tipo " + nombre + " con nombre"
		}
		return nil, nuevoError(`No conozco ningun tipo llamado "`+nombre+`".`, ln, pista, ClaseNombre)
	}

	campos := tipo.Campos()
	objeto := &Objeto{Tipo: tipo, Campos: map[string]Valor{}}
	for _, c := range campos {
		objeto.Pon(c, nil)
	}

	// Si el tipo trae una funcion "crea", ella decide como se rellena.
	if constructor, propietario := tipo.BuscaMetodo("crea"); constructor != nil {
		if _, err := in.llamar(constructor, args, objeto, ln, propietario); err != nil {
			return nil, err
		}
		return objeto, nil
	}

	if len(args) > len(campos) {
		return nil, errTipo(`El tipo "`+tipo.Nombre+`" tiene `+itoa(len(campos))+
			" campo(s), pero le diste "+itoa(len(args))+".", ln, "")
	}
	for i, v := range args {
		objeto.Pon(campos[i], v)
	}
	return objeto, nil
}

func (in *Interprete) invocar(nombre string, args []Valor, ln int) (Valor, *ErrorFal) {
	if f, hay := in.funciones[nombre]; hay {
		return in.llamar(f, args, nil, ln, nil)
	}
	if info, hay := integradas[nombre]; hay {
		if len(args) < info.min || len(args) > info.max {
			cuantos := itoa(info.min)
			if info.min != info.max {
				cuantos = "entre " + itoa(info.min) + " y " + itoa(info.max)
			}
			pista := ""
			if len(args) < info.min {
				pista = `Si encadenaste varios "de", el "con" se lo queda el primero. ` +
					`Usa parentesis:   primero de (` + nombre + ` de texto con ",")`
			}
			return nil, errTipo(`"`+nombre+`" necesita `+cuantos+" dato(s), pero le diste "+
				itoa(len(args))+".", ln, pista)
		}
		return info.fn(in, args, ln)
	}
	pista := sugerir(nombre, in.nombresLlamables())
	if pista == "" {
		pista = "Las funciones se crean asi:   funcion " + nombre + " con dato"
	}
	return nil, nuevoError(`No conozco ninguna funcion llamada "`+nombre+`".`, ln, pista, ClaseNombre)
}

func (in *Interprete) llamar(f *Funcion, args []Valor, mi *Objeto, ln int, tipoActual *Tipo) (Valor, *ErrorFal) {
	if f.Integrada != "" {
		return in.invocar(f.Integrada, args, ln)
	}

	nombre := f.Nombre
	if nombre == "" {
		nombre = "una funcion sin nombre"
	}
	if len(args) != len(f.Parametros) {
		detalle := "no necesita datos"
		if len(f.Parametros) > 0 {
			detalle = "necesita " + itoa(len(f.Parametros)) + " dato(s) (" +
				unirTextos(f.Parametros, ", ") + ")"
		}
		return nil, errTipo(`La funcion "`+nombre+`" `+detalle+", pero le diste "+
			itoa(len(args))+".", ln, "")
	}

	if len(in.pila) >= limiteLlamadas {
		return nil, nuevoError("Se han encadenado mas de "+itoa(limiteLlamadas)+
			" llamadas sin terminar ninguna.", ln,
			`Casi siempre es una funcion que se llama a si misma sin un "si" que la pare.`,
			ClaseLimite)
	}

	// La memoria cuelga de donde nacio la funcion, no de donde se llama:
	// asi una funcion recuerda las variables de su alrededor.
	padre := f.Entorno
	if padre == nil {
		padre = in.global
	}
	local := nuevaMemoria(padre, true)
	if mi != nil {
		local.datos["mi"] = mi
		t := tipoActual
		if t == nil {
			t = mi.Tipo
		}
		local.datos["_tipo_actual"] = t
	}
	for i, p := range f.Parametros {
		local.datos[p] = args[i]
	}

	in.pila = append(in.pila, Marco{nombre, ln})
	r, err := in.ejecutarBloque(f.Cuerpo, local)
	in.pila = in.pila[:len(in.pila)-1]

	if err != nil {
		// Se va apuntando el camino para poder enseñarlo despues.
		err.Pila = append(err.Pila, Marco{nombre, ln})
		return nil, err
	}
	if r.sig == sigDevuelve {
		return r.valor, nil
	}
	return nil, nil
}

func unirTextos(xs []string, sep string) string {
	salida := ""
	for i, x := range xs {
		if i > 0 {
			salida += sep
		}
		salida += x
	}
	return salida
}

func sortStrings(xs []string) { sort.Strings(xs) }
