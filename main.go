package main

import (
	"flag"
	"log"
	"os"
	"sync"

	"github.com/shamlikt/dnsflixer/dns"
	"github.com/shamlikt/dnsflixer/httpserver"
	"github.com/pelletier/go-toml"
)


type Config struct {
	FilePath string `toml:"file_path"`
	DNSPort  string `toml:"dns_port"`
	HTTPPort string `toml:"http_port"`
}

var config Config

// Load configuration from TOML file
func loadConfig(configFile string) {
	file, err := os.Open(configFile)
	if err != nil {
		log.Fatalf("Failed to open config file: %v", err)
	}
	defer file.Close()

	decoder := toml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	log.Printf("Config loaded: %+v", config)
}

func main() {
	// Parse command-line arguments
	configFile := flag.String("config", "config.toml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	loadConfig(*configFile)

	// Ensure the files directory exists
	if err := os.MkdirAll(config.FilePath, os.ModePerm); err != nil {
		log.Fatalf("Failed to create files directory: %v", err)
	}

	// Run DNS and HTTP servers concurrently
	var wg sync.WaitGroup
	wg.Add(2)

	// Start DNS server
	go func() {
		defer wg.Done()
		log.Printf("Starting DNS server on %s...", config.DNSPort)
		if err := dns.StartServer(config.DNSPort, config.FilePath); err != nil {
			log.Fatalf("DNS server failed: %v", err)
		}
	}()

	// Start HTTP server
	go func() {
		defer wg.Done()
		log.Printf("Starting HTTP server on %s...", config.HTTPPort)
		if err := httpserver.StartServer(config.HTTPPort, config.FilePath); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	wg.Wait()
}
