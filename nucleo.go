package main

// El compilador del nucleo: convierte un programa de Fal en una imagen que un
// ordenador arranca sin ningun sistema operativo debajo.
//
// Compila un Fal mas pequeño que el del interprete. Hay numeros enteros de 64
// bits, textos escritos en el programa, verdadero, falso, nada y regiones de
// memoria, y hay funciones. No hay listas ni decimales porque los dos
// necesitan pedir memoria mientras el programa corre, y en un nucleo eso es
// precisamente lo que todavia no existe: lo tendra que escribir el propio
// sistema operativo.
//
// Cada valor ocupa dos palabras, su etiqueta y su dato, y todas las
// operaciones miran la etiqueta antes de hacer nada. Asi un numero nunca se
// puede usar como direccion: la unica manera de tocar memoria es una region,
// y cada acceso a una region comprueba que no se sale. Con eso desaparece la
// clase de fallo que mas agujeros ha dado en los nucleos escritos en C.

import (
	_ "embed"
	"encoding/binary"
	"sort"
	"strings"
)

//go:embed nucleo/preludio.fal
var preludioNucleo string

//go:embed nucleo/programa.fal
var preludioPrograma string

const (
	etiquetaNada = iota
	etiquetaNumero
	etiquetaBooleano
	etiquetaTexto
	etiquetaRegion
)

type funcionNucleo struct {
	nombre   string
	params   []string
	cuerpo   []Instruccion
	linea    int
	entrada  etiqueta
	preludio bool
}

type bucleNucleo struct{ sigue, sale etiqueta }

type panicoPendiente struct {
	sitio          etiqueta
	mensaje, donde string
}

type compiladorNucleo struct {
	e         *ensamblador
	prueba    bool
	funciones map[string]*funcionNucleo
	orden     []*funcionNucleo
	globales  map[string]int
	textos    map[string]string
	pilaRota  etiqueta
	rtLugar   etiqueta
	rtLee     etiqueta
	rtPon     etiqueta
	rtLargo   etiqueta
	// Las funciones que atienden alguna interrupcion. Cada una necesita su
	// propia entrada, que guarda los registros y vuelve con iretq.
	atenciones map[string]bool
	// Lo mismo para los fallos de los programas: cada una, con sus tres datos.
	fallos map[string]bool

	programa bool
	leer     func(ruta string) (string, error)
	// Los programas que el sistema lleva dentro, ya compilados.
	incluidos map[string]int
	imagenes  [][]byte

	// Lo de la funcion que se esta compilando.
	lineas     []string
	enPreludio bool
	arriba     bool // el cuerpo del programa, fuera de toda funcion
	locales    map[string]int32
	huecos     int
	bucles     []bucleNucleo
	salida     etiqueta
	pendientes []panicoPendiente
	yaAvisado  map[string]etiqueta
	// Donde van el tamaño del marco y las palabras a limpiar, que no se
	// saben hasta haber compilado el cuerpo entero.
	marco [2]int
}

// opcionesNucleo: con prueba, al terminar o al reventar avisa a QEMU para que
// se cierre. Con programa, sale un programa suelto para que lo cargue un
// sistema, en vez de una imagen que arranca. leer es como se llega a los
// archivos de incluye_programa, que en la web no existen.
type opcionesNucleo struct {
	prueba   bool
	programa bool
	leer     func(ruta string) (string, error)
}

// compilarNucleo devuelve la imagen lista para arrancar.
func compilarNucleo(fuente string, prueba bool) ([]byte, *ErrorFal) {
	return compilarImagen(fuente, opcionesNucleo{prueba: prueba})
}

func compilarImagen(fuente string, op opcionesNucleo) ([]byte, *ErrorFal) {
	usuario, err := armarFuente(fuente)
	if err != nil {
		return nil, err
	}
	textoPreludio := preludioNucleo
	base := int64(cargaEn)
	if op.programa {
		textoPreludio, base = preludioPrograma, cargaPrograma
	}
	preludio, err := armarFuente(textoPreludio)
	if err != nil {
		err.Mensaje = "El preludio del nucleo no compila: " + err.Mensaje
		return nil, err
	}

	e := nuevoEnsamblador(base)
	c := &compiladorNucleo{
		e: e, prueba: op.prueba, programa: op.programa, leer: op.leer,
		funciones: map[string]*funcionNucleo{},
		globales:  map[string]int{},
		textos:    map[string]string{},
		incluidos: map[string]int{},
		fallos:    map[string]bool{},
	}
	c.pilaRota, c.rtLugar, c.rtLee, c.rtPon, c.rtLargo = e.nueva(), e.nueva(), e.nueva(), e.nueva(), e.nueva()
	c.atenciones = map[string]bool{}

	lineasPreludio := strings.Split(textoPreludio, "\n")
	lineasUsuario := strings.Split(fuente, "\n")
	if err := c.recogerFunciones(preludio, true); err != nil {
		return nil, err
	}
	if err := c.recogerFunciones(usuario, false); err != nil {
		return nil, err
	}
	if err := c.recogerGlobales(preludio, true); err != nil {
		return nil, err
	}
	if err := c.recogerGlobales(usuario, false); err != nil {
		return nil, err
	}

	principal := e.nueva()
	if op.programa {
		emitirCabeceraPrograma(e)
		emitirEntradaPrograma(e, principal)
	} else {
		emitirCabecera(e)
		emitirArranque(e, principal, op.prueba)
	}

	// El programa de arriba es una funcion mas, sin parametros, en la que
	// los nombres son globales. Primero va el preludio, que prepara lo que
	// el resto da por hecho.
	e.aqui(principal)
	c.empezarFuncion()
	c.arriba = true
	c.lineas, c.enPreludio = lineasPreludio, true
	if err := c.bloque(sinFunciones(preludio)); err != nil {
		return nil, err
	}
	c.lineas, c.enPreludio = lineasUsuario, false
	if err := c.bloque(sinFunciones(usuario)); err != nil {
		return nil, err
	}
	c.terminarFuncion()

	for _, f := range c.orden {
		c.lineas, c.enPreludio = lineasUsuario, f.preludio
		if f.preludio {
			c.lineas = lineasPreludio
		}
		if err := c.funcion(f); err != nil {
			return nil, err
		}
	}

	c.emitirRutinas()
	c.emitirTextos()
	if op.programa {
		colocarBSSPrograma(e, len(c.globales))
	} else {
		c.emitirImagenes()
		emitirGDT(e)
		colocarBSS(e, len(c.globales))
	}
	if falta := e.resolver(); falta != "" {
		return nil, nuevoError("Fallo del compilador del nucleo: no encuentro "+falta+".", 0,
			"Esto no es culpa de tu programa. Avisa en el foro con el programa que lo provoca.", ClasePrograma)
	}
	return e.b, nil
}

func armarFuente(fuente string) ([]Instruccion, *ErrorFal) {
	piezas, err := leer(quitarBOM(fuente))
	if err != nil {
		return nil, err
	}
	return armar(piezas)
}

func sinFunciones(programa []Instruccion) []Instruccion {
	var salida []Instruccion
	for _, ins := range programa {
		if _, es := ins.(*InsFuncion); !es {
			salida = append(salida, ins)
		}
	}
	return salida
}

// Lo que sabe hacer el nucleo sin que nadie lo escriba, con los datos que
// pide cada una. Las que empiezan por nucleo_ solo las usa el preludio.
var integradasNucleo = map[string][2]int{
	"region":          {2, 3},
	"largo":           {1, 1},
	"cociente":        {2, 2},
	"lee_puerto":      {1, 1},
	"lee_palabra":     {1, 1},
	"pon_en_puerto":   {2, 2},
	"congela":         {0, 0},
	"nucleo_etiqueta": {1, 1},

	"atiende_interrupcion":     {2, 2},
	"activa_interrupciones":    {0, 0},
	"desactiva_interrupciones": {0, 0},
	"espera_interrupcion":      {0, 0},

	"datos_de_arranque": {0, 0},
	"final_del_nucleo":  {0, 0},
	"inicio_del_codigo": {0, 0},
	"final_del_codigo":  {0, 0},
	"hay_no_ejecutar":   {0, 0},
	"usa_tablas":        {1, 1},

	"contexto":         {0, 0},
	"pon_contexto":     {1, 1},
	"pila_de_entrada":  {1, 1},
	"atiende_llamadas": {1, 1},
	"atiende_fallos":   {1, 1},
	"llama_al_sistema": {1, 3},
	"incluye_programa": {1, 1},
	"mapa_de_puertos":  {0, 0},
}

// Lo que no tiene sentido en un programa porque habla de cosas que solo
// existen en el nucleo. El resto de lo privilegiado (los puertos, cli, hlt,
// usa_tablas) si compila: lo para el procesador, que es quien manda. Pararlo
// aqui no protegeria nada, porque un programa no tiene por que venir de este
// compilador.
var soloEnElNucleo = map[string]bool{
	"atiende_interrupcion": true, "contexto": true, "pon_contexto": true, "pila_de_entrada": true,
	"atiende_llamadas": true, "atiende_fallos": true, "incluye_programa": true,
	"mapa_de_puertos": true,
}

func (c *compiladorNucleo) recogerFunciones(programa []Instruccion, preludio bool) *ErrorFal {
	for _, ins := range programa {
		f, es := ins.(*InsFuncion)
		if !es {
			continue
		}
		nombre := clave(f.Nombre)
		if err := c.nombrePermitido(nombre, f.Ln, preludio); err != nil {
			return err
		}
		if _, hay := integradasNucleo[nombre]; hay {
			return nuevoError(`Ya hay una funcion "`+nombre+`" en el nucleo.`, f.Ln,
				"Ponle otro nombre a la tuya.", ClaseNombre)
		}
		if _, hay := c.funciones[nombre]; hay {
			return nuevoError(`La funcion "`+nombre+`" esta escrita dos veces.`, f.Ln, "", ClaseNombre)
		}
		var params []string
		for _, p := range f.Parametros {
			params = append(params, clave(p))
		}
		nueva := &funcionNucleo{nombre, params, f.Cuerpo, f.Ln, c.e.nueva(), preludio}
		c.funciones[nombre] = nueva
		c.orden = append(c.orden, nueva)
	}
	return nil
}

func (c *compiladorNucleo) nombrePermitido(nombre string, linea int, preludio bool) *ErrorFal {
	if !preludio && strings.HasPrefix(nombre, "nucleo_") {
		return nuevoError(`Los nombres que empiezan por "nucleo_" son del compilador.`, linea,
			"Ponle otro nombre.", ClaseNombre)
	}
	return nil
}

// recogerGlobales apunta todo lo que se guarda arriba, fuera de las
// funciones, y lo que las funciones comparten. Se hace antes de compilar
// nada para que una funcion pueda leer una global que se guarda mas abajo
// en el archivo, igual que en el interprete.
func (c *compiladorNucleo) recogerGlobales(programa []Instruccion, preludio bool) *ErrorFal {
	arriba := map[string]bool{}
	compartidos := map[string]bool{}
	recogerNombres(sinFunciones(programa), arriba, compartidos)
	for _, ins := range programa {
		if f, es := ins.(*InsFuncion); es {
			recogerNombres(f.Cuerpo, map[string]bool{}, compartidos)
		}
	}
	for nombre := range compartidos {
		arriba[nombre] = true
	}
	for _, nombre := range ordenados(arriba) {
		if err := c.nombrePermitido(nombre, 0, preludio); err != nil {
			return err
		}
		if _, hay := c.funciones[nombre]; hay {
			return nuevoError(`"`+nombre+`" es a la vez una funcion y una variable.`, 0,
				"En el nucleo un nombre es una cosa o la otra. Cambia uno de los dos.", ClaseNombre)
		}
		if _, hay := c.globales[nombre]; !hay {
			c.globales[nombre] = len(c.globales)
		}
	}
	return nil
}

// ordenados existe porque repartir los huecos en el orden de un mapa hacia
// que la misma fuente diera cada vez una imagen distinta.
func ordenados(nombres map[string]bool) []string {
	salida := make([]string, 0, len(nombres))
	for n := range nombres {
		salida = append(salida, n)
	}
	sort.Strings(salida)
	return salida
}

// recogerNombres junta los nombres que un bloque guarda, sin entrar en las
// funciones que haya dentro.
func recogerNombres(bloque []Instruccion, guardados, compartidos map[string]bool) {
	for _, ins := range bloque {
		switch x := ins.(type) {
		case *InsGuarda:
			if n, es := x.Objetivo.(*ExNombre); es {
				guardados[clave(n.Nombre)] = true
			}
		case *InsComparte:
			for _, n := range x.Nombres {
				compartidos[clave(n)] = true
			}
		case *InsSi:
			recogerNombres(x.Cuerpo, guardados, compartidos)
			recogerNombres(x.Sino, guardados, compartidos)
		case *InsMientras:
			recogerNombres(x.Cuerpo, guardados, compartidos)
		case *InsRepite:
			recogerNombres(x.Cuerpo, guardados, compartidos)
		case *InsCuenta:
			guardados[clave(x.Nombre)] = true
			recogerNombres(x.Cuerpo, guardados, compartidos)
		case *InsRecorre:
			guardados[clave(x.Nombre)] = true
			recogerNombres(x.Cuerpo, guardados, compartidos)
		}
	}
}

// Funciones

// Marco de una funcion, desde rbp:
//
//	rbp+16 ...  los parametros, el ultimo el mas cerca
//	rbp+8       la vuelta
//	rbp         el rbp de quien llamo
//	rbp-16 ...  las variables locales y los huecos de los bucles
//
// Cada hueco son 16 bytes: la etiqueta y, 8 bytes mas arriba, el dato.
func (c *compiladorNucleo) empezarFuncion() {
	c.locales = map[string]int32{}
	c.huecos = 0
	c.bucles = nil
	c.pendientes = nil
	c.yaAvisado = map[string]etiqueta{}
	c.salida = c.e.nueva()
	c.arriba = false
	e := c.e

	e.push(rbp)
	e.mov(rbp, rsp)
	// Sin esto, una funcion que se llama a si misma sin parar bajaria la
	// pila hasta pisar las tablas de paginas, y el procesador se reiniciaria
	// sin decir nada.
	// El suelo esta en memoria porque cambia: quien atiende una interrupcion
	// corre en otra pila.
	e.leeAbs(rcx, "suelo")
	e.cmp(rsp, rcx)
	e.jcc(siDebajo, c.pilaRota)

	e.bytes(0x48, 0x81, 0xEC) // sub rsp, tamaño del marco
	c.marco[0] = len(e.b)
	e.d32(0)
	// Las variables empiezan valiendo nada (etiqueta 0), no lo que hubiera
	// en la pila.
	e.mov(rdi, rsp)
	e.bytes(0x48, 0xC7, 0xC1) // mov rcx, palabras del marco
	c.marco[1] = len(e.b)
	e.d32(0)
	e.movImm(rax, 0)
	e.bytes(0xF3, 0x48, 0xAB) // rep stosq
}

func (c *compiladorNucleo) terminarFuncion() {
	e := c.e
	e.movImm(rax, etiquetaNada)
	e.movImm(rdx, 0)
	e.aqui(c.salida)
	e.mov(rsp, rbp)
	e.pop(rbp)
	e.ret()

	binary.LittleEndian.PutUint32(e.b[c.marco[0]:], uint32(16*c.huecos))
	binary.LittleEndian.PutUint32(e.b[c.marco[1]:], uint32(2*c.huecos))

	for _, p := range c.pendientes {
		e.aqui(p.sitio)
		c.llamarPanico(p.mensaje, p.donde)
	}
}

func (c *compiladorNucleo) funcion(f *funcionNucleo) *ErrorFal {
	c.e.aqui(f.entrada)
	c.empezarFuncion()
	for i, p := range f.params {
		c.locales[p] = int32(16 + 16*(len(f.params)-1-i))
	}
	guardados := map[string]bool{}
	compartidos := map[string]bool{}
	recogerNombres(f.cuerpo, guardados, compartidos)
	for _, nombre := range ordenados(guardados) {
		if _, esParam := c.locales[nombre]; esParam || compartidos[nombre] {
			continue
		}
		if _, hay := c.funciones[nombre]; hay {
			return nuevoError(`"`+nombre+`" ya es el nombre de una funcion.`, f.linea,
				"Usa otro nombre para la variable.", ClaseNombre)
		}
		c.locales[nombre] = c.hueco()
	}
	if err := c.bloque(f.cuerpo); err != nil {
		return err
	}
	c.terminarFuncion()
	return nil
}

func (c *compiladorNucleo) hueco() int32 {
	c.huecos++
	return int32(-16 * c.huecos)
}

// Los avisos

// panico devuelve a donde saltar para reventar con este mensaje. El texto
// se arma aqui, al compilar, con la linea del programa dentro: en la
// maquina no hay archivo fuente que consultar.
func (c *compiladorNucleo) panico(linea int, mensaje, pista string) etiqueta {
	donde := c.donde(linea, pista)
	llave := mensaje + "\x00" + donde
	if l, hay := c.yaAvisado[llave]; hay {
		return l
	}
	l := c.e.nueva()
	c.yaAvisado[llave] = l
	c.pendientes = append(c.pendientes, panicoPendiente{l, mensaje, donde})
	return l
}

func (c *compiladorNucleo) llamarPanico(mensaje, donde string) {
	e := c.e
	for _, t := range []string{mensaje, donde} {
		e.movSimbolo(rdx, c.texto(t), 0)
		e.push(rdx)
		e.movImm(rax, etiquetaTexto)
		e.push(rax)
	}
	e.call(c.funciones["nucleo_panico"].entrada)
}

// exige revienta si la etiqueta que hay en r no es la esperada.
func (c *compiladorNucleo) exige(r, etiq, linea int, mensaje, pista string) {
	c.e.cmpImm(r, int32(etiq))
	c.e.jcc(siDistinto, c.panico(linea, mensaje, pista))
}

func (c *compiladorNucleo) texto(t string) string {
	if s, hay := c.textos[t]; hay {
		return s
	}
	s := "texto:" + itoa(len(c.textos))
	c.textos[t] = s
	return s
}

func noHayTodavia(que string, linea int) *ErrorFal {
	return nuevoError("En el nucleo todavia no hay "+que+".", linea,
		"El nucleo compila un Fal mas pequeño: numeros enteros, textos, verdadero y falso, "+
			"regiones de memoria y funciones. Lo que falta lo ira escribiendo el propio sistema.",
		ClaseSintaxis)
}
