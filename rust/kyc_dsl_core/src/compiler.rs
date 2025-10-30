use crate::parser::Expr;
use crate::Instruction;

/// Compile an AST into a sequence of executable instructions
pub fn compile(ast: Expr) -> Result<Vec<Instruction>, String> {
    let mut instructions = Vec::new();
    compile_expr(&ast, &mut instructions)?;
    Ok(instructions)
}

/// Recursively compile an expression into instructions
fn compile_expr(expr: &Expr, instructions: &mut Vec<Instruction>) -> Result<(), String> {
    match expr {
        Expr::Call(name, args) => {
            // Handle special forms
            match name.as_str() {
                "kyc-case" => compile_kyc_case(name, args, instructions)?,
                "nature-purpose" => compile_form(name, args, instructions)?,
                "ownership-structure" => compile_form(name, args, instructions)?,
                "data-dictionary" => compile_form(name, args, instructions)?,
                "document-requirements" => compile_form(name, args, instructions)?,
                _ => compile_form(name, args, instructions)?,
            }
        }
        Expr::Atom(_) => {
            // Atoms at top level are not compiled to instructions
            // They're typically arguments to calls
        }
    }
    Ok(())
}

/// Compile a kyc-case form
fn compile_kyc_case(
    _name: &str,
    args: &[Expr],
    instructions: &mut Vec<Instruction>,
) -> Result<(), String> {
    if args.is_empty() {
        return Err("kyc-case requires at least a name".to_string());
    }

    // Extract case name
    let case_name = match &args[0] {
        Expr::Atom(s) => s.clone(),
        _ => return Err("kyc-case name must be an atom".to_string()),
    };

    // Add case initialization instruction
    instructions.push(Instruction {
        name: "init-case".to_string(),
        args: vec![case_name.clone()],
    });

    // Compile all sub-forms
    for arg in &args[1..] {
        compile_expr(arg, instructions)?;
    }

    // Add case finalization instruction
    instructions.push(Instruction {
        name: "finalize-case".to_string(),
        args: vec![case_name],
    });

    Ok(())
}

/// Compile a generic form (function call with arguments)
fn compile_form(
    name: &str,
    args: &[Expr],
    instructions: &mut Vec<Instruction>,
) -> Result<(), String> {
    // Extract arguments as strings
    let mut arg_strings = Vec::new();
    for arg in args {
        arg_strings.push(expr_to_string(arg));
    }

    instructions.push(Instruction {
        name: name.to_string(),
        args: arg_strings,
    });

    Ok(())
}

/// Convert an expression to a string representation
fn expr_to_string(expr: &Expr) -> String {
    match expr {
        Expr::Atom(s) => s.clone(),
        Expr::Call(name, args) => {
            let args_str = args
                .iter()
                .map(expr_to_string)
                .collect::<Vec<_>>()
                .join(" ");
            if args_str.is_empty() {
                format!("({})", name)
            } else {
                format!("({} {})", name, args_str)
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_compile_simple_case() {
        let ast = Expr::Call(
            "kyc-case".to_string(),
            vec![Expr::Atom("TEST-CASE".to_string())],
        );

        let result = compile(ast);
        assert!(result.is_ok());

        let instructions = result.unwrap();
        assert_eq!(instructions.len(), 2);
        assert_eq!(instructions[0].name, "init-case");
        assert_eq!(instructions[0].args[0], "TEST-CASE");
        assert_eq!(instructions[1].name, "finalize-case");
        assert_eq!(instructions[1].args[0], "TEST-CASE");
    }

    #[test]
    fn test_compile_with_nested_forms() {
        let ast = Expr::Call(
            "kyc-case".to_string(),
            vec![
                Expr::Atom("TEST-CASE".to_string()),
                Expr::Call(
                    "nature".to_string(),
                    vec![Expr::Atom("Corporate".to_string())],
                ),
                Expr::Call(
                    "purpose".to_string(),
                    vec![Expr::Atom("Investment".to_string())],
                ),
            ],
        );

        let result = compile(ast);
        assert!(result.is_ok());

        let instructions = result.unwrap();
        assert_eq!(instructions.len(), 4);
        assert_eq!(instructions[0].name, "init-case");
        assert_eq!(instructions[1].name, "nature");
        assert_eq!(instructions[1].args[0], "Corporate");
        assert_eq!(instructions[2].name, "purpose");
        assert_eq!(instructions[2].args[0], "Investment");
        assert_eq!(instructions[3].name, "finalize-case");
    }

    #[test]
    fn test_expr_to_string() {
        let expr = Expr::Call(
            "owner".to_string(),
            vec![
                Expr::Atom("ACME-Corp".to_string()),
                Expr::Atom("45.5%".to_string()),
            ],
        );

        let result = expr_to_string(&expr);
        assert_eq!(result, "(owner ACME-Corp 45.5%)");
    }
}
