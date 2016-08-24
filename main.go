package main

import (
	"flag"

	"github.com/abourget/slick"
	// _ "github.com/abourget/slick/web"
	// _ "github.com/abourget/slick/webauth"
	_ "github.com/itsoneiota/llong/deploybot"
)

var configFile = flag.String("config", "slick.conf", "config file")

func main() {
	flag.Parse()

	bot := slick.New(*configFile)

	bot.Run()
}
