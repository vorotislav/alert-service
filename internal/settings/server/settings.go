package server

// Settings представляет настройки для сервера.
type Settings struct {
	StorageCfg
	NetworkCfg
	HashKey   string `env:"KEY"`
	CryptoKey string `env:"CRYPTO_KEY" json:"crypto_key"`
	Config    string `env:"CONFIG"`
}

type NetworkCfg struct {
	Address       string `env:"ADDRESS" json:"address"`
	TrustedSubnet string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
	GAddress      string `ENV:"GADDRESS" json:"g_address"`
}

type StorageCfg struct {
	StoreInterval   *int   `env:"STORE_INTERVAL" json:"store_interval,omitempty"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	Restore         *bool  `env:"RESTORE" json:"restore,omitempty"`
	DatabaseDSN     string `env:"DATABASE_DSN" json:"database_dsn"`
}
