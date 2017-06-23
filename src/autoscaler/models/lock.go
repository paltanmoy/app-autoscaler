package models

type Lock struct {
	Type                    string
	Owner                   string
	Last_Modified_Timestamp int64
	Ttl                     int
}
