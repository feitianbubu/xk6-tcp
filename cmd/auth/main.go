package main

import (
	_ "tevat.nd.org/basecode/loader-jsoner/toml"
	"tevat.nd.org/framework/app"
	_ "tevat.nd.org/provider/elastic"
	_ "tevat.nd.org/provider/jaeger"
	_ "tevat.nd.org/provider/prometheus"
	_ "tevat.nd.org/provider/zap"
)

type Config struct {
	app.SdkConfig
}

func main() {
	var cfg Config

	app.Init(&cfg)
	app.Run()
}
