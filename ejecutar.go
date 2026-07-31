package main

import (
	"os"
	"path/filepath"
	"strings"
)

func (in *Interprete) ejecutar(ins Instruccion, mem *Memoria) (resultado, *ErrorFal) {
	switch x := ins.(type) {

	case *InsEscribe:
		v, err := in.evaluar(x.Valor, mem)
		if err != nil {
			return seguir, err
		}
		in.escribir(textoDe(v), x.Salto)
		return seguir, nil

	case *InsGuarda:
		v, err := in.evaluar(x.Valor, mem)
		if err != nil {
			return seguir, err
		}
		return seguir, in.asignar(x.Objetivo, v, mem)

	case *InsSi:
		cond, err := in.evaluar(x.Condicion, mem)
		if err != nil {
			return seguir, err
		}
		if esVerdad(cond) {
			return in.ejecutarBloque(x.Cuerpo, mem)
		}
		return in.ejecutarBloque(x.Sino, mem)

	case *InsMientras:
		for {
			cond, err := in.evaluar(x.Condicion, mem)
			if err != nil {
				return seguir, err
			}
			if !esVerdad(cond) {
				return seguir, nil
			}
			sigue, r, err := in.bucle(x.Cuerpo, mem)
			if err != nil {
				return seguir, err
			}
			if !sigue {
				return r, nil
			}
		}

	case *InsRepite:
		v, err := in.evaluar(x.Veces, mem)
		if err != nil {
			return seguir, err
		}
		n, ok := v.(Num)
		if !ok {
			return seguir, errTipo(`Despues de "repite" hace falta un numero, pero llego `+
				nombreTipo(v)+".", x.Ln, "")
		}
		for i := int64(0); i < n.Int(); i++ {
			sigue, r, err := in.bucle(x.Cuerpo, mem)
			if err != nil {
				return seguir, err
			}
			if !sigue {
				return r, nil
			}
		}
		return seguir, nil

	case *InsCuenta:
		desde, err := in.numeroDe(x.Desde, mem, "para cada", x.Ln)
		if err != nil {
			return seguir, err
		}
		hasta, err := in.numeroDe(x.Hasta, mem, "para cada", x.Ln)
		if err != nil {
			return seguir, err
		}
		// La direccion se deduce sola: "desde 10 hasta 1" cuenta hacia atras.
		direccion := int64(1)
		if hasta < desde {
			direccion = -1
		}
		paso := direccion
		if x.Paso != nil {
			p, err := in.numeroDe(x.Paso, mem, "para cada", x.Ln)
			if err != nil {
				return seguir, err
			}
			if p < 0 {
				p = -p
			}
			if p == 0 {
				return seguir, errValor("El paso de la cuenta no puede ser 0.", x.Ln, "")
			}
			paso = p * direccion
		}
		for v := desde; (direccion > 0 && v <= hasta) || (direccion < 0 && v >= hasta); v += paso {
			mem.guardar(x.Nombre, Entero(v))
			sigue, r, err := in.bucle(x.Cuerpo, mem)
			if err != nil {
				return seguir, err
			}
			if !sigue {
				return r, nil
			}
		}
		return seguir, nil

	case *InsRecorre:
		col, err := in.evaluar(x.Coleccion, mem)
		if err != nil {
			return seguir, err
		}
		elementos, err := in.elementosDe(col, x.Ln)
		if err != nil {
			return seguir, err
		}
		for _, v := range elementos {
			mem.guardar(x.Nombre, v)
			sigue, r, err := in.bucle(x.Cuerpo, mem)
			if err != nil {
				return seguir, err
			}
			if !sigue {
				return r, nil
			}
		}
		return seguir, nil

	case *InsFuncion, *InsTipo:
		return seguir, in.registrar([]Instruccion{ins})

	case *InsDevuelve:
		if x.Valor == nil {
			return resultado{sigDevuelve, nil}, nil
		}
		v, err := in.evaluar(x.Valor, mem)
		if err != nil {
			return seguir, err
		}
		return resultado{sigDevuelve, v}, nil

	case *InsDetente:
		return resultado{sigDetente, nil}, nil

	case *InsContinua:
		return resultado{sigContinua, nil}, nil

	case *InsComparte:
		for _, n := range x.Nombres {
			mem.declararCompartida(n)
		}
		return seguir, nil

	case *InsFalla:
		v, err := in.evaluar(x.Mensaje, mem)
		if err != nil {
			return seguir, err
		}
		if o, ok := v.(*Objeto); ok && o.Tipo.EsError {
			return seguir, nuevoError(textoDe(o.Campos["mensaje"]), x.Ln, "",
				Clase(textoDe(o.Campos["clase"])))
		}
		return seguir, nuevoError(textoDe(v), x.Ln, "", ClasePrograma)

	case *InsRelanza:
		v, hay := mem.leer("error")
		o, ok := v.(*Objeto)
		if !hay || !ok || !o.Tipo.EsError {
			return seguir, nuevoError(`"relanza" solo vale dentro de un "si falla".`,
				x.Ln, "", ClaseSintaxis)
		}
		linea := x.Ln
		if n, ok := o.Campos["linea"].(Num); ok {
			linea = int(n.Int())
		}
		return seguir, nuevoError(textoDe(o.Campos["mensaje"]), linea,
			textoDe(o.Campos["pista"]), Clase(textoDe(o.Campos["clase"])))

	case *InsIntenta:
		return in.ejecutarIntenta(x, mem)

	case *InsUsa:
		v, err := in.evaluar(x.Ruta, mem)
		if err != nil {
			return seguir, err
		}
		return seguir, in.cargar(textoDe(v), x.Apodo, x.Ln)

	case *InsAgrega:
		destino, err := in.evaluar(x.Destino, mem)
		if err != nil {
			return seguir, err
		}
		valor, err := in.evaluar(x.Valor, mem)
		if err != nil {
			return seguir, err
		}
		switch d := destino.(type) {
		case *Lista:
			d.Datos = append(d.Datos, valor)
			return seguir, nil
		case *Conjunto:
			d.Pon(valor)
			return seguir, nil
		}
		return seguir, errTipo("Solo se puede agregar a una lista o a un conjunto, y esto es "+
			nombreTipo(destino)+".", x.Ln, "Crea la lista antes:   cosas es lista vacia")

	case *InsQuita:
		destino, err := in.evaluar(x.Destino, mem)
		if err != nil {
			return seguir, err
		}
		pos, err := in.evaluar(x.Posicion, mem)
		if err != nil {
			return seguir, err
		}
		switch d := destino.(type) {
		case *Conjunto:
			if !d.Quita(pos) {
				return seguir, errValor("El conjunto no tiene ese elemento.", x.Ln, "")
			}
			return seguir, nil
		case *Dicc:
			if !d.Quita(pos) {
				return seguir, errValor(`El diccionario no tiene la clave "`+textoDe(pos)+`".`, x.Ln, "")
			}
			return seguir, nil
		case *Lista:
			i, err := indice(pos, len(d.Datos), x.Ln)
			if err != nil {
				return seguir, err
			}
			d.Datos = append(d.Datos[:i], d.Datos[i+1:]...)
			return seguir, nil
		}
		return seguir, errTipo("Solo se puede quitar de una lista, un diccionario o un "+
			"conjunto, y esto es "+nombreTipo(destino)+".", x.Ln, "")

	case *InsSuelta:
		v, err := in.evaluar(x.Valor, mem)
		if err != nil {
			return seguir, err
		}
		// Una funcion sola en su linea se entiende como "llamala".
		if f, ok := v.(*Funcion); ok && len(f.Parametros) == 0 && f.Integrada == "" {
			_, err = in.llamar(f, nil, nil, x.Ln, nil)
			return seguir, err
		}
		return seguir, nil
	}

	return seguir, errValor("Instruccion desconocida.", ins.linea(), "")
}

func (in *Interprete) ejecutarIntenta(x *InsIntenta, mem *Memoria) (resultado, *ErrorFal) {
	r, err := in.ejecutarBloque(x.Cuerpo, mem)

	if err != nil {
		atendido := false
		for _, rescate := range x.Rescates {
			if len(rescate.Clases) == 0 || contieneTexto(rescate.Clases, string(err.Clase)) {
				mem.guardar("error", in.objetoError(err))
				atendido = true
				r, err = in.ejecutarBloque(rescate.Cuerpo, mem)
				break
			}
		}
		if !atendido {
			// Aunque nadie lo cace, el "finalmente" tiene que correr.
			if len(x.Finalmente) > 0 {
				if _, err2 := in.ejecutarBloque(x.Finalmente, mem); err2 != nil {
					return seguir, err2
				}
			}
			return seguir, err
		}
	}

	if len(x.Finalmente) > 0 {
		rf, errf := in.ejecutarBloque(x.Finalmente, mem)
		if errf != nil {
			return seguir, errf
		}
		if rf.sig != sigNada {
			return rf, nil
		}
	}
	return r, err
}

func (in *Interprete) asignar(objetivo Expresion, valor Valor, mem *Memoria) *ErrorFal {
	switch x := objetivo.(type) {

	case *ExNombre:
		mem.guardar(x.Nombre, valor)
		return nil

	case *ExElemento:
		destino, err := in.evaluar(x.Coleccion, mem)
		if err != nil {
			return err
		}
		pos, err := in.evaluar(x.Posicion, mem)
		if err != nil {
			return err
		}
		switch d := destino.(type) {
		case *Dicc:
			d.Pon(pos, valor)
			return nil
		case *Lista:
			i, err := indice(pos, len(d.Datos), x.Ln)
			if err != nil {
				return err
			}
			d.Datos[i] = valor
			return nil
		}
		return errTipo("Solo se pueden cambiar elementos de una lista o un diccionario, "+
			"y esto es "+nombreTipo(destino)+".", x.Ln, "")

	case *ExDe:
		destino, err := in.evaluar(x.Objeto, mem)
		if err != nil {
			return err
		}
		switch d := destino.(type) {
		case *Objeto:
			if _, hay := d.Campos[x.Nombre]; !hay {
				return nuevoError("Un "+d.Tipo.Nombre+` no tiene ningun campo llamado "`+
					x.Nombre+`".`, x.Ln, sugerir(x.Nombre, d.Orden), ClaseNombre)
			}
			d.Pon(x.Nombre, valor)
			return nil
		case *Dicc:
			d.Pon(x.Nombre, valor)
			return nil
		}
		return errTipo("Solo los objetos y los diccionarios tienen campos, y esto es "+
			nombreTipo(destino)+".", x.Ln, "")
	}
	return errValor("Aqui no se puede guardar nada.", objetivo.linea(), "")
}

// Ayudas

func (in *Interprete) numeroDe(e Expresion, mem *Memoria, quien string, ln int) (int64, *ErrorFal) {
	v, err := in.evaluar(e, mem)
	if err != nil {
		return 0, err
	}
	n, ok := v.(Num)
	if !ok {
		return 0, errTipo(`"`+quien+`" necesita un numero, pero le llego `+nombreTipo(v)+".", ln, "")
	}
	return n.Int(), nil
}

func (in *Interprete) elementosDe(col Valor, ln int) ([]Valor, *ErrorFal) {
	switch x := col.(type) {
	case *Lista:
		return append([]Valor{}, x.Datos...), nil
	case *Dicc:
		salida := make([]Valor, 0, len(x.Claves))
		for _, k := range x.Claves {
			salida = append(salida, x.Crudas[k])
		}
		return salida, nil
	case *Conjunto:
		return x.Elementos(), nil
	case string:
		var salida []Valor
		for _, r := range x {
			salida = append(salida, string(r))
		}
		return salida, nil
	}
	return nil, errTipo("Solo se pueden recorrer listas, textos, diccionarios y conjuntos, "+
		"y esto es "+nombreTipo(col)+".", ln,
		"Si querias contar, usa:   para cada n desde 1 hasta 10")
}

// indice pasa de la posicion que escribe la persona (empieza en 1, y el -1
// es el ultimo) a la posicion interna (empieza en 0).
func indice(pos Valor, largo int, ln int) (int, *ErrorFal) {
	n, ok := pos.(Num)
	if !ok {
		return 0, errValor("La posicion tiene que ser un numero, pero es "+
			nombreTipo(pos)+".", ln, "")
	}
	p := int(n.Int())
	original := p
	if largo == 0 {
		return 0, errValor("Esta vacio, no tiene ningun elemento.", ln, "")
	}
	if p < 0 {
		p = largo + p + 1
	}
	if p < 1 || p > largo {
		return 0, errValor("Pediste el elemento "+itoa(original)+", pero solo hay "+
			itoa(largo)+".", ln,
			"El primer elemento es el 1 y el ultimo es el "+itoa(largo)+" (o el -1).")
	}
	return p - 1, nil
}

// Otros archivos

func (in *Interprete) cargar(ruta, apodo string, ln int) *ErrorFal {
	if !strings.HasSuffix(ruta, ".fal") {
		ruta += ".fal"
	}
	completa, _ := filepath.Abs(in.rutaDe(ruta))

	if ya, hay := in.cargados[completa]; hay {
		if apodo != "" && ya != nil {
			in.global.datos[apodo] = ya
		}
		return nil
	}

	datos, e := os.ReadFile(completa)
	if e != nil {
		return nuevoError(`No encontre el archivo "`+ruta+`".`, ln,
			`Se busca al lado de tu programa, en "`+in.carpeta+`".`, ClaseArchivo)
	}

	piezas, err := leer(quitarBOM(string(datos)))
	if err != nil {
		return err
	}
	instrucciones, err := armar(piezas)
	if err != nil {
		return err
	}

	carpetaAntes := in.carpeta
	in.carpeta = filepath.Dir(completa)
	defer func() { in.carpeta = carpetaAntes }()

	if apodo == "" {
		in.cargados[completa] = nil
		return in.correr(instrucciones)
	}

	// Con apodo se ejecuta en su propio mundo y lo que salga se empaqueta
	// en un objeto, para que nada choque con lo tuyo.
	globalAntes, funcionesAntes, tiposAntes := in.global, in.funciones, in.tipos
	in.global = nuevaMemoria(nil, false)
	in.funciones = map[string]*Funcion{}
	in.tipos = map[string]*Tipo{"error": in.tipoError}

	errCorrer := in.correr(instrucciones)
	moduloGlobal, moduloFunciones := in.global, in.funciones
	in.global, in.funciones, in.tipos = globalAntes, funcionesAntes, tiposAntes
	if errCorrer != nil {
		return errCorrer
	}

	caja := &Tipo{Nombre: apodo, Metodos: moduloFunciones}
	objeto := &Objeto{Tipo: caja, Campos: map[string]Valor{}}
	var nombres []string
	for n := range moduloGlobal.datos {
		nombres = append(nombres, n)
	}
	sortStrings(nombres)
	for _, n := range nombres {
		objeto.Pon(n, moduloGlobal.datos[n])
	}
	in.cargados[completa] = objeto
	in.global.datos[apodo] = objeto
	return nil
}

func quitarBOM(s string) string { return strings.TrimPrefix(s, "\ufeff") }
