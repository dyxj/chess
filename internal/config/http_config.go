package config

import "time"

type HTTPServerConfig struct {
	HostEV                string        `env:"HOST"`
	PortEV                int           `env:"PORT"`
	ReadHeaderTimeoutEV   time.Duration `env:"READ_HEADER_TIMEOUT"`
	ReadTimeoutEV         time.Duration `env:"READ_TIMEOUT"`
	IdleTimeoutEV         time.Duration `env:"IDLE_TIMEOUT"`
	HandlerTimeoutEV      time.Duration `env:"HANDLER_TIMEOUT"`
	ShutDownTimeoutEV     time.Duration `env:"SHUT_DOWN_TIMEOUT"`
	ShutDownHardTimeoutEV time.Duration `env:"SHUT_DOWN_HARD_TIMEOUT"`
	ShutDownReadyDelayEV  time.Duration `env:"SHUT_DOWN_READY_DELAY"`
}

func (c *HTTPServerConfig) Host() string {
	return c.HostEV
}

func (c *HTTPServerConfig) Port() int {
	return c.PortEV
}

func (c *HTTPServerConfig) ReadHeaderTimeout() time.Duration {
	return c.ReadHeaderTimeoutEV
}

func (c *HTTPServerConfig) ReadTimeout() time.Duration {
	return c.ReadTimeoutEV
}

func (c *HTTPServerConfig) IdleTimeout() time.Duration {
	return c.IdleTimeoutEV
}

func (c *HTTPServerConfig) HandlerTimeout() time.Duration {
	return c.HandlerTimeoutEV
}

func (c *HTTPServerConfig) ShutDownTimeout() time.Duration {
	return c.ShutDownTimeoutEV
}

func (c *HTTPServerConfig) ShutDownHardTimeout() time.Duration {
	return c.ShutDownHardTimeoutEV
}

func (c *HTTPServerConfig) ShutDownReadyDelay() time.Duration {
	return c.ShutDownReadyDelayEV
}
