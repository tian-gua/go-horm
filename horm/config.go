package horm

var isPrintLog bool = true

func DisableLog() {
	isPrintLog = false
}

func EnableLog() {
	isPrintLog = true
}
