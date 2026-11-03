use chrono::{DateTime, Duration, Utc};
use std::collections::HashMap;

/// Active seconds on one project before Sablier offers a timer.
pub const SUGGEST_AFTER: i64 = 300;
/// How long "Not now" keeps a project quiet.
pub const SNOOZE: i64 = 3600;
/// The most a single sample may add, so a suspended laptop does not count as work.
pub const MAX_STEP: i64 = 60;

/// Decides when a rule-matched project has had enough attention to suggest a timer.
#[derive(Debug, Default)]
pub struct Suggester {
    current: Option<String>,
    active: i64,
    last: Option<DateTime<Utc>>,
    snoozed: HashMap<String, DateTime<Utc>>,
}

impl Suggester {
    /// Feeds one window sample and returns the project to suggest, at most once per snooze.
    ///
    /// `project` is the matched rule's project; `None` for a window no rule matched, which
    /// pauses the count without resetting it.
    pub fn observe(
        &mut self,
        now: DateTime<Utc>,
        project: Option<&str>,
        active: bool,
        timer_running: bool,
    ) -> Option<String> {
        let step = self
            .last
            .map_or(0, |last| (now - last).num_seconds().clamp(0, MAX_STEP));
        self.last = Some(now);
        if timer_running {
            self.current = None;
            self.active = 0;
            return None;
        }
        let project = project.filter(|_| active)?;
        if self.current.as_deref() != Some(project) {
            self.current = Some(project.to_string());
            self.active = 0;
            return None;
        }
        self.active += step;
        if self.active < SUGGEST_AFTER || self.is_snoozed(project, now) {
            return None;
        }
        self.active = 0;
        self.snooze(project, now);
        Some(project.to_string())
    }

    /// Keeps `project` from being suggested for the next hour.
    pub fn snooze(&mut self, project: &str, now: DateTime<Utc>) {
        self.snoozed
            .insert(project.to_string(), now + Duration::seconds(SNOOZE));
    }

    fn is_snoozed(&self, project: &str, now: DateTime<Utc>) -> bool {
        self.snoozed.get(project).is_some_and(|until| *until > now)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn at(seconds: i64) -> DateTime<Utc> {
        DateTime::from_timestamp(1_800_000_000 + seconds, 0).unwrap()
    }

    fn feed(
        suggester: &mut Suggester,
        from: i64,
        to: i64,
        project: Option<&str>,
    ) -> Vec<(i64, String)> {
        (from..=to)
            .step_by(15)
            .filter_map(|t| {
                suggester
                    .observe(at(t), project, true, false)
                    .map(|p| (t, p))
            })
            .collect()
    }

    #[test]
    fn suggests_after_five_minutes_on_the_same_project_then_snoozes() {
        let mut suggester = Suggester::default();
        assert_eq!(
            feed(&mut suggester, 0, 3600, Some("Bulle")),
            vec![(300, "Bulle".to_string())]
        );
        assert_eq!(feed(&mut suggester, 3615, 4000, Some("Bulle")).len(), 1);
    }

    #[test]
    fn switching_project_restarts_the_count_and_unmatched_windows_pause_it() {
        let mut suggester = Suggester::default();
        assert!(feed(&mut suggester, 0, 240, Some("A")).is_empty());
        assert!(feed(&mut suggester, 255, 400, None).is_empty());
        assert_eq!(
            feed(&mut suggester, 415, 480, Some("A")),
            vec![(460, "A".to_string())]
        );
        assert!(feed(&mut suggester, 495, 700, Some("B")).is_empty());
    }

    #[test]
    fn a_running_timer_or_idleness_never_suggests() {
        let mut suggester = Suggester::default();
        for t in (0..1000).step_by(15) {
            assert!(suggester.observe(at(t), Some("A"), true, true).is_none());
            assert!(suggester.observe(at(t), Some("B"), false, false).is_none());
        }
    }

    #[test]
    fn not_now_snoozes_a_project_for_an_hour() {
        let mut suggester = Suggester::default();
        suggester.snooze("A", at(0));
        assert!(feed(&mut suggester, 0, 3500, Some("A")).is_empty());
        assert_eq!(feed(&mut suggester, 3515, 3700, Some("A")).len(), 1);
    }
}
