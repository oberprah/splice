use chrono::{DateTime, Utc};

pub fn format_relative_time(date: DateTime<Utc>, now: DateTime<Utc>) -> String {
    let seconds = now.signed_duration_since(date).num_seconds().max(0);

    if seconds < 60 {
        return format!("{seconds}s ago");
    }

    let minutes = seconds / 60;
    if minutes < 60 {
        return format!("{minutes}m ago");
    }

    let hours = minutes / 60;
    if hours < 24 {
        return format!("{hours}h ago");
    }

    let days = hours / 24;
    if days < 30 {
        return format!("{days}d ago");
    }

    let months = days / 30;
    if days < 365 {
        return format!("{months}mo ago");
    }

    let years = days / 365;
    format!("{years}y ago")
}

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::TimeZone;

    #[test]
    fn formats_seconds() {
        let now = Utc.with_ymd_and_hms(2026, 1, 1, 0, 0, 10).unwrap();
        let date = Utc.with_ymd_and_hms(2026, 1, 1, 0, 0, 0).unwrap();
        assert_eq!(format_relative_time(date, now), "10s ago");
    }

    #[test]
    fn formats_minutes() {
        let now = Utc.with_ymd_and_hms(2026, 1, 1, 0, 10, 0).unwrap();
        let date = Utc.with_ymd_and_hms(2026, 1, 1, 0, 0, 0).unwrap();
        assert_eq!(format_relative_time(date, now), "10m ago");
    }

    #[test]
    fn formats_hours() {
        let now = Utc.with_ymd_and_hms(2026, 1, 1, 10, 0, 0).unwrap();
        let date = Utc.with_ymd_and_hms(2026, 1, 1, 0, 0, 0).unwrap();
        assert_eq!(format_relative_time(date, now), "10h ago");
    }

    #[test]
    fn formats_days() {
        let now = Utc.with_ymd_and_hms(2026, 1, 11, 0, 0, 0).unwrap();
        let date = Utc.with_ymd_and_hms(2026, 1, 1, 0, 0, 0).unwrap();
        assert_eq!(format_relative_time(date, now), "10d ago");
    }

    #[test]
    fn formats_months() {
        let now = Utc.with_ymd_and_hms(2026, 5, 1, 0, 0, 0).unwrap();
        let date = Utc.with_ymd_and_hms(2026, 1, 1, 0, 0, 0).unwrap();
        assert_eq!(format_relative_time(date, now), "4mo ago");
    }

    #[test]
    fn formats_years() {
        let now = Utc.with_ymd_and_hms(2026, 1, 1, 0, 0, 0).unwrap();
        let date = Utc.with_ymd_and_hms(2020, 1, 1, 0, 0, 0).unwrap();
        assert_eq!(format_relative_time(date, now), "6y ago");
    }

    #[test]
    fn formats_364_days_as_months() {
        let now = Utc.with_ymd_and_hms(2026, 1, 1, 0, 0, 0).unwrap();
        let date = now - chrono::Duration::days(364);
        assert_eq!(format_relative_time(date, now), "12mo ago");
    }

    #[test]
    fn formats_365_days_as_years() {
        let now = Utc.with_ymd_and_hms(2026, 1, 1, 0, 0, 0).unwrap();
        let date = now - chrono::Duration::days(365);
        assert_eq!(format_relative_time(date, now), "1y ago");
    }

    #[test]
    fn clamps_future_dates_to_zero() {
        let now = Utc.with_ymd_and_hms(2026, 1, 1, 0, 0, 0).unwrap();
        let date = Utc.with_ymd_and_hms(2026, 1, 1, 0, 0, 10).unwrap();
        assert_eq!(format_relative_time(date, now), "0s ago");
    }
}
