package logger

import (
	"encoding/json"
	"log"
	"time"
)

var format string = "2006-01-02T15:04:05.9999Z07:00"

func Info(message string, scope *map[string]any) {
	log.Printf("%s [INF] %s %s", time.Now().UTC().Format(format), message, asJson(scope))
}

func Debug(message string, scope *map[string]any) {
	log.Printf("%s [DBG] %s %s", time.Now().UTC().Format(format), message, asJson(scope))
}

func Error(err error, scope *map[string]any) {
	log.Printf("%s [ERR] %s %s", time.Now().UTC().Format(format), err.Error(), asJson(scope))
}

func asJson(d interface{}) []byte {
	b, _ := json.Marshal(d)
	return b
}
