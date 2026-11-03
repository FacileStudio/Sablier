use crate::config::Credentials;
use crate::heartbeat::Heartbeat;
use chrono::{DateTime, Utc};
use serde::de::DeserializeOwned;
use serde::{Deserialize, Serialize};
use std::time::Duration;

/// A Sablier time entry, as `/api/time-entries` returns it.
#[derive(Debug, Clone, PartialEq, Deserialize)]
pub struct Entry {
    pub id: i64,
    pub project_id: i64,
    pub task_id: i64,
    #[serde(default)]
    pub task_name: String,
    pub started_at: DateTime<Utc>,
    pub paused_at: Option<DateTime<Utc>>,
    #[serde(default)]
    pub paused_duration_ms: i64,
}

impl Entry {
    /// Seconds of work so far: wall time minus every pause, the current one included.
    pub fn elapsed(&self, now: DateTime<Utc>) -> i64 {
        let current_pause = self
            .paused_at
            .map_or(0, |paused| (now - paused).num_milliseconds());
        let worked =
            (now - self.started_at).num_milliseconds() - self.paused_duration_ms - current_pause;
        worked.max(0) / 1000
    }
}

/// A project the user can track time on.
#[derive(Debug, Clone, Deserialize)]
pub struct Project {
    pub id: i64,
    pub name: String,
    pub updated_at: DateTime<Utc>,
}

/// A task inside a project.
#[derive(Debug, Clone, Deserialize)]
pub struct Task {
    pub id: i64,
    pub name: String,
    #[serde(default)]
    pub status: String,
}

#[derive(Deserialize)]
struct RunningResponse {
    entry: Option<Entry>,
}

#[derive(Deserialize)]
struct ProjectsResponse {
    projects: Vec<Project>,
}

#[derive(Deserialize)]
struct TasksResponse {
    tasks: Vec<Task>,
}

#[derive(Serialize)]
struct EntryUpdate {
    project_id: i64,
    task_id: i64,
    started_at: DateTime<Utc>,
    stopped_at: DateTime<Utc>,
}

#[derive(Serialize)]
struct HeartbeatBatch<'a> {
    heartbeats: &'a [Heartbeat],
}

/// Why a heartbeat batch did not go through.
#[derive(Debug, PartialEq)]
pub enum SendError {
    Retry,
    Drop,
}

/// A blocking client for the Sablier API, authenticated with the CLI's token.
pub struct Api {
    agent: ureq::Agent,
    base: String,
    authorization: String,
}

impl Api {
    /// A client for the server and token in `credentials`.
    pub fn new(credentials: &Credentials) -> Self {
        let agent = ureq::Agent::config_builder()
            .timeout_global(Some(Duration::from_secs(15)))
            .user_agent(concat!("sablier-desktop/", env!("CARGO_PKG_VERSION")))
            .build()
            .into();
        Self {
            agent,
            base: format!("{}/api", credentials.server_url),
            authorization: format!("Bearer {}", credentials.token),
        }
    }

    fn get<T: DeserializeOwned>(&self, path: &str) -> Result<T, String> {
        self.agent
            .get(format!("{}{path}", self.base))
            .header("Authorization", &self.authorization)
            .call()
            .and_then(|mut response| response.body_mut().read_json())
            .map_err(|err| format!("GET {path}: {err}"))
    }

    fn post(&self, path: &str) -> Result<(), String> {
        self.agent
            .post(format!("{}{path}", self.base))
            .header("Authorization", &self.authorization)
            .send_empty()
            .map(drop)
            .map_err(|err| format!("POST {path}: {err}"))
    }

    fn send<T: Serialize>(&self, method: &str, path: &str, body: &T) -> Result<(), ureq::Error> {
        let url = format!("{}{path}", self.base);
        let request = match method {
            "PUT" => self.agent.put(url),
            _ => self.agent.post(url),
        };
        request
            .header("Authorization", &self.authorization)
            .send_json(body)
            .map(drop)
    }

    /// The caller's running entry, if any.
    pub fn running(&self) -> Result<Option<Entry>, String> {
        self.get::<RunningResponse>("/time-entries/running")
            .map(|response| response.entry)
    }

    /// Every project the caller can see.
    pub fn projects(&self) -> Result<Vec<Project>, String> {
        self.get::<ProjectsResponse>("/projects/")
            .map(|response| response.projects)
    }

    /// The tasks of one project.
    pub fn tasks(&self, project_id: i64) -> Result<Vec<Task>, String> {
        self.get::<TasksResponse>(&format!("/projects/{project_id}/tasks"))
            .map(|response| response.tasks)
    }

    /// Starts a timer on a task.
    pub fn start(&self, project_id: i64, task_id: i64) -> Result<(), String> {
        let body = serde_json::json!({ "project_id": project_id, "task_id": task_id });
        self.send("POST", "/time-entries/start", &body)
            .map_err(|err| format!("POST /time-entries/start: {err}"))
    }

    /// Stops, pauses or resumes the running timer.
    pub fn timer(&self, action: &str) -> Result<(), String> {
        self.post(&format!("/time-entries/{action}"))
    }

    /// Ends `entry` at `at`, or deletes it when it started after that.
    pub fn end_at(&self, entry: &Entry, at: DateTime<Utc>) -> Result<(), String> {
        let path = format!("/time-entries/{}", entry.id);
        if at <= entry.started_at {
            return self
                .agent
                .delete(format!("{}{path}", self.base))
                .header("Authorization", &self.authorization)
                .call()
                .map(drop)
                .map_err(|err| format!("DELETE {path}: {err}"));
        }
        let body = EntryUpdate {
            project_id: entry.project_id,
            task_id: entry.task_id,
            started_at: entry.started_at,
            stopped_at: at,
        };
        self.send("PUT", &path, &body)
            .map_err(|err| format!("PUT {path}: {err}"))
    }

    /// Posts a batch of heartbeats. A 400 means the batch can never succeed.
    pub fn heartbeats(&self, batch: &[Heartbeat]) -> Result<(), SendError> {
        match self.send("POST", "/heartbeats", &HeartbeatBatch { heartbeats: batch }) {
            Ok(()) => Ok(()),
            Err(ureq::Error::StatusCode(400 | 413 | 422)) => Err(SendError::Drop),
            Err(_) => Err(SendError::Retry),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn elapsed_time_leaves_out_past_and_current_pauses() {
        let entry: Entry = serde_json::from_str(
            r#"{"id":1,"project_id":2,"task_id":3,"started_at":"2026-11-03T10:00:00Z",
                "stopped_at":null,"paused_at":"2026-11-03T11:00:00Z","paused_duration_ms":600000}"#,
        )
        .unwrap();
        let now = DateTime::parse_from_rfc3339("2026-11-03T11:30:00Z")
            .unwrap()
            .to_utc();
        assert_eq!(entry.elapsed(now), 50 * 60);
    }

    #[test]
    fn heartbeats_serialize_to_the_contract_shape() {
        let at = DateTime::parse_from_rfc3339("2026-11-03T10:00:00Z")
            .unwrap()
            .to_utc();
        let batch = [Heartbeat::desktop(at, "Code", "Bulle")];
        let json = serde_json::to_value(HeartbeatBatch { heartbeats: &batch }).unwrap();
        assert_eq!(
            json,
            serde_json::json!({"heartbeats":[{"at":"2026-11-03T10:00:00Z","source":"desktop","entity":"Code","project":"Bulle"}]})
        );
    }
}
