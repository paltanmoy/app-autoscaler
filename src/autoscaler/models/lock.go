package models

import (
	"time"
)

type Lock struct {
	Type                    string
	Owner                   string
	Last_Modified_Timestamp int64
	Ttl                     time.Duration
}
