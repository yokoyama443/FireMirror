package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"
)

type ServerConfig struct {
	Path   string `yaml:"path"`
	Target string `yaml:"target"`
}

type Response struct {
	// Adjust the fields according to the actual response structure.
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func sendChatRequest(inputText string) (string, error) {

	apiKey := "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"

	url := "https://api.groq.com/openai/v1/chat/completions"
	contentType := "application/json"
	message := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": inputText + "\nPlease tell me if the above text is a string used for OS command injection or SQL injection.\n Short words and common sentences are safe, but those containing multiple os commands or special characters such as <> are dangerous.\n Please write only Yes/No in your answer."},
		},
		"model": "llama3-8b-8192",
	}

	body, err := json.Marshal(message)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", contentType)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response Response
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return "", err
	}
	//fmt.Println("responce message", response.Choices[0].Message.Content)
	return response.Choices[0].Message.Content, nil
}

func generateRandomString(length int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}
	return string(ret), nil
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

func checkForMaliciousContent(body string, signatures []string) bool {
	for _, signature := range signatures {
		if signature == "" || signature == "\n" || signature == " " {
			continue
		}
		if strings.Contains(strings.ToLower(body), strings.ToLower(signature)) {
			fmt.Println("Malicious content detected: ", signature)
			return true
		}
	}
	return false
}

func main() {
	// 設定ファイル読み込み (YAML)
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
					//tmp, _ := sendChatRequest(strBody)
					//tmpF := strings.Contains(tmp, "Yes")
					if checkForMaliciousContent(strBody, signatures) {
						fmt.Println("悪性通信")
						http.Error(w, "悪性通信のためブロックしました", http.StatusForbidden)
						ip := strings.Split(r.RemoteAddr, ":")[0]
						if ip[0] == '[' {
							return
						}
						fmt.Println("Evil SourceIP: ", ip)
						cmd := exec.Command("sh", "-c", "hydra -l ubuntu -P password.lst "+ip+" ssh > tmp")
						output, err := cmd.Output()
						if err != nil {
							fmt.Println(err)
						}
						fmt.Println(string(output))
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
				//fmt.Println("Proxying to", targetURL)
				proxy := httputil.NewSingleHostReverseProxy(targetURL)
				//fmt.Println("Proxying to")
				proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
					log.Printf("Proxy error: %v", err)
					http.Error(w, "Bad Gateway", http.StatusBadGateway)
				}

				// // リバースプロ将棋サーバー
				// proxy.ModifyResponse = func(res *http.Response) error {
				// 	if r.Header.Get("X-Shogi") == "Hello" {
				// 		res.Header.Set("X-Shogi", "World")
				// 	} else {
				// 		randomString, err := generateRandomString(32)
				// 		if err != nil {
				// 			log.Printf("Error generating random string: %v", err)
				// 			return err
				// 		}
				// 		res.Header.Set("X-Shogi", randomString)
				// 	}
				// 	return nil
				// }

				proxy.ServeHTTP(w, r)
				return
			}
		}

		http.Error(w, "Not Found", http.StatusNotFound)
	})

	fmt.Println("Reverse proxy listening on :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
