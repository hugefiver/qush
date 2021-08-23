module github.com/hugefiver/qush

go 1.16

require (
	github.com/creack/pty v1.1.15
	github.com/go-ini/ini v1.62.0
	github.com/hugefiver/quic v0.23.1-0.20210823085152-576d9dc5351f
	github.com/mattn/go-colorable v0.1.8
	github.com/rs/zerolog v1.21.0
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/spf13/pflag v1.0.5
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b
	golang.org/x/term v0.0.0-20210421210424-b80969c67360
	gopkg.in/ini.v1 v1.62.0 // indirect
)

replace github.com/hugefiver/quic => ./quic
