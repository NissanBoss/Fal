// Convierte MANUAL.md en la pagina del manual.
//
//   go run ./herramientas/manual MANUAL.md docs/manual.html
//
// Existe para que el manual no tenga dos copias. La documentacion que se
// duplica se separa: alguien arregla una frase en el markdown, nadie se
// acuerda del html, y a los tres meses la web dice una cosa y el
// repositorio otra. Aqui el markdown es el original y la pagina se fabrica
// en cada publicacion.
//
// No usa ninguna libreria de markdown a proposito, igual que el resto del
// proyecto no usa ninguna libreria de nada. Tampoco hace falta una: el
// manual escribe markdown de andar por casa, y lo que sale de aqui es
// exactamente lo que ese archivo necesita y ni una linea mas. Lo que no
// entiende, lo deja pasar como texto en vez de inventarselo.

package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type seccion struct {
	Nivel  int
	Titulo string
	Ancla  string
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "uso: manual <entrada.md> <salida.html>")
		os.Exit(2)
	}
	crudo, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "manual:", err)
		os.Exit(1)
	}

	cuerpo, indice := convertir(string(crudo))
	pagina := envolver(cuerpo, indice)
	if err := os.WriteFile(os.Args[2], []byte(pagina), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "manual:", err)
		os.Exit(1)
	}
	fmt.Printf("manual: %d secciones, %d KB\n", len(indice), len(pagina)/1024)
}

func convertir(texto string) (string, []seccion) {
	var (
		salida   strings.Builder
		indice   []seccion
		lineas   = strings.Split(strings.ReplaceAll(texto, "\r\n", "\n"), "\n")
		usadas   = map[string]int{}
		enCodigo bool
		enLista  string
		enTabla  bool
	)

	cerrarLista := func() {
		if enLista != "" {
			salida.WriteString("</" + enLista + ">\n")
			enLista = ""
		}
	}
	cerrarTabla := func() {
		if enTabla {
			salida.WriteString("</tbody></table>\n")
			enTabla = false
		}
	}

	for i := 0; i < len(lineas); i++ {
		linea := lineas[i]

		// Dentro de un bloque de codigo no se interpreta nada. Es lo primero
		// que se mira porque el manual esta lleno de ejemplos que llevan
		// almohadillas, guiones y asteriscos que no son markdown.
		if strings.HasPrefix(strings.TrimSpace(linea), "```") {
			if enCodigo {
				salida.WriteString("</code></pre>\n")
				enCodigo = false
				continue
			}
			cerrarLista()
			cerrarTabla()
			idioma := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(linea), "```"))
			clase := ""
			if idioma != "" {
				clase = ` class="lenguaje-` + escapar(idioma) + `"`
			}
			salida.WriteString("<pre" + clase + "><code>")
			enCodigo = true
			continue
		}
		if enCodigo {
			salida.WriteString(escapar(linea) + "\n")
			continue
		}

		recortada := strings.TrimSpace(linea)

		if recortada == "" {
			cerrarLista()
			cerrarTabla()
			continue
		}

		// Titulos. Cada uno se queda con un ancla estable, que es lo que
		// usan el indice de la izquierda y los enlaces internos del propio
		// manual.
		if m := regexp.MustCompile(`^(#{1,4})\s+(.*)$`).FindStringSubmatch(recortada); m != nil {
			cerrarLista()
			cerrarTabla()
			nivel := len(m[1])
			titulo := strings.TrimSpace(m[2])
			ancla := anclar(titulo)
			// Dos secciones pueden llamarse igual ("Listas" sale tres veces).
			// Sin este contador, el indice mandaria a las tres al mismo sitio.
			if n := usadas[ancla]; n > 0 {
				ancla = fmt.Sprintf("%s-%d", ancla, n)
			}
			usadas[anclar(titulo)]++
			if nivel <= 2 {
				indice = append(indice, seccion{Nivel: nivel, Titulo: titulo, Ancla: ancla})
			}
			salida.WriteString(fmt.Sprintf(
				`<h%d id="%s">%s<a class="enlazar" href="#%s" aria-label="Enlace a esta seccion">#</a></h%d>`+"\n",
				nivel, ancla, enLinea(titulo), ancla, nivel))
			continue
		}

		if recortada == "---" || recortada == "***" {
			cerrarLista()
			cerrarTabla()
			salida.WriteString("<hr>\n")
			continue
		}

		// Tablas. La segunda linea es la de guiones y no se pinta: solo
		// sirve para saber que lo de arriba era una cabecera.
		if strings.HasPrefix(recortada, "|") {
			celdas := partirFila(recortada)
			esSeparador := true
			for _, c := range celdas {
				if strings.Trim(c, "-: ") != "" {
					esSeparador = false
					break
				}
			}
			if esSeparador {
				continue
			}
			if !enTabla {
				cerrarLista()
				salida.WriteString("<table><thead><tr>")
				for _, c := range celdas {
					salida.WriteString("<th>" + enLinea(c) + "</th>")
				}
				salida.WriteString("</tr></thead><tbody>\n")
				enTabla = true
				continue
			}
			salida.WriteString("<tr>")
			for _, c := range celdas {
				salida.WriteString("<td>" + enLinea(c) + "</td>")
			}
			salida.WriteString("</tr>\n")
			continue
		}
		cerrarTabla()

		if m := regexp.MustCompile(`^[-*]\s+(.*)$`).FindStringSubmatch(recortada); m != nil {
			if enLista != "ul" {
				cerrarLista()
				salida.WriteString("<ul>\n")
				enLista = "ul"
			}
			salida.WriteString("<li>" + enLinea(m[1]) + "</li>\n")
			continue
		}
		if m := regexp.MustCompile(`^\d+\.\s+(.*)$`).FindStringSubmatch(recortada); m != nil {
			if enLista != "ol" {
				cerrarLista()
				salida.WriteString("<ol>\n")
				enLista = "ol"
			}
			salida.WriteString("<li>" + enLinea(m[1]) + "</li>\n")
			continue
		}
		cerrarLista()

		// Un parrafo puede ocupar varias lineas seguidas. Se juntan aqui en
		// vez de sacar un <p> por linea, que es lo que hacia que el texto
		// saliera cortado a lo ancho de como estuviera escrito el markdown.
		parrafo := []string{recortada}
		for i+1 < len(lineas) {
			siguiente := strings.TrimSpace(lineas[i+1])
			if siguiente == "" || empiezaBloque(siguiente) {
				break
			}
			parrafo = append(parrafo, siguiente)
			i++
		}
		salida.WriteString("<p>" + enLinea(strings.Join(parrafo, " ")) + "</p>\n")
	}
	cerrarLista()
	cerrarTabla()
	if enCodigo {
		salida.WriteString("</code></pre>\n")
	}
	return salida.String(), indice
}

// empiezaBloque dice si una linea abre algo que no puede seguir dentro del
// parrafo de arriba.
func empiezaBloque(linea string) bool {
	if strings.HasPrefix(linea, "#") || strings.HasPrefix(linea, "|") ||
		strings.HasPrefix(linea, "```") || linea == "---" {
		return true
	}
	return regexp.MustCompile(`^([-*]\s+|\d+\.\s+)`).MatchString(linea)
}

func partirFila(linea string) []string {
	linea = strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(linea), "|"), "|")
	trozos := strings.Split(linea, "|")
	for i := range trozos {
		trozos[i] = strings.TrimSpace(trozos[i])
	}
	return trozos
}

var (
	codigoEnLinea = regexp.MustCompile("`([^`]+)`")
	negrita       = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	cursiva       = regexp.MustCompile(`(^|[^*])\*([^*]+)\*`)
	enlace        = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
)

// enLinea es el formato de dentro de una frase. El orden importa: el codigo
// va primero y se guarda aparte, porque el manual esta lleno de trozos como
// `*` y `**` que son ejemplos y no ordenes de poner negrita.
func enLinea(texto string) string {
	var guardados []string
	texto = codigoEnLinea.ReplaceAllStringFunc(texto, func(m string) string {
		guardados = append(guardados, escapar(strings.Trim(m, "`")))
		return fmt.Sprintf("\x00%d\x00", len(guardados)-1)
	})

	texto = escapar(texto)
	texto = enlace.ReplaceAllStringFunc(texto, func(m string) string {
		partes := enlace.FindStringSubmatch(m)
		return `<a href="` + destino(partes[2]) + `">` + partes[1] + `</a>`
	})
	texto = negrita.ReplaceAllString(texto, "<strong>$1</strong>")
	texto = cursiva.ReplaceAllString(texto, "$1<em>$2</em>")

	for i, g := range guardados {
		texto = strings.ReplaceAll(texto, fmt.Sprintf("\x00%d\x00", i), "<code>"+g+"</code>")
	}
	return texto
}

// destino arregla los enlaces que en el repositorio apuntan a un archivo de
// al lado. En la web ese archivo no esta, asi que [LEEME.md](LEEME.md) daba
// un enlace roto en la unica pagina que la gente va a leer entera.
func destino(url string) string {
	if strings.HasPrefix(url, "#") || strings.Contains(url, "://") ||
		strings.HasPrefix(url, "mailto:") || strings.HasPrefix(url, "/") {
		return url
	}
	return "https://github.com/NissanBoss/Fal/blob/main/" + url
}

func escapar(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// anclar hace el trozo que va detras de la almohadilla. Sigue la misma regla
// que GitHub porque el propio manual ya tiene enlaces internos escritos
// contra ella, y cambiarla los romperia todos de golpe.
func anclar(titulo string) string {
	s := strings.ToLower(strings.TrimSpace(titulo))
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == ' ':
			b.WriteRune('-')
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		case strings.ContainsRune("áéíóúüñ", r):
			b.WriteRune(r)
		}
	}
	return strings.Trim(b.String(), "-")
}
