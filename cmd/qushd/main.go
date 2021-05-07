package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	golog "log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/hugefiver/qush/consits"

	"github.com/hugefiver/qush/auth"

	"github.com/hugefiver/qush/wrap"

	"github.com/hugefiver/qush/ssh"

	"github.com/hugefiver/qush/quic"

	"github.com/hugefiver/qush/key"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"

	serverConfig "github.com/hugefiver/qush/config"
	"github.com/hugefiver/qush/config/conf"
	"github.com/hugefiver/qush/logger"
	"github.com/hugefiver/qush/util"
)

var version = "0.0.1"
var buildTime = "unknown"
var serverVersion = "QUSH-0.0.1"

//var keyRings agent.Agent

func main() {
	f := processArgs()

	c, err := serverConfig.LoadServerConfig(f.ConfigPath)
	if err != nil {
		golog.Fatalln("Cannot load server config:", err)
	}

	loadLogger(c.LogPath, c.LogLevel, f.Verbose)

	// if `genkey`
	if f.GenKey {
		path := f.KeyPath
		if path == "" {
			path = c.HostKeyPath
		}
		path = util.GetPath(path)

		log.Warn().Msgf("Generate host private key to %s", path)
		err := genKey(path)
		if err != nil {
			log.Fatal().Msgf("Cannot generate private key: %v", err)
		}
		return
	}

	var tlsConfig *tls.Config
	// make TLS Config
	{
		// Generate a TLS config from exist key files

		path := util.GetPath(c.HostKeyPath)
		file, err := os.Open(path)
		if err != nil {
			log.Fatal().Msgf("Open file %s failed", path)
		}
		pri, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatal().Msgf("Cannot read file %s", path)
		}

		path = path + ".pub"
		file, err = os.Open(path)
		if err != nil {
			log.Fatal().Msgf("Open file %s failed", path)
		}
		pub, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatal().Msgf("Cannot read file %s", path)
		}

		tlsConfig, err = key.GenTlsConfig(pub, pri)
		tlsConfig.NextProtos = []string{"qush"}
		if err != nil {
			log.Fatal().Err(err).Msgf("Parse TLS config failed")
		}

		// set application field
		tlsConfig.NextProtos = []string{"qush"}
	}

	quicConfig := &quic.Config{
		KeepAlive: true,
	}
	listener, err := quic.ListenAddr(fmt.Sprintf("%v:%v", c.Addr, c.Port), tlsConfig, quicConfig)
	if err != nil {
		log.Err(err).Msg("")
		log.Fatal().Msgf("QUSHD is exiting, cause can't listen at %v:%v", c.Addr, c.Port)
	} else {
		log.Warn().Msgf("Server is listening at UDP %v:%v", c.Addr, c.Port)
	}
	defer listener.Close()

	// config ssh server config
	serverConf := &ssh.ServerConfig{
		Config:                      ssh.Config{},
		NoClientAuth:                false,
		MaxAuthTries:                3,
		PasswordCallback:            auth.PasswordAuthFunc,
		PublicKeyCallback:           nil,
		KeyboardInteractiveCallback: nil,
		AuthLogCallback:             logAuthLog,
		ServerVersion:               serverVersion,
		BannerCallback:              nil,
		GSSAPIWithMICConfig:         nil,
	}

	// add host key
	k, err := key.LoadHostKey(c.HostKeyPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot load host key")
	}
	hostKey, err := ssh.NewSignerFromKey(k)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot parse host key")
	}
	serverConf.AddHostKey(hostKey)

	for {
		session, err := listener.Accept(context.Background())
		if err != nil {
			log.Info().Err(err).Msg("Failed to accept a QUIC session")
			continue
		}
		log.Info().Msgf("Accepted a QUIC session from %v", session.RemoteAddr())
		go handleQUICSession(session, serverConf)
	}

}

func handleQUICSession(session quic.Session, serverConf *ssh.ServerConfig) {
	defer session.CloseWithError(consits.DISCONNECT, "server disconnected")
	if s, err := session.AcceptStream(context.Background()); err != nil {
		addr := session.RemoteAddr()
		log.Debug().Err(err).Msgf("Cannot accept stream from %v, connection will close", addr)
		_ = session.CloseWithError(1, "Session closed")
	} else {
		log.Debug().Msgf("Accept a QUIC stream #%d from %v", s.StreamID(), session.RemoteAddr())
		conn, channels, reqs, err := ssh.NewServerConn(wrap.From(s, session), serverConf)
		if err != nil {
			addr := session.RemoteAddr()
			log.Info().Err(err).Msgf("Disconnected from %v", addr)
			return
		}
		log.Debug().Fields(map[string]interface{}{
			"conn":     conn,
			"channels": channels,
			"requests": reqs,
		}).Msg("Information of SSH connection")

		// Serve request channel
		go ssh.DiscardRequests(reqs)

		// Service the incoming Channel channel.
		for newChannel := range channels {
			if t := newChannel.ChannelType(); t != "session" {
				_ = newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
				continue
			}

			go handleSSHChannel(newChannel, conn.User())
			{
				//	// Channels have a type, depending on the application level
				//	// protocol intended. In the case of a shell, the type is
				//	// "session" and ServerShell may be used to present a simple
				//	// terminal interface.
				//	if newChannel.ChannelType() != "session" {
				//		_ = newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
				//		continue
				//	}
				//	channel, requests, err := newChannel.Accept()
				//	if err != nil {
				//		log.Debug().Err(err).Msg("Could not accept channel")
				//	}
				//
				//	// Sessions have out-of-band requests such as "shell",
				//	// "pty-req" and "env".  Here we handle only the
				//	// "shell" request.
				//	go func(in <-chan *ssh.Request) {
				//		for req := range in {
				//			_ = req.Reply(req.Type == "shell", nil)
				//		}
				//	}(requests)
				//
				//	terminal := term.NewTerminal(channel, "> ")
				//
				//	go func() {
				//		defer channel.Close()
				//		for {
				//			line, err := terminal.ReadLine()
				//			if err != nil {
				//				break
				//			}
				//			fmt.Println(line)
				//		}
				//	}()
			}
		}
	}

}

func loadLogger(path string, lvl string, verbose int) {
	if path != "none" {
		dir := filepath.Dir(util.GetPath(path))
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0775)
			if err != nil {
				golog.Fatalln("Cannot create log file folder:", err)
			}
		}
	}

	// Set logger
	fileWriter := logger.Writer(util.GetPath(path))

	//zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.TimeFieldFormat = time.RFC3339

	fileLogger := logger.NewWriterFilter(fileWriter, logger.Level(lvl))
	stdoutLogger := logger.NewWriterFilter(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		NoColor:    true,
		TimeFormat: "2006/01/02 15:04:05",
	}, logger.LevelN(verbose))
	muxWriter := zerolog.MultiLevelWriter(stdoutLogger, fileLogger)
	log.Logger = zerolog.New(muxWriter).
		Level(zerolog.DebugLevel).
		With().
		Timestamp().Logger()
}

func genKey(path string) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		log.Warn().Msgf("Key file at %s is already exist", path)
	} else if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0640)
		if err != nil {
			return err
		}
	}

	pub, pri, err := key.CreateEd25519Key()
	if err != nil {
		return err
	}

	priFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer func() { _ = priFile.Close() }()

	pubFile, err := os.OpenFile(path+".pub", os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer func() { _ = pubFile.Close() }()

	priBytes, err := key.MarshalPriKey(pri)
	if err != nil {
		return err
	}

	pubBytes, err := key.MarshalPubKey(pub)
	if err != nil {
		return err
	}

	_, err = priFile.Write(priBytes)
	if err != nil {
		return err
	}

	_, err = pubFile.Write(pubBytes)
	if err != nil {
		return err
	}

	return nil
}

func processArgs() *Flags {
	f := ParseFlags()
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Help of %s:\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "usages: qushd [-hvV] [--config] [--genconf] [--genkey [--keypath]]")
		pflag.PrintDefaults()
	}

	if f.Version {
		showVerbose()
		os.Exit(0)
	}

	if f.Help {
		pflag.Usage()
		os.Exit(0)
	}

	if f.ConfigPath == "" {
		if runtime.GOOS == "windows" {
			f.ConfigPath = filepath.Join(os.Getenv("LOCALAPPDATA"), "qush", "qushd_config.ini")
		} else {
			f.ConfigPath = "/etc/qushd/qushd_config.ini"
		}
	}

	if f.GenConfig {
		path := filepath.Dir(f.ConfigPath)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			err := os.MkdirAll(path, 0660)
			if err != nil {
				golog.Fatalln("Cannot make folder:", err)
			}
		}
		file, err := os.OpenFile(f.ConfigPath, os.O_CREATE|os.O_WRONLY, 0660)
		if err != nil {
			golog.Fatalln("Cannot write config file:", err)
		}

		_, err = file.WriteString(conf.DefaultServerConfig)
		if err != nil {
			golog.Fatalln("Cannot write config file:", err)
		}

		os.Exit(0)
	}

	return f
}

func showVerbose() {
	fmt.Println("QUSH - Quick UDP Shell")
	fmt.Printf("Server version %s, build time %s \n", version, buildTime)
	fmt.Println("Author Hugefiver<i@iruri.moe> 2021")
}
