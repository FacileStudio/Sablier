use chrono::{DateTime, Duration, Utc};

/// Seconds without input after which the user counts as away.
pub const DETECT_AFTER: i64 = 180;
/// The shortest absence worth asking about (Toggl's default).
pub const PROMPT_MIN: i64 = 300;
/// Absences this long are a night or a weekend, not a break: never ask.
pub const PROMPT_MAX: i64 = 10 * 3600;

/// A closed interval of time.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct Span {
    pub since: DateTime<Utc>,
    pub until: DateTime<Utc>,
}

impl Span {
    /// Whole seconds between the two ends.
    pub fn seconds(&self) -> i64 {
        (self.until - self.since).num_seconds()
    }
}

/// Turns periodic "seconds since last input" samples into absences.
#[derive(Debug, Default)]
pub struct IdleTracker {
    since: Option<DateTime<Utc>>,
}

impl IdleTracker {
    /// True while the last sample said the user is away.
    pub fn is_idle(&self) -> bool {
        self.since.is_some()
    }

    /// Feeds one sample and returns the absence that just ended, if any.
    pub fn sample(&mut self, now: DateTime<Utc>, idle_seconds: u64) -> Option<Span> {
        let idle = i64::try_from(idle_seconds).unwrap_or(i64::MAX / 2);
        let last_input = now - Duration::seconds(idle);
        if idle >= DETECT_AFTER {
            self.since.get_or_insert(last_input);
            return None;
        }
        self.since.take().map(|since| Span {
            since,
            until: last_input.max(since),
        })
    }
}

/// The part of an absence a running timer counted, when it is worth a prompt.
///
/// `timer_started` is the running, unpaused entry's start; `None` means nothing to discard.
pub fn prompt_span(absence: Span, timer_started: Option<DateTime<Utc>>) -> Option<Span> {
    let since = absence.since.max(timer_started?);
    let span = Span {
        since,
        until: absence.until,
    };
    (PROMPT_MIN..PROMPT_MAX)
        .contains(&span.seconds())
        .then_some(span)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn at(minutes: i64) -> DateTime<Utc> {
        DateTime::from_timestamp(1_800_000_000 + minutes * 60, 0).unwrap()
    }

    #[test]
    fn an_absence_ends_when_input_comes_back() {
        let mut tracker = IdleTracker::default();
        assert_eq!(tracker.sample(at(0), 10), None);
        assert_eq!(tracker.sample(at(4), 200), None);
        assert!(tracker.is_idle());
        assert_eq!(tracker.sample(at(9), 300), None);
        let span = tracker.sample(at(12), 60).unwrap();
        assert_eq!(span.since, at(4) - Duration::seconds(200));
        assert_eq!(span.until, at(11));
        assert!(!tracker.is_idle());
    }

    #[test]
    fn short_breaks_are_not_absences() {
        let mut tracker = IdleTracker::default();
        assert_eq!(tracker.sample(at(1), 120), None);
        assert_eq!(tracker.sample(at(2), 5), None);
    }

    #[test]
    fn prompts_only_for_five_minutes_to_ten_hours_of_a_running_timer() {
        let absence = Span {
            since: at(0),
            until: at(20),
        };
        assert_eq!(prompt_span(absence, None), None);
        assert_eq!(prompt_span(absence, Some(at(-30))), Some(absence));
        assert_eq!(prompt_span(absence, Some(at(10))).unwrap().since, at(10));
        assert_eq!(prompt_span(absence, Some(at(17))), None);
        assert_eq!(
            prompt_span(
                Span {
                    since: at(0),
                    until: at(4)
                },
                Some(at(-1))
            ),
            None
        );
        assert_eq!(
            prompt_span(
                Span {
                    since: at(0),
                    until: at(600)
                },
                Some(at(-1))
            ),
            None
        );
    }
}
