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
| `pon_en_puerto con p y v` | manda un byte a un puerto |
| `congela` | para el procesador para siempre |

Se cuenta desde el 1 y `menos 1` es el último, igual que en todo Fal.

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
serie y lo pinta en la pantalla en blanco sobre rojo.

## Lo que no hay, y por qué

- **Listas, diccionarios, decimales y unir textos.** Todos necesitan pedir
  memoria mientras el programa corre, y en un núcleo eso no existe hasta que
  el propio sistema lo escribe.
- **Interrupciones, teclado, temporizador, procesos.** Son trabajo del sistema
  operativo, no del lenguaje.
- **Más de un giga de memoria.** El arranque deja mapeado el primer giga, con
  la dirección virtual igual a la física. Una región que se salga de ahí se
  rechaza, porque tocarla haría que el procesador se reiniciase sin decir
  nada. Ampliarlo es cambiar las tablas de páginas, que también es trabajo
  del sistema.
- **Funciones guardadas en variables, tipos y `intenta`.**
- **Varios archivos.** Por ahora un núcleo es un solo `.fal`.

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
