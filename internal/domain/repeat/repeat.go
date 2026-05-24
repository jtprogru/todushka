package repeat

import (
	"errors"
	"fmt"
	"time"
)

type Kind string

const (
	KindDaily          Kind = "daily"
	KindEveryNDays     Kind = "every_n_days"
	KindWeeklyWeekdays Kind = "weekly_weekdays"
	KindMonthlyDay     Kind = "monthly_day"
)

type Rule struct {
	Kind     Kind           `json:"kind"`
	N        int            `json:"n,omitempty"`
	Weekdays []time.Weekday `json:"weekdays,omitempty"`
	Day      int            `json:"day,omitempty"`
}

var (
	ErrInvalidRepeatRule         = errors.New("repeat: invalid rule")
	ErrMonthlyDayUnsupportedInV1 = errors.New("repeat: monthly day-of-month > 28 will be supported in v2")
)

// Validate enforces structural invariants on a Rule per REQ-7.1, 7.4, 7.5.
func Validate(r Rule) error {
	switch r.Kind {
	case KindDaily:
		return nil
	case KindEveryNDays:
		if r.N < 1 {
			return fmt.Errorf("%w: n must be >= 1", ErrInvalidRepeatRule)
		}
		return nil
	case KindWeeklyWeekdays:
		if len(r.Weekdays) == 0 {
			return fmt.Errorf("%w: weekdays must be non-empty", ErrInvalidRepeatRule)
		}
		seen := make(map[time.Weekday]struct{}, len(r.Weekdays))
		for _, wd := range r.Weekdays {
			if wd < time.Sunday || wd > time.Saturday {
				return fmt.Errorf("%w: invalid weekday %d", ErrInvalidRepeatRule, wd)
			}
			if _, dup := seen[wd]; dup {
				return fmt.Errorf("%w: duplicate weekday %d", ErrInvalidRepeatRule, wd)
			}
			seen[wd] = struct{}{}
		}
		return nil
	case KindMonthlyDay:
		if r.Day < 1 {
			return fmt.Errorf("%w: day must be >= 1", ErrInvalidRepeatRule)
		}
		if r.Day > 28 {
			return ErrMonthlyDayUnsupportedInV1
		}
		return nil
	default:
		return fmt.Errorf("%w: unknown kind %q", ErrInvalidRepeatRule, r.Kind)
	}
}

// NextOccurrence computes the next firing time strictly after `after` per
// the rule. Result is normalized to local midnight (00:00:00) of the matched
// calendar day.
func NextOccurrence(r Rule, after time.Time) (time.Time, error) {
	if err := Validate(r); err != nil {
		return time.Time{}, err
	}
	day := startOfDay(after)
	switch r.Kind {
	case KindDaily:
		return day.AddDate(0, 0, 1), nil
	case KindEveryNDays:
		return day.AddDate(0, 0, r.N), nil
	case KindWeeklyWeekdays:
		wd := make(map[time.Weekday]struct{}, len(r.Weekdays))
		for _, w := range r.Weekdays {
			wd[w] = struct{}{}
		}
		for i := 1; i <= 7; i++ {
			cand := day.AddDate(0, 0, i)
			if _, ok := wd[cand.Weekday()]; ok {
				return cand, nil
			}
		}
		// Unreachable given Validate ensures non-empty weekdays.
		return time.Time{}, fmt.Errorf("%w: no weekday match", ErrInvalidRepeatRule)
	case KindMonthlyDay:
		cand := time.Date(day.Year(), day.Month(), r.Day, 0, 0, 0, 0, day.Location())
		if !cand.After(after) {
			cand = time.Date(day.Year(), day.Month()+1, r.Day, 0, 0, 0, 0, day.Location())
		}
		return cand, nil
	default:
		return time.Time{}, ErrInvalidRepeatRule
	}
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
