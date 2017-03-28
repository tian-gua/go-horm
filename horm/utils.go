package horm

import (
	"bytes"
	"github.com/fatih/color"
	"log"
	"runtime"
	"strconv"
)

//获取goroutine id
func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func printLog(s string) {
	if isPrintLog {
		formatS := color.GreenString("%s", s)
		log.Printf("[horm]:%s", formatS)
	}
}
