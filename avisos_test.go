package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func choquesDe(t *testing.T, fuente string) []Choque {
	t.Helper()
	piezas, err := leer(fuente)
	if err != nil {
		t.Fatal(err.Mensaje)
	}
	return choquesDeNombres(piezas)
}

func TestAvisaDeLasConstantesTapadas(t *testing.T) {
	casos := []struct {
		nombre, fuente, pisa string
		linea                int
	}{
		// Los tres fallos de verdad de FalOS, reducidos a lo minimo.
		{"una variable local, como SUBIDAS", "SUBIDAS es 2\nfuncion f\n    subidas es 0\n    devuelve SUBIDAS\nfin\n", "subidas", 3},
		{"dos globales, como HISTORIAL", "HISTORIAL es 16\nhistorial es 0\n", "historial", 2},
		{"antes de la constante", "marco es 1\nMARCO es 2\n", "marco", 1},
		{"un parametro", "MARCO es 2\nfuncion f con a y marco\n    devuelve MARCO\nfin\n", "marco", 2},
		{"una funcion sin nombre", "MARCO es 2\ng es funcion con marco\n    devuelve MARCO\nfin\n", "marco", 2},
		{"el contador de un bucle", "TOPE es 2\npara cada tope desde 1 hasta 3\n    escribe TOPE\nfin\n", "tope", 2},
		{"con otra mezcla de mayusculas", "TOPE es 2\nTope es 3\n", "Tope", 2},
		{"con tildes", "AÑO es 2026\naño es 1\n", "año", 2},
		{"con cifras y barra baja", "BYTE_2 es 1\nbyte_2 es 1\n", "byte_2", 2},
	}
	for _, c := range casos {
		choques := choquesDe(t, c.fuente)
		if len(choques) != 1 || choques[0].Nombre != c.pisa || choques[0].Linea != c.linea {
			t.Errorf("%s: esperaba %s en la linea %d y sale %+v", c.nombre, c.pisa, c.linea, choques)
		}
	}
}

func TestNoAvisaSinMotivo(t *testing.T) {
	casos := map[string]string{
		"leerla sin mas":            "MARCO es 2\nescribe marco\n",
		"escrita siempre igual":     "MARCO es 2\nMARCO es 3\n",
		"sin ninguna en mayusculas": "edad es 1\nEdad es 2\n",
		"una sola letra":            "X es 1\nx es 2\n",
		"nombres que solo se leen":  "MARCO es 2\nfuncion f con a\n    devuelve a mas marco\nfin\n",
		"un campo no es un nombre":  "MARCO es 2\nescribe marco de cosa\n",
		"la tilde cambia el nombre": "AÑO es 2026\nano es 1\n",
		"una comparacion":           "MARCO es 2\nsi marco es 2\n    escribe 1\nfin\n",
	}
	for nombre, fuente := range casos {
		if choques := choquesDe(t, fuente); len(choques) != 0 {
			t.Errorf("%s: no esperaba aviso y sale %+v", nombre, choques)
		}
	}
}

// Treinta veces el mismo nombre en una funcion no pueden tapar otro aviso.
func TestUnChoquePorNombre(t *testing.T) {
	fuente := "UNO es 1\nDOS es 2\nuno es 3\nuno es 4\ndos es 5\n"
	choques := choquesDe(t, fuente)
	if len(choques) != 2 || choques[0].Linea != 3 || choques[1].Linea != 5 {
		t.Fatalf("salen %+v", choques)
	}
}

func TestComoSeLeeElAviso(t *testing.T) {
	var b strings.Builder
	avisarChoques(&b, "pantalla.fal", "SUBIDAS es 2\nfuncion f\n    subidas es 0\nfin\n")
	salida := b.String()
	for _, trozo := range []string{"pantalla.fal", "subidas y la constante SUBIDAS (linea 1)", "linea 3 |  subidas es 0", "Pista:"} {
		if !strings.Contains(salida, trozo) {
			t.Errorf("falta %q en:\n%s", trozo, salida)
		}
	}
	b.Reset()
	avisarChoques(&b, "roto.fal", "\"sin cerrar\n")
	if b.Len() != 0 {
		t.Errorf("un programa que no se lee no da avisos, y salio:\n%s", b.String())
	}
}

// Un aviso que salta en programas sanos acaba sin leerse, asi que los
// ejemplos y las pruebas del repositorio tienen que salir limpios.
func TestLosProgramasDelRepositorioNoAvisan(t *testing.T) {
	for _, patron := range []string{"ejemplos/*.fal", "pruebas/*.fal", "nucleo/*.fal"} {
		rutas, e := filepath.Glob(patron)
		if e != nil {
			t.Fatal(e)
		}
		for _, ruta := range rutas {
			datos, e := os.ReadFile(ruta)
			if e != nil {
				t.Fatal(e)
			}
			piezas, err := leer(quitarBOM(string(datos)))
			if err != nil {
				continue
			}
			if choques := choquesDeNombres(piezas); len(choques) != 0 {
				t.Errorf("%s: %+v", ruta, choques)
			}
		}
	}
}
