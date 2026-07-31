package main

// Memoria, entorno y estado del interprete.

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const limiteLlamadas = 3000

// Memoria

type Memoria struct {
	datos       map[string]Valor
	padre       *Memoria
	esFuncion   bool
	compartidas map[string]bool
}

func nuevaMemoria(padre *Memoria, esFuncion bool) *Memoria {
	return &Memoria{datos: map[string]Valor{}, padre: padre, esFuncion: esFuncion}
}

func (m *Memoria) existe(n string) bool {
	for l := m; l != nil; l = l.padre {
		if _, hay := l.datos[n]; hay {
			return true
		}
	}
	return false
}

func (m *Memoria) leer(n string) (Valor, bool) {
	for l := m; l != nil; l = l.padre {
		if v, hay := l.datos[n]; hay {
			return v, true
		}
	}
	return nil, false
}

func (m *Memoria) raiz() *Memoria {
	l := m
	for l.padre != nil {
		l = l.padre
	}
	return l
}

func (m *Memoria) declararCompartida(n string) {
	if m.compartidas == nil {
		m.compartidas = map[string]bool{}
	}
	m.compartidas[n] = true
}

// guardar deja lo que escribes DENTRO de la funcion donde estas, salvo que
// hayas dicho "comparte". Sin esto, una funcion podria pisar las variables
// de fuera sin avisar, que es el peor tipo de error: el que no se ve.
func (m *Memoria) guardar(n string, v Valor) {
	if m.compartidas[n] {
		for l := m.padre; l != nil; l = l.padre {
			if _, hay := l.datos[n]; hay {
				l.datos[n] = v
				return
			}
		}
		m.raiz().datos[n] = v
		return
	}
	for l := m; l != nil; l = l.padre {
		if _, hay := l.datos[n]; hay {
			l.datos[n] = v
			return
		}
		if l.esFuncion {
			break
		}
	}
	m.datos[n] = v
}

func (m *Memoria) nombres() []string {
	vistos := map[string]bool{}
	var salida []string
	for l := m; l != nil; l = l.padre {
		for n := range l.datos {
			if !vistos[n] {
				vistos[n] = true
				salida = append(salida, n)
			}
		}
	}
	return salida
}

// Señales de control (no son errores)

type senal int

const (
	sigNada senal = iota
	sigDetente
	sigContinua
	sigDevuelve
)

type resultado struct {
	sig   senal
	valor Valor
}

var seguir = resultado{}

// El interprete

type Interprete struct {
	global     *Memoria
	funciones  map[string]*Funcion
	tipos      map[string]*Tipo
	carpeta    string
	argumentos []string
	cargados   map[string]Valor
	pila       []Marco
	tipoError  *Tipo
	salida     *bufio.Writer
	entrada    *bufio.Reader
}

func nuevoInterprete(carpeta string, args []string) *Interprete {
	in := &Interprete{
		global:     nuevaMemoria(nil, false),
		funciones:  map[string]*Funcion{},
		tipos:      map[string]*Tipo{},
		carpeta:    carpeta,
		argumentos: args,
		cargados:   map[string]Valor{},
		salida:     bufio.NewWriter(os.Stdout),
		entrada:    bufio.NewReader(os.Stdin),
	}
	in.tipoError = &Tipo{Nombre: "Error", Propios: []string{"mensaje", "clase", "linea", "pista"},
		Metodos: map[string]*Funcion{}, EsError: true}
	in.tipos["error"] = in.tipoError
	// pi y e con muchos decimales, guardados como fraccion exacta.
	pi, _ := NumDesdeTexto("3.14159265358979323846")
	e, _ := NumDesdeTexto("2.71828182845904523536")
	in.global.datos["pi"] = pi
	in.global.datos["e"] = e
	interpreteActual = in
	return in
}

func (in *Interprete) rutaDe(r string) string {
	if filepath.IsAbs(r) {
		return r
	}
	return filepath.Join(in.carpeta, r)
}

func (in *Interprete) escribir(s string, salto bool) {
	in.salida.WriteString(s)
	if salto {
		in.salida.WriteString("\n")
	}
}

func (in *Interprete) preguntar(mensaje string) (string, bool) {
	in.salida.WriteString(mensaje)
	in.salida.Flush()
	linea, err := in.entrada.ReadString('\n')
	if err != nil && linea == "" {
		return "", false
	}
	return strings.TrimRight(linea, "\r\n"), true
}

func (in *Interprete) correr(instrucciones []Instruccion) *ErrorFal {
	if err := in.registrar(instrucciones); err != nil {
		return err
	}
	for _, ins := range instrucciones {
		switch ins.(type) {
		case *InsFuncion, *InsTipo:
			continue
		}
		if _, err := in.ejecutar(ins, in.global); err != nil {
			return err
		}
	}
	return nil
}

// registrar apunta funciones y tipos antes de empezar, para poder usarlos
// desde arriba del archivo aunque esten definidos mas abajo.
func (in *Interprete) registrar(instrucciones []Instruccion) *ErrorFal {
	for _, ins := range instrucciones {
		switch x := ins.(type) {
		case *InsFuncion:
			in.funciones[x.Nombre] = &Funcion{Nombre: x.Nombre, Parametros: x.Parametros,
				Cuerpo: x.Cuerpo, Entorno: in.global}
		case *InsTipo:
			metodos := map[string]*Funcion{}
			for _, m := range x.Metodos {
				metodos[m.Nombre] = &Funcion{Nombre: m.Nombre, Parametros: m.Parametros,
					Cuerpo: m.Cuerpo, Entorno: in.global}
			}
			var madre *Tipo
			if x.Madre != "" {
				m, hay := in.tipos[x.Madre]
				if !hay {
					return nuevoError(`No conozco ningun tipo llamado "`+x.Madre+`" para heredar.`,
						x.Ln, sugerirDeMapaTipos(x.Madre, in.tipos), ClaseNombre)
				}
				madre = m
			}
			in.tipos[x.Nombre] = &Tipo{Nombre: x.Escrito, Propios: x.Campos,
				Metodos: metodos, Madre: madre}
		}
	}
	return nil
}

func (in *Interprete) ejecutarBloque(cuerpo []Instruccion, mem *Memoria) (resultado, *ErrorFal) {
	for _, ins := range cuerpo {
		r, err := in.ejecutar(ins, mem)
		if err != nil {
			return seguir, err
		}
		if r.sig != sigNada {
			return r, nil
		}
	}
	return seguir, nil
}

// bucle ejecuta el cuerpo y dice si hay que seguir dando vueltas.
func (in *Interprete) bucle(cuerpo []Instruccion, mem *Memoria) (bool, resultado, *ErrorFal) {
	r, err := in.ejecutarBloque(cuerpo, mem)
	if err != nil {
		return false, seguir, err
	}
	switch r.sig {
	case sigDetente:
		return false, seguir, nil
	case sigDevuelve:
		return false, r, nil
	}
	return true, seguir, nil
}

func (in *Interprete) objetoError(e *ErrorFal) *Objeto {
	o := &Objeto{Tipo: in.tipoError, Campos: map[string]Valor{}}
	o.Pon("mensaje", e.Mensaje)
	o.Pon("clase", string(e.Clase))
	o.Pon("linea", Entero(int64(e.Linea)))
	o.Pon("pista", e.Pista)
	return o
}

func sugerirDeMapaTipos(nombre string, tipos map[string]*Tipo) string {
	var claves []string
	for k := range tipos {
		claves = append(claves, k)
	}
	sort.Strings(claves)
	if s := sugerir(nombre, claves); s != "" {
		return s
	}
	return "El tipo del que se hereda tiene que estar definido antes."
}
