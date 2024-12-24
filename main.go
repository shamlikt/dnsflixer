package main

import (
	"flag"
	"log"
	"os"
	"sync"
	"io"
	"github.com/shamlikt/dnsflixer/dns"
	"github.com/shamlikt/dnsflixer/httpserver"
	"github.com/pelletier/go-toml"
)

type Config struct {
	FilePath string `toml:"file_path"`
	DNSPort  string `toml:"dns_port"`
	HTTPPort string `toml:"http_port"`
	LogFile  string `toml:"log_file"`
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

// Setup logging to a file
func setupLogging(logFile string) {
	// Open the log file
	logFileHandle, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}


	multiWriter := io.MultiWriter(logFileHandle, os.Stdout)


	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Println("Logging initialized.")
}

func main() {
	// Parse command-line arguments
	configFile := flag.String("config", "config.toml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	loadConfig(*configFile)

	// Setup logging
	setupLogging(config.LogFile)

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
		if err := dns.StartServer(config.DNSPort, config.FilePath, logConnection); err != nil {
			log.Fatalf("DNS server failed: %v", err)
		}
	}()

	// Start HTTP server
	go func() {
		defer wg.Done()
		log.Printf("Starting HTTP server on %s...", config.HTTPPort)
		if err := httpserver.StartServer(config.HTTPPort, config.FilePath, logConnection); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	wg.Wait()
}


func logConnection(serverType, clientAddr, details string) {
	log.Printf("[%s] Connection from %s - Details: %s", serverType, clientAddr, details)
}
