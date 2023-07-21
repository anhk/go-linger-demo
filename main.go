package main

import (
	"errors"
	"fmt"
	"go-linger-init/checker"
	"net"
	"runtime/debug"
	"syscall"
	"time"
)

// 7个包
//
//	A->B Flags [S]
//	B->A Flags [S.]
//	A->B Flags [.]
//	A->B Flags [F.]
//	B->A Flags [.]
//	B->A Flags [F.]
//	A->B Flags [.]
func doProbe001() {
	conn, err := net.DialTimeout("tcp", "baidu.com:80", time.Second*2)
	Throw(err)

	defer conn.Close()
}

// 4个包
//
//	A->B Flags [S]
//	B->A Flags [S.]
//	A->B Flags [.]
//	A->B Flags [R.]
func doProbe002() {
	conn, err := net.DialTimeout("tcp", "baidu.com:80", time.Second*2)
	Throw(err)

	Throw(conn.(*net.TCPConn).SetLinger(0))

	defer conn.Close()
}

// 3个包 On Linux
//
//	A->B Flags [S]
//	B->A Flags [S.]
//	A->B Flags [R.]
func doProbe004() {
	c, err := checker.NewChecker()
	Throw(err)
	defer c.Close()

	c.TimeoutSecond = 2

	doCheck := func(addr string) {
		fmt.Printf("check %v ", addr)
		err := c.Check(addr)
		if err == nil {
			fmt.Println("healthy")
		} else if isNetworkError(err) {
			fmt.Println("unhealthy, because of network error: ", err)
		} else {
			fmt.Println("failed:", err)
		}
	}

	doCheck("www.baidu.com:80")
	doCheck("www.baidu.com:81")
	doCheck("www.baidu.cm:80")
}

func main() {
	doProbe001() // 7个包
	doProbe002() // 4个包
	doProbe004() // 3个包
}

func Throw(e any) {
	if e != nil {
		fmt.Printf("%s\n", e)
		fmt.Printf("%s\n", debug.Stack())
		panic(e)
	}
}

// 判断错误类型，看是不是网络原因引起的，其他原因则设置为healthy
// - ConnectTimeout
// - Unknown Host
// - Reset
// - Connection Refused
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	if e, ok := err.(net.Error); ok && e.Timeout() { // timeout
		return true
	}

	if _, ok := err.(*net.DNSError); ok {
		return true
	}

	if e, ok := err.(*net.OpError); ok {
		if e.Op == "dial" { // unknown host
			return true
		} else if e.Op == "read" { // maybe reset
			return true
		}
	}

	return errors.Is(err, syscall.ECONNREFUSED) // connect refused
}
