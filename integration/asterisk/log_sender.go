package main

import (
	"flag"
	"io"
	"net"
	"os"
	"time"
)

func main() {
	addr := flag.String("addr", "", "Addr (host:port) to send logs")
	flag.Parse()

	if *addr == "" {
		flag.Usage()
		return
	}

	var (
		conn net.Conn
		err  error
	)

	for {
		conn, err = net.Dial("tcp", *addr)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		break
	}
	defer conn.Close()

	w := io.MultiWriter(conn, os.Stdout)

	_, err = io.Copy(w, os.Stdin)
	if err != nil {
		return
	}
}
