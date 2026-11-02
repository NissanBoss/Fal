package main

// Lo que hace falta para que un ordenador arranque el programa: la cabecera
// multiboot, el paso del modo de 32 bits al de 64 y la memoria que el
// programa encuentra ya preparada.
//
// La imagen es un binario plano y no un ELF porque el cargador multiboot de
// QEMU rechaza los ELF de 64 bits ("Cannot load x86-64 image, give a 32bit
// one"). Con los campos de direccion de la cabecera rellenos (el bit 16 de
// las banderas) lo carga tal cual, y GRUB tambien.

const (
	cargaEn        = 0x100000 // el primer mega: lo de debajo es de la BIOS y la pantalla
	tamanoPila     = 256 << 10
	margenPila     = 16 << 10 // lo que queda libre para poder contar el desborde
	memoriaVisible = 1 << 30  // el primer giga, que es lo que mapean las tablas de paginas
)

func emitirCabecera(e *ensamblador) {
	const magia, banderas = 0x1BADB002, 0x00010002
	e.d32(magia)
	e.d32(banderas)
	suma := uint32(magia + banderas)
	e.d32(-suma) // las tres palabras tienen que sumar 0
	e.ref(4, "cabecera", 0)
	e.ref(4, "cabecera", 0) // se carga desde la cabecera misma
	e.ref(4, "fin_archivo", 0)
	e.ref(4, "fin_bss", 0)
	e.ref(4, "entrada32", 0)
}

// emitirArranque deja la maquina en 64 bits con el primer giga mapeado tal
// cual (la direccion virtual es la fisica) y llama a principal. Al volver se
// queda parada. Las paginas son de 2 megas para no necesitar mas que tres
// tablas.
func emitirArranque(e *ensamblador, principal etiqueta, prueba bool) {
	e.marca("entrada32")
	e.bytes(0xFA) // cli
	e.bytes(0xBC) // mov esp, pila_tope
	e.ref(4, "pila_tope", 0)

	// La BIOS deja basura en la memoria. Hay que limpiarla aunque QEMU y
	// GRUB ya la entreguen a cero, porque las tablas y las globales cuentan
	// con empezar a cero y no todos los cargadores lo garantizan.
	e.bytes(0xBF)
	e.ref(4, "inicio_bss", 0) // mov edi, inicio_bss
	e.bytes(0xB9)
	e.ref(4, "palabras_bss", 0) // mov ecx, palabras
	e.bytes(0x31, 0xC0)         // xor eax, eax
	e.bytes(0xF3, 0xAB)         // rep stosd

	e.bytes(0xB8)
	e.ref(4, "pdpt", 3) // mov eax, pdpt | presente | escribible
	e.bytes(0xA3)
	e.ref(4, "pml4", 0) // mov [pml4], eax
	e.bytes(0xB8)
	e.ref(4, "pd", 3)
	e.bytes(0xA3)
	e.ref(4, "pdpt", 0)

	e.bytes(0xBF)
	e.ref(4, "pd", 0)                  // mov edi, pd
	e.bytes(0xB8, 0x83, 0, 0, 0)       // mov eax, presente | escribible | pagina grande
	e.bytes(0xB9, 0x00, 0x02, 0, 0)    // mov ecx, 512
	e.bytes(0x89, 0x07)                // bucle: mov [edi], eax
	e.bytes(0x05, 0x00, 0x00, 0x20, 0) // add eax, 2 megas
	e.bytes(0x83, 0xC7, 0x08)          // add edi, 8
	e.bytes(0xE2, 0xF4)                // loop bucle

	e.bytes(0xB8)
	e.ref(4, "pml4", 0)
	e.bytes(0x0F, 0x22, 0xD8)             // mov cr3, eax
	e.bytes(0x0F, 0x20, 0xE0)             // mov eax, cr4
	e.bytes(0x83, 0xC8, 0x20)             // or eax, PAE
	e.bytes(0x0F, 0x22, 0xE0)             // mov cr4, eax
	e.bytes(0xB9, 0x80, 0x00, 0x00, 0xC0) // mov ecx, EFER
	e.bytes(0x0F, 0x32)                   // rdmsr
	e.bytes(0x0D, 0x00, 0x01, 0x00, 0x00) // or eax, modo largo
	e.bytes(0x0F, 0x30)                   // wrmsr
	e.bytes(0x0F, 0x20, 0xC0)             // mov eax, cr0
	e.bytes(0x0D, 0x01, 0x00, 0x00, 0x80) // or eax, paginacion | modo protegido
	e.bytes(0x0F, 0x22, 0xC0)             // mov cr0, eax
	e.bytes(0x0F, 0x01, 0x15)
	e.ref(4, "gdtr", 0) // lgdt [gdtr]
	e.bytes(0xEA)
	e.ref(4, "entrada64", 0)
	e.bytes(0x08, 0x00) // jmp 0x08:entrada64

	e.marca("entrada64")
	e.bytes(0x66, 0xB8, 0x10, 0x00) // mov ax, 0x10
	for _, segmento := range []byte{0xD8, 0xC0, 0xD0, 0xE0, 0xE8} {
		e.bytes(0x8E, segmento) // ds, es, ss, fs, gs
	}
	e.movSimbolo(rsp, "pila_tope", 0)

	// Las 256 interrupciones a sus trampas antes de nada: desde aqui, un
	// fallo del procesador ya se cuenta en vez de reiniciar la maquina.
	e.movSimbolo(rsi, "idt", 0)
	e.movSimbolo(rax, "trampas", 0)
	e.movImm(rbx, 0)
	cada := e.nueva()
	e.aqui(cada)
	emitirPuerta(e)
	e.addImm(rax, 16)
	e.addImm(rsi, 16)
	e.addImm(rbx, 1)
	e.cmpImm(rbx, 256)
	e.jcc(siMenor, cada)
	e.movSimbolo(rax, "idtr", 0)
	e.bytes(0x0F, 0x01, 0x18) // lidt [rax]

	e.call(principal)
	emitirCongela(e, prueba, 0)
}

// emitirPuerta escribe en rsi la entrada de la tabla de interrupciones que
// salta a la direccion de rax. Es una puerta de interrupcion (0x8E: entra
// con las interrupciones cerradas) en el segmento de codigo 0x08. La
// direccion va partida en tres trozos, que es como la quiere el procesador.
func emitirPuerta(e *ensamblador) {
	e.mov(rcx, rax)
	e.andImm(rcx, 0xFFFF)
	e.mov(rdx, rax)
	e.shr(rdx, 16)
	e.andImm(rdx, 0xFFFF)
	e.shl(rdx, 48)
	e.add(rcx, rdx)
	e.movImm(rdx, 0x8E00<<32|0x08<<16)
	e.add(rcx, rdx)
	e.guarda(rsi, 0, rcx)
	e.mov(rdx, rax)
	e.shr(rdx, 32)
	e.guarda(rsi, 8, rdx)
}

// emitirCongela para el procesador para siempre. En las pruebas antes avisa
// a QEMU por el puerto de isa-debug-exit, que si no se quedaria esperando y
// cada prueba tardaria lo que tarde en saltar el reloj.
func emitirCongela(e *ensamblador, prueba bool, codigo int64) {
	if prueba {
		e.movImm(rdx, 0xF4)
		e.movImm(rax, codigo)
		e.salePuerto()
	}
	e.bytes(0xFA, 0xF4, 0xEB, 0xFD) // cli; hlt; jmp al hlt
}

func emitirGDT(e *ensamblador) {
	e.alinea(8)
	e.marca("gdt")
	e.d64(0)
	e.d64(0x00AF9A000000FFFF) // codigo de 64 bits
	e.d64(0x00CF92000000FFFF) // datos
	e.marca("gdtr")
	e.bytes(23, 0)
	e.ref(4, "gdt", 0)
	e.alinea(8)
	e.marca("idtr")
	e.bytes(0xFF, 0x0F) // 256 entradas de 16 bytes, menos uno
	e.ref(8, "idt", 0)
}

// colocarBSS reparte la memoria que va a cero detras del archivo. La pila va
// justo despues de las tablas de paginas y antes de las globales: si se
// desborda pisaria las tablas, y por eso cada funcion mira antes que quede
// sitio (ver pila_suelo).
func colocarBSS(e *ensamblador, globales int) {
	e.marca("fin_archivo")
	pos := (cargaEn + int64(len(e.b)) + 4095) &^ 4095
	poner := func(simbolo string, tamano int64) {
		e.simbolos[simbolo] = pos
		pos += tamano
	}
	poner("inicio_bss", 0)
	poner("pml4", 4096)
	poner("pdpt", 4096)
	poner("pd", 4096)
	poner("idt", 4096)
	poner("pila", tamanoPila)
	e.simbolos["pila_suelo"] = e.simbolos["pila"] + margenPila
	e.simbolos["pila_tope"] = pos
	poner("globales", int64(16*globales))
	e.simbolos["fin_bss"] = pos
	e.simbolos["palabras_bss"] = (pos - e.simbolos["inicio_bss"]) / 4
	e.simbolos["cabecera"] = cargaEn
}
