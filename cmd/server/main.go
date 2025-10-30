package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/adamtc007/KYC-DSL/api/pb"
	"github.com/adamtc007/KYC-DSL/internal/service"
	"github.com/adamtc007/KYC-DSL/internal/storage"
)

func main() {
	log.Println("🚀 Starting gRPC Server...")

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

	if err := db.Ping(); err != nil {
		log.Fatalf("❌ Database ping failed: %v", err)
	}
	log.Println("✅ Database connected successfully")

	// Create gRPC server
	grpcSrv := grpc.NewServer()

	// Register services
	pb.RegisterKycCaseServiceServer(grpcSrv, service.NewKycCaseService(db))
	pb.RegisterDslServiceServer(grpcSrv, service.NewDslService(db))
	pb.RegisterRagServiceServer(grpcSrv, service.NewRagService(db))

	// Enable reflection for grpcurl
	reflection.Register(grpcSrv)

	// Listen on port 50051
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("❌ Failed to listen: %v", err)
	}

	log.Println("🌐 gRPC server listening on :50051")
	log.Println()
	log.Println("📋 Available services:")
	log.Println("   • kyc.KycCaseService")
	log.Println("   • kyc.dsl.DslService")
	log.Println("   • kyc.rag.RagService")
	log.Println()
	log.Println("💡 Test with grpcurl:")
	log.Println("   grpcurl -plaintext localhost:50051 list")
	log.Println("   grpcurl -plaintext localhost:50051 kyc.rag.RagService/HealthCheck")
	log.Println()

	if err := grpcSrv.Serve(lis); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}
