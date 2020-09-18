package consul

import (
	"crypto/x509"
	"strings"
)

type Config struct {
	GatewayName    string
	GatewayID      string
	GatewayAddress string
	GatewayPort    int
	CAsPool        *x509.CertPool
	CAs            [][]byte
	Services       []Service
}

type Service struct {
	Name       string
	Instances  []*Instance
	Intentions []string
	ProxyTLS   *ProxyTLS
	TLS
}

type Instance struct {
	ID      string
	Address string
	Port    int
}

type TLS struct {
	Cert []byte
	Key  []byte
	CAs  [][]byte
}

type ProxyTLS struct {
	CAFile   string `json:",omitempty"`
	CertFile string `json:",omitempty"`
	KeyFile  string `json:",omitempty"`
	SNI      string `json:",omitempty"`
}

func (t *TLS) CertString() string {
	return string(t.Cert)
}

func (t *TLS) KeyString() string {
	return string(t.Key)
}

func (t *TLS) CAsString() string {
	if len(t.CAs) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, ca := range t.CAs {
		_, _ = builder.WriteString(string(ca))
	}

	return builder.String()
}

func NewService(svc *service) Service {
	downstream := Service{
		Name: svc.name,
		ProxyTLS: &ProxyTLS{
			CAFile:   svc.gatewayService.CAFile,
			CertFile: svc.gatewayService.CertFile,
			KeyFile:  svc.gatewayService.KeyFile,
			SNI:      svc.gatewayService.SNI,
		},
		TLS: TLS{
			Cert: svc.leaf.Cert,
			Key:  svc.leaf.Key,
		},
	}
	for _, i := range svc.instances {
		newInstance := &Instance{
			ID:   i.Service.ID,
			Port: i.Service.Port,
		}
		if i.Service.Address == "" {
			newInstance.Address = i.Node.Address
		} else {
			newInstance.Address = i.Service.Address
		}
		downstream.Instances = append(downstream.Instances, newInstance)
	}

	for _, i := range svc.intentions {
		if i.Action == "allow" {
			downstream.Intentions = append(downstream.Intentions, i.SourceName)
		}
	}
	return downstream
}
