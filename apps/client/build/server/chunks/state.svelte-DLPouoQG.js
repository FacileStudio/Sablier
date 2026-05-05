import { B as noop } from './index2-BK72vU4V.js';
import './root-C01qvPdc.js';

const is_legacy = noop.toString().includes("$$") || /function \w+\(\) \{\}/.test(noop.toString());
const placeholder_url = "a:";
if (is_legacy) {
  ({
    url: new URL(placeholder_url)
  });
}
//# sourceMappingURL=state.svelte-DLPouoQG.js.map
