package hub

import "encoding/json"

type IncomingMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type OutgoingMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type ReadyPayload struct {
	Ready bool `json:"ready"`
}

type ErrorPayload struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func NewOutgoing(msgType string, payload interface{}) OutgoingMessage {
	return OutgoingMessage{Type: msgType, Payload: payload}
}
