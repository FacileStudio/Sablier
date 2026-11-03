use tauri::ipc::CapabilityBuilder;
use tauri::{AppHandle, Manager, Url, WebviewUrl, WebviewWindow, WebviewWindowBuilder};

/// Label of the window that shows the Sablier web app.
pub const MAIN: &str = "main";

/// Server opened when no `~/.sablier.yml` names one.
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

/// The URL the main window opens for `server` and `path`, when `server` is an http(s) origin.
pub fn page_url(server: &str, path: &str) -> Option<Url> {
    let base = Url::parse(server.trim_end_matches('/')).ok()?;
    if !matches!(base.scheme(), "https" | "http") || base.host_str().is_none() {
        return None;
    }
    base.join(path).ok()
}

/// The remote URL pattern that may drag the window: every page of the server's origin.
pub fn remote_pattern(url: &Url) -> String {
    format!("{}/*", url.origin().ascii_serialization())
}

/// Shows the Sablier window on `path`, creating it on first use.
pub fn show(app: &AppHandle, server: &str, path: &str) {
    let Some(url) = page_url(server, path) else {
        return eprintln!("sablier: {server:?} is not an http(s) server URL");
    };
    let handle = app.clone();
    let result = app.run_on_main_thread(move || {
        if let Err(err) = open(&handle, url) {
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

fn open(app: &AppHandle, url: Url) -> tauri::Result<()> {
    #[cfg(target_os = "macos")]
    app.set_activation_policy(tauri::ActivationPolicy::Regular)?;
    if let Some(window) = app.get_webview_window(MAIN) {
        if window
            .url()
            .map(|current| current.origin() != url.origin())
            .unwrap_or(true)
        {
            window.navigate(url)?;
        }
        return focus(&window);
    }
    app.add_capability(
        CapabilityBuilder::new("main-window-drag")
            .remote(remote_pattern(&url))
            .local(false)
            .window(MAIN)
            .permission("core:window:allow-start-dragging")
            .permission("core:window:allow-internal-toggle-maximize"),
    )?;
    let builder = WebviewWindowBuilder::new(app, MAIN, WebviewUrl::External(url))
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

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn page_urls_join_the_path_onto_the_server() {
        let url = page_url("https://sablier.facile.studio/", "/dashboard").unwrap();
        assert_eq!(url.as_str(), "https://sablier.facile.studio/dashboard");
        assert_eq!(remote_pattern(&url), "https://sablier.facile.studio/*");
    }

    #[test]
    fn non_web_servers_are_refused() {
        assert!(page_url("file:///etc/passwd", "/").is_none());
        assert!(page_url("javascript:alert(1)", "/").is_none());
        assert!(page_url("not a url", "/").is_none());
    }
}
