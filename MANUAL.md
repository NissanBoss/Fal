# Fal

Un lenguaje de programación **completo** escrito íntegramente en español y **sin un solo
símbolo raro**: 42 palabras, 83 funciones, y todo se lee en voz alta.

Se reparte como un **único ejecutable** que no necesita nada instalado. Para ponerlo en
marcha, mira [LEEME.md](LEEME.md).

```
tipo Circulo hereda de Figura con radio
    funcion area
        devuelve redondea con (pi por radio de mi por radio de mi) y 2
    fin
fin

figuras es lista con (nuevo Circulo con 3) y (nuevo Cuadrado con 6)
escribe ordena con figuras y funcion area_de
```

## Cómo usarlo

```bash
fal ejemplos/avanzado.fal
```

Sin argumentos abre la consola interactiva:

```bash
fal
```

Para comprobar que todo sigue bien después de tocar algo:

```bash
fal --probar
```

## Qué hay en la carpeta

| | |
|---|---|
| `dist/` | El ejecutable, uno por sistema. Es todo lo que necesitas. |
| `LEEME.md` | Cómo instalarlo y usarlo |
| `pruebas/` | 13 pruebas con su salida esperada. `fal --probar` |
| `ejemplos/` | Programas de ejemplo, de lo básico a lo avanzado |
| `editor/` | Resaltado de sintaxis para VS Code. `fal --editor` |
| `*.go` | El código fuente del intérprete |

## Ejemplos

| Archivo | Qué enseña |
|---|---|
| `hola.fal` | Lo mínimo: variables, condiciones, repetir |
| `listas.fal` | Listas, bucles, funciones |
| `adivina.fal` | Un juego con `pregunta` y `mientras` |
| `completo.fal` | Recorrido por lo básico del lenguaje |
| `agenda.fal` | Programa real: objetos + archivos + menú |
| `avanzado.fal` | Clausuras, herencia, JSON, fechas, patrones, internet |
| `snake.fal` | El juego de la serpiente: `limpia`, `comparte`, listas |
| `gastos.fal` | Programa de verdad: CSV, diccionarios y dinero exacto |
| `cifrado.fal` | Cifrado Cesar, y como se rompe probando las 25 claves |
| `romanos.fal` | Numeros romanos en los dos sentidos, con autocomprobacion |
| `hanoi.fal` | Torres de Hanoi: una funcion que se llama a si misma |

---

## Las 8 reglas de diseño

1. **Todo en español.** Con tildes o sin ellas, mayúsculas o minúsculas: `FUNCIÓN`, `funcion` y `Función` son la misma palabra.
2. **Cero símbolos.** No hay `;` `{}` `==` `!=` `&&` `+` `>` `.`. Solo palabras y paréntesis.
3. **Una sola forma de hacer cada cosa.**
4. **Todo se llama igual.** `nombre de dato` o `nombre con dato1 y dato2`. Da igual si es una función tuya, una integrada o el campo de un objeto.
5. **Los bloques cierran con `fin`.** La sangría es decorativa.
6. **Una línea, una instrucción.**
7. **Se cuenta desde el 1.** El primero es el 1, el último es el `menos 1`.
8. **Los errores enseñan.** Dicen la línea, el camino que siguió el programa, y cómo se arregla.

---

# Lo básico

## Guardar valores

```
edad es 25
nombre es "Ana"
precio es 19.99
mayor es verdadero
sin_datos es nada
```

Los valores son: **números**, **textos**, **verdadero/falso**, **nada**, **listas**,
**diccionarios**, **conjuntos**, **objetos** y **funciones**.

## Escribir y preguntar

```
escribe "Hola mundo"
escribe "Cargando" sin salto      # se queda en la misma línea
escribe                            # una línea en blanco

nombre es pregunta "¿Cómo te llamas?"
edad es numero de pregunta "¿Cuántos años tienes?"
```

`pregunta` siempre devuelve texto; conviértelo con `numero de`.

## Cuentas

| Escribes | Significa |
|---|---|
| `3 mas 4` | 7 |
| `10 menos 4` | 6 |
| `3 por 4` | 12 |
| `10 entre 2` | 5 |
| `17 resto 5` | 2 (lo que sobra) |

**Los decimales son exactos.** Esta es la trampa clásica de la que Fal se libra:

```
escribe 0.1 mas 0.2              → 0.3          (en Python: 0.30000000000000004)
escribe (0.1 mas 0.2) es 0.3     → verdadero    (en Python: False)
escribe 19.99 por 3              → 59.97        exacto, sirve para dinero
```

`mas` también une textos, listas y diccionarios. Los paréntesis funcionan como en el
colegio: `(2 mas 3) por 4` da 20.

## Comparar

| Escribes | Significa |
|---|---|
| `a es b` | son iguales |
| `a no es igual a b` | son distintos |
| `a es mayor que b` | a > b |
| `a es menor o igual que b` | a <= b |
| `a esta en b` | a está dentro de b |
| `a no esta en b` | a no está dentro de b |

Para combinar: `y`, `o`, `no`.

## Comentarios

```
# El programa ignora todo lo que va después de #
```

---

# Controlar el programa

```
si edad es mayor o igual que 18
    escribe "Puedes pasar"
si no si edad es mayor que 15
    escribe "Casi"
si no
    escribe "Todavía no"
fin
```

```
repite 3 veces ... fin
para cada n desde 1 hasta 10 ... fin
para cada n desde 10 hasta 1 de 2 ... fin      # hacia atrás, de 2 en 2
para cada x en coleccion ... fin               # listas, textos, diccionarios, conjuntos
mientras condicion ... fin
```

Dentro de un bucle: `detente` lo corta, `continua` salta a la vuelta siguiente.

---

# Datos

## Listas

```
frutas es lista con "manzana" y "pera" y "uva"
vacia es lista vacia

escribe largo de frutas
escribe elemento 1 de frutas
escribe elemento menos 1 de frutas    # el último
escribe primero de frutas
escribe ultimo de frutas

agrega "kiwi" a frutas
elemento 2 de frutas es "plátano"
quita 3 de frutas
```

## Diccionarios

```
edades es diccionario vacio
elemento "ana" de edades es 25

escribe elemento "ana" de edades
escribe claves de edades
escribe valores de edades
quita "ana" de edades

precios es diccionario con "pan" y 1.20 y "leche" y 0.99    # por parejas
```

Recorrer un diccionario te da las **claves**.

## Conjuntos

Una bolsa sin repetidos y sin orden.

```
a es conjunto con 1 y 2 y 3 y 2      # queda {1, 2, 3}
b es conjunto con 3 y 4
vacio es conjunto vacio
distintas es conjunto de una_lista   # quita los repetidos de una lista

escribe union con a y b              # {1, 2, 3, 4}
escribe interseccion con a y b       # {3}
escribe diferencia con a y b         # {1, 2}
agrega 9 a a
```

---

# Funciones

```
funcion sumar con a y b
    devuelve a mas b
fin

escribe sumar con 3 y 4
escribe sumar de 3 con 4      # lo mismo
```

## Las funciones son datos

Se pueden guardar, pasar y devolver como cualquier otro valor.

```
f es funcion sumar                    # una que ya existe
g es funcion mayusculas               # también las integradas

triple es funcion con n               # una sin nombre, escrita ahí mismo
    devuelve n por 3
fin
escribe triple con 7
```

Y una función puede **recordar** dónde nació (una clausura):

```
funcion contador_desde con inicio
    cuenta es inicio
    devuelve funcion
        comparte cuenta
        cuenta es cuenta mas 1
        devuelve cuenta
    fin
fin

siguiente es contador_desde con 100
escribe siguiente     → 101
escribe siguiente     → 102
```

**La regla:** un nombre a secas que apunta a una función **sin datos** se entiende como
"llámala". Si quieres la función en sí, escribe `funcion <nombre>`.

## Trabajar con listas de golpe

```
escribe mapa con numeros y funcion con n
    devuelve n por n
fin

escribe filtra con numeros y funcion con n
    devuelve n resto 2 es 0
fin

escribe reduce con numeros y funcion con a y b
    devuelve a mas b
fin

escribe ordena con gente y funcion edad_de        # ordena por lo que diga la función
escribe cuenta_si con numeros y funcion es_par
escribe alguno con numeros y funcion es_par
escribe todos con numeros y funcion es_par
```

## Ámbito: las funciones no pisan tus variables

Lo que guardas dentro de una función **se queda dentro**:

```
total es "esto no se toca"
funcion suma con l
    total es 0            # crea SU total, no toca el de fuera
    ...
fin
```

Si de verdad quieres tocar la de fuera, hay que pedirlo:

```
funcion sube
    comparte contador
    contador es contador mas 1
fin
```

`comparte` vale tanto para una variable de arriba del todo como para la de la función que
te envuelve. En otros lenguajes hacen falta dos palabras (`global` y `nonlocal`); aquí una.

---

# Objetos

```
tipo Perro con nombre y edad
    funcion ladra
        escribe nombre de mi mas " dice guau"
    fin
fin

toby es nuevo Perro con "Toby" y 3
escribe nombre de toby         # leer un campo
nombre de toby es "Tobías"     # cambiar un campo
ladra de toby                  # llamar a una función suya
```

Dentro de un tipo, **`mi`** es el propio objeto: `nombre de mi`.

## Herencia

```
tipo Animal con nombre
    funcion habla
        devuelve "..."
    fin
    funcion presenta
        devuelve nombre de mi mas " dice " mas habla de mi
    fin
fin

tipo Perro hereda de Animal
    funcion habla
        devuelve "guau"
    fin
fin
```

`presenta de un_perro` usa la versión heredada, pero llama al `habla` del perro.

**`padre`** llama a la versión del tipo del que heredas:

```
funcion presenta
    devuelve (presenta de padre) mas " y es verde"
fin
```

## Dos funciones con nombre especial

**`crea`** es el constructor: si existe, manda ella sobre cómo se rellena el objeto.

```
tipo Loro hereda de Animal con colores
    funcion crea con n y c
        nombre de mi es n
        colores de mi es c
    fin
fin
```

**`texto`** decide cómo se ve el objeto al escribirlo.

```
funcion texto
    devuelve "el loro " mas nombre de mi
fin
```

---

# Cuando algo falla

```
intenta
    edad es numero de pregunta "¿Tu edad?"
si falla de valor
    escribe "Eso no era un número"
si falla de matematica o archivo
    escribe "Otro problema: " mas error
si falla
    escribe "Cualquier otra cosa"
finalmente
    escribe "esto se hace siempre, falle o no"
fin
```

`error` se puede escribir tal cual (sale el mensaje) o mirarle las piezas:
`mensaje de error`, `clase de error`, `linea de error`, `pista de error`.

**`relanza`** vuelve a lanzar el mismo error para que lo cace alguien de fuera.

Para provocar uno tú: `falla "no se puede dividir entre cero"`.

## Las 9 clases de error

| Clase | Cuándo |
|---|---|
| `matematica` | dividir entre cero, raíz de un negativo |
| `valor` | un dato con la forma equivocada |
| `tipo` | sumar un texto con un número |
| `nombre` | usar algo que no existe |
| `archivo` | no se pudo leer o escribir |
| `red` | fallo de internet |
| `programa` | los que lanzas tú con `falla` |
| `sintaxis` | el programa está mal escrito |
| `limite` | demasiadas llamadas anidadas |

## Los errores dicen por dónde pasaron

```
  X  No se puede dividir entre cero.

     linea 5 |  devuelve x entre 0

     Se llego aqui asi:
       dentro de b, llamada en la linea 2  |  devuelve b con x
       dentro de a, llamada en la linea 7  |  escribe a con 10

     Pista: Comprueba el valor antes de dividir.
```

---

# Archivos

```
graba con "notas.txt" y "hola\nadiós"
anexa con "notas.txt" y "\notra línea"
escribe lee de "notas.txt"
para cada l en lineas de "notas.txt" ... fin
escribe existe de "notas.txt"
borra de "notas.txt"
escribe archivos de "."
```

Las rutas se buscan **al lado de tu programa**.

---

# Varios archivos

```
usa "utiles.fal"                  # mezcla todo con lo tuyo
usa "matematicas.fal" como mate   # lo deja en su propia caja
```

Con `como`, nada choca: dos módulos pueden tener una función `doble` y tú otra distinta.
Se usan igual que un objeto: `doble de mate con 21`.

---

# Todo lo que trae puesto (83 funciones)

Todas se llaman igual: `nombre de dato` o `nombre con dato1 y dato2`.

## Textos

`largo` · `mayusculas` · `minusculas` · `capitaliza` · `recorta` · `parte` · `une` ·
`reemplaza` · `posicion` · `trozo` · `invierte` · `empieza` · `acaba` · `repetido` ·
`contiene` · `cuenta` · `numerico` · `formato`

```
parte de "a,b,c" con ","                     → [a, b, c]
une de lista con " | "
reemplaza de frase con "viejo" y "nuevo"
trozo de "abcdefg" con 2 y 4                 → bcd   (ambos incluidos)
formato de "Hola {}, tienes {}" con nombre y edad
```

## Números

`numero` · `texto` · `redondea` · `arriba` · `abajo` · `absoluto` · `raiz` · `potencia` ·
`minimo` · `maximo` · `azar`

```
redondea de pi con 4        → 3.1416
azar                        → decimal entre 0 y 1
azar de 6                   → entero de 1 a 6
azar entre 10 y 20
```

Constantes: `pi` y `e`.

## Listas

`largo` · `primero` · `ultimo` · `contiene` · `posicion` · `cuenta` · `trozo` · `ordena` ·
`invierte` · `unicos` · `copia` · `inserta` · `suma` · `promedio` · `elige` · `mezcla` ·
`rango` · `mapa` · `filtra` · `reduce` · `cuenta_si` · `alguno` · `todos`

## Diccionarios y conjuntos

`largo` · `claves` · `valores` · `contiene` · `copia` · `conjunto` · `union` ·
`interseccion` · `diferencia`

## Archivos

`lee` · `lineas` · `graba` · `anexa` · `existe` · `borra` · `archivos`

## JSON

```
texto es a_json de dato          # a texto JSON
texto es a_json con dato y verdadero   # bonito, con saltos de línea
dato es de_json de texto         # de vuelta
```

## Fechas

Son textos con la forma `"2026-07-31"`.

`ahora` · `hoy` · `anio` · `mes` · `dia` · `nombremes` · `diasemana` · `dias` · `avanza` ·
`fecha`

```
escribe diasemana de "2026-07-31"        → viernes
escribe dias con "2026-01-01" y hoy      → cuántos días han pasado
escribe avanza con hoy y 45              → la fecha 45 días después
escribe fecha con 2026 y 2 y 14          → 2026-02-14
```

## Patrones

`coincide` · `busca` · `cambia`

```
coincide con correo y "^[a-z]+@[a-z]+\.[a-z]+$"
busca con texto y "[0-9]{9}"              → todos los que encajan
cambia con texto y "[0-9]{4}" y "AAAA"
```

## Tablas (CSV)

```
filas es tabla de texto_csv       # lista de listas
escribe a_csv de filas
```

## Internet

```
pagina es pide de "example.com"
pagina es pide con url y 5        # con 5 segundos de espera máxima
```

## Tiempo y sistema

`reloj` · `espera` · `argumentos` · `termina`

## Pantalla y azar reproducible

| Función | Qué hace |
|---|---|
| `limpia` | borra la pantalla, para poder repintar en el sitio |
| `semilla de n` | fija el azar: el programa hará siempre lo mismo |

`limpia` es lo que permite hacer juegos que se redibujan, como `ejemplos/snake.fal`.

`semilla` sirve para dos cosas muy prácticas: repetir una partida exactamente igual
mientras arreglas algo, y poder **probar** un programa que usa azar. Sin ella, cada
ejecución sale distinta y no hay forma de comprobar nada:

```
semilla de 7
escribe azar entre 1 y 100      # siempre sale lo mismo
```

---

# Palabras reservadas (42)

`agrega` `como` `comparte` `con` `continua` `de` `detente` `devuelve` `entre` `es`
`escribe` `esta` `falla` `falso` `fin` `finalmente` `funcion` `hereda` `igual`
`intenta` `mas` `menos` `mi` `mientras` `nada` `no` `nuevo` `o` `padre` `para` `por`
`pregunta` `quita` `relanza` `repite` `resto` `retorna` `si` `sino` `tipo` `usa`
`verdadero`

## Y nueve palabras del lenguaje que **sí** puedes usar como variables

`y` `veces` `hasta` `mayor` `menor` `que` `cada` `en` `desde`

Son palabras del lenguaje, pero no están prohibidas como nombres. Y no hay que
elegir: la misma palabra puede ser las dos cosas en la misma línea.

```
veces es 2
repite veces veces ... fin           # repite <la variable veces> veces

hasta es 3
para cada n desde 1 hasta hasta      # hasta <la variable hasta>

mayor es 5
si 9 es mayor que mayor ... fin

en es lista con 3
escribe 3 esta en en

para cada cada en lista ... fin

y es 5
escribe y por 2                      # 10
tipo Punto con x y y                 # campos x e y, sin renombrar nada
escribe lista con y y y              # [valor de y, valor de y]
```

Funciona porque las dos cosas nunca caen en el mismo sitio. Como palabra del
lenguaje van siempre **detrás** de algo (`repite 3 veces`, `es mayor que`,
`desde 1 hasta 10`); como nombre, una variable va siempre donde se espera un
**valor**. El armador las distingue por la posición, así que no hace falta
prohibirlas.

Esto importa más de lo que parece: `veces`, `mayor` y `hasta` son nombres de
variable normalísimos, y en casi cualquier lenguaje tendrías que renombrarlos
por culpa de la gramática.

Solo hay un sitio donde tienes que echar una mano. `si x es mayor` se puede
entender de dos formas, y gana la comparación:

```
si x es mayor            # entiende "x es mayor que...", y falta el que
si x es (mayor)          # con paréntesis, compara x con la variable mayor
```

Pasa únicamente con `mayor` y `menor`, y el error te lo recuerda.

Las 83 funciones integradas tampoco son palabras reservadas: puedes tener una variable
llamada `numero`, `lista` o `suma` sin problema.

---

# La única regla rara

Cuando encadenas varios `de`, el `con` se lo queda **el primero**, que es como se lee en
voz alta:

```
parte de recorta de frase con " "
```
→ *"parte, de (recorta de frase), con espacio"*. Correcto.

Si querías lo contrario, hacen falta paréntesis:

```
primero de (parte de "a,b,c" con ",")
```

Si te equivocas, el error te enseña dónde ponerlos.

---

# Si vienes de otro lenguaje

| En otros lenguajes | En Fal |
|---|---|
| `x = 5` / `x == 5` | `x es 5` |
| `x != 5` | `x no es igual a 5` |
| `a + b` / `a % b` | `a mas b` / `a resto b` |
| `a && b` / `a \|\| b` / `!a` | `a y b` / `a o b` / `no a` |
| `x in lista` | `x esta en lista` |
| `print(x)` / `input()` | `escribe x` / `pregunta` |
| `for i in range(1,11)` | `para cada i desde 1 hasta 10` |
| `break` / `continue` | `detente` / `continua` |
| `def f(a, b):` / `return x` | `funcion f con a y b` / `devuelve x` |
| `lambda n: n*2` | `funcion con n` … `fin` |
| `map` / `filter` / `sorted(key=)` | `mapa` / `filtra` / `ordena con` |
| `global` / `nonlocal` | `comparte` (las dos) |
| `lista[0]` / `lista[-1]` | `elemento 1 de lista` / `elemento menos 1 de lista` |
| `len(x)` / `x.append(y)` | `largo de x` / `agrega y a x` |
| `dic["k"] = v` | `elemento "k" de dic es v` |
| `obj.campo` / `obj.metodo(a)` | `campo de obj` / `metodo de obj con a` |
| `class C(P):` / `self` / `super()` | `tipo C hereda de P` / `mi` / `padre` |
| `__init__` / `__str__` | `funcion crea` / `funcion texto` |
| `try/except/finally` / `raise` | `intenta / si falla / finalmente` / `falla` |
| `import m` / `import m as n` | `usa "m.fal"` / `usa "m.fal" como n` |
| `set()` / `\|` / `&` / `-` | `conjunto` / `union` / `interseccion` / `diferencia` |
| `json.dumps` / `json.loads` | `a_json` / `de_json` |
| `re.search` / `re.findall` / `re.sub` | `coincide` / `busca` / `cambia` |
| `}` o indentación | `fin` |

---

# Cómo está hecho por dentro

Fal está escrito en Go, sin ninguna librería de terceros, y se reparte como un
**único ejecutable**: no hace falta instalar nada para usarlo. Por dentro son cuatro partes:

1. **El lector** (`lector.go`): parte el texto en piezas.
2. **El armador** (`armador*.go`): junta las piezas en un árbol de instrucciones.
3. **El ejecutor** (`interprete.go`, `ejecutar.go`, `evaluar.go`): recorre el árbol y hace lo que dice.
4. **La consola** (`main.go`): ejecuta archivos y presenta los errores.

Los números exactos viven aparte, en `numero.go`: un número es un entero rápido
mientras quepa, una fracción exacta cuando tiene decimales, y solo se vuelve
aproximado si no queda más remedio (una raíz cuadrada, por ejemplo).

**Para añadir una función al lenguaje** basta una entrada en la tabla de `biblioteca.go`.
No hay que tocar ni el lector ni el armador:

```go
integrada("dobla", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
    n, err := pideNum(a[0], ln, "dobla")
    if err != nil {
        return nil, err
    }
    return multiplicaNum(n, Entero(2)), nil
})
```

Y ya puedes escribir `escribe dobla de 21`. Después, `fal --editor` actualiza
el coloreado del editor solo.

Para añadir una **instrucción** (una palabra que va al principio de una línea) hay que
tocar dos sitios: `Armador.instruccion` en `armador_ins.go` y
`Interprete.ejecutar` en `ejecutar.go`.

---

# Lo que sigue faltando

Para ser honestos, esto es lo que Fal todavía no tiene frente a Python:

- **Velocidad.** Sigue recorriendo el árbol en cada ejecución. Medido con `fib(24)`:
  unas **8,6 veces** más lento que Python nativo descontando el arranque (antes eran 55).
  Bajar de ahí es compilar a bytecode, y es el problema que menos molesta en la práctica.
- **Paquetes de otra gente.** No hay un `pip` de Fal.
- **Depurador** con puntos de parada y ejecución paso a paso.
- **Hilos y concurrencia.**
- **Generar un `.exe`**: hace falta tener Python instalado.
