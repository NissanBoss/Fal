package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// arrancarEnQemu compila el programa en modo prueba, lo arranca sin
// pantalla y devuelve lo que salio por el puerto serie. reventado dice si
// paro el preludio (codigo 3 de isa-debug-exit) en vez de acabar bien (1).
func arrancarEnQemu(t *testing.T, programa string) (salida string, reventado bool) {
	t.Helper()
	qemu := buscarQemu()
	if qemu == "" {
		t.Skip("no hay QEMU instalado")
	}
	imagen, err := compilarNucleo(programa, true)
	if err != nil {
		t.Fatalf("no compila: %s (linea %d)", err.Mensaje, err.Linea)
	}
	carpeta := t.TempDir()
	if e := os.WriteFile(filepath.Join(carpeta, "prueba.bin"), imagen, 0644); e != nil {
		t.Fatal(e)
	}
	ctx, cancelar := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelar()
	orden := exec.CommandContext(ctx, qemu, "-kernel", "prueba.bin", "-display", "none",
		"-serial", "file:serie.txt", "-no-reboot",
		"-device", "isa-debug-exit,iobase=0xf4,iosize=0x04")
	orden.Dir = carpeta
	e := orden.Run()
	codigo := 0
	var salio *exec.ExitError
	if errors.As(e, &salio) {
		codigo = salio.ExitCode()
	}
	if ctx.Err() != nil {
		t.Fatal("QEMU no termino: el nucleo se quedo colgado")
	}
	datos, _ := os.ReadFile(filepath.Join(carpeta, "serie.txt"))
	salida = strings.ReplaceAll(string(datos), "\r", "")
	switch codigo {
	case 1:
		return salida, false
	case 3:
		return salida, true
	}
	t.Fatalf("QEMU salio con %d (%v); lo que llego a escribir:\n%s", codigo, e, salida)
	return
}

// Lo de cada prueba va en un solo arranque: QEMU tarda mas en arrancar que
// en correr todo esto.
func TestNucleoFunciona(t *testing.T) {
	programa := `
funcion fib con n
    si n es menor que 2
        devuelve n
    fin
    devuelve (fib con n menos 1) mas (fib con n menos 2)
fin

funcion suma_hasta con n
    total es 0
    para cada i desde 1 hasta n
        total es total mas i
    fin
    devuelve total
fin

contador es 0
funcion cuenta_una
    comparte contador
    contador es contador mas 1
fin

escribe fib con 20
escribe suma_hasta con 100
cuenta_una
cuenta_una
escribe contador

escribe "atras:" sin salto
para cada i desde 10 hasta 1 de 3
    escribe " " sin salto
    escribe i sin salto
fin
escribe

veces es 0
repite 4 veces
    veces es veces mas 1
fin
escribe veces

x es 0
mientras verdadero
    x es x mas 1
    si x es 3
        continua
    fin
    si x es mayor que 5
        detente
    fin
fin
escribe x

escribe 17 resto 5
escribe (0 menos 17) resto 5
escribe cociente con 0 menos 17 y 5
escribe 12 entre 4
escribe verdadero y no falso
escribe falso o (1 es 2)
escribe "abc" es "abd"
escribe "abc" no es igual a "ab"
escribe nada

letras es 0
para cada b en "Fal"
    letras es letras mas b
fin
escribe letras
escribe elemento menos 1 de "hola"
escribe largo de "dos"

memoria es region con 1048576 mas 8000000 y 4 y 8
elemento 1 de memoria es 0 menos 1
elemento 4 de memoria es 9223372036854775807
escribe elemento 1 de memoria
escribe elemento menos 1 de memoria
bytes es region con 1048576 mas 8000000 y 32
escribe elemento 1 de bytes
escribe elemento 32 de bytes
dobles es region con 1048576 mas 8000000 y 16 y 2
escribe elemento 13 de dobles
escribe (region con 0 y 10) es (region con 0 y 10)
escribe (region con 0 y 10) es (region con 0 y 10 y 2)
`
	salida, reventado := arrancarEnQemu(t, programa)
	if reventado {
		t.Fatalf("revento:\n%s", salida)
	}
	quiero := `6765
5050
2
atras: 10 7 4 1
4
6
2
-2
-3
3
verdadero
falso
falso
verdadero
nada
275
97
3
-1
9223372036854775807
255
127
65535
verdadero
falso
`
	if salida != quiero {
		t.Errorf("salida distinta.\nquiero:\n%s\nsale:\n%s", quiero, salida)
	}
}

func TestNucleoAvisa(t *testing.T) {
	casos := []struct{ programa, aviso string }{
		{"escribe \"antes\"\nescribe elemento 5 de \"hola\"", "No hay ningun elemento en esa posicion."},
		{"x es 9223372036854775807\nx es x mas 1", "El resultado no cabe en un numero del nucleo."},
		{"escribe 7 entre 2", "La division no da exacta"},
		{"escribe 7 entre 0", "No se puede dividir entre cero."},
		{"funcion sin_fin con n\n    devuelve sin_fin con n mas 1\nfin\nsin_fin con 1", "se acabo la pila"},
		{"falla \"el disco no responde\"", "el disco no responde"},
		{"falla \"esa página ya está libre\"", "esa página ya está libre"},
		{"r es region con 1048576 y 4\nelemento 1 de r es 300", "Ese numero no cabe"},
		{"escribe 1 mas \"uno\"", `"mas" necesita dos numeros, y el segundo no lo es.`},
		{"si 3\nfin", "Una condicion tiene que dar verdadero o falso."},
		{"r es region con 1073741824 y 1", "se sale del primer giga"},
		{"elemento 1 de \"hola\" es 65", "Los textos del nucleo no se pueden cambiar."},
		{"pon_en_puerto con 1016 y 256", "A un puerto se le manda un byte"},
	}
	for _, caso := range casos {
		salida, reventado := arrancarEnQemu(t, caso.programa)
		if !reventado {
			t.Errorf("%q no revento; salio:\n%s", caso.programa, salida)
			continue
		}
		if !strings.Contains(salida, caso.aviso) {
			t.Errorf("%q: esperaba %q en el aviso, y salio:\n%s", caso.programa, caso.aviso, salida)
		}
	}
}

// El controlador de interrupciones con las de los aparatos de la 32 en
// adelante, y todas tapadas menos el reloj.
const remapeaYDejaElReloj = `
pon_en_puerto con 32 y 17
pon_en_puerto con 160 y 17
pon_en_puerto con 33 y 32
pon_en_puerto con 161 y 40
pon_en_puerto con 33 y 4
pon_en_puerto con 161 y 2
pon_en_puerto con 33 y 1
pon_en_puerto con 161 y 1
pon_en_puerto con 33 y 254
pon_en_puerto con 161 y 255
`

// El reloj interrumpe de verdad mientras el programa hace cuentas, y las
// cuentas tienen que salir igual: si la entrada a la interrupcion se dejase
// un registro sin guardar, la suma saldria mal.
func TestNucleoInterrupciones(t *testing.T) {
	// El reloj a unos 1000 tics por segundo (1193182 / 1193): con los 18 de
	// fabrica la suma acababa antes del primer tic y no probaba nada.
	programa := remapeaYDejaElReloj + `
pon_en_puerto con 67 y 52
pon_en_puerto con 64 y 169
pon_en_puerto con 64 y 4
vueltas es 0
funcion reloj
    comparte vueltas
    vueltas es vueltas mas 1
    pon_en_puerto con 32 y 32
fin

funcion suma_hasta con n
    total es 0
    para cada i desde 1 hasta n
        total es total mas i
    fin
    devuelve total
fin

atiende_interrupcion con 32 y reloj
activa_interrupciones
mientras vueltas es menor que 3
    espera_interrupcion
fin
antes es vueltas
escribe suma_hasta con 2000000
escribe vueltas es mayor que antes
desactiva_interrupciones
`
	salida, reventado := arrancarEnQemu(t, programa)
	if reventado {
		t.Fatalf("revento:\n%s", salida)
	}
	if salida != "2000001000000\nverdadero\n" {
		t.Errorf("salida distinta:\n%s", salida)
	}
}

func TestNucleoAvisaDeInterrupciones(t *testing.T) {
	casos := []struct{ programa, aviso string }{
		{remapeaYDejaElReloj + "activa_interrupciones\nmientras verdadero\n    espera_interrupcion\nfin",
			"Ha llegado la interrupcion 32 y nadie la atiende."},
		// Sin remapear, el reloj de la BIOS llega por la 8 y el aviso tiene
		// que reconocer la confusion.
		{"pon_en_puerto con 33 y 254\nactiva_interrupciones\nmientras verdadero\n    espera_interrupcion\nfin",
			"Ha llegado la interrupcion de un aparato por la 8"},
		{"funcion f\nfin\natiende_interrupcion con 13 y f", "Las interrupciones de la 0 a la 31"},
	}
	for _, caso := range casos {
		salida, reventado := arrancarEnQemu(t, caso.programa)
		if !reventado {
			t.Errorf("no revento; salio:\n%s", salida)
			continue
		}
		if !strings.Contains(salida, caso.aviso) {
			t.Errorf("esperaba %q en el aviso, y salio:\n%s", caso.aviso, salida)
		}
	}
}

// Lo que deja el cargador se lee igual que la pantalla, con una region. El
// mapa de memoria va en entradas de 24 bytes, cada una con su largo delante,
// y QEMU por defecto da 128 megas.
func TestNucleoDatosDeArranque(t *testing.T) {
	programa := `
info es region con datos_de_arranque y 13 y 4
escribe ((elemento 1 de info) resto 128) es mayor o igual que 64
mapa es elemento 13 de info
fin_mapa es mapa mas elemento 12 de info
libre es 0
mientras mapa es menor que fin_mapa
    entrada es region con mapa y 6 y 4
    si elemento 6 de entrada es 1
        libre es libre mas elemento 4 de entrada
    fin
    mapa es mapa mas 4 mas elemento 1 de entrada
fin
megas es cociente con libre y 1048576
escribe megas es mayor o igual que 126 y megas es menor o igual que 128
escribe (final_del_nucleo resto 4096) es 0
escribe final_del_nucleo es mayor que 1048576
`
	salida, reventado := arrancarEnQemu(t, programa)
	if reventado {
		t.Fatalf("revento:\n%s", salida)
	}
	if salida != "verdadero\nverdadero\nverdadero\nverdadero\n" {
		t.Errorf("salida distinta:\n%s", salida)
	}
}

// Unas tablas como las que montaria un sistema: el primer mega con paginas
// de 4 KB, la 0 sin poner, el codigo de solo lectura y todo lo demas sin
// poder ejecutarse. Despues de usa_tablas el programa tiene que seguir igual,
// y lo prohibido tiene que acabar en el aviso que explica que paso.
const tablasProtegidas = `
PAGINA es 4096
NX es 0 menos 9223372036854775807 menos 1
base es 8388608
pml4 es region con base y 512 y 8
pdpt es region con base mas PAGINA y 512 y 8
pd es region con base mas 2 por PAGINA y 512 y 8
pt es region con base mas 3 por PAGINA y 512 y 8
para cada i desde 1 hasta 512
    elemento i de pml4 es 0
    elemento i de pdpt es 0
fin
elemento 1 de pml4 es base mas PAGINA mas 3
elemento 1 de pdpt es base mas 2 por PAGINA mas 3
elemento 1 de pd es base mas 3 por PAGINA mas 3
para cada t desde 2 hasta 512
    elemento t de pd es (t menos 1) por 2097152 mas 131 mas NX
fin
elemento 1 de pt es 0
para cada i desde 2 hasta 512
    d es (i menos 1) por PAGINA
    si d es mayor o igual que inicio_del_codigo y d es menor que final_del_codigo
        elemento i de pt es d mas 1
    si no
        elemento i de pt es d mas 3 mas NX
    fin
fin
escribe hay_no_ejecutar
usa_tablas con base
codigo es region con inicio_del_codigo y 4 y 4
escribe (elemento 1 de codigo) es 464367618
`

func TestNucleoMemoriaProtegida(t *testing.T) {
	salida, reventado := arrancarEnQemu(t, tablasProtegidas+"escribe \"sigue\"")
	if reventado || salida != "verdadero\nverdadero\nsigue\n" {
		t.Fatalf("con las tablas puestas no sigue igual (reventado %v):\n%s", reventado, salida)
	}
	casos := []struct{ programa, aviso string }{
		{"elemento 1 de codigo es 0", "que es de solo lectura."},
		{"cero es region con 0 y 1\nescribe elemento 1 de cero", "que no existe."},
	}
	for _, caso := range casos {
		salida, reventado := arrancarEnQemu(t, tablasProtegidas+caso.programa)
		if !reventado || !strings.Contains(salida, caso.aviso) {
			t.Errorf("%q: esperaba %q en el aviso, y salio:\n%s", caso.programa, caso.aviso, salida)
		}
	}
}

// El aviso dice la linea y la copia, igual que el interprete.
func TestNucleoAvisoConLinea(t *testing.T) {
	salida, _ := arrancarEnQemu(t, "a es 1\nb es 0\nescribe a entre b")
	if !strings.Contains(salida, "linea 3 |  escribe a entre b") {
		t.Errorf("el aviso no dice la linea:\n%s", salida)
	}
}

func TestNucleoNoCompila(t *testing.T) {
	casos := []struct{ programa, error string }{
		{"x es lista con 1 y 2", "listas"},
		{"x es 1.5", "solo hay numeros enteros"},
		{"nucleo_pantalla es 3", `"nucleo_"`},
		{"nucleo_serie con 65", `"nucleo_"`},
		{"escribe no_existe con 3", `No hay ninguna funcion que se llame "no_existe"`},
		{"funcion f con a\n    devuelve a\nfin\nescribe f con 1 y 2", `"f" necesita 1 datos y le das 2`},
		{"escribe y_esto", `No hay nada que se llame "y_esto"`},
		{"devuelve 3", `"devuelve" solo tiene sentido dentro de una funcion`},
		{"si verdadero\n    funcion g\n    fin\nfin", "arriba del todo"},
		{"funcion region\nfin", `Ya hay una funcion "region"`},
		{"escribe pregunta \"hola\"", "pregunta"},
		{"detente", "dentro de un bucle"},
		{"atiende_interrupcion con 33 y 5", "el nombre de una funcion tuya"},
		{"funcion f con x\nfin\natiende_interrupcion con 33 y f", "pide datos"},
	}
	for _, caso := range casos {
		_, err := compilarNucleo(caso.programa, true)
		if err == nil {
			t.Errorf("%q compilo y no deberia", caso.programa)
			continue
		}
		if !strings.Contains(err.Mensaje+" "+err.Pista, caso.error) {
			t.Errorf("%q: esperaba %q y dijo %q", caso.programa, caso.error, err.Mensaje)
		}
	}
}

// La misma fuente tiene que dar siempre la misma imagen: si no, no se
// puede comparar un arranque con otro ni fiarse de un binario publicado.
func TestNucleoReproducible(t *testing.T) {
	programa := "escribe \"a\"\nescribe \"b\"\nescribe \"c\"\nescribe \"a\" es \"b\""
	primera, err := compilarNucleo(programa, false)
	if err != nil {
		t.Fatal(err.Mensaje)
	}
	for i := 0; i < 5; i++ {
		otra, _ := compilarNucleo(programa, false)
		if string(otra) != string(primera) {
			t.Fatal("dos compilaciones de lo mismo dan imagenes distintas")
		}
	}
}
