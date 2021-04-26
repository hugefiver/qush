package main

import (
	"fmt"
	golog "log"
	"os"
	"regexp"

	flag "github.com/spf13/pflag"
)

type Flags struct {
	Verbose int

	User string
	Port uint16
	Host string

	Cmd []string

	Version bool
	Help    bool
}

func ParseFlags() *Flags {
	f := Flags{}

	flag.StringVarP(&f.User, "login", "l", "", "Login user name")
	flag.Uint16VarP(&f.Port, "port", "p", 22, "QUSH server port")

	flag.CountVarP(&f.Verbose, "verbose", "v", "Set for debug")

	flag.BoolVarP(&f.Version, "version", "V", false, "Show information")
	flag.BoolVarP(&f.Help, "help", "h", false, "Show this help page")

	flag.Parse()

	if flag.NArg() == 0 {
		golog.Fatalln("host must be specialized")
	}

	// special args
	arg := flag.Args()
	host := arg[0]
	cmd := arg[1:]

	golog.Println("host:", host)
	golog.Println("cmd:", cmd)

	// parse user and hostname
	{
		patt := regexp.MustCompile(`^((\w+)@)?([^@]+)$`)
		ok := patt.MatchString(host)
		if !ok {
			golog.Fatalln(`host must be like "[user@]hostname"`)
		}
		m := patt.FindStringSubmatch(host)
		user := m[2]
		host := m[3]

		if user != "" && f.User == "" {
			f.User = user
		}
		f.Host = host
	}

	// cmd
	f.Cmd = cmd

	// check
	if f.User == "" {
		golog.Fatalln("please special a login user")
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Help of %s:\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "usages: qush [-hvV] [-l user] [user@]host ")
		flag.PrintDefaults()
	}

	return &f
}
