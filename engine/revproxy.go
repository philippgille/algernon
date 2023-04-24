package engine

import (
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
	"github.com/xyproto/algernon/utils"
)

type ReverseProxy struct {
	PathPrefix string
	Endpoint   url.URL
}

func (rp *ReverseProxy) DoProxyPass(req http.Request) (*http.Response, error) {
	client := &http.Client{}
	endpoint := rp.Endpoint
	req.RequestURI = ""
	req.URL.Path = req.URL.Path[len(rp.PathPrefix):]
	req.URL.Scheme = endpoint.Scheme
	req.URL.Host = endpoint.Host
	res, err := client.Do(&req)
	if err != nil {
		log.Errorf("reverse proxy error: %v\n", err)
		return nil, err
	}
	return res, nil
}

type ReverseProxyConfig struct {
	ReverseProxies []ReverseProxy
	proxyMatcher   utils.PrefixMatch
	prefix2rproxy  map[string]int
}

func NewReverseProxyConfig() *ReverseProxyConfig {
	return &ReverseProxyConfig{}
}

// Add a ReverseProxy and also initialize the internal proxy matcher
func (rc *ReverseProxyConfig) Add(rp *ReverseProxy) {
	rc.ReverseProxies = append(rc.ReverseProxies, *rp)
	rc.Init()
}

func (rc *ReverseProxyConfig) Init() {
	keys := make([]string, 0, len(rc.ReverseProxies))
	rc.prefix2rproxy = make(map[string]int)
	for i, rp := range rc.ReverseProxies {
		keys = append(keys, rp.PathPrefix)
		rc.prefix2rproxy[rp.PathPrefix] = i
	}
	rc.proxyMatcher.Build(keys)
}

func (rc *ReverseProxyConfig) FindMatchingReverseProxy(path string) *ReverseProxy {
	matches := rc.proxyMatcher.Match(path)
	if len(matches) == 0 {
		return nil
	}
	if len(matches) > 1 {
		log.Warnf("found more than one reverse proxy for `%s`: %+v. returning the longest", matches, path)
	}
	var match *ReverseProxy
	maxlen := 0
	for _, prefix := range matches {
		if len(prefix) > maxlen {
			maxlen = len(prefix)
			match = &rc.ReverseProxies[rc.prefix2rproxy[prefix]]
		}
	}
	return match
}