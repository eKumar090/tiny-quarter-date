# tiny-quarter-date

Maps calendar dates to fiscal quarters for a configurable fiscal-year start month and computes quarter boundaries.

```go
package main

import (
	"fmt"
	"time"

	tqd "tiny_quarter_date"
)

func main() {
	t := time.Date(2025, time.February, 14, 0, 0, 0, 0, time.UTC)

	// Fiscal year starts in October.
	q, err := tqd.NewQuarter(t, 10)
	if err != nil {
		panic(err)
	}
	fmt.Println(q) // {2024 2}

	start, _ := q.Start(10)
	end, _ := q.End(10)
	fmt.Println(start.Format(time.DateOnly)) // 2025-01-01
	fmt.Println(end.Format(time.DateOnly))   // 2025-03-31
}
```

## Why this exists

Most date libraries assume the fiscal year starts in January and the calendar year and fiscal year are the same. That is wrong for a large number of organisations: a UK-style tax year, an October-start retail year, a July-start university year. The only thing that varies is the start month, but getting the fiscal-year label right is the part people get wrong, because for a start month other than January the first quarter belongs to the *previous* calendar year.

The trade-off is scope. This library does only one thing: given a start month in 1..12, it tells you which fiscal quarter a date falls in and where that quarter begins and ends. It does not model 4-4-5 calendars, week-based quarters, or partial first periods, because those cannot be derived from a start month alone. If you need them, you need a different library.

## API

- `NewQuarter(t time.Time, startMonth int) (Quarter, error)` returns the fiscal quarter containing `t`.
- `Quarter{FiscalYear int, Number int}` is the result. `Number` is 1..4, where 1 contains the start month. `FiscalYear` is the calendar year in which the fiscal year begins.
- `(Quarter) Start(startMonth int) (time.Time, error)` returns the first day of the quarter at 00:00:00 UTC.
- `(Quarter) End(startMonth int) (time.Time, error)` returns the last nanosecond of the quarter, i.e. `23:59:59.999999999` UTC on the final day.
- `(Quarter) Contains(t time.Time, startMonth int) (bool, error)` reports whether `t` is within the quarter, inclusive of both endpoints.
- `(Quarter) Calendar(startMonth int) (year, month int, err error)` returns the calendar year and month in which the quarter begins.
- `StartMonth(t time.Time, startMonth int) (int, error)` validates and returns the start month.
- `ErrInvalidStartMonth` is returned when `startMonth` is outside 1..12.

## The awkward edge

When the start month is not January, the fiscal year label is not the calendar year of the date. For a fiscal year starting in October, 1 February 2025 is in fiscal year 2024, quarter 2. The `FiscalYear` field always refers to the calendar year in which the fiscal year *begins*, which is the convention most accounting systems use. If your organisation labels that period "FY25", you will need to add one to `FiscalYear` yourself; the library does not guess.

## Time zones

All comparisons and boundary construction use UTC. `NewQuarter` and `Contains` convert their argument to UTC first, so a `time.Time` carrying a local zone is interpreted at the corresponding UTC instant. This is deliberate: quarter boundaries are instants, and mixing zones silently produces off-by-one-day errors.
