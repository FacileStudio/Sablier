#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod actions;
mod api;
mod config;
mod heartbeat;
mod idle;
mod platform;
mod popup;
mod rules;
mod suggest;
mod tray;
mod window;
mod worker;

use popup::{Command, Popup};
use std::sync::mpsc::{Sender, channel};
use tauri::image::Image;
use tauri::tray::TrayIconBuilder;
use tauri::{App, Manager, RunEvent, WindowEvent};
use tauri_plugin_autostart::{MacosLauncher, ManagerExt as _};

const FIRST_RUN_MARKER: &str = "open-at-login-offered";

fn tray_icon() -> tauri::Result<Image<'static>> {
    if cfg!(target_os = "macos") {
        Image::from_bytes(include_bytes!("../icons/tray-template.png"))
    } else {
        Image::from_bytes(include_bytes!("../icons/tray.png"))
    }
}

fn offer_open_at_login(app: &App) {
    let Ok(dir) = app.path().app_config_dir() else {
        return;
    };
    let marker = dir.join(FIRST_RUN_MARKER);
    if cfg!(debug_assertions) || marker.exists() {
        return;
    }
    if let Err(err) = app.autolaunch().enable() {
        eprintln!("sablier: cannot open at login: {err}");
    }
    let _ = std::fs::create_dir_all(&dir).and_then(|()| std::fs::write(marker, b""));
}

fn setup(
    app: &mut App,
    commands: Sender<Command>,
    inbox: std::sync::mpsc::Receiver<Command>,
) -> tauri::Result<()> {
    #[cfg(target_os = "macos")]
    app.set_activation_policy(tauri::ActivationPolicy::Accessory);
    offer_open_at_login(app);
    let icon = TrayIconBuilder::with_id("sablier")
        .icon(tray_icon()?)
        .icon_as_template(cfg!(target_os = "macos"))
        .tooltip("Sablier")
        .show_menu_on_left_click(true)
        .on_menu_event(move |app, event| match event.id().as_ref() {
            "quit" => app.exit(0),
            id => {
                let _ = commands.send(Command::Menu(id.to_string()));
            }
        })
        .build(app)?;
    let worker = worker::Worker::new(app.handle().clone(), inbox, tray::Tray::new(icon));
    std::thread::Builder::new()
        .name("sablier-worker".into())
        .spawn(move || worker.run())?;
    Ok(())
}

/// Starts the tray app: one instance, a hidden popup for prompts, and the Sablier window on demand.
fn main() {
    let (commands, inbox) = channel();
    let menu_commands = commands.clone();
    tauri::Builder::default()
        .plugin(tauri_plugin_single_instance::init(|_, _, _| {}))
        .plugin(tauri_plugin_autostart::init(
            MacosLauncher::LaunchAgent,
            None,
        ))
        .plugin(tauri_plugin_notification::init())
        .plugin(tauri_plugin_opener::init())
        .manage(Popup::new(commands))
        .invoke_handler(tauri::generate_handler![
            popup::current_prompt,
            popup::answer
        ])
        .setup(move |app| Ok(setup(app, menu_commands, inbox)?))
        .on_window_event(|window, event| {
            if let WindowEvent::CloseRequested { api, .. } = event {
                api.prevent_close();
                window::hide(window);
            }
        })
        .build(tauri::generate_context!())
        .expect("Sablier failed to start")
        .run(|_, event| {
            if let RunEvent::ExitRequested {
                code: None, api, ..
            } = event
            {
                api.prevent_exit();
            }
        });
}
