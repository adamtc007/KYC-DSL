package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/adamtc007/KYC-DSL/api/pb/kycdata"
	pbOntology "github.com/adamtc007/KYC-DSL/api/pb/kycontology"
	"github.com/adamtc007/KYC-DSL/internal/dataservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	log.Println("🚀 Starting KYC Data Service...")
	log.Println()

	// Initialize database connection pool
	log.Println("📊 Initializing database connection pool...")
	if err := dataservice.InitDB(); err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}
	defer dataservice.CloseDB()

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Create and register Data Service (implements both Dictionary and Case services)
	dataService := dataservice.NewDataService()
	pb.RegisterDictionaryServiceServer(grpcServer, dataService)
	pb.RegisterCaseServiceServer(grpcServer, dataService)

	// Create and register Ontology Service (entities, CBUs, attributes, control graph)
	ontologyService := dataservice.NewOntologyService()
	pbOntology.RegisterOntologyServiceServer(grpcServer, ontologyService)

	// TODO: Dictionary and DocMaster services temporarily disabled for debugging
	// They are causing the gRPC server to hang/block on initialization
	//
	// Create and register Dictionary Service (attribute data model)
	// dictionaryService := dictionary.NewServer()
	// pbDictionary.RegisterDictionaryServiceServer(grpcServer, dictionaryService)
	//
	// Create and register DocMaster Service (document catalog)
	// docMasterService := docmaster.NewServer()
	// pbDocMaster.RegisterDocMasterServiceServer(grpcServer, docMasterService)

	// Enable gRPC reflection for grpcurl/grpcui
	reflection.Register(grpcServer)

	// Listen on port 50070
	lis, err := net.Listen("tcp", ":50070") //nolint:gosec
	if err != nil {
		log.Fatalf("❌ Failed to listen on :50070: %v", err)
	}

	log.Println("✅ Data Service initialized successfully")
	log.Println()
	log.Println("📋 Available services:")
	log.Println("   • kyc.data.DictionaryService - Ontology data (attributes, documents)")
	log.Println("   • kyc.data.CaseService - Case version management")
	log.Println("   • kyc.ontology.OntologyService - Full ontology API (entities, CBUs, control graph)")
	log.Println("   • kyc.dictionary.DictionaryService - [DISABLED - debugging]")
	log.Println("   • kyc.docmaster.DocMasterService - [DISABLED - debugging]")
	log.Println()
	log.Println("🌐 gRPC server listening on :50070")
	log.Println()
	log.Println("💡 Test with grpcurl:")
	log.Println("   grpcurl -plaintext localhost:50070 list")
	log.Println("   grpcurl -plaintext localhost:50070 kyc.ontology.OntologyService/SearchAttributes -d '{\"query\":\"ownership\",\"limit\":10}'")
	log.Println("   grpcurl -plaintext localhost:50070 kyc.ontology.OntologyService/ListEntities -d '{\"limit\":5}'")
	log.Println("   grpcurl -plaintext localhost:50070 kyc.data.DictionaryService/ListAttributes -d '{\"limit\":5}'")
	log.Println()
	log.Println("🔗 Consumer clients:")
	log.Println("   • Go CLI (kycctl) - connects to this service for data operations")
	log.Println("   • Rust DSL Engine - can query ontology data via gRPC")
	log.Println("   • UI/Frontend - case and dictionary data access")
	log.Println()

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println()
		log.Println("🛑 Shutting down gracefully...")
		grpcServer.GracefulStop()
	}()

	// Start serving
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}
