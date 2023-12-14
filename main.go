package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type ResolveResponse struct {
	IPAddresses []string `json:"ip_addresses"`
}

func resolveIPs(w http.ResponseWriter, r *http.Request) {
	domain := chi.URLParam(r, "domain")

	// Step 1: Create a recon-ng script file with the necessary commands
	scriptContent := `
marketplace install recon/hosts-hosts/resolve
modules load recon/hosts-hosts/resolve
options set SOURCE ` + domain + `
run
exit
`
	scriptFile := "/app/recon_script.rc"
	if err := ioutil.WriteFile(scriptFile, []byte(scriptContent), 0644); err != nil {
		log.Printf("Error writing script file: %v", err)
		http.Error(w, "Failed to write script file", http.StatusInternalServerError)
		return
	}

	// Step 2: Run recon-ng with the script file
	cmd := exec.Command("/app/recon-ng/recon-ng", "-r", scriptFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Command execution error: %v", err)
		http.Error(w, "Failed to resolve IPs", http.StatusInternalServerError)
		return
	}

	// Regular expression to match IP addresses
	ipRegex := regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`)

	// Parse the output and extract IP addresses
	ipAddresses := []string{}
	for _, line := range strings.Split(string(output), "\n") {
		if ip := ipRegex.FindString(line); ip != "" {
			ipAddresses = append(ipAddresses, ip)
		}
	}

	// Remove duplicates by converting to a map, then back to a slice
	uniqueIPs := make(map[string]struct{})
	for _, ip := range ipAddresses {
		uniqueIPs[ip] = struct{}{}
	}

	ipList := []string{}
	for ip := range uniqueIPs {
		ipList = append(ipList, ip)
	}

	// Respond with JSON containing only the IP addresses
	response := ResolveResponse{IPAddresses: ipList}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	// Clean up by removing the script file
	os.Remove(scriptFile)
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/resolve/{domain}", resolveIPs)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
