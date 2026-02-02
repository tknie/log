/*
* Copyright 2022-2026 Thorsten A. Knieling
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
	"io"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func initTestLogWithFile(t *testing.T, fileName string) error {
	err := InitZapLogLevelWithFile(fileName, zapcore.DebugLevel)
	if !assert.NoError(t, err) {
		t.Fatalf("error opening file: %v", err)
	}
	return err
}

func newWinFileSink(u *url.URL) (zap.Sink, error) {
	// Remove leading slash left by url.Parse()
	return os.OpenFile(u.Path[1:], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}

func doTrackCall() {
	defer TimeTrack(time.Now(), "Time Track Unit test ")

}

func TestLogZap(t *testing.T) {
	fileName := "zap.log"
	fileName = os.TempDir() + "/" + fileName
	fmt.Println("Use log file:", fileName)
	os.Remove(fileName)
	err := initTestLogWithFile(t, fileName)
	if !assert.NoError(t, err) {
		fmt.Println(err)
		return
	}

	d := IsDebugLevel()
	SetDebugLevel(true)
	doTrackCall()

	hallo := "HELLO"
	Log.Debugf("This is a test of data %s", hallo)

	LogMultiLineString(true, "ABC\nXXXX\n")
	SetDebugLevel(false)
	SetDebugLevel(d)

	flog, err := os.Open(fileName)
	if !assert.NoError(t, err) {
		return
	}
	logInfo, err := io.ReadAll(flog)
	if !assert.NoError(t, err) {
		return
	}
	assert.Regexp(t, `....-..-.. ..:..:..\tinfo\tStart logging with level debug
....-..-.. ..:..:..\tinfo\tTime Track Unit test  took [0-8]*ns
....-..-.. ..:..:..\tdebug\tThis is a test of data HELLO
....-..-.. ..:..:..\tdebug\tABC
....-..-.. ..:..:..\tdebug\tXXXX
....-..-.. ..:..:..\tdebug\t
`, string(logInfo))
}

func TestLogrus(t *testing.T) {
	fileName := "logrus.log"
	os.Remove(os.TempDir() + "/" + fileName)
	log := logrus.StandardLogger()

	fmt.Println("Init logging")
	SetDebugLevel(true)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02T15:04:05",
	})
	log.SetLevel(logrus.DebugLevel)
	p := os.Getenv("LOGPATH")
	if p == "" {
		p = os.TempDir()
	}
	f, err := os.OpenFile(p+"/"+fileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if !assert.NoError(t, err) {
		fmt.Println("Error opening log:", err)
		return
	}
	log.SetOutput(f)
	log.Infof("Init logrus")
	InitLog(log)
	fmt.Println("Logging running")

	flog, err := os.Open(os.TempDir() + "/" + fileName)
	if !assert.NoError(t, err) {
		return
	}
	logInfo, err := io.ReadAll(flog)
	if !assert.NoError(t, err) {
		return
	}
	assert.Regexp(t, "time=\"20..-..-..T..:..:..\" level=info msg=\"Init logrus\"\n", string(logInfo))
}

const testResult = `info:INFO: Pre-log information
error:ERROR: Pre-log information
debug:DEBUG: Post-log information
info:INFO: Post-log information
error:ERROR: Post-log information
`

type testLogger struct {
	testLog string
}

var testLog = &testLogger{}

func (t *testLogger) Debugf(format string, args ...interface{}) {
	t.testLog += "debug:" + fmt.Sprintf(format+"\n", args...)
}

func (t *testLogger) Infof(format string, args ...interface{}) {
	t.testLog += "info:" + fmt.Sprintf(format+"\n", args...)
}

func (t *testLogger) Errorf(format string, args ...interface{}) {
	t.testLog += "error:" + fmt.Sprintf(format+"\n", args...)
}

func (t *testLogger) Fatal(args ...interface{}) {
}

func (t *testLogger) Fatalf(format string, args ...interface{}) {
	t.testLog += "fatal:" + fmt.Sprintf(format+"\n", args...)
	os.Exit(1)
}

func TestCache(t *testing.T) {
	disableLog()
	Log.Debugf("DEBUG: Pre-log information")
	Log.Infof("INFO: Pre-log information")
	Log.Errorf("ERROR: Pre-log information")
	InitLog(testLog)
	Log.Debugf("DEBUG: Post-log information")
	Log.Infof("INFO: Post-log information")
	Log.Errorf("ERROR: Post-log information")
	assert.Equal(t, testResult, testLog.testLog)
}

func TestStackTest(t *testing.T) {
	InitZapLogLevelWithFile("StackTest.log", zapcore.DebugLevel)
	Log.Debugf("DEBUG: Post-log information")
	Log.Infof("INFO: Post-log information")
	Log.Errorf("ERROR: Post-log information")
	stack1()
}

func stack1() {
	LogFunctionStart()
	stack2()
	defer LogFunctionEnd(time.Now())

}

func stack2() {
	LogFunctionStart()
	stack3()
	defer LogFunctionEnd(time.Now())

}

func stack3() {
	LogFunctionStarts("stack3")
	Log.Debugf("In stack3")
	defer LogFunctionEnds(time.Now(), "stack3")
}
