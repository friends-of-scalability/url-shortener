package config

type Config struct {
	HTTPAddress    string
	EnableFakeLoad bool
	Postgresql     PostgresqlConfig
	StorageType    string
}

type PostgresqlConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}
