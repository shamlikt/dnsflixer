package dns

import (
	"fmt"
	"github.com/miekg/dns"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var queryRegex = regexp.MustCompile(`^[a-zA-Z0-9]+:\d+:\d+$`)

func StartServer(port string, filePath string,  logConnection func(serverType, clientAddr, query string)) error {
	server := &dns.Server{
		Addr: ":" + port,
		Net:  "udp",
	}
	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		clientIP := w.RemoteAddr().String()
		for _, q := range r.Question {
			logConnection("DNS", clientIP, q.Name)
			handleDNSRequest(w, r, filePath)
		}
	})
	log.Printf("DNS server listening on port %s...", port)
	return server.ListenAndServe()
	
}

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg, filePath string) {
	clientIP := w.RemoteAddr().String() // Get client's IP address
	log.Printf("Received DNS query from %s", clientIP)

	for _, q := range r.Question {
		log.Printf("Query from %s for: %s", clientIP, q.Name)

		trimmedName := strings.TrimSuffix(q.Name, ".")
		if !queryRegex.MatchString(trimmedName) {
			sendDNSError(w, r, q.Name, "Invalid query format")
			log.Printf("Invalid query format from %s for: %s", clientIP, q.Name)
			return
		}

		parts := strings.Split(trimmedName, ":")
		fileHash := parts[0]
		index, _ := strconv.Atoi(parts[1])
		size, _ := strconv.Atoi(parts[2])

		file, err := os.Open(fmt.Sprintf("%s/%s.b64", filePath, fileHash))
		if err != nil {
			sendDNSError(w, r, q.Name, "File not found")
			log.Printf("File not found for query %s from %s", q.Name, clientIP)
			return
		}
		defer file.Close()

		start := int64(index) * int64(size)
		file.Seek(start, io.SeekStart)

		buffer := make([]byte, size)
		bytesRead, _ := io.ReadFull(file, buffer)

		responseText := string(buffer[:bytesRead])
		if bytesRead < size {
			responseText += "$"
		}

		log.Printf("Query response for %s from %s: %s", q.Name, clientIP, responseText)

		response := new(dns.Msg)
		response.SetReply(r)
		response.Authoritative = true
		response.Answer = append(response.Answer, &dns.TXT{
			Hdr: dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    60,
			},
			Txt: []string{responseText},
		})

		w.WriteMsg(response)
	}
}

func sendDNSError(w dns.ResponseWriter, r *dns.Msg, name, message string) {
	clientIP := w.RemoteAddr().String() // Get client's IP address
	log.Printf("Error for query %s from %s: %s", name, clientIP, message)

	response := new(dns.Msg)
	response.SetReply(r)
	response.Authoritative = true
	response.Answer = append(response.Answer, &dns.TXT{
		Hdr: dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeTXT,
			Class:  dns.ClassINET,
			Ttl:    60,
		},
		Txt: []string{message},
	})
	w.WriteMsg(response)
}
