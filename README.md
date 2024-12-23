package main

import (
	"fmt"
	"log"
	"net"

	"github.com/miekg/dns"
)

// handleDNSRequest processes incoming DNS queries and logs the DNS name.
func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	for _, q := range r.Question {
		log.Printf("Received query for: %s", q.Name)

		// Prepare the DNS response.
		response := new(dns.Msg)
		response.SetReply(r)

		// Add a fake response for demonstration purposes.
		aRecord := &dns.A{
			Hdr: dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    600,
			},
			A: net.ParseIP("127.0.0.1"), // Respond with localhost IP.
		}
		response.Answer = append(response.Answer, aRecord)

		// Write the response back to the client.
		if err := w.WriteMsg(response); err != nil {
			log.Printf("Failed to write response: %v", err)
		}
	}
}

func main() {
	// Create a DNS server.
	server := &dns.Server{
		Addr: ":5353", // DNS server listens on port 5353.
		Net:  "udp",
	}

	// Attach the request handler to process DNS queries.
	dns.HandleFunc(".", handleDNSRequest)

	log.Println("Starting DNS server on :5353...")

	// Start the server.
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start DNS server: %v", err)
	}

	defer server.Shutdown()
}
# dnsflixer
