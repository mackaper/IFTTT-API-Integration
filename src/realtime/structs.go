package realtime

import (
	"pvk/API/athena"
	"pvk/API/db"
)

type IFTTTNotification struct {
	Data []IFTTTNotificationItem `json:"data"`
}

type IFTTTNotificationItem struct {
	TriggerIdentity string `json:"trigger_identity"`
}

type Patch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

type SocketSeries struct {
	IsSent  bool `json:"is_sent"`
	Payload struct {
		Patch []Patch       `json:"patch"`
		State athena.Series `json:"state"`
	} `json:"payload"`
}

type StateManager struct {
	CurrentSeries []int
	Cache         athena.Athena
	Client        *db.DBClient
}
