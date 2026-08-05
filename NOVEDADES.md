# Novedades

Lo que cambia en cada versión. Lo de aquí arriba es lo más nuevo.

El automatismo de publicar coge de este archivo la sección de la versión que
se está etiquetando y la pone como texto de la release, así que lo que
escribas aquí es lo que lee quien vaya a descargarla.

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
