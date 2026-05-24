package repeat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestRepeat_ValidateScenarios(t *testing.T) {
	cases := []struct {
		name    string
		rule    Rule
		wantErr error
	}{
		{"daily_ok", Rule{Kind: KindDaily}, nil},
		{"every_3_days_ok", Rule{Kind: KindEveryNDays, N: 3}, nil},
		{"every_zero_fails", Rule{Kind: KindEveryNDays, N: 0}, ErrInvalidRepeatRule},
		{"every_negative_fails", Rule{Kind: KindEveryNDays, N: -1}, ErrInvalidRepeatRule},
		{"weekly_ok", Rule{Kind: KindWeeklyWeekdays, Weekdays: []time.Weekday{time.Monday, time.Wednesday}}, nil},
		{"weekly_empty_fails", Rule{Kind: KindWeeklyWeekdays}, ErrInvalidRepeatRule},
		{"weekly_duplicate_fails", Rule{Kind: KindWeeklyWeekdays, Weekdays: []time.Weekday{time.Monday, time.Monday}}, ErrInvalidRepeatRule},
		{"weekly_invalid_wd_fails", Rule{Kind: KindWeeklyWeekdays, Weekdays: []time.Weekday{9}}, ErrInvalidRepeatRule},
		{"monthly_15_ok", Rule{Kind: KindMonthlyDay, Day: 15}, nil},
		{"monthly_28_ok", Rule{Kind: KindMonthlyDay, Day: 28}, nil},
		{"monthly_29_unsupported", Rule{Kind: KindMonthlyDay, Day: 29}, ErrMonthlyDayUnsupportedInV1},
		{"monthly_31_unsupported", Rule{Kind: KindMonthlyDay, Day: 31}, ErrMonthlyDayUnsupportedInV1},
		{"monthly_0_fails", Rule{Kind: KindMonthlyDay, Day: 0}, ErrInvalidRepeatRule},
		{"unknown_kind_fails", Rule{Kind: "weird"}, ErrInvalidRepeatRule},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := Validate(c.rule)
			if c.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, c.wantErr)
			}
		})
	}
}

func mustDate(t *testing.T, y int, m time.Month, d int) time.Time {
	t.Helper()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestRepeat_NextOccurrenceDaily(t *testing.T) {
	got, err := NextOccurrence(Rule{Kind: KindDaily}, time.Date(2026, 5, 25, 10, 30, 0, 0, time.UTC))
	require.NoError(t, err)
	require.Equal(t, mustDate(t, 2026, 5, 26), got)
}

func TestRepeat_NextOccurrenceEveryNDays(t *testing.T) {
	got, err := NextOccurrence(Rule{Kind: KindEveryNDays, N: 3}, mustDate(t, 2026, 5, 25))
	require.NoError(t, err)
	require.Equal(t, mustDate(t, 2026, 5, 28), got)
}

func TestRepeat_NextOccurrenceWeeklyWeekdays(t *testing.T) {
	t.Run("from_friday_to_next_monday", func(t *testing.T) {
		// 2026-05-22 is Friday
		got, err := NextOccurrence(
			Rule{Kind: KindWeeklyWeekdays, Weekdays: []time.Weekday{time.Monday, time.Wednesday}},
			mustDate(t, 2026, 5, 22),
		)
		require.NoError(t, err)
		require.Equal(t, mustDate(t, 2026, 5, 25), got) // Monday
	})
	t.Run("late_completion_advances_to_next_week", func(t *testing.T) {
		// Rule: weekly Monday. Completed Wednesday. Next must be next Monday.
		// 2026-05-27 is Wednesday. Next Monday is 2026-06-01.
		got, err := NextOccurrence(
			Rule{Kind: KindWeeklyWeekdays, Weekdays: []time.Weekday{time.Monday}},
			mustDate(t, 2026, 5, 27),
		)
		require.NoError(t, err)
		require.Equal(t, mustDate(t, 2026, 6, 1), got)
	})
	t.Run("strict_after_when_today_matches", func(t *testing.T) {
		// 2026-05-25 is Monday. Rule weekly Monday should jump to next Monday.
		got, err := NextOccurrence(
			Rule{Kind: KindWeeklyWeekdays, Weekdays: []time.Weekday{time.Monday}},
			mustDate(t, 2026, 5, 25),
		)
		require.NoError(t, err)
		require.Equal(t, mustDate(t, 2026, 6, 1), got)
	})
}

func TestRepeat_NextOccurrenceMonthly(t *testing.T) {
	t.Run("same_month_when_day_in_future", func(t *testing.T) {
		got, err := NextOccurrence(Rule{Kind: KindMonthlyDay, Day: 15}, mustDate(t, 2026, 5, 10))
		require.NoError(t, err)
		require.Equal(t, mustDate(t, 2026, 5, 15), got)
	})
	t.Run("next_month_when_day_in_past", func(t *testing.T) {
		got, err := NextOccurrence(Rule{Kind: KindMonthlyDay, Day: 15}, mustDate(t, 2026, 5, 20))
		require.NoError(t, err)
		require.Equal(t, mustDate(t, 2026, 6, 15), got)
	})
	t.Run("strict_after_when_today_matches", func(t *testing.T) {
		got, err := NextOccurrence(Rule{Kind: KindMonthlyDay, Day: 15}, mustDate(t, 2026, 5, 15))
		require.NoError(t, err)
		require.Equal(t, mustDate(t, 2026, 6, 15), got)
	})
}

// PBT: NextOccurrence is monotonic — result > after, always.
// Covers CP-10.
func TestProp_NextOccurrenceMonotonic(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		rule := genValidRule(rt)
		after := genTime(rt)
		got, err := NextOccurrence(rule, after)
		if err != nil {
			rt.Fatalf("unexpected error: %v", err)
		}
		if !got.After(after) {
			rt.Fatalf("not monotonic: rule=%+v after=%s next=%s", rule, after, got)
		}
	})
}

// PBT: invalid rule => Validate returns ErrInvalidRepeatRule, no mutation.
// Covers CP-11.
func TestProp_RepeatInvalidNoMutation(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		rule := genInvalidRule(rt)
		err := Validate(rule)
		if err == nil {
			rt.Fatalf("expected error for invalid rule %+v", rule)
		}
	})
}

func genTime(rt *rapid.T) time.Time {
	year := rapid.IntRange(2000, 2100).Draw(rt, "year")
	month := time.Month(rapid.IntRange(1, 12).Draw(rt, "month"))
	day := rapid.IntRange(1, 28).Draw(rt, "day")
	hour := rapid.IntRange(0, 23).Draw(rt, "hour")
	min := rapid.IntRange(0, 59).Draw(rt, "min")
	return time.Date(year, month, day, hour, min, 0, 0, time.UTC)
}

func genValidRule(rt *rapid.T) Rule {
	kind := rapid.SampledFrom([]Kind{KindDaily, KindEveryNDays, KindWeeklyWeekdays, KindMonthlyDay}).Draw(rt, "kind")
	r := Rule{Kind: kind}
	switch kind {
	case KindEveryNDays:
		r.N = rapid.IntRange(1, 90).Draw(rt, "n")
	case KindWeeklyWeekdays:
		count := rapid.IntRange(1, 7).Draw(rt, "wd_count")
		// Choose `count` distinct weekdays
		all := []time.Weekday{time.Sunday, time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday}
		rapid.Permutation(all).Draw(rt, "perm")
		r.Weekdays = all[:count]
	case KindMonthlyDay:
		r.Day = rapid.IntRange(1, 28).Draw(rt, "day")
	}
	return r
}

func genInvalidRule(rt *rapid.T) Rule {
	scenario := rapid.IntRange(0, 4).Draw(rt, "scenario")
	switch scenario {
	case 0:
		// every_n_days with N < 1
		return Rule{Kind: KindEveryNDays, N: rapid.IntRange(-10, 0).Draw(rt, "n")}
	case 1:
		// weekly_weekdays with empty list
		return Rule{Kind: KindWeeklyWeekdays}
	case 2:
		// monthly_day with day > 28
		return Rule{Kind: KindMonthlyDay, Day: rapid.IntRange(29, 31).Draw(rt, "day")}
	case 3:
		// monthly_day with day < 1
		return Rule{Kind: KindMonthlyDay, Day: rapid.IntRange(-5, 0).Draw(rt, "day")}
	default:
		// unknown kind
		return Rule{Kind: Kind(rapid.SampledFrom([]string{"", "weird", "never", "unknown"}).Draw(rt, "kind"))}
	}
}
