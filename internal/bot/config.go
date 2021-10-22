package bot

type Config struct {
	PathToIndex    string `env:"PATH_TO_INDEX,default=./bin/cities.idx"`
	ProxySchema    string `env:"TELEGRAM_PROXY_SCHEMA,default=http"`
	ProxyAddr      string `env:"TELEGRAM_PROXY_ADDR,default=127.0.0.1:8081"`
	WebHookURL     string `env:"TELEGRAM_WEBHOOK_URL"`
	WebHookAddr    string `env:"TELEGRAM_WEBHOOK_ADDR"`
	Token          string `env:"TELEGRAM_TOKEN"`
	PollingTimeout int    `env:"TELEGRAM_POLLING_TIMEOUT,default=10"`
	MaxWorkers     int    `env:"TELEGRAM_POLLING_TIMEOUT,default=10"`
}
