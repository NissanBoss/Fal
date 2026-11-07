package main

// Un ensamblador de x86-64 que solo sabe las instrucciones que usa el
// compilador del nucleo. Se codifican a mano en vez de llamar a nasm o a gcc
// para que "fal --nucleo" siga siendo un unico programa sin nada instalado al
// lado: quien lo siga en los videos no tiene que montarse una cadena cruzada.
//
// Solo se usan los ocho registros de siempre. Con ellos no hace falta el
// prefijo REX.R ni REX.B, y la codificacion queda en una tabla que se puede
// comprobar a ojo con objdump. Los otros ocho solo se guardan y se devuelven
// al entrar y salir de una interrupcion (ver guardaAltos).

import "encoding/binary"

const (
	rax = iota
	rcx
	rdx
	rbx
	rsp
	rbp
	rsi
	rdi
)

// Condiciones de los saltos, tal como van en el segundo byte de 0F 8x.
const (
	siDesborda    = 0x0
	siDebajo      = 0x2 // sin signo
	siIgual       = 0x4
	siDistinto    = 0x5
	siMenor       = 0xC
	siMayorOIgual = 0xD
	siMenorOIgual = 0xE
	siMayor       = 0xF
)

type etiqueta int

type salto struct {
	pos     int
	destino etiqueta
}

// Una direccion que no se sabe hasta montar la imagen entera: los textos y
// las variables globales van detras del codigo, y el codigo no sabe cuanto
// ocupa hasta que se acaba de compilar.
type referencia struct {
	pos, ancho int
	simbolo    string
	mas        int64
}

type ensamblador struct {
	base     int64 // donde va a estar cargado el primer byte
	b        []byte
	sitios   []int
	saltos   []salto
	refs     []referencia
	simbolos map[string]int64
}

func nuevoEnsamblador(base int64) *ensamblador {
	return &ensamblador{base: base, simbolos: map[string]int64{}}
}

func (e *ensamblador) nueva() etiqueta {
	e.sitios = append(e.sitios, -1)
	return etiqueta(len(e.sitios) - 1)
}

func (e *ensamblador) aqui(l etiqueta) { e.sitios[l] = len(e.b) }

// marca apunta la direccion final de lo que venga a continuacion.
func (e *ensamblador) marca(simbolo string) { e.simbolos[simbolo] = e.base + int64(len(e.b)) }

func (e *ensamblador) bytes(b ...byte) { e.b = append(e.b, b...) }
func (e *ensamblador) d32(v uint32)    { e.b = binary.LittleEndian.AppendUint32(e.b, v) }
func (e *ensamblador) d64(v uint64)    { e.b = binary.LittleEndian.AppendUint64(e.b, v) }

func (e *ensamblador) alinea(n int) {
	for len(e.b)%n != 0 {
		e.b = append(e.b, 0)
	}
}

func (e *ensamblador) ref(ancho int, simbolo string, mas int64) {
	e.refs = append(e.refs, referencia{len(e.b), ancho, simbolo, mas})
	if ancho == 4 {
		e.d32(0)
	} else {
		e.d64(0)
	}
}

func (e *ensamblador) rel32(l etiqueta) {
	e.saltos = append(e.saltos, salto{len(e.b), l})
	e.d32(0)
}

func (e *ensamblador) jmp(l etiqueta)          { e.bytes(0xE9); e.rel32(l) }
func (e *ensamblador) call(l etiqueta)         { e.bytes(0xE8); e.rel32(l) }
func (e *ensamblador) jcc(cc byte, l etiqueta) { e.bytes(0x0F, 0x80|cc); e.rel32(l) }
func (e *ensamblador) ret()                    { e.bytes(0xC3) }
func (e *ensamblador) push(r int)              { e.bytes(0x50 | byte(r)) }
func (e *ensamblador) pop(r int)               { e.bytes(0x58 | byte(r)) }

// movImm no toca las banderas, y el codigo de las comparaciones cuenta con
// ello: pone el resultado a 0 entre el cmp y el salto.
func (e *ensamblador) movImm(r int, v int64) {
	if v >= -1<<31 && v < 1<<31 {
		e.bytes(0x48, 0xC7, 0xC0|byte(r))
		e.d32(uint32(int32(v)))
		return
	}
	e.bytes(0x48, 0xB8|byte(r))
	e.d64(uint64(v))
}

func (e *ensamblador) movSimbolo(r int, simbolo string, mas int64) {
	e.bytes(0x48, 0xB8|byte(r))
	e.ref(8, simbolo, mas)
}

// leeAbs y guardaAbs van a una direccion fija sin tocar ningun otro registro:
// es lo unico que se puede hacer al entrar en una interrupcion, cuando aun no
// hay ninguno libre. La direccion va en 32 bits, y toda la imagen cabe.
func (e *ensamblador) leeAbs(r int, simbolo string) {
	e.bytes(0x48, 0x8B, 0x04|byte(r)<<3, 0x25)
	e.ref(4, simbolo, 0)
}

func (e *ensamblador) guardaAbs(simbolo string, r int) {
	e.bytes(0x48, 0x89, 0x04|byte(r)<<3, 0x25)
	e.ref(4, simbolo, 0)
}

// guardaAltos mete r8 a r15 en la pila, y devuelveAltos los saca. El codigo
// compilado no los usa, pero un programa podria dejar algo ahi y leer lo que
// dejo otro: lo que no se guarda al cambiar de tarea, se filtra.
func (e *ensamblador) guardaAltos() {
	for i := byte(0); i < 8; i++ {
		e.bytes(0x41, 0x50|i)
	}
}

func (e *ensamblador) devuelveAltos() {
	for i := 7; i >= 0; i-- {
		e.bytes(0x41, 0x58|byte(i))
	}
}

func (e *ensamblador) mov(dst, src int) { e.bytes(0x48, 0x89, 0xC0|byte(src)<<3|byte(dst)) }

// lee y guarda usan siempre desplazamiento de 32 bits: asi rbp no necesita
// el caso especial de mod=00. La base nunca es rsp, que pediria un byte SIB.
func (e *ensamblador) lee(dst, base int, desp int32) {
	e.bytes(0x48, 0x8B, 0x80|byte(dst)<<3|byte(base))
	e.d32(uint32(desp))
}

func (e *ensamblador) guarda(base int, desp int32, src int) {
	e.bytes(0x48, 0x89, 0x80|byte(src)<<3|byte(base))
	e.d32(uint32(desp))
}

func (e *ensamblador) add(dst, src int)  { e.bytes(0x48, 0x01, 0xC0|byte(src)<<3|byte(dst)) }
func (e *ensamblador) sub(dst, src int)  { e.bytes(0x48, 0x29, 0xC0|byte(src)<<3|byte(dst)) }
func (e *ensamblador) cmp(dst, src int)  { e.bytes(0x48, 0x39, 0xC0|byte(src)<<3|byte(dst)) }
func (e *ensamblador) imul(dst, src int) { e.bytes(0x48, 0x0F, 0xAF, 0xC0|byte(dst)<<3|byte(src)) }
func (e *ensamblador) cqo()              { e.bytes(0x48, 0x99) }
func (e *ensamblador) idiv(r int)        { e.bytes(0x48, 0xF7, 0xF8|byte(r)) }
func (e *ensamblador) neg(r int)         { e.bytes(0x48, 0xF7, 0xD8|byte(r)) }

func (e *ensamblador) shl(r int, n byte) { e.bytes(0x48, 0xC1, 0xE0|byte(r), n) }
func (e *ensamblador) shr(r int, n byte) { e.bytes(0x48, 0xC1, 0xE8|byte(r), n) }

func (e *ensamblador) opImm(extension byte, r int, v int32) {
	e.bytes(0x48, 0x81, 0xC0|extension<<3|byte(r))
	e.d32(uint32(v))
}

func (e *ensamblador) addImm(r int, v int32) { e.opImm(0, r, v) }
func (e *ensamblador) andImm(r int, v int32) { e.opImm(4, r, v) }
func (e *ensamblador) subImm(r int, v int32) { e.opImm(5, r, v) }
func (e *ensamblador) xorImm(r int, v int32) { e.opImm(6, r, v) }
func (e *ensamblador) cmpImm(r int, v int32) { e.opImm(7, r, v) }

// Leer y escribir en la direccion que hay en rsi, con el ancho del elemento
// de la region. Lo leido se extiende con ceros.
func (e *ensamblador) leeAncho(ancho int) {
	switch ancho {
	case 1:
		e.bytes(0x0F, 0xB6, 0x06)
	case 2:
		e.bytes(0x0F, 0xB7, 0x06)
	case 4:
		e.bytes(0x8B, 0x06)
	default:
		e.bytes(0x48, 0x8B, 0x06)
	}
}

func (e *ensamblador) guardaAncho(ancho int) {
	switch ancho {
	case 1:
		e.bytes(0x88, 0x06)
	case 2:
		e.bytes(0x66, 0x89, 0x06)
	case 4:
		e.bytes(0x89, 0x06)
	default:
		e.bytes(0x48, 0x89, 0x06)
	}
}

func (e *ensamblador) entraPuerto() { e.bytes(0xEC, 0x0F, 0xB6, 0xC0) } // in al, dx; movzx eax, al
func (e *ensamblador) salePuerto()  { e.bytes(0xEE) }                   // out dx, al

// entraPalabra lee 16 bits de golpe. Un disco ATA da sus datos asi y no se
// los puede pedir byte a byte.
func (e *ensamblador) entraPalabra() { e.bytes(0x66, 0xED, 0x0F, 0xB7, 0xC0) } // in ax, dx; movzx eax, ax

// Las tarjetas graficas y el bus PCI tampoco aceptan bytes sueltos.
func (e *ensamblador) salePalabra() { e.bytes(0x66, 0xEF) } // out dx, ax
func (e *ensamblador) saleDoble()   { e.bytes(0xEF) }       // out dx, eax
func (e *ensamblador) entraDoble()  { e.bytes(0xED) }       // in eax, dx

// resolver rellena los saltos y las direcciones. Devuelve el simbolo que
// falte, que solo puede ser un fallo del propio compilador.
func (e *ensamblador) resolver() string {
	for _, s := range e.saltos {
		destino := e.sitios[s.destino]
		if destino < 0 {
			return "una etiqueta sin colocar"
		}
		binary.LittleEndian.PutUint32(e.b[s.pos:], uint32(int32(destino-(s.pos+4))))
	}
	for _, r := range e.refs {
		v, hay := e.simbolos[r.simbolo]
		if !hay {
			return r.simbolo
		}
		v += r.mas
		if r.ancho == 4 {
			binary.LittleEndian.PutUint32(e.b[r.pos:], uint32(v))
		} else {
			binary.LittleEndian.PutUint64(e.b[r.pos:], uint64(v))
		}
	}
	return ""
}
