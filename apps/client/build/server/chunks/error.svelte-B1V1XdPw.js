import { d as escape_html, f as getContext } from './index2-BK72vU4V.js';
import './state.svelte-DLPouoQG.js';
import { s as stores } from './client-cHGcHw9I.js';
import './root-C01qvPdc.js';
import './index-C_-iMXF1.js';

({
  check: stores.updated.check
});
function context() {
  return getContext("__request__");
}
const page$1 = {
  get error() {
    return context().page.error;
  },
  get status() {
    return context().page.status;
  }
};
const page = page$1;
function Error$1($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    $$renderer2.push(`<h1>${escape_html(page.status)}</h1> <p>${escape_html(page.error?.message)}</p>`);
  });
}

export { Error$1 as default };
//# sourceMappingURL=error.svelte-B1V1XdPw.js.map
