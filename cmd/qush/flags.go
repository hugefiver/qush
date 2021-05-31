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

	IgnorePubKey bool

	PasswdTries int

	Cmd []string

	Version bool
	Help    bool

	DebugOnlyPasswd *string
}

func ParseFlags() *Flags {
	f := Flags{}

	flag.StringVarP(&f.User, "login", "l", "", "Login user name")
	flag.Uint16VarP(&f.Port, "port", "p", 22, "QUSH server port")

	flag.BoolVarP(&f.IgnorePubKey, "ignore", "I", false,
		"Set for ignore server public confirm")

	flag.IntVar(&f.PasswdTries, "tries", -1, "max tries of input password")

	flag.CountVarP(&f.Verbose, "verbose", "v", "Set for debug")

	flag.BoolVarP(&f.Version, "version", "V", false, "Show information")
	flag.BoolVarP(&f.Help, "help", "h", false, "Show this help page")

	flag.String("debug-passwd", "", "[only for debug]")

	// password from args only for debug
	passwd := flag.Lookup("debug-passwd")
	// hide from usage
	passwd.Hidden = true

	flag.Parse()

	// if `help` or `version` is set
	// return directly
	if f.Help || f.Version {
		return &f
	}

	if passwd.Changed {
		p := passwd.Value.String()
		f.DebugOnlyPasswd = &p
	}

	// `host` must be set
	if flag.NArg() == 0 {
		golog.Fatalln("host must be specialized")
	}

	// special args
	arg := flag.Args()
	host := arg[0]
	cmd := arg[1:]

	if f.Verbose > 0 {
		golog.Println("[debug] host:", host)
		golog.Println("[debug] cmd:", cmd)
	}

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
