// Mete la politica de contenidos dentro de las paginas de fal-lang.org.
//
//   go run ./herramientas/politica docs/index.html docs/probar.html ...
//
// GitHub Pages no deja mandar cabeceras propias, asi que la politica va en
// una etiqueta meta. Y una meta no admite frame-ancestors: contra que metan
// la pagina en un marco ajeno se pone ademas un trozo que la esconde si se
// abre dentro de otra.
//
// Los scripts de dentro de la pagina se permiten por su hash y no con
// 'unsafe-inline', que dejaria pasar tambien cualquiera que alguien consiga
// colar. Por eso esto corre al publicar, sobre la copia que se sube: un hash
// escrito a mano se queda viejo en cuanto alguien toca una linea, y la
// pagina deja de funcionar sin decir por que.
//
// Hace lo mismo que web/poner-politica.mjs en el repositorio de la academia.
// Si se cambia una regla aqui, seguramente haya que cambiarla alli.

package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const ancla = "<meta charset=\"utf-8\">\n"

// Escondida si va dentro de un marco. Lo que se ve no se puede pulsar, y lo
// que no se ve tampoco: nadie puede ponerle encima un boton falso.
const contraMarcos = "<style>html.en-marco{display:none}</style>\n" +
	"<script>if (window.top !== window.self) document.documentElement.classList.add(\"en-marco\");</script>\n"

var (
	yaPuesta = regexp.MustCompile(`(?i)<meta http-equiv="Content-Security-Policy"[^>]*>\n?`)
	scripts  = regexp.MustCompile(`(?is)<script([^>]*)>(.*?)</script>`)
	conSrc   = regexp.MustCompile(`(?i)\ssrc\s*=`)
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "uso: politica <pagina.html> [...]")
		os.Exit(2)
	}
	for _, archivo := range os.Args[1:] {
		crudo, err := os.ReadFile(archivo)
		if err != nil {
			fmt.Fprintln(os.Stderr, "politica:", err)
			os.Exit(1)
		}
		// El navegador convierte los \r\n en \n antes de calcular el hash, asi
		// que aqui tambien, o una copia sacada en Windows no coincidiria.
		html := strings.ReplaceAll(string(crudo), "\r\n", "\n")
		html = yaPuesta.ReplaceAllString(html, "")
		html = strings.Replace(html, contraMarcos, "", 1)
		if !strings.Contains(html, ancla) {
			fmt.Fprintln(os.Stderr, archivo+": no encuentro donde meterla")
			os.Exit(1)
		}
		html = strings.Replace(html, ancla, ancla+contraMarcos, 1)

		// Se calcula despues de meter el trozo contra marcos, que tambien es
		// un script de dentro y tambien necesita su hash.
		politica, cuantos := politicaDe(html)
		etiqueta := "<meta http-equiv=\"Content-Security-Policy\" content=\"" + politica + "\">\n"
		html = strings.Replace(html, ancla, ancla+etiqueta, 1)

		if err := os.WriteFile(archivo, []byte(html), 0644); err != nil {
			fmt.Fprintln(os.Stderr, "politica:", err)
			os.Exit(1)
		}
		fmt.Printf("%s: politica puesta, %d script(s) por hash\n", archivo, cuantos)
	}
}

func politicaDe(html string) (string, int) {
	var hashes []string
	for _, m := range scripts.FindAllStringSubmatch(html, -1) {
		if conSrc.MatchString(m[1]) {
			continue
		}
		suma := sha256.Sum256([]byte(m[2]))
		hashes = append(hashes, "'sha256-"+base64.StdEncoding.EncodeToString(suma[:])+"'")
	}
	reglas := []string{
		"default-src 'none'",
		// wasm-unsafe-eval es para el interprete, que es WebAssembly.
		strings.TrimSpace("script-src 'self' 'wasm-unsafe-eval' " + strings.Join(hashes, " ")),
		// Los estilos con unsafe-inline es a proposito: un estilo colado
		// maquilla la pagina pero no ejecuta nada, y atar por hash una hoja
		// que se toca a diario rompe la pagina cada vez que cambia un color.
		"style-src 'self' 'unsafe-inline'",
		// Ninguna pagina de aqui habla con la API de cuentas: si un dia lo
		// hace, tiene que decirse aqui a proposito.
		"connect-src 'self'",
		"img-src 'self' data:",
		"media-src 'self'",
		"font-src 'self'",
		"worker-src 'self'",
		"base-uri 'none'",
		"form-action 'none'",
	}
	return strings.Join(reglas, "; "), len(hashes)
}
