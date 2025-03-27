package safe

import (
	"log"
	"runtime/debug"
)

type fnc func()

func GoFunc(fn fnc) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Print("Recovering from panic: ", r, ", stacktrace:\n", string(debug.Stack()))
			}
		}()

		fn()
	}()
}
