use super::Credentials;
use crate::token::Vault;
use std::cell::RefCell;
use std::path::PathBuf;

struct FakeVault {
    available: bool,
    keeps: bool,
    token: RefCell<Option<String>>,
}

fn vault(available: bool, keeps: bool) -> FakeVault {
    FakeVault {
        available,
        keeps,
        token: RefCell::new(None),
    }
}

impl Vault for FakeVault {
    fn get(&self) -> Result<Option<String>, String> {
        if !self.available {
            return Err("no secret service".into());
        }
        Ok(self.token.borrow().clone())
    }

    fn set(&self, token: &str) -> Result<(), String> {
        if !self.available {
            return Err("no secret service".into());
        }
        if self.keeps {
            *self.token.borrow_mut() = Some(token.to_string());
        }
        Ok(())
    }

    fn delete(&self) -> Result<(), String> {
        if !self.available {
            return Err("no secret service".into());
        }
        self.token.borrow_mut().take();
        Ok(())
    }
}

const SIGNED_IN: &str = "server_url: https://sablier.test/api\ntoken: secret\n";

fn scratch(name: &str, contents: &str) -> PathBuf {
    let path =
        std::env::temp_dir().join(format!("sablier-desktop-{}-{name}.yml", std::process::id()));
    std::fs::write(&path, contents).unwrap();
    path
}

#[test]
fn a_file_token_moves_to_the_keyring_and_leaves_the_file() {
    let path = scratch("migrate", SIGNED_IN);
    let keyring = vault(true, true);

    let loaded = Credentials::load_from(&path, &keyring).unwrap();

    assert_eq!(loaded.server_url, "https://sablier.test");
    assert_eq!(loaded.token, "secret");
    assert_eq!(keyring.token.borrow().as_deref(), Some("secret"));
    assert_eq!(
        std::fs::read_to_string(&path).unwrap(),
        "server_url: https://sablier.test/api\n"
    );
    assert_eq!(Credentials::load_from(&path, &keyring), Some(loaded));
    std::fs::remove_file(&path).unwrap();
}

#[test]
fn without_a_keyring_the_token_stays_in_the_file() {
    for (name, keyring) in [
        ("absent", vault(false, false)),
        ("dropped", vault(true, false)),
    ] {
        let path = scratch(name, SIGNED_IN);

        assert_eq!(
            Credentials::load_from(&path, &keyring).unwrap().token,
            "secret"
        );
        assert_eq!(std::fs::read_to_string(&path).unwrap(), SIGNED_IN);
        std::fs::remove_file(&path).unwrap();
    }
}

#[test]
fn sign_in_writes_the_token_to_the_keyring_or_else_the_file() {
    let path = scratch("save", "");

    Credentials::save_to(&path, &vault(true, true), "https://sablier.test/", "secret").unwrap();
    assert_eq!(
        std::fs::read_to_string(&path).unwrap(),
        "server_url: https://sablier.test/api\n"
    );

    Credentials::save_to(
        &path,
        &vault(false, false),
        "https://sablier.test",
        "secret",
    )
    .unwrap();
    assert_eq!(std::fs::read_to_string(&path).unwrap(), SIGNED_IN);
    std::fs::remove_file(&path).unwrap();
}

#[test]
fn a_token_only_in_the_keyring_still_signs_in() {
    let path = scratch("keyring-only", "server_url: https://sablier.test/api\n");
    let keyring = vault(true, true);

    assert!(Credentials::load_from(&path, &keyring).is_none());
    keyring.set("secret").unwrap();
    assert_eq!(
        Credentials::load_from(&path, &keyring).unwrap().token,
        "secret"
    );
    std::fs::remove_file(&path).unwrap();
}

#[test]
fn sign_out_clears_the_keyring_and_the_file() {
    let path = scratch("forget", SIGNED_IN);
    let keyring = vault(true, true);
    keyring.set("secret").unwrap();

    Credentials::forget_at(&path, &keyring).unwrap();

    assert!(keyring.token.borrow().is_none());
    assert!(!path.exists());
    Credentials::forget_at(&path, &vault(false, false)).unwrap();
}
