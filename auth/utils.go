package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func MustGetEnvInt(scope string) int {
	key, port := mustGetEnv(scope)
	portInt, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf(`invalid non-integral value %s=%s`, key, port)
	}

	return portInt
}

func MustGetEnvString(scope string) string {
	_, str := mustGetEnv(scope)
	return str
}

func mustGetEnv(scope string) (string, string) {
	key := scopeKey(scope)
	str := os.Getenv(key)
	if str == "" {
		log.Fatalf(`%s is not defined`, key)
	}
	return key, str
}

func scopeKey(scope string) string {
	return fmt.Sprintf(`ATES_%s`, scope)
}
