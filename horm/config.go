package horm

var printLog bool = false

func DisableLog() {
	printLog = false
}

func EnableLog() {
	printLog = true
}