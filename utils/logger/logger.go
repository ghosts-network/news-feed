package logger

import (
	"encoding/json"
	"log"
	"time"
)

var ApplicationName = ""

func Info(message string, scope *map[string]any) {
	log.Printf("%s %s [INFO] %s %s", time.Now().UTC().Format(time.RFC3339), ApplicationName, message, asJson(scope))
}

func Debug(message string, scope *map[string]any) {
	log.Printf("%s %s [DEBUG] %s %s", time.Now().UTC().Format(time.RFC3339), ApplicationName, message, asJson(scope))
}

func Error(err error, scope *map[string]any) {
	log.Printf("%s %s [ERROR] %s %s", time.Now().UTC().Format(time.RFC3339), ApplicationName, err.Error(), asJson(scope))
}

func asJson(d interface{}) []byte {
	b, _ := json.Marshal(d)
	return b
}
