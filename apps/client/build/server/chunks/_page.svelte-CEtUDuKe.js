import { h as head, d as escape_html, F as attr_class, G as stringify } from './index2-BK72vU4V.js';
import './root-C01qvPdc.js';
import './state.svelte-DLPouoQG.js';
import { B as Button } from './create-id-C41XNBK8.js';
import { L as Label, I as Input } from './label-Bwk0JVVZ.js';

function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let email = "";
    let password = "";
    let busy = false;
    let $$settled = true;
    let $$inner_renderer;
    function $$render_inner($$renderer3) {
      head("1x05zx6", $$renderer3, ($$renderer4) => {
        $$renderer4.title(($$renderer5) => {
          $$renderer5.push(`<title>${escape_html("Log in")} — Sablier</title>`);
        });
      });
      $$renderer3.push(`<div class="grid min-h-screen place-items-center bg-background px-4 py-12"><div class="w-full max-w-sm"><a href="/" class="mb-8 block text-center text-xl font-semibold tracking-tight text-foreground">Sablier</a> <div class="rounded-xl border border-border bg-card p-6 shadow-sm"><div class="mb-6"><h1 class="text-lg font-semibold text-foreground">${escape_html("Welcome back")}</h1> <p class="mt-1 text-sm text-muted-foreground">${escape_html("Log in to your account.")}</p></div> <div class="mb-6 flex rounded-md border border-border p-1"><button${attr_class(`flex-1 rounded py-1.5 text-sm font-medium transition-colors ${stringify(
        "bg-foreground text-background"
      )}`)}>Log in</button> <button${attr_class(`flex-1 rounded py-1.5 text-sm font-medium transition-colors ${stringify("text-muted-foreground hover:text-foreground")}`)}>Register</button></div> <form class="space-y-4"><div class="space-y-1.5">`);
      Label($$renderer3, {
        for: "email",
        children: ($$renderer4) => {
          $$renderer4.push(`<!---->Email`);
        },
        $$slots: { default: true }
      });
      $$renderer3.push(`<!----> `);
      Input($$renderer3, {
        id: "email",
        type: "email",
        placeholder: "you@example.com",
        required: true,
        get value() {
          return email;
        },
        set value($$value) {
          email = $$value;
          $$settled = false;
        }
      });
      $$renderer3.push(`<!----></div> <div class="space-y-1.5">`);
      Label($$renderer3, {
        for: "password",
        children: ($$renderer4) => {
          $$renderer4.push(`<!---->Password`);
        },
        $$slots: { default: true }
      });
      $$renderer3.push(`<!----> `);
      Input($$renderer3, {
        id: "password",
        type: "password",
        placeholder: "••••••••",
        required: true,
        get value() {
          return password;
        },
        set value($$value) {
          password = $$value;
          $$settled = false;
        }
      });
      $$renderer3.push(`<!----></div> `);
      {
        $$renderer3.push("<!--[-1-->");
      }
      $$renderer3.push(`<!--]--> `);
      Button($$renderer3, {
        type: "submit",
        class: "w-full",
        disabled: busy,
        children: ($$renderer4) => {
          $$renderer4.push(`<!---->${escape_html("Log in")}`);
        },
        $$slots: { default: true }
      });
      $$renderer3.push(`<!----></form></div></div></div>`);
    }
    do {
      $$settled = true;
      $$inner_renderer = $$renderer2.copy();
      $$render_inner($$inner_renderer);
    } while (!$$settled);
    $$renderer2.subsume($$inner_renderer);
  });
}

export { _page as default };
//# sourceMappingURL=_page.svelte-CEtUDuKe.js.map
