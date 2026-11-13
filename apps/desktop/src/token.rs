/// The keyring entry the desktop app shares with the `sablier` CLI (sablier-cli's src/vault.rs).
pub const SERVICE: &str = "studio.facile.sablier";
pub const ACCOUNT: &str = "session";

/// Where the bearer token lives when it is not in `~/.sablier.yml`. Each error is the reason the
/// store cannot be used.
pub trait Vault {
    fn get(&self) -> Result<Option<String>, String>;
    fn set(&self, token: &str) -> Result<(), String>;
    fn delete(&self) -> Result<(), String>;
}

/// The OS keyring: Keychain on macOS, Credential Manager on Windows, the Secret Service on Linux.
pub struct Keyring;

fn entry() -> Result<keyring::Entry, String> {
    keyring::Entry::new(SERVICE, ACCOUNT).map_err(|err| err.to_string())
}

impl Vault for Keyring {
    fn get(&self) -> Result<Option<String>, String> {
        match entry()?.get_password() {
            Ok(token) => Ok(Some(token)),
            Err(keyring::Error::NoEntry) => Ok(None),
            Err(err) => Err(err.to_string()),
        }
    }

    fn set(&self, token: &str) -> Result<(), String> {
        entry()?.set_password(token).map_err(|err| err.to_string())
    }

    fn delete(&self) -> Result<(), String> {
        match entry()?.delete_credential() {
            Ok(()) | Err(keyring::Error::NoEntry) => Ok(()),
            Err(err) => Err(err.to_string()),
        }
    }
}

/// Writes `token` and reads it back, so a store that accepts a write and keeps nothing never costs
/// the copy on disk.
pub fn store(vault: &dyn Vault, token: &str) -> Result<(), String> {
    vault.set(token)?;
    match vault.get()? {
        Some(stored) if stored == token => Ok(()),
        _ => Err("the keyring did not keep the token".into()),
    }
}

/// Says once per run that the token stays on disk, and why.
pub fn fallback_notice(path: &std::path::Path, reason: &str) {
    static ONCE: std::sync::Once = std::sync::Once::new();
    ONCE.call_once(|| {
        eprintln!(
            "sablier: no usable OS keyring ({reason}); the token stays in {} (mode 0600)",
            path.display()
        );
    });
}
