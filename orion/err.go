package orion

import "fmt"

func handle(err error, desc string, args ...any) {
	if err != nil {
		text := fmt.Sprintf(desc, args...)
		panic(text + ": " + err.Error())
	}
}
