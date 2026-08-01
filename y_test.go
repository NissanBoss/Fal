package main

import "testing"

func TestLaYComoVariable(t *testing.T) {
	casos := []struct{ programa, espera, porque string }{
		{"y es 5\nescribe y", "5", "guardar y leer una variable llamada y"},
		{"y es 3\nescribe y por 2", "6", "usarla en una cuenta"},
		{"y es 10\nx es y mas 5\nescribe x", "15", "leerla al asignar otra"},

		// El caso que parecia imposible: operador y variable en la misma linea.
		{"x es verdadero\ny es verdadero\nescribe x y y", "verdadero", "si x Y la variable y"},
		{"x es verdadero\ny es falso\nescribe x y y", "falso", "lo mismo, dando falso"},

		// Separador de argumentos junto a variable con el mismo nombre.
		{"y es 7\nescribe lista con 1 y y", "[1, 7]", "separador y variable seguidos"},
		{"y es 7\nescribe lista con y y y", "[7, 7]", "la variable a los dos lados"},

		// Coordenadas: el caso real que dolia.
		{"tipo Punto con x y y\n    funcion muestra\n        devuelve \"(\" mas x de mi mas \", \" mas y de mi mas \")\"\n    fin\nfin\np es nuevo Punto con 3 y 4\nescribe muestra de p", "(3, 4)", "campos x e y en un tipo"},
		{"funcion suma_xy con x y y\n    devuelve x mas y\nfin\nescribe suma_xy con 3 y 4", "7", "parametros llamados x e y"},

		// Bucles y comparaciones.
		{"para cada y desde 1 hasta 3\n    escribe y\nfin", "1\n2\n3", "variable de bucle llamada y"},
		{"y es 8\nsi y es mayor que 5\n    escribe \"grande\"\nfin", "grande", "y dentro de una condicion"},
		{"x es verdadero\ny es 8\nsi x y y es mayor que 5\n    escribe \"vale\"\nfin", "vale", "operador y variable con comparacion"},

		// Que el "y" logico de siempre siga funcionando igual.
		{"escribe verdadero y falso", "falso", "el y logico de toda la vida"},
		{"escribe suma de (lista con 1 y 2 y 3)", "6", "separador de siempre"},
	}
	for _, c := range casos {
		salida, err := ejecutarEnMemoria(c.programa)
		if err != nil {
			t.Errorf("%s -> fallo: %s", c.porque, err.Mensaje)
			continue
		}
		if salida != c.espera {
			t.Errorf("%s\n  esperaba: %q\n  obtuve:   %q", c.porque, c.espera, salida)
		}
	}
}
