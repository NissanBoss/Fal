// Pintar código de Fal, compartido por la academia, el foro y el playground.
// La portada lleva su propia copia dentro para no pedir un archivo más.

// Las 42 palabras más las nueve que el lenguaje usa sin prohibirlas como
// nombre: en un ejemplo corto casi siempre van de palabra.
const PALABRAS = new Set(("agrega como comparte con continua de detente devuelve entre es " +
  "escribe esta falla falso fin finalmente funcion hereda igual intenta mas menos mi mientras " +
  "nada no nuevo o padre para por pregunta quita relanza repite resto retorna si sino tipo usa " +
  "verdadero y veces hasta mayor menor que cada en desde").split(" "));

export function escapar(t) {
  return t.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}

// FUNCIÓN y funcion son la misma palabra, así que también se pintan igual.
function sinTildes(p) {
  return p.normalize("NFD").replace(/[̀-ͯ]/g, "").toLowerCase();
}

export function pintarLinea(linea) {
  let html = "";
  let i = 0;
  while (i < linea.length) {
    const resto = linea.slice(i);
    if (resto[0] === "#") {
      html += '<span class="com">' + escapar(resto) + "</span>";
      break;
    }
    if (resto[0] === '"') {
      let cierre = linea.indexOf('"', i + 1);
      if (cierre < 0) cierre = linea.length - 1;
      html += '<span class="txt">' + escapar(linea.slice(i, cierre + 1)) + "</span>";
      i = cierre + 1;
      continue;
    }
    const palabra = /^[\p{L}_][\p{L}\p{N}_]*/u.exec(resto);
    if (palabra) {
      const p = palabra[0];
      html += PALABRAS.has(sinTildes(p)) ? '<span class="pal">' + p + "</span>" : escapar(p);
      i += p.length;
      continue;
    }
    html += escapar(resto[0]);
    i++;
  }
  return html;
}

export function pintar(codigo) {
  return codigo.split("\n").map(pintarLinea).join("\n");
}

// Engancha un .editor ya puesto en la página: el textarea de encima, el pre
// pintado de debajo y la columna de números. Devuelve con qué repintarlo
// cuando el texto cambia desde fuera, que el evento input no se entera.
export function montarEditor(caja, alCambiar) {
  const area = caja.querySelector("textarea");
  const pintado = caja.querySelector("pre");
  const numeros = caja.querySelector(".numeros");
  function repintar() {
    // El espacio del final cuenta: sin él, una línea vacía al acabar no ocupa
    // alto en el pre y el texto de encima se descuadra una línea.
    pintado.innerHTML = pintar(area.value) + "\n ";
    const lineas = area.value.split("\n").length;
    numeros.textContent = Array.from({ length: lineas }, (_, n) => n + 1).join("\n");
  }
  area.addEventListener("input", () => { repintar(); if (alCambiar) alCambiar(); });
  area.addEventListener("scroll", () => {
    pintado.style.transform = `translateX(${-area.scrollLeft}px)`;
  });
  // El tabulador mete cuatro espacios en vez de saltar al siguiente botón.
  // Escape lo devuelve a su sitio, para que nadie se quede atrapado dentro
  // del editor usando solo el teclado.
  let atrapaTab = true;
  area.addEventListener("keydown", ev => {
    if (ev.key === "Escape") { atrapaTab = false; return; }
    if (ev.key !== "Tab" || !atrapaTab || ev.shiftKey) return;
    ev.preventDefault();
    const { selectionStart: a, selectionEnd: b, value } = area;
    area.value = value.slice(0, a) + "    " + value.slice(b);
    area.selectionStart = area.selectionEnd = a + 4;
    area.dispatchEvent(new Event("input"));
  });
  area.addEventListener("focus", () => { atrapaTab = true; });
  repintar();
  return {
    area,
    poner(texto) {
      area.value = texto;
      area.scrollLeft = 0;
      pintado.style.transform = "";
      repintar();
    },
    repintar,
  };
}

// El mismo formato que usa el playground al compartir: el programa en base64
// de la variante sin + ni /, detrás de #p=.
export function aTextoDeEnlace(texto) {
  const bytes = new TextEncoder().encode(texto);
  let binario = "";
  for (const b of bytes) binario += String.fromCharCode(b);
  return btoa(binario).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}

export function deTextoDeEnlace(texto) {
  const base = texto.replace(/-/g, "+").replace(/_/g, "/");
  const binario = atob(base + "===".slice((base.length + 3) % 4));
  return new TextDecoder().decode(Uint8Array.from(binario, c => c.charCodeAt(0)));
}

// Cada raya de un dibujo se traza detrás de la anterior, en el orden en que
// la tortuga la hizo: así se entiende que lo ha dibujado un bucle y no que
// es una imagen pegada.
export function ponerDibujo(caja, svg) {
  caja.innerHTML = svg || "";
  const rayas = caja.querySelectorAll("line, path, polyline");
  if (matchMedia("(prefers-reduced-motion: reduce)").matches || !rayas.length) return;
  const paso = Math.min(45, 1800 / rayas.length);
  rayas.forEach((r, n) => {
    const largo = r.getTotalLength ? r.getTotalLength() : 200;
    r.style.strokeDasharray = largo;
    r.style.strokeDashoffset = largo;
    r.getBoundingClientRect();
    r.style.transition = `stroke-dashoffset ${paso * 2}ms ease-out ${n * paso}ms`;
    r.style.strokeDashoffset = 0;
  });
}
