package elliptics

//+build linux,cgo

/*
#cgo LDFLAGS: -lelliptics_cpp -lpthread
#cgo CXXFLAGS: -std=c++0x -W -Wall
*/
import "C"
