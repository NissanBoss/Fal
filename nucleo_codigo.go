package main

// La traduccion de cada instruccion y cada expresion a codigo maquina.
//
// Es la forma mas sencilla que hay: cada expresion deja su valor en rax (la
// etiqueta) y rdx (el dato), y lo que hay que guardar mientras se calcula la
// otra mitad va a la pila. Sale un codigo lento al lado del de gcc, pero
// cada trozo se puede leer y comprobar por separado.

import (
	"sort"
	"strings"
)

func (c *compiladorNucleo) bloque(instrucciones []Instruccion) *ErrorFal {
	for _, ins := range instrucciones {
		if err := c.instruccion(ins); err != nil {
			return err
		}
	}
	return nil
}

func (c *compiladorNucleo) instruccion(ins Instruccion) *ErrorFal {
	e := c.e
	ln := ins.linea()
	switch x := ins.(type) {
	case *InsEscribe:
		if err := c.expr(x.Valor); err != nil {
			return err
		}
		e.push(rdx)
		e.push(rax)
		salto := int64(0)
		if x.Salto {
			salto = 1
		}
		e.movImm(rdx, salto)
		e.push(rdx)
		e.movImm(rax, etiquetaBooleano)
		e.push(rax)
		e.call(c.funciones["nucleo_escribe"].entrada)
		e.addImm(rsp, 32)

	case *InsGuarda:
		return c.guardar(x)

	case *InsSuelta:
		return c.expr(x.Valor)

	case *InsSi:
		if err := c.condicion(x.Condicion, ln); err != nil {
			return err
		}
		sino, fin := e.nueva(), e.nueva()
		e.jcc(siIgual, sino)
		if err := c.bloque(x.Cuerpo); err != nil {
			return err
		}
		e.jmp(fin)
		e.aqui(sino)
		if err := c.bloque(x.Sino); err != nil {
			return err
		}
		e.aqui(fin)

	case *InsMientras:
		inicio, fin := e.nueva(), e.nueva()
		e.aqui(inicio)
		if err := c.condicion(x.Condicion, ln); err != nil {
			return err
		}
		e.jcc(siIgual, fin)
		if err := c.cuerpoDeBucle(x.Cuerpo, inicio, fin); err != nil {
			return err
		}
		e.jmp(inicio)
		e.aqui(fin)

	case *InsRepite:
		if err := c.expr(x.Veces); err != nil {
			return err
		}
		c.exige(rax, etiquetaNumero, ln, `"repite" necesita un numero de veces.`, "")
		quedan := c.hueco()
		e.guarda(rbp, quedan+8, rdx)
		inicio, fin := e.nueva(), e.nueva()
		e.aqui(inicio)
		e.lee(rdx, rbp, quedan+8)
		e.cmpImm(rdx, 0)
		e.jcc(siMenorOIgual, fin)
		e.subImm(rdx, 1)
		e.guarda(rbp, quedan+8, rdx)
		if err := c.cuerpoDeBucle(x.Cuerpo, inicio, fin); err != nil {
			return err
		}
		e.jmp(inicio)
		e.aqui(fin)

	case *InsCuenta:
		return c.cuenta(x)

	case *InsRecorre:
		return c.recorre(x)

	case *InsDevuelve:
		if c.arriba {
			return nuevoError(`"devuelve" solo tiene sentido dentro de una funcion.`, ln, "", ClaseSintaxis)
		}
		if x.Valor != nil {
			if err := c.expr(x.Valor); err != nil {
				return err
			}
		} else {
			e.movImm(rax, etiquetaNada)
			e.movImm(rdx, 0)
		}
		e.jmp(c.salida)

	case *InsDetente, *InsContinua:
		if len(c.bucles) == 0 {
			return nuevoError("Esto solo se puede usar dentro de un bucle.", ln, "", ClaseSintaxis)
		}
		b := c.bucles[len(c.bucles)-1]
		if _, es := x.(*InsDetente); es {
			e.jmp(b.sale)
		} else {
			e.jmp(b.sigue)
		}

	case *InsFalla:
		if err := c.expr(x.Mensaje); err != nil {
			return err
		}
		e.push(rdx)
		e.push(rax)
		e.movSimbolo(rdx, c.texto(c.donde(ln, "")), 0)
		e.push(rdx)
		e.movImm(rax, etiquetaTexto)
		e.push(rax)
		e.call(c.funciones["nucleo_panico"].entrada)

	case *InsComparte:
		// Ya se tuvo en cuenta al repartir las variables de la funcion.

	case *InsFuncion:
		return nuevoError("Las funciones del nucleo se escriben arriba del todo, no dentro de otro bloque.",
			ln, "", ClaseSintaxis)
	case *InsTipo:
		return noHayTodavia("tipos", ln)
	case *InsIntenta, *InsRelanza:
		return noHayTodavia(`"intenta"`, ln)
	case *InsUsa:
		return noHayTodavia(`"usa": el nucleo se compila de un solo archivo`, ln)
	case *InsAgrega, *InsQuita:
		return noHayTodavia("listas", ln)
	default:
		return noHayTodavia("esa instruccion", ln)
	}
	return nil
}

func (c *compiladorNucleo) cuerpoDeBucle(cuerpo []Instruccion, sigue, sale etiqueta) *ErrorFal {
	c.bucles = append(c.bucles, bucleNucleo{sigue, sale})
	err := c.bloque(cuerpo)
	c.bucles = c.bucles[:len(c.bucles)-1]
	return err
}

// condicion deja las banderas listas para un "salta si es falso" (siIgual).
func (c *compiladorNucleo) condicion(x Expresion, ln int) *ErrorFal {
	if err := c.expr(x); err != nil {
		return err
	}
	c.exige(rax, etiquetaBooleano, ln, "Una condicion tiene que dar verdadero o falso.", "")
	c.e.cmpImm(rdx, 0)
	return nil
}

// cuenta sigue al interprete: la direccion sale sola de desde y hasta, el
// paso se toma sin signo, y cambiar la variable dentro no mueve la cuenta.
func (c *compiladorNucleo) cuenta(x *InsCuenta) *ErrorFal {
	e := c.e
	ln := x.Ln
	if err := c.expr(x.Desde); err != nil {
		return err
	}
	c.exige(rax, etiquetaNumero, ln, `"desde" necesita un numero.`, "")
	e.push(rdx)
	if err := c.expr(x.Hasta); err != nil {
		return err
	}
	c.exige(rax, etiquetaNumero, ln, `"hasta" necesita un numero.`, "")
	hasta := c.hueco()
	e.guarda(rbp, hasta+8, rdx)
	if x.Paso != nil {
		if err := c.expr(x.Paso); err != nil {
			return err
		}
		c.exige(rax, etiquetaNumero, ln, "El paso de la cuenta tiene que ser un numero.", "")
	} else {
		e.movImm(rdx, 1)
	}
	positivo := e.nueva()
	e.cmpImm(rdx, 0)
	e.jcc(siMayorOIgual, positivo)
	e.neg(rdx)
	e.jcc(siDesborda, c.panico(ln, "Ese paso es demasiado grande.", ""))
	e.aqui(positivo)
	e.cmpImm(rdx, 0)
	e.jcc(siIgual, c.panico(ln, "El paso de la cuenta no puede ser 0.", ""))

	e.pop(rbx)
	e.lee(rcx, rbp, hasta+8)
	adelante := e.nueva()
	e.cmp(rcx, rbx)
	e.jcc(siMayorOIgual, adelante)
	e.neg(rdx)
	e.aqui(adelante)
	paso, va := c.hueco(), c.hueco()
	e.guarda(rbp, paso+8, rdx)
	e.guarda(rbp, va+8, rbx)

	mira, atras, cuerpo, sigue, fin := e.nueva(), e.nueva(), e.nueva(), e.nueva(), e.nueva()
	e.aqui(mira)
	e.lee(rdx, rbp, va+8)
	e.lee(rcx, rbp, hasta+8)
	e.lee(rbx, rbp, paso+8)
	e.cmpImm(rbx, 0)
	e.jcc(siMenor, atras)
	e.cmp(rdx, rcx)
	e.jcc(siMayor, fin)
	e.jmp(cuerpo)
	e.aqui(atras)
	e.cmp(rdx, rcx)
	e.jcc(siMenor, fin)
	e.aqui(cuerpo)
	e.movImm(rax, etiquetaNumero)
	c.guardarVariable(clave(x.Nombre))
	if err := c.cuerpoDeBucle(x.Cuerpo, sigue, fin); err != nil {
		return err
	}
	e.aqui(sigue)
	e.lee(rdx, rbp, va+8)
	e.lee(rbx, rbp, paso+8)
	e.add(rdx, rbx)
	// Pasarse del numero mas grande solo puede pasar despues de la ultima
	// vuelta, asi que es el final de la cuenta y no un error.
	e.jcc(siDesborda, fin)
	e.guarda(rbp, va+8, rdx)
	e.jmp(mira)
	e.aqui(fin)
	return nil
}

// recorre da los bytes de un texto o los elementos de una region, uno a uno.
func (c *compiladorNucleo) recorre(x *InsRecorre) *ErrorFal {
	e := c.e
	ln := x.Ln
	if err := c.expr(x.Coleccion); err != nil {
		return err
	}
	coleccion, vuelta := c.hueco(), c.hueco()
	e.guarda(rbp, coleccion, rax)
	e.guarda(rbp, coleccion+8, rdx)
	e.call(c.rtLargo)
	c.exige(rax, etiquetaNumero, ln, `"para cada ... en" recorre un texto o una region.`,
		"En el nucleo todavia no hay listas.")
	e.movImm(rdx, 0)
	e.guarda(rbp, vuelta+8, rdx)

	inicio, fin := e.nueva(), e.nueva()
	e.aqui(inicio)
	e.lee(rax, rbp, coleccion)
	e.lee(rdx, rbp, coleccion+8)
	e.call(c.rtLargo)
	e.lee(rcx, rbp, vuelta+8)
	e.addImm(rcx, 1)
	e.cmp(rcx, rdx)
	e.jcc(siMayor, fin)
	e.guarda(rbp, vuelta+8, rcx)
	e.mov(rdx, rcx)
	e.lee(rcx, rbp, coleccion)
	e.lee(rbx, rbp, coleccion+8)
	e.movImm(rax, etiquetaNumero)
	e.call(c.rtLugar)
	e.call(c.rtLee)
	c.guardarVariable(clave(x.Nombre))
	if err := c.cuerpoDeBucle(x.Cuerpo, inicio, fin); err != nil {
		return err
	}
	e.jmp(inicio)
	e.aqui(fin)
	return nil
}

// Variables

// variable dice donde vive un nombre: en el marco (desp) o entre las
// globales (global >= 0).
func (c *compiladorNucleo) variable(nombre string) (desp int32, global int, hay bool) {
	if !c.arriba {
		if d, hay := c.locales[nombre]; hay {
			return d, -1, true
		}
	}
	if g, hay := c.globales[nombre]; hay {
		return 0, g, true
	}
	return 0, -1, false
}

func (c *compiladorNucleo) cargarVariable(desp int32, global int) {
	if global >= 0 {
		c.e.movSimbolo(rbx, "globales", int64(16*global))
		c.e.lee(rax, rbx, 0)
		c.e.lee(rdx, rbx, 8)
		return
	}
	c.e.lee(rax, rbp, desp)
	c.e.lee(rdx, rbp, desp+8)
}

// guardarVariable guarda rax y rdx. El nombre siempre existe: se repartieron
// todos antes de compilar.
func (c *compiladorNucleo) guardarVariable(nombre string) {
	desp, global, _ := c.variable(nombre)
	if global >= 0 {
		c.e.movSimbolo(rbx, "globales", int64(16*global))
		c.e.guarda(rbx, 0, rax)
		c.e.guarda(rbx, 8, rdx)
		return
	}
	c.e.guarda(rbp, desp, rax)
	c.e.guarda(rbp, desp+8, rdx)
}

func (c *compiladorNucleo) guardar(x *InsGuarda) *ErrorFal {
	e := c.e
	ln := x.Ln
	switch obj := x.Objetivo.(type) {
	case *ExNombre:
		if err := c.expr(x.Valor); err != nil {
			return err
		}
		c.guardarVariable(clave(obj.Nombre))
		return nil

	case *ExElemento:
		if err := c.expr(x.Valor); err != nil {
			return err
		}
		e.push(rdx)
		e.push(rax)
		if err := c.expr(obj.Coleccion); err != nil {
			return err
		}
		e.push(rdx)
		e.push(rax)
		if err := c.expr(obj.Posicion); err != nil {
			return err
		}
		e.pop(rcx)
		e.pop(rbx)
		e.cmpImm(rcx, etiquetaTexto)
		e.jcc(siIgual, c.panico(ln, "Los textos del nucleo no se pueden cambiar.",
			"Estan escritos dentro del programa. Para cambiar bytes, usa una region."))
		e.call(c.rtLugar)
		c.erroresDeLugar(ln)
		e.pop(rax)
		e.pop(rdx)
		c.exige(rax, etiquetaNumero, ln, "En una region solo se guardan numeros.", "")
		e.call(c.rtPon)
		e.cmpImm(rax, 0)
		e.jcc(siDistinto, c.panico(ln, "Ese numero no cabe en un elemento de esta region.",
			"Un elemento de 1 byte guarda de 0 a 255; de 2, hasta 65535; de 4, hasta 4294967295. "+
				"Los de 8 guardan cualquier numero."))
		return nil
	}
	return noHayTodavia("objetos con campos", ln)
}

func (c *compiladorNucleo) erroresDeLugar(ln int) {
	e := c.e
	e.cmpImm(rax, 1)
	e.jcc(siIgual, c.panico(ln, "Solo se pueden sacar elementos de un texto o de una region.", ""))
	e.cmpImm(rax, 2)
	e.jcc(siIgual, c.panico(ln, "La posicion de un elemento tiene que ser un numero.", ""))
	e.cmpImm(rax, 3)
	e.jcc(siIgual, c.panico(ln, "No hay ningun elemento en esa posicion.",
		"Se cuenta desde el 1, y con menos 1 llegas al ultimo. Mira el largo antes."))
}

// Expresiones

func (c *compiladorNucleo) expr(x Expresion) *ErrorFal {
	e := c.e
	ln := x.linea()
	switch x := x.(type) {
	case *ExValor:
		return c.constante(x.V, ln)

	case *ExNombre:
		nombre := clave(x.Nombre)
		if desp, global, hay := c.variable(nombre); hay {
			c.cargarVariable(desp, global)
			return nil
		}
		// Un nombre suelto que es una funcion se llama, como en el interprete.
		if _, hay := c.funciones[nombre]; hay {
			return c.llamada(nombre, nil, ln)
		}
		if _, hay := integradasNucleo[nombre]; hay {
			return c.llamada(nombre, nil, ln)
		}
		return nuevoError(`No hay nada que se llame "`+nombre+`".`, ln,
			"Si es una variable, guardale algo antes:   "+nombre+" es 0", ClaseNombre)

	case *ExBinaria:
		switch x.Op {
		case "y", "o":
			if err := c.expr(x.Izq); err != nil {
				return err
			}
			c.exige(rax, etiquetaBooleano, ln, `"`+x.Op+`" une condiciones, y esto no es verdadero ni falso.`, "")
			fin := e.nueva()
			e.cmpImm(rdx, 0)
			if x.Op == "y" {
				e.jcc(siIgual, fin)
			} else {
				e.jcc(siDistinto, fin)
			}
			if err := c.expr(x.Der); err != nil {
				return err
			}
			c.exige(rax, etiquetaBooleano, ln, `"`+x.Op+`" une condiciones, y esto no es verdadero ni falso.`, "")
			e.aqui(fin)
			return nil
		}
		return c.aritmetica(x.Op, x.Izq, x.Der, ln)

	case *ExCompara:
		return c.compara(x.Op, x.Izq, x.Der, ln)

	case *ExNego:
		if err := c.expr(x.Valor); err != nil {
			return err
		}
		c.exige(rax, etiquetaBooleano, ln, `"no" necesita verdadero o falso.`, "")
		e.xorImm(rdx, 1)
		return nil

	case *ExElemento:
		if err := c.expr(x.Coleccion); err != nil {
			return err
		}
		e.push(rdx)
		e.push(rax)
		if err := c.expr(x.Posicion); err != nil {
			return err
		}
		e.pop(rcx)
		e.pop(rbx)
		e.call(c.rtLugar)
		c.erroresDeLugar(ln)
		e.call(c.rtLee)
		return nil

	case *ExLlama:
		return c.llamada(clave(x.Nombre), x.Args, ln)

	case *ExDe:
		return c.llamada(clave(x.Nombre), append([]Expresion{x.Objeto}, x.Extras...), ln)

	case *ExLista:
		return noHayTodavia("listas", ln)
	case *ExDicc:
		return noHayTodavia("diccionarios", ln)
	case *ExConjunto:
		return noHayTodavia("conjuntos", ln)
	case *ExPregunta:
		return noHayTodavia(`"pregunta": el teclado lo tiene que leer el propio sistema`, ln)
	case *ExNuevo, *ExMi, *ExPadre:
		return noHayTodavia("tipos ni objetos", ln)
	case *ExFuncionValor:
		return noHayTodavia("funciones guardadas en variables", ln)
	case *ExDentro:
		return noHayTodavia(`"esta en"`, ln)
	}
	return noHayTodavia("esa expresion", ln)
}

func (c *compiladorNucleo) constante(v Valor, ln int) *ErrorFal {
	e := c.e
	switch v := v.(type) {
	case Num:
		if v.esRat || v.esFlt {
			return nuevoError("En el nucleo solo hay numeros enteros.", ln,
				"Los decimales exactos de Fal necesitan pedir memoria, y el nucleo todavia no sabe.",
				ClaseValor)
		}
		e.movImm(rax, etiquetaNumero)
		e.movImm(rdx, v.i)
	case string:
		e.movImm(rax, etiquetaTexto)
		e.movSimbolo(rdx, c.texto(v), 0)
	case bool:
		e.movImm(rax, etiquetaBooleano)
		if v {
			e.movImm(rdx, 1)
		} else {
			e.movImm(rdx, 0)
		}
	case nil:
		e.movImm(rax, etiquetaNada)
		e.movImm(rdx, 0)
	default:
		return noHayTodavia("ese tipo de valor", ln)
	}
	return nil
}

// dosValores deja el de la izquierda en rax/rdx y el de la derecha en
// rcx/rbx.
func (c *compiladorNucleo) dosValores(izq, der Expresion) *ErrorFal {
	e := c.e
	if err := c.expr(izq); err != nil {
		return err
	}
	e.push(rdx)
	e.push(rax)
	if err := c.expr(der); err != nil {
		return err
	}
	e.mov(rcx, rax)
	e.mov(rbx, rdx)
	e.pop(rax)
	e.pop(rdx)
	return nil
}

func (c *compiladorNucleo) aritmetica(op string, izq, der Expresion, ln int) *ErrorFal {
	e := c.e
	switch op {
	case "mas", "menos", "por", "entre", "resto", "cociente":
	default:
		return noHayTodavia(`"`+op+`"`, ln)
	}
	if err := c.dosValores(izq, der); err != nil {
		return err
	}
	pista := ""
	if op == "mas" {
		pista = "En el nucleo no se pueden unir textos: no hay memoria donde dejar el resultado."
	}
	c.exige(rax, etiquetaNumero, ln, `"`+op+`" necesita dos numeros, y el primero no lo es.`, pista)
	c.exige(rcx, etiquetaNumero, ln, `"`+op+`" necesita dos numeros, y el segundo no lo es.`, pista)

	desborda := c.panico(ln, "El resultado no cabe en un numero del nucleo.",
		"Los numeros del nucleo van de menos 9223372036854775808 a 9223372036854775807.")
	switch op {
	case "mas":
		e.add(rdx, rbx)
		e.jcc(siDesborda, desborda)
	case "menos":
		e.sub(rdx, rbx)
		e.jcc(siDesborda, desborda)
	case "por":
		e.imul(rdx, rbx)
		e.jcc(siDesborda, desborda)
	default:
		c.division(op, ln, desborda)
	}
	return nil
}

// division divide rdx entre rbx. El caso de menos 1 va aparte porque el
// numero mas pequeño entre menos 1 no cabe, y el procesador no avisa con un
// desborde sino con una excepcion que, sin nadie que la atienda, reinicia la
// maquina.
func (c *compiladorNucleo) division(op string, ln int, desborda etiqueta) {
	e := c.e
	e.cmpImm(rbx, 0)
	e.jcc(siIgual, c.panico(ln, "No se puede dividir entre cero.", ""))
	menosUno, fin := e.nueva(), e.nueva()
	e.cmpImm(rbx, -1)
	e.jcc(siIgual, menosUno)
	e.mov(rax, rdx)
	e.cqo()
	e.idiv(rbx) // rax el cociente, rdx el resto
	switch op {
	case "entre":
		e.cmpImm(rdx, 0)
		e.jcc(siDistinto, c.panico(ln, "La division no da exacta, y en el nucleo no hay decimales.",
			"Usa  cociente con a y b  para la parte entera y  a resto b  para lo que sobra."))
		e.mov(rdx, rax)
	case "cociente":
		e.mov(rdx, rax)
	}
	e.jmp(fin)
	e.aqui(menosUno)
	if op == "resto" {
		e.movImm(rdx, 0)
	} else {
		e.neg(rdx)
		e.jcc(siDesborda, desborda)
	}
	e.aqui(fin)
	e.movImm(rax, etiquetaNumero)
}

var comoSeDice = map[string]string{
	">": "mayor que", ">=": "mayor o igual que", "<": "menor que", "<=": "menor o igual que",
}

func (c *compiladorNucleo) compara(op string, izq, der Expresion, ln int) *ErrorFal {
	e := c.e
	if err := c.dosValores(izq, der); err != nil {
		return err
	}
	if op == "==" {
		c.iguales()
		return nil
	}
	mensaje := `"es ` + comoSeDice[op] + `" compara numeros, y esto no lo es.`
	c.exige(rax, etiquetaNumero, ln, mensaje, "")
	c.exige(rcx, etiquetaNumero, ln, mensaje, "")
	// Se salta por la condicion contraria para dejar el 1 solo si se cumple.
	contraria := map[string]byte{">": siMenorOIgual, ">=": siMenor, "<": siMayorOIgual, "<=": siMayor}[op]
	fin := e.nueva()
	e.cmp(rdx, rbx)
	e.movImm(rdx, 0)
	e.jcc(contraria, fin)
	e.movImm(rdx, 1)
	e.aqui(fin)
	e.movImm(rax, etiquetaBooleano)
	return nil
}

// iguales compara dos valores de cualquier tipo. De tipos distintos nunca
// son iguales, igual que en el interprete. Dos textos se comparan letra a
// letra en el preludio; dos regiones son iguales si cubren la misma memoria.
func (c *compiladorNucleo) iguales() {
	e := c.e
	falso, textos, fin, listo := e.nueva(), e.nueva(), e.nueva(), e.nueva()
	e.cmp(rax, rcx)
	e.jcc(siDistinto, falso)
	e.cmpImm(rax, etiquetaTexto)
	e.jcc(siIgual, textos)
	e.cmp(rdx, rbx)
	e.jcc(siDistinto, falso)
	e.movImm(rdx, 1)
	e.jmp(fin)
	e.aqui(falso)
	e.movImm(rdx, 0)
	e.jmp(fin)
	e.aqui(textos)
	e.push(rdx)
	e.push(rax)
	e.push(rbx)
	e.push(rcx)
	e.call(c.funciones["nucleo_textos_iguales"].entrada)
	e.addImm(rsp, 32)
	e.jmp(listo)
	e.aqui(fin)
	e.movImm(rax, etiquetaBooleano)
	e.aqui(listo)
}

func (c *compiladorNucleo) llamada(nombre string, args []Expresion, ln int) *ErrorFal {
	e := c.e
	if cuantos, hay := integradasNucleo[nombre]; hay {
		if strings.HasPrefix(nombre, "nucleo_") && !c.enPreludio {
			return c.nombrePermitido(nombre, ln, false)
		}
		if len(args) < cuantos[0] || len(args) > cuantos[1] {
			pide := itoa(cuantos[0])
			if cuantos[1] != cuantos[0] {
				pide += " o " + itoa(cuantos[1])
			}
			return nuevoError(`"`+nombre+`" necesita `+pide+` datos y le das `+itoa(len(args))+`.`,
				ln, "", ClaseValor)
		}
		return c.integrada(nombre, args, ln)
	}

	f, hay := c.funciones[nombre]
	if !hay {
		if _, _, esVariable := c.variable(nombre); esVariable {
			return nuevoError(`"`+nombre+`" es una variable, no una funcion.`, ln, "", ClaseTipo)
		}
		var nombres []string
		for n, otra := range c.funciones {
			if !otra.preludio {
				nombres = append(nombres, n)
			}
		}
		for n := range integradasNucleo {
			if !strings.HasPrefix(n, "nucleo_") {
				nombres = append(nombres, n)
			}
		}
		return nuevoError(`No hay ninguna funcion que se llame "`+nombre+`".`, ln,
			sugerir(nombre, nombres), ClaseNombre)
	}
	if err := c.nombrePermitido(nombre, ln, c.enPreludio); err != nil {
		return err
	}
	if len(args) != len(f.params) {
		return nuevoError(`"`+nombre+`" necesita `+itoa(len(f.params))+` datos y le das `+
			itoa(len(args))+`.`, ln, "", ClaseValor)
	}
	for _, a := range args {
		if err := c.expr(a); err != nil {
			return err
		}
		e.push(rdx)
		e.push(rax)
	}
	e.call(f.entrada)
	if len(args) > 0 {
		e.addImm(rsp, int32(16*len(args)))
	}
	return nil
}

func (c *compiladorNucleo) integrada(nombre string, args []Expresion, ln int) *ErrorFal {
	e := c.e
	switch nombre {
	case "cociente":
		return c.aritmetica("cociente", args[0], args[1], ln)

	case "largo":
		if err := c.expr(args[0]); err != nil {
			return err
		}
		e.call(c.rtLargo)
		c.exige(rax, etiquetaNumero, ln, `"largo" mide un texto o una region.`, "")

	case "lee_puerto":
		if err := c.expr(args[0]); err != nil {
			return err
		}
		c.puerto(rax, rdx, ln)
		e.entraPuerto()
		e.mov(rdx, rax)
		e.movImm(rax, etiquetaNumero)

	case "pon_en_puerto":
		if err := c.dosValores(args[0], args[1]); err != nil {
			return err
		}
		c.exige(rcx, etiquetaNumero, ln, "A un puerto solo se le mandan numeros.", "")
		e.cmpImm(rbx, 0)
		fuera := c.panico(ln, "A un puerto se le manda un byte: de 0 a 255.", "")
		e.jcc(siMenor, fuera)
		e.cmpImm(rbx, 255)
		e.jcc(siMayor, fuera)
		c.puerto(rax, rdx, ln)
		e.mov(rax, rbx)
		e.salePuerto()
		e.movImm(rax, etiquetaNada)
		e.movImm(rdx, 0)

	case "congela":
		// En las pruebas, que el preludio congele es que algo revento.
		codigo := int64(0)
		if c.enPreludio {
			codigo = 1
		}
		emitirCongela(e, c.prueba, codigo)

	case "nucleo_etiqueta":
		if err := c.expr(args[0]); err != nil {
			return err
		}
		e.mov(rdx, rax)
		e.andImm(rdx, 0xFF)
		e.movImm(rax, etiquetaNumero)

	case "region":
		return c.region(args, ln)

	case "atiende_interrupcion":
		return c.atiende(args, ln)

	case "activa_interrupciones":
		e.bytes(0xFB) // sti
	case "desactiva_interrupciones":
		e.bytes(0xFA) // cli
	case "espera_interrupcion":
		// sti y hlt juntos: sti no deja entrar ninguna interrupcion hasta
		// despues de la instruccion siguiente. Asi la que llegue justo aqui
		// despierta al hlt en vez de colarse antes y dejarlo dormido sin
		// nadie que lo despierte.
		e.bytes(0xFB, 0xF4)

	case "datos_de_arranque":
		e.movSimbolo(rdx, "datos_arranque", 0)
		e.lee(rdx, rdx, 0)
		e.movImm(rax, etiquetaNumero)
	case "final_del_nucleo":
		e.movSimbolo(rdx, "fin_nucleo", 0)
		e.movImm(rax, etiquetaNumero)
	}
	if strings.HasSuffix(nombre, "interrupciones") || nombre == "espera_interrupcion" {
		e.movImm(rax, etiquetaNada)
		e.movImm(rdx, 0)
	}
	return nil
}

// atiende pone una funcion del programa a atender una interrupcion. La
// funcion se dice por su nombre y se resuelve al compilar: en el nucleo no
// hay funciones guardadas en variables, pero el numero de la interrupcion
// si puede salir de una cuenta.
func (c *compiladorNucleo) atiende(args []Expresion, ln int) *ErrorFal {
	e := c.e
	nombre := ""
	if n, es := args[1].(*ExNombre); es {
		nombre = clave(n.Nombre)
	}
	f, hay := c.funciones[nombre]
	if !hay || (f.preludio && !c.enPreludio) {
		return nuevoError(`"atiende_interrupcion" necesita el nombre de una funcion tuya.`, ln,
			"Por ejemplo:   atiende_interrupcion con 33 y teclado", ClaseValor)
	}
	if len(f.params) > 0 {
		return nuevoError(`"`+nombre+`" no puede atender una interrupcion porque pide datos.`, ln,
			"Una interrupcion llega sin datos. Lo que necesite, que lo lea de una global o de un puerto.",
			ClaseValor)
	}
	c.atenciones[nombre] = true

	if err := c.expr(args[0]); err != nil {
		return err
	}
	c.exige(rax, etiquetaNumero, ln, "El numero de una interrupcion tiene que ser un numero.", "")
	// Las 32 primeras son las excepciones del procesador. Algunas meten un
	// codigo de error en la pila y otras no, y una funcion que no lo espera
	// volveria a una direccion equivocada.
	e.cmpImm(rdx, 32)
	e.jcc(siMenor, c.panico(ln, "Las interrupciones de la 0 a la 31 son las excepciones del procesador.",
		"Remapea el controlador de interrupciones para que las de los aparatos lleguen de la 32 en adelante."))
	e.cmpImm(rdx, 255)
	e.jcc(siMayor, c.panico(ln, "Solo hay 256 interrupciones, de la 0 a la 255.", ""))

	// La entrada de la tabla son dos palabras. Si llegase esa misma
	// interrupcion entre una y otra, saltaria a media direccion.
	e.bytes(0x9C, 0xFA) // pushfq; cli
	e.mov(rbx, rdx)
	e.shl(rbx, 4)
	e.movSimbolo(rsi, "idt", 0)
	e.add(rsi, rbx)
	e.movSimbolo(rax, "atencion:"+nombre, 0)
	emitirPuerta(e)
	e.bytes(0x9D) // popfq
	e.movImm(rax, etiquetaNada)
	e.movImm(rdx, 0)
	return nil
}

func (c *compiladorNucleo) puerto(etiq, dato, ln int) {
	c.exige(etiq, etiquetaNumero, ln, "El numero de puerto tiene que ser un numero.", "")
	fuera := c.panico(ln, "Los puertos van del 0 al 65535.", "")
	c.e.cmpImm(dato, 0)
	c.e.jcc(siMenor, fuera)
	c.e.cmpImm(dato, 65535)
	c.e.jcc(siMayor, fuera)
}

// region crea una ventana a la memoria: desde donde empieza, cuantos
// elementos tiene y cuantos bytes ocupa cada uno. Todo cabe en el propio
// valor, sin tabla aparte: la etiqueta lleva el ancho en el byte 1 y los
// elementos del bit 16 en adelante, y el dato es la direccion.
func (c *compiladorNucleo) region(args []Expresion, ln int) *ErrorFal {
	e := c.e
	for _, a := range args {
		if err := c.expr(a); err != nil {
			return err
		}
		e.push(rdx)
		e.push(rax)
	}
	if len(args) == 2 {
		e.movImm(rdx, 1)
		e.push(rdx)
		e.movImm(rax, etiquetaNumero)
		e.push(rax)
	}
	e.mov(rsi, rsp)
	for i := int32(0); i < 3; i++ {
		e.lee(rax, rsi, 16*i)
		c.exige(rax, etiquetaNumero, ln, `Los datos de "region" tienen que ser numeros.`, "")
	}
	e.lee(rdi, rsi, 8)
	e.lee(rcx, rsi, 24)
	e.lee(rbx, rsi, 40)
	e.addImm(rsp, 48)

	bueno := e.nueva()
	for _, ancho := range []int32{1, 2, 4, 8} {
		e.cmpImm(rdi, ancho)
		e.jcc(siIgual, bueno)
	}
	e.jmp(c.panico(ln, "Cada elemento de una region ocupa 1, 2, 4 u 8 bytes.", ""))
	e.aqui(bueno)
	e.cmpImm(rcx, 1)
	e.jcc(siMenor, c.panico(ln, "Una region tiene que tener al menos un elemento.", ""))

	// Lo que no mapean las tablas de paginas del arranque no existe todavia:
	// tocarlo seria un fallo de pagina sin nadie que lo atienda.
	fuera := c.panico(ln, "Esa region se sale del primer giga de memoria, que es lo unico que el nucleo ve por ahora.", "")
	e.cmpImm(rbx, 0)
	e.jcc(siMenor, fuera)
	e.cmpImm(rbx, memoriaVisible)
	e.jcc(siMayor, fuera)
	e.cmpImm(rcx, memoriaVisible)
	e.jcc(siMayor, fuera)
	e.mov(rax, rcx)
	e.imul(rax, rdi)
	e.add(rax, rbx)
	e.cmpImm(rax, memoriaVisible)
	e.jcc(siMayor, fuera)

	e.mov(rax, rcx)
	e.shl(rax, 16)
	e.mov(rdx, rdi)
	e.shl(rdx, 8)
	e.add(rax, rdx)
	e.addImm(rax, etiquetaRegion)
	e.mov(rdx, rbx)
	return nil
}

// donde arma la parte del aviso que cuenta en que linea fue.
func (c *compiladorNucleo) donde(linea int, pista string) string {
	donde := "linea " + itoa(linea)
	if c.enPreludio {
		donde += " del preludio"
	}
	if linea >= 1 && linea <= len(c.lineas) {
		donde += " |  " + strings.TrimSpace(c.lineas[linea-1])
	}
	if pista != "" {
		donde += "\n\nPista: " + pista
	}
	return donde
}

// Rutinas que comparte todo el programa. Devuelven en rax si salio bien, y
// quien las llama decide el mensaje, que necesita su numero de linea.
func (c *compiladorNucleo) emitirRutinas() {
	e := c.e

	e.aqui(c.pilaRota)
	// Con la pila agotada ni siquiera se puede llamar al aviso, asi que se
	// empieza una nueva: de esta ya no se vuelve.
	e.movSimbolo(rsp, "pila_tope", 0)
	c.llamarPanico("Demasiadas llamadas, una dentro de otra: se acabo la pila.",
		"Pista: una funcion que se llama a si misma necesita un caso en el que deja de hacerlo.")

	c.emitirTrampas()
	nombres := make([]string, 0, len(c.atenciones))
	for n := range c.atenciones {
		nombres = append(nombres, n)
	}
	sort.Strings(nombres)
	for _, n := range nombres {
		// El codigo compilado solo usa estos siete registros, asi que
		// guardando estos el programa interrumpido no nota nada.
		e.marca("atencion:" + n)
		guardados := []int{rax, rcx, rdx, rbx, rsi, rdi, rbp}
		for _, r := range guardados {
			e.push(r)
		}
		e.call(c.funciones[n].entrada)
		for i := len(guardados) - 1; i >= 0; i-- {
			e.pop(guardados[i])
		}
		e.bytes(0x48, 0xCF) // iretq
	}

	// rtLugar: coleccion en rcx/rbx y posicion en rax/rdx. Deja en rsi la
	// direccion y en rdi el ancho. rax: 0 bien, 1 no es una coleccion, 2 la
	// posicion no es un numero, 3 se sale.
	malIndice, fuera, texto, region, comun, positivo := e.nueva(), e.nueva(), e.nueva(), e.nueva(), e.nueva(), e.nueva()
	e.aqui(c.rtLugar)
	e.cmpImm(rax, etiquetaNumero)
	e.jcc(siDistinto, malIndice)
	e.cmpImm(rcx, etiquetaTexto)
	e.jcc(siIgual, texto)
	e.mov(rax, rcx)
	e.andImm(rax, 0xFF)
	e.cmpImm(rax, etiquetaRegion)
	e.jcc(siIgual, region)
	e.movImm(rax, 1)
	e.ret()
	e.aqui(texto)
	e.lee(rcx, rbx, 0)
	e.mov(rsi, rbx)
	e.addImm(rsi, 8)
	e.movImm(rdi, 1)
	e.jmp(comun)
	e.aqui(region)
	e.mov(rsi, rbx)
	e.mov(rdi, rcx)
	e.shr(rdi, 8)
	e.andImm(rdi, 0xFF)
	e.shr(rcx, 16)
	e.aqui(comun)
	e.cmpImm(rdx, 0)
	e.jcc(siMayorOIgual, positivo)
	e.add(rdx, rcx)
	e.addImm(rdx, 1)
	e.aqui(positivo)
	e.cmpImm(rdx, 1)
	e.jcc(siMenor, fuera)
	e.cmp(rdx, rcx)
	e.jcc(siMayor, fuera)
	e.subImm(rdx, 1)
	e.imul(rdx, rdi)
	e.add(rsi, rdx)
	e.movImm(rax, 0)
	e.ret()
	e.aqui(malIndice)
	e.movImm(rax, 2)
	e.ret()
	e.aqui(fuera)
	e.movImm(rax, 3)
	e.ret()

	// rtLee: lee lo que hay en rsi con el ancho de rdi y lo deja en rax/rdx.
	e.aqui(c.rtLee)
	listo := e.nueva()
	anchos := []int{1, 2, 4}
	saltos := []etiqueta{e.nueva(), e.nueva(), e.nueva()}
	for i, a := range anchos {
		e.cmpImm(rdi, int32(a))
		e.jcc(siIgual, saltos[i])
	}
	e.leeAncho(8)
	e.jmp(listo)
	for i, a := range anchos {
		e.aqui(saltos[i])
		e.leeAncho(a)
		e.jmp(listo)
	}
	e.aqui(listo)
	e.mov(rdx, rax)
	e.movImm(rax, etiquetaNumero)
	e.ret()

	// rtPon: guarda rdx en rsi con el ancho de rdi. rax: 0 bien, 4 no cabe.
	// Recortar en silencio un 300 a un byte es de los fallos que mas cuesta
	// encontrar, asi que no se recorta: se avisa.
	e.aqui(c.rtPon)
	ocho, noCabe, mira, hecho := e.nueva(), e.nueva(), e.nueva(), e.nueva()
	uno, dos := e.nueva(), e.nueva()
	e.cmpImm(rdi, 8)
	e.jcc(siIgual, ocho)
	e.cmpImm(rdx, 0)
	e.jcc(siMenor, noCabe)
	e.cmpImm(rdi, 1)
	e.jcc(siIgual, uno)
	e.cmpImm(rdi, 2)
	e.jcc(siIgual, dos)
	e.movImm(rcx, 1<<32)
	e.jmp(mira)
	e.aqui(uno)
	e.movImm(rcx, 1<<8)
	e.jmp(mira)
	e.aqui(dos)
	e.movImm(rcx, 1<<16)
	e.aqui(mira)
	e.cmp(rdx, rcx)
	e.jcc(siMayorOIgual, noCabe)
	e.mov(rax, rdx)
	guardaCon := []etiqueta{e.nueva(), e.nueva()}
	e.cmpImm(rdi, 1)
	e.jcc(siIgual, guardaCon[0])
	e.cmpImm(rdi, 2)
	e.jcc(siIgual, guardaCon[1])
	e.guardaAncho(4)
	e.jmp(hecho)
	e.aqui(guardaCon[0])
	e.guardaAncho(1)
	e.jmp(hecho)
	e.aqui(guardaCon[1])
	e.guardaAncho(2)
	e.jmp(hecho)
	e.aqui(ocho)
	e.mov(rax, rdx)
	e.guardaAncho(8)
	e.aqui(hecho)
	e.movImm(rax, 0)
	e.ret()
	e.aqui(noCabe)
	e.movImm(rax, 4)
	e.ret()

	// rtLargo: el largo del valor de rax/rdx en rdx, con rax a etiquetaNumero.
	// Si no tiene largo, rax queda a 0.
	noTiene, deRegion := e.nueva(), e.nueva()
	e.aqui(c.rtLargo)
	e.cmpImm(rax, etiquetaTexto)
	e.jcc(siDistinto, deRegion)
	e.lee(rdx, rdx, 0)
	e.movImm(rax, etiquetaNumero)
	e.ret()
	e.aqui(deRegion)
	e.mov(rcx, rax)
	e.andImm(rcx, 0xFF)
	e.cmpImm(rcx, etiquetaRegion)
	e.jcc(siDistinto, noTiene)
	e.mov(rdx, rax)
	e.shr(rdx, 16)
	e.movImm(rax, etiquetaNumero)
	e.ret()
	e.aqui(noTiene)
	e.movImm(rax, 0)
	e.ret()
}

// Las excepciones que meten en la pila un codigo de error. Las demas no, y
// para que todas lleguen igual al aviso, a esas se les mete un 0 a mano.
var conCodigoDeError = map[int]bool{8: true, 10: true, 11: true, 12: true, 13: true, 14: true,
	17: true, 21: true, 29: true, 30: true}

// emitirTrampas pone una entrada para cada una de las 256 interrupciones.
// Al arrancar todas apuntan aqui, y lo que no tenga una funcion que lo
// atienda acaba en un aviso que dice cual fue. Sin esto, una interrupcion
// sin dueño o un fallo del procesador reiniciaban la maquina sin decir
// nada. Cada entrada mide 16 bytes justos para que el arranque las pueda
// recorrer sumando.
func (c *compiladorNucleo) emitirTrampas() {
	e := c.e
	comun := e.nueva()
	e.marca("trampas")
	for v := 0; v < 256; v++ {
		inicio := len(e.b)
		if !conCodigoDeError[v] {
			e.bytes(0x6A, 0x00) // push 0
		}
		e.bytes(0x68) // push el numero de la interrupcion
		e.d32(uint32(v))
		e.jmp(comun)
		for len(e.b)-inicio < 16 {
			e.bytes(0x90)
		}
	}

	// En la pila: el numero, el codigo de error y la direccion donde iba.
	// De aqui no se vuelve, asi que se empieza una pila nueva: la de antes
	// puede ser justo lo que se rompio.
	e.aqui(comun)
	e.bytes(0xFA)
	e.mov(rsi, rsp)
	e.lee(rbx, rsi, 0)
	e.lee(rcx, rsi, 8)
	e.lee(rdi, rsi, 16)
	e.bytes(0x0F, 0x20, 0xD2) // mov rdx, cr2: la direccion de un fallo de pagina
	e.movSimbolo(rsp, "pila_tope", 0)
	for _, r := range []int{rbx, rcx, rdi, rdx} {
		e.push(r)
		e.movImm(rax, etiquetaNumero)
		e.push(rax)
	}
	e.call(c.funciones["nucleo_excepcion"].entrada)
}

// Los textos van detras del codigo con su largo delante, en bytes de UTF-8.
func (c *compiladorNucleo) emitirTextos() {
	e := c.e
	ordenados := make([]string, len(c.textos))
	for t, simbolo := range c.textos {
		n := 0
		for _, r := range strings.TrimPrefix(simbolo, "texto:") {
			n = n*10 + int(r-'0')
		}
		ordenados[n] = t
	}
	for _, t := range ordenados {
		e.alinea(8)
		e.marca(c.textos[t])
		e.d64(uint64(len(t)))
		e.bytes([]byte(t)...)
	}
}
