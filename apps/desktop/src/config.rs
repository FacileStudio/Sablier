use serde::{Deserialize, Serialize};
use std::path::PathBuf;

/// The server and API token `sablier login` stores in `~/.sablier.yml`.
#[derive(Debug, Clone, Deserialize, Serialize, PartialEq)]
pub struct Credentials {
    pub server_url: String,
    #[serde(default)]
    pub token: String,
}

fn credentials_path() -> Option<PathBuf> {
    Some(dirs::home_dir()?.join(".sablier.yml"))
}

impl Credentials {
    /// Reads the CLI's credentials, or `None` when the user has not signed in.
    pub fn load() -> Option<Self> {
        Self::parse(&std::fs::read_to_string(credentials_path()?).ok()?)
    }

    /// Writes `server` and `token` to `~/.sablier.yml` in the CLI's format, readable by the owner only.
    pub fn save(server: &str, token: &str) -> std::io::Result<()> {
        let path = credentials_path().ok_or_else(|| std::io::Error::other("no home directory"))?;
        let text = Self::render(server, token).map_err(std::io::Error::other)?;
        write_private(&path, &text)
    }

    /// Deletes `~/.sablier.yml`, signing out the app, the tray and the CLI together.
    pub fn forget() -> std::io::Result<()> {
        let Some(path) = credentials_path() else {
            return Ok(());
        };
        match std::fs::remove_file(path) {
            Err(err) if err.kind() != std::io::ErrorKind::NotFound => Err(err),
            _ => Ok(()),
        }
    }

    /// The YAML the CLI expects: the server with its `/api` suffix, then the token.
    pub fn render(server: &str, token: &str) -> Result<String, serde_yaml::Error> {
        serde_yaml::to_string(&Self {
            server_url: format!("{}/api", server.trim().trim_end_matches('/')),
            token: token.trim().to_string(),
        })
    }

    /// Parses the YAML the CLI writes, dropping the `/api` suffix it may store; an empty token or server counts as signed out.
    pub fn parse(text: &str) -> Option<Self> {
        let mut credentials: Self = serde_yaml::from_str(text).ok()?;
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
mod tests {
    use super::*;

    #[test]
    fn reads_the_cli_config_and_trims_the_trailing_slash() {
        let parsed =
            Credentials::parse("server_url: https://sablier.facile.studio/\ntoken: abc\n").unwrap();
        assert_eq!(parsed.server_url, "https://sablier.facile.studio");
        assert_eq!(parsed.token, "abc");
    }

    #[test]
    fn strips_the_api_suffix_the_cli_stores() {
        let parsed =
            Credentials::parse("server_url: https://sablier.facile.studio/api/\ntoken: abc\n")
                .unwrap();
        assert_eq!(parsed.server_url, "https://sablier.facile.studio");
    }

    #[test]
    fn rendered_credentials_read_back_as_the_cli_writes_them() {
        let text = Credentials::render("https://sablier.facile.studio/", "abc").unwrap();
        assert!(text.contains("server_url: https://sablier.facile.studio/api"));
        let parsed = Credentials::parse(&text).unwrap();
        assert_eq!(parsed.server_url, "https://sablier.facile.studio");
        assert_eq!(parsed.token, "abc");
    }

    #[test]
    fn a_signed_out_config_is_no_credentials() {
        assert!(Credentials::parse("server_url: https://x.test\n").is_none());
        assert!(Credentials::parse("server_url: https://x.test\ntoken: ''\n").is_none());
        assert!(Credentials::parse("not yaml: [").is_none());
    }
}
