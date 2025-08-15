package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var service string

func init() {
	// Register a command-line flag to filter logs by service name
	flag.StringVar(&service, "service", "", "filter which service to see")

	// Ignore SIGINT (Ctrl+C) to avoid accidental termination in log pipelines
	signal.Ignore(syscall.SIGINT)
}

func main() {
	// Parse CLI flags
	flag.Parse()

	var b strings.Builder
	service := strings.ToLower(service) // Normalize service filter to lowercase

	// Create a scanner to read input line-by-line from stdin
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		s := scanner.Text()
		m := make(map[string]any)

		// Try to parse the log line as JSON
		if err := json.Unmarshal([]byte(s), &m); err != nil {
			// If parsing fails and no service filter is set, print raw line
			if service == "" {
				fmt.Println(s)
			}
			continue
		}

		// If service filter is set, skip non-matching logs
		if service != "" && strings.ToLower(m["service"].(string)) != service {
			continue
		}

		// Default trace ID if missing
		traceID := "00000000-0000-0000-0000-000000000000"
		if v, ok := m["trace_id"]; ok {
			traceID = fmt.Sprintf("%v", v)
		}

		// Reset string builder to reuse memory
		b.Reset()
		// Format the main log fields in a fixed order
		b.WriteString(fmt.Sprintf(
			"%s: %s: %s: %s: %s: %s: ",
			m["service"],
			m["time"],
			m["file"],
			m["level"],
			traceID,
			m["msg"],
		))

		// Append additional fields (exclude main ones)
		for k, v := range m {
			switch k {
			case "service", "time", "file", "level", "trace_id", "msg":
				continue
			}
			b.WriteString(fmt.Sprintf("%s[%v]: ", k, v))
		}

		// Remove the last ": " and print
		out := b.String()
		fmt.Println(out[:len(out)-2])
	}

	// Handle possible scanner errors
	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}
