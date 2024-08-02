package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type ServerConfig struct {
	Path   string `yaml:"path"`
	Target string `yaml:"target"`
}

func loadSignatures(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var signatures []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		signatures = append(signatures, scanner.Text())
	}
	return signatures, scanner.Err()
}

func checkBadContent(body string, signatures []string) bool {
	for _, signature := range signatures {
		if signature == "" || signature == "\n" || signature == " " {
			continue
		}
		if strings.Contains(strings.ToLower(body), strings.ToLower(signature)) {
			fmt.Println("Bad content detected: ", signature)
			return true
		}
	}
	return false
}

func main() {
	configData, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
	signatures, err := loadSignatures("signature.txt")
	if err != nil {
		log.Fatalf("Error loading signatures: %v", err)
	}

	var config struct {
		Servers []ServerConfig `yaml:"servers"`
	}
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		for _, server := range config.Servers {
			if r.URL.Path == server.Path {
				bufOfRequestBody, _ := io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(bufOfRequestBody))
				strBody := string(bufOfRequestBody)
				strBody, _ = url.QueryUnescape(strBody)
				fmt.Println("Request-Client-IP", r.RemoteAddr)
				fmt.Println("Request-Body-Decode", strBody)
				if strBody != "" {
					if checkBadContent(strBody, signatures) {
						fmt.Println("悪性通信")
						http.Error(w, "悪性通信のためブロックしました", http.StatusForbidden)
						return
					} else {
						fmt.Println("良性通信")
					}
				}

				targetURL, err := url.Parse(server.Target)
				if err != nil {
					log.Printf("Error parsing target URL: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				proxy := httputil.NewSingleHostReverseProxy(targetURL)
				proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
					log.Printf("Proxy error: %v", err)
					http.Error(w, "Bad Gateway", http.StatusBadGateway)
				}

				proxy.ServeHTTP(w, r)
				return
			}
		}

		http.Error(w, "Not Found", http.StatusNotFound)
	})

	fmt.Println("Reverse proxy listening on :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
