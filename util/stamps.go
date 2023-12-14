package util

import "time"

func NumericalTimeStamp() string {
	return time.Now().Format("20060102150405")
}
