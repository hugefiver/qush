package main

import (
	"fmt"
	"os"

	"golang.org/x/term"

	"github.com/hugefiver/qush/ssh"
)

func EnterPasswd(user string, max int) ssh.AuthMethod {
	callback := func() (string, error) {
		fmt.Printf("Input password for %s: ", user)
		p, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		return string(p), err
	}

	//return ssh.PasswordCallback(callback)
	return ssh.RetryableAuthMethod(ssh.PasswordCallback(callback), max)
}
