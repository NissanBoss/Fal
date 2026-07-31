//go:build !windows

package main

// En Mac y en Linux las terminales entienden estas secuencias desde
// siempre, asi que no hay nada que preparar.

func prepararConsola() bool { return true }
