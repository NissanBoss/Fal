package main

import "testing"

func TestPalabrasLiberadas(t *testing.T) {
	casos := []struct{ programa, espera, porque string }{
		{"veces es 3\nescribe veces", "3", "veces como variable"},
		{"repite 3 veces\n escribe \"ok\"\nfin", "ok\nok\nok", "y repite sigue funcionando"},
		{"hasta es 10\nescribe hasta", "10", "hasta como variable"},
		{"para cada n desde 1 hasta 3\n escribe n\nfin", "1\n2\n3", "y desde/hasta siguen"},
		{"mayor es 9\nescribe mayor", "9", "mayor como variable"},
		{"a es 5\nsi a es mayor que 3\n escribe \"si\"\nfin", "si", "y la comparacion sigue"},
		{"menor es 1\nque es 2\nescribe menor mas que", "3", "menor y que como variables"},
		{"cada es \"uno\"\nescribe cada", "uno", "cada como variable"},
		{"en es 7\nescribe en", "7", "en como variable"},
		{"desde es 4\nescribe desde", "4", "desde como variable"},
		{"l es lista con 1 y 2\nescribe 1 esta en l", "verdadero", "y esta-en sigue"},
		{"l es lista con \"a\"\npara cada x en l\n escribe x\nfin", "a", "y para-cada-en sigue"},
		// El caso mas retorcido: la palabra como variable Y como palabra clave
		{"veces es 2\nrepite veces veces\n escribe \"x\"\nfin", "x\nx", "repite <variable veces> veces"},
		{"hasta es 3\npara cada n desde 1 hasta hasta\n escribe n\nfin", "1\n2\n3", "desde 1 hasta <variable hasta>"},
		{"mayor es 5\nsi 9 es mayor que mayor\n escribe \"si\"\nfin", "si", "es mayor que <variable mayor>"},
		{"en es lista con 3\nescribe 3 esta en en", "verdadero", "esta en <variable en>"},
		{"l es lista con 8\npara cada cada en l\n escribe cada\nfin", "8", "para cada <variable cada> en l"},
	}
	for _, c := range casos {
		salida, err := ejecutarEnMemoria(c.programa)
		if err != nil {
			t.Errorf("%-38s FALLA: %s", c.porque, err.Mensaje)
			continue
		}
		if salida != c.espera {
			t.Errorf("%-38s esperaba %q, dio %q", c.porque, c.espera, salida)
		}
	}
}
