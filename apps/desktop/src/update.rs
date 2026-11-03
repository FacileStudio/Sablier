use std::sync::Mutex;
use std::time::Duration;
use tauri::{AppHandle, Manager};
use tauri_plugin_notification::NotificationExt as _;
use tauri_plugin_updater::{Update, UpdaterExt as _};

const FIRST_CHECK: Duration = Duration::from_secs(30);
const EVERY: Duration = Duration::from_secs(6 * 60 * 60);

/// The update found by the last check, and the last version the user was told about.
#[derive(Default)]
pub struct Updates {
    available: Mutex<Option<Update>>,
    announced: Mutex<Option<String>>,
    installing: Mutex<bool>,
}

impl Updates {
    /// The version waiting to be installed, if a check found one.
    pub fn available_version(&self) -> Option<String> {
        self.available
            .lock()
            .ok()?
            .as_ref()
            .map(|update| update.version.clone())
    }

    /// True while an update downloads and installs.
    pub fn installing(&self) -> bool {
        self.installing.lock().map(|flag| *flag).unwrap_or(false)
    }

    fn store(&self, update: Option<Update>) {
        if let Ok(mut slot) = self.available.lock() {
            *slot = update;
        }
    }

    fn take(&self) -> Option<Update> {
        self.available.lock().ok()?.clone()
    }

    fn first_announcement(&self, version: &str) -> bool {
        let Ok(mut announced) = self.announced.lock() else {
            return false;
        };
        should_announce(announced.as_deref(), version) && {
            *announced = Some(version.to_string());
            true
        }
    }

    fn set_installing(&self, value: bool) {
        if let Ok(mut flag) = self.installing.lock() {
            *flag = value;
        }
    }
}

/// True when `version` differs from the one already announced.
pub fn should_announce(announced: Option<&str>, version: &str) -> bool {
    announced != Some(version)
}

/// Checks for updates 30 seconds after launch, then every 6 hours. Release builds only.
pub fn spawn_checks(app: AppHandle) {
    if cfg!(debug_assertions) {
        return;
    }
    tauri::async_runtime::spawn(async move {
        tokio_sleep(FIRST_CHECK).await;
        loop {
            check(&app, false).await;
            tokio_sleep(EVERY).await;
        }
    });
}

/// Runs one check now and says so when Sablier is already up to date.
pub fn check_now(app: &AppHandle) {
    let app = app.clone();
    tauri::async_runtime::spawn(async move { check(&app, true).await });
}

/// Downloads and installs the waiting update, then restarts Sablier.
pub fn install(app: &AppHandle) {
    let app = app.clone();
    tauri::async_runtime::spawn(async move {
        let updates = app.state::<Updates>();
        let Some(update) = updates.take() else {
            return check(&app, true).await;
        };
        if updates.installing() {
            return;
        }
        updates.set_installing(true);
        notify(&app, &format!("Installing Sablier {}…", update.version));
        match update.download_and_install(|_, _| {}, || {}).await {
            Ok(()) => app.restart(),
            Err(err) => {
                updates.set_installing(false);
                eprintln!("sablier: cannot install the update: {err}");
                notify(&app, &format!("The update failed: {err}"));
            }
        }
    });
}

async fn check(app: &AppHandle, manual: bool) {
    let result = match app.updater() {
        Ok(updater) => updater.check().await,
        Err(err) => Err(err),
    };
    let updates = app.state::<Updates>();
    match result {
        Ok(Some(update)) => {
            let version = update.version.clone();
            updates.store(Some(update));
            if manual || updates.first_announcement(&version) {
                notify(
                    app,
                    &format!("Sablier {version} is available. Install it from the tray menu."),
                );
            }
        }
        Ok(None) => {
            updates.store(None);
            if manual {
                let current = app.package_info().version.to_string();
                notify(app, &format!("Sablier {current} is up to date."));
            }
        }
        Err(err) => {
            eprintln!("sablier: update check failed: {err}");
            if manual {
                notify(app, &format!("Cannot check for updates: {err}"));
            }
        }
    }
}

fn notify(app: &AppHandle, body: &str) {
    if let Err(err) = app
        .notification()
        .builder()
        .title("Sablier")
        .body(body)
        .show()
    {
        eprintln!("sablier: cannot notify: {err}");
    }
}

async fn tokio_sleep(duration: Duration) {
    let _ = tauri::async_runtime::spawn_blocking(move || std::thread::sleep(duration)).await;
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn a_version_is_announced_once() {
        assert!(should_announce(None, "0.2.0"));
        assert!(!should_announce(Some("0.2.0"), "0.2.0"));
        assert!(should_announce(Some("0.2.0"), "0.3.0"));
    }
}
