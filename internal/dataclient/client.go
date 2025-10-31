package dataclient

import (
	"context"
	"fmt"
	"time"

	pb "github.com/adamtc007/KYC-DSL/api/pb/kycdata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DataClient wraps the gRPC connection to the Data Service
type DataClient struct {
	conn           *grpc.ClientConn
	dictClient     pb.DictionaryServiceClient
	caseClient     pb.CaseServiceClient
	defaultTimeout time.Duration
}

// NewDataClient creates a new data service client
func NewDataClient(addr string) (*DataClient, error) {
	if addr == "" {
		addr = "localhost:50070"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to data service at %s: %w", addr, err)
	}

	return &DataClient{
		conn:           conn,
		dictClient:     pb.NewDictionaryServiceClient(conn),
		caseClient:     pb.NewCaseServiceClient(conn),
		defaultTimeout: 30 * time.Second,
	}, nil
}

// Close closes the gRPC connection
func (c *DataClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetAttribute retrieves an attribute by code
func (c *DataClient) GetAttribute(code string) (*pb.Attribute, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	req := &pb.GetAttributeRequest{
		Id: code,
	}

	resp, err := c.dictClient.GetAttribute(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get attribute %s: %w", code, err)
	}

	return resp, nil
}

// ListAttributes retrieves all attributes
func (c *DataClient) ListAttributes() ([]*pb.Attribute, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	req := &pb.ListAttributesRequest{}

	resp, err := c.dictClient.ListAttributes(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list attributes: %w", err)
	}

	return resp.Attributes, nil
}

// SaveCaseVersion saves a new version of a case
func (c *DataClient) SaveCaseVersion(caseName, dslText string) (*pb.CaseVersionResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	req := &pb.CaseVersionRequest{
		CaseId:    caseName,
		DslSource: dslText,
	}

	resp, err := c.caseClient.SaveCaseVersion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to save case version for %s: %w", caseName, err)
	}

	return resp, nil
}

// GetCaseVersion retrieves a specific version of a case
func (c *DataClient) GetCaseVersion(caseName string, version int32) (*pb.CaseVersion, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	req := &pb.GetCaseRequest{
		CaseId: caseName,
	}

	resp, err := c.caseClient.GetCaseVersion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get case %s version %d: %w", caseName, version, err)
	}

	return resp, nil
}

// GetLatestCaseVersion retrieves the latest version of a case
func (c *DataClient) GetLatestCaseVersion(caseName string) (*pb.CaseVersion, error) {
	return c.GetCaseVersion(caseName, 0) // 0 = latest
}

// ListCaseVersions retrieves all versions of a case
func (c *DataClient) ListCaseVersions(caseName string) ([]*pb.CaseVersion, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	req := &pb.ListCaseVersionsRequest{
		CaseId: caseName,
	}

	resp, err := c.caseClient.ListCaseVersions(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list case versions for %s: %w", caseName, err)
	}

	return resp.Versions, nil
}

// ListAllCases retrieves all cases with summary information
func (c *DataClient) ListAllCases(limit, offset int32, statusFilter string) ([]*pb.CaseSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	req := &pb.ListAllCasesRequest{
		Limit:        limit,
		Offset:       offset,
		StatusFilter: statusFilter,
	}

	resp, err := c.caseClient.ListAllCases(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list all cases: %w", err)
	}

	return resp.Cases, nil
}
