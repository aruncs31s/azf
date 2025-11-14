package utils

import "os"

func GetEnv(varName string) (string, error) {
	varVal := os.Getenv(varName)
	if varVal == "" {
		return "", ErrNoEnvVar
	}
	return varVal, nil
}
