package common

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

type SSH struct {
	user           string
	password       string
	privateKeyFile string
	host           string
	client         *ssh.Client
	session        *ssh.Session
}

func NewPasswordSSH(host string, user string, password string) *SSH {
	return &SSH{host: host, user: user, password: password}
}

func (s *SSH) appendPrivateKeyAuthMethod(autoMethods *[]ssh.AuthMethod) {
	if s.privateKeyFile != "" {
		key, err := ioutil.ReadFile(s.privateKeyFile)
		if err != nil {
			log.Printf("host:%s\tThe private key file %s can not be parsed, Error:%s", s.host, s.privateKeyFile, err)
			return
		}

		signer, err := ssh.ParsePrivateKey([]byte(key))
		if err != nil {
			log.Printf("host:%s\tThe private key file %s can not be parsed, Error:%s", s.host, s.privateKeyFile, err)
			return
		}
		*autoMethods = append(*autoMethods, ssh.PublicKeys(signer))
	}
}

func (s *SSH) appendPasswordAuthMethod(autoMethods *[]ssh.AuthMethod) {
	if s.password != "" {
		*autoMethods = append(*autoMethods, ssh.Password(s.password))
	}

}

func (s *SSH) ConnectToHost() error {
	var err error
	autoMethods := make([]ssh.AuthMethod, 0)
	s.appendPrivateKeyAuthMethod(&autoMethods)
	s.appendPasswordAuthMethod(&autoMethods)
	sshConfig := &ssh.ClientConfig{
		User:            s.user,
		Auth:            autoMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	s.client, err = ssh.Dial("tcp", s.host, sshConfig)
	if err != nil {
		return err
	}

	s.session, err = s.client.NewSession()
	if err != nil {
		s.client.Close()
		return err
	}
	return nil
}

func (s *SSH) StartConsole(localIn *os.File, localOut *os.File) (*Console, error) {
	var state *terminal.State
	tty := Tty{}
	ttyWidth, err := tty.Width()
	if err != nil {
		return nil, err
	}
	ttyHeight, err := tty.Height()
	if err != nil {
		return nil, err
	}
	if terminal.IsTerminal(int(localIn.Fd())) {
		if err != nil {
			return nil, err
		}
		state, err = terminal.MakeRaw(int(localIn.Fd()))
		if err != nil {
			return nil, err
		}
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // Disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	if err := s.session.RequestPty("xterm-256color", ttyWidth, ttyHeight, modes); err != nil {
		terminal.Restore(int(localIn.Fd()), state)
		return nil, err
	}
	sshIn, err := s.session.StdinPipe()
	if err != nil {
		terminal.Restore(int(localIn.Fd()), state)
		return nil, err
	}
	sshOut, err := s.session.StdoutPipe()
	if err != nil {
		terminal.Restore(int(localIn.Fd()), state)
		return nil, err
	}
	// Start remote shell
	if err := s.session.Shell(); err != nil {
		terminal.Restore(int(localIn.Fd()), state)
		return nil, err
	}
	console := (localIn, localOut, sshIn, sshOut, state, s)
	return console, nil
}

func (s *SSH) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

func (s *SSH) Wait() error {
	if s.session != nil {
		return s.session.Wait()
	}
	return nil
}
