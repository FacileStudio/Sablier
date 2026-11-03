use chrono::{DateTime, Duration, Utc};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::path::PathBuf;

/// The shortest gap between two heartbeats for the same entity.
pub const MIN_INTERVAL: i64 = 120;
/// The most heartbeats one POST may carry.
pub const BATCH: usize = 500;
/// The queue keeps at most this many heartbeats; the oldest go first.
pub const MAX_QUEUED: usize = 5000;
/// The server refuses heartbeats older than seven days; drop them a little before.
pub const MAX_AGE: i64 = 7 * 24 * 3600 - 3600;

/// What the desktop reports: when, which app, and the project hint. Never a title or path.
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Heartbeat {
    pub at: DateTime<Utc>,
    pub source: String,
    pub entity: String,
    pub project: String,
}

impl Heartbeat {
    /// A desktop heartbeat for the focused app.
    pub fn desktop(at: DateTime<Utc>, app: &str, project: &str) -> Self {
        Self {
            at,
            source: "desktop".into(),
            entity: app.into(),
            project: project.into(),
        }
    }
}

/// Lets through at most one heartbeat per entity and project every two minutes.
#[derive(Debug, Default)]
pub struct Throttle {
    last: HashMap<(String, String), DateTime<Utc>>,
}

impl Throttle {
    /// True when this heartbeat should be sent, and records it.
    pub fn admit(&mut self, heartbeat: &Heartbeat) -> bool {
        let key = (heartbeat.entity.clone(), heartbeat.project.clone());
        let due = self
            .last
            .get(&key)
            .is_none_or(|last| (heartbeat.at - *last).num_seconds() >= MIN_INTERVAL);
        if due {
            self.last
                .retain(|_, last| (heartbeat.at - *last).num_seconds() < MIN_INTERVAL);
            self.last.insert(key, heartbeat.at);
        }
        due
    }
}

/// Heartbeats waiting for the server, mirrored to a JSON file so they survive restarts.
#[derive(Debug)]
pub struct Queue {
    path: Option<PathBuf>,
    items: Vec<Heartbeat>,
}

impl Queue {
    /// Loads the queue saved at `path`; a missing or broken file starts empty.
    pub fn open(path: Option<PathBuf>) -> Self {
        let items = path
            .as_ref()
            .and_then(|path| std::fs::read(path).ok())
            .and_then(|bytes| serde_json::from_slice(&bytes).ok())
            .unwrap_or_default();
        Self { path, items }
    }

    /// Adds a heartbeat, dropping the oldest when the queue is full.
    pub fn push(&mut self, heartbeat: Heartbeat) {
        self.items.push(heartbeat);
        let excess = self.items.len().saturating_sub(MAX_QUEUED);
        self.items.drain(..excess);
        self.save();
    }

    /// Drops heartbeats the server would refuse as too old.
    pub fn prune(&mut self, now: DateTime<Utc>) {
        let before = self.items.len();
        let oldest = now - Duration::seconds(MAX_AGE);
        self.items.retain(|heartbeat| heartbeat.at >= oldest);
        if self.items.len() != before {
            self.save();
        }
    }

    /// The next batch to send, oldest first.
    pub fn batch(&self) -> &[Heartbeat] {
        &self.items[..self.items.len().min(BATCH)]
    }

    /// Removes the first `count` heartbeats once the server took them.
    pub fn ack(&mut self, count: usize) {
        self.items.drain(..count.min(self.items.len()));
        self.save();
    }

    fn save(&self) {
        let Some(path) = &self.path else { return };
        let write = || -> std::io::Result<()> {
            if let Some(dir) = path.parent() {
                std::fs::create_dir_all(dir)?;
            }
            let temporary = path.with_extension("json.tmp");
            std::fs::write(&temporary, serde_json::to_vec(&self.items)?)?;
            std::fs::rename(&temporary, path)
        };
        if let Err(err) = write() {
            eprintln!("sablier: cannot save the heartbeat queue: {err}");
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn at(seconds: i64) -> DateTime<Utc> {
        DateTime::from_timestamp(1_800_000_000 + seconds, 0).unwrap()
    }

    #[test]
    fn throttles_each_entity_to_one_heartbeat_per_two_minutes() {
        let mut throttle = Throttle::default();
        assert!(throttle.admit(&Heartbeat::desktop(at(0), "Code", "Bulle")));
        assert!(!throttle.admit(&Heartbeat::desktop(at(119), "Code", "Bulle")));
        assert!(throttle.admit(&Heartbeat::desktop(at(30), "Firefox", "")));
        assert!(throttle.admit(&Heartbeat::desktop(at(60), "Code", "Sablier")));
        assert!(throttle.admit(&Heartbeat::desktop(at(120), "Code", "Bulle")));
    }

    #[test]
    fn the_queue_survives_a_restart_and_batches_oldest_first() {
        let path =
            std::env::temp_dir().join(format!("sablier-desktop-{}-queue.json", std::process::id()));
        let _ = std::fs::remove_file(&path);
        let mut queue = Queue::open(Some(path.clone()));
        for second in 0..3 {
            queue.push(Heartbeat::desktop(at(second), "Code", ""));
        }
        queue.ack(1);
        let reopened = Queue::open(Some(path.clone()));
        assert_eq!(
            reopened.batch().iter().map(|h| h.at).collect::<Vec<_>>(),
            vec![at(1), at(2)]
        );
        std::fs::remove_file(&path).unwrap();
    }

    #[test]
    fn the_queue_is_bounded_and_forgets_what_the_server_would_refuse() {
        let mut queue = Queue::open(None);
        for second in 0..(MAX_QUEUED as i64 + 10) {
            queue.push(Heartbeat::desktop(at(second), "Code", ""));
        }
        assert_eq!(queue.items.len(), MAX_QUEUED);
        assert_eq!(queue.batch().len(), BATCH);
        assert_eq!(queue.batch()[0].at, at(10));
        queue.prune(at(MAX_AGE + 100));
        assert_eq!(queue.batch()[0].at, at(100));
    }
}
