use crate::token::{self, Keyring, Vault};
use serde::{Deserialize, Serialize};
use std::path::{Path, PathBuf};

/// The server `sablier login` stores in `~/.sablier.yml` and the token it keeps in the OS keyring.
#[derive(Debug, Clone, Deserialize, Serialize, PartialEq)]
pub struct Credentials {
    pub server_url: String,
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub token: String,
}

fn credentials_path() -> Option<PathBuf> {
    Some(dirs::home_dir()?.join(".sablier.yml"))
}

impl Credentials {
    /// Reads the shared sign-in, or `None` when the user has not signed in.
    pub fn load() -> Option<Self> {
        Self::load_from(&credentials_path()?, &Keyring)
    }

    /// Stores `token` in the OS keyring and `server` in `~/.sablier.yml`, in the CLI's format. Without
    /// a usable keyring the token goes into the file, readable by the owner only.
    pub fn save(server: &str, token: &str) -> std::io::Result<()> {
        let path = credentials_path().ok_or_else(|| std::io::Error::other("no home directory"))?;
        Self::save_to(&path, &Keyring, server, token)
    }

    /// Deletes the keyring entry and `~/.sablier.yml`, signing out the app, the tray and the CLI together.
    pub fn forget() -> std::io::Result<()> {
        let path = credentials_path().ok_or_else(|| std::io::Error::other("no home directory"))?;
        Self::forget_at(&path, &Keyring)
    }

    /// Removes the file, then the keyring entry. A keyring that cannot be reached is not an error
    /// unless it still hands the token out.
    fn forget_at(path: &Path, vault: &dyn Vault) -> std::io::Result<()> {
        match std::fs::remove_file(path) {
            Err(err) if err.kind() != std::io::ErrorKind::NotFound => return Err(err),
            _ => {}
        }
        match vault.delete() {
            Err(reason) if matches!(vault.get(), Ok(Some(_))) => Err(std::io::Error::other(reason)),
            _ => Ok(()),
        }
    }

    /// Reads `path` and the keyring. A token still in the file moves to the keyring and leaves the
    /// file only once the keyring returned it; without a usable keyring it stays in the file.
    fn load_from(path: &Path, vault: &dyn Vault) -> Option<Self> {
        let mut stored: Self = serde_yaml::from_str(&std::fs::read_to_string(path).ok()?).ok()?;
        stored.token = stored.token.trim().to_string();
        if stored.token.is_empty() {
            stored.token = vault.get().ok().flatten()?;
            return Self::normalize(stored);
        }
        match token::store(vault, &stored.token) {
            Ok(()) => {
                let server_only = Self {
                    server_url: stored.server_url.clone(),
                    token: String::new(),
                };
                if let Ok(text) = serde_yaml::to_string(&server_only) {
                    let _ = write_private(path, &text);
                }
            }
            Err(reason) => token::fallback_notice(path, &reason),
        }
        Self::normalize(stored)
    }

    fn save_to(path: &Path, vault: &dyn Vault, server: &str, token: &str) -> std::io::Result<()> {
        let kept = match token::store(vault, token.trim()) {
            Ok(()) => "",
            Err(reason) => {
                token::fallback_notice(path, &reason);
                token
            }
        };
        write_private(
            path,
            &Self::render(server, kept).map_err(std::io::Error::other)?,
        )
    }

    /// The YAML the CLI expects: the server with its `/api` suffix, then the token when it has to
    /// live in the file.
    pub fn render(server: &str, token: &str) -> Result<String, serde_yaml::Error> {
        serde_yaml::to_string(&Self {
            server_url: format!("{}/api", server.trim().trim_end_matches('/')),
            token: token.trim().to_string(),
        })
    }

    /// Drops the `/api` suffix the CLI stores; an empty token or server counts as signed out.
    fn normalize(mut credentials: Self) -> Option<Self> {
        credentials.server_url = credentials
            .server_url
            .trim()
            .trim_end_matches('/')
            .trim_end_matches("/api")
            .trim_end_matches('/')
            .to_string();
        if credentials.server_url.is_empty() || credentials.token.trim().is_empty() {
            return None;
        }
        Some(credentials)
    }
}

#[cfg(unix)]
fn write_private(path: &std::path::Path, text: &str) -> std::io::Result<()> {
    use std::io::Write;
    use std::os::unix::fs::OpenOptionsExt;
    let mut file = std::fs::OpenOptions::new()
        .write(true)
        .create(true)
        .truncate(true)
        .mode(0o600)
        .open(path)?;
    file.write_all(text.as_bytes())
}

#[cfg(not(unix))]
fn write_private(path: &std::path::Path, text: &str) -> std::io::Result<()> {
    std::fs::write(path, text)
}

/// The rules file, `~/.config/sablier/rules.yml` on every platform.
pub fn rules_path() -> Option<PathBuf> {
    Some(
        dirs::home_dir()?
            .join(".config")
            .join("sablier")
            .join("rules.yml"),
    )
}

/// Where the offline heartbeat queue lives: the XDG state dir, else the local data dir.
pub fn queue_path() -> Option<PathBuf> {
    let base = dirs::state_dir().or_else(dirs::data_local_dir)?;
    Some(base.join("sablier").join("heartbeats.json"))
}

#[cfg(test)]
mod keyring_tests;

#[cfg(test)]
mod tests {
    use super::*;

    fn parse(text: &str) -> Option<Credentials> {
        Credentials::normalize(serde_yaml::from_str(text).ok()?)
    }

    #[test]
    fn reads_the_cli_config_and_trims_the_trailing_slash() {
        let parsed = parse("server_url: https://sablier.facile.studio/\ntoken: abc\n").unwrap();
        assert_eq!(parsed.server_url, "https://sablier.facile.studio");
        assert_eq!(parsed.token, "abc");
    }

    #[test]
    fn strips_the_api_suffix_the_cli_stores() {
        let parsed = parse("server_url: https://sablier.facile.studio/api/\ntoken: abc\n").unwrap();
        assert_eq!(parsed.server_url, "https://sablier.facile.studio");
    }

    #[test]
    fn rendered_credentials_read_back_as_the_cli_writes_them() {
        let text = Credentials::render("https://sablier.facile.studio/", "abc").unwrap();
        assert!(text.contains("server_url: https://sablier.facile.studio/api"));
        let parsed = parse(&text).unwrap();
        assert_eq!(parsed.server_url, "https://sablier.facile.studio");
        assert_eq!(parsed.token, "abc");
    }

    #[test]
    fn a_signed_out_config_is_no_credentials() {
        assert!(parse("server_url: https://x.test\n").is_none());
        assert!(parse("server_url: https://x.test\ntoken: ''\n").is_none());
        assert!(parse("not yaml: [").is_none());
    }
}
