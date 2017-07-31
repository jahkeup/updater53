package whatip

import (
	"fmt"
	"io/ioutil"
	"net"
	http "net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

const maxHTTPRequestTime = time.Second * 120

var (
	IfconfigMeHTTP = newHTTP("http://ifconfig.me/ip")
	ICanHazIPHTTP  = newHTTP("http://icanhazip.com")
	AWSHTTP        = newHTTP("http://checkip.amazonaws.com")
)

type HTTP struct {
	url *url.URL
}

func (h *HTTP) GetIP() (ip net.IP, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), maxHTTPRequestTime)
	backoff := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
	defer cancel()

	client := &http.Client{
		Timeout: time.Second * 15,
	}

	req, _ := http.NewRequest("GET", h.url.String(), nil)
	req.Header.Add("User-Agent", "curl/7.53.1")
	req.Header.Add("Accept", "text/plain")

	// try once
	if resp, err := ctxhttp.Do(ctx, client, req); err == nil {
		return h.readIPResponse(resp)
	}

	// and if you fail, try try try again until you can't no more!
	for {
		select {
		case <-ctx.Done():
			return ip, fmt.Errorf("could not retrieve IP from web source at %s, exceeded %s", h.url, maxHTTPRequestTime)
		case <-time.After(backoff.NextBackOff()):
			resp, err := ctxhttp.Do(ctx, client, req)
			if err != nil {
				continue
			}
			return h.readIPResponse(resp)
		}
	}
}

func (h *HTTP) readIPResponse(resp *http.Response) (ip net.IP, err error) {
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

func NewHTTP(source string) (h *HTTP, err error) {
	u, err := url.Parse(source)
	return &HTTP{url: u}, err
}

func newHTTP(source string) (h *HTTP) {
	h, err := NewHTTP(source)
	if err != nil {
		panic(err)
	}
	return h
}
