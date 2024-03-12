package io

import (
	"bufio"
	"fmt"
	"io"
	"syscall"

	"golang.org/x/term"
)

type StdinReader func(sensitive bool) string

func NewStdinReader(inputoutput InputOutput) StdinReader {
	return func(sensitive bool) string {
		if sensitive {
			bytepw, _ := term.ReadPassword(int(syscall.Stdin))
			fmt.Fprintln(inputoutput.Stdout)
			return string(bytepw)
		} else {
			stringreader := bufio.NewReader(inputoutput.Stdin)
			switch str, err := stringreader.ReadString('\n'); err {
			case nil:
				return str
			case io.EOF:
				fmt.Fprintln(inputoutput.Stdout)
				return ""
			default:
				panic(err)
			}
		}
	}
}
