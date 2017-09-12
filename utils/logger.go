// Copyright 2017 The Xuefei Chen Authors. All rights reserved.
// Created on 2017/9/5 15:55
// Email chenxuefei_pp@163.com
// Logger

package utils

import (
    "sync"
    "log"
    "fmt"
    "os"
)

// Logger Level
const (
    DEBUG = 0
    INFO  = 1
    WARN  = 2
    ERROR = 3
    FATAL = 4
)

// Format Level to string
func prefix(level int) string {
    switch level {
    case DEBUG:
        return "[DEBUG]"
    case INFO:
        return "[INFO]"
    case WARN:
        return "[WARN]"
    case ERROR:
        return "[ERROR]"
    case FATAL:
        return "[FATAL]"
    }
    return "[UNKONWN]"
}

// Logger Config
type LoggerConfig struct {
    FileName  string    // Output filename default ""
    Console   bool      // Whether console output
    Date      bool      // Whether have date header
    Time      bool      // Whether have time header
    Micros    bool      // Whether have microseconds header
    LongName  bool      // Whether have long file name header
    ShortName bool      // Whether have short file name header
    UTC       bool      // Whether UTC header
    Level     int       // Logger level default is DEBUG
}

// Logger struct
type logger struct {
    config LoggerConfig
    l      *log.Logger
    mu     sync.Mutex
    file   *os.File
}

// Update config recreate logger
func (l *logger) UpdateConfig(config LoggerConfig) {
    l.update(config)
}

// Get current logger config
func (l *logger) GetConfig() LoggerConfig {
    l.mu.Lock()
    defer l.mu.Unlock()
    return l.config
}

func (l *logger) update(config LoggerConfig) {
    l.mu.Lock()
    defer l.mu.Unlock()

    options := int(0)
    var err error

    if config.Date {
        options |= log.Ldate
    }
    if config.Time {
        options |= log.Ltime
    }
    if config.Micros {
        options |= log.Lmicroseconds
    }
    if config.LongName {
        options |= log.Llongfile
    }
    if config.ShortName {
        options |= log.Lshortfile
    }
    if config.UTC {
        options |= log.LUTC
    }

    if l.config.FileName != config.FileName {
        if l.file != nil {
            l.file.Close()
        }
        if config.FileName != "" {
            l.file, err = os.OpenFile(config.FileName, os.O_RDWR|os.O_APPEND|os.O_SYNC|os.O_CREATE, 0666)
            if err != nil {
                panic(fmt.Sprintf("Cannot create file: %s\n", config.FileName))
            }
        } else {
            l.file = nil
        }
    }
    l.config = config

    l.l = log.New(l, "", options)

    if l.l == nil{
        panic("Logger create failed!")
    }
}

// Writer interface implement
func (l *logger) Write(p []byte) (n int, err error) {
    l.mu.Lock()
    defer l.mu.Unlock()

    var _err error
    var wl int
    if l.config.Console && l.config.Level > WARN {
        wl, _err = os.Stderr.Write(p)
    } else if l.config.Console {
        wl, _err = os.Stdout.Write(p)
    }
    if _err != nil {
        return wl, _err
    }
    if l.file != nil {
        wl, _err = l.file.Write(p)
        if _err != nil {
            return wl, _err
        }
    }
    return len(p), nil
}

// Output log string
func (l *logger) output(level int, format string, v ...interface{}) {
    if l.config.Level > level {
        return
    }
    if l.l == nil {
        l.update(l.config)
    }
    l.l.SetPrefix(prefix(level))
    switch level {
    case ERROR:
        l.l.Fatalf(format, v...)
    case FATAL:
        l.l.Panicf(format, v...)
    default:
        l.l.Printf(format, v...)
    }
}

// Single instance of logger
var Logger *logger = &logger{
    config: LoggerConfig{
        FileName:  "",
        Console:   true,
        Date:      true,
        Time:      true,
        Micros:    false,
        LongName:  false,
        ShortName: true,
        UTC:       false,
        Level:     DEBUG,
    },
    mu:   sync.Mutex{},
    l:    nil,
    file: nil,
}

// Debug logger
func Debug(format string, v ...interface{}) {
    Logger.output(DEBUG, format, v...)
}

// Info logger
func Info(format string, v ...interface{}) {
    Logger.output(INFO, format, v...)
}

// Warn logger
func Warn(format string, v ...interface{}) {
    Logger.output(WARN, format, v...)
}

// Error logger
func Error(format string, v ...interface{}) {
    Logger.output(ERROR, format, v...)
}

// Fatal logger
func Fatal(format string, v ...interface{}) {
    Logger.output(FATAL, format, v...)
}
