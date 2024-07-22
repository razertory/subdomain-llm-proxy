package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

var apiEndpoints = map[string]string{
	"openai": "https://api.openai.com",
	"claude": "https://api.anthropic.com",
	"gemini": "https://generativelanguage.googleapis.com",
}

func main() {
	http.HandleFunc("/", handleRequest)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.Host, ".")
	if len(parts) < 2 {
		http.Error(w, "Invalid domain", http.StatusBadRequest)
		return
	}
	subdomain := parts[0]

	targetURL, ok := apiEndpoints[subdomain]
	if !ok {
		http.Error(w, "Unknown API", http.StatusNotFound)
		return
	}

	target, err := url.Parse(targetURL)
	if err != nil {
		http.Error(w, "Invalid target URL", http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
		req.URL.Host = target.Host
		req.URL.Scheme = target.Scheme
		req.Header.Del("X-Forwarded-For")
		req.Header.Del("X-Real-IP")
		req.Header.Del("CF-Connecting-IP")
		req.Header.Set("X-Forwarded-For", "0.0.0.0")
		req.Header.Del("User-Agent")
		req.Header.Del("Referer")
	}

	log.Printf("Proxying request to %s", targetURL)

	// 转发请求
	proxy.ServeHTTP(w, r)
}
