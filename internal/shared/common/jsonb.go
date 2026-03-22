package common

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONB is a generic type for storing any Go type as JSONB in PostgreSQL.
// Usage: type ExamplesJSON = jsonb.JSONB[[]domain.Example]
type JSONB[T any] struct {
	Data T
}

func NewJSONB[T any](data T) JSONB[T] {
	return JSONB[T]{Data: data}
}

func (jsonb JSONB[T]) Value() (driver.Value, error) {
	bytes, err := json.Marshal(jsonb.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSONB: %w", err)
	}
	return string(bytes), nil
}

func (jsonb *JSONB[T]) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("JSONB.Scan: unsupported type %T", value)
	}
	return json.Unmarshal(bytes, &jsonb.Data)
}
