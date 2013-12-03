package elliptics

// #include "logger.h"
// #include <stdlib.h>
import "C"

import (
	"unsafe"
)

type Logger struct {
	logger unsafe.Pointer
}

const (
	ERROR = iota
	WARNING
	INFO
	DEBUG
)

func NewFileLogger(file string) (logger *Logger, err error) {
	cfile := C.CString(file)
	defer C.free(unsafe.Pointer(cfile))

	ellLogger, err := C.new_file_logger(cfile)
	if err != nil {
		return
	}
	logger = &Logger{ellLogger}
	return
}

func (logger *Logger) Free() {
	C.delete_file_logger(logger.logger)
}

func (logger *Logger) Log(level int, msg string) {
	cmsg := C.CString(msg)
	defer C.free(unsafe.Pointer(cmsg))

	C.file_logger_log(logger.logger, C.int(level), cmsg)
}

func (logger *Logger) GetLevel() (lvl int) {
	return int(C.file_logger_get_level(logger.logger))
}
