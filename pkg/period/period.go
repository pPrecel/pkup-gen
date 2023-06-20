package period

import "time"

func GetLastPKUP(before int) (time.Time, time.Time) {
	actualMonth := time.Date(
		time.Now().Year(),
		time.Now().AddDate(0, before, 0).Month(),
		18,
		23, 59, 59, 0, time.Now().Local().Location(),
	)

	pastMonth := actualMonth.AddDate(0, -1, 1)
	pastMonth = time.Date(
		pastMonth.Year(),
		pastMonth.Month(),
		19,
		0, 0, 0, 0, time.Now().Local().Location(),
	)

	return pastMonth.UTC(), actualMonth.UTC()
}
