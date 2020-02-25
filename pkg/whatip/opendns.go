package whatip

import (
	"fmt"
	"math/rand"
	"net"

	"github.com/miekg/dns"
)

var (
	// OpenDNS provides an IP address lookup mechanism via DNS query.
	OpenDNS = &openDNS{}

	opendnsResolvers = []string{
		"208.67.222.123:53",
		"208.67.220.123:53",
	}
)

type openDNS struct{}

func (openDNS) GetIP() (ip net.IP, err error) {
	client := new(dns.Client)
	msg := &dns.Msg{
		Question: []dns.Question{
			{Name: "myip.opendns.com.", Qclass: dns.ClassINET, Qtype: dns.TypeA},
		},
	}

	msg.RecursionDesired = true
	msg.Id = dns.Id()

	r, _, err := client.Exchange(msg, opendnsResolvers[rand.Intn(len(opendnsResolvers))])
	if err != nil {
		return ip, err
	}

	if len(r.Answer) != 1 {
		return ip, fmt.Errorf("opendns response did not contain 1 answer")
	}

	if record, ok := r.Answer[0].(*dns.A); ok {
		return record.A, nil
	} else {
		return ip, fmt.Errorf("opendns response was not an A record")
	}
}
