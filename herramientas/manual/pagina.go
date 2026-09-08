package main

// La pagina que envuelve al manual: el mismo papel y la misma tinta que la
// portada, con el indice a un lado.
//
// El buscador filtra el indice y no el texto. Es una decision: buscar dentro
// de novecientas lineas pide un indice invertido y una caja de resultados, y
// lo que la gente escribe en un manual de referencia es el nombre de lo que
// busca. Filtrar los titulos contesta eso en el momento y cabe en diez
// lineas de javascript.

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
		nav.WriteString(fmt.Sprintf(
			`<a class="%s" href="#%s">%s</a>`+"\n", clase, s.Ancla, s.Titulo))
	}

	return `<!doctype html>
<html lang="es">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Manual de Fal</title>
<meta name="description" content="El manual completo de Fal: las 42 palabras del lenguaje, las 93 funciones de la biblioteca, los objetos, los errores, los archivos y una tabla que lo traduce todo a Python y JavaScript.">
<link rel="canonical" href="https://fal-lang.org/manual.html">
<meta property="og:title" content="Manual de Fal">
<meta property="og:description" content="Todo el lenguaje explicado, de las variables a las clausuras.">
<meta property="og:type" content="article">
<meta property="og:url" content="https://fal-lang.org/manual.html">
<style>
:root {
  --papel: #fbfaf7; --caja: #ffffff; --tinta: #1a1a18; --suave: #5f5c55;
  --borde: #e2ded4; --azul: #1c4f8b; --tierra: #9c4420; --codigo-fondo: #f6f4ef;
  --serif: "Iowan Old Style", "Palatino Linotype", Palatino, "Book Antiqua", Georgia, serif;
  --sans: system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", sans-serif;
  --mono: ui-monospace, "Cascadia Code", "SF Mono", Consolas, "Liberation Mono", monospace;
}
@media (prefers-color-scheme: dark) {
  :root {
    --papel: #16171a; --caja: #1d1f23; --tinta: #e8e6e1; --suave: #a2a099;
    --borde: #30323a; --azul: #79b0e4; --tierra: #e0865c; --codigo-fondo: #1a1c20;
  }
}
* { box-sizing: border-box; }
body { margin: 0; background: var(--papel); color: var(--tinta); font: 17px/1.65 var(--sans); }
a { color: var(--azul); }
h1, h2, h3, h4 { font-family: var(--serif); font-weight: 600; line-height: 1.25; }

nav.barra { border-bottom: 1px solid var(--borde); background: var(--papel); position: sticky; top: 0; z-index: 30; }
nav.barra .fila { display: flex; align-items: center; gap: 18px; height: 56px; padding: 0 20px; max-width: 1200px; margin: 0 auto; }
.marca { font-family: var(--serif); font-size: 20px; font-weight: 600; color: var(--tinta); text-decoration: none; }
.marca span { color: var(--tierra); }
.aqui { color: var(--suave); font-size: 15px; }
nav.barra .lejos { margin-left: auto; display: flex; gap: 18px; font-size: 15px; }
nav.barra .lejos a { color: var(--suave); text-decoration: none; }
nav.barra .lejos a:hover { color: var(--tinta); }
#abrirIndice { display: none; background: none; border: 1px solid var(--borde); border-radius: 6px; color: var(--tinta); padding: 5px 11px; font: inherit; font-size: 14px; cursor: pointer; }

.marco { display: grid; grid-template-columns: 262px minmax(0, 1fr); max-width: 1200px; margin: 0 auto; }
aside {
  border-right: 1px solid var(--borde); padding: 18px 0 60px;
  position: sticky; top: 56px; max-height: calc(100vh - 56px); overflow-y: auto;
}
aside .buscar { padding: 0 18px 12px; }
aside input {
  width: 100%; padding: 7px 10px; font: inherit; font-size: 14px;
  border: 1px solid var(--borde); border-radius: 6px;
  background: var(--caja); color: var(--tinta);
}
aside a { display: block; text-decoration: none; color: var(--tinta); font-size: 14.5px; padding: 4px 18px; border-left: 2px solid transparent; }
aside a:hover { background: var(--codigo-fondo); }
aside a.n1 { font-weight: 600; margin-top: 12px; }
aside a.n2 { padding-left: 30px; color: var(--suave); font-size: 14px; }
aside a.viendo { border-left-color: var(--tierra); background: var(--codigo-fondo); }
aside .vacio { padding: 10px 18px; color: var(--suave); font-size: 14px; }

main { padding: 30px 34px 90px; min-width: 0; }
main h1 { font-size: 30px; margin: 34px 0 10px; }
main h1:first-child { margin-top: 4px; }
main h2 { font-size: 23px; margin: 30px 0 8px; }
main h3 { font-size: 18.5px; margin: 24px 0 6px; }
main h4 { font-size: 16.5px; margin: 20px 0 6px; font-family: var(--sans); }
main p { max-width: 40em; }
main ul, main ol { max-width: 40em; }
main li { margin: 4px 0; }
main hr { border: 0; border-top: 1px solid var(--borde); margin: 34px 0; }

pre {
  background: var(--codigo-fondo); border: 1px solid var(--borde); border-radius: 8px;
  padding: 14px 16px; overflow-x: auto; font-size: 14.5px; line-height: 1.55;
  font-family: var(--mono);
}
pre code { font-family: inherit; }
:not(pre) > code {
  font-family: var(--mono); background: var(--codigo-fondo); border: 1px solid var(--borde);
  border-radius: 4px; padding: 1px 5px; font-size: .87em;
}
table { border-collapse: collapse; margin: 14px 0; font-size: 15px; display: block; overflow-x: auto; max-width: 100%; }
th { text-align: left; font-weight: 600; font-size: 12.5px; text-transform: uppercase; letter-spacing: .06em; color: var(--suave); padding: 0 16px 7px 0; border-bottom: 1px solid var(--borde); white-space: nowrap; }
td { padding: 8px 16px 8px 0; border-bottom: 1px solid var(--borde); vertical-align: top; }

.enlazar { color: transparent; text-decoration: none; margin-left: 8px; font-weight: 400; }
h1:hover .enlazar, h2:hover .enlazar, h3:hover .enlazar, h4:hover .enlazar { color: var(--borde); }
.enlazar:hover { color: var(--tierra) !important; }

@media (max-width: 900px) {
  .marco { grid-template-columns: 1fr; }
  aside { display: none; position: static; max-height: none; border-right: 0; border-bottom: 1px solid var(--borde); }
  aside.abierto { display: block; }
  #abrirIndice { display: block; }
  main { padding: 22px 20px 70px; }
}
</style>
</head>
<body>

<nav class="barra">
  <div class="fila">
    <a class="marca" href="/">Fal<span>.</span></a>
    <span class="aqui">Manual</span>
    <button id="abrirIndice" aria-expanded="false">Índice</button>
    <div class="lejos">
      <a href="/probar.html">Probar</a>
      <a href="https://aprende.fal-lang.org">Aprender</a>
      <a href="https://github.com/NissanBoss/Fal">Código</a>
    </div>
  </div>
</nav>

<div class="marco">
  <aside id="indice">
    <div class="buscar">
      <input id="filtro" type="search" placeholder="Buscar en el índice" aria-label="Buscar en el índice">
    </div>
    <div id="listaIndice">
` + nav.String() + `    </div>
    <p class="vacio" id="sinNada" hidden>Nada con ese nombre.</p>
  </aside>

  <main>
` + cuerpo + `  </main>
</div>

<script>
// Filtrar el indice segun se escribe.
const filtro = document.getElementById("filtro");
const enlaces = [...document.querySelectorAll("#listaIndice a")];
const sinNada = document.getElementById("sinNada");

filtro.addEventListener("input", () => {
  // Sin tildes y en minusculas por los dos lados, para que "funcion"
  // encuentre "Funciones" y "Ámbito" salga buscando "ambito".
  const pelado = t => t.toLowerCase().normalize("NFD").replace(/[̀-ͯ]/g, "");
  const busca = pelado(filtro.value.trim());
  let visibles = 0;
  for (const a of enlaces) {
    const cabe = busca === "" || pelado(a.textContent).includes(busca);
    a.hidden = !cabe;
    if (cabe) visibles++;
  }
  sinNada.hidden = visibles > 0;
});

// Marcar en el indice la seccion que se esta leyendo.
const porAncla = new Map(enlaces.map(a => [a.getAttribute("href").slice(1), a]));
const vigilante = new IntersectionObserver(entradas => {
  for (const e of entradas) {
    if (!e.isIntersecting) continue;
    const a = porAncla.get(e.target.id);
    if (!a) continue;
    enlaces.forEach(x => x.classList.remove("viendo"));
    a.classList.add("viendo");
    // Que el indice acompañe a la lectura sin arrastrar la pagina entera.
    if (a.offsetTop < a.parentElement.parentElement.scrollTop ||
        a.offsetTop > a.parentElement.parentElement.scrollTop + window.innerHeight - 140) {
      a.scrollIntoView({ block: "center" });
    }
  }
}, { rootMargin: "0px 0px -75% 0px" });

for (const id of porAncla.keys()) {
  const seccion = document.getElementById(id);
  if (seccion) vigilante.observe(seccion);
}

const abrir = document.getElementById("abrirIndice");
abrir.addEventListener("click", () => {
  const indice = document.getElementById("indice");
  const abierto = indice.classList.toggle("abierto");
  abrir.setAttribute("aria-expanded", abierto ? "true" : "false");
});
document.getElementById("listaIndice").addEventListener("click", () => {
  document.getElementById("indice").classList.remove("abierto");
});
</script>

</body>
</html>
`
}
