package config

type Config struct {
	// Context represents currently used auth context.
	Context string
	// Agent holds all available Agent's configuration. Stored as map for O(n) access, we don't care that it's not sorted.
	Agent map[string]Agent
}

type Agent struct {
	// Alias specifies Agent's alias. Necessary for storing configuration for the same Agent's server URL multiple times.
	Alias string
	// ServerURL specifies Agent's server URL.
	ServerURL string
	// AgentCAFilePath specifies CA file path. It is used to validate if we can trust Agent's server.
	AgentCAFilePath string
	// ClientAuth holds client configuration for authorization.
	ClientAuth ClientAuth
}

type ClientAuth struct {
	// TODO(simplification): later we can add different auth method options and use command pattern
	// to select a proper one based on exposed `--method` flag.
	//
	// Method specifies the client auth type.
	// Method string

	// ClientCertAuth holds client certs configuration
	ClientCertAuth
}

type ClientCertAuth struct {
	CertFilePath string
	KeyFilePath  string
}

func (a *Agent) IsEmpty() bool {
	if a == nil {
		return true
	}
	if *a == (Agent{}) {
		return true
	}
	return false
}
