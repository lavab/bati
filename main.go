package main

import (
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/lavab/kiri"
	"github.com/namsral/flag"
)

var (
	configFlag       = flag.String("config", "", "Config file to read")                  // Enable config file cuntionality
	logFormatterType = flag.String("log_formatter_type", "text", "Log formatter to use") // Logrus log formatter
	logForceColors   = flag.Bool("log_force_colors", false, "Force colored log output?") // Logrus force colors
	etcdAddress      = flag.String("etcd_addresses", "", "Addresses of the etcd servers to use")

	services = flag.String("services", "", "Services list. Syntax: name,address,tag=val,tag=val;name,address")
	stores   = flag.String("stores", "", "Stores list. Syntax: kind,path;kind,path")
)

func main() {
	// Parse the flags
	flag.Parse()

	// Set up a new logger
	log := logrus.New()

	// Set the formatter depending on the passed flag's value
	if *logFormatterType == "text" {
		log.Formatter = &logrus.TextFormatter{
			ForceColors: *logForceColors,
		}
	} else if *logFormatterType == "json" {
		log.Formatter = &logrus.JSONFormatter{}
	}

	// Split the addresses
	etcds := strings.Split(*etcdAddress, ",")

	// Create a new kiri client
	sd := kiri.New(etcds)

	// Parse and register the services
	for i, service := range strings.Split(*services, ";") {
		parts := strings.Split(service, ",")

		if len(parts) < 2 {
			log.Fatalf("Invalid service parts count in %d", i)
		}

		var tags map[string]interface{}
		if len(parts) > 2 {
			tags = map[string]interface{}{}
			for i := 3; i < len(parts); i++ {
				fields := strings.Split(parts[i], "=")
				tags[fields[0]] = fields[1]
			}
		}

		sd.Register(parts[0], parts[1], tags)
	}

	// Set up stores
	for i, store := range strings.Split(*stores, ";") {
		parts := strings.Split(store, ",")

		if len(parts) != 2 {
			log.Fatalf("Invalid store parts count in %d", i)
		}

		var kind kiri.Format

		switch parts[0] {
		case "default":
			kind = kiri.Default
		case "puro":
			kind = kiri.Puro
		default:
			log.Fatalf("Invalid kind of store in %d", i)
		}

		sd.Store(kind, parts[1])
	}

	// Lock up the process
	select {}
}
