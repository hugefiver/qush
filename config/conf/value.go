// +build !windows

package conf

var serverValue = serverMaterial{
	HostKeyPath: "/etc/qush/qush_host_key",
	LogPath:     "/var/log/qush.log",
}
