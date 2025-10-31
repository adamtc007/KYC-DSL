package rustclient

import (
	"context"
	"fmt"
	"os"
	"time"

	pb "github.com/adamtc007/KYC-DSL/api/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DslClient wraps the Rust DSL gRPC service
type DslClient struct {
	conn   *grpc.ClientConn
	client pb.DslServiceClient
	addr   string
}

// NewDslClient creates a new connection to the Rust DSL service
// Default address is localhost:50060, can be overridden with RUST_DSL_SERVICE_ADDR env var
func NewDslClient(addr string) (*DslClient, error) {
	if addr == "" {
		addr = os.Getenv("RUST_DSL_SERVICE_ADDR")
		if addr == "" {
			addr = "localhost:50060"
		}
	}

	// Create connection with retry logic
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Rust DSL service at %s: %w (is the service running?)", addr, err)
	}

	return &DslClient{
		conn:   conn,
		client: pb.NewDslServiceClient(conn),
		addr:   addr,
	}, nil
}

// Close closes the gRPC connection
func (c *DslClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// ParseDSL parses DSL text into structured format
func (c *DslClient) ParseDSL(dsl string) (*pb.ParseResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.Parse(ctx, &pb.ParseRequest{Dsl: dsl})
	if err != nil {
		return nil, fmt.Errorf("parse RPC failed: %w", err)
	}

	return resp, nil
}

// ValidateDSL checks if a DSL case is valid
func (c *DslClient) ValidateDSL(dsl string) (*pb.ValidationResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := c.client.Validate(ctx, &pb.ValidateRequest{Dsl: dsl})
	if err != nil {
		return nil, fmt.Errorf("validate RPC failed: %w", err)
	}

	return result, nil
}

// ValidateCaseByID validates an existing case by ID
func (c *DslClient) ValidateCaseByID(caseID string) (*pb.ValidationResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := c.client.Validate(ctx, &pb.ValidateRequest{CaseId: caseID})
	if err != nil {
		return nil, fmt.Errorf("validate RPC failed: %w", err)
	}

	return result, nil
}

// ExecuteCase runs a specific function on a KYC case
func (c *DslClient) ExecuteCase(caseID, functionName string) (*pb.ExecuteResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.Execute(ctx, &pb.ExecuteRequest{
		CaseId:       caseID,
		FunctionName: functionName,
	})
	if err != nil {
		return nil, fmt.Errorf("execute RPC failed: %w", err)
	}

	return resp, nil
}

// AmendCase applies a predefined amendment to a case
func (c *DslClient) AmendCase(caseName, amendmentType string) (*pb.AmendResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.client.Amend(ctx, &pb.AmendRequest{
		CaseName:      caseName,
		AmendmentType: amendmentType,
	})
	if err != nil {
		return nil, fmt.Errorf("amend RPC failed: %w", err)
	}

	return resp, nil
}

// SerializeCase converts structured case back to DSL
func (c *DslClient) SerializeCase(kycCase *pb.ParsedCase) (*pb.SerializeResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.Serialize(ctx, &pb.SerializeRequest{Case: kycCase})
	if err != nil {
		return nil, fmt.Errorf("serialize RPC failed: %w", err)
	}

	return resp, nil
}

// GetGrammar returns the current DSL grammar definition
func (c *DslClient) GetGrammar() (*pb.GrammarResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.GetGrammar(ctx, &pb.GetGrammarRequest{})
	if err != nil {
		return nil, fmt.Errorf("get grammar RPC failed: %w", err)
	}

	return resp, nil
}

// ListAmendments returns available amendment types
func (c *DslClient) ListAmendments() (*pb.ListAmendmentsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.ListAmendments(ctx, &pb.ListAmendmentsRequest{})
	if err != nil {
		return nil, fmt.Errorf("list amendments RPC failed: %w", err)
	}

	return resp, nil
}

// HealthCheck verifies the Rust DSL service is responsive
func (c *DslClient) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Use GetGrammar as a health check
	_, err := c.client.GetGrammar(ctx, &pb.GetGrammarRequest{})
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	return nil
}

// Addr returns the address of the connected Rust service
func (c *DslClient) Addr() string {
	return c.addr
}
