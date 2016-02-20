package liberty

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gnanderson/trie"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

// Proxy defines the configuration of a reverse proxy entry in the router.
type Proxy struct {
	HostPath    string `yaml:"hostPath"`
	RemoteHost  string `yaml:"remoteHost"`
	rHostIPs    []net.IP
	HostAlias   []string `yaml:"hostAlias"`
	HostIP      string   `yaml:"hostIP"`
	HostPort    int      `yaml:"hostPort"`
	Tls         bool     `yaml:"tls"`
	TlsRedirect bool     `yaml:"tlsRedirect"`
	Ws          bool     `yaml:"ws"`
	HandlerType string   `yaml:"handlerType"`
	IPs         []string `yaml:"ips, flow"`
	Cors        []string `yaml:"cors, flow"`
}

var muxers map[int]*http.ServeMux

func getMux(port int) *http.ServeMux {
	if muxers == nil {
		muxers = make(map[int]*http.ServeMux)
	}
	if m, ok := muxers[port]; ok {
		return m
	}
	muxers[port] = http.NewServeMux()
	return muxers[port]
}

// Configure a proxy for use with the paramaters from the parsed yaml config. If
// a remote host resolves to more than one IP address, we'll create a server and
// for each. This works because under the hood we're using SO_REUSEPORT.
//
// It's important to note at the moment there's nothing to stop you serving
// different back ends from the same 'hostpath'. This will lead to unpredictable
// results so please ensure you only use identical backends for each proxy
// entry/hostpath
func (p *Proxy) configure() ([]*http.Server, error) {
	var servers []*http.Server

	// In order to avoid ambiguity,  each entry should have a port to listen on.
	switch {
	case !p.Tls && p.HostPort == 0:
		p.HostPort = 80
	case p.Tls && p.HostPort == 0:
		p.HostPort = 443
	}

	if p.HostIP == "" {
		p.HostIP = "0.0.0.0"
	}

	// lookup the backend host and create a server/mux combination for each
	// resolved IP address if this remote host is not a valid resource we can't
	// continue.
	// TODO skip this proxy if error here?
	remote, err := url.Parse(p.RemoteHost)
	if err != nil {
		panic(fmt.Sprintf("Invalid proxy host: %s", err))
	}

	// add an additional redirect from port 80
	if p.TlsRedirect && p.HostPort == 443 {
		s := &http.Server{
			Addr: fmt.Sprintf("%s:80", p.HostIP),
		}
		m := getMux(80)
		m.HandleFunc(p.HostPath, redir)
		s.Handler = m
		servers = append(servers, s)
	}

	chunks := strings.Split(remote.Host, ":")
	var hostName string
	if len(chunks) > 1 {
		hostName = chunks[0]
	} else {
		hostName = remote.Host
	}
	ips, err := net.LookupIP(hostName)
	if err != nil {
		return nil, err
	}

	// now the server (or servers) for this proxy entry
	for _, ip := range ips {
		s := &http.Server{
			Addr: fmt.Sprintf("%s:%d", p.HostIP, p.HostPort),
		}
		mux := getMux(p.HostPort)

		if p.Tls {
			setTLSConfig(s)
		}

		// if this is a websocket proxy, we need to hijack the connection. We'll
		// have to treat this a little differently.
		if p.Ws {
			mux.Handle(p.HostPath, websocketProxy(p.RemoteHost))
			s.Handler = mux
			servers = append(servers, s)
			continue
		}

		// now configure the reverse proxy
		/*m, ok := s.Handler.(*http.ServeMux)
		if !ok {
			return nil, fmt.Errorf("cannot configure reverse proxy for '%s'", p.HostPath)
		}
		*/
		reverseProxy(p, mux, strings.Replace(p.RemoteHost, remote.Host, ip.String(), 1))

		servers = append(servers, s)
	}

	return servers, nil
}

// build a chain of handlers, with the last one actually performing the reverse
// proxy to the remote resource.
func reverseProxy(p *Proxy, mux *http.ServeMux, remoteUrl string) {

	// if this remote host is not a valid resource we can't continue
	remote, err := url.Parse(remoteUrl)
	if err != nil {
		glog.Fatalf("Invalid proxy host: %s", err)
		panic(err)
	}

	// the first handler should be a prometheus instrumented handler
	handlers := make([]Chainable, 0)
	handlers = append(handlers, &InstrumentedHandler{Name: p.HostPath})

	// next we check for restrictions based on location / IP
	if len(p.IPs) > 0 {
		nets := ips2nets(p.IPs)
		restricted := &IPRestrictedHandler{Allowed: nets}
		restricted.handlerType = p.HandlerType

		// if this is also an API handler, pass in the open paths
		if restricted.handlerType == apiHandler {
			restricted.openPaths = &trie.Trie{}
			for _, wl := range conf.Whitelist {
				restricted.openPaths.Put(wl.Path, true)
			}
		}
		handlers = append(handlers, restricted)
	}

	// use a standard library reverse proxy, but use our own transport so that
	// we can further update the response
	reverseProxy := httputil.NewSingleHostReverseProxy(remote)
	reverseProxy.Transport = &Transport{
		tr:   http.DefaultTransport,
		tls:  p.Tls,
		cors: p.Cors,
	}

	// now we should decided what type of resource the request is for, there's
	// only really three basic types at the moment: web, api, metrics
	var final http.Handler
	switch p.HandlerType {
	default:
		final = reverseProxy
	case apiHandler:
		handlers = append(handlers, NewApiHandler(p))
		final = reverseProxy
	case promHandler:
		final = prometheus.InstrumentHandler(hostname, prometheus.Handler())
	case redirectHandler:
		final = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, fmt.Sprintf("%s/%s", p.RemoteHost, r.URL.Path), 301)
		})
	}

	// link the handler chain
	chain := NewChain(handlers...).Link(final)
	if len(p.HostAlias) > 0 {
		chunks := strings.Split(p.HostPath, ".")
		for _, alias := range p.HostAlias {
			chunks[0] = alias
			mux.Handle(strings.Join(chunks, "."), chain)
		}
	} else {
		mux.Handle(p.HostPath, chain)
	}
}

// convert a list of IP address strings in CIDR format to IPNets
func ips2nets(ips []string) []*net.IPNet {
	nets := make([]*net.IPNet, 0)
	for _, ipRange := range ips {
		_, ipNet, err := net.ParseCIDR(ipRange)
		if err != nil {
			glog.Fatalf("Invalid proxy host: %s", err)
			panic(err)
		}
		nets = append(nets, ipNet)
	}
	return nets
}
