package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	serverPort string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run HTTP server for Monday workflow",
	Long: `Start an HTTP server that exposes endpoints to trigger the Monday workflow:
- GET /health - Health check endpoint
- POST /trigger - Trigger workflow with linear_id and github_url`,
	RunE: runServer,
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVar(&serverPort, "port", "", "HTTP server port (default: 8080 or $PORT)")
}

func runServer(cmd *cobra.Command, args []string) error {
	initLogger()
	
	port := serverPort
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}

	apiKey := os.Getenv("SERVER_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("SERVER_API_KEY environment variable is required")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/trigger", makeTriggerHandler(logger, apiKey))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	logger.Info("Starting Monday HTTP server", zap.String("port", port))
	fmt.Printf("🚀 Monday server starting on port %s\n", port)
	fmt.Printf("📋 Health check: GET http://localhost:%s/health\n", port)
	fmt.Printf("🔗 Trigger workflow: POST http://localhost:%s/trigger\n", port)
	
	return srv.ListenAndServe()
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

type triggerRequest struct {
	LinearID  string `json:"linear_id"`
	GithubURL string `json:"github_url"`
}

type triggerResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func makeTriggerHandler(logger *zap.Logger, apiKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if r.Header.Get("X-API-Key") != apiKey {
			logger.Warn("Unauthorized request", zap.String("remote_addr", r.RemoteAddr))
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req triggerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("Failed to decode request", zap.Error(err))
			http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		if req.LinearID == "" || req.GithubURL == "" {
			http.Error(w, "linear_id and github_url are required", http.StatusBadRequest)
			return
		}

		logger.Info("Received workflow trigger request", 
			zap.String("linear_id", req.LinearID),
			zap.String("github_url", req.GithubURL),
			zap.String("remote_addr", r.RemoteAddr))

		go func() {
			if err := runWorkflow(req.LinearID, req.GithubURL); err != nil {
				logger.Error("Workflow failed", zap.Error(err),
					zap.String("linear_id", req.LinearID),
					zap.String("github_url", req.GithubURL))
			} else {
				logger.Info("Workflow completed successfully",
					zap.String("linear_id", req.LinearID),
					zap.String("github_url", req.GithubURL))
			}
		}()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		
		response := triggerResponse{
			Status:  "started",
			Message: fmt.Sprintf("Workflow started for Linear issue %s", req.LinearID),
		}
		
		json.NewEncoder(w).Encode(response)
	}
}
