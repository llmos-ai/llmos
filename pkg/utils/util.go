package utils

import (
	"fmt"
)

func AddEnv(env []string, key, value string) []string {
	return append(env, fmt.Sprintf("%s=%s", key, value))
}
