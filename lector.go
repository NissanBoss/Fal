package main

// Parte el texto en piezas:  "edad es 25"  ->  edad | es | 25

import (
	"strings"
	"unicode"
)

type TipoPieza int

const (
	PNumero TipoPieza = iota
	PTexto
	PPalabra
	PAbre
	PCierra
	PFinLinea
	PFinArchivo
)

type Pieza struct {
	Tipo  TipoPieza
	Texto string // el texto tal cual se escribio
	Clave string // normalizado: sin tildes y en minusculas
	Num   Num
	Linea int
}

// Las tildes se quitan solo para comparar palabras del lenguaje, de forma
// que "función" y "funcion" sean la misma. Los nombres que escribes tu
// conservan sus tildes.
var sinTilde = strings.NewReplacer(
	"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n",
	"Á", "a", "É", "e", "Í", "i", "Ó", "o", "Ú", "u", "Ü", "u", "Ñ", "n",
)

func clave(s string) string { return sinTilde.Replace(strings.ToLower(s)) }

// Si escribes un simbolo de otro lenguaje, te decimos como se dice en Fal.
var traduccionSimbolos = map[rune]string{
	'=': "Para guardar un valor se escribe:   nombre es valor",
	'+': `Para sumar se usa la palabra "mas":   3 mas 4`,
	'-': `Para restar se usa la palabra "menos":   10 menos 4`,
	'*': `Para multiplicar se usa la palabra "por":   3 por 4`,
	'/': `Para dividir se usa la palabra "entre":   10 entre 2`,
	'%': `Para el resto de una division se usa "resto":   10 resto 3`,
	'>': `Para comparar se escribe "es mayor que":   si edad es mayor que 18`,
	'<': `Para comparar se escribe "es menor que":   si edad es menor que 18`,
	'!': `Para negar se usa la palabra "no":   si no encontrado`,
	'&': `Para unir condiciones se usa la palabra "y":   si a y b`,
	'|': `Para elegir entre condiciones se usa la palabra "o":   si a o b`,
	';': "En Fal no hacen falta los puntos y coma. Puedes borrarlo.",
	'{': `Los bloques no llevan llaves: terminan con la palabra "fin".`,
	'}': `Los bloques no llevan llaves: terminan con la palabra "fin".`,
	'[': `Para crear una lista se escribe:   frutas es lista con "pera" y "uva"`,
	']': `Para crear una lista se escribe:   frutas es lista con "pera" y "uva"`,
	':': "En Fal no hacen falta los dos puntos. Puedes borrarlo.",
	'.': `Para llegar a un campo se escribe al reves:   nombre de persona`,
}

var cierreComilla = map[rune]rune{'"': '"', '\'': '\'', '“': '”', '‘': '’'}

func esInicioNombre(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func esParteNombre(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func leer(fuente string) ([]Pieza, *ErrorFal) {
	runas := []rune(fuente)
	piezas := []Pieza{}
	i, n, linea := 0, len(runas), 1

	for i < n {
		c := runas[i]

		switch {
		case c == ' ' || c == '\t' || c == '\r' || c == ' ':
			i++
			continue

		case c == '#':
			for i < n && runas[i] != '\n' {
				i++
			}
			continue

		case c == '\n':
			piezas = append(piezas, Pieza{Tipo: PFinLinea, Linea: linea})
			linea++
			i++
			continue

		case c == ',':
			// La coma solo separa; equivale a "y".
			piezas = append(piezas, Pieza{Tipo: PPalabra, Texto: ",", Clave: "y", Linea: linea})
			i++
			continue

		case c == '(':
			piezas = append(piezas, Pieza{Tipo: PAbre, Texto: "(", Linea: linea})
			i++
			continue

		case c == ')':
			piezas = append(piezas, Pieza{Tipo: PCierra, Texto: ")", Linea: linea})
			i++
			continue
		}

		if cierre, esComilla := cierreComilla[c]; esComilla {
			i++
			var sb strings.Builder
			cerrado := false
			for i < n {
				if runas[i] == cierre || runas[i] == c {
					cerrado = true
					i++
					break
				}
				if runas[i] == '\n' {
					break
				}
				if runas[i] == '\\' && i+1 < n {
					switch runas[i+1] {
					case 'n':
						sb.WriteRune('\n')
					case 't':
						sb.WriteRune('\t')
					default:
						sb.WriteRune(runas[i+1])
					}
					i += 2
					continue
				}
				sb.WriteRune(runas[i])
				i++
			}
			if !cerrado {
				return nil, nuevoError("Un texto se quedo sin cerrar.", linea,
					`Falta la comilla final. Los textos van asi: "hola"`, ClaseSintaxis)
			}
			piezas = append(piezas, Pieza{Tipo: PTexto, Texto: sb.String(), Linea: linea})
			continue
		}

		if unicode.IsDigit(c) {
			inicio := i
			for i < n && unicode.IsDigit(runas[i]) {
				i++
			}
			if i+1 < n && runas[i] == '.' && unicode.IsDigit(runas[i+1]) {
				i++
				for i < n && unicode.IsDigit(runas[i]) {
					i++
				}
			}
			texto := string(runas[inicio:i])
			num, ok := NumDesdeTexto(texto)
			if !ok {
				return nil, nuevoError("No entiendo el numero \""+texto+"\".", linea, "", ClaseSintaxis)
			}
			piezas = append(piezas, Pieza{Tipo: PNumero, Texto: texto, Num: num, Linea: linea})
			continue
		}

		if esInicioNombre(c) {
			inicio := i
			for i < n && esParteNombre(runas[i]) {
				i++
			}
			texto := string(runas[inicio:i])
			piezas = append(piezas, Pieza{Tipo: PPalabra, Texto: texto, Clave: clave(texto), Linea: linea})
			continue
		}

		pista := traduccionSimbolos[c]
		return nil, nuevoError("Fal no usa el simbolo \""+string(c)+"\".", linea, pista, ClaseSintaxis)
	}

	piezas = append(piezas, Pieza{Tipo: PFinArchivo, Linea: linea})
	return piezas, nil
}
