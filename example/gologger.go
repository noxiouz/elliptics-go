package main

import "unsafe"
import "fmt"
import "github.com/bioothod/elliptics-go/elliptics"
import "log"
import "os"
import "C"

type MyLogger struct {
	l *log.Logger
}

func GoLogFunc(priv unsafe.Pointer, level int, msg *C.char) {
	var log *MyLogger
	log = (*MyLogger)(priv)

	log.l.Printf("%d: %s", level, C.GoString(msg))
}

var GoLogVar = GoLogFunc

func main() {
	var log MyLogger = MyLogger{l: log.New(os.Stdout, "test-prefix:", log.LstdFlags | log.Lmicroseconds) }

	log.l.Printf("log ptr: %p\n", &log)

	n, err := elliptics.NewNodeLog(unsafe.Pointer(&GoLogVar), unsafe.Pointer(&log), 4)
	if err != nil {
		fmt.Println("failed to create new node: ", err)
		return
	}
	defer n.Free()
}
