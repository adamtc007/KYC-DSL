package cli

import (
	"fmt"

	pb "github.com/adamtc007/KYC-DSL/api/pb/kycdata"
	"github.com/adamtc007/KYC-DSL/internal/dataclient"
)

// RunGetCaseCommand retrieves and displays DSL from the database.
func RunGetCaseCommand(caseName string, version int) error {
	// Connect to data service
	client, err := dataclient.NewDataClient("")
	if err != nil {
		return fmt.Errorf("failed to connect to data service: %w", err)
	}
	defer client.Close()

	// Get case version via gRPC
	var caseVersion *pb.CaseVersion
	if version > 0 {
		caseVersion, err = client.GetCaseVersion(caseName, int32(version))
	} else {
		caseVersion, err = client.GetLatestCaseVersion(caseName)
	}

	if err != nil {
		return fmt.Errorf("failed to retrieve case: %w", err)
	}

	// Display metadata
	fmt.Printf("📦 Case: %s\n", caseVersion.CaseId)
	fmt.Printf("🔑 ID: %s\n", caseVersion.Id)
	fmt.Printf("📅 Created: %s\n", caseVersion.CreatedAt)
	fmt.Printf("📊 Status: %s\n", caseVersion.Status)
	fmt.Println("─────────────────────────────────────────────")
	fmt.Println()

	// Display DSL content
	fmt.Println(caseVersion.DslSource)
	fmt.Println()

	return nil
}

// RunListCaseVersionsCommand lists all versions of a case.
func RunListCaseVersionsCommand(caseName string) error {
	// Connect to data service
	client, err := dataclient.NewDataClient("")
	if err != nil {
		return fmt.Errorf("failed to connect to data service: %w", err)
	}
	defer client.Close()

	// Get version list via gRPC
	versions, err := client.ListCaseVersions(caseName)
	if err != nil {
		return fmt.Errorf("failed to list versions for case '%s': %w", caseName, err)
	}

	if len(versions) == 0 {
		fmt.Printf("ℹ️  No versions found for case: %s\n", caseName)
		return nil
	}

	// Display header
	fmt.Printf("📦 Case: %s\n", caseName)
	fmt.Printf("📊 Total Versions: %d\n\n", len(versions))
	fmt.Println("ID                                   │ Status    │ Created At")
	fmt.Println("─────────────────────────────────────┼───────────┼─────────────────────")

	// Display each version
	for _, v := range versions {
		shortId := v.Id
		if len(shortId) > 36 {
			shortId = shortId[:36]
		}
		fmt.Printf("%-36s │ %-9s │ %s\n", shortId, v.Status, v.CreatedAt)
	}
	fmt.Println()

	return nil
}

// RunListAllCasesCommand lists all cases in the database.
func RunListAllCasesCommand() error {
	// Connect to data service
	client, err := dataclient.NewDataClient("")
	if err != nil {
		return fmt.Errorf("failed to connect to data service: %w", err)
	}
	defer client.Close()

	// Get all cases via gRPC
	cases, err := client.ListAllCases(0, 0, "")
	if err != nil {
		return fmt.Errorf("failed to list cases: %w", err)
	}

	if len(cases) == 0 {
		fmt.Println("ℹ️  No cases found in database")
		return nil
	}

	// Display header
	fmt.Printf("📋 Total Cases: %d\n\n", len(cases))
	fmt.Println("Case Name                        │ Versions │ Status    │ Last Updated")
	fmt.Println("─────────────────────────────────┼──────────┼───────────┼─────────────────────")

	// Display each case
	for _, c := range cases {
		fmt.Printf("%-32s │ %-8d │ %-9s │ %s\n",
			truncate(c.CaseId, 32),
			c.VersionCount,
			c.Status,
			c.LastUpdated)
	}
	fmt.Println()

	return nil
}

// truncate truncates a string to maxLen and adds "..." if needed
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
