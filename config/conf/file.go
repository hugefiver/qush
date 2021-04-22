package conf

import (
	"bytes"
	_ "embed"
	"text/template"
)

//go:embed server.ini
var serverTemp string

//go:embed client.ini
var clientTemp string

var DefaultServerConfig string
var DefaultClientConfig string

func init() {
	buff := &bytes.Buffer{}
	_ = template.Must(template.New("server").Parse(serverTemp)).Execute(buff, serverValue)
	DefaultServerConfig = buff.String()

	DefaultClientConfig = clientTemp
}
