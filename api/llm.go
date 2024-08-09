package api

import (
	"io"
	"net/http"
	"time"
)

var (
	client = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: 30 * time.Second,
			// DisableCompression: true,
		},
	}
	llmMap = map[string]string{
		"groq":       "https://api.groq.com/openai/v1/chat/completions",
		"openai":     "https://api.openai.com/v1/chat/completions",
		"gemini":     "https://generativelanguage.googleapis.com/v1beta/models/",
		"openrouter": "https://openrouter.ai/api/v1/chat/completions",
	}
)

func proxyHandler(href string, w http.ResponseWriter, r *http.Request) {
	// client := &http.Client{}

	// Create a new request using http
	req, err := http.NewRequest(r.Method, href, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy the headers from the original request
	for name, values := range r.Header {
		// Loop over all values for the name.
		for _, value := range values {
			req.Header.Set(name, value)
		}
	}

	// Send the request via a client
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	// Copy the headers from the response
	for name, values := range resp.Header {
		// Loop over all values for the name.
		for _, value := range values {
			w.Header().Set(name, value)
		}
	}

	// Set the status code from the response
	w.WriteHeader(resp.StatusCode)

	// Copy the body from the response to the writer
	io.Copy(w, resp.Body)
}

func LLM(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.Header.Get("Authorization")
	if key == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	targetURL := switchllm(r)

	proxyHandler(targetURL, w, r)
}

func switchllm(r *http.Request) (targetURL string) {
	p := r.URL.Query().Get("p")
	switch p {
	case "groq":
		targetURL = llmMap["groq"]
	case "openai":
		targetURL = llmMap["openai"]
	case "openrouter":
		targetURL = llmMap["openrouter"]
	case "gemini":
		targetURL = llmMap["gemini"]
		model := r.URL.Query().Get("model")
		targetURL = targetURL + model + ":generateContent?key=" + r.Header.Get("Authorization")
		r.Header.Del("Authorization")
	default:
		targetURL = llmMap["openai"]
	}
	return
}
