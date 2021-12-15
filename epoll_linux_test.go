// +build linux

package xlistener

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"
)

const (
	DELIMITER byte = '\n'
)

func ReadLine(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)
	var buffer bytes.Buffer
	for {
		ba, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		buffer.Write(ba)
		if !isPrefix {
			break
		}
	}
	return buffer.String(), nil
}

func WriteLine(conn net.Conn, content string) (int, error) {
	content = fmt.Sprintf("%s%c", content, DELIMITER)
	writer := bufio.NewWriter(conn)
	number, err := writer.WriteString(content)
	if err == nil {
		err = writer.Flush()
	}
	return number, err
}

func findPort() net.Addr {
	if l, err := net.Listen("tcp", "127.0.0.1:0"); err != nil {
		panic("Could not find available server port")
	} else {
		defer func() { _ = l.Close() }()
		return l.Addr()
	}
}

func TestEpollListener(t *testing.T) {
	addr := findPort()
	ln, err := Listen(func() (listener net.Listener, e error) {
		return net.Listen(addr.Network(), addr.String())
	})
	if err != nil {
		t.Fatal(err)
	}
	connAccepted := make([]int, 0)
	defer func() { _ = ln.Close() }()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("epool listener close with:", err)
				return
			}
			go func(connIn net.Conn) {
				req, err := ReadLine(connIn)
				if err != nil {
					t.Fatal(err)
				}
				id, err := strconv.Atoi(req)
				if err != nil {
					t.Fatalf("get io error:%s  req:%s", err.Error(), req)
				}
				connAccepted = append(connAccepted, id)

				_, err = WriteLine(connIn, req)
				if err != nil {
					t.Fatal(err)
				}
			}(conn)
		}
	}()

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		conn, err := net.Dial(addr.Network(), addr.String())
		if err != nil {
			t.Fatal(err)
		}
		go func(connIn net.Conn, index int) {
			time.AfterFunc(time.Duration(1)*time.Second, func() {
				_ = connIn.Close()
				wg.Done()
			})
		}(conn, i)
	}
	connShouldAccepted := make([]int, 0)
	for i := 10; i < 20; i++ {
		wg.Add(1)
		conn, err := net.Dial(addr.Network(), addr.String())
		if err != nil {
			t.Fatal(err)
		}
		connShouldAccepted = append(connShouldAccepted, i)
		go func(connIn net.Conn, iIn int) {
			req := fmt.Sprintf("%d", iIn)
			_, err = WriteLine(connIn, req)
			if err != nil {
				t.Fatal(err)
			}
			rsp, err := ReadLine(connIn)
			if err != nil {
				t.Fatal(err)
			}
			if req != rsp {
				t.Fatalf("req rsp not equal,expected:%s got:%s", req, rsp)
			}
			_ = connIn.Close()
			wg.Done()
		}(conn, i)
	}
	wg.Wait()

	sort.Ints(connAccepted)
	sort.Ints(connShouldAccepted)
	if !reflect.DeepEqual(connAccepted, connShouldAccepted) {
		t.Fatalf("conn not equal,should: %v got: %v", connShouldAccepted, connAccepted)
	}
}
