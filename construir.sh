#!/bin/sh
# Construye Fal y prepara los paquetes que se reparten.
#
#   sh construir.sh
#
# Deja en dist/ un paquete por sistema. Cada paquete lleva el programa,
# su instalador, los ejemplos y la documentacion. Quien lo reciba solo
# tiene que descomprimir y ejecutar el instalador.

set -e
cd "$(dirname "$0")"
rm -rf dist
mkdir -p dist

# -s -w quita la informacion de depuracion: el archivo pesa bastante menos.
BANDERAS="-s -w"

echo "Comprobando antes de construir..."
if gofmt -l . | grep -q .; then
    echo "  hay archivos sin formatear. Ejecuta:  gofmt -w ."
    exit 1
fi
go vet ./...
go test ./... >/dev/null
echo "  todo en orden"
echo ""

# armar <carpeta> <sistema> <arquitectura> <nombre del binario>
#
# OJO: el automatismo de publicar (.github/workflows/publicar.yml) repite
# esto mismo por su cuenta. Si tocas aqui, toca alli tambien.
armar() {
    CARPETA="dist/$1"
    mkdir -p "$CARPETA/ejemplos"
    # -trimpath quita del binario la ruta desde la que se compilo. Si no,
    # cualquiera que abra el ejecutable ve la carpeta de quien lo hizo.
    GOOS="$2" GOARCH="$3" go build -trimpath -ldflags "$BANDERAS" -o "$CARPETA/$4" .
    # Todo lo que hay en ejemplos, no solo los .fal: gastos.fal necesita su
    # gastos.csv al lado y copiando solo *.fal se quedaba fuera, asi que el
    # ejemplo fallaba nada mas descargarlo.
    cp -r ejemplos/. "$CARPETA/ejemplos/"
    # Lo que dejan los ejemplos al ejecutarse aqui no tiene que viajar.
    rm -f "$CARPETA/ejemplos/"*.svg "$CARPETA/ejemplos/agenda.txt" \
          "$CARPETA/ejemplos/temporal.txt" "$CARPETA/ejemplos/notas.txt"
    # El banco de pruebas tambien va dentro: son 63 KB y hacen que
    # "fal --probar" funcione recien descargado, que es lo que promete el
    # README y documentan LEEME y MANUAL.
    cp -r pruebas "$CARPETA/"
    cp README.md LEEME.md MANUAL.md NOVEDADES.md LICENSE "$CARPETA/"
    echo "  $1"
}

echo "Construyendo..."
armar fal-windows   windows amd64 fal.exe
armar fal-mac-apple darwin  arm64 fal
armar fal-mac-intel darwin  amd64 fal
armar fal-linux     linux   amd64 fal
armar fal-linux-arm linux   arm64 fal

# Cada paquete se lleva el instalador que le toca.
cp instalador/instalar.bat instalador/desinstalar.bat dist/fal-windows/
for P in fal-mac-apple fal-mac-intel fal-linux fal-linux-arm; do
    cp instalador/instalar.sh "dist/$P/"
    chmod +x "dist/$P/instalar.sh" 2>/dev/null || true
done

# La extension de VS Code, solo si hay vsce a mano.
if command -v vsce >/dev/null 2>&1; then
    (cd editor/vscode-fal && vsce package --out ../../dist/fal-vscode.vsix >/dev/null 2>&1)
    echo "  fal-vscode.vsix"
fi

echo ""
echo "Comprimiendo..."
cd dist
for P in fal-windows fal-mac-apple fal-mac-intel fal-linux fal-linux-arm; do
    if [ "$P" = "fal-windows" ]; then
        # Windows abre los .zip con doble clic, sin instalar nada.
        if command -v powershell >/dev/null 2>&1; then
            powershell -NoProfile -Command \
                "Compress-Archive -Path '$P' -DestinationPath '$P.zip' -Force" >/dev/null
        else
            tar -a -c -f "$P.zip" "$P"
        fi
    else
        tar -czf "$P.tar.gz" "$P"
    fi
done
cd ..

echo ""
for F in dist/*.zip dist/*.tar.gz; do
    [ -f "$F" ] && printf "  %-26s %s\n" "$(basename "$F")" "$(du -h "$F" | cut -f1)"
done
echo ""
echo "Listo. Reparte a cada persona el paquete de su sistema:"
echo "  descomprimir y ejecutar el instalador que viene dentro."
