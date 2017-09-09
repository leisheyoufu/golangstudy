package main

import (
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"log"
	"fmt"
	"os/signal"
	"github.com/leisheyoufu/golangstudy/ssh/common"
	"time"
)

const	(
	exitSequence = "\x05c"  // ctrl-e, c
)

var (
	state *terminal.State
	escape int
)

func checkEscape(b []byte, n int)(int) {
	for i:=0; i<n; i++ {
		c := b[i]
		if c == '\x05' {
			escape = 1
		} else if c== 'c' {
			if escape == 1 {
				escape = 2
			}
		} else if c == '.' {
			if escape == 2 {
				return -1
			}
		}else {
			escape = 0
		}
	}
	return 0
}

func connectToHost(user, host string) (*ssh.Client, *ssh.Session, error) {
	var pass string
	// fmt.Print("Password: ")
	//fmt.Scanf("%s\n", &pass)
	pass = fmt.Sprintf("stg096917")

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(pass)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	client, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
	}

	return client, session, nil
}

func main() {
	if len(os.Args) != 4 {
		log.Fatalf("Usage: %s <user> <host:port> <command>", os.Args[0])
	}

	client, session, err := connectToHost(os.Args[1], os.Args[2])
	if err != nil {
		panic(err)
	}

	tty := common.Tty{}
	ttyWidth, err:= tty.Width()
	if err != nil {
		panic(err)
	}
	ttyHeight, err:= tty.Height()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d %d\n", ttyWidth,ttyHeight)


	if terminal.IsTerminal(int(os.Stdin.Fd())) {
		//state, err = terminal.GetState(int(os.Stdin.Fd()))
		if err != nil {
			panic(err)
		}
		state, err = terminal.MakeRaw(int(os.Stdin.Fd()))
	}
	defer terminal.Restore(int(os.Stdin.Fd()), state)

	modes := ssh.TerminalModes{
		ssh.ECHO:  1, // Disable echoing
		//ssh.IGNCR: 1, // Ignore CR on input.
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	// Request pseudo terminal
	//if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
	if err := session.RequestPty("xterm-256color", ttyWidth, ttyHeight, modes); err != nil {
	//if err := session.RequestPty("vt100", 80, 40, modes); err != nil {
		//if err := session.RequestPty("vt220", 80, 40, modes); err != nil {
		log.Fatalf("request for pseudo terminal failed: %s", err)
	}
	sshIn, _ := session.StdinPipe()
	sshOut,_ := session.StdoutPipe()
	// Start remote shell
	if err := session.Shell(); err != nil {
		log.Fatalf("failed to start shell: %s", err)
	}

	// Handle control + C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	escape = 0

	go func() {
		for {
			<-c
			fmt.Println("^C")
			fmt.Fprint(sshIn, "\n")
			os.Exit(1)
		}
	}()
	go func() {
		b := make([]byte, 4096)
		for {
			n, _ := os.Stdin.Read(b)
			exit := checkEscape(b, n)
			if n>0 {
				sshIn.Write(b[:n])
			}
			if exit == -1 {
				terminal.Restore(int(os.Stdin.Fd()), state)
				os.Exit(0)
			}
		}
	}()
	go func() {
		b := make([]byte, 4096)
		for {
			n, _ := sshOut.Read(b)
			if n > 0 {
				os.Stdout.Write(b[:n])
			}
		}


	}()
	// Accepting commands
	//for {
	//	reader := bufio.NewReader(os.Stdin)
	//	str, _ := reader.ReadString('\n')
	//	fmt.Fprint(in, str)
	//}
	session.Wait()

	client.Close()
}
