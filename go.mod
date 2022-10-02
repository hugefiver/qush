module github.com/hugefiver/qush

go 1.19

require (
	github.com/creack/pty v1.1.18
	github.com/go-ini/ini v1.67.0
	// github.com/hugefiver/quic v0.23.1-0.20210823085152-576d9dc5351f
	github.com/mattn/go-colorable v0.1.13
	github.com/rs/zerolog v1.28.0
	github.com/spf13/pflag v1.0.5
	golang.org/x/crypto v0.0.0-20220926161630-eccd6366d1be
	golang.org/x/term v0.0.0-20220919170432-7a66f970e087
)

// for github.com/hugefiver/qush/quic
require (
	github.com/cheekybits/genny v1.0.0 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.20.2 // indirect
	golang.org/x/net v0.0.0-20221002022538-bcab6841153b // indirect
	golang.org/x/sys v0.0.0-20220928140112-f11e5e49a4ec // indirect
)

require (
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
)

require (
	github.com/go-task/slim-sprig v0.0.0-20210107165309-348f09dbbbc0 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/marten-seemann/qtls-go1-18 v0.1.2 // indirect
	github.com/marten-seemann/qtls-go1-19 v0.1.0 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
	golang.org/x/exp v0.0.0-20221002003631-540bb7301a08 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/tools v0.1.12 // indirect
)

require github.com/hugefiver/quic v0.23.1

replace github.com/hugefiver/quic => ./quic
