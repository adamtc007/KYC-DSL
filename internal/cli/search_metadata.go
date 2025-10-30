package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/adamtc007/KYC-DSL/internal/ontology"
	"github.com/adamtc007/KYC-DSL/internal/rag"
	"github.com/adamtc007/KYC-DSL/internal/storage"
)

// RunSearchMetadataCommand performs semantic search on attribute metadata
func RunSearchMetadataCommand(query string, limit int) error {
	if query == "" {
		return fmt.Errorf("search query cannot be empty")
	}

	if limit <= 0 {
		limit = 10
	}

	fmt.Printf("🔍 Semantic Search: \"%s\"\n", query)
	fmt.Println("================================================")

	// Connect to database
	db, err := storage.ConnectPostgres()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Initialize repositories and embedder
	repo := ontology.NewMetadataRepo(db)
	embedder := rag.NewEmbedder()
	ctx := context.Background()

	// Generate embedding for the query
	fmt.Println("\n⚡ Generating query embedding...")
	queryEmbedding, err := embedder.GenerateEmbeddingFromText(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Perform vector search
	fmt.Printf("🔎 Searching for top %d matches...\n\n", limit)
	results, err := repo.SearchByVector(ctx, queryEmbedding, limit)
	if err != nil {
		return fmt.Errorf("failed to search: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("❌ No results found.")
		return nil
	}

	// Display results
	fmt.Printf("📊 Found %d matches:\n\n", len(results))

	for i, result := range results {
		fmt.Printf("─────────────────────────────────────────────────\n")
		fmt.Printf("Rank #%d\n", i+1)
		fmt.Printf("─────────────────────────────────────────────────\n")
		fmt.Printf("🏷️  Code:           %s\n", result.AttributeCode)
		fmt.Printf("📈 Similarity:      %.4f (distance: %.4f)\n", result.SimilarityScore, result.Distance)
		fmt.Printf("⚠️  Risk Level:      %s\n", result.RiskLevel)
		fmt.Printf("📝 Data Type:       %s\n", result.DataType)

		if len(result.Synonyms) > 0 {
			fmt.Printf("🔤 Synonyms:        %s\n", strings.Join(result.Synonyms, ", "))
		}

		if result.BusinessContext != "" {
			context := result.BusinessContext
			if len(context) > 150 {
				context = context[:150] + "..."
			}
			fmt.Printf("📖 Context:         %s\n", context)
		}

		if len(result.RegulatoryCitations) > 0 {
			fmt.Printf("📜 Citations:       %s\n", strings.Join(result.RegulatoryCitations, ", "))
		}

		if len(result.ExampleValues) > 0 && len(result.ExampleValues) <= 5 {
			fmt.Printf("💡 Examples:        %s\n", strings.Join(result.ExampleValues, ", "))
		}

		fmt.Println()
	}

	fmt.Println("================================================")
	fmt.Printf("✅ Search complete! Found %d relevant attributes.\n", len(results))

	return nil
}

// RunSimilarAttributesCommand finds attributes similar to a given attribute
func RunSimilarAttributesCommand(attributeCode string, limit int) error {
	if attributeCode == "" {
		return fmt.Errorf("attribute code cannot be empty")
	}

	if limit <= 0 {
		limit = 10
	}

	fmt.Printf("🔍 Finding Similar Attributes to: %s\n", attributeCode)
	fmt.Println("================================================")

	// Connect to database
	db, err := storage.ConnectPostgres()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Initialize repository
	repo := ontology.NewMetadataRepo(db)
	ctx := context.Background()

	// Get the source attribute first
	fmt.Println("\n📋 Source Attribute:")
	sourceMetadata, err := repo.GetMetadata(ctx, attributeCode)
	if err != nil {
		return fmt.Errorf("failed to get source attribute: %w", err)
	}

	fmt.Printf("  Code:        %s\n", sourceMetadata.AttributeCode)
	fmt.Printf("  Risk Level:  %s\n", sourceMetadata.RiskLevel)
	fmt.Printf("  Data Type:   %s\n", sourceMetadata.DataType)
	if len(sourceMetadata.Synonyms) > 0 {
		fmt.Printf("  Synonyms:    %s\n", strings.Join(sourceMetadata.Synonyms, ", "))
	}
	if sourceMetadata.BusinessContext != "" {
		context := sourceMetadata.BusinessContext
		if len(context) > 150 {
			context = context[:150] + "..."
		}
		fmt.Printf("  Context:     %s\n", context)
	}

	// Find similar attributes
	fmt.Printf("\n🔎 Finding top %d similar attributes...\n\n", limit)
	results, err := repo.FindSimilarAttributes(ctx, attributeCode, limit)
	if err != nil {
		return fmt.Errorf("failed to find similar attributes: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("❌ No similar attributes found.")
		return nil
	}

	// Display results
	fmt.Printf("📊 Found %d similar attributes:\n\n", len(results))

	for i, result := range results {
		fmt.Printf("─────────────────────────────────────────────────\n")
		fmt.Printf("Rank #%d\n", i+1)
		fmt.Printf("─────────────────────────────────────────────────\n")
		fmt.Printf("🏷️  Code:           %s\n", result.AttributeCode)
		fmt.Printf("📈 Similarity:      %.4f (distance: %.4f)\n", result.SimilarityScore, result.Distance)
		fmt.Printf("⚠️  Risk Level:      %s\n", result.RiskLevel)
		fmt.Printf("📝 Data Type:       %s\n", result.DataType)

		if len(result.Synonyms) > 0 {
			fmt.Printf("🔤 Synonyms:        %s\n", strings.Join(result.Synonyms, ", "))
		}

		if result.BusinessContext != "" {
			context := result.BusinessContext
			if len(context) > 120 {
				context = context[:120] + "..."
			}
			fmt.Printf("📖 Context:         %s\n", context)
		}

		fmt.Println()
	}

	fmt.Println("================================================")
	fmt.Printf("✅ Search complete! Found %d similar attributes.\n", len(results))

	// Suggest potential clustering
	if len(results) > 0 {
		fmt.Println("\n💡 Clustering Suggestion:")
		fmt.Printf("   These attributes could form a cluster with %s\n", attributeCode)
		fmt.Println("   based on semantic similarity.")
	}

	return nil
}

// RunTextSearchCommand performs traditional text-based search
func RunTextSearchCommand(searchTerm string) error {
	if searchTerm == "" {
		return fmt.Errorf("search term cannot be empty")
	}

	fmt.Printf("🔍 Text Search: \"%s\"\n", searchTerm)
	fmt.Println("================================================")

	// Connect to database
	db, err := storage.ConnectPostgres()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Initialize repository
	repo := ontology.NewMetadataRepo(db)
	ctx := context.Background()

	// Perform text search
	fmt.Println("\n🔎 Searching attributes and synonyms...")
	results, err := repo.SearchByText(ctx, searchTerm)
	if err != nil {
		return fmt.Errorf("failed to search: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("❌ No results found.")
		return nil
	}

	// Display results
	fmt.Printf("📊 Found %d matches:\n\n", len(results))

	for i, result := range results {
		fmt.Printf("─────────────────────────────────────────────────\n")
		fmt.Printf("Result #%d\n", i+1)
		fmt.Printf("─────────────────────────────────────────────────\n")
		fmt.Printf("🏷️  Code:           %s\n", result.AttributeCode)
		fmt.Printf("⚠️  Risk Level:      %s\n", result.RiskLevel)
		fmt.Printf("📝 Data Type:       %s\n", result.DataType)

		if len(result.Synonyms) > 0 {
			fmt.Printf("🔤 Synonyms:        %s\n", strings.Join(result.Synonyms, ", "))
		}

		if result.BusinessContext != "" {
			context := result.BusinessContext
			if len(context) > 150 {
				context = context[:150] + "..."
			}
			fmt.Printf("📖 Context:         %s\n", context)
		}

		if len(result.ExampleValues) > 0 && len(result.ExampleValues) <= 5 {
			fmt.Printf("💡 Examples:        %s\n", strings.Join(result.ExampleValues, ", "))
		}

		fmt.Println()
	}

	fmt.Println("================================================")
	fmt.Printf("✅ Search complete! Found %d matching attributes.\n", len(results))

	return nil
}

// RunMetadataStatsCommand displays statistics about the metadata repository
func RunMetadataStatsCommand() error {
	fmt.Println("📊 Attribute Metadata Statistics")
	fmt.Println("================================================")

	// Connect to database
	db, err := storage.ConnectPostgres()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Initialize repository
	repo := ontology.NewMetadataRepo(db)
	ctx := context.Background()

	// Get stats
	stats, err := repo.GetMetadataStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	fmt.Println("\n📈 Overview:")
	fmt.Printf("  Total Attributes:         %v\n", stats["total_attributes"])
	fmt.Printf("  With Embeddings:          %v\n", stats["attributes_with_embeddings"])
	fmt.Printf("  Embedding Coverage:       %.1f%%\n", stats["embedding_coverage_percent"])

	// Risk distribution
	if riskDist, ok := stats["risk_distribution"].([]struct {
		RiskLevel string `db:"risk_level"`
		Count     int    `db:"count"`
	}); ok && len(riskDist) > 0 {
		fmt.Println("\n⚠️  Risk Level Distribution:")
		for _, rd := range riskDist {
			fmt.Printf("  %-12s  %d attributes\n", rd.RiskLevel, rd.Count)
		}
	}

	// Get count of attributes without embeddings
	noEmbeddings, err := repo.GetAttributesWithoutEmbeddings(ctx)
	if err != nil {
		return fmt.Errorf("failed to get attributes without embeddings: %w", err)
	}

	if len(noEmbeddings) > 0 {
		fmt.Printf("\n⚠️  Attributes Missing Embeddings: %d\n", len(noEmbeddings))
		fmt.Println("   Run './kycctl seed-metadata' to generate embeddings")
	} else {
		fmt.Println("\n✅ All attributes have embeddings!")
	}

	fmt.Println("\n================================================")
	fmt.Println("✅ Statistics retrieved successfully.")

	return nil
}
