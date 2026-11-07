# Fal sin sistema debajo

`fal --nucleo` convierte un programa de Fal en una imagen que un ordenador
arranca directamente, sin Windows, Linux ni nada por debajo. Es lo que hace
falta para escribir un sistema operativo en Fal.

```
fal --nucleo sistema.fal     deja al lado sistema.bin
fal --arranca sistema.fal    lo compila y lo arranca en QEMU
```

`--arranca` necesita [QEMU](https://www.qemu.org), un ordenador virtual. Lo que
el programa manda con `escribe` sale en la terminal; la ventana de QEMU es la
pantalla del ordenador.

Todo sale del propio `fal`: no hace falta gcc, ni nasm, ni un enlazador. Fal
traduce el programa a código máquina x86-64 él solo.

## Un ejemplo

```
pantalla es region con 753664 y 2000 y 2

para cada i desde 1 hasta 2000
    elemento i de pantalla es 7936       # espacio, blanco sobre azul
fin

mensaje es "Hola, soy un sistema escrito en Fal"
para cada i desde 1 hasta largo de mensaje
    elemento 164 mas i de pantalla es 7936 mas (elemento i de mensaje)
fin

escribe "pantalla pintada"
```

La pantalla de texto vive en la dirección 753664 (`0xB8000`): 80 por 25
casillas de 2 bytes, la letra y su color.

## Qué hay

Un Fal más pequeño que el del intérprete:

- **Números enteros** de 64 bits. Si una cuenta se sale, el programa se para y
  lo dice; nunca da la vuelta en silencio.
- **Textos** escritos en el programa. Se leen byte a byte con `elemento`, se
  miden con `largo` y se recorren con `para cada`.
- `verdadero`, `falso` y `nada`.
- **Regiones**: ventanas a la memoria. Son la única forma de tocarla.
- Funciones, `si`, `mientras`, `repite`, `para cada`, `detente`, `continua`,
  `comparte`, `falla` y `escribe`.

`entre` sigue siendo exacto, como en el resto de Fal: si `7 entre 2` no da
entero, el programa se para en vez de inventarse un decimal que no existe.
Para la parte entera está `cociente con 7 y 2`, y para lo que sobra,
`7 resto 2`.

### Lo que el núcleo trae hecho

| Escribes | Hace |
|---|---|
| `region con inicio y cuantos` | una ventana de `cuantos` bytes desde `inicio` |
| `region con inicio y cuantos y ancho` | lo mismo con elementos de 1, 2, 4 u 8 bytes |
| `elemento i de r` | lee el elemento `i` de un texto o una región |
| `elemento i de r es v` | lo escribe (solo en regiones) |
| `largo de x` | cuántos elementos tiene |
| `cociente con a y b` | la división entera |
| `lee_puerto de p` | lee un byte de un puerto de entrada y salida |
| `lee_palabra de p` | lee 16 bits de un puerto, como pide un disco ATA |
| `pon_en_puerto con p y v` | manda un byte a un puerto |
| `pon_palabra con p y v` | manda 16 bits a un puerto |
| `pon_doble con p y v`, `lee_doble de p` | manda o lee 32 bits, como pide el bus PCI |
| `copia_region con d y i y o y j y n` | copia `n` elementos de `o` desde `j` a `d` desde `i`, de golpe y comprobado |
| `congela` | para el procesador para siempre |
| `atiende_interrupcion con n y f` | la función `f` atiende la interrupción `n` |
| `activa_interrupciones` | deja que lleguen |
| `desactiva_interrupciones` | las cierra |
| `espera_interrupcion` | duerme el procesador hasta que llegue una |
| `datos_de_arranque` | dónde dejó el cargador sus datos, o 0 si no hay |
| `final_del_nucleo` | la primera dirección libre después del programa, en un múltiplo de 4096 |
| `inicio_del_codigo`, `final_del_codigo` | dónde está el código del programa (con sus textos) |
| `hay_no_ejecutar` | si el procesador sabe prohibir que se ejecute una página |
| `usa_tablas con d` | pasa a usar las tablas de páginas que empiezan en `d` |
| `incluye_programa de "ruta"` | compila otro `.fal` como programa y lo mete en la imagen |
| `incluye_archivo de "ruta"` | mete un archivo en la imagen y da una región de bytes de solo lectura |
| `atiende_llamadas con f` | `f` atiende las llamadas al sistema de los programas |
| `atiende_fallos con f` | `f` atiende lo que rompan los programas |
| `llama_al_sistema con n y a y b` | desde un programa: pide algo al sistema |
| `contexto` | dónde está guardado lo que se acaba de interrumpir |
| `pon_contexto con d` | al salir de la interrupción, sigue por otro contexto |
| `pila_de_entrada con d` | la pila a la que salta el procesador al entrar desde un programa |

Se cuenta desde el 1 y `menos 1` es el último, igual que en todo Fal.

## La memoria que hay

El núcleo no reparte memoria: eso lo decide el sistema. Lo que da es lo que
hace falta para saber cuánta hay y qué parte está ya ocupada.

`datos_de_arranque` es la dirección de la [información multiboot](https://www.gnu.org/software/grub/manual/multiboot/multiboot.html#Boot-information-format)
que dejan QEMU y GRUB, y se lee con una región como todo lo demás. Ahí está
el mapa de la memoria: qué trozos son RAM libre y cuáles son de la BIOS o de
los aparatos.

```
info es region con datos_de_arranque y 13 y 4
mapa es elemento 13 de info
fin_mapa es mapa mas elemento 12 de info
mientras mapa es menor que fin_mapa
    entrada es region con mapa y 6 y 4
    # desde: elemento 2, cuanto: elemento 4, tipo (1 es libre): elemento 6
    mapa es mapa mas 4 mas elemento 1 de entrada
fin
```

`final_del_nucleo` dice dónde acaba lo que ocupa el programa: el código, sus
tablas, la pila y las variables. De ahí para arriba, lo que el mapa diga que
está libre es del sistema.

## Proteger la memoria

Las tablas de páginas del arranque dejan todo el primer giga para leer,
escribir y ejecutar. El sistema puede montar las suyas, con los permisos que
quiera para cada página de 4 KB, y pasar a usarlas con `usa_tablas`. Lo
normal es:

- el código, entre `inicio_del_codigo` y `final_del_codigo`, de solo lectura;
- todo lo demás, sin poder ejecutarse (el bit 63 de cada entrada, si
  `hay_no_ejecutar`);
- la página 0, sin poner, para que tocar la dirección 0 pare el sistema.

El arranque ya deja encendido lo necesario para que el procesador lo haga
cumplir también al propio núcleo, que si no se salta las páginas de solo
lectura. Y cuando algo choca con un permiso, el aviso lo dice con palabras:
"quiso escribir en 0x100000, que es de solo lectura". Las pruebas montan unas
tablas así y comprueban las dos cosas: que el programa sigue igual y que lo
prohibido acaba en ese aviso.

## Interrupciones

Una interrupción es un aparato que avisa: una tecla pulsada, un tic del
reloj. En vez de preguntarle al teclado sin parar si hay algo, el sistema
duerme con `espera_interrupcion` y el procesador lo despierta cuando llega.

```
pulsadas es 0

funcion teclado
    comparte pulsadas
    codigo es lee_puerto de 96
    pulsadas es pulsadas mas 1
    pon_en_puerto con 32 y 32       # "ya está" al controlador
fin

atiende_interrupcion con 33 y teclado
activa_interrupciones
mientras verdadero
    espera_interrupcion
fin
```

La función que atiende una interrupción es una función normal de Fal, sin
datos. El compilador le pone delante una entrada que guarda los registros y
detrás un `iretq`, así que el programa interrumpido no nota nada. Las
pruebas lo comprueban con el reloj interrumpiendo cientos de veces a mitad de
una suma.

Lo que no hace el compilador es configurar los aparatos: remapear el
controlador de interrupciones (el 8259), destapar las líneas y avisarle de
que ya se atendió cada una. Eso es trabajo del sistema, y se hace con
`pon_en_puerto`.

Las interrupciones de la 0 a la 31 son las excepciones del procesador y no se
pueden atender: si salta una, el núcleo para y dice cuál fue y dónde. Si llega
una interrupción que nadie atiende, también lo dice, con su número. Y si un
aparato llega por el número de una excepción, que es lo que pasa al activar
las interrupciones sin remapear el controlador, lo reconoce y lo explica en
vez de dejarte con un "fallo doble" que no lo es.

## Programas sin permisos

El procesador tiene cuatro anillos, y usa dos: el núcleo va en el 0 y los
programas, en el 3. En el anillo 3 no se puede tocar un puerto, cerrar las
interrupciones, parar la máquina ni cambiar de tablas de páginas. No es que el
compilador no deje: esas instrucciones existen y compilan, y es el procesador
quien las para. Tiene que ser así, porque un programa no tiene por qué haber
salido de este compilador.

### Compilar un programa

```
sumador es incluye_programa de "programas/sumador.fal"
```

Compila el archivo aparte, para que viva en la dirección `0x30000000`, y mete
el resultado en la imagen. Devuelve una región de palabras de 8 bytes con el
programa entero. La ruta se cuenta desde la carpeta del sistema, tiene que ir
escrita entre comillas porque se resuelve al compilar, y un fallo dentro del
programa se avisa con su archivo y su línea.

Un programa lleva delante cuatro palabras:

| Palabra | Es |
|---|---|
| 1 | la firma, `20071938299543878` ("FALPROG") |
| 2 | dónde empieza |
| 3 | cuántos bytes ocupa el archivo: código y textos, de solo lectura |
| 4 | cuánta memoria necesita en total: el resto es su pila y sus variables, a cero |

Cargarlo es trabajo del sistema: pedir páginas, copiar el archivo y montarle
unas tablas donde solo lo suyo lleve el bit de usuario (el 4).

### Entrar y salir

Cada vez que entra una interrupción, el compilador guarda los quince registros
y deja la dirección en `contexto`: 20 palabras, de abajo arriba r15 a r8, rdi,
rsi, rbp, rbx, rdx, rcx, rax, y luego rip, cs, rflags, rsp y ss. Se leen y se
cambian con una región:

```
marco es region con contexto y 20 y 8
```

Al salir, el procesador sigue por lo que diga `contexto`, sea o no lo que
entró. Con eso, cambiar de tarea es una línea, y empezar un programa nuevo es
escribir un contexto a mano y ponerlo:

| Palabra | Valor |
|---|---|
| 16 (rip) | la palabra 2 del programa |
| 17 (cs) | 27: el segmento de código, en el anillo 3 |
| 18 (rflags) | 514: con las interrupciones abiertas |
| 20 (ss) | 35: el segmento de datos, en el anillo 3 |

Cada programa necesita una página propia para entrar: con `pila_de_entrada`
se le dice al procesador dónde acaba, y el contexto de un programa recién
hecho va 160 bytes por debajo. La función que atiende la interrupción no corre
ahí sino en una pila aparte, una sola para todo el sistema, como en seL4: las
interrupciones están cerradas mientras dura y nada la pisa.

### Llamadas al sistema

Un programa pide cosas con `llama_al_sistema con n y a y b`, que da rax, rdi y
rsi en las palabras 15, 9 y 10 del contexto y devuelve lo que el sistema deje
en la 15. Es un `int 0x80`, la única puerta que el anillo 3 puede cruzar. Hay
dos llamadas fijas, porque las usa el propio lenguaje:

- la 0 es salir con el código `a`: es lo que hace un programa al acabar o al
  reventar;
- la 1 manda el byte `a`: es lo que hace `escribe`.

Va byte a byte a propósito: así el sistema no tiene que leer memoria de un
programa, y una dirección que viene de fuera es lo primero que se usa para
engañar a un núcleo. El resto de las llamadas las decide el sistema.

Todo lo que llega en una llamada es hostil. Si el sistema suma un número que le
pasa un programa sin mirarlo, el programa puede hacerle desbordar la cuenta y
pararlo entero: mira cada dato antes de usarlo.

### Cuando un programa rompe algo

```
funcion al_fallar con numero y codigo y direccion
    ...
fin

atiende_fallos con al_fallar
```

Lo que rompa un programa llega ahí en vez de a la pantalla roja: el número de
la excepción, su código de error y, si fue un fallo de página, la dirección. Se
cierra el programa y el resto sigue. Lo que rompa el núcleo sigue siendo un
aviso en rojo: eso sí que no tiene arreglo.

## La seguridad no es un añadido

Cada valor lleva su tipo pegado y todas las operaciones lo miran antes de
hacer nada. Por eso un número no se puede usar como dirección: para tocar
memoria hay que crear una región, y cada acceso comprueba que no se sale de
ella. Un byte de una región de bytes no acepta un 300: en vez de recortarlo,
el programa se para y lo cuenta.

Con eso desaparece entera la clase de fallo que más agujeros ha dado en los
núcleos escritos en C: leer o escribir fuera de donde se debía.

Cuando algo revienta, el núcleo hace lo mismo que el intérprete: dice qué
pasó, en qué línea y con qué texto, y da una pista. Lo manda por el puerto
serie y lo pinta en la pantalla en blanco sobre rojo. Lo mismo si el que se
para es el procesador: un fallo de página, por ejemplo, dice qué instrucción
iba a ejecutar y qué dirección quiso tocar.

## Lo que no hay, y por qué

- **Listas, diccionarios, decimales y unir textos.** Todos necesitan pedir
  memoria mientras el programa corre, y en un núcleo eso no existe hasta que
  el propio sistema lo escribe. El núcleo dice cuánta hay y dónde; el
  repartidor es del sistema.
- **Teclado, temporizador, procesos.** Son trabajo del sistema operativo,
  no del lenguaje. El núcleo da las interrupciones y el cambio de contexto;
  a quién le toca lo decide el sistema.
- **`syscall` y `sysret`.** Las llamadas van por `int 0x80`, que es más lento y
  mucho más difícil de hacer mal: `sysret` tiene un agujero conocido si la
  dirección de vuelta no es válida.
- **Más de un giga de memoria.** El arranque deja mapeado el primer giga, con
  la dirección virtual igual a la física. Una región que se salga de ahí se
  rechaza, porque tocarla haría que el procesador se reiniciase sin decir
  nada. Ampliarlo es cambiar las tablas de páginas, que también es trabajo
  del sistema.
- **Funciones guardadas en variables, tipos y `intenta`.**
- **Varios archivos.** Un núcleo es un solo `.fal`, más los programas que
  incluya con `incluye_programa`, que van aparte.

Si usas algo de esto, `fal --nucleo` lo dice al compilar, con la línea.

## Cómo funciona

- La imagen lleva una cabecera **multiboot**, así que la arrancan QEMU (con
  `-kernel`) y GRUB. Es un binario plano y no un ELF porque QEMU no carga los
  ELF de 64 bits por multiboot.
- El arranque pasa del modo de 32 bits al de 64, mapea el primer giga con
  páginas de 2 megas y llama al programa. Al terminar, el procesador se para.
- Delante de tu programa va un **preludio** escrito en el mismo Fal
  ([nucleo/preludio.fal](nucleo/preludio.fal)): saca `escribe` por el puerto
  serie y pinta los avisos. Lo que hace él lo puede hacer tu sistema.
- El código máquina se genera de la forma más sencilla que hay, con una pila
  para los cálculos a medias. Es mucho más lento que el de gcc, y a cambio
  cada trozo se puede leer y comprobar por separado.
- La misma fuente da siempre la misma imagen, byte a byte.

Las pruebas (`go test -run Nucleo`) arrancan cada programa de verdad en QEMU
y comparan lo que sale por el puerto serie. Si no hay QEMU, se saltan.

## Puertos para un programa

El mapa de puertos del procesador va detrás del TSS y nace lleno de unos: un bit
a 1 es un puerto cerrado. `mapa_de_puertos` da su dirección, y el sistema abre
los que quiera poniendo a 0 el bit de cada uno:

```fal
mapa es region con mapa_de_puertos y 8192
elemento 13 de mapa es 254      # abre el puerto 96 y deja cerrado el 100
```

El mapa es uno solo, así que el sistema lo cambia al cambiar de tarea. Abrir un
puerto no lo abre para siempre: lo abre para quien corre en ese momento.

`fal --arranca` conecta el `disco.img` que haya al lado del fuente.

## Modo gráfico

El núcleo no sabe de gráficos: eso es del sistema, que habla con la tarjeta por
sus puertos. La que trae QEMU se enciende escribiendo el ancho, el alto y los
bits por los puertos 462 y 463, y la dirección de su imagen se lee del bus PCI
con `pon_doble` y `lee_doble`.

Si algo revienta con los gráficos encendidos, el preludio los apaga y vuelve al
modo de texto para pintar el aviso rojo. Para que eso funcione, el sistema no
puede pisar los primeros 256 KB de la memoria de la tarjeta, que es donde vive la
letra del modo de texto: FalOS desplaza la imagen 70 filas.
