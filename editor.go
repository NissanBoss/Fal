//go:build !js

package main

// Genera la extension de VS Code que colorea los archivos .fal.
//
//   fal --editor [carpeta]
//
// Se genera a partir de las tablas del propio lenguaje, asi que el
// coloreado nunca se queda desfasado: si añades una funcion, vuelves a
// lanzar esto y ya sale pintada.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Variantes con tilde que el lector acepta pero que en las tablas estan
// normalizadas.
var conTilde = map[string]string{
	"funcion": "función", "numero": "número", "mas": "más", "esta": "está",
	"vacia": "vacía", "vacio": "vacío", "mayusculas": "mayúsculas",
	"minusculas": "minúsculas", "posicion": "posición", "anio": "año",
	"interseccion": "intersección", "diasemana": "díasemana", "dia": "día",
	"ultimo": "último", "unicos": "únicos",
}

// Que letras cuentan como parte de un nombre. Sirve para que "mas" no se
// pinte dentro de "masa".
const letraNombre = `[0-9A-Za-z_À-ſ]`

func grupoPalabras(palabras []string) string {
	var formas []string
	for _, p := range palabras {
		formas = append(formas, p)
		if t, hay := conTilde[p]; hay {
			formas = append(formas, t)
		}
	}
	// Las mas largas primero, para que gane la coincidencia completa.
	sort.Slice(formas, func(i, j int) bool { return len(formas[i]) > len(formas[j]) })
	return "(?i:" + strings.Join(formas, "|") + ")"
}

func patronPalabras(palabras []string, alcance string) map[string]interface{} {
	return map[string]interface{}{
		"name":  alcance,
		"match": "(?<!" + letraNombre + ")" + grupoPalabras(palabras) + "(?!" + letraNombre + ")",
	}
}

var grupoControl = strings.Fields(`si sino fin mientras repite veces para cada en desde
	hasta detente continua devuelve retorna intenta falla finalmente relanza escribe
	pregunta agrega quita`)
var grupoDeclara = strings.Fields(`funcion tipo hereda nuevo usa como comparte`)
var grupoOperador = strings.Fields(`es no y o mas menos por entre resto mayor menor
	igual que esta de con a`)
var grupoConstante = strings.Fields(`verdadero falso nada pi`)
var grupoEspecial = strings.Fields(`mi padre error`)
var grupoEstructura = strings.Fields(`lista diccionario conjunto elemento vacia vacio`)

func generarEditor(destino string) int {
	var funciones []string
	for n := range integradas {
		funciones = append(funciones, n)
	}
	sort.Strings(funciones)

	gramatica := map[string]interface{}{
		"name":      "Fal",
		"scopeName": "source.fal",
		"patterns": []interface{}{
			map[string]interface{}{"name": "comment.line.number-sign.fal", "match": `#.*$`},
			map[string]interface{}{
				"name": "string.quoted.double.fal", "begin": `["“]`, "end": `["”]`,
				"patterns": []interface{}{map[string]interface{}{
					"name": "constant.character.escape.fal", "match": `\\.`}},
			},
			map[string]interface{}{
				"name": "string.quoted.single.fal", "begin": `['‘]`, "end": `['’]`,
				"patterns": []interface{}{map[string]interface{}{
					"name": "constant.character.escape.fal", "match": `\\.`}},
			},
			map[string]interface{}{"name": "constant.numeric.fal",
				"match": `(?<!` + letraNombre + `)\d+(\.\d+)?(?!` + letraNombre + `)`},
			patronPalabras(grupoConstante, "constant.language.fal"),
			patronPalabras(grupoEspecial, "variable.language.fal"),
			patronPalabras(grupoControl, "keyword.control.fal"),
			patronPalabras(grupoDeclara, "storage.type.fal"),
			patronPalabras(grupoEstructura, "support.type.fal"),
			patronPalabras(grupoOperador, "keyword.operator.word.fal"),
			patronPalabras(funciones, "support.function.fal"),
			// El nombre que va detras de "funcion" o "tipo".
			map[string]interface{}{
				"match": `(?i:(?<!` + letraNombre + `)(funci[oó]n|tipo)(?!` + letraNombre + `))\s+(` + letraNombre + `+)`,
				"captures": map[string]interface{}{
					"1": map[string]string{"name": "storage.type.fal"},
					"3": map[string]string{"name": "entity.name.function.fal"},
				},
			},
		},
	}

	configuracion := map[string]interface{}{
		"comments": map[string]string{"lineComment": "#"},
		"brackets": [][]string{{"(", ")"}},
		"autoClosingPairs": []interface{}{
			map[string]interface{}{"open": "(", "close": ")"},
			map[string]interface{}{"open": `"`, "close": `"`, "notIn": []string{"string"}},
		},
		"surroundingPairs": [][]string{{"(", ")"}, {`"`, `"`}},
		"indentationRules": map[string]string{
			"increaseIndentPattern": `^\s*(?i:si|sino|mientras|repite|para|funci[oó]n|tipo|intenta|finalmente)\b.*$`,
			"decreaseIndentPattern": `^\s*(?i:fin|sino|finalmente)\b.*$`,
		},
	}

	paquete := map[string]interface{}{
		"name":        "fal",
		"displayName": "Fal",
		"description": "Resaltado de sintaxis para el lenguaje Fal",
		"version":     "1.0.0", // lo exige VS Code, no es la version de Fal
		"engines":     map[string]string{"vscode": "^1.60.0"},
		"categories":  []string{"Programming Languages"},
		"contributes": map[string]interface{}{
			"languages": []interface{}{map[string]interface{}{
				"id": "fal", "aliases": []string{"Fal", "fal"},
				"extensions":    []string{".fal"},
				"configuration": "./language-configuration.json",
			}},
			"grammars": []interface{}{map[string]interface{}{
				"language": "fal", "scopeName": "source.fal",
				"path": "./syntaxes/fal.tmLanguage.json",
			}},
		},
	}

	if err := os.MkdirAll(filepath.Join(destino, "syntaxes"), 0755); err != nil {
		fmt.Fprintln(os.Stderr, "No pude crear la carpeta:", err)
		return 1
	}
	archivos := map[string]interface{}{
		filepath.Join(destino, "syntaxes", "fal.tmLanguage.json"): gramatica,
		filepath.Join(destino, "language-configuration.json"):     configuracion,
		filepath.Join(destino, "package.json"):                    paquete,
	}
	for ruta, datos := range archivos {
		texto, err := json.MarshalIndent(datos, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, "No pude generar", ruta, err)
			return 1
		}
		if err := os.WriteFile(ruta, append(texto, '\n'), 0644); err != nil {
			fmt.Fprintln(os.Stderr, "No pude escribir", ruta, err)
			return 1
		}
	}

	fmt.Println("Extension de VS Code generada en " + destino)
	fmt.Printf("  %d palabras del lenguaje, %d funciones integradas\n",
		len(reservadas), len(funciones))
	fmt.Println()
	fmt.Println("Para instalarla, copia esa carpeta dentro de:")
	fmt.Println(`  Windows      (tu carpeta de usuario)\.vscode\extensions\fal`)
	fmt.Println("  Mac y Linux  ~/.vscode/extensions/fal")
	return 0
}
