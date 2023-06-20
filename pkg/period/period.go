package period

import "time"

func GetLastPKUP(before int) (time.Time, time.Time) {
	actualMonth := time.Date(
		time.Now().Year(),
		time.Now().AddDate(0, before, 0).Month(),
		19,
		23, 59, 59, 0, time.Now().Local().Location(),
	).UTC()
	return actualMonth.AddDate(0, -1, 0).UTC(), actualMonth
}
