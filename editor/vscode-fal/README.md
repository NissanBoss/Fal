# Fal para VS Code

Colorea los archivos `.fal` del lenguaje de programación
[Fal](https://github.com/NissanBoss/Fal), y pone la sangría sola después de
`si`, `funcion` o `repite`, quitándola cuando escribes `fin`.

```fal
tipo Perro hereda de Animal con nombre
    funcion habla
        devuelve nombre de mi mas " dice guau"
    fin
fin
```

Fal es un lenguaje en español y sin símbolos: nada de `;`, `{}`, `==` ni `+`.
Está pensado para que aprender a programar no exija pelearse antes con el
inglés y con la sintaxis.

Se puede probar en el navegador, sin instalar nada:
**[nissanboss.github.io/Fal](https://nissanboss.github.io/Fal/)**

El coloreado se genera a partir de las tablas del propio intérprete, con
`fal --editor`, así que nunca se queda desfasado.
