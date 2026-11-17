const sinMovimiento = matchMedia("(prefers-reduced-motion: reduce)").matches;

const NEGRO = 0, AZUL = 1, ROJO = 4, GRIS = 7, VERDE_CLARO = 10, CIAN_CLARO = 11, ROJO_CLARO = 12, AMARILLO = 14, BLANCO = 15;
const TINTAS = [
  [13, 17, 23], [31, 58, 138], [46, 139, 87], [32, 142, 158], [192, 57, 43], [142, 68, 173], [180, 120, 30], [196, 203, 212],
  [110, 118, 129], [88, 166, 255], [86, 211, 100], [57, 197, 207], [248, 113, 104], [210, 168, 255], [227, 179, 65], [240, 246, 252]
].map(([r, g, b]) => (255 << 24 | b << 16 | g << 8 | r) >>> 0);
const EN_437 = { "á": 160, "é": 130, "í": 161, "ó": 162, "ú": 163, "ñ": 164, "Ñ": 165, "ü": 129, "¿": 168, "¡": 173 };
const ALTO = 25;

function leerLetra(texto) {
  const chicas = new Map();
  let codigo = 0, filas = [];
  for (const linea of texto.split(/\r?\n/)) {
    if (/^\d/.test(linea)) {
      codigo = Number(linea);
      filas = [];
    } else if (linea[0] === "." || linea[0] === "#") {
      filas.push([...linea.slice(0, 6)].map(c => c === "#" ? 1 : 0));
      if (filas.length === 12) chicas.set(codigo, filas);
    }
  }
  const letras = new Map();
  for (const [codigo, chica] of chicas) letras.set(codigo, agrandar(chica));
  return letras;
}

function agrandar(chica) {
  const en = (x, y) => chica[Math.min(Math.max(y, 0), 11)][Math.min(Math.max(x, 0), 5)];
  const grande = new Uint8Array(288);
  for (let y = 0; y < 12; y++) {
    for (let x = 0; x < 6; x++) {
      const b = en(x, y - 1), d = en(x - 1, y), e = en(x, y), f = en(x + 1, y), h = en(x, y + 1);
      let a1 = e, a2 = e, a3 = e, a4 = e;
      if (b !== h && d !== f) {
        if (d === b) a1 = d;
        if (b === f) a2 = f;
        if (d === h) a3 = d;
        if (h === f) a4 = f;
      }
      const sitio = y * 24 + x * 2;
      grande[sitio] = a1;
      grande[sitio + 1] = a2;
      grande[sitio + 12] = a3;
      grande[sitio + 13] = a4;
    }
  }
  return grande;
}

function codigos(texto) {
  return [...texto].map(c => EN_437[c] || c.charCodeAt(0));
}

function partir(trozos, ancho) {
  const casillas = [];
  for (const [texto, tinta] of trozos) for (const codigo of codigos(texto)) casillas.push([codigo, tinta]);
  const renglones = [];
  let resto = casillas;
  while (resto.length > ancho) {
    let corte = ancho;
    while (corte > 2 && resto[corte][0] !== 32) corte--;
    if (corte <= 2) corte = ancho;
    renglones.push(resto.slice(0, corte));
    resto = [[32, GRIS], [32, GRIS], ...resto.slice(corte).filter((c, i) => i > 0 || c[0] !== 32)];
  }
  renglones.push(resto);
  return renglones;
}

function crearPantalla(lienzo, letras) {
  const ctx = lienzo.getContext("2d");
  let ancho = 0, imagen = null, pixeles = null;
  const estado = { consola: [], ventana: null, aviso: null, cursor: null };

  function medir(columnas) {
    if (columnas === ancho) return;
    ancho = columnas;
    lienzo.width = ancho * 12;
    lienzo.height = ALTO * 24;
    imagen = ctx.createImageData(lienzo.width, lienzo.height);
    pixeles = new Uint32Array(imagen.data.buffer);
  }

  function casilla(rejilla, fila, col, codigo, tinta, fondo) {
    if (fila < 0 || fila >= ALTO || col < 0 || col >= ancho) return;
    rejilla[fila * ancho + col] = [codigo, tinta, fondo];
  }

  function frase(rejilla, fila, col, texto, tinta, fondo) {
    codigos(texto).forEach((codigo, i) => casilla(rejilla, fila, col + i, codigo, tinta, fondo));
  }

  function barra(rejilla) {
    for (let c = 0; c < ancho; c++) casilla(rejilla, 0, c, 32, BLANCO, AZUL);
    frase(rejilla, 0, 2, "FalOS", BLANCO, AZUL);
    const ahora = new Date();
    const dos = n => String(n).padStart(2, "0");
    const hora = `${dos(ahora.getUTCHours())}:${dos(ahora.getUTCMinutes())} UTC`;
    const ancha = ancho >= 80;
    frase(rejilla, 0, ancha ? 60 : ancho - 11, hora, BLANCO, AZUL);
    if (ancha) frase(rejilla, 0, 73, "día 16", BLANCO, AZUL);
    if (estado.aviso) {
      const col = ancha ? 23 : 8;
      casilla(rejilla, 0, col, 32, NEGRO, estado.aviso.tinta);
      frase(rejilla, 0, col + 1, ` ${estado.aviso.texto} `, NEGRO, AMARILLO);
    }
  }

  function ventana(rejilla, v) {
    const va = v.columnas + 2;
    for (let c = 0; c < va; c++) casilla(rejilla, v.y, v.x + c, 32, NEGRO, v.tinta);
    casilla(rejilla, v.y, v.x + 1, 48 + v.numero, NEGRO, v.tinta);
    frase(rejilla, v.y, v.x + 3, v.nombre, NEGRO, v.tinta);
    frase(rejilla, v.y, v.x + va - 4, " x ", BLANCO, ROJO);
    const vistas = [...v.lineas, ""].slice(-v.renglones);
    for (let r = 0; r < v.renglones; r++) {
      const fila = v.y + 1 + r;
      casilla(rejilla, fila, v.x, 179, v.tinta, NEGRO);
      for (let c = 0; c < v.columnas; c++) casilla(rejilla, fila, v.x + 1 + c, 32, BLANCO, NEGRO);
      frase(rejilla, fila, v.x + 1, (vistas[r] || "").slice(0, v.columnas), BLANCO, NEGRO);
      casilla(rejilla, fila, v.x + va - 1, 179, v.tinta, NEGRO);
    }
    const pie = v.y + v.renglones + 1;
    casilla(rejilla, pie, v.x, 192, v.tinta, NEGRO);
    for (let c = 1; c < va - 1; c++) casilla(rejilla, pie, v.x + c, 196, v.tinta, NEGRO);
    casilla(rejilla, pie, v.x + va - 1, 217, v.tinta, NEGRO);
  }

  function componer() {
    const rejilla = new Array(ancho * ALTO).fill([32, GRIS, NEGRO]);
    const renglones = estado.consola.flatMap(trozos => partir(trozos, ancho));
    if (estado.cursor === "nuevo") renglones.push([[62, CIAN_CLARO], [32, GRIS]]);
    const vistos = renglones.slice(-(ALTO - 2));
    vistos.forEach((renglon, r) => renglon.forEach(([codigo, tinta], c) => casilla(rejilla, 2 + r, c, codigo, tinta, NEGRO)));
    barra(rejilla);
    let cursor = null;
    if (estado.cursor && vistos.length) cursor = { fila: 1 + vistos.length, col: vistos[vistos.length - 1].length };
    if (estado.ventana && ancho >= 80) {
      ventana(rejilla, estado.ventana);
      const v = estado.ventana;
      if (cursor && cursor.col >= v.x && cursor.fila >= v.y && cursor.fila <= v.y + v.renglones + 1) cursor = null;
    }
    return { rejilla, cursor };
  }

  function pintar() {
    const { rejilla, cursor } = componer();
    const fila = ancho * 12;
    rejilla.forEach(([codigo, tinta, fondo], i) => {
      const letra = letras.get(codigo) || letras.get(63);
      let sitio = Math.floor(i / ancho) * 24 * fila + (i % ancho) * 12;
      for (let y = 0, p = 0; y < 24; y++, sitio += fila) {
        for (let x = 0; x < 12; x++, p++) pixeles[sitio + x] = TINTAS[letra[p] ? tinta : fondo];
      }
    });
    if (cursor) {
      let sitio = (cursor.fila * 24 + 22) * fila + cursor.col * 12;
      for (let y = 0; y < 2; y++, sitio += fila) pixeles.fill(TINTAS[BLANCO], sitio, sitio + 12);
    }
    ctx.putImageData(imagen, 0, 0);
  }

  return { estado, medir, pintar };
}

const ACTOS = [
  {
    pie: 'Día 16: <code>busca</code> quiere preguntar al DNS. Hasta que le dejas, no sale nada.',
    pasos: [
      ["teclea", "lanza busca"],
      ["consola", [["  busca: ", GRIS], ["30", BLANCO], [" ", GRIS], ["KB", BLANCO], [" leídos del disco. Nace sin permisos, en su propia caja.", GRIS]]],
      ["espera", 300],
      ["ventana", { numero: 1, nombre: "busca", tinta: VERDE_CLARO, x: 42, y: 15, columnas: 34, renglones: 6, lineas: ["pregunto al DNS por fal-lang.org"] }],
      ["espera", 250],
      ["consola", []],
      ["consola", [["  ", GRIS], ["busca", VERDE_CLARO], [" quiere conectarse a 10.0.2.3, puerto 53 (DNS).", BLANCO]]],
      ["consola", [["  Mandaría 30 bytes. Si le dejas: deja 1. Si no: niega 1.", GRIS]]],
      ["consola", [["  Hasta que me contestes, espera, y no sale nada.", GRIS]]],
      ["consola", []],
      ["cursor"],
      ["espera", 2400],
      ["teclea", "deja 1"],
      ["consola", [["  busca puede hablar con 10.0.2.3:53 hasta que se cierre. Con nadie más.", VERDE_CLARO]]],
      ["consola", []],
      ["cursor"],
      ["aviso", { tinta: VERDE_CLARO, texto: "busca habla con 10.0.2.3" }],
      ["espera", 300],
      ["escribe", "fal-lang.org esta en:", "  185.199.111.153", "  185.199.110.153", "  185.199.108.153", "  185.199.109.153"],
      ["espera", 2700],
      ["aviso", null],
      ["espera", 2600]
    ]
  },
  {
    pie: 'Día 16: <code>chivato</code> quiere contar lo que haces. Le dices que no, lo intenta otra vez, y FalOS ya ni te pregunta.',
    pasos: [
      ["teclea", "lanza chivato"],
      ["consola", [["  chivato: ", GRIS], ["21", BLANCO], [" ", GRIS], ["KB", BLANCO], [" leídos del disco. Nace sin permisos, en su propia caja.", GRIS]]],
      ["espera", 300],
      ["ventana", { numero: 1, nombre: "chivato", tinta: VERDE_CLARO, x: 42, y: 16, columnas: 34, renglones: 6, lineas: ["quiero contarle a un servidor", "lo que haces en FalOS"] }],
      ["espera", 250],
      ["consola", []],
      ["consola", [["  ", GRIS], ["chivato", VERDE_CLARO], [" quiere conectarse a 203.0.113.66, puerto 4444.", BLANCO]]],
      ["consola", [["  Mandaría 24 bytes. Si le dejas: deja 1. Si no: niega 1.", GRIS]]],
      ["consola", [["  Hasta que me contestes, espera, y no sale nada.", GRIS]]],
      ["consola", []],
      ["cursor"],
      ["espera", 2400],
      ["teclea", "niega 1"],
      ["consola", [["  chivato no se conectará a 203.0.113.66:4444. No ha salido nada.", ROJO_CLARO]]],
      ["consola", []],
      ["cursor"],
      ["espera", 300],
      ["escribe", "  no me dejan"],
      ["espera", 1800],
      ["escribe", "lo intento otra vez"],
      ["espera", 150],
      ["escribe", "  no me dejan"],
      ["espera", 1800],
      ["escribe", "pues a otro servidor"],
      ["espera", 200],
      ["sincursor"],
      ["consola", [["  ", GRIS], ["chivato", VERDE_CLARO], [" quiere conectarse a 198.51.100.23, puerto 4444.", BLANCO]]],
      ["consola", [["  Mandaría 24 bytes. Si le dejas: deja 1. Si no: niega 1.", GRIS]]],
      ["consola", [["  Hasta que me contestes, espera, y no sale nada.", GRIS]]],
      ["consola", []],
      ["cursor"],
      ["espera", 4200]
    ]
  }
];

const $foto = document.getElementById("foto");
const $lienzo = document.getElementById("lienzo");
const $pie = document.getElementById("pie");

async function arrancar() {
  const respuesta = await fetch("/falos/letra.txt");
  if (!respuesta.ok) return;
  const pantalla = crearPantalla($lienzo, leerLetra(await respuesta.text()));
  const { estado } = pantalla;
  const ajustar = () => pantalla.medir($lienzo.parentElement.clientWidth < 520 ? 48 : 80);
  ajustar();

  let vuelta = 0, enPantalla = false, turno = 0;

  const esperar = (ms, mia) => new Promise((bien, mal) => setTimeout(() => (mia === vuelta ? bien() : mal()), ms));

  async function teclear(orden, mia) {
    const linea = [[">", CIAN_CLARO], [" ", GRIS]];
    estado.consola.push(linea);
    estado.cursor = "final";
    for (const letra of orden) {
      await esperar(60 + Math.random() * 80, mia);
      linea.push([letra, GRIS]);
      pantalla.pintar();
    }
    await esperar(420, mia);
    estado.cursor = null;
  }

  async function representar(n) {
    const mia = ++vuelta;
    turno = n;
    const acto = ACTOS[n];
    Object.assign(estado, { consola: [], ventana: null, aviso: null, cursor: "nuevo" });
    $pie.innerHTML = acto.pie;
    pantalla.pintar();
    $lienzo.classList.remove("apagado");
    try {
      await esperar(700, mia);
      for (const [que, ...datos] of acto.pasos) {
        if (que === "teclea") await teclear(datos[0], mia);
        else if (que === "espera") await esperar(datos[0], mia);
        else if (que === "consola") estado.consola.push(datos[0]);
        else if (que === "cursor") estado.cursor = "nuevo";
        else if (que === "sincursor") estado.cursor = null;
        else if (que === "ventana") estado.ventana = { ...datos[0], lineas: [...datos[0].lineas] };
        else if (que === "escribe") estado.ventana.lineas.push(...datos);
        else if (que === "aviso") estado.aviso = datos[0];
        pantalla.pintar();
      }
      $lienzo.classList.add("apagado");
      await esperar(450, mia);
    } catch (e) {
      return;
    }
    if (enPantalla) representar((n + 1) % ACTOS.length);
  }

  $foto.hidden = true;
  $lienzo.hidden = false;
  new ResizeObserver(() => { ajustar(); pantalla.pintar(); }).observe($lienzo.parentElement);

  const vigia = new IntersectionObserver(entradas => {
    const antes = enPantalla;
    enPantalla = entradas[0].isIntersecting && !document.hidden;
    if (enPantalla && !antes) representar(turno);
    if (!enPantalla) vuelta++;
  }, { threshold: 0.4 });
  vigia.observe($lienzo);
  document.addEventListener("visibilitychange", () => {
    if (document.hidden) { enPantalla = false; vuelta++; }
    else if ($lienzo.getBoundingClientRect().bottom > 0) { enPantalla = true; representar(turno); }
  });
}

if (!sinMovimiento && "IntersectionObserver" in window && "ResizeObserver" in window) {
  arrancar();

  const tarjetas = document.querySelectorAll(".dias li, .reglas div");
  tarjetas.forEach(t => t.classList.add("aparece"));
  const vigiaTarjetas = new IntersectionObserver(entradas => {
    for (const e of entradas) {
      if (!e.isIntersecting) continue;
      e.target.classList.add("visto");
      vigiaTarjetas.unobserve(e.target);
    }
  }, { rootMargin: "0px 0px -8% 0px" });
  tarjetas.forEach(t => vigiaTarjetas.observe(t));
}
