/*
* Copyright 2022-2024 Thorsten A. Knieling
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
)

const maxBufferEntries = 100

type errorLevel byte

const (
	fataL errorLevel = iota
	errorL
	infoL
	debugL
)

type tempStoreElement struct {
	level   errorLevel
	message string
}

var tempStore = make([]tempStoreElement, 0)

type nilLogger struct {
}

func lognil() *nilLogger {
	return &nilLogger{}
}

func disableLog() {
	Log = lognil()
}

func shrinkTempStore() {
	if len(tempStore) > maxBufferEntries {
		tempStore = tempStore[len(tempStore)-maxBufferEntries:]
	}
}

func (*nilLogger) Debugf(format string, args ...interface{}) {
}

func (*nilLogger) Infof(format string, args ...interface{}) {
	tempStore = append(tempStore, tempStoreElement{infoL, fmt.Sprintf(format, args...)})
	shrinkTempStore()
}

func (*nilLogger) Errorf(format string, args ...interface{}) {
	tempStore = append(tempStore, tempStoreElement{errorL, fmt.Sprintf(format, args...)})
	shrinkTempStore()
}

func (*nilLogger) Fatal(args ...interface{}) {
}

func (*nilLogger) Fatalf(format string, args ...interface{}) {
	tempStore = append(tempStore, tempStoreElement{fataL, fmt.Sprintf(format, args...)})
	os.Exit(1)
}

// Log defines the log interface to manage other Log output frameworks
type LogI interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

// Central central configuration
var Log = LogI(lognil())
var debug = false

func InitLog(newLog LogI) {
	Log = newLog
	if len(tempStore) > 0 {
		for _, t := range tempStore {
			switch t.level {
			case fataL:
				Log.Fatalf(t.message)
			case errorL:
				Log.Errorf(t.message)
			case infoL:
				Log.Infof(t.message)
			default:
			}
		}
	}
	tempStore = make([]tempStoreElement, 0)
}

func IsDebugLevel() bool {
	return debug
}

func SetDebugLevel(debugIn bool) {
	debug = debugIn
	if debug {
		fmt.Println("Warning DB debug is enabled")
	}
}

// LogMultiLineString log multi line string to log. This prevent the \n display in log.
// Instead multiple lines are written to log
func LogMultiLineString(debug bool, logOutput string) {
	if debug && !IsDebugLevel() {
		return
	}
	columns := strings.Split(logOutput, "\n")
	for _, c := range columns {
		if debug {
			Log.Debugf("%s", c)
		} else {
			Log.Errorf("%s", c)
		}
	}
}

// TimeTrack defer function measure the difference end log it to log management, like
//
//	defer TimeTrack(time.Now(), "Info")
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	Log.Infof("%s took %s", name, elapsed)
}
