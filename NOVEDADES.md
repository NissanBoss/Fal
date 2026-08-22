# Novedades

Lo que cambia en cada versión. Lo de aquí arriba es lo más nuevo.

El automatismo de publicar coge de este archivo la sección de la versión que
se está etiquetando y la pone como texto de la release, así que lo que
escribas aquí es lo que lee quien vaya a descargarla.

## v8.1

Un arreglo de empaquetado. El lenguaje no cambia.

- La extensión de VS Code seguía colgándose con el número `1.0.0`, que es
  justo lo que v8 decía haber arreglado. Estaba a medias: el número se
  calculaba bien al generar la extensión con `fal --editor`, pero al
  construir se empaquetaba la carpeta guardada en el repositorio, donde ese
  número está escrito a mano y no se mueve. Ahora se regenera antes de
  empaquetarla, así que lleva la versión que se está publicando.

## v8

**El Snake se juega de verdad, y en el navegador.** Estaba escrito con
`pregunta`, que se para hasta que le das a Enter: había que teclear `dddw` y
mirar cómo pasaban cuatro movimientos de golpe. Ahora usa `tecla`, así que la
serpiente avanza sola y tú solo la tuerces, con `wasd` o con las flechas. Y
está en el playground, que es donde se puede jugar sin instalar nada.

Curioso el detalle: `tecla` se añadió en v7 justo para esto, y el juego se
quedó sin enterarse.

**Lo que escribes en el playground ya no se pierde.** Se guarda solo, según lo
escribes. Antes bastaba con recargar sin querer, darle a atrás o cerrar la
pestaña para quedarte sin nada, y también al ir a las lecciones y volver.

**Compartir con un enlace.** Hay un botón que te da una dirección con tu
programa dentro. Se la pasas a quien quieras, la abre, y ahí está. Sin cuentas,
sin servidor y sin nada que guardar en ninguna parte.

**`fal --version`.** No había forma de saber qué versión tenías instalada. De
paso arregla la extensión de VS Code, que salía en todas las versiones marcada
como `1.0.0`: el Marketplace no deja publicar dos veces el mismo número, así
que estaba de hecho bloqueada.

**Cuatro funciones nuevas:** `seno`, `coseno`, `tangente` y `logaritmo`. Los
ángulos van en grados, igual que `gira`, para que la trigonometría y la tortuga
hablen el mismo idioma. Ya se puede dibujar un círculo o una onda.

**Arreglado, en los números:**

- `potencia` con exponente negativo se iba al flotante, así que
  `potencia con 10 y (menos 2)` daba un 0.01 que luego no cuadraba al sumarlo.
  Ahora es exacto, que es lo que este lenguaje promete.
- `arriba` y `abajo` pasaban por el flotante: perdían precisión con números
  grandes y se salían del entero sin avisar.
- `escribe azar` soltaba veintiocho decimales
  (`0.9188921592527634629732347094`). Ahora da nueve.

**Arreglado, en el taller:** los `.go` no estaban fijados a saltos de línea LF,
así que clonando en Windows salían los 38 archivos marcados como sin formatear
y `construir.sh` se negaba a compilar. Y el automatismo de publicar repetía por
su cuenta lo que hace `construir.sh`; ahora lo llama, que de tener dos copias
separándose salió que los paquetes fueran meses sin el `gastos.csv`.

Van 93 funciones y siguen siendo 42 palabras.

## v7.1

Arreglos de empaquetado. El lenguaje no cambia.

- `ejemplos/gastos.fal` no funcionaba al descargarlo: el paquete llevaba el
  programa pero no su `gastos.csv`, así que fallaba nada más abrirlo.
- `fal --probar` tampoco: el banco de pruebas no viajaba dentro del paquete,
  aunque el manual lo documentara. Ahora van las 13 pruebas dentro.

## v7

**Dibujar con la tortuga.** Un lápiz que se arrastra por la pantalla y deja
raya. Cinco palabras nuevas: `camina`, `gira`, `levanta`, `apoya` y `color`.

```
repite 36 veces
    camina de 100
    gira de 170
fin
```

Cuatro líneas y sale una estrella de 36 puntas. Y esa es la gracia: es el
mismo bucle de siempre, pero se ve. En el navegador el dibujo sale debajo
del texto; desde la terminal se guarda en un `.svg` junto a tu programa.

**Juegos de verdad, con `tecla`.** Dice qué tecla acabas de pulsar, o texto
vacío si ninguna, y no espera a nadie. `pregunta` sigue siendo lo contrario,
que se para hasta el Enter: una sirve para pedir datos y la otra para jugar.
Está en `ejemplos/mueve.fal`.

**El playground va en vivo.** Lo que escribe el programa aparece según pasa,
en vez de todo de golpe al terminar. Eso es lo que permite que algo se mueva,
que `limpia` sirva para repintar en el sitio, y que se acabara el corte a los
cinco segundos. Hay un botón de Parar para los bucles que se van de las manos.

**Arreglado:** la función de textos que dice si algo acaba en un trozo no
había funcionado nunca. Compartía nombre con la de cortar el programa y se
quedaba sin registrar. Ahora se llama `acaba` y hace pareja con `empieza`.

Van 89 funciones y siguen siendo 42 palabras.
