package models

type CustomMetric struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	Value      float64 `json:"value"`
	Unit       string  `json:"unit"`
	Timestamp  int64   `json:"timestamp"`
	InstanceID string  `json:"instance_id"`
	AppGUID    string  `json:"app_guid"`
}
