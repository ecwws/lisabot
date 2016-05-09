package main

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
)

var logPrefix = [5]string{
	"[FATAL] ", "[ERROR] ", "[WARN] ", "[INFO] ", "[DEBUG] ",
}

// Logger struct is used for error checking and logging
type lisaLog struct {
	Debug *log.Logger
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger
	Level string
}

func NewLogger(writer io.Writer, level string) (*lisaLog, error) {

	logger := lisaLog{}
	nullLog := log.New(ioutil.Discard, "", 0)

	switch level {
	case "debug":
		logger.Debug = log.New(writer, "[DEBUG] ", log.LstdFlags|log.Lshortfile)
		fallthrough
	case "info":
		logger.Info = log.New(writer, "[INFO] ", log.LstdFlags|log.Lshortfile)
		fallthrough
	case "warn":
		logger.Warn = log.New(writer, "[WARN] ", log.LstdFlags|log.Lshortfile)
		fallthrough
	case "error":
		logger.Error = log.New(writer, "[ERROR] ", log.LstdFlags|log.Lshortfile)
		fallthrough
	default:
		switch {
		case logger.Error == nil:
			logger.Error = nullLog
			fallthrough
		case logger.Warn == nil:
			logger.Warn = nullLog
			fallthrough
		case logger.Info == nil:
			logger.Info = nullLog
			fallthrough
		case logger.Debug == nil:
			logger.Debug = nullLog
		}
	}

	if logger.Error == nullLog {
		return &logger, errors.New(
			"Log level has to one of debug, info, warn, error")
	}

	logger.Level = level

	return &logger, nil
}
