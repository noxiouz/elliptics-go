/*
* 2013+ Copyright (c) Anton Tyurin <noxiouz@yandex.ru>
* All rights reserved.
*
* This program is free software; you can redistribute it and/or modify
* it under the terms of the GNU General Public License as published by
* the Free Software Foundation; either version 2 of the License, or
* (at your option) any later version.
*
* This program is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
* GNU General Public License for more details.
 */

package elliptics

// #include "logger.h"
// #include <stdlib.h>
import "C"

import (
	"fmt"
	"unsafe"
)

type Logger struct {
	logger unsafe.Pointer
}

//Constants for log level of Logger.
const (
	//Log very important data. Practically nothing is written
	LOGDATA = iota
	//Log critical errors that materially affect the work
	LOGERROR
	//Log messages about the time of the various operations
	LOGINFO
	//It's a first level of debugging
	LOGNOTICE
	//Logs all sort of information about errors and work
	LOGDEBUG
)

/*NewFileLogger returns new Logger.
file is path to logfile.
level is log level.
*/
func NewFileLogger(file string, level int) (logger *Logger, err error) {
	cfile := C.CString(file)
	defer C.free(unsafe.Pointer(cfile))

	ellLogger, err := C.new_file_logger(cfile, C.int(level))
	if err != nil {
		return
	}
	logger = &Logger{ellLogger}
	return
}

func (logger *Logger) Free() {
	C.delete_file_logger(logger.logger)
}

func (logger *Logger) Log(level int, format string, args ...interface{}) {

	str := fmt.Sprintf(format, args...)
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))

	C.file_logger_log(logger.logger, C.int(level), cstr)
}

func (logger *Logger) GetLevel() (lvl int) {
	return int(C.file_logger_get_level(logger.logger))
}
