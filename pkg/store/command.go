package store

import "encoding/json"

type Command struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func NewCommand(t string, v any) Command {
	b, _ := json.Marshal(v)
	return Command{Type: t, Payload: b}
}

