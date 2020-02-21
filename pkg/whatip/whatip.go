package whatip

import (
	"net"
)

var (
	// Default provides a safe bet for public ip determination.
	Default = OpenDNS
)

// IPers can return the public IP address of the caller.
type IPResolver interface {
	GetIP() (ip net.IP, err error)
}
