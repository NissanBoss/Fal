package main

// La pagina que envuelve al manual. Lleva la hoja y el resaltado que
// comparten el curso, el foro y el playground (comun.css y pintar.js), para
// que el manual no vuelva a ser la unica pagina con otra letra y otros
// colores, que es como estuvo hasta octubre de 2026.
//
// El buscador filtra el indice y no el texto. Es una decision: buscar dentro
// de novecientas lineas pide un indice invertido y una caja de resultados, y
// lo que la gente escribe en un manual de referencia es el nombre de lo que
// busca. Filtrar los titulos contesta eso en el momento.

import (
	"fmt"
	"strings"
)

func envolver(cuerpo string, indice []seccion) string {
	var nav strings.Builder
	for _, s := range indice {
		clase := "n1"
		if s.Nivel == 2 {
			clase = "n2"
		}
		titulo := sinMarcas(s.Titulo)
		// La cabecera de la pagina ya dice que esto es Fal; en el indice
		// esa primera seccion es la que cuenta que es.
		if s.Ancla == "fal" {
			titulo = "Qué es Fal"
		}
		nav.WriteString(fmt.Sprintf(
			`<a class="%s" href="#%s">%s</a>`+"\n", clase, s.Ancla, titulo))
	}

	// El titulo y la primera frase del markdown los dice la cabecera con
	// letra grande, asi que aqui se quitan para no contarlo dos veces.
	if strings.HasPrefix(cuerpo, `<h1 id="fal">`) {
		if fin := strings.Index(cuerpo, "</p>\n"); fin > 0 {
			cuerpo = cuerpo[fin+len("</p>\n"):]
		}
	}

	return `<!doctype html>
<html lang="es">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Manual de Fal</title>
<meta name="description" content="El manual completo de Fal: las 42 palabras del lenguaje, las 93 funciones de la biblioteca, los objetos, los errores, los archivos y una tabla que lo traduce todo a Python y JavaScript.">
<link rel="canonical" href="https://fal-lang.org/manual.html">
<link rel="icon" href="/favicon.ico" sizes="48x48">
<link rel="icon" href="/favicon.svg" type="image/svg+xml">
<link rel="apple-touch-icon" href="/apple-touch-icon.png">
<link rel="preload" href="fuentes/newsreader.woff2" as="font" type="font/woff2" crossorigin>
<link rel="preload" href="fuentes/plex-mono.woff2" as="font" type="font/woff2" crossorigin>
<meta name="theme-color" content="#fbfaf7" media="(prefers-color-scheme: light)">
<meta name="theme-color" content="#16171a" media="(prefers-color-scheme: dark)">
<meta property="og:title" content="Manual de Fal">
<meta property="og:description" content="Todo el lenguaje explicado, de las variables a las clausuras.">
<meta property="og:type" content="article">
<meta property="og:url" content="https://fal-lang.org/manual.html">
<link rel="stylesheet" href="comun.css">
<style>
nav.sitio .fila { max-width: none; padding: 0 24px; }
@media (max-width: 520px) { nav.sitio .fila { padding: 0 16px; } }

.disposicion { display: grid; grid-template-columns: 280px minmax(0, 1fr); }

aside {
  border-right: 1px solid var(--borde); padding: 22px 0 50px;
  position: sticky; top: var(--alto-barra); align-self: start;
  max-height: calc(100vh - var(--alto-barra)); overflow-y: auto; overscroll-behavior: contain;
}
.buscar { position: relative; padding: 0 20px 14px; }
.buscar input {
  width: 100%; padding: 8px 34px 8px 11px; font: inherit; font-size: 14.5px;
  border: 1px solid var(--borde); border-radius: 8px; background: var(--caja); color: var(--tinta);
}
.buscar input:focus { outline: none; border-color: var(--azul); box-shadow: 0 0 0 3px color-mix(in srgb, var(--azul) 16%, transparent); }
.buscar kbd { position: absolute; right: 29px; top: 9px; color: var(--suave); background: var(--papel); }
.buscar input:focus + kbd, .buscar input:not(:placeholder-shown) + kbd { display: none; }

/* Las secciones grandes en cursiva y las de dentro colgando de una raya,
   como las del curso. Antes eran todas la misma lista en negrita o en gris
   y no se veía dónde empezaba cada parte. */
#listaIndice a {
  display: block; text-decoration: none; color: var(--suave);
  font-size: 14px; line-height: 1.35; padding: 4px 20px 4px 33px;
  border-left: 2px solid transparent; margin-left: 20px;
}
#listaIndice a:hover { color: var(--tinta); }
#listaIndice a.n1 {
  font: italic 400 17px/1.3 var(--serif); color: var(--tinta);
  margin: 16px 0 3px; padding: 0 20px; border: 0;
}
#listaIndice a.n1:first-child { margin-top: 4px; }
#listaIndice a.n2 { border-left-color: var(--borde); padding-left: 13px; }
#listaIndice a.viendo { color: var(--tinta); border-left-color: var(--tierra); }
#listaIndice a.n1.viendo { color: var(--tierra); }
#listaIndice mark { background: var(--tierra-velo); color: inherit; border-radius: 2px; }
.vacio { padding: 6px 20px; color: var(--suave); font-size: 14px; }

main { padding: 34px 48px 100px; min-width: 0; }
.ancho { max-width: 760px; }

.cabeza { padding-bottom: 30px; margin-bottom: 8px; border-bottom: 1px solid var(--borde); }
.cabeza h1 { font-size: clamp(40px, 6vw, 60px); font-weight: 500; letter-spacing: -0.02em; margin: 6px 0 14px; }
.cabeza .entrada { font: 400 21px/1.5 var(--serif); margin: 0 0 26px; max-width: 34em; color: var(--tinta); }
.cifras { display: grid; grid-template-columns: repeat(4, 1fr); gap: 1px; background: var(--borde); border: 1px solid var(--borde); border-radius: 12px; overflow: hidden; }
.cifras a { background: var(--papel); padding: 14px 16px 13px; text-decoration: none; color: var(--suave); font-size: 13.5px; line-height: 1.3; transition: background-color .15s; }
.cifras a:hover { background: var(--caja); color: var(--tinta); }
.cifras b { display: block; font: 400 36px/1 var(--serif); color: var(--tierra); letter-spacing: -0.02em; margin-bottom: 5px; }
@media (max-width: 560px) { .cifras { grid-template-columns: repeat(2, 1fr); } }

.texto { font-size: 16.5px; line-height: 1.7; }
.texto h1 { font-size: 36px; font-weight: 500; letter-spacing: -0.01em; margin: 0 0 14px; padding-top: 6px; }
.texto h2 { font-size: 26px; font-weight: 500; margin: 42px 0 10px; }
.texto h3 { font-size: 20px; font-weight: 600; margin: 30px 0 8px; }
.texto h4 { font: 600 16px/1.4 var(--sans); margin: 24px 0 6px; }
.texto p, .texto ul, .texto ol { max-width: 40em; text-wrap: pretty; }
.texto li { margin: 5px 0; }
.texto li::marker { color: var(--suave); }
.texto ol > li::marker { font-family: var(--serif); }
.texto strong { font-weight: 600; }
.texto :not(pre) > code { white-space: nowrap; }
/* Las rayas de separación marcan dónde empieza otra parte del manual:
   un hueco con más aire que el de una sección cualquiera. */
.texto hr { border: 0; height: 0; margin: 64px 0 40px; border-top: 1px solid var(--borde); }

.enlazar {
  color: var(--borde); text-decoration: none; font-weight: 400; margin-left: 10px;
  opacity: 0; transition: opacity .15s;
}
.texto :is(h1, h2, h3, h4):hover .enlazar, .enlazar:focus-visible { opacity: 1; }
.enlazar:hover { color: var(--tierra); }
@media (hover: none) { .enlazar { display: none; } }

.texto pre {
  position: relative; margin: 16px 0 20px;
  background: var(--codigo-fondo); border: 1px solid var(--borde); border-radius: 10px;
  padding: 15px 18px; overflow-x: auto; font-size: 14.5px; line-height: 1.6;
}
.texto pre code { font-family: inherit; }
.texto pre.lenguaje-bash { background: transparent; border-style: dashed; }
.texto pre.lenguaje-bash code::before { content: "$ "; color: var(--suave); user-select: none; }
.texto pre.lenguaje-go { color: var(--suave); }

/* Copiar y probar salen al pasar por encima. En el móvil no hay encima y
   puestos sobre el código tapaban el final de las líneas largas, así que
   allí bajan a una fila propia debajo del bloque. */
.bloque { position: relative; margin: 16px 0 20px; }
.bloque pre { margin: 0; }
.acciones {
  position: absolute; top: 8px; right: 8px; display: flex; gap: 6px;
  opacity: 0; transition: opacity .15s;
}
.bloque:hover .acciones, .acciones:focus-within { opacity: 1; }
.acciones :is(a, button) {
  font: 500 12.5px/1 var(--sans); padding: 6px 9px; border-radius: 6px;
  border: 1px solid var(--borde); background: var(--caja); color: var(--suave); text-decoration: none;
}
.acciones :is(a, button):hover { color: var(--tinta); border-color: color-mix(in srgb, var(--suave) 50%, var(--borde)); }
@media (hover: none) {
  .acciones { position: static; opacity: 1; justify-content: flex-end; margin-top: 6px; }
}

.texto table { border-collapse: collapse; margin: 18px 0 24px; font-size: 15px; line-height: 1.5; display: block; overflow-x: auto; max-width: 100%; }
.texto th {
  text-align: left; font-weight: 400; color: var(--suave); white-space: nowrap;
  letter-spacing: 0.08em; font-variant-caps: all-small-caps; font-size: 15px;
  padding: 0 22px 8px 0; border-bottom: 1px solid var(--tinta);
}
.texto td { padding: 9px 22px 9px 0; border-bottom: 1px solid var(--borde); vertical-align: top; }
.texto td:first-child { white-space: nowrap; }
.texto tr:last-child td { border-bottom: 0; }

.subir {
  position: fixed; right: 22px; bottom: 22px; z-index: 20;
  width: 42px; height: 42px; border-radius: 50%; border: 1px solid var(--borde);
  background: var(--caja); color: var(--suave); box-shadow: var(--sombra);
  display: grid; place-items: center; opacity: 0; pointer-events: none; transition: opacity .2s;
}
.subir.ver { opacity: 1; pointer-events: auto; }
.subir:hover { color: var(--tinta); }
.subir svg { width: 14px; height: 14px; }

.abrir-indice { display: none; }
@media (max-width: 900px) {
  .disposicion { grid-template-columns: 1fr; }
  aside {
    position: fixed; z-index: 40; top: var(--alto-barra); left: 0; bottom: 0; width: min(320px, 86vw);
    max-height: none; background: var(--papel); box-shadow: var(--sombra);
    /* Con el start de arriba el cajón medía lo que su contenido y, como no
       cabía, el navegador lo subía hasta tapar la barra. */
    align-self: stretch;
    transform: translateX(-102%); visibility: hidden;
    transition: transform .3s cubic-bezier(.3, .7, .2, 1), visibility 0s .3s;
  }
  /* Escondido de verdad y no solo fuera de la pantalla: asomaba la sombra
     por el borde y el tabulador se metía en enlaces que no se veían. */
  aside.abierto { transform: none; visibility: visible; transition-delay: 0s; }
  .velo { position: fixed; inset: var(--alto-barra) 0 0 0; z-index: 35; background: rgba(20, 18, 14, .35); }
  main { padding: 20px 16px 70px; }
  /* Una franja pegada bajo la barra y no un botón flotante: el botón
     tapaba media línea del título que tenía debajo. */
  html { scroll-padding-top: calc(var(--alto-barra) + 60px); }
  .abrir-indice {
    display: flex; align-items: center; gap: 10px; width: calc(100% + 32px); margin: -20px -16px 22px;
    position: sticky; top: var(--alto-barra); z-index: 10;
    padding: 11px 16px; border: 0; border-bottom: 1px solid var(--borde); font-size: 14.5px; text-align: left;
    background: color-mix(in srgb, var(--papel) 90%, transparent);
    -webkit-backdrop-filter: blur(10px); backdrop-filter: blur(10px);
  }
  .abrir-indice b { font-weight: 600; flex: none; }
  .abrir-indice span { color: var(--suave); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; flex: 1; }
  .abrir-indice svg { width: 12px; height: 12px; flex: none; color: var(--suave); }
  .texto h1 { font-size: 30px; }
  .texto h2 { font-size: 23px; }
  .subir { right: 16px; bottom: 16px; }
}
</style>
</head>
<body>

<nav class="sitio">
  <div class="fila">
    <a class="marca" href="./">Fal<span>.</span></a>
    <div class="enlaces">
      <a href="probar.html">Probar</a>
      <a href="https://aprende.fal-lang.org">Aprender</a>
      <a href="https://foro.fal-lang.org/">Foro</a>
      <a class="opcional" href="manual.html" aria-current="page">Manual</a>
      <a class="opcional" href="https://github.com/NissanBoss/Fal">Código</a>
    </div>
  </div>
</nav>

<div class="disposicion">
  <aside id="indice" aria-label="Índice del manual">
    <div class="buscar">
      <input id="filtro" type="search" placeholder="Buscar en el índice" aria-label="Buscar en el índice" autocomplete="off">
      <kbd class="teclas">/</kbd>
    </div>
    <nav id="listaIndice">
` + nav.String() + `    </nav>
    <p class="vacio" id="sinNada" hidden>Nada con ese nombre.</p>
  </aside>
  <div class="velo" id="velo" hidden></div>

  <main>
  <div class="ancho">
    <button class="abrir-indice" id="abrirIndice" aria-expanded="false" aria-controls="indice" aria-label="Abrir el índice">
      <b>Manual</b> <span id="dondeEstoy"></span>
      <svg viewBox="0 0 12 12" aria-hidden="true"><path d="M2.5 4.5 6 8l3.5-3.5" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
    </button>

    <header class="cabeza">
      <p class="versalitas">Manual</p>
      <h1 id="fal">Todo Fal en una página</h1>
      <p class="entrada">Un lenguaje de programación completo escrito en español y sin un solo símbolo raro. Todo se lee en voz alta.</p>
      <div class="cifras">
        <a href="#palabras-reservadas-42"><b>42</b>palabras reservadas</a>
        <a href="#todo-lo-que-trae-puesto-93-funciones"><b>93</b>funciones de serie</a>
        <a href="#las-9-clases-de-error"><b>9</b>clases de error</a>
        <a href="#si-vienes-de-otro-lenguaje"><b>1</b>tabla para pasar a Python</a>
      </div>
    </header>

    <article class="texto">
` + cuerpo + `    </article>
  </div>
  </main>
</div>

<button class="subir" id="subir" aria-label="Volver arriba">
  <svg viewBox="0 0 14 14" aria-hidden="true"><path d="M7 12V2M2.5 6.5 7 2l4.5 4.5" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/></svg>
</button>

<footer class="pie">
  <div class="centro">
    <a href="aviso-legal.html">Aviso legal</a>
    <a href="privacidad.html">Privacidad</a>
    <a href="cookies.html">Cookies</a>
    <a href="https://github.com/NissanBoss/Fal">Código</a>
    <span class="lejos">Cuarenta y dos palabras. Ni un símbolo.</span>
  </div>
</footer>

<script type="module">
import { pintar, aTextoDeEnlace } from "./pintar.js";

const $ = (id) => document.getElementById(id);
const enlaces = [...document.querySelectorAll("#listaIndice a")];
const textos = new Map(enlaces.map((a) => [a, a.textContent]));

// Los bloques sin clase son de Fal; los de la consola y el de Go llevan la
// suya y se quedan sin pintar, que si no el resaltado marcaba "func" y "if"
// como si fueran palabras nuestras.
for (const pre of document.querySelectorAll(".texto pre")) {
  const codigo = pre.querySelector("code");
  const fuente = codigo.textContent.replace(/\n$/, "");
  const esFal = !pre.className || pre.classList.contains("corre");
  if (esFal) codigo.innerHTML = pintar(fuente);

  const acciones = document.createElement("div");
  acciones.className = "acciones";
  // Probar solo en los que corren tal cual. En el primero, por ejemplo,
  // falta Figura y llevarlo al playground era estrenarse con un error.
  if (pre.classList.contains("corre")) {
    const probar = document.createElement("a");
    probar.href = "probar.html#p=" + aTextoDeEnlace(fuente);
    probar.textContent = "Probar";
    acciones.append(probar);
  }
  const copiar = document.createElement("button");
  copiar.type = "button";
  copiar.textContent = "Copiar";
  copiar.onclick = async () => {
    try {
      await navigator.clipboard.writeText(fuente);
      copiar.textContent = "Copiado";
    } catch {
      copiar.textContent = "No se pudo";
    }
    setTimeout(() => { copiar.textContent = "Copiar"; }, 1600);
  };
  acciones.append(copiar);
  const bloque = document.createElement("div");
  bloque.className = "bloque";
  pre.replaceWith(bloque);
  bloque.append(pre, acciones);
}

// Filtrar el índice según se escribe. Sin tildes y en minúsculas por los
// dos lados, para que "funcion" encuentre "Funciones" y "Ámbito" salga
// buscando "ambito".
const pelado = (t) => t.toLowerCase().normalize("NFD").replace(/[̀-ͯ]/g, "");
const filtro = $("filtro");
filtro.addEventListener("input", () => {
  const busca = pelado(filtro.value.trim());
  let visibles = 0;
  for (const a of enlaces) {
    const original = textos.get(a);
    const donde = busca ? pelado(original).indexOf(busca) : -1;
    const cabe = busca === "" || donde >= 0;
    a.hidden = !cabe;
    if (!cabe) continue;
    visibles++;
    // Quitar las tildes no cambia la longitud del texto (cada letra suelta
    // su marca y nada más), así que la posición vale para el original.
    if (donde >= 0) {
      a.replaceChildren(original.slice(0, donde), Object.assign(document.createElement("mark"), {
        textContent: original.slice(donde, donde + busca.length),
      }), original.slice(donde + busca.length));
    } else {
      a.textContent = original;
    }
  }
  $("sinNada").hidden = visibles > 0;
});
filtro.addEventListener("keydown", (ev) => {
  if (ev.key === "Enter") {
    const primero = enlaces.find((a) => !a.hidden);
    if (primero) { primero.click(); filtro.blur(); }
  }
  if (ev.key === "Escape" && filtro.value) {
    filtro.value = "";
    filtro.dispatchEvent(new Event("input"));
    ev.stopPropagation();
  }
});
addEventListener("keydown", (ev) => {
  if (ev.key === "/" && !/^(INPUT|TEXTAREA)$/.test(document.activeElement.tagName)) {
    ev.preventDefault();
    abrirIndice();
    filtro.focus();
  }
});

// Marcar en el índice la sección que se está leyendo: la última cuyo título
// ya ha pasado por arriba. El vigilante de antes solo se enteraba cuando un
// título entraba en pantalla, y en una sección larga se quedaba marcada la
// anterior o ninguna.
const titulos = enlaces.map((a) => $(decodeURIComponent(a.hash.slice(1)))).filter(Boolean);
const porId = new Map(enlaces.map((a) => [decodeURIComponent(a.hash.slice(1)), a]));
let marcado = null;
function marcar() {
  const linea = parseFloat(getComputedStyle(document.documentElement).scrollPaddingTop) + 30;
  let actual = titulos[0];
  for (const t of titulos) {
    if (t.getBoundingClientRect().top <= linea) actual = t;
    else break;
  }
  const a = porId.get(actual.id);
  $("subir").classList.toggle("ver", scrollY > innerHeight * 1.5);
  if (a === marcado) return;
  marcado?.classList.remove("viendo");
  a.classList.add("viendo");
  marcado = a;
  $("dondeEstoy").textContent = textos.get(a);
  acercar(a);
}

// Mover el índice sin scrollIntoView, que arrastraba también la página y
// daba tirones mientras leías. El aside está posicionado (sticky o fixed),
// así que offsetTop ya se mide desde él.
function acercar(a) {
  const indice = $("indice");
  if (a.offsetTop < indice.scrollTop + 60 || a.offsetTop > indice.scrollTop + indice.clientHeight - 80) {
    indice.scrollTo({ top: a.offsetTop - indice.clientHeight / 3 });
  }
}
let pedido = false;
addEventListener("scroll", () => {
  if (pedido) return;
  pedido = true;
  requestAnimationFrame(() => { pedido = false; marcar(); });
}, { passive: true });
marcar();

$("subir").onclick = () => {
  scrollTo({ top: 0, behavior: matchMedia("(prefers-reduced-motion: reduce)").matches ? "auto" : "smooth" });
};

// El índice del móvil, igual que el del curso.
function abrirIndice() {
  if (!matchMedia("(max-width: 900px)").matches) return;
  $("indice").classList.add("abierto");
  $("velo").hidden = false;
  $("abrirIndice").setAttribute("aria-expanded", "true");
  if (marcado) acercar(marcado);
}
function cerrarIndice() {
  $("indice").classList.remove("abierto");
  $("velo").hidden = true;
  $("abrirIndice").setAttribute("aria-expanded", "false");
}
$("abrirIndice").onclick = abrirIndice;
$("velo").onclick = cerrarIndice;
$("listaIndice").addEventListener("click", cerrarIndice);
addEventListener("keydown", (ev) => { if (ev.key === "Escape") cerrarIndice(); });
</script>

</body>
</html>
`
}

// sinMarcas deja un titulo en texto llano para el indice. Antes se ponia tal
// cual y uno salia con los asteriscos de la negrita a la vista.
func sinMarcas(titulo string) string {
	titulo = strings.ReplaceAll(titulo, "**", "")
	titulo = strings.ReplaceAll(titulo, "`", "")
	return escapar(titulo)
}
