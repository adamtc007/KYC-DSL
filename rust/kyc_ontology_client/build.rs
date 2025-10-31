fn main() -> Result<(), Box<dyn std::error::Error>> {
    let proto_dir = "../../proto_shared";

    // Compile ontology_service.proto for client usage
    tonic_build::configure()
        .build_server(false)
        .build_client(true)
        .compile(
            &[
                &format!("{}/ontology_service.proto", proto_dir),
                &format!("{}/data_service.proto", proto_dir),
            ],
            &[proto_dir],
        )?;

    // Rerun build if proto files change
    println!(
        "cargo:rerun-if-changed={}/ontology_service.proto",
        proto_dir
    );
    println!("cargo:rerun-if-changed={}/data_service.proto", proto_dir);

    Ok(())
}
