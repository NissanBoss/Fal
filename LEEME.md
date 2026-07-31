# Fal

El lenguaje de programación más fácil del mundo, ahora **independiente**.

Un solo archivo ejecutable. Sin Python. Sin máquina virtual. Sin instalador.
Lo copias y funciona.

```
tipo Cuenta con titular y saldo
    funcion ingresa con cantidad
        saldo de mi es saldo de mi mas cantidad
    fin
fin

cuenta es nuevo Cuenta con "Pedro" y 100
ingresa de cuenta con 50
escribe saldo de cuenta        # 150
```

---

## Instalación

Descarga el paquete de tu sistema, **descomprímelo entero** y ejecuta el instalador
que viene dentro.

| Tu sistema | El paquete | El instalador |
|---|---|---|
| Windows | `fal-windows.zip` | doble clic en `instalar.bat` |
| Mac con chip Apple (M1, M2, M3...) | `fal-mac-apple.tar.gz` | `sh instalar.sh` |
| Mac con Intel | `fal-mac-intel.tar.gz` | `sh instalar.sh` |
| Linux | `fal-linux.tar.gz` | `sh instalar.sh` |
| Linux en ARM (Raspberry Pi...) | `fal-linux-arm.tar.gz` | `sh instalar.sh` |

El instalador **no pide permisos de administrador**. Deja todo en tu carpeta de usuario y:

1. Copia el programa (`%LOCALAPPDATA%\Fal` en Windows, `~/.local/bin` en Mac y Linux).
2. Añade Fal al PATH, para poder escribir `fal` desde cualquier sitio.
3. Asocia los archivos `.fal` (solo en Windows), para poder abrirlos con doble clic.
4. Instala el coloreado de VS Code, si tienes VS Code.
5. Copia los ejemplos.

Cuando acabe, **abre una terminal nueva** (la que ya tenías no ve el PATH nuevo) y prueba:

```bash
fal --ayuda
```

### Para desinstalar

- **Windows**: `desinstalar.bat`, que viene en el mismo paquete.
- **Mac y Linux**: `sh instalar.sh --quitar`.

### Si tu sistema bloquea el programa

Le pasa a todo programa nuevo que no venga firmado por una empresa con certificado.
No es un virus: es que Fal es tuyo y aún no tiene firma.

- **Windows**: botón derecho sobre `fal.exe` → **Propiedades** → marca **Desbloquear**.
  Si además tienes activado *Smart App Control*, hay que desactivarlo en
  *Seguridad de Windows → Control de aplicaciones y navegador*.
- **Mac**: *Ajustes → Privacidad y seguridad*, y pulsa **Abrir igualmente**.

### Sin instalador

Si prefieres no instalar nada, el programa funciona suelto. Sácalo del paquete,
dale permiso de ejecución en Mac y Linux (`chmod +x fal`) y ejecútalo donde esté.

## Cómo se usa

```bash
fal programa.fal      ejecuta un programa
fal                     abre la consola interactiva
fal --probar [carpeta]  ejecuta el banco de pruebas
fal --editor [carpeta]  genera el coloreado para VS Code
fal --ayuda             recuerda todo esto
```

Pruébalo con el recorrido completo por el lenguaje:

```bash
fal ejemplos/avanzado.fal
```

---

## Color en el editor

```bash
fal --editor
```

Genera la carpeta `vscode-fal`. Cópiala dentro de:

- **Windows**: `%USERPROFILE%\.vscode\extensions\fal`
- **Mac y Linux**: `~/.vscode/extensions/fal`

Reinicia VS Code y los `.fal` salen coloreados, con sangría automática
después de `si` o `funcion` y quitándola al escribir `fin`.

El coloreado se genera leyendo las tablas del propio lenguaje, así que si
añades una función nueva y vuelves a lanzarlo, sale pintada sola.

---

## Qué hay en cada archivo

| Archivo | Qué es |
|---|---|
| `dist/` | Los ejecutables, uno por sistema |
| `LEEME.md` | Esto |
| `MANUAL.md` | El manual completo del lenguaje |
| `ejemplos/` | 6 programas, de `hola.fal` a `avanzado.fal` |
| `pruebas/` | 13 pruebas con su salida exacta esperada |
| `construir.sh` | Vuelve a compilar para todos los sistemas |

Y el código fuente del intérprete, en Go:

| Archivo | Parte |
|---|---|
| `lector.go` | Texto → piezas |
| `armador.go` `armador_ins.go` `armador_exp.go` | Piezas → árbol |
| `interprete.go` `ejecutar.go` `evaluar.go` `operaciones.go` | Árbol → resultado |
| `numero.go` | Los números exactos |
| `valor.go` | Los tipos de dato y los errores |
| `integradas.go` `integradas2.go` | Las 80 funciones |
| `main.go` `probar.go` `editor.go` | Consola, pruebas y editor |

---

## Volver a compilar

Solo hace falta [Go](https://go.dev) instalado.

```bash
go build -o fal .        # solo para tu sistema
sh construir.sh            # para los cinco sistemas
go test ./...              # las pruebas del interprete
fal --probar             # las pruebas del lenguaje
```

**Para añadir una función al lenguaje** basta con una entrada en la tabla de
`integradas.go`. No hay que tocar ni el lector ni el armador:

```go
integrada("dobla", 1, 1, func(in *Interprete, a []Valor, ln int) (Valor, *ErrorFal) {
    n, err := pideNum(a[0], ln, "dobla")
    if err != nil {
        return nil, err
    }
    return multiplicaNum(n, Entero(2)), nil
})
```

Y ya puedes escribir `escribe dobla de 21`.

---

## Qué cambió respecto a la versión en Python

| | Antes (Python) | Ahora (Go) |
|---|---|---|
| Hace falta instalar | Python 3.8+ | **nada** |
| Repartir un programa | "instálate Python primero" | copias un archivo |
| Arranque | 149 ms | **15 ms** |
| `fib(24)` | 889 ms | **50 ms** (18x más rápido) |
| Comparado con Python nativo | 55x más lento | **8,6x** más lento |
| Sistemas | donde haya Python | Windows, Mac Intel, Mac Apple, Linux, Linux ARM |
| Banco de pruebas | `probar.py` | dentro del propio ejecutable |
| Generar el coloreado | `generar.py` | dentro del propio ejecutable |

**El lenguaje no cambió en nada.** Las 13 pruebas dan exactamente la misma
salida, byte a byte, que daban con el intérprete de Python.

Los números exactos, que eran lo más delicado de portar, siguen igual de bien:

```
escribe 0.1 mas 0.2              → 0.3         (Python dice 0.30000000000000004)
escribe (0.1 mas 0.2) es 0.3     → verdadero   (Python dice False)
escribe 19.99 por 3              → 59.97       sirve para dinero
```

Por dentro ahora son fracciones exactas (`math/big.Rat`) con un atajo rápido
para los enteros pequeños, que es lo que hace que siga yendo deprisa.
