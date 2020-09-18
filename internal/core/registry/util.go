package registry

import (
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"

	"crypto/tls"
	"fmt"

	"github.com/leon-yc/ggs/internal/core/common"
	"github.com/leon-yc/ggs/internal/core/config/model"
	ggsTLS "github.com/leon-yc/ggs/internal/core/tls"
	"github.com/leon-yc/ggs/internal/pkg/util/iputil"
	"github.com/leon-yc/ggs/pkg/qlog"
	"github.com/cenkalti/backoff"
)

const protocolSymbol = "://"

//GetProtocolMap returns the protocol map
func GetProtocolMap(eps []string) (map[string]string, string) {
	m := make(map[string]string)
	var p string
	for _, ep := range eps {
		u, err := url.Parse(ep)
		if err != nil {
			qlog.Errorf("Can not parse %s: %s", ep, err.Error())
			continue
		}
		proto := u.Scheme
		ipPort := u.Host
		if proto == "" {
			m["unknown"] = ipPort
		} else {
			m[proto] = ipPort
			p = proto
		}
	}
	return m, p
}

//GetProtocolList returns the protocol list
func GetProtocolList(m map[string]string) []string {
	eps := []string{}
	for p, ep := range m {
		uri := p + protocolSymbol + ep
		eps = append(eps, uri)
	}
	return eps
}

//MakeEndpoints returns the endpoints
func MakeEndpoints(m map[string]model.Protocol) []string {
	var eps = make([]string, 0)
	for name, protocol := range m {
		ep := protocol.Advertise
		if ep == "" {
			if protocol.Listen != "" {
				ep = protocol.Listen
			} else {
				ep = iputil.DefaultEndpoint4Protocol(name)
			}
		}
		ep = strings.Join([]string{name, ep}, protocolSymbol)
		eps = append(eps, ep)
	}
	return eps
}

//MakeEndpointMap returns the endpoints map
func MakeEndpointMap(m map[string]model.Protocol) (map[string]string, error) {
	eps := make(map[string]string, 0)
	for name, protocol := range m {
		ep := protocol.Listen
		if len(protocol.Advertise) > 0 {
			ep = protocol.Advertise
		}

		host, port, err := net.SplitHostPort(ep)
		if err != nil {
			return nil, err
		}
		if host == "" || port == "" {
			return nil, fmt.Errorf("listen address is invalid [%s]", protocol.Listen)
		}

		ip, err := fillUnspecifiedIP(host)
		if err != nil {
			return nil, err
		}
		eps[name] = net.JoinHostPort(ip, port)
	}
	return eps, nil
}

// fillUnspecifiedIP replace 0.0.0.0 or :: IPv4 and IPv6 unspecified IP address with local NIC IP.
func fillUnspecifiedIP(host string) (string, error) {
	var addr string
	ip := net.ParseIP(host)
	if ip == nil {
		return "", fmt.Errorf("invalid IP address %s", host)
	}

	addr = host
	if ip.IsUnspecified() {
		if iputil.IsIPv6Address(ip) {
			addr = iputil.GetLocalIPv6()
		} else {
			addr = iputil.GetLocalIP()
		}
		if len(addr) == 0 {
			return addr, fmt.Errorf("failed to get local IP address")
		}
	}
	return addr, nil
}

//Microservice2ServiceKeyStr prepares a microservice key
func Microservice2ServiceKeyStr(m *MicroService) string {
	return strings.Join([]string{m.ServiceName, m.Version, m.AppID}, ":")
}

const (
	initialInterval = 5 * time.Second
	maxInterval     = 3 * time.Minute
)

func startBackOff(operation func() error) {
	backOff := &backoff.ExponentialBackOff{
		InitialInterval:     initialInterval,
		MaxInterval:         maxInterval,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		Clock:               backoff.SystemClock,
	}
	for {
		qlog.Infof("start backoff with initial interval %v", initialInterval)
		err := backoff.Retry(operation, backOff)
		if err == nil {
			return
		}
	}
}

//URIs2Hosts return hosts and scheme
func URIs2Hosts(uris []string) ([]string, string, error) {
	hosts := make([]string, 0)
	var scheme string
	var hostPortRegex = "(\\.*://.*):(\\d*)\\/?(.*)"
	for _, addr := range uris {
		ok, err := regexp.MatchString(hostPortRegex, addr)
		if err != nil {
			return nil, "", err
		}
		if ok {
			u, e := url.Parse(addr)
			if e != nil {
				qlog.Warn("registry address is invalid:" + addr)
				continue
			}
			if len(u.Host) == 0 {
				continue
			}
			if len(scheme) != 0 && u.Scheme != scheme {
				return nil, "", fmt.Errorf("inconsistent scheme found in registry address")
			}
			scheme = u.Scheme
			hosts = append(hosts, u.Host)
		} else {
			//not uri. but still permitted, like zookeeper,file system
			hosts = append(hosts, addr)
			continue
		}
	}
	return hosts, scheme, nil
}
func getTLSConfig(scheme, t string) (*tls.Config, error) {
	var tlsConfig *tls.Config
	secure := scheme == common.HTTPS
	if secure {
		sslTag := t + "." + common.Consumer
		tmpTLSConfig, sslConfig, err := ggsTLS.GetTLSConfigByService(t, "", common.Consumer)
		if err != nil {
			if ggsTLS.IsSSLConfigNotExist(err) {
				tmpErr := fmt.Errorf("%s tls mode, but no ssl config", sslTag)
				qlog.Error(tmpErr.Error() + ", err: " + err.Error())
				return nil, tmpErr
			}
			qlog.Errorf("Load %s TLS config failed: %s", err)
			return nil, err
		}
		qlog.Warnf("%s TLS mode, verify peer: %t, cipher plugin: %s.",
			sslTag, sslConfig.VerifyPeer, sslConfig.CipherPlugin)
		tlsConfig = tmpTLSConfig
	}
	return tlsConfig, nil
}
