use serde::Deserialize;
use std::path::PathBuf;

/// The server and API token `sablier login` stores in `~/.sablier.yml`.
#[derive(Debug, Clone, Deserialize, PartialEq)]
pub struct Credentials {
    pub server_url: String,
    #[serde(default)]
    pub token: String,
}

impl Credentials {
    /// Reads the CLI's credentials, or `None` when the user has not signed in.
    pub fn load() -> Option<Self> {
        let path = dirs::home_dir()?.join(".sablier.yml");
        Self::parse(&std::fs::read_to_string(path).ok()?)
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
    fn a_signed_out_config_is_no_credentials() {
        assert!(Credentials::parse("server_url: https://x.test\n").is_none());
        assert!(Credentials::parse("server_url: https://x.test\ntoken: ''\n").is_none());
        assert!(Credentials::parse("not yaml: [").is_none());
    }
}
