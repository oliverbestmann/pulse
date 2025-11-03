package orion

func handle(err error, desc string) {
	if err != nil {
		panic(desc + ": " + err.Error())
	}
}
