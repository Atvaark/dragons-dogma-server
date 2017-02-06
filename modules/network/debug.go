package network

import "log"

var debug bool

func printf(format string, v ...interface{}) {
	if debug {
		log.Printf(format, v...)
	}
}
