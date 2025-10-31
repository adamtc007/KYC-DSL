use clap::{Parser, Subcommand};
use colored::*;
use tonic::Request;

// Include generated protobuf code
pub mod kyc {
    pub mod ontology {
        tonic::include_proto!("kyc.ontology");
    }
    pub mod data {
        tonic::include_proto!("kyc.data");
    }
}

use kyc::ontology::ontology_service_client::OntologyServiceClient;
use kyc::data::dictionary_service_client::DictionaryServiceClient;

#[derive(Parser)]
#[command(name = "ontology_cli")]
#[command(about = "KYC Ontology gRPC Client", long_about = None)]
struct Cli {
    #[arg(short, long, default_value = "http://localhost:50070")]
    server: String,

    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// Search attributes by keyword
    SearchAttributes {
        #[arg(help = "Search query (e.g., 'ownership', 'tax')")]
        query: String,
        #[arg(short, long, default_value = "10")]
        limit: i32,
    },
    /// List all attributes
    ListAttributes {
        #[arg(short, long, default_value = "20")]
        limit: i32,
        #[arg(short, long, default_value = "0")]
        offset: i32,
    },
    /// Get a specific attribute by ID
    GetAttribute {
        #[arg(help = "Attribute ID (UUID)")]
        id: String,
    },
    /// List entities
    ListEntities {
        #[arg(short, long, default_value = "10")]
        limit: i32,
        #[arg(short, long, default_value = "0")]
        offset: i32,
    },
    /// Search entities
    SearchEntities {
        #[arg(help = "Search query")]
        query: String,
        #[arg(short, long, default_value = "10")]
        limit: i32,
    },
    /// Get entity by ID
    GetEntity {
        #[arg(help = "Entity ID (UUID)")]
        id: String,
    },
    /// Search concepts
    SearchConcepts {
        #[arg(help = "Search query")]
        query: String,
        #[arg(short, long, default_value = "10")]
        limit: i32,
    },
    /// List regulations
    ListRegulations {
        #[arg(short, long, default_value = "20")]
        limit: i32,
    },
    /// List documents
    ListDocuments {
        #[arg(short, long, default_value = "20")]
        limit: i32,
    },
    /// Get CBU by ID
    GetCbu {
        #[arg(help = "CBU ID (UUID)")]
        id: String,
    },
    /// Test connection
    Ping,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let cli = Cli::parse();

    println!(
        "{} {}",
        "🔗 Connecting to OntologyService at".cyan(),
        cli.server.green()
    );

    match cli.command {
        Commands::SearchAttributes { query, limit } => {
            search_attributes(&cli.server, &query, limit).await?;
        }
        Commands::ListAttributes { limit, offset } => {
            list_attributes(&cli.server, limit, offset).await?;
        }
        Commands::GetAttribute { id } => {
            get_attribute(&cli.server, &id).await?;
        }
        Commands::ListEntities { limit, offset } => {
            list_entities(&cli.server, limit, offset).await?;
        }
        Commands::SearchEntities { query, limit } => {
            search_entities(&cli.server, &query, limit).await?;
        }
        Commands::GetEntity { id } => {
            get_entity(&cli.server, &id).await?;
        }
        Commands::SearchConcepts { query, limit } => {
            search_concepts(&cli.server, &query, limit).await?;
        }
        Commands::ListRegulations { limit } => {
            list_regulations(&cli.server, limit).await?;
        }
        Commands::ListDocuments { limit } => {
            list_documents(&cli.server, limit).await?;
        }
        Commands::GetCbu { id } => {
            get_cbu(&cli.server, &id).await?;
        }
        Commands::Ping => {
            ping(&cli.server).await?;
        }
    }

    Ok(())
}

async fn search_attributes(
    server: &str,
    query: &str,
    limit: i32,
) -> Result<(), Box<dyn std::error::Error>> {
    let mut client = OntologyServiceClient::connect(server.to_string()).await?;

    println!(
        "{} '{}' {}",
        "🔍 Searching attributes for".yellow(),
        query.bright_white(),
        format!("(limit: {})", limit).dimmed()
    );

    let request = Request::new(kyc::ontology::SearchRequest {
        query: query.to_string(),
        limit,
        offset: 0,
        domain: String::new(),
        similarity_threshold: 0.0,
    });

    let response = client.search_attributes(request).await?;
    let attribute_list = response.into_inner();

    println!(
        "\n{} {} {}",
        "✅ Found".green(),
        attribute_list.attributes.len().to_string().bright_white(),
        format!("attributes (total: {})", attribute_list.total_count).dimmed()
    );

    for attr in attribute_list.attributes {
        println!("\n  {} {}", "▶".blue(), attr.code.bright_cyan());
        println!("    {} {}", "Name:".dimmed(), attr.name);
        println!("    {} {}", "Type:".dimmed(), attr.attr_type);
        if !attr.jurisdiction.is_empty() {
            println!("    {} {}", "Jurisdiction:".dimmed(), attr.jurisdiction);
        }
        if !attr.sink_table.is_empty() && !attr.sink_column.is_empty() {
            println!(
                "    {} {}.{}",
                "Target:".dimmed(),
                attr.sink_table.yellow(),
                attr.sink_column.yellow()
            );
        }
        if !attr.description.is_empty() {
            println!("    {} {}", "Description:".dimmed(), attr.description);
        }
    }

    Ok(())
}

async fn list_attributes(
    server: &str,
    limit: i32,
    offset: i32,
) -> Result<(), Box<dyn std::error::Error>> {
    let mut client = OntologyServiceClient::connect(server.to_string()).await?;

    println!(
        "{} {}",
        "📖 Listing attributes".yellow(),
        format!("(limit: {}, offset: {})", limit, offset).dimmed()
    );

    let request = Request::new(kyc::ontology::ListAttributesRequest {
        limit,
        offset,
        jurisdiction: String::new(),
        attr_type: String::new(),
        is_required: false,
    });

    let response = client.list_attributes(request).await?;
    let attribute_list = response.into_inner();

    println!(
        "\n{} {} {}",
        "✅ Listed".green(),
        attribute_list.attributes.len().to_string().bright_white(),
        format!("attributes (total: {})", attribute_list.total_count).dimmed()
    );

    for attr in attribute_list.attributes {
        println!(
            "  {} {} → {}.{}",
            attr.code.bright_cyan(),
            format!("[{}]", attr.jurisdiction).dimmed(),
            attr.sink_table.yellow(),
            attr.sink_column.yellow()
        );
    }

    Ok(())
}

async fn get_attribute(server: &str, id: &str) -> Result<(), Box<dyn std::error::Error>> {
    let mut client = OntologyServiceClient::connect(server.to_string()).await?;

    println!("{} {}", "📖 Getting attribute".yellow(), id.bright_white());

    let request = Request::new(kyc::ontology::GetAttributeRequest {
        id: id.to_string(),
    });

    let response = client.get_attribute(request).await?;
    let attr = response.into_inner();

    println!("\n{}", "✅ Attribute Details:".green());
    println!("  {} {}", "Code:".dimmed(), attr.code.bright_cyan());
    println!("  {} {}", "Name:".dimmed(), attr.name);
    println!("  {} {}", "Type:".dimmed(), attr.attr_type);
    println!("  {} {}", "Jurisdiction:".dimmed(), attr.jurisdiction);
    println!("  {} {}", "Required:".dimmed(), attr.is_required);
    println!("  {} {}", "PII:".dimmed(), attr.is_pii);
    if !attr.sink_table.is_empty() {
        println!(
            "  {} {}.{}",
            "Target:".dimmed(),
            attr.sink_table.yellow(),
            attr.sink_column.yellow()
        );
    }
    println!("  {} {}", "Description:".dimmed(), attr.description);

    Ok(())
}

async fn list_entities(
    server: &str,
    limit: i32,
    offset: i32,
) -> Result<(), Box<dyn std::error::Error>> {
    let mut client = OntologyServiceClient::connect(server.to_string()).await?;

    println!(
        "{} {}",
        "📦 Listing entities".yellow(),
        format!("(limit: {}, offset: {})", limit, offset).dimmed()
    );

    let request = Request::new(kyc::ontology::ListEntitiesRequest {
        limit,
        offset,
        entity_type: String::new(),
        jurisdiction: String::new(),
        status: String::new(),
    });

    let response = client.list_entities(request).await?;
    let entity_list = response.into_inner();

    println!(
        "\n{} {} {}",
        "✅ Listed".green(),
        entity_list.entities.len().to_string().bright_white(),
        format!("entities (total: {})", entity_list.total_count).dimmed()
    );

    for entity in entity_list.entities {
        println!(
            "  {} {} [{}] - {}",
            entity.name.bright_cyan(),
            format!("({})", entity.entity_type).dimmed(),
            entity.jurisdiction,
            entity.status.green()
        );
        if !entity.lei_code.is_empty() {
            println!("    {} {}", "LEI:".dimmed(), entity.lei_code.yellow());
        }
    }

    Ok(())
}

async fn search_entities(
    server: &str,
    query: &str,
    limit: i32,
) -> Result<(), Box<dyn std::error::Error>> {
    let mut client = OntologyServiceClient::connect(server.to_string()).await?;

    println!(
        "{} '{}'",
        "🔍 Searching entities for".yellow(),
        query.bright_white()
    );

    let request = Request::new(kyc::ontology::SearchRequest {
        query: query.to_string(),
        limit,
        offset: 0,
        domain: String::new(),
        similarity_threshold: 0.0,
    });

    let response = client.search_entities(request).await?;
    let entity_list = response.into_inner();

    println!(
        "\n{} {} entities",
        "✅ Found".green(),
        entity_list.entities.len().to_string().bright_white()
    );

    for entity in entity_list.entities {
        println!("  {} {}", "▶".blue(), entity.name.bright_cyan());
        println!("    {} {}", "Type:".dimmed(), entity.entity_type);
        println!("    {} {}", "Jurisdiction:".dimmed(), entity.jurisdiction);
        if !entity.description.is_empty() {
            println!("    {} {}", "Description:".dimmed(), entity.description);
        }
    }

    Ok(())
}

async fn get_entity(server: &str, id: &str) -> Result<(), Box<dyn std::error::Error>> {
    let mut client = OntologyServiceClient::connect(server.to_string()).await?;

    println!("{} {}", "📦 Getting entity".yellow(), id.bright_white());

    let request = Request::new(kyc::ontology::GetEntityRequest {
        id: id.to_string(),
    });

    let response = client.get_entity(request).await?;
    let entity = response.into_inner();

    println!("\n{}", "✅ Entity Details:".green());
    println!("  {} {}", "Name:".dimmed(), entity.name.bright_cyan());
    println!("  {} {}", "Type:".dimmed(), entity.entity_type);
    println!("  {} {}", "Jurisdiction:".dimmed(), entity.jurisdiction);
    println!("  {} {}", "Status:".dimmed(), entity.status.green());
    if !entity.legal_form.is_empty() {
        println!("  {} {}", "Legal Form:".dimmed(), entity.legal_form);
    }
    if !entity.lei_code.is_empty() {
        println!("  {} {}", "LEI:".dimmed(), entity.lei_code.yellow());
    }
    if !entity.registration_number.is_empty() {
        println!(
            "  {} {}",
            "Registration:".dimmed(),
            entity.registration_number
        );
    }
    if !entity.description.is_empty() {
        println!("  {} {}", "Description:".dimmed(), entity.description);
    }

    Ok(())
}

async fn search_concepts(
    server: &str,
    query: &str,
    limit: i32,
) -> Result<(), Box<dyn std::error::Error>> {
    let mut client = OntologyServiceClient::connect(server.to_string()).await?;

    println!(
        "{} '{}'",
        "💡 Searching concepts for".yellow(),
        query.bright_white()
    );

    let request = Request::new(kyc::ontology::SearchRequest {
        query: query.to_string(),
        limit,
        offset: 0,
        domain: String::new(),
        similarity_threshold: 0.0,
    });

    let response = client.search_concepts(request).await?;
    let concept_list = response.into_inner();

    println!(
        "\n{} {} concepts",
        "✅ Found".green(),
        concept_list.concepts.len().to_string().bright_white()
    );

    for concept in concept_list.concepts {
        println!("  {} {}", "▶".blue(), concept.name.bright_cyan());
        println!("    {} {}", "Code:".dimmed(), concept.code.yellow());
        println!("    {} {}", "Domain:".dimmed(), concept.domain);
        if !concept.synonyms.is_empty() {
            println!(
                "    {} {}",
                "Synonyms:".dimmed(),
                concept.synonyms.join(", ")
            );
        }
        if !concept.description.is_empty() {
            println!("    {} {}", "Description:".dimmed(), concept.description);
        }
    }

    Ok(())
}

async fn list_regulations(server: &str, limit: i32) -> Result<(), Box<dyn std::error::Error>> {
    let mut client = OntologyServiceClient::connect(server.to_string()).await?;

    println!("{} {}", "📜 Listing regulations".yellow(), format!("(limit: {})", limit).dimmed());

    let request = Request::new(kyc::ontology::ListRegulationsRequest {
        limit,
        offset: 0,
        jurisdiction: String::new(),
        status: String::new(),
    });

    let response = client.list_regulations(request).await?;
    let regulation_list = response.into_inner();

    println!(
        "\n{} {} {}",
        "✅ Listed".green(),
        regulation_list.regulations.len().to_string().bright_white(),
        format!("regulations (total: {})", regulation_list.total_count).dimmed()
    );

    for reg in regulation_list.regulations {
        println!(
            "  {} {} [{}]",
            reg.code.bright_cyan(),
            reg.name,
            reg.jurisdiction.yellow()
        );
        if !reg.authority.is_empty() {
            println!("    {} {}", "Authority:".dimmed(), reg.authority);
        }
    }

    Ok(())
}

async fn list_documents(server: &str, limit: i32) -> Result<(), Box<dyn std::error::Error>> {
    let mut client = OntologyServiceClient::connect(server.to_string()).await?;

    println!("{} {}", "📄 Listing documents".yellow(), format!("(limit: {})", limit).dimmed());

    let request = Request::new(kyc::ontology::ListDocumentsRequest {
        limit,
        offset: 0,
        jurisdiction: String::new(),
        category: String::new(),
        is_mandatory: false,
    });

    let response = client.list_documents(request).await?;
    let document_list = response.into_inner();

    println!(
        "\n{} {} {}",
        "✅ Listed".green(),
        document_list.documents.len().to_string().bright_white(),
        format!("documents (total: {})", document_list.total_count).dimmed()
    );

    for doc in document_list.documents {
        println!(
            "  {} {} [{}]",
            doc.code.bright_cyan(),
            doc.title,
            doc.category.yellow()
        );
        if !doc.jurisdiction.is_empty() {
            println!("    {} {}", "Jurisdiction:".dimmed(), doc.jurisdiction);
        }
    }

    Ok(())
}

async fn get_cbu(server: &str, id: &str) -> Result<(), Box<dyn std::error::Error>> {
    let mut client = OntologyServiceClient::connect(server.to_string()).await?;

    println!("{} {}", "🏢 Getting CBU".yellow(), id.bright_white());

    let request = Request::new(kyc::ontology::GetCbuRequest {
        id: id.to_string(),
    });

    let response = client.get_cbu(request).await?;
    let cbu = response.into_inner();

    println!("\n{}", "✅ CBU Details:".green());
    println!("  {} {}", "Name:".dimmed(), cbu.name.bright_cyan());
    println!("  {} {}", "Code:".dimmed(), cbu.code.yellow());
    println!("  {} {}", "Domicile:".dimmed(), cbu.domicile);
    if !cbu.sponsor_entity_id.is_empty() {
        println!("  {} {}", "Sponsor Entity:".dimmed(), cbu.sponsor_entity_id);
    }
    if !cbu.description.is_empty() {
        println!("  {} {}", "Description:".dimmed(), cbu.description);
    }

    Ok(())
}

async fn ping(server: &str) -> Result<(), Box<dyn std::error::Error>> {
    let mut client = OntologyServiceClient::connect(server.to_string()).await?;

    println!("{} {}", "🏓 Testing connection to".yellow(), server.green());

    // Try a simple list operation
    let request = Request::new(kyc::ontology::ListRegulationsRequest {
        limit: 1,
        offset: 0,
        jurisdiction: String::new(),
        status: String::new(),
    });

    match client.list_regulations(request).await {
        Ok(_) => {
            println!("\n{} {}", "✅ Connection successful!".green(), "🎉".bright_white());
            println!("   {} is {}", server.cyan(), "UP".green());
        }
        Err(e) => {
            println!("\n{} {}", "❌ Connection failed!".red(), e);
        }
    }

    Ok(())
}
