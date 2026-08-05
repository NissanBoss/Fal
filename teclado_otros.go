//go:build !windows && !linux && !darwin && !js

package main

// En un sistema que no conocemos no hay forma de leer teclas sueltas sin
// meter dependencias. El programa sigue funcionando: "tecla" contesta
// siempre que no hay ninguna pulsada.

func soltarTeclado() {}

func leerTecla() string { return "" }
