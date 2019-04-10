package server

type Config struct {
	ConfigDir          string `env:"config_dir"`
	RequestLogCapacity int    `env:"request_log_capacity" default:"0"`
}
