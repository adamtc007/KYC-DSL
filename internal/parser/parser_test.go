package parser

import (
	"strings"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple s-expression",
			input:    "(kyc-case TEST)",
			expected: []string{"(", "kyc-case", "TEST", ")"},
		},
		{
			name:     "quoted string",
			input:    `(nature "This is a test")`,
			expected: []string{"(", "nature", `"This is a test"`, ")"},
		},
		{
			name:     "quoted string with spaces",
			input:    `(purpose "Multi word purpose statement")`,
			expected: []string{"(", "purpose", `"Multi word purpose statement"`, ")"},
		},
		{
			name:     "nested expressions",
			input:    "(nature-purpose (nature \"test\") (purpose \"example\"))",
			expected: []string{"(", "nature-purpose", "(", "nature", `"test"`, ")", "(", "purpose", `"example"`, ")", ")"},
		},
		{
			name:     "comment line",
			input:    "; This is a comment\n(test value)",
			expected: []string{"(", "test", "value", ")"},
		},
		{
			name:     "hash comment",
			input:    "# Another comment\n(test value)",
			expected: []string{"(", "test", "value", ")"},
		},
		{
			name:     "empty lines",
			input:    "\n\n(test value)\n\n",
			expected: []string{"(", "test", "value", ")"},
		},
		{
			name:     "identifiers with hyphens",
			input:    "(client-business-unit AVIVA-EU-FUNDS)",
			expected: []string{"(", "client-business-unit", "AVIVA-EU-FUNDS", ")"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := tokenize(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("tokenize failed: %v", err)
			}
			if len(tokens) != len(tt.expected) {
				t.Fatalf("expected %d tokens, got %d\nexpected: %v\ngot: %v",
					len(tt.expected), len(tokens), tt.expected, tokens)
			}
			for i, tok := range tokens {
				if tok != tt.expected[i] {
					t.Errorf("token %d: expected %q, got %q", i, tt.expected[i], tok)
				}
			}
		})
	}
}

func TestParse(t *testing.T) {
	input := `
(kyc-case TEST-CASE
  (nature-purpose
    (nature "Test nature")
    (purpose "Test purpose")
  )
  (client-business-unit TEST-CBU)
  (function DISCOVER-POLICIES)
  (policy TEST-POLICY)
  (obligation TEST-OBL)
  (kyc-token "pending")
)
`
	dsl, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(dsl.Cases) != 1 {
		t.Fatalf("expected 1 case, got %d", len(dsl.Cases))
	}

	c := dsl.Cases[0]
	if c.Name != "TEST-CASE" {
		t.Errorf("expected case name 'TEST-CASE', got %q", c.Name)
	}

	if c.Body == nil {
		t.Fatal("case body is nil")
	}

	if len(c.Body.Args) < 6 {
		t.Errorf("expected at least 6 body sections, got %d", len(c.Body.Args))
	}
}

func TestBind(t *testing.T) {
	input := `
(kyc-case BIND-TEST
  (nature-purpose
    (nature "Binding test nature")
    (purpose "Binding test purpose")
  )
  (client-business-unit BIND-CBU)
  (function DISCOVER-POLICIES)
  (policy POL-001)
  (obligation OBL-001)
  (kyc-token "pending")
)
`
	dsl, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	cases, err := Bind(dsl)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	if len(cases) != 1 {
		t.Fatalf("expected 1 case, got %d", len(cases))
	}

	c := cases[0]

	// Check basic fields
	if c.Name != "BIND-TEST" {
		t.Errorf("expected name 'BIND-TEST', got %q", c.Name)
	}

	if c.Nature != "Binding test nature" {
		t.Errorf("expected nature 'Binding test nature', got %q", c.Nature)
	}

	if c.Purpose != "Binding test purpose" {
		t.Errorf("expected purpose 'Binding test purpose', got %q", c.Purpose)
	}

	if c.CBU.Name != "BIND-CBU" {
		t.Errorf("expected CBU 'BIND-CBU', got %q", c.CBU.Name)
	}

	// Check functions
	if len(c.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(c.Functions))
	}
	if c.Functions[0].Action != "DISCOVER-POLICIES" {
		t.Errorf("expected function 'DISCOVER-POLICIES', got %q", c.Functions[0].Action)
	}

	// Check policies
	if len(c.Policies) != 1 {
		t.Fatalf("expected 1 policy, got %d", len(c.Policies))
	}
	if c.Policies[0].Code != "POL-001" {
		t.Errorf("expected policy 'POL-001', got %q", c.Policies[0].Code)
	}

	// Check obligations
	if len(c.Obligations) != 1 {
		t.Fatalf("expected 1 obligation, got %d", len(c.Obligations))
	}
	if c.Obligations[0].PolicyCode != "OBL-001" {
		t.Errorf("expected obligation 'OBL-001', got %q", c.Obligations[0].PolicyCode)
	}

	// Check token
	if c.Token == nil {
		t.Fatal("token is nil")
	}
	if c.Token.Status != "pending" {
		t.Errorf("expected token status 'pending', got %q", c.Token.Status)
	}
}

func TestSerializeCases(t *testing.T) {
	input := `
(kyc-case SERIALIZE-TEST
  (nature-purpose
    (nature "Serialize nature")
    (purpose "Serialize purpose")
  )
  (client-business-unit SERIALIZE-CBU)
  (function EXTRACT-DATA)
  (policy POL-SERIALIZE)
  (obligation OBL-SERIALIZE)
  (kyc-token "approved")
)
`
	dsl, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	cases, err := Bind(dsl)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	serialized := SerializeCases(cases)

	// Check that serialized output contains key elements
	expectedStrings := []string{
		"(kyc-case SERIALIZE-TEST",
		"(nature-purpose",
		`(nature "Serialize nature")`,
		`(purpose "Serialize purpose")`,
		"(client-business-unit SERIALIZE-CBU)",
		"(function EXTRACT-DATA)",
		"(policy POL-SERIALIZE)",
		"(obligation OBL-SERIALIZE)",
		`(kyc-token "approved")`,
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(serialized, expected) {
			t.Errorf("serialized output missing expected string: %q", expected)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	input := `
(kyc-case ROUNDTRIP-TEST
  (nature-purpose
    (nature "Round trip nature")
    (purpose "Round trip purpose")
  )
  (client-business-unit RT-CBU)
  (function VERIFY-OWNERSHIP)
  (policy POL-RT-001)
  (obligation OBL-RT-001)
  (kyc-token "review")
)
`
	// Parse original
	dsl1, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("First parse failed: %v", err)
	}

	// Bind to models
	cases1, err := Bind(dsl1)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	// Serialize back to DSL
	serialized := SerializeCases(cases1)

	// Parse the serialized version
	dsl2, err := Parse(strings.NewReader(serialized))
	if err != nil {
		t.Fatalf("Second parse failed: %v", err)
	}

	// Bind again
	cases2, err := Bind(dsl2)
	if err != nil {
		t.Fatalf("Second bind failed: %v", err)
	}

	// Compare the two bound cases
	c1 := cases1[0]
	c2 := cases2[0]

	if c1.Name != c2.Name {
		t.Errorf("name mismatch: %q vs %q", c1.Name, c2.Name)
	}
	if c1.Nature != c2.Nature {
		t.Errorf("nature mismatch: %q vs %q", c1.Nature, c2.Nature)
	}
	if c1.Purpose != c2.Purpose {
		t.Errorf("purpose mismatch: %q vs %q", c1.Purpose, c2.Purpose)
	}
	if c1.CBU.Name != c2.CBU.Name {
		t.Errorf("CBU mismatch: %q vs %q", c1.CBU.Name, c2.CBU.Name)
	}
	if c1.Token.Status != c2.Token.Status {
		t.Errorf("token status mismatch: %q vs %q", c1.Token.Status, c2.Token.Status)
	}
}

func TestTrimQuotes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"quoted"`, "quoted"},
		{`"with spaces"`, "with spaces"},
		{"unquoted", "unquoted"},
		{`""`, ""},
		// Single quote doesn't get trimmed (not a pair)
	}

	for _, tt := range tests {
		result := trimQuotes(tt.input)
		if result != tt.expected {
			t.Errorf("trimQuotes(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestParseMultipleCases(t *testing.T) {
	input := `
(kyc-case CASE-ONE
  (nature-purpose
    (nature "First case")
    (purpose "First purpose")
  )
  (client-business-unit CBU-1)
  (kyc-token "pending")
)

(kyc-case CASE-TWO
  (nature-purpose
    (nature "Second case")
    (purpose "Second purpose")
  )
  (client-business-unit CBU-2)
  (kyc-token "approved")
)
`
	dsl, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(dsl.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(dsl.Cases))
	}

	if dsl.Cases[0].Name != "CASE-ONE" {
		t.Errorf("first case name: expected 'CASE-ONE', got %q", dsl.Cases[0].Name)
	}
	if dsl.Cases[1].Name != "CASE-TWO" {
		t.Errorf("second case name: expected 'CASE-TWO', got %q", dsl.Cases[1].Name)
	}
}
