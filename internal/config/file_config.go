package config

// AgentConfigFile describes the JSON configuration format for the agent.
type AgentConfigFile struct {
	Address        string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
	GRPCAddress    string `json:"grpc_address"`
	GRPCCAFile     string `json:"grpc_ca_file"`
}

// ServerConfigFile describes the JSON configuration format for the server.
type ServerConfigFile struct {
	Address       string `json:"address"`
	Restore       *bool  `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDSN   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
	TrustedSubnet string `json:"trusted_subnet"`
	GRPCAddress   string `json:"grpc_address"`
	GRPCCertFile  string `json:"grpc_cert_file"`
	GRPCKeyFile   string `json:"grpc_key_file"`
}
