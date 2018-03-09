package handler

import (
	"fmt"
	"github.com/kr/pty"
	"golang.org/x/net/websocket"
	"os"
	"os/exec"
	"strings"
)

const (
	CMD = "screen"
)

func NewWebSocketHandler() websocket.Handler {
	return websocket.Handler(echoHandler)
}
func echoHandler(ws *websocket.Conn) {
	session := NewSession(ws)
	go session.readThread()
	session.writeThread()
}

type Session struct {
	cmd *exec.Cmd
	pty *os.File
	ws  *websocket.Conn
}

func NewSession(ws *websocket.Conn) *Session {
	var err error
	args := strings.Split(CMD, " ")
	session := new(Session)
	session.ws = ws
	session.cmd = exec.Command(args[0], args[1:]...)
	session.pty, err = pty.Start(session.cmd)
	if err != nil {
		panic(err)
	}
	return session
}

func (self *Session) readThread() {
	var n int
	var err error
	b := make([]byte, 1024)
	for {
		n, err = self.pty.Read(b)
		if err != nil {
			fmt.Errorf("Failed to read from pty %s", err.Error())
			break
		}
		err = self.send(b[:n])
		if err != nil {
			return
		}
	}
}

func (self *Session) send(b []byte) error {
	var count int
	var tmp int
	var err error
	n := len(b)
	tmp = 0
	for n > 0 {
		count, err = self.ws.Write(b[tmp:])
		if err != nil {
			fmt.Errorf("Failed to write to websocket %s", err.Error())
			return err
		}
		tmp += count
		n -= count
	}
	return nil
}

func (self *Session) writeThread() {
	var n int
	var err error
	b := make([]byte, 1024)
	for {
		n, err = self.ws.Read(b)
		if err != nil {
			fmt.Errorf("Failed to read from websocket %s", err.Error())
			return
		}
		err = self.output(b[:n])
		if err != nil {
			return
		}
	}
}

func (self *Session) output(b []byte) error {
	var count int
	var tmp int
	var err error
	tmp = 0
	n := len(b)
	for n > 0 {
		count, err = self.pty.Write(b[tmp:])
		if err != nil {
			fmt.Errorf("Failed to write to pty %s", err.Error())
			return err

		}
		tmp += count
		n -= count
	}
	return nil
}
