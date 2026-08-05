package main

// La biblioteca: lo que toca el mundo de fuera. Archivos, JSON, patrones,
// tablas, internet, fechas y reloj.

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Archivos

func registrarArchivos() {
	integrada("lee", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		r, err := pideTexto(a[0], ln, "lee")
		if err != nil {
			return nil, err
		}
		datos, e := os.ReadFile(in.rutaDe(r))
		if e != nil {
			return nil, nuevoError(`No pude leer el archivo "`+r+`".`, ln, e.Error(), ClaseArchivo)
		}
		return quitarBOM(string(datos)), nil
	})

	integrada("lineas", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		v, err := integradas["lee"].fn(in, a, ln)
		if err != nil {
			return nil, err
		}
		texto := strings.ReplaceAll(v.(string), "\r\n", "\n")
		texto = strings.TrimSuffix(texto, "\n")
		if texto == "" {
			return lista(nil), nil
		}
		return listaDeTextos(strings.Split(texto, "\n")), nil
	})

	escribirArchivo := func(nombre string, banderas int) {
		integrada(nombre, 2, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
			r, err := pideTexto(a[0], ln, nombre)
			if err != nil {
				return nil, err
			}
			f, e := os.OpenFile(in.rutaDe(r), banderas, 0644)
			if e != nil {
				return nil, nuevoError(`No pude escribir en el archivo "`+r+`".`, ln,
					e.Error(), ClaseArchivo)
			}
			defer f.Close()
			if _, e := f.WriteString(textoDe(a[1])); e != nil {
				return nil, nuevoError(`No pude escribir en el archivo "`+r+`".`, ln,
					e.Error(), ClaseArchivo)
			}
			return nil, nil
		})
	}
	escribirArchivo("graba", os.O_CREATE|os.O_WRONLY|os.O_TRUNC)
	escribirArchivo("anexa", os.O_CREATE|os.O_WRONLY|os.O_APPEND)

	integrada("existe", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		r, err := pideTexto(a[0], ln, "existe")
		if err != nil {
			return nil, err
		}
		_, e := os.Stat(in.rutaDe(r))
		return e == nil, nil
	})

	integrada("borra", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		r, err := pideTexto(a[0], ln, "borra")
		if err != nil {
			return nil, err
		}
		if e := os.Remove(in.rutaDe(r)); e != nil {
			return nil, nuevoError(`No pude borrar "`+r+`".`, ln, e.Error(), ClaseArchivo)
		}
		return nil, nil
	})

	integrada("archivos", 0, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		carpeta := in.carpeta
		if len(a) > 0 {
			r, err := pideTexto(a[0], ln, "archivos")
			if err != nil {
				return nil, err
			}
			carpeta = in.rutaDe(r)
		}
		entradas, e := os.ReadDir(carpeta)
		if e != nil {
			return nil, nuevoError("No pude mirar dentro de esa carpeta.", ln, e.Error(), ClaseArchivo)
		}
		var nombres []string
		for _, x := range entradas {
			nombres = append(nombres, x.Name())
		}
		sort.Strings(nombres)
		return listaDeTextos(nombres), nil
	})
}

// JSON, patrones, tablas e internet

func registrarDatos() {
	integrada("a_json", 1, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		var b strings.Builder
		bonito := len(a) > 1 && esVerdad(a[1])
		if err := escribirJSON(&b, a[0], bonito, 0, ln); err != nil {
			return nil, err
		}
		return b.String(), nil
	})

	integrada("de_json", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		s, err := pideTexto(a[0], ln, "de_json")
		if err != nil {
			return nil, err
		}
		var crudo interface{}
		d := json.NewDecoder(strings.NewReader(s))
		d.UseNumber()
		if e := d.Decode(&crudo); e != nil {
			return nil, errValor("Ese texto no es JSON valido.", ln, e.Error())
		}
		return deJSON(crudo), nil
	})

	compilar := func(v Valor, ln int, quien string) (*regexp.Regexp, *ErrorFal) {
		s, err := pideTexto(v, ln, quien)
		if err != nil {
			return nil, err
		}
		re, e := regexp.Compile(s)
		if e != nil {
			return nil, errValor("Ese patron esta mal escrito.", ln, e.Error())
		}
		return re, nil
	}

	integrada("coincide", 2, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		s, err := pideTexto(a[0], ln, "coincide")
		if err != nil {
			return nil, err
		}
		re, err := compilar(a[1], ln, "coincide")
		if err != nil {
			return nil, err
		}
		return re.MatchString(s), nil
	})

	integrada("busca", 2, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		s, err := pideTexto(a[0], ln, "busca")
		if err != nil {
			return nil, err
		}
		re, err := compilar(a[1], ln, "busca")
		if err != nil {
			return nil, err
		}
		return listaDeTextos(re.FindAllString(s, -1)), nil
	})

	integrada("cambia", 3, 3, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		s, err := pideTexto(a[0], ln, "cambia")
		if err != nil {
			return nil, err
		}
		re, err := compilar(a[1], ln, "cambia")
		if err != nil {
			return nil, err
		}
		nuevo, err := pideTexto(a[2], ln, "cambia")
		if err != nil {
			return nil, err
		}
		return re.ReplaceAllLiteralString(s, nuevo), nil
	})

	integrada("tabla", 1, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		s, err := pideTexto(a[0], ln, "tabla")
		if err != nil {
			return nil, err
		}
		lector := csv.NewReader(strings.NewReader(s))
		lector.FieldsPerRecord = -1
		if len(a) > 1 {
			sep, err := pideTexto(a[1], ln, "tabla")
			if err != nil {
				return nil, err
			}
			if sep != "" {
				lector.Comma = []rune(sep)[0]
			}
		}
		filas, e := lector.ReadAll()
		if e != nil {
			return nil, errValor("Ese texto no es una tabla valida.", ln, e.Error())
		}
		var salida []Valor
		for _, f := range filas {
			if len(f) > 0 {
				salida = append(salida, listaDeTextos(f))
			}
		}
		return lista(salida), nil
	})

	integrada("a_csv", 1, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		l, err := pideLista(a[0], ln, "a_csv")
		if err != nil {
			return nil, err
		}
		var b strings.Builder
		escritor := csv.NewWriter(&b)
		if len(a) > 1 {
			sep, err := pideTexto(a[1], ln, "a_csv")
			if err != nil {
				return nil, err
			}
			if sep != "" {
				escritor.Comma = []rune(sep)[0]
			}
		}
		for _, fila := range l.Datos {
			sub, err := pideLista(fila, ln, "a_csv")
			if err != nil {
				return nil, err
			}
			campos := make([]string, len(sub.Datos))
			for i, v := range sub.Datos {
				campos[i] = textoDe(v)
			}
			escritor.Write(campos)
		}
		escritor.Flush()
		return strings.TrimRight(strings.ReplaceAll(b.String(), "\r\n", "\n"), "\n"), nil
	})

	integrada("pide", 1, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		url, err := pideTexto(a[0], ln, "pide")
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "https://" + url
		}
		segundos := 20
		if len(a) > 1 {
			if segundos, err = pideEntero(a[1], ln, "pide"); err != nil {
				return nil, err
			}
		}
		cliente := &http.Client{Timeout: time.Duration(segundos) * time.Second}
		peticion, _ := http.NewRequest("GET", url, nil)
		peticion.Header.Set("User-Agent", "Fal")
		respuesta, e := cliente.Do(peticion)
		if e != nil {
			return nil, nuevoError("No pude conectar con esa direccion.", ln,
				url+"  ("+e.Error()+")", ClaseRed)
		}
		defer respuesta.Body.Close()
		if respuesta.StatusCode >= 400 {
			return nil, nuevoError("La pagina contesto con el error "+
				itoa(respuesta.StatusCode)+".", ln, url, ClaseRed)
		}
		cuerpo, e := io.ReadAll(io.LimitReader(respuesta.Body, 50<<20))
		if e != nil {
			return nil, nuevoError("Se corto la descarga.", ln, e.Error(), ClaseRed)
		}
		return string(cuerpo), nil
	})
}

// escribirJSON pasa un valor de Fal a texto JSON.
//
// Se escribe a mano en vez de usar la libreria para dos cosas que
// importan: que las claves salgan siempre en el mismo orden, y que los
// numeros exactos no pierdan precision al convertirse.
func escribirJSON(b *strings.Builder, v Valor, bonito bool, nivel int, ln int) *ErrorFal {
	sangria := func(n int) {
		if bonito {
			b.WriteString("\n")
			b.WriteString(strings.Repeat("  ", n))
		}
	}
	dosPuntos := ": "
	coma := ", "
	if bonito {
		coma = ","
	}

	switch x := v.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		if x {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case string:
		datos, _ := json.Marshal(x)
		b.Write(datos)
	case Num:
		b.WriteString(x.Texto())

	case *Lista:
		if len(x.Datos) == 0 {
			b.WriteString("[]")
			return nil
		}
		b.WriteString("[")
		for i, e := range x.Datos {
			if i > 0 {
				b.WriteString(coma)
			}
			sangria(nivel + 1)
			if err := escribirJSON(b, e, bonito, nivel+1, ln); err != nil {
				return err
			}
		}
		sangria(nivel)
		b.WriteString("]")

	case *Conjunto:
		return escribirJSON(b, lista(x.Elementos()), bonito, nivel, ln)

	case *Dicc:
		return escribirObjetoJSON(b, x.Claves, func(k string) Valor { return x.Datos[k] },
			bonito, nivel, ln, coma, dosPuntos, sangria)

	case *Objeto:
		return escribirObjetoJSON(b, x.Orden, func(k string) Valor { return x.Campos[k] },
			bonito, nivel, ln, coma, dosPuntos, sangria)

	default:
		return errTipo("Una funcion o un tipo no se pueden pasar a JSON.", ln, "")
	}
	return nil
}

func escribirObjetoJSON(b *strings.Builder, claves []string, dame func(string) Valor,
	bonito bool, nivel, ln int, coma, dosPuntos string, sangria func(int)) *ErrorFal {
	if len(claves) == 0 {
		b.WriteString("{}")
		return nil
	}
	b.WriteString("{")
	for i, k := range claves {
		if i > 0 {
			b.WriteString(coma)
		}
		sangria(nivel + 1)
		datos, _ := json.Marshal(k)
		b.Write(datos)
		b.WriteString(dosPuntos)
		if err := escribirJSON(b, dame(k), bonito, nivel+1, ln); err != nil {
			return err
		}
	}
	sangria(nivel)
	b.WriteString("}")
	return nil
}

func deJSON(v interface{}) Valor {
	switch x := v.(type) {
	case nil:
		return nil
	case bool:
		return x
	case string:
		return x
	case json.Number:
		if n, ok := NumDesdeTexto(x.String()); ok {
			return n
		}
		f, _ := x.Float64()
		return Flotante(f)
	case []interface{}:
		salida := make([]Valor, 0, len(x))
		for _, e := range x {
			salida = append(salida, deJSON(e))
		}
		return lista(salida)
	case map[string]interface{}:
		d := nuevoDicc()
		var claves []string
		for k := range x {
			claves = append(claves, k)
		}
		sort.Strings(claves)
		for _, k := range claves {
			d.Pon(k, deJSON(x[k]))
		}
		return d
	}
	return nil
}

// Fechas

var diasSemana = []string{"domingo", "lunes", "martes", "miercoles", "jueves", "viernes", "sabado"}
var mesesAnio = []string{"enero", "febrero", "marzo", "abril", "mayo", "junio",
	"julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"}

func pideFecha(v Valor, ln int, quien string) (time.Time, *ErrorFal) {
	s, err := pideTexto(v, ln, quien)
	if err != nil {
		return time.Time{}, err
	}
	s = strings.TrimSpace(s)
	if len(s) > 10 {
		s = s[:10]
	}
	t, e := time.Parse("2006-01-02", s)
	if e != nil {
		return time.Time{}, errValor(`"`+textoDe(v)+`" no es una fecha.`, ln,
			`Las fechas se escriben asi: "2026-07-31".`)
	}
	return t, nil
}

func registrarFechas() {
	integrada("ahora", 0, 0, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		return time.Now().Format("2006-01-02 15:04:05"), nil
	})
	integrada("hoy", 0, 0, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		return time.Now().Format("2006-01-02"), nil
	})

	deFecha := func(nombre string, f func(time.Time) Valor) {
		integrada(nombre, 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
			t, err := pideFecha(a[0], ln, nombre)
			if err != nil {
				return nil, err
			}
			return f(t), nil
		})
	}
	deFecha("anio", func(t time.Time) Valor { return Entero(int64(t.Year())) })
	deFecha("mes", func(t time.Time) Valor { return Entero(int64(t.Month())) })
	deFecha("dia", func(t time.Time) Valor { return Entero(int64(t.Day())) })
	deFecha("diasemana", func(t time.Time) Valor { return diasSemana[int(t.Weekday())] })
	deFecha("nombremes", func(t time.Time) Valor { return mesesAnio[int(t.Month())-1] })

	integrada("dias", 2, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		desde, err := pideFecha(a[0], ln, "dias")
		if err != nil {
			return nil, err
		}
		hasta, err := pideFecha(a[1], ln, "dias")
		if err != nil {
			return nil, err
		}
		return Entero(int64(hasta.Sub(desde).Hours() / 24)), nil
	})

	integrada("avanza", 2, 2, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		t, err := pideFecha(a[0], ln, "avanza")
		if err != nil {
			return nil, err
		}
		n, err := pideEntero(a[1], ln, "avanza")
		if err != nil {
			return nil, err
		}
		return t.AddDate(0, 0, n).Format("2006-01-02"), nil
	})

	integrada("fecha", 3, 3, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		anio, err := pideEntero(a[0], ln, "fecha")
		if err != nil {
			return nil, err
		}
		mes, err := pideEntero(a[1], ln, "fecha")
		if err != nil {
			return nil, err
		}
		dia, err := pideEntero(a[2], ln, "fecha")
		if err != nil {
			return nil, err
		}
		t := time.Date(anio, time.Month(mes), dia, 0, 0, 0, 0, time.UTC)
		if t.Year() != anio || int(t.Month()) != mes || t.Day() != dia {
			return nil, errValor("Esa fecha no existe.", ln, "")
		}
		return t.Format("2006-01-02"), nil
	})
}

// Tiempo y sistema

func registrarSistema() {
	// limpia borra la pantalla. Si la consola no entiende las secuencias
	// de control, se apana empujando el texto viejo hacia arriba.
	integrada("limpia", 0, 0, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		if soportaSecuencias {
			in.salida.WriteString("\x1b[2J\x1b[H")
		} else {
			in.salida.WriteString(strings.Repeat("\n", 50))
		}
		return nil, nil
	})

	integrada("reloj", 0, 0, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		r := new(big.Rat).SetFloat64(float64(time.Now().UnixNano()) / 1e9)
		return Racional(r), nil
	})

	integrada("espera", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		n, err := pideNum(a[0], ln, "espera")
		if err != nil {
			return nil, err
		}
		in.salida.Flush()
		time.Sleep(time.Duration(n.Float() * float64(time.Second)))
		return nil, nil
	})

	integrada("argumentos", 0, 0, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		return listaDeTextos(in.argumentos), nil
	})

	integrada("termina", 0, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
		codigo := 0
		if len(a) > 0 {
			c, err := pideEntero(a[0], ln, "termina")
			if err != nil {
				return nil, err
			}
			codigo = c
		}
		in.salida.Flush()
		soltarTeclado() // os.Exit no deja correr ningun defer
		os.Exit(codigo)
		return nil, nil
	})
}
