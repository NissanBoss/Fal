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
