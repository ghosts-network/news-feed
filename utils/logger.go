package utils

import (
	"encoding/json"
	"log"
	"time"
)

type Logger struct {
	appName string
	scope   *map[string]any
}

func NewLogger(appName string) *Logger {
	log.SetFlags(0)
	return &Logger{
		appName: appName,
		scope:   &map[string]any{},
	}
}

func NewLoggerWithContext(appName string, context *map[string]any) *Logger {
	return &Logger{
		appName: appName,
		scope:   context,
	}
}

func (l *Logger) WithValue(key string, value any) *Logger {
	newScope := make(map[string]any)
	for k, v := range *l.scope {
		newScope[k] = v
	}
	newScope[key] = value
	return NewLoggerWithContext(l.appName, &newScope)
}

func (l *Logger) Info(message string) *Logger {
	log.Printf("%s %s [INFO] %s %s", time.Now().UTC().Format(time.RFC3339), l.appName, message, asJson(l.scope))
	return l
}

func (l *Logger) Debug(message string) *Logger {
	log.Printf("%s %s [DEBUG] %s %s", time.Now().UTC().Format(time.RFC3339), l.appName, message, asJson(l.scope))
	return l
}

func (l *Logger) Error(err error) *Logger {
	log.Printf("%s %s [ERROR] %s %s", time.Now().UTC().Format(time.RFC3339), l.appName, err.Error(), asJson(l.scope))
	return l
}

func asJson(d interface{}) []byte {
	b, _ := json.Marshal(d)
	return b
}
