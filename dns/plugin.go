package dns

import (
	"context"
	"fmt"
	"strings"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/go-redis/redis"
	"github.com/miekg/dns"
)

// HealthCheckPluginName is the name of the health check plugin needed
// to implement the Handler interface
const HealthCheckPluginName = "healthcheck"

// HealthCheckPlugin is a CoreDNS plugin that emulate various email
// authentication states.
type HealthCheckPlugin struct {
	Redis *redis.Client
	Next  plugin.Handler
}

// Name implements the Handler interface.
func (hc HealthCheckPlugin) Name() string {
	return HealthCheckPluginName
}

// ServeDNS retrieves the health check configuration for the requested message
// and returns an appropriate response.
func (hc HealthCheckPlugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	if state.QType() != dns.TypeTXT && state.QType() != dns.TypeSPF && state.QType() != dns.TypeMX {
		return plugin.NextOrFailure(hc.Name(), hc.Next, ctx, w, r)
	}

	a := new(dns.Msg)
	a.SetReply(r)
	a.Authoritative = true
	a.Compress = true

	var rr dns.RR

	messageID := strings.Split(state.QName(), ".")[0]
	// Lookup the message
	fmt.Println(messageID)

	switch state.QType() {
	case dns.TypeTXT:
		rr = new(dns.TXT)
		rr.(*dns.TXT).Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeTXT, Class: state.QClass()}
		rr.(*dns.TXT).Txt = []string{"It works!"}

	case dns.TypeSPF:
		rr = new(dns.SPF)
		rr.(*dns.SPF).Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeSPF, Class: state.QClass()}
		rr.(*dns.SPF).Txt = []string{"SPF Response"}

	case dns.TypeMX:
		fmt.Println("MX")
		rr = new(dns.MX)
		rr.(*dns.MX).Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeMX, Class: state.QClass()}
		rr.(*dns.MX).Preference = 10
		rr.(*dns.MX).Mx = "example.com."
	}

	a.Answer = []dns.RR{rr}

	state.SizeAndDo(a)
	w.WriteMsg(a)

	return dns.RcodeSuccess, nil
}
