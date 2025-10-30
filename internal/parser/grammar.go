package parser

// CurrentGrammarEBNF returns the canonical EBNF that aligns with the parser.
// Version 1.1 adds Ownership and Control constructs.
func CurrentGrammarEBNF() string {
	return `
(* KYC-DSL Grammar v1.1 *)

DSL              = { KycCase } ;

KycCase          = "(" "kyc-case" Identifier
                    { Section }
                   ")" ;

Section          = NaturePurpose
                 | ClientBusinessUnit
                 | Function
                 | Policy
                 | Obligation
                 | KycToken
                 | OwnershipStructure ;

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

(* Ownership & Control Section *)

OwnershipStructure = "(" "ownership-structure"
                        { OwnershipNode }
                     ")" ;

OwnershipNode      = Owner | BeneficialOwner | Controller ;

Owner              = "(" "owner" Identifier Number "%" ")" ;
BeneficialOwner    = "(" "beneficial-owner" Identifier Number "%" ")" ;
Controller         = "(" "controller" Identifier QuotedString ")" ;

Identifier       = { Letter | Digit | "-" | "_" } ;
QuotedString     = '"' { Character | Space | Punctuation } '"' ;
Number           = Digit { Digit } ;
`
}
