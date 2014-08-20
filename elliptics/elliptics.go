//+build linux,cgo

package elliptics

/*
#cgo LDFLAGS: -lelliptics_cpp -lpthread
#cgo CXXFLAGS: -std=c++0x -W -Wall
*/
import "C"
