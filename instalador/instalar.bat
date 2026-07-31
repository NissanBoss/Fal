@echo off
setlocal EnableDelayedExpansion
chcp 65001 >nul 2>&1
title Instalar Fal

rem ---------------------------------------------------------------------
rem  Instalador de Fal para Windows.
rem
rem  No hace falta ser administrador: todo se instala en la carpeta del
rem  usuario. Para quitarlo, ejecuta desinstalar.bat.
rem ---------------------------------------------------------------------

set "ORIGEN=%~dp0"
set "DESTINO=%LOCALAPPDATA%\Fal"

echo.
echo   ================================================
echo      FAL - el lenguaje de programacion mas facil
echo   ================================================
echo.

if not exist "%ORIGEN%fal.exe" (
    echo   [X] No encuentro fal.exe al lado de este instalador.
    echo       Descomprime el ZIP entero antes de ejecutarlo.
    echo.
    pause
    exit /b 1
)

echo   Se instalara en:
echo     %DESTINO%
echo.
set /p RESPUESTA="   Continuar? (s/n): "
if /i not "%RESPUESTA%"=="s" (
    echo.
    echo   Cancelado. No se ha tocado nada.
    echo.
    pause
    exit /b 0
)

echo.
echo   [1/4] Copiando el programa...
if not exist "%DESTINO%" mkdir "%DESTINO%" 2>nul
copy /y "%ORIGEN%fal.exe" "%DESTINO%\fal.exe" >nul
if errorlevel 1 (
    echo   [X] No pude copiar fal.exe. Cierra cualquier ventana de Fal abierta.
    echo.
    pause
    exit /b 1
)
if exist "%ORIGEN%ejemplos" (
    if not exist "%DESTINO%\ejemplos" mkdir "%DESTINO%\ejemplos" 2>nul
    copy /y "%ORIGEN%ejemplos\*.fal" "%DESTINO%\ejemplos\" >nul 2>&1
)
for %%A in (LEEME.md MANUAL.md) do (
    if exist "%ORIGEN%%%A" copy /y "%ORIGEN%%%A" "%DESTINO%\%%A" >nul 2>&1
)

echo   [2/4] Anadiendo Fal al PATH...
rem  Se usa PowerShell porque setx corta el PATH si pasa de 1024 caracteres.
powershell -NoProfile -ExecutionPolicy Bypass -Command ^
  "$d = $env:LOCALAPPDATA + '\Fal';" ^
  "$p = [Environment]::GetEnvironmentVariable('Path','User');" ^
  "if ($null -eq $p) { $p = '' };" ^
  "if (($p -split ';') -notcontains $d) {" ^
  "  $nuevo = if ($p.TrimEnd(';') -eq '') { $d } else { $p.TrimEnd(';') + ';' + $d };" ^
  "  [Environment]::SetEnvironmentVariable('Path', $nuevo, 'User');" ^
  "  Write-Host '        anadido' -ForegroundColor DarkGray" ^
  "} else { Write-Host '        ya estaba' -ForegroundColor DarkGray }"

echo   [3/4] Asociando los archivos .fal...
reg add "HKCU\Software\Classes\.fal" /ve /t REG_SZ /d "Fal.Programa" /f >nul 2>&1
reg add "HKCU\Software\Classes\Fal.Programa" /ve /t REG_SZ /d "Programa de Fal" /f >nul 2>&1
reg add "HKCU\Software\Classes\Fal.Programa\DefaultIcon" /ve /t REG_SZ /d "\"%DESTINO%\fal.exe\",0" /f >nul 2>&1
reg add "HKCU\Software\Classes\Fal.Programa\shell\open\command" /ve /t REG_SZ /d "\"%DESTINO%\fal.exe\" \"%%1\"" /f >nul 2>&1

echo   [4/4] Coloreado para VS Code...
if exist "%USERPROFILE%\.vscode\extensions" (
    "%DESTINO%\fal.exe" --editor "%USERPROFILE%\.vscode\extensions\fal" >nul 2>&1
    if errorlevel 1 (
        echo         no pude generarlo, hazlo luego con:  fal --editor
    ) else (
        echo         instalado, reinicia VS Code
    )
) else (
    echo         VS Code no esta instalado, me lo salto
)

echo.
echo   ================================================
echo      Listo. Fal esta instalado.
echo   ================================================
echo.
echo   IMPORTANTE: abre una ventana NUEVA de la terminal
echo   para que reconozca el comando (esta ya no vale).
echo.
echo   Luego prueba:
echo.
echo       fal --ayuda
echo       fal "%DESTINO%\ejemplos\snake.fal"
echo.
echo   Si Windows dice que bloqueo el programa, es porque
echo   no esta firmado. Boton derecho sobre fal.exe,
echo   Propiedades, marca Desbloquear, Aceptar.
echo.
pause
