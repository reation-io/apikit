package spec

import "encoding/json"

// marshalMap is a helper function to marshal a map to JSON
func marshalMap(m any) ([]byte, error) {
	return json.Marshal(m)
}

// unmarshalMap is a helper function to unmarshal JSON into a map
func unmarshalMap(data []byte, m any) error {
	return json.Unmarshal(data, m)
}
