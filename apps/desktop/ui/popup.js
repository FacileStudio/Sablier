const { invoke } = window.__TAURI__.core;
const { listen } = window.__TAURI__.event;

function button(action) {
  const element = document.createElement("button");
  element.textContent = action.label;
  element.className = action.primary ? "primary" : "";
  element.addEventListener("click", () => invoke("answer", { action: action.id }));
  return element;
}

async function render() {
  const prompt = await invoke("current_prompt");
  document.getElementById("title").textContent = prompt ? prompt.title : "";
  document.getElementById("detail").textContent = prompt ? prompt.detail : "";
  document.getElementById("actions").replaceChildren(...(prompt ? prompt.actions.map(button) : []));
}

listen("prompt", render);
render();
