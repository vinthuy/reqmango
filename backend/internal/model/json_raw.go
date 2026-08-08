package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONRawMessage is a json.RawMessage that implements sql.Scanner and driver.Valuer
// for SQLite compatibility (SQLite stores JSON as TEXT, but GORM can't scan text into []byte).
type JSONRawMessage json.RawMessage

func (j *JSONRawMessage) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("JSONRawMessage.Scan: unsupported type %T", value)
	}
	*j = make(JSONRawMessage, len(bytes))
	copy(*j, bytes)
	return nil
}

func (j JSONRawMessage) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return string(j), nil
}

func (j JSONRawMessage) MarshalJSON() ([]byte, error) {
	return json.RawMessage(j).MarshalJSON()
}

func (j *JSONRawMessage) UnmarshalJSON(data []byte) error {
	return (*json.RawMessage)(j).UnmarshalJSON(data)
}

// ToRawMessage converts to json.RawMessage for API usage.
func (j JSONRawMessage) ToRawMessage() json.RawMessage {
	return json.RawMessage(j)
}

// FromRawMessage creates a JSONRawMessage from json.RawMessage.
func FromRawMessage(r json.RawMessage) JSONRawMessage {
	return JSONRawMessage(r)
}
