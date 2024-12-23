package dns

import (
	"fmt"
	"github.com/miekg/dns"
	"log"
	"os"
	"io"
	"regexp"
	"strconv"
	"strings"
)

var queryRegex = regexp.MustCompile(`^[a-zA-Z0-9]+:\d+:\d+$`)

func StartServer(port string, filePath string) error {
	server := &dns.Server{
		Addr: ":" + port,
		Net:  "udp",
	}
	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		handleDNSRequest(w, r, filePath)
	})
	return server.ListenAndServe()
}

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg, filePath string) {
	for _, q := range r.Question {
		log.Printf("Received query for: %s", q.Name)

		trimmedName := strings.TrimSuffix(q.Name, ".")
		if !queryRegex.MatchString(trimmedName) {
			sendDNSError(w, r, q.Name, "Invalid query format")
			return
		}

		parts := strings.Split(trimmedName, ":")
		fileHash := parts[0]
		index, _ := strconv.Atoi(parts[1])
		size, _ := strconv.Atoi(parts[2])

		file, err := os.Open(fmt.Sprintf("%s/%s.b64", filePath, fileHash))
		if err != nil {
			sendDNSError(w, r, q.Name, "File not found")
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
	log.Printf("Error for query %s: %s", name, message)
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
