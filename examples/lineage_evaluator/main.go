package main

import (
	"fmt"
	"log"

	"github.com/adamtc007/KYC-DSL/internal/lineage"
	"github.com/adamtc007/KYC-DSL/internal/model"
)

func main() {
	fmt.Println("=== Lineage Evaluator Example ===")
	fmt.Println()

	// =====================================================
	// 1. Setup: Public attribute values (from documents)
	// =====================================================
	fmt.Println("📋 Step 1: Loading public attribute values from case data...")
	caseData := map[string]any{
		"TAX_RESIDENCY_COUNTRY":      "IR",
		"INCORPORATION_JURISDICTION": "US",
		"REGISTERED_NAME":            "BlackRock Global Fund",
		"INCORPORATION_DATE":         "2010-01-15",
		"UBO_NAME":                   []string{"Larry Fink", "Institutional Investors", "Vanguard Group"},
		"UBO_PERCENT":                []float64{35.0, 45.0, 20.0},
		"PEP_STATUS":                 false,
	}

	for k, v := range caseData {
		fmt.Printf("  • %s = %v\n", k, v)
	}
	fmt.Println()

	// =====================================================
	// 2. Define derived attributes with rules
	// =====================================================
	fmt.Println("📋 Step 2: Defining derived attributes with rule expressions...")
	derivations := []model.DerivedAttribute{
		{
			DerivedAttribute: "HIGH_RISK_JURISDICTION_FLAG",
			SourceAttributes: []string{"TAX_RESIDENCY_COUNTRY"},
			RuleExpression:   `TAX_RESIDENCY_COUNTRY in ["IR", "KP", "SY", "YE", "AF", "MM"]`,
			Jurisdiction:     "GLOBAL",
			RegulationCode:   "AMLD5",
		},
		{
			DerivedAttribute: "SANCTIONED_COUNTRY_FLAG",
			SourceAttributes: []string{"TAX_RESIDENCY_COUNTRY", "INCORPORATION_JURISDICTION"},
			RuleExpression:   `TAX_RESIDENCY_COUNTRY in ["IR", "KP", "SY", "CU", "RU"] || INCORPORATION_JURISDICTION in ["IR", "KP", "SY", "CU", "RU"]`,
			Jurisdiction:     "GLOBAL",
			RegulationCode:   "BSAAML",
		},
		{
			DerivedAttribute: "PEP_EXPOSURE_FLAG",
			SourceAttributes: []string{"PEP_STATUS"},
			RuleExpression:   `PEP_STATUS == true`,
			Jurisdiction:     "GLOBAL",
			RegulationCode:   "AMLD5",
		},
		{
			DerivedAttribute: "UBO_CONCENTRATION_SCORE",
			SourceAttributes: []string{"UBO_PERCENT"},
			RuleExpression:   `max(UBO_PERCENT)`,
			Jurisdiction:     "GLOBAL",
			RegulationCode:   "AMLD5",
		},
		{
			DerivedAttribute: "COMPLEX_STRUCTURE_FLAG",
			SourceAttributes: []string{"UBO_NAME"},
			RuleExpression:   `len(UBO_NAME) > 3`,
			Jurisdiction:     "GLOBAL",
			RegulationCode:   "AMLD5",
		},
	}

	for _, d := range derivations {
		fmt.Printf("  • %s\n", d.DerivedAttribute)
		fmt.Printf("    Sources: %v\n", d.SourceAttributes)
		fmt.Printf("    Rule: %s\n", d.RuleExpression)
	}
	fmt.Println()

	// =====================================================
	// 3. Create evaluator and compile rules
	// =====================================================
	fmt.Println("📋 Step 3: Compiling rule expressions...")
	eval := lineage.NewEvaluator(caseData)

	if err := eval.CompileDerivations(derivations); err != nil {
		log.Fatalf("❌ Compilation failed: %v", err)
	}
	fmt.Println("✅ All rules compiled successfully")
	fmt.Println()

	// =====================================================
	// 4. Evaluate all derived attributes
	// =====================================================
	fmt.Println("📋 Step 4: Evaluating derived attributes...")
	results := eval.Evaluate(derivations)
	fmt.Println()

	// =====================================================
	// 5. Display results with explanations
	// =====================================================
	fmt.Println("📊 Evaluation Results:")
	fmt.Println()

	successCount := 0
	failCount := 0

	for i, r := range results {
		fmt.Printf("Result #%d\n", i+1)
		fmt.Println(r.ExplainResult())
		fmt.Println()

		if r.Success {
			successCount++
		} else {
			failCount++
		}
	}

	// =====================================================
	// 6. Summary statistics
	// =====================================================
	fmt.Println("=== Summary ===")
	fmt.Printf("Total Evaluations: %d\n", len(results))
	fmt.Printf("✅ Successful: %d\n", successCount)
	fmt.Printf("❌ Failed: %d\n", failCount)
	fmt.Println()

	// =====================================================
	// 7. Access specific derived values
	// =====================================================
	fmt.Println("=== Derived Values ===")

	if val, ok := eval.GetValue("HIGH_RISK_JURISDICTION_FLAG"); ok {
		fmt.Printf("HIGH_RISK_JURISDICTION_FLAG = %v\n", val)
	}

	if val, ok := eval.GetValue("UBO_CONCENTRATION_SCORE"); ok {
		fmt.Printf("UBO_CONCENTRATION_SCORE = %v%%\n", val)
	}

	if val, ok := eval.GetValue("COMPLEX_STRUCTURE_FLAG"); ok {
		fmt.Printf("COMPLEX_STRUCTURE_FLAG = %v\n", val)
	}
	fmt.Println()

	// =====================================================
	// 8. Demonstrate explainability
	// =====================================================
	fmt.Println("=== Explainability Demo ===")
	fmt.Println("Question: Why is HIGH_RISK_JURISDICTION_FLAG = true?")
	fmt.Println("Answer:")

	for _, r := range results {
		if r.DerivedCode == "HIGH_RISK_JURISDICTION_FLAG" {
			fmt.Printf("  The flag is %v because:\n", r.Value)
			fmt.Printf("  - TAX_RESIDENCY_COUNTRY = '%v'\n", r.Inputs["TAX_RESIDENCY_COUNTRY"])
			fmt.Printf("  - Rule: %s\n", r.Rule)
			fmt.Printf("  - 'IR' is in the high-risk jurisdiction list\n")
			break
		}
	}
	fmt.Println()

	// =====================================================
	// 9. Demonstrate cascading evaluations
	// =====================================================
	fmt.Println("=== Cascading Evaluation Demo ===")
	fmt.Println("Derived values are available for subsequent rules:")

	env := eval.GetEnvironment()
	fmt.Println("\nFinal Environment (Public + Derived):")
	for k, v := range env {
		fmt.Printf("  %s = %v\n", k, v)
	}
	fmt.Println()

	// =====================================================
	// 10. Error handling demo
	// =====================================================
	fmt.Println("=== Error Handling Demo ===")
	fmt.Println("Testing invalid rule expression...")

	invalidDerivation := []model.DerivedAttribute{
		{
			DerivedAttribute: "INVALID_RULE",
			SourceAttributes: []string{"TAX_RESIDENCY_COUNTRY"},
			RuleExpression:   `NONEXISTENT_FIELD == "test"`,
		},
	}

	eval2 := lineage.NewEvaluator(caseData)
	if err := eval2.CompileDerivations(invalidDerivation); err != nil {
		fmt.Printf("✅ Compilation correctly failed: %v\n", err)
	}
	fmt.Println()

	fmt.Println("=== Example Complete ===")
}
