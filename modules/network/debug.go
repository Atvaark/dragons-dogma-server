package network

import "log"

var debug bool = false

func printf(format string, v ...interface{}) {
	if debug {
		log.Printf(format, v...)
	}
}
