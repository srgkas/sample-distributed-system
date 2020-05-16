package dto

import (
	"encoding/gob"
	"time"
)

type SensorMessage struct {
	Name      string `json:"name"`
	Value     float64 `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

func init() {
	gob.Register(SensorMessage{})
}
