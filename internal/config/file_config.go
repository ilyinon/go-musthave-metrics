package config

// AgentConfigFile describes the JSON configuration format for the agent.
type AgentConfigFile struct {
	Address        string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
}

// ServerConfigFile describes the JSON configuration format for the server.
type ServerConfigFile struct {
	Address       string `json:"address"`
	Restore       *bool  `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDSN   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
}
