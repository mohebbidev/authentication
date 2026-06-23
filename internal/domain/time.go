package domain

import "time"

type RevokeTime time.Time

func Now() RevokeTime {
	return RevokeTime(time.Now().UTC())
}