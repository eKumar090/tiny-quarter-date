package tiny_quarter_date

import (
	"errors"
	"testing"
	"time"
)

func mustDate(t *testing.T, year int, month time.Month, day int) time.Time {
	t.Helper()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestNewQuarterCalendarYearStart(t *testing.T) {
	// Start month January: fiscal year == calendar year, quarter == calendar
	// quarter. This is the simplest case and anchors the rest.
	cases := []struct {
		date time.Time
		want Quarter
	}{
		{mustDate(t, 2024, time.January, 1), Quarter{2024, 1}},
		{mustDate(t, 2024, time.March, 31), Quarter{2024, 1}},
		{mustDate(t, 2024, time.April, 1), Quarter{2024, 2}},
		{mustDate(t, 2024, time.June, 30), Quarter{2024, 2}},
		{mustDate(t, 2024, time.July, 1), Quarter{2024, 3}},
		{mustDate(t, 2024, time.September, 30), Quarter{2024, 3}},
		{mustDate(t, 2024, time.October, 1), Quarter{2024, 4}},
		{mustDate(t, 2024, time.December, 31), Quarter{2024, 4}},
	}
	for _, tc := range cases {
		got, err := NewQuarter(tc.date, 1)
		if err != nil {
			t.Fatalf("NewQuarter(%v, 1): %v", tc.date, err)
		}
		if got != tc.want {
			t.Errorf("NewQuarter(%v, 1) = %+v, want %+v", tc.date, got, tc.want)
		}
	}
}

func TestNewQuarterOffsetStart(t *testing.T) {
	// Start month October: Q1 is Oct-Dec of the previous calendar year, Q2 is
	// Jan-Mar, Q3 is Apr-Jun, Q4 is Jul-Sep. This is the case that trips up
	// naive implementations because the fiscal year label is the previous
	// calendar year for Q1.
	cases := []struct {
		date time.Time
		want Quarter
	}{
		{mustDate(t, 2024, time.October, 1), Quarter{2024, 1}},
		{mustDate(t, 2024, time.December, 31), Quarter{2024, 1}},
		{mustDate(t, 2025, time.January, 1), Quarter{2024, 2}},
		{mustDate(t, 2025, time.March, 31), Quarter{2024, 2}},
		{mustDate(t, 2025, time.April, 1), Quarter{2024, 3}},
		{mustDate(t, 2025, time.June, 30), Quarter{2024, 3}},
		{mustDate(t, 2025, time.July, 1), Quarter{2024, 4}},
		{mustDate(t, 2025, time.September, 30), Quarter{2024, 4}},
		{mustDate(t, 2025, time.October, 1), Quarter{2025, 1}},
	}
	for _, tc := range cases {
		got, err := NewQuarter(tc.date, 10)
		if err != nil {
			t.Fatalf("NewQuarter(%v, 10): %v", tc.date, err)
		}
		if got != tc.want {
			t.Errorf("NewQuarter(%v, 10) = %+v, want %+v", tc.date, got, tc.want)
		}
	}
}

func TestNewQuarterInvalidStartMonth(t *testing.T) {
	for _, sm := range []int{0, 13, -1} {
		_, err := NewQuarter(time.Now(), sm)
		if !errors.Is(err, ErrInvalidStartMonth) {
			t.Errorf("NewQuarter(startMonth=%d) error = %v, want ErrInvalidStartMonth", sm, err)
		}
	}
}

func TestQuarterStart(t *testing.T) {
	// Fiscal year starting October 2024.
	q := Quarter{FiscalYear: 2024, Number: 1}
	start, err := q.Start(10)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2024, time.October, 1, 0, 0, 0, 0, time.UTC)
	if !start.Equal(want) {
		t.Errorf("Start = %v, want %v", start, want)
	}

	q = Quarter{FiscalYear: 2024, Number: 2}
	start, err = q.Start(10)
	if err != nil {
		t.Fatal(err)
	}
	want = time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	if !start.Equal(want) {
		t.Errorf("Start = %v, want %v", start, want)
	}
}

func TestQuarterEnd(t *testing.T) {
	// The end of Q1 for a January start is March 31. The last instant is
	// one nanosecond before April 1.
	q := Quarter{FiscalYear: 2024, Number: 1}
	end, err := q.End(1)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2024, time.March, 31, 23, 59, 59, 999999999, time.UTC)
	if !end.Equal(want) {
		t.Errorf("End = %v, want %v", end, want)
	}

	// Leap-year February is inside Q1 for a January start; the end is still
	// March 31, but the day count differs. The instant is what matters.
	q = Quarter{FiscalYear: 2024, Number: 1}
	end, err = q.End(1)
	if err != nil {
		t.Fatal(err)
	}
	if end.Month() != time.March || end.Day() != 31 {
		t.Errorf("End month/day = %v/%d, want March/31", end.Month(), end.Day())
	}
}

func TestQuarterEndOctoberStart(t *testing.T) {
	// Q1 of fiscal 2024 starting October 2024 ends on December 31, 2024.
	q := Quarter{FiscalYear: 2024, Number: 1}
	end, err := q.End(10)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2024, time.December, 31, 23, 59, 59, 999999999, time.UTC)
	if !end.Equal(want) {
		t.Errorf("End = %v, want %v", end, want)
	}
}

func TestQuarterContains(t *testing.T) {
	q := Quarter{FiscalYear: 2024, Number: 2} // Jan-Mar 2025 for October start
	cases := []struct {
		t    time.Time
		want bool
	}{
		{time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC), true},
		{time.Date(2025, time.March, 31, 23, 59, 59, 999999999, time.UTC), true},
		{time.Date(2024, time.December, 31, 23, 59, 59, 999999999, time.UTC), false},
		{time.Date(2025, time.April, 1, 0, 0, 0, 0, time.UTC), false},
	}
	for _, tc := range cases {
		got, err := q.Contains(tc.t, 10)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Errorf("Contains(%v) = %v, want %v", tc.t, got, tc.want)
		}
	}
}

func TestQuarterCalendar(t *testing.T) {
	q := Quarter{FiscalYear: 2024, Number: 3}
	year, month, err := q.Calendar(10)
	if err != nil {
		t.Fatal(err)
	}
	if year != 2025 || month != 4 {
		t.Errorf("Calendar = %d-%02d, want 2025-04", year, month)
	}
}

func TestNewQuarterRoundTrip(t *testing.T) {
	// Every day in a leap year should map back to a quarter that contains it.
	start := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	for d := 0; d < 366; d++ {
		t1 := start.AddDate(0, 0, d)
		q, err := NewQuarter(t1, 4)
		if err != nil {
			t.Fatal(err)
		}
		ok, err := q.Contains(t1, 4)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatalf("quarter %+v does not contain %v", q, t1)
		}
	}
}

func TestStartMonth(t *testing.T) {
	got, err := StartMonth(time.Now(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if got != 7 {
		t.Errorf("StartMonth = %d, want 7", got)
	}
	if _, err := StartMonth(time.Now(), 0); !errors.Is(err, ErrInvalidStartMonth) {
		t.Errorf("StartMonth(0) error = %v, want ErrInvalidStartMonth", err)
	}
}
