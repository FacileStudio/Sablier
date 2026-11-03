use tauri::{AppHandle, Manager, WebviewUrl, WebviewWindow, WebviewWindowBuilder};

/// Label of the window that shows the Sablier web app.
pub const MAIN: &str = "main";

/// The Sablier server the bundled app and its sign-in talk to.
pub const DEFAULT_SERVER: &str = "https://sablier.facile.studio";

const MARK_DESKTOP: &str = r#"(function () {
  var platform = navigator.userAgent.indexOf("Mac OS X") >= 0 ? "macos" : "other";
  document.documentElement.dataset.desktop = platform;
  if (platform !== "macos") return;
  function addStrip() {
    if (document.getElementById("sablier-desktop-drag")) return;
    var strip = document.createElement("div");
    strip.id = "sablier-desktop-drag";
    strip.setAttribute("data-tauri-drag-region", "");
    strip.style.cssText = "position:fixed;top:0;left:0;right:0;height:var(--desktop-titlebar,40px);z-index:40;-webkit-user-select:none;user-select:none;";
    document.body.appendChild(strip);
  }
  if (document.body) addStrip();
  else document.addEventListener("DOMContentLoaded", addStrip);
})();"#;

/// Shows the Sablier window on `path` of the bundled app, creating it on first use.
pub fn show(app: &AppHandle, path: &str) {
    let handle = app.clone();
    let path = path.to_string();
    let result = app.run_on_main_thread(move || {
        if let Err(err) = open(&handle, &path) {
            eprintln!("sablier: cannot open the Sablier window: {err}");
        }
    });
    if let Err(err) = result {
        eprintln!("sablier: cannot reach the main thread: {err}");
    }
}

/// Hides the Sablier window and drops the dock icon again on macOS.
pub fn hide(window: &tauri::Window) {
    let _ = window.hide();
    #[cfg(target_os = "macos")]
    if window.label() == MAIN {
        let _ = window
            .app_handle()
            .set_activation_policy(tauri::ActivationPolicy::Accessory);
    }
}

fn open(app: &AppHandle, path: &str) -> tauri::Result<()> {
    #[cfg(target_os = "macos")]
    app.set_activation_policy(tauri::ActivationPolicy::Regular)?;
    if let Some(window) = app.get_webview_window(MAIN) {
        if path != "/" {
            let mut url = window.url()?;
            url.set_path(path);
            url.set_query(None);
            window.navigate(url)?;
        }
        return focus(&window);
    }
    let builder = WebviewWindowBuilder::new(app, MAIN, WebviewUrl::App(asset(path).into()))
        .title("Sablier")
        .inner_size(1200.0, 800.0)
        .min_inner_size(820.0, 560.0)
        .center()
        .initialization_script(MARK_DESKTOP);
    #[cfg(target_os = "macos")]
    let builder = builder
        .title_bar_style(tauri::TitleBarStyle::Overlay)
        .hidden_title(true)
        .traffic_light_position(tauri::LogicalPosition::new(20.0, 26.0));
    focus(&builder.build()?)
}

fn focus(window: &WebviewWindow) -> tauri::Result<()> {
    window.unminimize()?;
    window.show()?;
    window.set_focus()
}

/// The bundled asset a route opens from: the SPA entry for the root, the route itself otherwise.
fn asset(path: &str) -> &str {
    match path.trim_start_matches('/') {
        "" => "index.html",
        route => route,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn routes_open_from_the_bundled_app() {
        assert_eq!(asset("/"), "index.html");
        assert_eq!(asset("/dashboard"), "dashboard");
    }
}
