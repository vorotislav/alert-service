package server

type Settings struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   *int   `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         *bool  `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	HashKey         string `env:"KEY"`
}
