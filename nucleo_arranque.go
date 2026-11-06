package main

// Lo que hace falta para que un ordenador arranque el programa: la cabecera
// multiboot, el paso del modo de 32 bits al de 64 y la memoria que el
// programa encuentra ya preparada.
//
// La imagen es un binario plano y no un ELF porque el cargador multiboot de
// QEMU rechaza los ELF de 64 bits ("Cannot load x86-64 image, give a 32bit
// one"). Con los campos de direccion de la cabecera rellenos (el bit 16 de
// las banderas) lo carga tal cual, y GRUB tambien.

import "encoding/binary"

const (
	cargaEn        = 0x100000 // el primer mega: lo de debajo es de la BIOS y la pantalla
	tamanoPila     = 256 << 10
	tamanoAtencion = 64 << 10 // donde corre la funcion que atiende una interrupcion
	margenPila     = 16 << 10 // lo que queda libre para poder contar el desborde
	memoriaVisible = 1 << 30  // el primer giga, que es lo que mapean las tablas de paginas

	// Un programa se compila para estar en 768 megas: dentro del primer giga
	// para que el sistema lo alcance con una region, y lejos de todo lo suyo.
	cargaPrograma  = 0x30000000
	pilaPrograma   = 64 << 10
	margenPrograma = 8 << 10
	magiaPrograma  = 0x00474F52504C4146 // "FALPROG"

	segmentoTSS = 0x28
	tamanoTSS   = 104
	// Un bit por puerto, mas el byte de unos que el procesador exige al final.
	tamanoMapaPuertos = 65536/8 + 1
	vectorLlamadas    = 0x80
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

	// El cargador deja en ebx donde estan sus datos (cuanta memoria hay y
	// donde), y en eax su firma. Sin firma, ebx es lo que hubiera y no se
	// puede leer como una direccion.
	e.bytes(0x3D, 0x02, 0xB0, 0xAD, 0x2B) // cmp eax, 0x2BADB002
	e.bytes(0x74, 0x02)                   // je +2
	e.bytes(0x31, 0xDB)                   // xor ebx, ebx

	// La BIOS deja basura en la memoria. Hay que limpiarla aunque QEMU y
	// GRUB ya la entreguen a cero, porque las tablas y las globales cuentan
	// con empezar a cero y no todos los cargadores lo garantizan.
	e.bytes(0xBF)
	e.ref(4, "inicio_bss", 0) // mov edi, inicio_bss
	e.bytes(0xB9)
	e.ref(4, "palabras_bss", 0) // mov ecx, palabras
	e.bytes(0x31, 0xC0)         // xor eax, eax
	e.bytes(0xF3, 0xAB)         // rep stosd
	e.bytes(0x89, 0x1D)
	e.ref(4, "datos_arranque", 0) // mov [datos_arranque], ebx

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
	e.bytes(0x0F, 0x22, 0xD8) // mov cr3, eax
	e.bytes(0x0F, 0x20, 0xE0) // mov eax, cr4
	e.bytes(0x83, 0xC8, 0x20) // or eax, PAE
	e.bytes(0x0F, 0x22, 0xE0) // mov cr4, eax
	// Se enciende el bit que prohibe ejecutar una pagina si el procesador lo
	// tiene: con el apagado, una tabla que lo use es una tabla invalida. Las
	// del arranque no lo usan; es para las que monte el sistema.
	e.bytes(0xB8, 0x01, 0x00, 0x00, 0x80)       // mov eax, 0x80000001
	e.bytes(0x0F, 0xA2)                         // cpuid
	e.bytes(0x89, 0xD6)                         // mov esi, edx
	e.bytes(0xB9, 0x80, 0x00, 0x00, 0xC0)       // mov ecx, EFER
	e.bytes(0x0F, 0x32)                         // rdmsr
	e.bytes(0x0D, 0x00, 0x01, 0x00, 0x00)       // or eax, modo largo
	e.bytes(0xF7, 0xC6, 0x00, 0x00, 0x10, 0x00) // test esi, NX
	e.bytes(0x74, 0x0C)                         // jz +12
	e.bytes(0x0D, 0x00, 0x08, 0x00, 0x00)       // or eax, NX
	e.bytes(0xC6, 0x05)
	e.ref(4, "hay_nx", 0)
	e.bytes(1)                // mov byte [hay_nx], 1
	e.bytes(0x0F, 0x30)       // wrmsr
	e.bytes(0x0F, 0x20, 0xC0) // mov eax, cr0
	// Paginacion y modo protegido, y que el nucleo tambien respete las
	// paginas de solo lectura (WP): sin ese bit, el procesador solo se las
	// hace respetar a los programas.
	e.bytes(0x0D, 0x01, 0x00, 0x01, 0x80) // or eax, 0x80010001
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
	e.movSimbolo(rax, "pila_suelo", 0)
	e.guardaAbs("suelo", rax)

	// Las 256 interrupciones a sus trampas antes de nada: desde aqui, un
	// fallo del procesador ya se cuenta en vez de reiniciar la maquina.
	e.movSimbolo(rsi, "idt", 0)
	e.movSimbolo(rax, "trampas", 0)
	e.movImm(rbx, 0)
	cada := e.nueva()
	e.aqui(cada)
	emitirPuerta(e, puertaNucleo)
	e.addImm(rax, 16)
	e.addImm(rsi, 16)
	e.addImm(rbx, 1)
	e.cmpImm(rbx, 256)
	e.jcc(siMenor, cada)
	e.movSimbolo(rax, "idtr", 0)
	e.bytes(0x0F, 0x01, 0x18) // lidt [rax]

	// Lo que el procesador necesita para volver del anillo 3: la pila a la
	// que salta al entrar. Sale de aqui hasta que el sistema diga otra, para
	// que olvidarse no acabe en un reinicio mudo.
	e.movSimbolo(rsi, "tss", 0)
	e.movSimbolo(rax, "pila_entrada_tope", 0)
	e.guarda(rsi, 4, rax)
	// Donde empieza el mapa de puertos permitidos. A 0, que es como queda sin
	// tocarlo, el mapa serian los primeros bytes del TSS, y sus ceros le
	// abririan a cualquier programa los puertos que coincidieran. El mapa va
	// detras del TSS y nace lleno de unos: un bit a 1 es un puerto cerrado,
	// asi que abrir uno es algo que el sistema tiene que hacer a proposito.
	e.movImm(rax, tamanoTSS)
	e.bytes(0x66, 0x89, 0x46, 102) // mov [rsi+102], ax
	e.movSimbolo(rdi, "tss", tamanoTSS)
	e.movImm(rcx, tamanoMapaPuertos)
	e.movImm(rax, 0xFF)
	e.bytes(0xF3, 0xAA) // rep stosb
	e.movImm(rax, segmentoTSS)
	e.bytes(0x0F, 0x00, 0xD8) // ltr ax

	e.call(principal)
	emitirCongela(e, prueba, 0)
}

// Las dos entran con las interrupciones cerradas. La de usuario es la unica
// que un programa puede cruzar con "int".
const (
	puertaNucleo  = 0x8E00
	puertaUsuario = 0xEE00
)

// emitirPuerta escribe en rsi la entrada de la tabla de interrupciones que
// salta a la direccion de rax, en el segmento de codigo 0x08. La direccion
// va partida en tres trozos, que es como la quiere el procesador.
func emitirPuerta(e *ensamblador, tipo int64) {
	e.mov(rcx, rax)
	e.andImm(rcx, 0xFFFF)
	e.mov(rdx, rax)
	e.shr(rdx, 16)
	e.andImm(rdx, 0xFFFF)
	e.shl(rdx, 48)
	e.add(rcx, rdx)
	e.movImm(rdx, tipo<<32|0x08<<16)
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
	// Con el bit de "usado" ya puesto: el procesador lo escribe la primera
	// vez que carga cada uno, y si el sistema deja este trozo de solo
	// lectura, esa escritura seria un fallo de pagina.
	e.d64(0x00AF9B000000FFFF) // codigo de 64 bits
	e.d64(0x00CF93000000FFFF) // datos
	e.d64(0x00AFFB000000FFFF) // los mismos dos, en el anillo 3
	e.d64(0x00CFF3000000FFFF)
	// El TSS ocupa dos entradas. Su direccion va partida a trozos y no se
	// sabe hasta colocar la memoria: la rellena ponerTSS.
	e.marca("gdt_tss")
	e.d64(0)
	e.d64(0)
	e.marca("gdtr")
	e.bytes(segmentoTSS+15, 0)
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
	poner("pila_atencion", tamanoAtencion)
	e.simbolos["suelo_atencion"] = e.simbolos["pila_atencion"] + margenPila
	e.simbolos["pila_atencion_tope"] = pos
	poner("pila_entrada", 4096)
	e.simbolos["pila_entrada_tope"] = pos
	poner("globales", int64(16*globales))
	poner("datos_arranque", 8)
	poner("hay_nx", 8)
	poner("suelo", 8)
	poner("contexto", 8)
	poner("manejador_fallos", 8)
	poner("fallo_numero", 8)
	poner("fallo_codigo", 8)
	poner("tss", tamanoTSS+tamanoMapaPuertos)
	e.simbolos["fin_bss"] = pos
	e.simbolos["fin_nucleo"] = (pos + 4095) &^ 4095
	e.simbolos["palabras_bss"] = (pos - e.simbolos["inicio_bss"]) / 4
	e.simbolos["cabecera"] = cargaEn
	ponerTSS(e)
}

func ponerTSS(e *ensamblador) {
	direccion := uint64(e.simbolos["tss"])
	sitio := e.simbolos["gdt_tss"] - e.base
	limite := uint64(tamanoTSS + tamanoMapaPuertos - 1)
	baja := limite&0xFFFF | (direccion&0xFFFFFF)<<16 | 0x89<<40 | (limite>>16&0xF)<<48 | (direccion>>24&0xFF)<<56
	binary.LittleEndian.PutUint64(e.b[sitio:], baja)
	binary.LittleEndian.PutUint64(e.b[sitio+8:], direccion>>32)
}

// Un programa suelto no arranca nada: lo carga el sistema, en el anillo 3.
// Delante lleva cuatro palabras para que el cargador sepa que es y cuanta
// memoria pedir.
func emitirCabeceraPrograma(e *ensamblador) {
	e.d64(magiaPrograma)
	e.ref(8, "entrada_programa", 0)
	e.ref(8, "tamano_archivo", 0)
	e.ref(8, "tamano_memoria", 0)
}

// emitirEntradaPrograma pone la pila del programa, lo llama y, si vuelve,
// sale por la llamada al sistema 0: en el anillo 3 nadie se puede parar solo.
func emitirEntradaPrograma(e *ensamblador, principal etiqueta) {
	e.marca("entrada_programa")
	e.movSimbolo(rsp, "pila_tope", 0)
	e.movSimbolo(rax, "pila_suelo", 0)
	e.guardaAbs("suelo", rax)
	e.call(principal)
	e.movImm(rax, 0)
	e.movImm(rdi, 0)
	e.bytes(0xCD, vectorLlamadas)
	e.bytes(0xEB, 0xFE) // aqui mismo, por si el sistema no lo cierra
}

// colocarBSSPrograma es colocarBSS para un programa: lo unico que tiene a
// cero son su pila y sus variables. Lo del nucleo aqui no existe, pero sus
// simbolos si, para que "datos_de_arranque" y compañia compilen.
func colocarBSSPrograma(e *ensamblador, globales int) {
	e.marca("fin_archivo")
	pos := (e.base + int64(len(e.b)) + 4095) &^ 4095
	poner := func(simbolo string, tamano int64) {
		e.simbolos[simbolo] = pos
		pos += tamano
	}
	poner("inicio_bss", 0)
	poner("pila", pilaPrograma)
	e.simbolos["pila_suelo"] = e.simbolos["pila"] + margenPrograma
	e.simbolos["pila_tope"] = pos
	poner("globales", int64(16*globales))
	poner("datos_arranque", 8)
	poner("hay_nx", 8)
	poner("suelo", 8)
	e.simbolos["fin_nucleo"] = (pos + 4095) &^ 4095
	e.simbolos["cabecera"] = e.base
	e.simbolos["tamano_archivo"] = e.simbolos["fin_archivo"] - e.base
	e.simbolos["tamano_memoria"] = e.simbolos["fin_nucleo"] - e.base
}
