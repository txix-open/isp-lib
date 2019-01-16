package utils

import "encoding/json"

func ConvertStructToMapInterface(data interface{}) (*map[string]interface{}, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	result := make(map[string]interface{}, 0)
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
