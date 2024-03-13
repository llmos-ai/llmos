package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

const (
	kernelParamPrefix = "LLMOS"
)

func ToEnv(prefix string, obj interface{}) ([]string, error) {
	data, err := EncodeToMap(obj)
	if err != nil {
		return nil, err
	}

	return mapToEnv(prefix, data), nil
}

func mapToEnv(prefix string, data map[string]interface{}) []string {
	var result []string
	for k, v := range data {
		keyName := strings.ToUpper(prefix + ToYAMLKey(k))
		if data, ok := v.(map[string]interface{}); ok {
			subResult := mapToEnv(keyName+"_", data)
			result = append(result, subResult...)
		} else {
			result = append(result, fmt.Sprintf("%s=%v", keyName, v))
		}
	}
	return result
}

func ToYAMLKey(str string) string {
	var result []rune
	cap := false

	for i, r := range []rune(str) {
		if i == 0 {
			if unicode.IsUpper(r) {
				cap = true
			}
			result = append(result, unicode.ToLower(r))
			continue
		}

		if unicode.IsUpper(r) {
			if cap {
				result = append(result, unicode.ToLower(r))
			} else {
				result = append(result, '_', unicode.ToLower(r))
			}
		} else {
			cap = false
			result = append(result, r)
		}
	}

	return string(result)
}

func EncodeToMap(obj interface{}) (map[string]interface{}, error) {
	if m, ok := obj.(map[string]interface{}); ok {
		return m, nil
	}

	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{}
	dec := json.NewDecoder(bytes.NewBuffer(b))
	dec.UseNumber()
	return result, dec.Decode(&result)
}
