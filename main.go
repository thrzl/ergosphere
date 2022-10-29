/*
Copyright © 2021 thrzl <thrizzle@skiff.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"github.com/phuslu/fastdns"
	"log"
	"net/netip"
	"net"
	"os"
	"github.com/miekg/dns"
)

var records = map[string]string{
	"test.service": "192.168.0.2",
}

var client = new(dns.Client)
type DNSHandler struct {
	Debug bool
}

func (h *DNSHandler) ServeDNS(rw fastdns.ResponseWriter, req *fastdns.Message) {
	if h.Debug {
		log.Printf("%s: CLASS %s TYPE %s\n", req.Domain, req.Question.Class, req.Question.Type)
	}

	m := new(dns.Msg)

	switch req.Question.Type {
	case fastdns.TypeA:
		m.SetQuestion(string(req.Domain), dns.TypeA)
		res, _, _ := client.Exchange(m, "1.1.1.1")
		var addrs []netip.Addr

		for _, r := range res.Answer {
			addrs = append(addrs, r.Header().Name())
		}

		fastdns.HOST(rw, req, 60, )
	case fastdns.TypeAAAA:
		fastdns.HOST(rw, req, 60, []netip.Addr{netip.MustParseAddr("2001:4860:4860::8888")})
		m.SetQuestion(string(req.Domain), dns.TypeAAAA)
		res, _, _ := client.Exchange(m, "1.1.1.1")
	case fastdns.TypeCNAME:
		fastdns.CNAME(rw, req, 60, []string{"dns.google"}, []netip.Addr{netip.MustParseAddr("8.8.8.8")})
		m.SetQuestion(string(req.Domain), dns.TypeCNAME)
		res, _, _ := client.Exchange(m, "1.1.1.1")
	case fastdns.TypeSRV:
		fastdns.SRV(rw, req, 60, []net.SRV{{"www.google.com", 443, 1000, 1000}})
		m.SetQuestion(string(req.Domain), dns.TypeSRV)
		res, _, _ := client.Exchange(m, "1.1.1.1")
	case fastdns.TypeNS:
		fastdns.NS(rw, req, 60, []net.NS{{"ns1.google.com"}, {"ns2.google.com"}})
		m.SetQuestion(string(req.Domain), dns.TypeNS)
		res, _, _ := client.Exchange(m, "1.1.1.1")
	case fastdns.TypeMX:
		fastdns.MX(rw, req, 60, []net.MX{{"mail.gmail.com", 10}, {"smtp.gmail.com", 10}})
		m.SetQuestion(string(req.Domain), dns.TypeMX)
		res, _, _ := client.Exchange(m, "1.1.1.1")
	case fastdns.TypeSOA:
		fastdns.SOA(rw, req, 60, net.NS{"ns1.google"}, net.NS{"ns2.google"}, 60, 90, 90, 180, 60)
		m.SetQuestion(string(req.Domain), dns.TypeSOA)
		res, _, _ := client.Exchange(m, "1.1.1.1")
	case fastdns.TypePTR:
		fastdns.PTR(rw, req, 0, "ptr.google.com")
		m.SetQuestion(string(req.Domain), dns.TypePTR)
		res, _, _ := client.Exchange(m, "1.1.1.1")
	case fastdns.TypeTXT:
		m.SetQuestion(string(req.Domain), dns.TypeTXT)
		res, _, _ := client.Exchange(m, "1.1.1.1")
		fastdns.TXT(rw, req, 60, res.String())
		
	default:
		fastdns.Error(rw, req, fastdns.RcodeNXDomain)
	}

	m.SetQuestion()
}

func main() {
	addr := ":53"

	server := &fastdns.ForkServer{
		Handler: &DNSHandler{
			Debug: os.Getenv("DEBUG") != "",
		},
		Stats: &fastdns.CoreStats{
			Prefix: "ergo_",
			Family: "1",
			Proto:  "udp",
			Server: "dns://" + addr,
			Zone:   ".",
		},
		ErrorLog: log.Default(),
	}

	err := server.ListenAndServe(addr)
	if err != nil {
		log.Fatalf("dnsserver error: %+v", err)
	}
}
