package token

import (
	"context"
	"fmt"
	"os"
	"strings"
)

const defaultTokenPath = "/var/lib/llmos/token"

func GetLocalToken(_ context.Context) (string, error) {
	token, err := os.ReadFile(defaultTokenPath)
	if err != nil {
		return "", err
	}
	if len(token) == 0 {
		return "", fmt.Errorf("token is empty")
	}

	parts := strings.Split(string(token), ":")
	if len(parts) > 1 {
		return parts[len(parts)-1], nil
	}

	return string(token), nil
}
