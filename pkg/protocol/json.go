package protocol

import "encoding/json"

type jsonRawEnvelope = json.RawMessage

func MarshalParams(value any) (jsonRawEnvelope, error) {
	if value == nil {
		return nil, nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func DecodeResult[T any](response Response) (T, error) {
	var result T
	if len(response.Result) == 0 {
		return result, nil
	}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return result, err
	}
	return result, nil
}

func UnmarshalParams(data jsonRawEnvelope, target any) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, target)
}
