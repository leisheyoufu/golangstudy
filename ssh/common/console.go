package common

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

const (
	exitSequence = "\x05c" // ctrl-e, c
)

type ConsoleSession interface {
	Wait() error
	Close() error
}

type Console struct {
	localIn      *os.File
	localOut     *os.File
	remoteIn     io.Writer
	remoteOut    io.Reader
	escape       int
	localInState *terminal.State
	sess         ConsoleSession
	inExit       chan bool
	outExit      chan bool
	sig          bool
}

func NewConsole(localIn *os.File, localOut *os.File, remoteIn io.Writer, remoteOut io.Reader, state *terminal.State, sess ConsoleSession) *Console {
	return &Console{localIn: localIn, localOut: localOut, remoteIn: remoteIn, remoteOut: remoteOut, localInState: state, sess: sess, inExit: make(chan bool, 1),
		outExit: make(chan bool, 1)}
}

func (c *Console) checkEscape(b []byte, n int) int {
	for i := 0; i < n; i++ {
		ch := b[i]
		if ch == '\x05' {
			c.escape = 1
		} else if ch == 'c' {
			if c.escape == 1 {
				c.escape = 2
			}
		} else if ch == '.' {
			if c.escape == 2 {
				return -1
			}
		} else {
			c.escape = 0
		}
	}
	return 0
}

func (c *Console) Start(path string) {
	defer c.close()
	go func() {
		defer c.close()
		b := make([]byte, 4096)
		for {
			select {
			case <-c.inExit:
				c.sig = true
			default:
			}
			n, err := c.localIn.Read(b)
			if err != nil {
				panic(err)
			}
			exit := c.checkEscape(b, n)
			if n > 0 {
				c.remoteIn.Write(b[:n])
			}
			if exit == -1 {
				c.close()
				return
			}
		}
	}()
	go func() {
		defer c.close()
		b := make([]byte, 4096)
		for {
			select {
			case <-c.outExit:
				c.sig = true
			default:
			}
			n, err := c.remoteOut.Read(b)
			if err != nil {
				panic(err)
			}
			if n > 0 {
				c.localOut.Write(b[:n])
				c.logger(path, b[:n])
			}
		}
	}()
	c.sess.Wait()
	terminal.Restore(int(c.localIn.Fd()), c.localInState)
	c.sess.Close()
}

func (c *Console) close() {
	if err := recover(); err != nil {
		if c.sig != true {
			fmt.Printf("Caught unexpected error: %s\n", err)
		}

	}
	terminal.Restore(int(c.localIn.Fd()), c.localInState)
	c.inExit <- true
	c.outExit <- true
	c.sess.Close()

}

func (c *Console) logger(path string, b []byte) error {
	if path != "" {
		var err error
		fd, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		l := len(b)
		for l > 0 {
			n, err := fd.Write(b)
			if err != nil {
				return err
			}
			l -= n
		}
		fd.Close()
	}
	return nil
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
