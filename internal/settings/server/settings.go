package server

// Settings представляет настройки для сервера.
type Settings struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   *int   `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         *bool  `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	HashKey         string `env:"KEY"`
	CryptoKey       string `env:"CRYPTO_KEY"`
	Config          string `env:"CONFIG"`
}

type Config struct {
	Address       string  `json:"address"`
	Restore       *bool   `json:"restore,omitempty"`
	StoreInterval *string `json:"store_interval,omitempty"`
	StoreFile     string  `json:"store_file"`
	DatabaseDsn   string  `json:"database_dsn"`
	CryptoKey     string  `json:"crypto_key"`
}
