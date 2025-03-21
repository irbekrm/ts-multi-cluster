package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	app string
)

func init() {
	app = os.Getenv("APP")
}

func getRegion() (string, error) {
	// Open and read /etc/resolv.conf
	file, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return "", fmt.Errorf("failed to open resolv.conf: %v", err)
	}
	defer file.Close()

	// Scan the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Look for the search line
		if strings.HasPrefix(line, "search") {
			// Split the line into parts
			parts := strings.Fields(line)
			// Look for the part containing the region info
			for _, part := range parts {
				if strings.Contains(part, "internal") {
					// Extract region from the hostname
					// Format: us-central1-c.c.tailscale-sandbox.internal
					region := strings.Split(part, ".")[0]
					return region, nil
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading resolv.conf: %v", err)
	}

	return "", fmt.Errorf("region information not found in resolv.conf")
}

func regionHandler(w http.ResponseWriter, r *http.Request) {
	region, err := getRegion()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting region: %v", err), http.StatusInternalServerError)
		return
	}
	ipp := strings.Split(r.RemoteAddr, ":")
	fmt.Fprintf(w, "Hello from app %s in %s. Received request from proxy with IP %s\n", app, region, ipp[0])
}

func main() {
	// Register handler for the root path
	http.HandleFunc("/", regionHandler)

	// Start server on port 8080
	port := ":8080"
	log.Printf("Starting server on port %s as app %s", port, app)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
