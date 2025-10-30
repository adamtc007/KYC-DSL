package lineage

import (
	"fmt"
	"time"

	"github.com/adamtc007/KYC-DSL/internal/model"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// EvaluationResult holds the outcome of a rule evaluation.
type EvaluationResult struct {
	DerivedCode string
	Value       any
	Success     bool
	Error       string
	Timestamp   time.Time
	Rule        string
	Inputs      map[string]any
}

// Evaluator runs derived-attribute rules in a sandboxed context.
type Evaluator struct {
	env     map[string]any
	program map[string]*vm.Program
	results []EvaluationResult
}

// NewEvaluator builds an evaluator with known public attributes.
func NewEvaluator(attrValues map[string]any) *Evaluator {
	return &Evaluator{
		env:     attrValues,
		program: make(map[string]*vm.Program),
		results: []EvaluationResult{},
	}
}

// CompileDerivations compiles all rule expressions ahead of time.
func (e *Evaluator) CompileDerivations(derivations []model.DerivedAttribute) error {
	for _, d := range derivations {
		prog, err := expr.Compile(d.RuleExpression, expr.Env(e.env))
		if err != nil {
			return fmt.Errorf("compile error for %s: %w", d.DerivedAttribute, err)
		}
		e.program[d.DerivedAttribute] = prog
	}
	return nil
}

// Evaluate runs all compiled expressions and stores results.
func (e *Evaluator) Evaluate(derivations []model.DerivedAttribute) []EvaluationResult {
	for _, d := range derivations {
		inputs := make(map[string]any)
		for _, src := range d.SourceAttributes {
			inputs[src] = e.env[src]
		}

		out := EvaluationResult{
			DerivedCode: d.DerivedAttribute,
			Rule:        d.RuleExpression,
			Inputs:      inputs,
			Timestamp:   time.Now(),
		}

		prog, ok := e.program[d.DerivedAttribute]
		if !ok {
			out.Success = false
			out.Error = "not compiled"
			e.results = append(e.results, out)
			continue
		}

		val, err := expr.Run(prog, e.env)
		if err != nil {
			out.Success = false
			out.Error = err.Error()
		} else {
			out.Success = true
			out.Value = val
			// Cascade: make derived value available for subsequent rules
			e.env[d.DerivedAttribute] = val
		}
		e.results = append(e.results, out)
	}
	return e.results
}

// Results returns all evaluation outcomes.
func (e *Evaluator) Results() []EvaluationResult {
	return e.results
}

// GetValue returns the evaluated value for a derived attribute.
func (e *Evaluator) GetValue(derivedCode string) (any, bool) {
	val, ok := e.env[derivedCode]
	return val, ok
}

// GetEnvironment returns the current evaluation environment (all attributes).
func (e *Evaluator) GetEnvironment() map[string]any {
	return e.env
}

// Reset clears evaluation results and derived values from environment.
func (e *Evaluator) Reset() {
	e.results = []EvaluationResult{}
	// Remove derived values from environment, keep only original public attributes
	for code := range e.program {
		delete(e.env, code)
	}
}

// ExplainResult generates human-readable explanation of evaluation.
func (r *EvaluationResult) ExplainResult() string {
	if !r.Success {
		return fmt.Sprintf("❌ %s failed: %s", r.DerivedCode, r.Error)
	}

	explanation := fmt.Sprintf("✅ %s = %v\n", r.DerivedCode, r.Value)
	explanation += fmt.Sprintf("   Rule: %s\n", r.Rule)
	explanation += "   Inputs:\n"
	for k, v := range r.Inputs {
		explanation += fmt.Sprintf("     • %s = %v\n", k, v)
	}
	explanation += fmt.Sprintf("   Evaluated at: %s", r.Timestamp.Format(time.RFC3339))
	return explanation
}
