package config

type Config struct {
	HTTPAddress      string
	ExposedHost      string
	ExposedPort      string
	EnableFakeLoad   bool
	Postgresql       PostgresqlConfig
	ServiceDiscovery ServiceDiscoveryConfig
	Role             string
	StorageType      string
}

type PostgresqlConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}

type ServiceDiscoveryConfig struct {
	Resolver  string
	Shortener string
}
