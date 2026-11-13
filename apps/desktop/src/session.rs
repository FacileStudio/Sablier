use crate::config::Credentials;
use crate::popup::{Command, Popup};
use crate::window::DEFAULT_SERVER;
use tauri::State;

/// Menu id that makes the worker re-read the credentials at once.
pub const REFRESH: &str = "refresh";

/// The shared token from the OS keyring (or `~/.sablier.yml` without one), when it belongs to the server the window talks to.
#[tauri::command]
pub fn stored_token() -> Option<String> {
    Credentials::load()
        .filter(|credentials| credentials.server_url == DEFAULT_SERVER)
        .map(|credentials| credentials.token)
}

/// Saves the token the window's sign-in returned, for the window, the tray and the CLI.
#[tauri::command]
pub fn store_token(token: String, popup: State<'_, Popup>) -> Result<(), String> {
    Credentials::save(DEFAULT_SERVER, &token).map_err(|err| err.to_string())?;
    popup.send(Command::Menu(REFRESH.into()));
    Ok(())
}

/// Signs the app out by deleting the shared credentials.
#[tauri::command]
pub fn forget_token(popup: State<'_, Popup>) -> Result<(), String> {
    Credentials::forget().map_err(|err| err.to_string())?;
    popup.send(Command::Menu(REFRESH.into()));
    Ok(())
}
