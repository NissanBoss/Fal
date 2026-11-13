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
	return arrancarConProgramas(t, programa, nil)
}

// arrancarConProgramas es arrancarEnQemu con archivos para incluye_programa.
func arrancarConProgramas(t *testing.T, programa string, archivos map[string]string, disco ...byte) (salida string, reventado bool) {
	t.Helper()
	qemu := buscarQemu()
	if qemu == "" {
		t.Skip("no hay QEMU instalado")
	}
	imagen, err := compilarImagen(programa, opcionesNucleo{prueba: true, leer: leerDe(archivos)})
	if err != nil {
		t.Fatalf("no compila: %s (linea %d)", err.Mensaje, err.Linea)
	}
	carpeta := t.TempDir()
	if e := os.WriteFile(filepath.Join(carpeta, "prueba.bin"), imagen, 0644); e != nil {
		t.Fatal(e)
	}
	ctx, cancelar := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelar()
	argumentos := []string{"-kernel", "prueba.bin", "-display", "none",
		"-serial", "file:serie.txt", "-no-reboot",
		"-device", "isa-debug-exit,iobase=0xf4,iosize=0x04"}
	if len(disco) > 0 {
		if e := os.WriteFile(filepath.Join(carpeta, "disco.img"), disco, 0644); e != nil {
			t.Fatal(e)
		}
		argumentos = append(argumentos, "-drive", "file=disco.img,format=raw,if=ide")
	}
	orden := exec.CommandContext(ctx, qemu, argumentos...)
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

func leerDe(archivos map[string]string) func(string) (string, error) {
	return func(ruta string) (string, error) {
		fuente, hay := archivos[ruta]
		if !hay {
			return "", os.ErrNotExist
		}
		return fuente, nil
	}
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

// Un sistema minimo: unas tablas con el programa en el anillo 3, un cargador y
// la llamada que entra en el. El numero 99 solo lo usa el nucleo para saltar
// al programa, y cada programa vuelve con la llamada 0 o con un fallo.
const anillo3 = `
USUARIO es 7
raiz es final_del_nucleo
tablas es region con raiz y 4096 y 8
para cada i desde 1 hasta 4096
    elemento i de tablas es 0
fin
arriba es region con raiz y 512 y 8
elemento 1 de arriba es raiz mas 4096 mas USUARIO
medio es region con raiz mas 4096 y 512 y 8
elemento 1 de medio es raiz mas 8192 mas USUARIO
directorio es region con raiz mas 8192 y 512 y 8
para cada i desde 1 hasta 512
    elemento i de directorio es (i menos 1) por 2097152 mas 131
fin
paginas es region con raiz mas 12288 y 512 y 8
elemento 385 de directorio es raiz mas 12288 mas USUARIO
usa_tablas con raiz

tope es raiz mas 20480
memoria es raiz mas 24576
vuelta es 0
entrada es 0
banderas es 2
fallos es 0

funcion cargar con imagen
    comparte entrada
    entrada es elemento 2 de imagen
    total es elemento 4 de imagen
    cuantas es cociente con (total mas 4095) y 4096
    para cada i desde 1 hasta cuantas
        elemento i de paginas es memoria mas (i menos 1) por 4096 mas USUARIO
    fin
    palabras es cociente con total y 8
    destino es region con memoria y palabras y 8
    para cada i desde 1 hasta palabras
        elemento i de destino es 0
    fin
    para cada i desde 1 hasta largo de imagen
        elemento i de destino es elemento i de imagen
    fin
    usa_tablas con raiz
fin

funcion llamada
    comparte vuelta
    marco es region con contexto y 20 y 8
    numero es elemento 15 de marco
    si numero es 99
        vuelta es contexto
        otro es region con tope menos 160 y 20 y 8
        para cada i desde 1 hasta 20
            elemento i de otro es 0
        fin
        elemento 16 de otro es entrada
        elemento 17 de otro es 27
        elemento 18 de otro es banderas
        elemento 20 de otro es 35
        pila_de_entrada con tope
        pon_contexto con tope menos 160
    si no si numero es 1
        pon_en_puerto con 1016 y elemento 9 de marco
        elemento 15 de marco es 77
    si no si numero es 0
        escribe "salio con"
        escribe elemento 9 de marco
        pon_contexto con vuelta
    fin
fin

funcion fallo con numero y codigo y direccion
    comparte fallos
    fallos es fallos mas 1
    escribe "fallo"
    escribe numero
    escribe codigo
    escribe direccion
    pon_contexto con vuelta
fin

atiende_llamadas con llamada
atiende_fallos con fallo
`

func TestNucleoAnillo3(t *testing.T) {
	programa := anillo3 + `
cargar con (incluye_programa de "hola.fal")
llama_al_sistema con 99
escribe "de vuelta en el nucleo"

cargar con (incluye_programa de "apaga.fal")
llama_al_sistema con 99
cargar con (incluye_programa de "espia.fal")
llama_al_sistema con 99
cargar con (incluye_programa de "duerme.fal")
llama_al_sistema con 99
cargar con (incluye_programa de "tablas.fal")
llama_al_sistema con 99
cargar con (incluye_programa de "revienta.fal")
llama_al_sistema con 99
cargar con (incluye_programa de "pila.fal")
llama_al_sistema con 99
escribe fallos
`
	salida, reventado := arrancarConProgramas(t, programa, map[string]string{
		"hola.fal": `
suma es 0
para cada i desde 1 hasta 10
    suma es suma mas i
fin
escribe "hola desde el anillo 3"
escribe suma
llama_al_sistema con 0 y (llama_al_sistema con 1 y 33)
`,
		"apaga.fal": `
escribe "voy"
pon_en_puerto con 100 y 254
escribe "he podido"
`,
		"espia.fal": `
nucleo es region con 1048576 y 4
escribe elemento 1 de nucleo
escribe "he podido"
`,
		"duerme.fal": `
espera_interrupcion
escribe "he podido"
`,
		"tablas.fal": `
usa_tablas con 4096
escribe "he podido"
`,
		"revienta.fal": `
x es 1 entre 0
`,
		"pila.fal": `
funcion sin_fin con n
    devuelve (sin_fin con n mas 1) mas 1
fin
escribe sin_fin con 1
`,
	})
	if reventado {
		t.Fatalf("el nucleo se paro:\n%s", salida)
	}
	if strings.Contains(salida, "he podido") {
		t.Errorf("un programa hizo lo que no podia:\n%s", salida)
	}
	quiere := []string{
		"hola desde el anillo 3\n55\n!salio con\n77\n", // la llamada 1 devuelve lo que puso el nucleo
		"de vuelta en el nucleo\n",
		"voy\nfallo\n13\n0\n",     // un puerto: fallo general
		"fallo\n14\n5\n1048576\n", // leer el nucleo: la pagina existe, sin permiso
		"entre 0",
		"salio con\n1\n",
		"se acabo la pila",
		"\n4\n", // los cuatro intentos acabaron en el manejador
	}
	for _, q := range quiere {
		if !strings.Contains(salida, q) {
			t.Errorf("no aparece %q en:\n%s", q, salida)
		}
	}
}

// El reloj tiene que sacar de la CPU a un programa que no la suelta nunca.
func TestNucleoEchaAUnPrograma(t *testing.T) {
	programa := anillo3 + `
tics es 0
en_programa es 0

funcion al_tic
    comparte tics y en_programa
    pon_en_puerto con 32 y 32
    tics es tics mas 1
    marco es region con contexto y 20 y 8
    si (elemento 17 de marco) es 27
        en_programa es en_programa mas 1
    fin
    si tics es 10
        pon_contexto con vuelta
    fin
fin

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
atiende_interrupcion con 32 y al_tic

banderas es 514
cargar con (incluye_programa de "bucle.fal")
llama_al_sistema con 99
desactiva_interrupciones
escribe tics
escribe en_programa
`
	salida, reventado := arrancarConProgramas(t, programa, map[string]string{
		"bucle.fal": `
mientras verdadero
fin
`,
	})
	if reventado {
		t.Fatalf("el nucleo se paro:\n%s", salida)
	}
	if !strings.Contains(salida, "10\n10\n") {
		t.Errorf("el reloj no echo al programa:\n%s", salida)
	}
}

func TestNucleoProgramasNoCompilan(t *testing.T) {
	casos := []struct {
		nombre, sistema, archivo, espera string
	}{
		{"ruta de una cuenta", `r es "a.fal"
incluye_programa de r`, "", "entre comillas"},
		{"archivo que no esta", `incluye_programa de "b.fal"`, "", "No pude abrir"},
		{"programa con un fallo", `incluye_programa de "a.fal"`, "x es\n", "a.fal, linea"},
		{"un programa no atiende interrupciones", `incluye_programa de "a.fal"`,
			"funcion f\nfin\natiende_interrupcion con 33 y f\n", "solo tiene sentido en el nucleo"},
		{"un programa no incluye otro", `incluye_programa de "a.fal"`,
			"incluye_programa de \"c.fal\"\n", "solo tiene sentido en el nucleo"},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			_, err := compilarImagen(caso.sistema, opcionesNucleo{prueba: true,
				leer: leerDe(map[string]string{"a.fal": caso.archivo})})
			if err == nil {
				t.Fatal("deberia no compilar")
			}
			if !strings.Contains(err.Mensaje, caso.espera) {
				t.Errorf("el aviso es %q", err.Mensaje)
			}
		})
	}
}

// Lo que abre el mapa de puertos para uno no puede abrirlo para otro, y lo que
// no se ha abierto sigue cerrado: el 113 y el 100 estan en el mismo byte que
// sus vecinos, asi que aqui se ve si el bit es el que toca.
func TestNucleoMapaDePuertos(t *testing.T) {
	programa := anillo3 + `
mapa es region con mapa_de_puertos y 8192
escribe elemento 1 de mapa
escribe elemento 8192 de mapa
elemento 15 de mapa es 254
cargar con (incluye_programa de "abierto.fal")
llama_al_sistema con 99
elemento 15 de mapa es 255
cargar con (incluye_programa de "abierto.fal")
llama_al_sistema con 99
escribe fallos
`
	salida, reventado := arrancarConProgramas(t, programa, map[string]string{
		"abierto.fal": `
pon_en_puerto con 112 y 0
escribe "puerto abierto"
pon_en_puerto con 113 y 0
escribe "no deberia llegar"
`,
	})
	if reventado {
		t.Fatalf("revento:\n%s", salida)
	}
	if !strings.HasPrefix(salida, "255\n255\n") {
		t.Errorf("el mapa no nace cerrado:\n%s", salida)
	}
	if strings.Count(salida, "puerto abierto") != 1 || strings.Contains(salida, "no deberia") {
		t.Errorf("el mapa no manda:\n%s", salida)
	}
}

// Un disco ATA da los datos de 16 en 16 bits. Si la lectura fuera de un byte,
// aqui saldria 72 o nada, porque el disco no adelanta su cola.
func TestNucleoLeePalabra(t *testing.T) {
	programa := `
pon_en_puerto con 502 y 224
pon_en_puerto con 498 y 1
pon_en_puerto con 499 y 0
pon_en_puerto con 500 y 0
pon_en_puerto con 501 y 0
pon_en_puerto con 503 y 32
mientras ((lee_puerto de 503) resto 16) es menor que 8
fin
escribe lee_palabra de 496
escribe lee_palabra de 496
`
	disco := make([]byte, 1024)
	copy(disco, "HOLA")
	salida, reventado := arrancarConProgramas(t, programa, nil, disco...)
	if reventado {
		t.Fatalf("revento:\n%s", salida)
	}
	if salida != "20296\n16716\n" {
		t.Errorf("no leyo el disco de dos en dos bytes:\n%q", salida)
	}
}

// El bus PCI solo habla de 32 en 32 bits. El primer aparato de QEMU es siempre
// el mismo puente, asi que si esto sale mal es la instruccion y no el aparato.
func TestNucleoPuertosGrandes(t *testing.T) {
	salida, reventado := arrancarEnQemu(t, `
pon_doble con 3320 y 2147483648
escribe lee_doble de 3324
pon_palabra con 462 y 0
escribe (lee_palabra de 463) es mayor o igual que 45248
`)
	if reventado {
		t.Fatalf("revento:\n%s", salida)
	}
	if salida != "305627270\nverdadero\n" {
		t.Errorf("no leyo lo que esperaba:\n%q", salida)
	}
}

func TestNucleoCopiaRegion(t *testing.T) {
	salida, reventado := arrancarEnQemu(t, `
r es region con 2097152 y 12 y 2
para cada i desde 1 hasta 12
    elemento i de r es i
fin
copia_region con r y 1 y r y 3 y 10
escribe elemento 1 de r
escribe elemento 10 de r
escribe elemento 12 de r
para cada i desde 1 hasta 12
    elemento i de r es i
fin
copia_region con r y 3 y r y 1 y 10
escribe elemento 3 de r
escribe elemento 12 de r
escribe elemento 2 de r
copia_region con r y 1 y r y 1 y 0
b es region con 2097408 y 21
para cada i desde 1 hasta 21
    elemento i de b es i
fin
copia_region con b y 2 y b y 1 y 20
escribe elemento 21 de b
escribe elemento 2 de b
`)
	if reventado {
		t.Fatalf("revento:\n%s", salida)
	}
	if salida != "3\n12\n12\n1\n10\n2\n20\n1\n" {
		t.Errorf("copio mal:\n%q", salida)
	}
}

func TestNucleoCopiaRegionNoSeSale(t *testing.T) {
	for _, caso := range []struct{ nombre, codigo, aviso string }{
		{"destino", "copia_region con a y 8 y b y 1 y 4", "se sale"},
		{"origen", "copia_region con a y 1 y b y 8 y 4", "se sale"},
		{"cuenta negativa", "copia_region con a y 1 y b y 1 y 0 menos 1", "se sale"},
		{"desde cero", "copia_region con a y 0 y b y 1 y 1", "se sale"},
		{"cuenta enorme", "copia_region con a y 2 y b y 2 y 9223372036854775807", "se sale"},
		{"anchos distintos", "copia_region con a y 1 y c y 1 y 1", "mismo ancho"},
		{"un numero", "copia_region con 5 y 1 y b y 1 y 1", "de una region a otra"},
	} {
		t.Run(caso.nombre, func(t *testing.T) {
			programa := `
a es region con 2097152 y 10
b es region con 2097408 y 10
c es region con 2097664 y 10 y 2
` + caso.codigo + `
escribe "no deberia llegar"
`
			salida, reventado := arrancarEnQemu(t, programa)
			if !reventado || !strings.Contains(salida, caso.aviso) {
				t.Errorf("reventado=%v:\n%s", reventado, salida)
			}
		})
	}
}

func TestNucleoIncluyeArchivo(t *testing.T) {
	salida, reventado := arrancarConProgramas(t, anillo3+`
letra es incluye_archivo de "letra.txt"
escribe largo de letra
escribe elemento 2 de letra
cargar con (incluye_programa de "usa.fal")
llama_al_sistema con 99
`, map[string]string{
		"letra.txt": "hola",
		"usa.fal": `
dato es incluye_archivo de "letra.txt"
escribe elemento 1 de dato
`,
	})
	if reventado {
		t.Fatalf("revento:\n%s", salida)
	}
	if !strings.HasPrefix(salida, "4\n111\n") || !strings.Contains(salida, "104") {
		t.Errorf("no llego el archivo:\n%s", salida)
	}
}

// Si el sistema estaba en modo grafico, el aviso rojo tiene que seguir saliendo.
// La vuelta al texto no se ve desde aqui, pero no puede parar el aviso.
func TestNucleoAvisoDesdeGraficos(t *testing.T) {
	salida, reventado := arrancarEnQemu(t, `
pon_palabra con 462 y 1
pon_palabra con 463 y 960
pon_palabra con 462 y 2
pon_palabra con 463 y 600
pon_palabra con 462 y 3
pon_palabra con 463 y 32
pon_palabra con 462 y 4
pon_palabra con 463 y 193
x es 1 entre 0
`)
	if !reventado || !strings.Contains(salida, "entre cero") {
		t.Errorf("reventado=%v:\n%s", reventado, salida)
	}
}

func TestBytesDeArchivos(t *testing.T) {
	carpeta := t.TempDir()
	viejo, _ := os.Getwd()
	if e := os.Chdir(carpeta); e != nil {
		t.Fatal(e)
	}
	defer os.Chdir(viejo)
	os.WriteFile("hola.fal", []byte("escribe \"hola\"\n"), 0644)
	salida, err := ejecutarEnMemoria(`
anexa_bytes con "b.bin" y (lista con 0 y 255 y 195)
anexa_bytes con "b.bin" y (lista con 7)
b es lee_bytes de "b.bin"
escribe largo de b
escribe elemento 2 de b
imagen es compila_programa de "hola.fal"
escribe (largo de imagen) es mayor que 100
escribe elemento 1 de imagen
`)
	if err != nil {
		t.Fatal(err.Mensaje)
	}
	if salida != "4\n255\nverdadero\n70" {
		t.Errorf("salio %q", salida)
	}
	if _, err := ejecutarEnMemoria(`anexa_bytes con "b.bin" y (lista con 256)`); err == nil {
		t.Error("un 256 no es un byte")
	}
}

// El cuarto de vuelta de ChaCha20 con los numeros del RFC 8439 (2.1.1). Si
// sale bien, las tres operaciones hacen lo que dicen, incluido que la suma
// de la vuelta al pasar de 32 bits.
func TestNucleoBitsDobles(t *testing.T) {
	salida, reventado := arrancarEnQemu(t, `
s es region con 2097152 y 4 y 8
elemento 1 de s es 286331153
elemento 2 de s es 16909060
elemento 3 de s es 2609737539
elemento 4 de s es 19088743
funcion mezcla con a y b y n
    elemento a de s es gira_doble con (o_exclusivo con (elemento a de s) y (elemento b de s)) y n
fin
funcion suma con a y b
    elemento a de s es suma_doble con (elemento a de s) y (elemento b de s)
fin
suma con 1 y 2
mezcla con 4 y 1 y 16
suma con 3 y 4
mezcla con 2 y 3 y 12
suma con 1 y 2
mezcla con 4 y 1 y 8
suma con 3 y 4
mezcla con 2 y 3 y 7
para cada i desde 1 hasta 4
    escribe elemento i de s
fin
escribe suma_doble con 4294967295 y 2
escribe gira_doble con 2147483648 y 1
escribe o_exclusivo con 4294967295 y 0
`)
	if reventado {
		t.Fatalf("revento:\n%s", salida)
	}
	if salida != "3928658676\n3407673550\n1166100270\n1484899515\n1\n1\n4294967295\n" {
		t.Errorf("el cuarto de vuelta no da lo del RFC:\n%q", salida)
	}
}

func TestNucleoBitsDoblesFuera(t *testing.T) {
	for _, programa := range []string{
		"escribe o_exclusivo con (0 menos 1) y 3",
		"escribe suma_doble con 4294967296 y 0",
		"escribe gira_doble con 5 y 32",
		`escribe o_exclusivo con "a" y 3`,
	} {
		salida, reventado := arrancarEnQemu(t, programa)
		if !reventado {
			t.Errorf("%s no revento:\n%q", programa, salida)
		}
	}
}

// Lo que pide BLAKE2b: sumar y multiplicar dando la vuelta a los 64 bits, y
// partir un numero en sus dos mitades sin signo.
func TestNucleoCuadruples(t *testing.T) {
	salida, reventado := arrancarEnQemu(t, `
escribe suma_cuadruple con 9223372036854775807 y 1
escribe producto_cuadruple con 4294967296 y 4294967296
escribe producto_cuadruple con 3 y (0 menos 1)
escribe o_exclusivo_cuadruple con (0 menos 1) y 0
escribe gira_cuadruple con 1 y 63
escribe gira_cuadruple con (0 menos 2) y 1
escribe mitad_baja con (0 menos 1)
escribe mitad_alta con (0 menos 1)
escribe mitad_alta con 4294967296
`)
	if reventado {
		t.Fatalf("revento:\n%s", salida)
	}
	quiero := "-9223372036854775808\n0\n-3\n-1\n-9223372036854775808\n-3\n4294967295\n4294967295\n1\n"
	if salida != quiero {
		t.Errorf("las cuentas de 64 bits no dan:\n%q", salida)
	}
}

func TestNucleoCuadruplesFuera(t *testing.T) {
	for _, programa := range []string{
		"escribe gira_cuadruple con 5 y 64",
		"escribe gira_cuadruple con 5 y (0 menos 1)",
		`escribe suma_cuadruple con "a" y 3`,
		`escribe mitad_baja con verdadero`,
	} {
		salida, reventado := arrancarEnQemu(t, programa)
		if !reventado {
			t.Errorf("%s no revento:\n%q", programa, salida)
		}
	}
}

// El QEMU de las pruebas no siempre trae RDRAND. Si lo trae, dos numeros
// seguidos no salen iguales; si no, azar_del_procesador revienta en vez de
// devolver algo que parezca azar.
func TestNucleoAzar(t *testing.T) {
	salida, reventado := arrancarEnQemu(t, `
si hay_azar
    escribe (azar_del_procesador) no es igual a (azar_del_procesador)
si no
    escribe "sin azar"
    escribe azar_del_procesador
fin
`)
	switch {
	case salida == "verdadero\n" && !reventado:
	case strings.HasPrefix(salida, "sin azar\n") && reventado:
	default:
		t.Errorf("el azar no se porta:\n%q (reventado %v)", salida, reventado)
	}
}

func TestNucleoUsa(t *testing.T) {
	salida, reventado := arrancarConProgramas(t, `usa "comun/dobla.fal"
escribe dobla de 21
escribe dobla de (dobla de 1)
`, map[string]string{"comun/dobla.fal": "funcion dobla con x\n    devuelve x mas x\nfin\n"})
	if reventado || salida != "42\n4\n" {
		t.Errorf("usa no trae la funcion:\n%q", salida)
	}
}

func TestNucleoUsaSoloFunciones(t *testing.T) {
	archivos := map[string]string{
		"global.fal": "cuenta es 3\nfuncion dobla con x\n    devuelve x mas x\nfin\n",
		"roto.fal":   "funcion dobla con x\n    devuelve x mas\nfin\n",
	}
	for programa, quiero := range map[string]string{
		`usa "global.fal"`:                          "solo puede tener funciones",
		`usa "roto.fal"`:                            "roto.fal, linea 2",
		`usa "no_esta.fal"`:                         "No pude abrir",
		"si verdadero\n    usa \"global.fal\"\nfin": "arriba del todo",
	} {
		_, err := compilarImagen(programa, opcionesNucleo{prueba: true, leer: leerDe(archivos)})
		if err == nil || !strings.Contains(err.Mensaje, quiero) {
			t.Errorf("%q: quiero un error con %q y sale %v", programa, quiero, err)
		}
	}
}
