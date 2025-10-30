use kyc_dsl_core::{compile_dsl, execute_plan, parser};
use tonic::{transport::Server, Request, Response, Status};
use tonic_reflection::server::Builder as ReflectionBuilder;

// Include the generated protobuf code (suppress warnings from generated code)
#[allow(dead_code, unused_imports, clippy::all)]
pub mod kyc {
    pub mod dsl {
        tonic::include_proto!("kyc.dsl");
    }
}

use kyc::dsl::dsl_service_server::{DslService, DslServiceServer};
use kyc::dsl::*;

/// Rust implementation of the DSL service
#[derive(Debug, Default)]
pub struct RustDslServer;

#[tonic::async_trait]
impl DslService for RustDslServer {
    /// Execute a DSL case with specified function
    async fn execute(
        &self,
        request: Request<ExecuteRequest>,
    ) -> Result<Response<ExecuteResponse>, Status> {
        let req = request.into_inner();

        println!("Execute request for case: {}", req.case_id);

        // For now, we'll use a simple DSL source if not provided
        let dsl_source = format!(
            "(kyc-case {} (function {}))",
            req.case_id, req.function_name
        );

        match compile_dsl(&dsl_source).and_then(|plan| execute_plan(&plan)) {
            Ok(_result) => Ok(Response::new(ExecuteResponse {
                updated_dsl: dsl_source,
                message: format!(
                    "Executed function '{}' on case '{}'",
                    req.function_name, req.case_id
                ),
                success: true,
                case_id: req.case_id,
                new_version: 1,
            })),
            Err(e) => Ok(Response::new(ExecuteResponse {
                updated_dsl: String::new(),
                message: format!("Execution failed: {}", e),
                success: false,
                case_id: req.case_id,
                new_version: 0,
            })),
        }
    }

    /// Validate a DSL case
    async fn validate(
        &self,
        request: Request<ValidateRequest>,
    ) -> Result<Response<ValidationResult>, Status> {
        let req = request.into_inner();

        let dsl_source = if !req.dsl.is_empty() {
            req.dsl
        } else {
            format!("(kyc-case {})", req.case_id)
        };

        println!("Validating DSL: {}", dsl_source);

        match compile_dsl(&dsl_source) {
            Ok(_) => Ok(Response::new(ValidationResult {
                valid: true,
                errors: vec![],
                warnings: vec![],
                issues: vec![],
            })),
            Err(e) => Ok(Response::new(ValidationResult {
                valid: false,
                errors: vec![e.to_string()],
                warnings: vec![],
                issues: vec![ValidationIssue {
                    severity: "error".to_string(),
                    message: e.to_string(),
                    code: "PARSE_ERROR".to_string(),
                    line: 0,
                    column: 0,
                }],
            })),
        }
    }

    /// Parse DSL text into structured format
    async fn parse(
        &self,
        request: Request<ParseRequest>,
    ) -> Result<Response<ParseResponse>, Status> {
        let req = request.into_inner();

        println!("Parsing DSL: {}", req.dsl);

        match parser::parse(&req.dsl) {
            Ok(ast) => {
                // Extract case information from AST
                let case_info = extract_case_info(&ast);

                Ok(Response::new(ParseResponse {
                    success: true,
                    message: "Parse successful".to_string(),
                    cases: vec![case_info],
                    errors: vec![],
                }))
            }
            Err(e) => Ok(Response::new(ParseResponse {
                success: false,
                message: format!("Parse failed: {}", e),
                cases: vec![],
                errors: vec![format!("Parse error: {}", e)],
            })),
        }
    }

    /// Serialize structured case back to DSL
    async fn serialize(
        &self,
        request: Request<SerializeRequest>,
    ) -> Result<Response<SerializeResponse>, Status> {
        let req = request.into_inner();

        if let Some(case) = req.case {
            let dsl = serialize_case(&case);

            Ok(Response::new(SerializeResponse {
                success: true,
                dsl,
                message: "Serialization successful".to_string(),
            }))
        } else {
            Ok(Response::new(SerializeResponse {
                success: false,
                dsl: String::new(),
                message: "No case provided".to_string(),
            }))
        }
    }

    /// Apply an amendment to a case
    async fn amend(
        &self,
        request: Request<AmendRequest>,
    ) -> Result<Response<AmendResponse>, Status> {
        let req = request.into_inner();

        println!(
            "Amending case '{}' with '{}'",
            req.case_name, req.amendment_type
        );

        // Generate amended DSL
        let amended_dsl = format!(
            "(kyc-case {}\n  (amendment {})\n  (kyc-token \"updated\"))",
            req.case_name, req.amendment_type
        );

        // Compute a simple hash
        let hash = format!("{:x}", md5::compute(&amended_dsl));

        Ok(Response::new(AmendResponse {
            success: true,
            message: format!("Applied amendment '{}'", req.amendment_type),
            updated_dsl: amended_dsl,
            new_version: 2,
            sha256_hash: hash,
        }))
    }

    /// List available amendment types
    async fn list_amendments(
        &self,
        _request: Request<ListAmendmentsRequest>,
    ) -> Result<Response<ListAmendmentsResponse>, Status> {
        let amendments = vec![
            AmendmentType {
                name: "policy-discovery".to_string(),
                description: "Add policy discovery function and policies".to_string(),
                parameters: vec!["policy_code".to_string()],
            },
            AmendmentType {
                name: "document-solicitation".to_string(),
                description: "Add document solicitation and obligations".to_string(),
                parameters: vec![],
            },
            AmendmentType {
                name: "document-discovery".to_string(),
                description: "Auto-populate documents from ontology".to_string(),
                parameters: vec!["jurisdiction".to_string()],
            },
            AmendmentType {
                name: "ownership-discovery".to_string(),
                description: "Add ownership structure and control hierarchy".to_string(),
                parameters: vec![],
            },
            AmendmentType {
                name: "risk-assessment".to_string(),
                description: "Add risk assessment function".to_string(),
                parameters: vec![],
            },
            AmendmentType {
                name: "approve".to_string(),
                description: "Finalize case as approved".to_string(),
                parameters: vec![],
            },
            AmendmentType {
                name: "decline".to_string(),
                description: "Finalize case as declined".to_string(),
                parameters: vec![],
            },
        ];

        Ok(Response::new(ListAmendmentsResponse { amendments }))
    }

    /// Get the current DSL grammar definition
    async fn get_grammar(
        &self,
        _request: Request<GetGrammarRequest>,
    ) -> Result<Response<GrammarResponse>, Status> {
        let ebnf = r#"
KYC-DSL Grammar (v1.2)

case        = "(kyc-case" IDENT form* ")"
form        = "(nature-purpose" nature purpose ")"
            | "(ownership-structure" entity owner* beneficial-owner* controller* ")"
            | "(data-dictionary" attribute* ")"
            | "(document-requirements" jurisdiction required ")"
            | "(kyc-token" STRING ")"
            | simple-form

simple-form = "(" IDENT value* ")"
value       = STRING | IDENT | PERCENT | form
IDENT       = [A-Z][A-Z0-9_-]*
STRING      = '"' [^"]* '"'
PERCENT     = [0-9]+ "." [0-9]+ "%"
"#;

        Ok(Response::new(GrammarResponse {
            ebnf: ebnf.to_string(),
            version: "1.2".to_string(),
            created_at: None,
        }))
    }
}

/// Extract case information from parsed AST
fn extract_case_info(ast: &parser::Expr) -> ParsedCase {
    let mut case = ParsedCase {
        name: "UNKNOWN".to_string(),
        ..Default::default()
    };

    if let parser::Expr::Call(name, args) = ast {
        if name == "kyc-case" && !args.is_empty() {
            if let parser::Expr::Atom(case_name) = &args[0] {
                case.name = case_name.clone();
            }

            // Parse nested forms
            for arg in &args[1..] {
                if let parser::Expr::Call(form_name, form_args) = arg {
                    match form_name.as_str() {
                        "nature" => {
                            if let Some(parser::Expr::Atom(val)) = form_args.first() {
                                case.nature = val.clone();
                            }
                        }
                        "purpose" => {
                            if let Some(parser::Expr::Atom(val)) = form_args.first() {
                                case.purpose = val.clone();
                            }
                        }
                        "client-business-unit" => {
                            if let Some(parser::Expr::Atom(val)) = form_args.first() {
                                case.client_business_unit = val.clone();
                            }
                        }
                        "policy" => {
                            if let Some(parser::Expr::Atom(val)) = form_args.first() {
                                case.policy = val.clone();
                            }
                        }
                        "function" => {
                            if let Some(parser::Expr::Atom(val)) = form_args.first() {
                                case.function = val.clone();
                            }
                        }
                        "kyc-token" => {
                            if let Some(parser::Expr::Atom(val)) = form_args.first() {
                                case.kyc_token = val.clone();
                            }
                        }
                        _ => {}
                    }
                }
            }
        }
    }

    case
}

/// Serialize a ParsedCase back to DSL format
fn serialize_case(case: &ParsedCase) -> String {
    let mut dsl = format!("(kyc-case {}\n", case.name);

    if !case.nature.is_empty() || !case.purpose.is_empty() {
        dsl.push_str("  (nature-purpose\n");
        if !case.nature.is_empty() {
            dsl.push_str(&format!("    (nature \"{}\")\n", case.nature));
        }
        if !case.purpose.is_empty() {
            dsl.push_str(&format!("    (purpose \"{}\")\n", case.purpose));
        }
        dsl.push_str("  )\n");
    }

    if !case.client_business_unit.is_empty() {
        dsl.push_str(&format!(
            "  (client-business-unit {})\n",
            case.client_business_unit
        ));
    }

    if !case.policy.is_empty() {
        dsl.push_str(&format!("  (policy {})\n", case.policy));
    }

    if !case.function.is_empty() {
        dsl.push_str(&format!("  (function {})\n", case.function));
    }

    if !case.obligation.is_empty() {
        dsl.push_str(&format!("  (obligation {})\n", case.obligation));
    }

    // Ownership structure
    if let Some(ownership) = &case.ownership {
        dsl.push_str("  (ownership-structure\n");
        if !ownership.entity_name.is_empty() {
            dsl.push_str(&format!("    (entity {})\n", ownership.entity_name));
        }
        for owner in &ownership.owners {
            dsl.push_str(&format!(
                "    (owner {} {}%)\n",
                owner.name, owner.percentage
            ));
        }
        for bo in &ownership.beneficial_owners {
            dsl.push_str(&format!(
                "    (beneficial-owner {} {}%)\n",
                bo.name, bo.percentage
            ));
        }
        for controller in &ownership.controllers {
            dsl.push_str(&format!(
                "    (controller {} \"{}\")\n",
                controller.name, controller.role
            ));
        }
        dsl.push_str("  )\n");
    }

    if !case.kyc_token.is_empty() {
        dsl.push_str(&format!("  (kyc-token \"{}\")\n", case.kyc_token));
    }

    dsl.push(')');
    dsl
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let addr = "[::1]:50060".parse()?;
    let service = RustDslServer;

    println!("🦀 Rust DSL gRPC Service");
    println!("========================");
    println!("Listening on: {}", addr);
    println!("Protocol: gRPC (HTTP/2)");
    println!("Service: kyc.dsl.DslService");
    println!();
    println!("Available RPCs:");
    println!("  - Execute");
    println!("  - Validate");
    println!("  - Parse");
    println!("  - Serialize");
    println!("  - Amend");
    println!("  - ListAmendments");
    println!("  - GetGrammar");
    println!();
    println!("Ready to accept connections...");

    // Build reflection service for grpcurl compatibility
    let reflection_service = ReflectionBuilder::configure()
        .register_encoded_file_descriptor_set(tonic::include_file_descriptor_set!("dsl_descriptor"))
        .build_v1()?;

    Server::builder()
        .add_service(DslServiceServer::new(service))
        .add_service(reflection_service)
        .serve(addr)
        .await?;

    Ok(())
}
