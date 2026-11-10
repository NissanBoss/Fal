
importScripts("wasm_exec.js");

const go = new Go();

const cargado = WebAssembly
  .instantiateStreaming(fetch("fal.wasm", { cache: "no-cache" }), go.importObject)
  .then(res => {
    go.run(res.instance);
    self.postMessage({ tipo: "listo" });
  });

self.onmessage = async ev => {
  const m = ev.data || {};

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
