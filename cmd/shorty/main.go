// Go shorty it's your birthday 🎂
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/NorskHelsenett/shorty/internal/config"
	docs "github.com/NorskHelsenett/shorty/internal/docs"
	"github.com/NorskHelsenett/shorty/internal/handlers"
	"github.com/NorskHelsenett/shorty/internal/metrics"
	"github.com/NorskHelsenett/shorty/internal/middleware"

	rlog "github.com/NorskHelsenett/ror/pkg/rlog"
	redis "github.com/go-redis/redis/v8"
	mux "github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	viper "github.com/spf13/viper"
	httpSwagger "github.com/swaggo/http-swagger"
)

const (
	defaultPort            = "8880"
	defaultHostname        = "localhost"
	defaultOIDCClientID    = "shortyfront"
	defaultOIDCProviderURL = "http://dex.localtest.me:5556/dex"
)

var (
	listener *HTTPServer
	server   *kvServer
	version  string
)

// HTTPServer represents an HTTP server configuration
type HTTPServer struct {
	Hostname   string
	DisableTLS bool
	Port       string
}

// NewHTTPServer creates a new HTTP server configuration from environment variables
func NewHTTPServer() *HTTPServer {
	return &HTTPServer{
		Hostname:   getStringWithDefault(viper.GetString("HOST"), defaultHostname),
		Port:       getStringWithDefault(viper.GetString("PORT"), defaultPort),
		DisableTLS: viper.GetBool("NO_TLS"),
	}
}

// Helper function to get string with default value
func getStringWithDefault(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return val
}

// GetURI returns the full URI for the server with appropriate protocol
func (l *HTTPServer) GetURI() string {
	if l.DisableTLS {
		return fmt.Sprintf("http://%s:%s", l.Hostname, l.Port)
	}
	return fmt.Sprintf("https://%s", l.Hostname)
}

// GetPort returns the server port
func (l *HTTPServer) GetPort() string {
	return l.Port
}

// GetHostPort returns hostname:port or just hostname depending on TLS setting
func (l *HTTPServer) GetHostPort() string {
	if l.DisableTLS {
		return fmt.Sprintf("%s:%s", l.GetHostname(), l.GetPort())
	}
	return l.GetHostname()
}

// GetHostname returns the server hostname
func (l *HTTPServer) GetHostname() string {
	return l.Hostname
}

type kvServer struct {
	rdb *redis.Client
}

func newServer(rdb *redis.Client) *kvServer {
	return &kvServer{rdb: rdb}
}

// HealthCheck handles health check requests by verifying Redis connectivity
// and returning appropriate status based on system health.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	status := "Healthy"
	statusCode := http.StatusOK

	// Check Redis connection
	if server != nil && server.rdb != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		if err := server.rdb.Ping(ctx).Err(); err != nil {
			status = "Unhealthy"
			statusCode = http.StatusServiceUnavailable
			rlog.Error("Redis health check failed", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if _, err := w.Write([]byte(fmt.Sprintf(`{"status":"%s"}`, status))); err != nil {
		rlog.Error("Failed to write health check response", err)
	}
}

func configureSwagger() {
	docs.SwaggerInfo.Host = listener.GetHostPort()
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Title = fmt.Sprintf("%s API", listener.GetHostname())
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Description = "Urlforkorter for Norsk helsenett"
}

func setupRouter(rdb *redis.Client) *mux.Router {
	r := mux.NewRouter()

	// metric
	r.Handle("/metrics", promhttp.Handler())

	// defines routes
	r.HandleFunc("/health", HealthCheck)
	r.HandleFunc("/{id}", handlers.Redirect(rdb)).Methods("GET")
	r.HandleFunc("/", handlers.Redirect(rdb)).Methods("GET")
	r.HandleFunc("", handlers.Redirect(rdb)).Methods("GET")

	adminRoute := r.PathPrefix("/v1").Subrouter()
	adminRoute.Use(middleware.AuthenticationMiddlewareWrapper(rdb))
	adminRoute.Use(middleware.AddAdminStatusMiddlewareWrapper(rdb))

	// User
	adminRoute.HandleFunc("/user", handlers.AddUserRedirect(rdb)).Methods("POST")
	adminRoute.HandleFunc("/user", handlers.GetAllUsersRedirect(rdb)).Methods("GET")
	adminRoute.HandleFunc("/user/{id}", handlers.DeleteUserRedirect(rdb)).Methods("DELETE")

	// URL
	urlRoute := adminRoute.PathPrefix("/").Subrouter()
	urlRoute.Use(middleware.IsOwnerMiddlewareWrapper(rdb))
	urlRoute.HandleFunc("/", handlers.AddRedirect(rdb)).Methods("POST")
	urlRoute.HandleFunc("/", handlers.GetAllRedirects(rdb)).Methods("GET")
	urlRoute.HandleFunc("/{id}", handlers.UpdateRedirect(rdb)).Methods("PATCH")
	urlRoute.HandleFunc("/{id}", handlers.DeleteRedirect(rdb)).Methods("DELETE")

	// QR-code
	adminRoute.HandleFunc("/qr/{id}", handlers.GenerateQRCode(rdb)).Methods("GET")
	qrRouter := r.PathPrefix("/qr").Subrouter()
	qrRouter.HandleFunc("/", handlers.GenerateQRCodeFromUrl()).Methods("GET")

	r.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("%s/swagger/doc.json", listener.GetHostname())),
	)).Methods(http.MethodGet)

	return r
}

// the main function
//
//	@contact.name	Containerplattformen

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apikey AccessToken
// @in header
// @name Authorization
func main() {
	// Set up context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	viper.SetDefault("OIDC_CLIENT_ID", defaultOIDCClientID)
	viper.SetDefault("OIDC_PROVIDER_URL", defaultOIDCProviderURL)
	viper.SetDefault("SKIPISSUERCHECK", false)
	viper.SetDefault("INSECURE_SKIP_SIGNATURE_CHECK", false)
	viper.AutomaticEnv()

	if version == "" {
		version = "v1-develop"
	}
	rlog.Info(fmt.Sprintf("## Starting k.nhn.no version %s", version))

	// initializes metrics
	metrics.InitMetrics()

	//loads listener config
	listener = NewHTTPServer()

	// create database client and server instance with error handling
	db, err := config.NewClient()
	if err != nil {
		rlog.Error("Failed to connect to Redis", err)
		os.Exit(1)
	}
	server = newServer(db)

	// Configure Swagger
	configureSwagger()

	// Set up router with all routes
	r := setupRouter(server.rdb)

	// Get allowed origins for CORS with empty check
	allowedOriginsString := viper.GetString("ALLOW_ORIGINS")
	var allowedOrigins []string
	if allowedOriginsString != "" {
		allowedOrigins = strings.Split(allowedOriginsString, ";")
	}

	// Configure HTTP server with timeouts
	url := fmt.Sprintf(":%s", listener.GetPort())
	httpServer := &http.Server{
		Handler:      middleware.RecoveryMiddleware(middleware.CORSMiddleware(r, allowedOrigins)),
		Addr:         url,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Set up graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		// Add more signals to handle
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
		sig := <-sigCh
		rlog.Info(fmt.Sprintf("Received signal: %v", sig))

		// Create a timeout context for shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
		defer shutdownCancel()

		rlog.Info("Shutting down server gracefully...")
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			rlog.Error("Server shutdown error", err)
		}

		// Perform cleanup operations
		if err := db.Close(); err != nil {
			rlog.Error("Error closing Redis connection", err)
		}

		// Add metrics cleanup if needed
		metrics.CleanupMetrics()

		cancel()
	}()

	// Start the HTTP server
	rlog.Info(fmt.Sprintf("Starting server on %s", httpServer.Addr))
	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		rlog.Error("HTTP server error", err)
		cancel() // Trigger graceful shutdown on server error
	}

	<-ctx.Done()

	rlog.Info("Server stopped gracefully")
}
