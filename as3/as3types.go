package as3

type (
	AS3Config struct {
		Schema      string      `json:"$schema"`
		Class       string      `json:"class"`
		Action      string      `json:"action"`
		Persist     bool        `json:"persist"`
		Declaration Declaration `json:"declaration"`
		JsonObj     string      `json:"-"`
	}

	Declaration struct {
		Class         string    `json:"class"`
		SchemaVersion string    `json:"schemaVersion"`
		ID            string    `json:"id"`
		Label         string    `json:"label"`
		Remark        string    `json:"remark"`
		Controls      *Controls `json:"controls,omitempty"`
		Tenant        Tenant    `json:"TGW_Tenant"`
	}

	Controls struct {
		Class     string `json:"class"`
		UserAgent string `json:"userAgent"`
	}

	Tenant struct {
		Name               string                 `json:"-"`
		Class              string                 `json:"class"`
		DefaultRouteDomain int                    `json:"defaultRouteDomain"`
		Application        map[string]interface{} `json:"TermatingGateway"`
	}

	Service struct {
		Name                   string   `json:"-"`
		Class                  string   `json:"class"`
		Layer4                 string   `json:"layer4,omitempty"`
		Source                 string   `json:"source,omitempty"`
		TranslateServerAddress bool     `json:"translateServerAddress,omitempty"`
		TranslateServerPort    bool     `json:"translateServerPort,omitempty"`
		ServerTLS              string   `json:"serverTLS,omitempty"`
		ClientTLS              string   `json:"clientTLS,omitempty"`
		VirtualPort            int      `json:"virtualPort"`
		VirtualAddresses       []string `json:"virtualAddresses"`
		SNAT                   string   `json:"snat,omitempty"`
		Pool                   string   `json:"pool,omitempty"`
		PersistenceMethods     []string `json:"persistenceMethods,omitempty"`
		PolicyEndpoint         string   `json:"policyEndpoint,omitempty"`
		IRules                 []string `json:"iRules,omitempty"`
		Redirect80             *bool    `json:"redirect80,omitempty"`
	}

	Pool struct {
		Name                 string            `json:"-"`
		Class                string            `json:"class"`
		Label                string            `json:"label"`
		Remark               string            `json:"remark,omitempty"`
		Members              []Member          `json:"members"`
		Monitors             []ResourcePointer `json:"monitors"`
		LoadBalancingMode    string            `json:"loadBalancingMode,omitempty"`
		MinimumMembersActive int               `json:"minimumMembersActive,omitempty"`
		ReselectTries        int               `json:"reselectTries,omitempty"`
		ServiceDownAction    string            `json:"serviceDownAction,omitempty"`
		SlowRampTime         int               `json:"slowRampTime,omitempty"`
		MinimumMonitors      int               `json:"minimumMonitors,omitempty"`
	}

	Member struct {
		ServicePort      int      `json:"servicePort"`
		ServerAddresses  []string `json:"serverAddresses"`
		AddressDiscovery string   `json:"addressDiscovery,omitempty"`
	}

	Monitor struct {
		Class             string  `json:"class,omitempty"`
		Interval          int     `json:"interval,omitempty"`
		MonitorType       string  `json:"monitorType,omitempty"`
		TargetAddress     *string `json:"targetAddress,omitempty"`
		Timeout           int     `json:"timeout,omitempty"`
		TimeUnitilUp      *int    `json:"timeUntilUp,omitempty"`
		Adaptive          *bool   `json:"adaptive,omitempty"`
		Dscp              *int    `json:"dscp,omitempty"`
		Receive           string  `json:"receive,omitempty"`
		Send              string  `json:"send,omitempty"`
		TargetPort        *int    `json:"targetPort,omitempty"`
		ClientCertificate string  `json:"clientCertificate,omitempty"`
		Ciphers           string  `json:"ciphers,omitempty"`
	}

	ServerTLS struct {
		Name                    string           `json:"-"`
		Class                   string           `json:"class"`
		Label                   string           `json:"label"`
		Remark                  string           `json:"remark,omitempty"`
		Certificates            []CertName       `json:"certificates"`
		RequireSNI              string           `json:"requireSNI,omitempty"`
		Ciphers                 string           `json:"ciphers,omitempty"`
		CipherGroup             *ResourcePointer `json:"cipherGroup,omitempty"`
		Tls1_3Enabled           bool             `json:"tls1_3Enabled,omitempty"`
		RenegotiationEnabled    *bool            `json:"renegotiationEnabled,omitempty"`
		AuthenticationTrustCA   string           `json:"authenticationTrustCA,omitempty"`
		AuthenticationMode      string           `json:"authenticationMode"`
		AuthenticationFrequency string           `json:"authenticationFrequency,omitempty"`
	}

	ClientTLS struct {
		Name                string `json:"-"`
		Class               string `json:"class"`
		Label               string `json:"label,omitempty"`
		Remark              string `json:"remark,omitempty"`
		SendSNI             string `json:"sendSNI,omitempty"`
		Ciphers             string `json:"ciphers,omitempty"`
		ServerName          string `json:"serverName,omitempty"`
		ValidateCertificate bool   `json:"validateCertificate,omitempty"`
		TrustCA             string `json:"trustCA,omitempty"`
		IgnoreExpired       bool   `json:"ignoreExpired,omitempty"`
		IgnoreUntrusted     bool   `json:"ignoreUntrusted,omitempty"`
		SessionTickets      bool   `json:"sessionTickets,omitempty"`
		ClientCertificate   string `json:"clientCertificate"`
	}

	CABundle struct {
		Name   string `json:"-"`
		Class  string `json:"class"`
		Bundle string `json:"bundle"`
	}
	CertName struct {
		Certificate string `json:"certificate"`
	}

	Certificate struct {
		Name        string `json:"-"`
		Class       string `json:"class"`
		Certificate string `json:"certificate"`
		PrivateKey  string `json:"privateKey"`
		ChainCA     string `json:"chainCA"`
	}

	PolicyEndpoint struct {
		Name     string        `json:"-"`
		Class    string        `json:"class"`
		Label    string        `json:"label"`
		Remark   string        `json:"remark"`
		Rules    []*PolicyRule `json:"rules"`
		Strategy string        `json:"strategy,omitempty"`
	}

	PolicyRule struct {
		Name       string       `json:"name"`
		Conditions []*Condition `json:"conditions"`
		Actions    []*Action    `json:"actions"`
	}

	Action struct {
		Type     string               `json:"type"`
		Event    string               `json:"event"`
		Enabled  bool                 `json:"enabled,omitempty"`
		Select   *ActionForwardSelect `json:"select,omitempty"`
		Policy   *ResourcePointer     `json:"policy,omitempty"`
		Location string               `json:"location,omitempty"`
		Replace  *ActionReplaceMap    `json:"replace,omitempty"`
	}

	ActionReplaceMap struct {
		Value string `json:"value,omitempty"`
		Name  string `json:"name,omitempty"`
		Path  string `json:"path,omitempty"`
	}

	// as3Condition maps to Policy_Condition in AS3 Resources
	Condition struct {
		Type        string               `json:"type,omitempty"`
		Name        string               `json:"name,omitempty"`
		Event       string               `json:"event,omitempty"`
		All         *PolicyCompareString `json:"all,omitempty"`
		Index       int                  `json:"index,omitempty"`
		ServerName  *PolicyCompareString `json:"serverName,omitempty"`
		Host        *PolicyCompareString `json:"host,omitempty"`
		PathSegment *PolicyCompareString `json:"pathSegment,omitempty"`
		Path        *PolicyCompareString `json:"path,omitempty"`
		Address     *PolicyCompareString `json:"address,omitempty"`
		Normalized  bool                 `json:"normalized"`
	}

	// as3PolicyCompareString maps to Policy_Compare_String in AS3 Resources
	PolicyCompareString struct {
		CaseSensitive bool     `json:"caseSensitive,omitempty"`
		Values        []string `json:"values,omitempty"`
		Operand       string   `json:"operand,omitempty"`
	}

	ActionForwardSelect struct {
		Pool    *ResourcePointer `json:"pool,omitempty"`
		Service *ResourcePointer `json:"service,omitempty"`
	}

	ResourcePointer struct {
		BigIP  string `json:"bigip,omitempty"`
		Use    string `json:"use,omitempty"`
		Base64 string `json:"base64,omitempty"`
	}

	IRule struct {
		Name  string           `json:"-"`
		Class string           `json:"class"`
		IRule *ResourcePointer `json:"iRule"`
	}

	DataGroup struct {
		Class       string    `json:"class"`
		Label       string    `json:"label"`
		StorageType string    `json:"storageType"`
		Name        string    `json:"name"`
		KeyDataType string    `json:"keyDataType"`
		Records     []*Record `json:"records"`
	}

	Record struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
)
