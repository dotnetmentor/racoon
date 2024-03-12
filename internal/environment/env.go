package environment

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func StringVar(key, defaultValue string) string {
	v := os.Getenv(key)
	if v != "" {
		return v
	}
	return defaultValue
}

func IntVar(key string, defaultValue int) (int, error) {
	v := StringVar(key, fmt.Sprintf("%d", defaultValue))
	pv, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("failed parsing int from environment variable %s", key)
	}
	return pv, nil
}

func BoolVar(key string, defaultValue bool) (bool, error) {
	v := StringVar(key, strconv.FormatBool(defaultValue))
	pv, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("failed parsing boolean from environment variable %s", key)
	}
	return pv, nil
}

func DurationVar(key string, defaultValue time.Duration) (time.Duration, error) {
	v := StringVar(key, defaultValue.String())
	pv, err := time.ParseDuration(v)
	if err != nil {
		return time.Second * 0, fmt.Errorf("failed parsing duration from environment variable %s", key)
	}
	return pv, nil
}
