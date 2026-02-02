/*
* Copyright 2026 Thorsten A. Knieling
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*    http://www.apache.org/licenses/LICENSE-2.0
*
 */

package log

import (
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var level = zapcore.ErrorLevel

func init() {
	ed := os.Getenv("ENABLE_LOGLEVEL")
	switch strings.ToLower(ed) {
	case "0", "error":
		level = zapcore.ErrorLevel
	case "1":
		level = zapcore.FatalLevel
	case "2", "warn", "warning":
		level = zapcore.WarnLevel
	case "3", "info":
		level = zapcore.InfoLevel
	case "4", "debug":
		level = zapcore.DebugLevel
	}
}

// InitZapLogWithFilename initializes the zap logger with a file name and default log level
func InitZapLogWithFilename(fileName string) (err error) {
	return InitZapLogLevelWithFilename(fileName, level)
}

// InitZapLogLevelWithFilename initializes the zap logger with a file name and log level
func InitZapLogLevelWithFilename(fileName string, levelIn zapcore.Level) (err error) {
	level = levelIn
	p := os.Getenv("LOGPATH")
	if p == "" {
		p = "."
	} else {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			err := os.Mkdir(p, os.ModePerm)
			if err != nil {
				fmt.Printf("Error creating log path '%s': %v\n", p, err)
				os.Exit(255)
			}
		}
	}

	name := p + string(os.PathSeparator) + fileName
	return InitZapLogWithFile(name)
}

// InitZapLogLevelWithFile initializes the zap logger with a file and log level
func InitZapLogLevelWithFile(fileName string, levelIn zapcore.Level) (err error) {
	level = levelIn

	return InitZapLogWithFile(fileName)
}

// InitZapLogWithFile initializes the zap logger with a file
func InitZapLogWithFile(name string) (err error) {
	lumberLog := &lumberjack.Logger{
		Filename:   name, // Location of the log file
		MaxSize:    10,   // Maximum file size (in MB)
		MaxBackups: 3,    // Maximum number of old files to retain
		MaxAge:     28,   // Maximum number of days to retain old files
		Compress:   true, // Whether to compress/archive old files
		LocalTime:  true, // Use local time for timestamps
	}
	writer := zapcore.AddSync(lumberLog)
	cfg := zap.NewProductionEncoderConfig()
	cfg.TimeKey = "timestamp"
	cfg.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg),
		writer,
		level,
	)
	logger := zap.New(core)
	defer logger.Sync() // Flush any buffered log entries

	sugar := logger.Sugar()

	sugar.Infof("Start logging with level %s", level)
	Log = sugar
	SetDebugLevel(level == zapcore.DebugLevel)

	return
}
