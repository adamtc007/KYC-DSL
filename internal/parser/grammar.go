package parser

// CurrentGrammarEBNF returns the canonical EBNF that aligns with the parser.
func CurrentGrammarEBNF() string {
	return `
(* KYC-DSL Grammar v1.0 *)

DSL              = { KycCase } ;

KycCase          = "(" "kyc-case" Identifier
                    { Section }
                   ")" ;

Section          = NaturePurpose
                 | ClientBusinessUnit
                 | Function
                 | Policy
                 | Obligation
                 | KycToken ;

NaturePurpose    = "(" "nature-purpose"
                      Nature
                      Purpose
                   ")" ;

Nature           = "(" "nature" QuotedString ")" ;
Purpose          = "(" "purpose" QuotedString ")" ;

ClientBusinessUnit = "(" "client-business-unit" Identifier ")" ;

Function         = "(" "function" Identifier ")" ;
Policy           = "(" "policy" Identifier ")" ;
Obligation       = "(" "obligation" Identifier ")" ;
KycToken         = "(" "kyc-token" QuotedString ")" ;

Identifier       = { Letter | Digit | "-" | "_" } ;
QuotedString     = '"' { Character | Space | Punctuation } '"' ;
`
}
