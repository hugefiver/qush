module github.com/hugefiver/qush

go 1.16

require (
	github.com/go-ini/ini v1.62.0
	github.com/rs/zerolog v1.21.0
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/spf13/pflag v1.0.5
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b
	golang.org/x/term v0.0.0-20210421210424-b80969c67360
	gopkg.in/ini.v1 v1.62.0 // indirect
)

require (
	github.com/creack/pty v1.1.15
	github.com/golang/mock v1.6.0 // indirect
	github.com/hugefiver/qush/quic v0.0.0-20210422142946-188f46b2db88
	github.com/marten-seemann/qtls-go1-16 v0.1.4 // indirect
	github.com/mattn/go-colorable v0.1.8
	github.com/onsi/ginkgo v1.16.4 // indirect
	github.com/onsi/gomega v1.13.0 // indirect
)

replace github.com/hugefiver/quic-go => ./quic
