package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/experiflow/proxy/internal/config"
	"github.com/experiflow/proxy/internal/middleware"
)

func main() {
	// Load configuration
	cfg := config.LoadFromEnv()

	log.Printf("[ExperiFlow Proxy] Starting on port %s", cfg.Port)
	log.Printf("[ExperiFlow Proxy] Origin: %s", cfg.OriginURL)
	log.Printf("[ExperiFlow Proxy] API: %s", cfg.APIBaseURL)
	log.Printf("[ExperiFlow Proxy] Fail Open: %v", cfg.FailOpen)

	// Parse origin URL
	originURL, err := url.Parse(cfg.OriginURL)
	if err != nil {
		log.Fatalf("Invalid origin URL: %v", err)
	}

	// Get experiment IDs from environment
	experimentIDs := getExperimentIDs()
	if len(experimentIDs) == 0 {
		log.Println("[ExperiFlow Proxy] WARNING: No experiment IDs configured. Set EXPERIMENT_IDS env var.")
	} else {
		log.Printf("[ExperiFlow Proxy] Active experiments: %v", experimentIDs)
	}

	// Create ExperiFlow middleware
	efMiddleware := middleware.NewExperiFlowMiddleware(cfg, experimentIDs)

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(originURL)

	// Customize proxy behavior
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// Preserve original host header
		req.Host = originURL.Host
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Header.Set("X-Forwarded-Proto", req.URL.Scheme)
	}

	// Add response modification
	proxy.ModifyResponse = func(resp *http.Response) error {
		return efMiddleware.ModifyResponse(resp, resp.Request)
	}

	// Error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("[ExperiFlow Proxy] Proxy error: %v", err)
		if cfg.FailOpen {
			// Try to pass through to origin directly
			log.Println("[ExperiFlow Proxy] Failing open - attempting direct connection")
		}
		http.Error(w, "Proxy error", http.StatusBadGateway)
	}

	// Health check handler
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"experiflow-proxy"}`))
	})
	mux.Handle("/", proxy)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	// Start server
	log.Printf("[ExperiFlow Proxy] Ready to accept requests on http://localhost:%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("[ExperiFlow Proxy] Server error: %v", err)
	}
}

// getExperimentIDs parses experiment IDs from environment variable
// Format: comma-separated list, e.g., "exp1,exp2,exp3"
func getExperimentIDs() []string {
	idsStr := os.Getenv("EXPERIMENT_IDS")
	if idsStr == "" {
		return nil
	}

	ids := strings.Split(idsStr, ",")
	var result []string
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" {
			result = append(result, id)
		}
	}
	return result
}
