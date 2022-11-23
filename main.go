package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/miekg/dns"
	"github.com/karlseguin/ccache/v3"
	"net/http"
	"io"
	"ergosphere/utils"
	"net"
)

func fetchHosts() ([]string) {
	blockResp, err := http.Get("https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts")
	if err != nil {
		panic(err)
	}
	defer blockResp.Body.Close()
	blockBody, err := io.ReadAll(blockResp.Body)
	if err != nil {
		panic(err)
	}
	hosts, err := utils.ParseHostFile(string(blockBody))
	if err != nil {
		panic(err)
	}
	log.Printf("[i] fetched %d hosts from StevenBlack/hosts\n", len(hosts))
	return hosts
}

var blockList = fetchHosts()
var cache = ccache.New(ccache.Configure[*dns.Msg]().MaxSize(10240).ItemsToPrune(500))

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	transport := "udp"
	if _, ok := w.RemoteAddr().(*net.TCPAddr); ok {
		transport = "tcp"
	}
	if cachedResp := cache.Get(r.Question[0].String()); cachedResp != nil {
		log.Printf("[i] cache hit: %s", r.Question[0].String())
		resp := cachedResp.Value()
		resp.Id = r.Id
		w.WriteMsg(resp)
		return
	}
	if utils.Contains(blockList, r.Question[0].Name) {
		log.Printf("[x] blocked %s\n", r.Question[0].Name)
		record, err := dns.NewRR(fmt.Sprintf("%s A %s", r.Question[0].Name, "0.0.0.0"))
		if err != nil {
			log.Printf("[!] error creating record: %s\n", err)
			return
		}
		r.Answer = []dns.RR{
			record,
		}
		r.Rcode = dns.RcodeSuccess
		log.Println(r.Answer[0].String())
		w.WriteMsg(r)
		return
	}
	c := &dns.Client{Net: transport}
	resp, _, err := c.Exchange(r, "1.1.1.1:53")
	if err != nil {
		log.Printf("[x] error: %s\n", err.Error())
		dns.HandleFailed(w, r)
		return
	}
	cache.Set(r.Question[0].String(), resp, 1)
	log.Println(resp.Answer[0].String())
	log.Printf("[i] cache miss: %s", r.Question[0].String())
	w.WriteMsg(resp)
}

func main() {
	// attach request handler func
	dns.HandleFunc(".", handleDnsRequest)

	// start server
	port := 53
	server := &dns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}
	log.Printf("[i] spinning up ergosphere on port %d\n", port)
	err := server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}
}