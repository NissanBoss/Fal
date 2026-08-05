// El interprete corre aqui dentro, en un hilo aparte.
//
// Antes corria en el mismo hilo que la pagina, y eso obligaba a cortar todo
// programa a los cinco segundos: un bucle largo dejaba la pestaña colgada y
// habia que cerrarla a la fuerza. Aqui puede tardar lo que quiera, porque la
// pagina sigue respondiendo, y si se va de las manos se le corta desde fuera.

importScripts("wasm_exec.js");

const go = new Go();

// no-cache no significa "no guardes", sino "pregunta si cambio". Sin esto,
// tras una actualizacion te puede tocar la pagina nueva con el interprete
// viejo durante unos minutos.
const cargado = WebAssembly
  .instantiateStreaming(fetch("fal.wasm", { cache: "no-cache" }), go.importObject)
  .then(res => {
    // go.run arranca el interprete y vuelve en cuanto queda a la espera.
    // No se puede esperar a que acabe: no acaba nunca, a proposito.
    go.run(res.instance);
    self.postMessage({ tipo: "listo" });
  });

self.onmessage = async ev => {
  const m = ev.data || {};

  // Las teclas solo entran aqui mientras el programa deja respirar al
  // navegador, que es lo que hace "espera".
  if (m.tipo === "tecla") {
    if (self.falTecla) self.falTecla(m.tecla);
    return;
  }

  if (m.tipo !== "correr") return;

  try {
    await cargado;
  } catch (e) {
    self.postMessage({ tipo: "fin", error: "No pude cargar el lenguaje: " + e });
    return;
  }

  // falEjecutar vuelve enseguida: el programa se queda corriendo por su
  // cuenta y avisa por el ultimo callback. Asi este hilo queda libre para ir
  // recogiendo las teclas que manda la pagina.
  try {
    falEjecutar(
      m.programa,
      m.respuestas,
      texto => self.postMessage({ tipo: "salida", texto }),
      r => self.postMessage({ tipo: "fin", error: r.error, dibujo: r.dibujo })
    );
  } catch (e) {
    self.postMessage({ tipo: "fin", error: "Algo se rompio por dentro: " + e });
  }
};
