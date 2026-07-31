#!/bin/sh
# Instalador de Fal para Mac y Linux.
#
#   sh instalar.sh
#
# No hace falta sudo: se instala en la carpeta del usuario.
# Para quitarlo:  sh instalar.sh --quitar

set -e
ORIGEN="$(cd "$(dirname "$0")" && pwd)"
DESTINO="$HOME/.local/bin"
DATOS="$HOME/.fal"

azul()  { printf '\033[36m%s\033[0m\n' "$1"; }
gris()  { printf '\033[90m%s\033[0m\n' "$1"; }
mal()   { printf '\033[31m%s\033[0m\n' "$1"; }

# ---------------------------------------------------------------------
# Desinstalar
# ---------------------------------------------------------------------
if [ "$1" = "--quitar" ] || [ "$1" = "--desinstalar" ]; then
    echo ""
    azul "Desinstalando Fal..."
    rm -f "$DESTINO/fal"
    rm -rf "$DATOS"
    echo "  Fal desinstalado."
    gris "  El coloreado de VS Code (~/.vscode/extensions/fal) se deja por si lo quieres."
    gris "  Si añadiste ~/.local/bin al PATH y ya no lo usas, quítalo a mano."
    echo ""
    exit 0
fi

# ---------------------------------------------------------------------
# Averiguar que binario toca
# ---------------------------------------------------------------------
SISTEMA="$(uname -s)"
ARQUITECTURA="$(uname -m)"

case "$SISTEMA" in
    Darwin)
        case "$ARQUITECTURA" in
            arm64|aarch64) BINARIO="fal-mac-apple" ;;
            *)             BINARIO="fal-mac-intel" ;;
        esac ;;
    Linux)
        case "$ARQUITECTURA" in
            aarch64|arm64) BINARIO="fal-linux-arm" ;;
            *)             BINARIO="fal-linux" ;;
        esac ;;
    *)
        mal "No reconozco el sistema \"$SISTEMA\"."
        mal "Copia a mano el binario que te toque y dale permiso con chmod +x."
        exit 1 ;;
esac

# El paquete puede traer el binario ya con el nombre corto.
if   [ -f "$ORIGEN/fal" ];       then FUENTE="$ORIGEN/fal"
elif [ -f "$ORIGEN/$BINARIO" ];  then FUENTE="$ORIGEN/$BINARIO"
else
    mal "No encuentro el programa ($BINARIO) al lado de este instalador."
    mal "Descomprime el paquete entero antes de ejecutarlo."
    exit 1
fi

echo ""
azul "================================================"
azul "   FAL - el lenguaje de programacion mas facil"
azul "================================================"
echo ""
echo "  Sistema:  $SISTEMA $ARQUITECTURA"
echo "  Programa: $DESTINO/fal"
echo "  Ejemplos: $DATOS/ejemplos"
echo ""
printf "  Continuar? (s/n): "
read RESPUESTA
case "$RESPUESTA" in
    s|S|si|SI|Si|y|Y) ;;
    *) echo ""; echo "  Cancelado. No se ha tocado nada."; echo ""; exit 0 ;;
esac

echo ""
echo "  [1/4] Copiando el programa..."
mkdir -p "$DESTINO"
cp "$FUENTE" "$DESTINO/fal"
chmod +x "$DESTINO/fal"

echo "  [2/4] Copiando los ejemplos..."
if [ -d "$ORIGEN/ejemplos" ]; then
    mkdir -p "$DATOS/ejemplos"
    cp "$ORIGEN"/ejemplos/*.fal "$DATOS/ejemplos/" 2>/dev/null || true
fi
for DOC in LEEME.md MANUAL.md; do
    [ -f "$ORIGEN/$DOC" ] && cp "$ORIGEN/$DOC" "$DATOS/$DOC"
done

echo "  [3/4] Comprobando el PATH..."
AVISO_PATH=""
case ":$PATH:" in
    *":$DESTINO:"*) gris "        ya estaba" ;;
    *)
        # Se añade al arranque del shell que use la persona.
        for PERFIL in "$HOME/.zshrc" "$HOME/.bashrc" "$HOME/.profile"; do
            if [ -f "$PERFIL" ]; then
                if ! grep -q '.local/bin' "$PERFIL" 2>/dev/null; then
                    printf '\n# Añadido por el instalador de Fal\nexport PATH="$HOME/.local/bin:$PATH"\n' >> "$PERFIL"
                    gris "        añadido a $(basename "$PERFIL")"
                fi
                AVISO_PATH="si"
                break
            fi
        done
        [ -z "$AVISO_PATH" ] && gris "        no encontre tu archivo de arranque; añade $DESTINO al PATH a mano"
        ;;
esac

echo "  [4/4] Coloreado para VS Code..."
if [ -d "$HOME/.vscode/extensions" ]; then
    if "$DESTINO/fal" --editor "$HOME/.vscode/extensions/fal" >/dev/null 2>&1; then
        gris "        instalado, reinicia VS Code"
    else
        gris "        no pude generarlo, hazlo luego con:  fal --editor"
    fi
else
    gris "        VS Code no esta instalado, me lo salto"
fi

echo ""
azul "================================================"
azul "   Listo. Fal esta instalado."
azul "================================================"
echo ""
if [ -n "$AVISO_PATH" ]; then
    echo "  Abre una terminal NUEVA para que reconozca el comando."
    echo ""
fi
echo "  Prueba:"
echo ""
echo "      fal --ayuda"
echo "      fal $DATOS/ejemplos/snake.fal"
echo ""
if [ "$SISTEMA" = "Darwin" ]; then
    gris "  Si macOS dice que no puede comprobar el programa, ve a"
    gris "  Ajustes > Privacidad y seguridad y pulsa \"Abrir igualmente\"."
    echo ""
fi
