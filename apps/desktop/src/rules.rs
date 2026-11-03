use regex::{Regex, RegexBuilder};
use serde::Deserialize;
use std::path::{Path, PathBuf};
use std::time::SystemTime;

/// What the desktop knows about the focused window. It never leaves the device.
#[derive(Debug, Clone, PartialEq)]
pub struct Window {
    pub app: String,
    pub title: Option<String>,
    pub path: Option<String>,
}

#[derive(Deserialize)]
struct RawRule {
    app: Option<String>,
    title: Option<String>,
    path: Option<String>,
    project: String,
    task: Option<String>,
    #[serde(default)]
    auto: bool,
}

/// One line of `rules.yml`: window patterns and the project they mean.
#[derive(Debug)]
pub struct Rule {
    app: Option<Regex>,
    title: Option<Regex>,
    path: Option<Regex>,
    pub project: String,
    pub task: Option<String>,
    pub auto: bool,
}

impl Rule {
    fn matches(&self, window: &Window) -> bool {
        field_matches(&self.app, Some(&window.app))
            && field_matches(&self.title, window.title.as_ref())
            && field_matches(&self.path, window.path.as_ref())
    }
}

fn field_matches(pattern: &Option<Regex>, value: Option<&String>) -> bool {
    match (pattern, value) {
        (None, _) => true,
        (Some(pattern), Some(value)) => pattern.is_match(value),
        (Some(_), None) => false,
    }
}

fn compile(pattern: Option<String>, index: usize) -> Result<Option<Regex>, String> {
    pattern
        .map(|source| {
            RegexBuilder::new(&source)
                .case_insensitive(true)
                .build()
                .map_err(|err| format!("rule {}: {err}", index + 1))
        })
        .transpose()
}

/// The ordered rule list; the first rule that matches a window wins.
#[derive(Debug, Default)]
pub struct Rules(Vec<Rule>);

impl Rules {
    /// Parses a YAML list of rules. Patterns are case-insensitive regexes.
    pub fn parse(text: &str) -> Result<Self, String> {
        if text.trim().is_empty() {
            return Ok(Self::default());
        }
        let raw: Vec<RawRule> = serde_yaml::from_str(text).map_err(|err| err.to_string())?;
        raw.into_iter()
            .enumerate()
            .map(|(index, rule)| {
                if rule.app.is_none() && rule.title.is_none() && rule.path.is_none() {
                    return Err(format!("rule {} needs app, title or path", index + 1));
                }
                if rule.project.trim().is_empty() {
                    return Err(format!("rule {} needs a project", index + 1));
                }
                Ok(Rule {
                    app: compile(rule.app, index)?,
                    title: compile(rule.title, index)?,
                    path: compile(rule.path, index)?,
                    project: rule.project,
                    task: rule.task,
                    auto: rule.auto,
                })
            })
            .collect::<Result<Vec<_>, _>>()
            .map(Self)
    }

    /// Returns the first rule that matches the window.
    pub fn matching(&self, window: &Window) -> Option<&Rule> {
        self.0.iter().find(|rule| rule.matches(window))
    }
}

/// The rules file, re-read whenever its modification time changes.
pub struct RulesFile {
    path: Option<PathBuf>,
    modified: Option<SystemTime>,
    rules: Rules,
}

impl RulesFile {
    /// Watches the file at `path`; a missing file means no rules.
    pub fn new(path: Option<PathBuf>) -> Self {
        Self {
            path,
            modified: None,
            rules: Rules::default(),
        }
    }

    /// Returns the current rules, reloading them if the file changed on disk.
    pub fn current(&mut self) -> &Rules {
        let modified = self.path.as_deref().and_then(modified_at);
        if modified != self.modified {
            self.modified = modified;
            self.rules = self.path.as_deref().map(load).unwrap_or_default();
        }
        &self.rules
    }
}

fn modified_at(path: &Path) -> Option<SystemTime> {
    std::fs::metadata(path)
        .and_then(|meta| meta.modified())
        .ok()
}

fn load(path: &Path) -> Rules {
    let text = std::fs::read_to_string(path).unwrap_or_default();
    Rules::parse(&text).unwrap_or_else(|err| {
        eprintln!("sablier: ignoring {}: {err}", path.display());
        Rules::default()
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    fn window(app: &str, title: Option<&str>) -> Window {
        Window {
            app: app.into(),
            title: title.map(Into::into),
            path: Some("/usr/bin/code".into()),
        }
    }

    const RULES: &str = r#"
- app: "^code$"
  title: "bulle"
  project: Bulle
  task: Desktop
- app: "firefox"
  project: Research
  auto: true
- path: "/usr/bin/"
  project: Fallback
"#;

    #[test]
    fn the_first_matching_rule_wins() {
        let rules = Rules::parse(RULES).unwrap();
        let hit = rules
            .matching(&window("Code", Some("main.rs - Bulle")))
            .unwrap();
        assert_eq!(
            (hit.project.as_str(), hit.task.as_deref()),
            ("Bulle", Some("Desktop"))
        );
        assert_eq!(
            rules
                .matching(&window("Code", Some("Sablier")))
                .unwrap()
                .project,
            "Fallback"
        );
        assert!(rules.matching(&window("firefox", None)).unwrap().auto);
    }

    #[test]
    fn a_title_pattern_never_matches_a_window_without_title() {
        let rules = Rules::parse("- title: x\n  project: P\n").unwrap();
        assert!(rules.matching(&window("Code", None)).is_none());
    }

    #[test]
    fn rejects_rules_that_would_match_everything_or_do_not_compile() {
        assert!(Rules::parse("- project: P\n").is_err());
        assert!(Rules::parse("- app: '('\n  project: P\n").is_err());
        assert!(
            Rules::parse("")
                .unwrap()
                .matching(&window("x", None))
                .is_none()
        );
    }
}
