/*
 * // Copyright 2022 The SimFG Authors
 * //
 * // Licensed under the Apache License, Version 2.0 (the "License");
 * // you may not use this file except in compliance with the License.
 * // You may obtain a copy of the License at
 * //
 * //     http://www.apache.org/licenses/LICENSE-2.0
 * //
 * // Unless required by applicable law or agreed to in writing, software
 * // distributed under the License is distributed on an "AS IS" BASIS,
 * // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * // See the License for the specific language governing permissions and
 * // limitations under the License.
 */

package tool

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var (
	logger   *zap.Logger
	record   *zap.Logger
	isRecord bool
	isDebug  bool
)

func init() {
	core := zapcore.NewCore(getEncoder(), getLoggerWriter("interfacer.log"), zapcore.InfoLevel)

	logger = zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)

	record = zap.New(zapcore.NewCore(getEmptyEncoder(), getLoggerWriter("record.log"), zapcore.InfoLevel),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)
}

func getLoggerWriter(fileName string) zapcore.WriteSyncer {
	file, _ := os.Create(fileName)
	return zapcore.AddSync(file)
}

func getEncoder() zapcore.Encoder {
	config := zap.NewProductionEncoderConfig()
	// readable time
	//config.EncodeTime = zapcore.ISO8601TimeEncoder
	// upper info, like INFO/WARN
	//config.EncodeLevel = zapcore.CapitalLevelEncoder
	config.TimeKey = ""
	config.LevelKey = ""
	return zapcore.NewConsoleEncoder(config)
}

func getEmptyEncoder() zapcore.Encoder {
	config := zapcore.EncoderConfig{}
	return zapcore.NewConsoleEncoder(config)
}

func EnableRecord(enable bool) {
	isRecord = enable
}

func EnableDebug(enable bool) {
	isDebug = enable
}

func Record(fields ...zap.Field) {
	if isRecord {
		record.Info("", fields...)
	}
	Info("Record", fields...)
}

func Info(msg string, fields ...zap.Field) {
	if isDebug {
		logger.Info(msg, fields...)
	}
}

func Warn(msg string, fields ...zap.Field) {
	if isDebug {
		logger.Warn(msg, fields...)
	}
}

func Panic(msg string, fields ...zap.Field) {
	logger.Panic(msg, fields...)
}
