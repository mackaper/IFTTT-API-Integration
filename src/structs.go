package main

type IFTTTTriggerReqBody struct {
	TriggerIdentity string            `json:"trigger_identity"`
	TriggerFields   map[string]string `json:"trigger_fields"`
	User            string            `json:"user"`
	Limit           *int              `json:"limit"`
}

type IFTTTRequestValue struct {
	DynamicValue string `json:"dynamic_value"`
}

type IFTTTRespValue struct {
	DataValid   bool   `json: "data_valid"`
	DataMessage string `json:"message"`
}

type IFTTTErrorMessage struct {
	Message string `json:"message"`
}

type IFTTTDataType struct {
	TournamentName string        `json:"tournament_name"`
	GameName       string        `json:"game_name"`
	Competitors    string        `json:"competitors"`
	CreatedAt      string        `json:"created_at"`
	Meta           IFTTTMetaData `json:"meta"`
}

type IFTTTPayload struct {
	Data []IFTTTDataType `json:"data"`
}

type IFTTTMetaData struct {
	ID        int   `json:"id"`
	Timestamp int64 `json:"timestamp"`
}
