package optional

import (
	"encoding/json"
)

var _ json.Marshaler = (*Null[string])(nil)

func (n Null[T]) MarshalJSON() ([]byte, error) {
	if !n.filled {
		return []byte("null"), nil
	}
	return json.Marshal(n.value)
}

var _ json.Unmarshaler = (*Null[string])(nil)

func (n *Null[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.filled = false
		return nil
	}
	err := json.Unmarshal(data, &n.value)
	n.filled = err == nil
	return err
}
