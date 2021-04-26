package main

import flag "github.com/spf13/pflag"

type Flags struct {
	ConfigPath string
	GenConfig  bool

	GenKey  bool
	KeyPath string

	Verbose int

	Version bool
	Help    bool
}

func ParseFlags() *Flags {
	f := Flags{}

	flag.BoolVar(&f.GenKey, "genkey", false, "Generate host private key")
	flag.StringVar(&f.KeyPath, "keypath", "", "Key path for genkey, default is /etc/qush/qush_host_key")

	flag.StringVar(&f.ConfigPath, "config", "", "Config path")
	flag.BoolVar(&f.GenConfig, "genconf", false, "Generate config file with default")

	flag.CountVarP(&f.Verbose, "verbose", "v", "Level for messages with count of \"v\";\n"+
		"default is Fatal, 1: Error, 2: Warning, 3: Info, 4:Debug")

	flag.BoolVarP(&f.Version, "version", "V", false, "Show information")
	flag.BoolVarP(&f.Help, "help", "h", false, "Show this help page")

	flag.Parse()
	return &f
}
