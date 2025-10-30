package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adamtc007/KYC-DSL/internal/api"
	"github.com/adamtc007/KYC-DSL/internal/rag"
	"github.com/adamtc007/KYC-DSL/internal/storage"
)

const (
	defaultPort         = "8080"
	defaultReadTimeout  = 15 * time.Second
	defaultWriteTimeout = 15 * time.Second
	defaultIdleTimeout  = 60 * time.Second
)

func main() {
	log.Println("🚀 Starting KYC-DSL RAG API Server...")

	// Get configuration from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Check OpenAI API key
	if os.Getenv("OPENAI_API_KEY") == "" {
		log.Fatal("❌ OPENAI_API_KEY environment variable not set")
	}

	// Connect to database
	log.Println("📊 Connecting to PostgreSQL...")
	db, err := storage.ConnectPostgres()
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("❌ Database ping failed: %v", err)
	}
	log.Println("✅ Database connected successfully")

	// Initialize embedder
	log.Println("🧠 Initializing OpenAI embedder...")
	embedder := rag.NewEmbedder()
	log.Printf("   Model: %s\n", embedder.GetModel())
	log.Printf("   Dimensions: %d\n", embedder.GetDimensions())

	// Initialize RAG handler
	ragHandler := api.NewRagHandler(db, embedder)

	// Create HTTP router
	mux := http.NewServeMux()

	// RAG endpoints
	mux.HandleFunc("/rag/attribute_search", corsMiddleware(ragHandler.HandleAttributeSearch))
	mux.HandleFunc("/rag/attribute_search_enriched", corsMiddleware(ragHandler.HandleEnrichedAttributeSearch))
	mux.HandleFunc("/rag/similar_attributes", corsMiddleware(ragHandler.HandleSimilarAttributes))
	mux.HandleFunc("/rag/text_search", corsMiddleware(ragHandler.HandleTextSearch))
	mux.HandleFunc("/rag/stats", corsMiddleware(ragHandler.HandleMetadataStats))
	mux.HandleFunc("/rag/health", corsMiddleware(ragHandler.HandleHealth))
	mux.HandleFunc("/rag/attribute/", corsMiddleware(ragHandler.HandleGetAttribute))

	// RAG Feedback endpoints
	mux.HandleFunc("/rag/feedback", corsMiddleware(ragHandler.HandleFeedback))
	mux.HandleFunc("/rag/feedback/recent", corsMiddleware(ragHandler.HandleRecentFeedback))
	mux.HandleFunc("/rag/feedback/analytics", corsMiddleware(ragHandler.HandleFeedbackAnalytics))
	mux.HandleFunc("/rag/feedback/attribute/", corsMiddleware(ragHandler.HandleFeedbackByAttribute))
	mux.HandleFunc("/rag/feedback/summary", corsMiddleware(ragHandler.HandleFeedbackSummary))

	// Root endpoint
	mux.HandleFunc("/", corsMiddleware(handleRoot))

	// Create server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		IdleTimeout:  defaultIdleTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🌐 Server listening on http://localhost:%s\n", port)
		log.Println("\n📋 Available endpoints:")
		log.Println("   GET  /                                   - API documentation")
		log.Println("   GET  /rag/health                         - Health check")
		log.Println("   GET  /rag/stats                          - Metadata statistics")
		log.Println("   GET  /rag/attribute_search?q=<query>     - Semantic search")
		log.Println("   GET  /rag/attribute_search_enriched?q=<query> - Enriched search with docs & regs")
		log.Println("   GET  /rag/similar_attributes?code=<code> - Similar attributes")
		log.Println("   GET  /rag/text_search?term=<term>        - Text search")
		log.Println("   GET  /rag/attribute/<code>               - Get attribute metadata")
		log.Println("   POST /rag/feedback                       - Submit feedback")
		log.Println("   GET  /rag/feedback/recent                - Recent feedback")
		log.Println("   GET  /rag/feedback/analytics             - Feedback analytics")
		log.Println("   GET  /rag/feedback/attribute/<code>      - Feedback by attribute")
		log.Println("   GET  /rag/feedback/summary               - Feedback summary")
		log.Println()

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("\n🛑 Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("⚠️  Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server stopped gracefully")
}

// handleRoot returns API documentation
func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
    <title>KYC-DSL RAG API</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; max-width: 1200px; margin: 40px auto; padding: 0 20px; }
        h1 { color: #2563eb; }
        h2 { color: #1e40af; margin-top: 30px; }
        .endpoint { background: #f3f4f6; padding: 15px; margin: 10px 0; border-radius: 8px; border-left: 4px solid #2563eb; }
        .method { display: inline-block; background: #2563eb; color: white; padding: 4px 12px; border-radius: 4px; font-weight: bold; font-size: 14px; }
        .path { font-family: "Courier New", monospace; font-size: 16px; margin-left: 10px; }
        .description { margin-top: 10px; color: #4b5563; }
        .param { background: #e5e7eb; padding: 2px 8px; border-radius: 4px; font-family: monospace; font-size: 14px; }
        .example { background: #1f2937; color: #f3f4f6; padding: 15px; border-radius: 8px; font-family: "Courier New", monospace; font-size: 14px; overflow-x: auto; margin-top: 10px; }
        a { color: #2563eb; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <h1>🚀 KYC-DSL RAG API</h1>
    <p>Semantic search and vector similarity API for regulatory compliance attributes.</p>

    <h2>📊 Health & Monitoring</h2>

    <div class="endpoint">
        <span class="method">GET</span><span class="path">/rag/health</span>
        <div class="description">Health check endpoint. Returns server status and embedding configuration.</div>
        <div class="example">curl http://localhost:8080/rag/health</div>
    </div>

    <div class="endpoint">
        <span class="method">GET</span><span class="path">/rag/stats</span>
        <div class="description">Get metadata repository statistics including embedding coverage and risk distribution.</div>
        <div class="example">curl http://localhost:8080/rag/stats</div>
    </div>

    <h2>🔍 Search Endpoints</h2>

    <div class="endpoint">
        <span class="method">GET</span><span class="path">/rag/attribute_search</span>
        <div class="description">
            Semantic search for attributes using vector embeddings.
            <br><strong>Parameters:</strong>
            <br>• <span class="param">q</span> (required) - Search query
            <br>• <span class="param">limit</span> (optional) - Max results (default: 10)
        </div>
        <div class="example">curl "http://localhost:8080/rag/attribute_search?q=tax%20reporting%20requirements&limit=5"</div>
    </div>

    <div class="endpoint">
        <span class="method">GET</span><span class="path">/rag/similar_attributes</span>
        <div class="description">
            Find attributes similar to a given attribute code.
            <br><strong>Parameters:</strong>
            <br>• <span class="param">code</span> (required) - Source attribute code
            <br>• <span class="param">limit</span> (optional) - Max results (default: 10)
        </div>
        <div class="example">curl "http://localhost:8080/rag/similar_attributes?code=UBO_NAME&limit=5"</div>
    </div>

    <div class="endpoint">
        <span class="method">GET</span><span class="path">/rag/text_search</span>
        <div class="description">
            Traditional text-based search (no embedding required).
            <br><strong>Parameters:</strong>
            <br>• <span class="param">term</span> (required) - Search term
        </div>
        <div class="example">curl "http://localhost:8080/rag/text_search?term=ownership"</div>
    </div>

    <div class="endpoint">
        <span class="method">GET</span><span class="path">/rag/attribute/{code}</span>
        <div class="description">Get complete metadata for a specific attribute code.</div>
        <div class="example">curl http://localhost:8080/rag/attribute/TAX_RESIDENCY_COUNTRY</div>
    </div>

    <h2>🔄 Feedback Endpoints</h2>

    <div class="endpoint">
        <span class="method">POST</span><span class="path">/rag/feedback</span>
        <div class="description">
            Submit feedback on search results to improve relevance scores.
            <br><strong>Body Parameters (JSON):</strong>
            <br>• <span class="param">query_text</span> (required) - Original search query
            <br>• <span class="param">attribute_code</span> (optional) - Attribute being rated
            <br>• <span class="param">document_code</span> (optional) - Document being rated
            <br>• <span class="param">regulation_code</span> (optional) - Regulation being rated
            <br>• <span class="param">feedback</span> (optional) - positive/negative/neutral (default: positive)
            <br>• <span class="param">confidence</span> (optional) - 0.0-1.0 weight (default: 1.0)
            <br>• <span class="param">agent_name</span> (optional) - Name of feedback provider
            <br>• <span class="param">agent_type</span> (optional) - human/ai/automated (default: human)
        </div>
        <div class="example">curl -X POST http://localhost:8080/rag/feedback \
  -H "Content-Type: application/json" \
  -d '{"query_text":"tax requirements","attribute_code":"TAX_RESIDENCY_COUNTRY","feedback":"positive","confidence":0.9}'</div>
    </div>

    <div class="endpoint">
        <span class="method">GET</span><span class="path">/rag/feedback/recent</span>
        <div class="description">
            Get recent feedback entries.
            <br><strong>Parameters:</strong>
            <br>• <span class="param">limit</span> (optional) - Max results (default: 50)
        </div>
        <div class="example">curl "http://localhost:8080/rag/feedback/recent?limit=10"</div>
    </div>

    <div class="endpoint">
        <span class="method">GET</span><span class="path">/rag/feedback/analytics</span>
        <div class="description">
            Get comprehensive feedback analytics including sentiment distribution and trends.
            <br><strong>Parameters:</strong>
            <br>• <span class="param">top</span> (optional) - Top N attributes to include (default: 10)
        </div>
        <div class="example">curl "http://localhost:8080/rag/feedback/analytics?top=20"</div>
    </div>

    <div class="endpoint">
        <span class="method">GET</span><span class="path">/rag/feedback/attribute/{code}</span>
        <div class="description">Get all feedback for a specific attribute code.</div>
        <div class="example">curl http://localhost:8080/rag/feedback/attribute/UBO_NAME</div>
    </div>

    <div class="endpoint">
        <span class="method">GET</span><span class="path">/rag/feedback/summary</span>
        <div class="description">
            Get aggregated feedback summary by sentiment and agent type.
            <br><strong>Parameters:</strong>
            <br>• <span class="param">limit</span> (optional) - Max attributes (default: 20)
        </div>
        <div class="example">curl http://localhost:8080/rag/feedback/summary</div>
    </div>

    <h2>📖 Example Queries</h2>
</text>


    <div class="endpoint">
        <strong>Find tax-related attributes:</strong>
        <div class="example">curl "http://localhost:8080/rag/attribute_search?q=tax%20compliance"</div>
    </div>

    <div class="endpoint">
        <strong>Find ownership attributes:</strong>
        <div class="example">curl "http://localhost:8080/rag/attribute_search?q=beneficial%20ownership"</div>
    </div>

    <div class="endpoint">
        <strong>Find risk indicators:</strong>
        <div class="example">curl "http://localhost:8080/rag/attribute_search?q=money%20laundering%20risk"</div>
    </div>

    <div class="endpoint">
        <strong>Find similar attributes:</strong>
        <div class="example">curl "http://localhost:8080/rag/similar_attributes?code=PEP_STATUS"</div>
    </div>

    <h2>📚 Documentation</h2>
    <p>
        For detailed documentation, see:
        <br>• <a href="https://github.com/adamtc007/KYC-DSL">GitHub Repository</a>
        <br>• RAG_VECTOR_SEARCH.md
        <br>• RAG_QUICKSTART.md
    </p>

    <hr style="margin: 40px 0;">
    <p style="color: #6b7280; font-size: 14px;">
        KYC-DSL v1.4 • Powered by OpenAI text-embedding-3-large • PostgreSQL + pgvector
    </p>
</body>
</html>`)
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		log.Printf("%s %s %d %v %s",
			r.Method,
			r.RequestURI,
			rw.statusCode,
			duration,
			r.RemoteAddr,
		)
	})
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write ensures status code is captured even if WriteHeader isn't called
func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}
