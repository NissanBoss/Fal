@echo off
setlocal
chcp 65001 >nul 2>&1
title Desinstalar Fal

set "DESTINO=%LOCALAPPDATA%\Fal"

echo.
echo   Desinstalar Fal
echo   ---------------
echo.
echo   Se borrara la carpeta:
echo     %DESTINO%
echo.
set /p RESPUESTA="   Seguro? (s/n): "
if /i not "%RESPUESTA%"=="s" (
    echo.
    echo   Cancelado. No se ha tocado nada.
    echo.
    pause
    exit /b 0
)

echo.
echo   [1/3] Quitando Fal del PATH...
powershell -NoProfile -ExecutionPolicy Bypass -Command ^
  "$d = $env:LOCALAPPDATA + '\Fal';" ^
  "$p = [Environment]::GetEnvironmentVariable('Path','User');" ^
  "if ($p) {" ^
  "  $limpio = ($p -split ';' | Where-Object { $_ -ne $d -and $_ -ne '' }) -join ';';" ^
  "  [Environment]::SetEnvironmentVariable('Path', $limpio, 'User')" ^
  "}"

echo   [2/3] Quitando la asociacion de los .fal...
reg delete "HKCU\Software\Classes\.fal" /f >nul 2>&1
reg delete "HKCU\Software\Classes\Fal.Programa" /f >nul 2>&1

echo   [3/3] Borrando los archivos...
if exist "%DESTINO%" rmdir /s /q "%DESTINO%"

rem El coloreado del editor se deja: puede que lo quieras conservar.
echo.
echo   Fal desinstalado.
echo.
echo   El coloreado de VS Code sigue instalado. Si tambien
echo   lo quieres fuera, borra esta carpeta:
echo     %USERPROFILE%\.vscode\extensions\fal
echo.
pause
