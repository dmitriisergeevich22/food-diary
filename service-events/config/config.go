package config

// Конфигурация приложения.
type Config struct {
	ServiceName  string  `env:"SERVICE_NAME" validate:"required"`         // Имя сервиса.
	PGWriterConn string  `env:"POSTGRES_WRITER_CONN" validate:"required"` // Строка подключения к БД для записи.
	PGReaderConn string  `env:"POSTGRES_READER_CONN"`                     // Опциональная. Строка подключения к БД для чтения.
	LogLevel     int     `env:"LOG_LEVEL, default=-4"`                    // debug = -4, info = 0, warn = 4
	Gateway      Gateway `env:", prefix=GATEWAY_"`
}

// Сетевые настройки.
type Gateway struct {
	AuthToken        string `env:"AUTH_TOKEN" validate:"required"`
	PathToSwaggerDir string `env:"SWAGGER_PATH, default=docs/swagger"`
	HTTP             Adr    `env:", prefix=HTTP_"`
	GRPC             Adr    `env:", prefix=GRPC_"`
}

type Adr struct {
	Host string `env:"HOST" validate:"required"`
	Port string `env:"PORT" validate:"required"`
}
