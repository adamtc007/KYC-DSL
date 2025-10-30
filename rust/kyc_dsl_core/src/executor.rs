use crate::Instruction;
use serde_json::from_str;
use std::collections::HashMap;

/// Execution context that maintains state during execution
#[derive(Debug, Default)]
pub struct ExecutionContext {
    /// Current case name being processed
    pub current_case: Option<String>,
    /// Variables and values accumulated during execution
    pub variables: HashMap<String, String>,
    /// Execution log for debugging
    pub log: Vec<String>,
}

impl ExecutionContext {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn log(&mut self, message: String) {
        self.log.push(message);
    }

    pub fn set_case(&mut self, name: String) {
        self.current_case = Some(name);
    }

    pub fn get_case(&self) -> Option<&str> {
        self.current_case.as_deref()
    }
}

/// Execute a compiled plan (JSON) and return the result
pub fn execute(plan_json: &str) -> Result<String, String> {
    let plan: Vec<Instruction> = from_str(plan_json).map_err(|e| e.to_string())?;

    let mut ctx = ExecutionContext::new();
    let mut results = Vec::new();

    for instruction in plan {
        let result = execute_instruction(&instruction, &mut ctx)?;
        results.push(result);
    }

    // Format the output
    let output = format!(
        "Execution completed successfully.\n\nResults:\n{}\n\nLog:\n{}",
        results.join("\n"),
        ctx.log.join("\n")
    );

    Ok(output)
}

/// Execute a single instruction
fn execute_instruction(
    instruction: &Instruction,
    ctx: &mut ExecutionContext,
) -> Result<String, String> {
    let result = match instruction.name.as_str() {
        "init-case" => execute_init_case(&instruction.args, ctx)?,
        "finalize-case" => execute_finalize_case(&instruction.args, ctx)?,
        "nature-purpose" => execute_nature_purpose(&instruction.args, ctx)?,
        "nature" => execute_nature(&instruction.args, ctx)?,
        "purpose" => execute_purpose(&instruction.args, ctx)?,
        "client-business-unit" => execute_cbu(&instruction.args, ctx)?,
        "policy" => execute_policy(&instruction.args, ctx)?,
        "function" => execute_function(&instruction.args, ctx)?,
        "obligation" => execute_obligation(&instruction.args, ctx)?,
        "ownership-structure" => execute_ownership(&instruction.args, ctx)?,
        "owner" => execute_owner(&instruction.args, ctx)?,
        "beneficial-owner" => execute_beneficial_owner(&instruction.args, ctx)?,
        "controller" => execute_controller(&instruction.args, ctx)?,
        "data-dictionary" => execute_data_dictionary(&instruction.args, ctx)?,
        "attribute" => execute_attribute(&instruction.args, ctx)?,
        "document-requirements" => execute_document_requirements(&instruction.args, ctx)?,
        "kyc-token" => execute_kyc_token(&instruction.args, ctx)?,
        _ => execute_generic(&instruction.name, &instruction.args, ctx)?,
    };

    Ok(result)
}

// Instruction executors

fn execute_init_case(args: &[String], ctx: &mut ExecutionContext) -> Result<String, String> {
    if args.is_empty() {
        return Err("init-case requires a case name".to_string());
    }
    let case_name = &args[0];
    ctx.set_case(case_name.clone());
    ctx.log(format!("Initialized case: {}", case_name));
    Ok(format!("✓ Case '{}' initialized", case_name))
}

fn execute_finalize_case(args: &[String], ctx: &mut ExecutionContext) -> Result<String, String> {
    if args.is_empty() {
        return Err("finalize-case requires a case name".to_string());
    }
    let case_name = &args[0];
    ctx.log(format!("Finalized case: {}", case_name));
    Ok(format!("✓ Case '{}' finalized", case_name))
}

fn execute_nature_purpose(args: &[String], ctx: &mut ExecutionContext) -> Result<String, String> {
    ctx.log("Processing nature-purpose section".to_string());
    Ok(format!(
        "✓ Nature-purpose defined with {} elements",
        args.len()
    ))
}

fn execute_nature(args: &[String], ctx: &mut ExecutionContext) -> Result<String, String> {
    if args.is_empty() {
        return Err("nature requires a value".to_string());
    }
    let nature = &args[0];
    ctx.variables.insert("nature".to_string(), nature.clone());
    ctx.log(format!("Set nature: {}", nature));
    Ok(format!("✓ Nature: {}", nature))
}

fn execute_purpose(args: &[String], ctx: &mut ExecutionContext) -> Result<String, String> {
    if args.is_empty() {
        return Err("purpose requires a value".to_string());
    }
    let purpose = &args[0];
    ctx.variables.insert("purpose".to_string(), purpose.clone());
    ctx.log(format!("Set purpose: {}", purpose));
    Ok(format!("✓ Purpose: {}", purpose))
}

fn execute_cbu(args: &[String], ctx: &mut ExecutionContext) -> Result<String, String> {
    if args.is_empty() {
        return Err("client-business-unit requires a value".to_string());
    }
    let cbu = &args[0];
    ctx.variables.insert("cbu".to_string(), cbu.clone());
    ctx.log(format!("Set CBU: {}", cbu));
    Ok(format!("✓ Client Business Unit: {}", cbu))
}

fn execute_policy(args: &[String], ctx: &mut ExecutionContext) -> Result<String, String> {
    if args.is_empty() {
        return Err("policy requires a value".to_string());
    }
    let policy = &args[0];
    ctx.variables.insert("policy".to_string(), policy.clone());
    ctx.log(format!("Set policy: {}", policy));
    Ok(format!("✓ Policy: {}", policy))
}

fn execute_function(args: &[String], ctx: &mut ExecutionContext) -> Result<String, String> {
    if args.is_empty() {
        return Err("function requires a value".to_string());
    }
    let function = &args[0];
    ctx.variables
        .insert("function".to_string(), function.clone());
    ctx.log(format!("Set function: {}", function));
    Ok(format!("✓ Function: {}", function))
}

fn execute_obligation(args: &[String], ctx: &mut ExecutionContext) -> Result<String, String> {
    if args.is_empty() {
        return Err("obligation requires a value".to_string());
    }
    let obligation = &args[0];
    ctx.variables
        .insert("obligation".to_string(), obligation.clone());
    ctx.log(format!("Set obligation: {}", obligation));
    Ok(format!("✓ Obligation: {}", obligation))
}

fn execute_ownership(args: &[String], ctx: &mut ExecutionContext) -> Result<String, String> {
    ctx.log("Processing ownership structure".to_string());
    Ok(format!(
        "✓ Ownership structure with {} elements",
        args.len()
    ))
}

fn execute_owner(args: &[String], ctx: &mut ExecutionContext) -> Result<String, String> {
    if args.len() < 2 {
        return Err("owner requires name and percentage".to_string());
    }
    let name = &args[0];
    let percentage = &args[1];
    ctx.log(format!("Added owner: {} ({})", name, percentage));
    Ok(format!("✓ Owner: {} - {}", name, percentage))
}

fn execute_beneficial_owner(args: &[String], ctx: &mut ExecutionContext) -> Result<String, String> {
    if args.len() < 2 {
        return Err("beneficial-owner requires name and percentage".to_string());
    }
    let name = &args[0];
    let percentage = &args[1];
    ctx.log(format!("Added beneficial owner: {} ({})", name, percentage));
    Ok(format!("✓ Beneficial Owner: {} - {}", name, percentage))
}

fn execute_controller(args: &[String], ctx: &mut ExecutionContext) -> Result<String, String> {
    if args.len() < 2 {
        return Err("controller requires name and role".to_string());
    }
    let name = &args[0];
    let role = &args[1];
    ctx.log(format!("Added controller: {} ({})", name, role));
    Ok(format!("✓ Controller: {} - {}", name, role))
}

fn execute_data_dictionary(args: &[String], ctx: &mut ExecutionContext) -> Result<String, String> {
    ctx.log("Processing data dictionary".to_string());
    Ok(format!("✓ Data dictionary with {} entries", args.len()))
}

fn execute_attribute(args: &[String], ctx: &mut ExecutionContext) -> Result<String, String> {
    if args.is_empty() {
        return Err("attribute requires a code".to_string());
    }
    let attr_code = &args[0];
    ctx.log(format!("Defined attribute: {}", attr_code));
    Ok(format!("✓ Attribute: {}", attr_code))
}

fn execute_document_requirements(
    args: &[String],
    ctx: &mut ExecutionContext,
) -> Result<String, String> {
    ctx.log("Processing document requirements".to_string());
    Ok(format!(
        "✓ Document requirements with {} elements",
        args.len()
    ))
}

fn execute_kyc_token(args: &[String], ctx: &mut ExecutionContext) -> Result<String, String> {
    if args.is_empty() {
        return Err("kyc-token requires a status".to_string());
    }
    let status = &args[0];
    ctx.variables
        .insert("kyc_token".to_string(), status.clone());
    ctx.log(format!("Set KYC token: {}", status));
    Ok(format!("✓ KYC Token: {}", status))
}

fn execute_generic(
    name: &str,
    args: &[String],
    ctx: &mut ExecutionContext,
) -> Result<String, String> {
    ctx.log(format!("Executed generic instruction: {}", name));
    Ok(format!("✓ {}: {} args", name, args.len()))
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::Instruction;

    #[test]
    fn test_execute_simple_plan() {
        let instructions = vec![
            Instruction {
                name: "init-case".to_string(),
                args: vec!["TEST-CASE".to_string()],
            },
            Instruction {
                name: "nature".to_string(),
                args: vec!["Corporate".to_string()],
            },
            Instruction {
                name: "finalize-case".to_string(),
                args: vec!["TEST-CASE".to_string()],
            },
        ];

        let plan_json = serde_json::to_string(&instructions).unwrap();
        let result = execute(&plan_json);

        assert!(result.is_ok());
        let output = result.unwrap();
        assert!(output.contains("TEST-CASE"));
        assert!(output.contains("Corporate"));
    }

    #[test]
    fn test_execution_context() {
        let mut ctx = ExecutionContext::new();

        ctx.set_case("MY-CASE".to_string());
        assert_eq!(ctx.get_case(), Some("MY-CASE"));

        ctx.log("Test log entry".to_string());
        assert_eq!(ctx.log.len(), 1);

        ctx.variables.insert("key".to_string(), "value".to_string());
        assert_eq!(ctx.variables.get("key"), Some(&"value".to_string()));
    }

    #[test]
    fn test_execute_init_case() {
        let mut ctx = ExecutionContext::new();
        let args = vec!["TEST-CASE".to_string()];

        let result = execute_init_case(&args, &mut ctx);
        assert!(result.is_ok());
        assert_eq!(ctx.get_case(), Some("TEST-CASE"));
    }

    #[test]
    fn test_execute_nature() {
        let mut ctx = ExecutionContext::new();
        let args = vec!["Corporate".to_string()];

        let result = execute_nature(&args, &mut ctx);
        assert!(result.is_ok());
        assert_eq!(ctx.variables.get("nature"), Some(&"Corporate".to_string()));
    }

    #[test]
    fn test_execute_owner() {
        let mut ctx = ExecutionContext::new();
        let args = vec!["ACME-Corp".to_string(), "45.5%".to_string()];

        let result = execute_owner(&args, &mut ctx);
        assert!(result.is_ok());
        assert!(result.unwrap().contains("ACME-Corp"));
    }

    #[test]
    fn test_invalid_json() {
        let result = execute("invalid json");
        assert!(result.is_err());
    }

    #[test]
    fn test_missing_args() {
        let mut ctx = ExecutionContext::new();
        let args = vec![];

        let result = execute_init_case(&args, &mut ctx);
        assert!(result.is_err());
    }
}
