package usecase

import "time"

func datesInRange(from, to time.Time) []time.Time {
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	to = time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)
	var dates []time.Time
	for d := from; d.Before(to); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d)
	}
	return dates
}
