package whatip

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
)

const (
	// Allowed total request time (across many requests).
	maxHTTPRequestTime = time.Second * 120
	// Allowed Per-request awaited response time.
	maxHTTPResponseTime = time.Second * 15
)

var (
	// IfconfigMeHTTP provides an IP resolver using ifconfig.me.
	IfconfigMeHTTP = builtinHTTPResolver("https://ifconfig.me/ip")
	// ICanHazIPHTTP provides an IP resolver using icanhazip.com.
	ICanHazIPHTTP = builtinHTTPResolver("https://icanhazip.com")
	// AWSHTTP provides an IP resolver using AWS' checkip.amazonaws.com
	// endpoint.
	AWSHTTP = builtinHTTPResolver("https://checkip.amazonaws.com")
)

type httpResolver struct {
	url *url.URL
}

func (h *httpResolver) GetIP() (ip net.IP, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), maxHTTPRequestTime)
	defer cancel()

	client := &http.Client{
		// Limit allowed response time, let requests error and be retried as
		// appropriate.
		Timeout: maxHTTPResponseTime,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", h.url.String(), nil)
	if err != nil {
		return ip, errors.Wrap(err, "unable to prepare HTTP request")
	}
	// Ask for a plain text response, which toggles to responses being the IP
	// for services that return HTML otherwise (icanhazip.com).
	req.Header.Add("Accept", "text/plain")

	// try once
	if resp, err := client.Do(req); err == nil {
		return h.readIPResponse(resp)
	}

	// and if you fail, try try try again until you can't no more!
	backoff := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
	for {
		select {
		case <-ctx.Done():
			return ip, fmt.Errorf("could not retrieve IP from web source at %s, exceeded %s", h.url, maxHTTPRequestTime)
		case <-time.After(backoff.NextBackOff()):
			resp, err := client.Do(req)
			if err != nil {
				continue
			}
			return h.readIPResponse(resp)
		}
	}
}

func (h *httpResolver) readIPResponse(resp *http.Response) (ip net.IP, err error) {
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ip, err
	}
	ip = net.ParseIP(strings.TrimSpace(string(data)))
	if ip == nil {
		fmt.Printf("%q", string(data))
		err = fmt.Errorf("could not parse ip from response")
	}
	return ip, err
}

func NewHTTPResolver(source string) (IPResolver, error) {
	u, err := url.Parse(source)
	return &httpResolver{url: u}, err
}

func builtinHTTPResolver(source string) IPResolver {
	h, err := NewHTTPResolver(source)
	if err != nil {
		panic(err)
	}
	return h
}
