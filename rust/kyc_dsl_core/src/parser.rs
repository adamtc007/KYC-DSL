use nom::{
    branch::alt,
    bytes::complete::take_while1,
    character::complete::{char, multispace0},
    combinator::map,
    multi::many0,
    sequence::{delimited, preceded, tuple},
    IResult,
};

/// Expression types in the KYC DSL
#[derive(Debug, Clone, PartialEq)]
pub enum Expr {
    /// Function call: (name arg1 arg2 ...)
    Call(String, Vec<Expr>),
    /// Atomic value: identifier, string, or number
    Atom(String),
}

/// Parse an atomic value (identifier, keyword, or literal)
fn atom(input: &str) -> IResult<&str, Expr> {
    map(
        take_while1(|c: char| c.is_alphanumeric() || "_-%.".contains(c)),
        |s: &str| Expr::Atom(s.to_string()),
    )(input)
}

/// Parse a quoted string
fn quoted_string(input: &str) -> IResult<&str, Expr> {
    map(
        delimited(char('"'), take_while1(|c: char| c != '"'), char('"')),
        |s: &str| Expr::Atom(s.to_string()),
    )(input)
}

/// Parse either an atom or a quoted string
fn atom_or_string(input: &str) -> IResult<&str, Expr> {
    alt((quoted_string, atom))(input)
}

/// Parse an S-expression recursively
fn expr(input: &str) -> IResult<&str, Expr> {
    alt((
        // S-expression: (name args...)
        delimited(
            tuple((char('('), multispace0)),
            map(
                tuple((atom_or_string, many0(preceded(multispace0, expr)))),
                |(f, args)| {
                    if let Expr::Atom(name) = f {
                        Expr::Call(name, args)
                    } else {
                        f
                    }
                },
            ),
            tuple((multispace0, char(')'))),
        ),
        // Simple atom or string
        atom_or_string,
    ))(input)
}

/// Parse a complete DSL source file
pub fn parse(src: &str) -> Result<Expr, nom::Err<nom::error::Error<&str>>> {
    let trimmed = src.trim();
    let (_, res) = expr(trimmed)?;
    Ok(res)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_atom() {
        let result = parse("kyc-case");
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), Expr::Atom("kyc-case".to_string()));
    }

    #[test]
    fn test_parse_simple_call() {
        let result = parse("(kyc-case TEST)");
        assert!(result.is_ok());
        match result.unwrap() {
            Expr::Call(name, args) => {
                assert_eq!(name, "kyc-case");
                assert_eq!(args.len(), 1);
            }
            _ => panic!("Expected Call"),
        }
    }

    #[test]
    fn test_parse_nested() {
        let result = parse("(kyc-case TEST (nature \"Corporate\"))");
        assert!(result.is_ok());
        match result.unwrap() {
            Expr::Call(name, args) => {
                assert_eq!(name, "kyc-case");
                assert_eq!(args.len(), 2);
            }
            _ => panic!("Expected Call"),
        }
    }

    #[test]
    fn test_parse_quoted_string() {
        let result = parse("\"Hello World\"");
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), Expr::Atom("Hello World".to_string()));
    }
}
