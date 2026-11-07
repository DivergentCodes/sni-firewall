package main

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"
)

func TestAllowConnectionForwardsBytesAlreadyRead(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	received := make(chan []byte, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			received <- nil
			return
		}
		defer conn.Close()
		if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
			received <- nil
			return
		}
		buf := make([]byte, len("GET / HTTP/1.0\r\n"))
		if _, err := io.ReadFull(conn, buf); err != nil {
			received <- nil
			return
		}
		received <- buf
	}()

	client, filterSide := net.Pipe()
	defer client.Close()
	go allowConnection(filterSide, listener.Addr().String(), []byte("G"), nil)

	if _, err := client.Write([]byte("ET / HTTP/1.0\r\n")); err != nil {
		t.Fatal(err)
	}

	select {
	case got := <-received:
		if want := []byte("GET / HTTP/1.0\r\n"); !bytes.Equal(got, want) {
			t.Errorf("destination received %q, want %q", got, want)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the destination")
	}
}
