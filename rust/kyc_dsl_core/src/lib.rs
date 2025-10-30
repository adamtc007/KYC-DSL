pub mod compiler;
pub mod executor;
pub mod parser;

use serde::{Deserialize, Serialize};
use thiserror::Error;

#[derive(Debug, Error)]
pub enum DslError {
    #[error("Parse error: {0}")]
    Parse(String),
    #[error("Compile error: {0}")]
    Compile(String),
    #[error("Execution error: {0}")]
    Exec(String),
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Instruction {
    pub name: String,
    pub args: Vec<String>,
}

/// Compile DSL source code into an execution plan (JSON)
pub fn compile_dsl(src: &str) -> Result<String, DslError> {
    let ast = parser::parse(src).map_err(|e| DslError::Parse(e.to_string()))?;
    let plan = compiler::compile(ast).map_err(|e| DslError::Compile(e.to_string()))?;
    Ok(serde_json::to_string(&plan).unwrap())
}

/// Execute a compiled plan (JSON) and return the result
pub fn execute_plan(plan_json: &str) -> Result<String, DslError> {
    executor::execute(plan_json).map_err(|e| DslError::Exec(e.to_string()))
}
