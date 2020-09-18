package gateway

import (
	"encoding/json"

	"github.com/f5devcentral/bigip-tgw/as3"
	"github.com/f5devcentral/bigip-tgw/consul"
	slog "github.com/go-eden/slf4go"
)

var log = slog.NewLogger("f5-writer")
var iruleEncoded = "d2hlbiBSVUxFX0lOSVQgewogICAgI3NldCBzdGF0aWM6OnNiX2RlYnVnIHRvIDIgaWYgeW91IHdhbnQgdG8gZW5hYmxlIGxvZ2dpbmcgdG8gdHJvdWJsZXNob290IHRoaXMgaVJ1bGUsIDEgZm9yIGluZm9ybWF0aW9uYWwgbWVzc2FnZXMsIG90aGVyd2lzZSBzZXQgdG8gMAogICAgc2V0IHN0YXRpYzo6c2JfZGVidWcgMgogICAgaWYgeyRzdGF0aWM6OnNiX2RlYnVnID4gMX0geyBsb2cgbG9jYWwwLiAicnVsZSBpbml0IiB9Cn0KCndoZW4gQ0xJRU5UU1NMX0NMSUVOVENFUlQgewogICAgaWYgeyRzdGF0aWM6OnNiX2RlYnVnID4gMX0ge2xvZyBsb2NhbDAuICJJbiBDTElFTlRTU0xfQ0xJRU5UQ0VSVCJ9CgogICAgc2V0IGNsaWVudF9jZXJ0IFtTU0w6OmNlcnQgMF0KICAKICAgIHNldCBzZXJpYWxfaWQgIiIKICAgIHNldCBzcGlmZmUgIiIKICAgIHNldCBsb2dfcHJlZml4ICJbSVA6OnJlbW90ZV9hZGRyXTpbVENQOjpyZW1vdGVfcG9ydCBjbGllbnRzaWRlXSBbSVA6OmxvY2FsX2FkZHJdOltUQ1A6OmxvY2FsX3BvcnQgY2xpZW50c2lkZV0iCgogICAgaWYgeyBbU1NMOjpjZXJ0IGNvdW50XSA+IDAgfSB7CiAgICAgICAgc2V0IHNwaWZmZSBbZmluZHN0ciBbWDUwOTo6ZXh0ZW5zaW9ucyBbU1NMOjpjZXJ0IDBdXSAiU3ViamVjdCBBbHRlcm5hdGl2ZSBOYW1lIiAzOSAiLCJdCiAgICAgICAgaWYgeyRzdGF0aWM6OnNiX2RlYnVnID4gMX0geyBsb2cgbG9jYWwwLiAiPCRsb2dfcHJlZml4PjogU0FOOiAkc3BpZmZlIn0KICAgICAgICBzZXQgc2VyaWFsX2lkIFtYNTA5OjpzZXJpYWxfbnVtYmVyICRjbGllbnRfY2VydF0KICAgICAgICBpZiB7JHN0YXRpYzo6c2JfZGVidWcgPiAxfSB7IGxvZyBsb2NhbDAuICI8JGxvZ19wcmVmaXg+OiBTZXJpYWxfSUQ6ICRzZXJpYWxfaWQifQogICAgfQogICAgaWYgeyRzdGF0aWM6OnNiX2RlYnVnID4gMX0geyBsb2cgbG9jYWwwLmluZm8gImhlcmUgaXMgc3BpZmZlOiAkc3BpZmZlIiB9CiAgICAgICAjcmVnZXhwIHteLipcL3tbYS16QS1aMC05XC1dKn19ICRzcGlmZmUgc3BpZmZlX3Jlc3VsdAogICAgc2V0IHNwaWZmZV9yZXN1bHQgW2dldGZpZWxkICRzcGlmZmUgIi8iIDldCiAgICBsb2cgbG9jYWwwLiAic3BpZmZlX3Jlc3VsdCArKysrKysrKysrKysrIGlzICRzcGlmZmVfcmVzdWx0IgogICAgc2V0IHRyaW1zcGlmZmUgW3N0cmluZyB0cmltICRzcGlmZmVfcmVzdWx0XQp9IAoKd2hlbiBDTElFTlRTU0xfSEFORFNIQUtFIHsKICAgIGlmIHsgW1NTTDo6ZXh0ZW5zaW9ucyBleGlzdHMgLXR5cGUgMF0gfSB7CiAgICAgICBiaW5hcnkgc2NhbiBbU1NMOjpleHRlbnNpb25zIC10eXBlIDBdIHtAOUEqfSBzbmlfbmFtZQogICAgICAgaWYgeyRzdGF0aWM6OnNiX2RlYnVnID4gMX0geyBsb2cgbG9jYWwwLiAic25pIG5hbWU6ICR7c25pX25hbWV9In0KICAgICAgIHJlZ2V4cCB7W14uXSp9ICRzbmlfbmFtZSBzbmlfcmVzdWx0CiAgICAgICBsb2cgbG9jYWwwLiAicmVzdWx0IGlzICRzbmlfcmVzdWx0IgogICAgfQoKICAgICMgdXNlIHRoZSB0ZXJuYXJ5IG9wZXJhdG9yIHRvIHJldHVybiB0aGUgc2VydmVybmFtZSBjb25kaXRpb25hbGx5CiAgICBpZiB7JHN0YXRpYzo6c2JfZGVidWcgPiAxfSB7IGxvZyBsb2NhbDAuICJzbmkgbmFtZTogW2V4cHIge1tpbmZvIGV4aXN0cyBzbmlfbmFtZV0gPyAke3NuaV9uYW1lfSA6IHtub3QgZm91bmR9IH1dIn0gICAgCiAgICAKICAgIHNldCBrZXkgW2NvbmNhdCAkdHJpbXNwaWZmZTokc25pX3Jlc3VsdF0KICAgIGxvZyBsb2NhbDAuICJoZXJlIGlzIHRoZSBrZXkgIC4uLi4gJGtleSIKICAgIGxvZyBsb2NhbDAuaW5mbyAidGFyZ2V0LWRnOiBbY2xhc3MgZ2V0IHRhcmdldC1kZ10iCiAgICBTU0w6OmhhbmRzaGFrZSBob2xkCiAgICBpZiB7W2NsYXNzIG1hdGNoICRrZXkgZXF1YWxzICJ0YXJnZXQtZGciXSB9IHsKICAgICAgICBsb2cgbG9jYWwwLiAic3VjY2VzcyIKICAgICAgICBzZXQgZ290U05JdmFsdWUgW2NsYXNzIG1hdGNoIC12YWx1ZSAiJGtleSIgZXF1YWxzICJ0YXJnZXQtZGciXQogICAgICAgIGxvZyBsb2NhbDAuICJ2YWx1ZSBpcyAkZ290U05JdmFsdWUiCiAgICB9CiAgICBlbHNlIHsKICAgICAgICBsb2cgbG9jYWwwLiAiU05JIG5vdCBpbiB0aGUgZGF0YSBncm91cCIKICAgICAgICByZWplY3QKICAgIH0KICAgIAogICAgaWYgeyAkZ290U05JdmFsdWUgZXEgImFsbG93IiB9IHRoZW4gewogICAgICAgIGxvZyBsb2NhbDAuICJ3ZSBhcmUgZ29vZCBwbGVhc2UgcHJvY2VlZCIKICAgICAgICBpZiB7JHN0YXRpYzo6c2JfZGVidWcgPiAxfSB7bG9nIGxvY2FsMC4gIklzIHRoZSBjb25uZWN0aW9uIGF1dGhvcml6ZWQ6ICRrZXkifQogICAgICAgIFNTTDo6aGFuZHNoYWtlIHJlc3VtZSAKICAgIH0KICAgIGVsc2UgewogICAgICAgIGlmIHskc3RhdGljOjpzYl9kZWJ1ZyA+IDF9IHtsb2cgbG9jYWwwLiAiQ29ubmVjdGlvbiBpcyBub3QgYXV0aG9yaXplZDogJGtleSJ9CiAgICAgICAgcmVqZWN0CiAgICB9Cn0="
var teemUAgent = "TGW Configured AS3"

type Bigip struct {
	Config as3.Params
	//Session bigip.BigIP
	CfgC    chan consul.Config
	ReqChan chan as3.AS3Config

	AS3Config *as3.AS3Config
}

func New(c as3.Params, watcherChan chan consul.Config, reqChan chan as3.AS3Config) *Bigip {
	log.Info("[INIT] Creating AS3 writer")

	return &Bigip{
		Config:  c,
		CfgC:    watcherChan,
		ReqChan: reqChan,
	}
}

func (f5 *Bigip) DeInit() error {
	close(f5.CfgC)
	close(f5.ReqChan)
	return nil
}

func (f5 *Bigip) Deploy(req as3.AS3Config) error {
	msgReq := req
	log.Info("[INFO] sending config to agent")
	select {
	case f5.ReqChan <- msgReq:
	case <-f5.ReqChan:
		f5.ReqChan <- msgReq
	}
	return nil
}

func (f5 *Bigip) Run() error {
	//go func() {
	for c := range f5.CfgC {
		log.Info("[INFO] Writer received configuration change")

		//Construct New AS3 Config
		f5.makeAppMap(c)

		//Get AS3 JSON from Structs
		jsonObj, err := json.Marshal(f5.AS3Config)
		if err != nil {
			log.Error(err)
		}
		f5.AS3Config.JsonObj = string(jsonObj)

		log.Debugf("[DEBUG] AS3 Declaration: %v", string(jsonObj))

		f5.Deploy(*f5.AS3Config)
	}
	return nil
}

func (f5 *Bigip) newAS3Config() *as3.AS3Config {
	stubConfig := as3.AS3Config{
		Schema:  f5.Config.Schema,
		Class:   "AS3",
		Action:  "deploy",
		Persist: true,
		Declaration: as3.Declaration{
			Class:         "ADC",
			SchemaVersion: f5.Config.SchemaVersion,
			Controls: &as3.Controls{
				Class:     "Controls",
				UserAgent: teemUAgent,
			},
			Tenant: as3.Tenant{
				Class:              "Tenant",
				DefaultRouteDomain: 0,
				Application:        make(map[string]interface{}),
			},
		},
	}

	return &stubConfig
}

func (f5 *Bigip) makeAppMap(c consul.Config) {
	f5.AS3Config = f5.newAS3Config()
	f5.AS3Config.Declaration.Tenant.Application["class"] = "Application"
	f5.AS3Config.Declaration.Tenant.Application["template"] = "generic"

	vServer := makeVserver(c)
	f5.AS3Config.Declaration.Tenant.Application[vServer.Name] = vServer

	pools := makePools(c)
	for _, p := range pools {
		f5.AS3Config.Declaration.Tenant.Application[p.Name] = p
	}

	CAs := makeCAs(c)
	f5.AS3Config.Declaration.Tenant.Application[CAs.Name] = CAs

	serverTLS := makeServerTLS(c)
	f5.AS3Config.Declaration.Tenant.Application[serverTLS.Name] = serverTLS

	//proxyTLS := makeProxyTLS(c)
	//for _, p := range proxyTLS {
	//	nextAS3.Declaration.Tenant.Application[p.Name] = p
	//}
	certs := makeCerts(c)
	for _, c := range certs {
		f5.AS3Config.Declaration.Tenant.Application[c.Name] = c
	}

	policy := makePolicies(c)
	f5.AS3Config.Declaration.Tenant.Application[policy.Name] = policy

	iRules := makeIRules(c)
	for _, i := range iRules {
		f5.AS3Config.Declaration.Tenant.Application[i.Name] = i
	}

	datagroups := makeDatagroups(c)
	for _, d := range datagroups {
		f5.AS3Config.Declaration.Tenant.Application[d.Name] = d
	}
}

func makePools(c consul.Config) []as3.Pool {
	pools := []as3.Pool{}

	for _, s := range c.Services {
		poolx := newPool()
		poolx.Name = s.Name + "-pool"

		// Add Pool Members
		for _, i := range s.Instances {
			poolx.Members = append(poolx.Members, as3.Member{
				ServicePort:     i.Port,
				ServerAddresses: []string{i.Address},
			})
		}

		pools = append(pools, *poolx)
	}
	return pools
}
func makePolicies(c consul.Config) as3.PolicyEndpoint {
	mySNI := as3.PolicyEndpoint{
		Name:  "SNIrouting",
		Class: "Endpoint_Policy",
		Label: "SNI Routing",
	}

	for _, s := range c.Services {
		myRule := &as3.PolicyRule{
			Name: "forward_to_" + s.Name,
		}
		myCondition := &as3.Condition{
			Type:       "sslExtension",
			Event:      "ssl-client-hello",
			Normalized: false,
		}
		myCondition.ServerName = &as3.PolicyCompareString{}
		myCondition.ServerName.Operand = "starts-with"
		myCondition.ServerName.CaseSensitive = false
		myCondition.ServerName.Values = append(myCondition.ServerName.Values, s.Name)
		myRule.Conditions = append(myRule.Conditions, myCondition)

		myAction := &as3.Action{
			Type:  "forward",
			Event: "ssl-client-hello",
		}
		myAction.Select = &as3.ActionForwardSelect{}
		myAction.Select.Pool = &as3.ResourcePointer{}
		myAction.Select.Pool.Use = s.Name + "-pool"
		myRule.Actions = append(myRule.Actions, myAction)
		mySNI.Rules = append(mySNI.Rules, myRule)
	}
	return mySNI
}
func makeServerTLS(c consul.Config) as3.ServerTLS {
	var server = as3.ServerTLS{
		Name:                    "webtls",
		Class:                   "TLS_Server",
		Label:                   "TLS Termination",
		AuthenticationMode:      "require",
		AuthenticationFrequency: "every-time",
		AuthenticationTrustCA:   "cabundle",
	}

	for _, s := range c.Services {
		server.Certificates = append(server.Certificates, as3.CertName{
			Certificate: s.Name + "-cert",
		})
	}
	return server
}

func makeProxyTLS(c consul.Config) []as3.ClientTLS {
	proxyTLS := []as3.ClientTLS{}
	for _, s := range c.Services {
		if s.ProxyTLS != nil {
			myProxyTLS := &as3.ClientTLS{
				Name:                s.Name + "-proxytls",
				Class:               "TLS_Client",
				SendSNI:             s.ProxyTLS.SNI,
				ValidateCertificate: true,
				TrustCA:             s.ProxyTLS.CAFile,
				IgnoreExpired:       false,
				IgnoreUntrusted:     false,
				ClientCertificate:   s.Name + "-proxycert",
			}
			proxyTLS = append(proxyTLS, *myProxyTLS)
		}
	}
	return proxyTLS
}

func makeCerts(c consul.Config) []as3.Certificate {
	var certs = []as3.Certificate{}

	for _, s := range c.Services {
		newCert := as3.Certificate{
			Name:        s.Name + "-cert",
			Class:       "Certificate",
			Certificate: s.CertString(),
			PrivateKey:  s.KeyString(),
			ChainCA:     s.CAsString(),
		}
		certs = append(certs, newCert)
	}
	return certs
}

func addCert(s consul.Service) *as3.Certificate {
	return &as3.Certificate{
		Class:       "Certificate",
		Certificate: s.CertString(),
		PrivateKey:  s.KeyString(),
		ChainCA:     s.CAsString(),
	}
}

func makeIRules(c consul.Config) []as3.IRule {
	var iRules []as3.IRule
	intentionRule := as3.IRule{
		Name:  "intentionRule",
		Class: "iRule",
	}
	intentionRule.IRule = &as3.ResourcePointer{}
	intentionRule.IRule.Base64 = iruleEncoded
	iRules = append(iRules, intentionRule)
	return iRules
}
func makeDatagroups(c consul.Config) []as3.DataGroup {
	var datagroups []as3.DataGroup
	intentions := as3.DataGroup{
		Class:       "Data_Group",
		StorageType: "internal",
		Name:        "target-dg",
		KeyDataType: "string",
	}
	intentions.Records = append(intentions.Records, &as3.Record{
		Key:   "dummy",
		Value: "disallow",
	})

	for _, s := range c.Services {
		for _, i := range s.Intentions {
			intentions.Records = append(intentions.Records, &as3.Record{
				Key:   i + ":" + s.Name,
				Value: "allow",
			})
		}
	}
	datagroups = append(datagroups, intentions)
	return datagroups
}
func makeVserver(c consul.Config) *as3.Service {
	stubVserver := as3.Service{
		Name:           "TG_Vserver",
		Class:          "Service_TCP",
		ServerTLS:      "webtls",
		PolicyEndpoint: "SNIrouting",
	}
	stubVserver.IRules = append(stubVserver.IRules, "intentionRule")
	stubVserver.VirtualAddresses = append(stubVserver.VirtualAddresses, c.GatewayAddress)
	stubVserver.VirtualPort = c.GatewayPort
	return &stubVserver
}

func makeCAs(c consul.Config) as3.CABundle {
	CA := as3.CABundle{
		Name:   "cabundle",
		Class:  "CA_Bundle",
		Bundle: "",
	}
	for _, cert := range c.CAs {
		CA.Bundle += string(cert)
	}
	return CA
}
func newCA(caList [][]byte) *as3.CABundle {
	stubCA := as3.CABundle{
		Class:  "CA_Bundle",
		Bundle: "",
	}
	for _, cert := range caList {
		stubCA.Bundle += string(cert)
	}
	return &stubCA
}

func newServerTLS() *as3.ServerTLS {
	return &as3.ServerTLS{
		Class:                   "TLS_Server",
		Label:                   "TLS Termination",
		AuthenticationMode:      "require",
		AuthenticationFrequency: "every-time",
		AuthenticationTrustCA:   "cabundle",
	}
}

func newPool() *as3.Pool {
	return &as3.Pool{
		Class:    "Pool",
		Monitors: []as3.ResourcePointer{},
		Members:  []as3.Member{},
	}
}
