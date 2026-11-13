// Package tiny_quarter_date maps calendar dates to fiscal quarters for an
// arbitrary fiscal-year start month, and computes the boundaries of each
// quarter.
//
// The zero value of Quarter is not meaningful; use NewQuarter to construct
// one. A Quarter is identified by its fiscal year and its 1-based quarter
// number within that fiscal year.
package tiny_quarter_date

import (
	"errors"
	"time"
)

// ErrInvalidStartMonth is returned when a fiscal-year start month is outside
// the range 1..12.
var ErrInvalidStartMonth = errors.New("tiny_quarter_date: start month must be in 1..12")

// Quarter is a fiscal quarter. FiscalYear is the calendar year in which the
// fiscal year begins. Number is 1..4, where 1 is the quarter that contains the
// fiscal-year start month.
type Quarter struct {
	FiscalYear int
	Number     int
}

// Calendar returns the quarter as a calendar-year quarter: the calendar year
// in which the quarter starts, and the calendar month (1..12) in which it
// starts. This is useful when the fiscal year does not align with the calendar
// year.
func (q Quarter) Calendar(startMonth int) (year, month int, err error) {
	if err := validateStartMonth(startMonth); err != nil {
		return 0, 0, err
	}
	// The first quarter of a fiscal year starts in the start month of the
	// fiscal year. Each subsequent quarter is three months later.
	offset := (q.Number - 1) * 3
	monthIndex := (startMonth - 1) + offset // 0-based month index
	year = q.FiscalYear + monthIndex/12
	month = monthIndex%12 + 1
	return year, month, nil
}

// Start returns the first day of the quarter.
func (q Quarter) Start(startMonth int) (time.Time, error) {
	year, month, err := q.Calendar(startMonth)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC), nil
}

// End returns the last day of the quarter, at 23:59:59.999999999 UTC.
func (q Quarter) End(startMonth int) (time.Time, error) {
	start, err := q.Start(startMonth)
	if err != nil {
		return time.Time{}, err
	}
	// Adding three months lands on the first day of the next quarter; the
	// last instant of this quarter is one nanosecond before that.
	return start.AddDate(0, 3, 0).Add(-time.Nanosecond), nil
}

// Contains reports whether t falls within the quarter. The comparison is
// inclusive of both endpoints. t is compared in UTC; if t carries a different
// location it is converted first.
func (q Quarter) Contains(t time.Time, startMonth int) (bool, error) {
	start, err := q.Start(startMonth)
	if err != nil {
		return false, err
	}
	end, err := q.End(startMonth)
	if err != nil {
		return false, err
	}
	u := t.UTC()
	return !u.Before(start) && !u.After(end), nil
}

// NewQuarter returns the fiscal quarter that contains t for the given fiscal
// start month. The start month is the calendar month (1..12) in which the
// fiscal year begins; it must be in 1..12.
func NewQuarter(t time.Time, startMonth int) (Quarter, error) {
	if err := validateStartMonth(startMonth); err != nil {
		return Quarter{}, err
	}
	u := t.UTC()
	y := u.Year()
	m := int(u.Month())

	// Express the month relative to the start month on a 0..11 scale, where 0
	// is the start month. This makes the quarter number a simple division.
	rel := (m - startMonth + 12) % 12
	quarterIndex := rel / 3 // 0..3

	// The fiscal year is the calendar year of the start month. If the
	// calendar month is before the start month, the fiscal year began in the
	// previous calendar year.
	fiscalYear := y
	if m < startMonth {
		fiscalYear = y - 1
	}

	return Quarter{FiscalYear: fiscalYear, Number: quarterIndex + 1}, nil
}

// StartMonth returns the start month of the fiscal year that contains t.
// It is a convenience for callers that need the resolved start month rather
// than a Quarter.
func StartMonth(t time.Time, startMonth int) (int, error) {
	if err := validateStartMonth(startMonth); err != nil {
		return 0, err
	}
	return startMonth, nil
}

func validateStartMonth(startMonth int) error {
	if startMonth < 1 || startMonth > 12 {
		return ErrInvalidStartMonth
	}
	return nil
}
