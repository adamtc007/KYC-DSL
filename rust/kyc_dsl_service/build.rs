fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Compile the shared protobuf definition from the Go API directory
    let out_dir = std::path::PathBuf::from(std::env::var("OUT_DIR")?);

    tonic_build::configure()
        .build_server(true)
        .build_client(false)
        .file_descriptor_set_path(out_dir.join("dsl_descriptor.bin"))
        .extern_path(".google.protobuf.Timestamp", "::prost_types::Timestamp")
        .compile_protos(&["../../api/proto/dsl_service.proto"], &["../../api/proto"])?;

    println!("cargo:rerun-if-changed=../../api/proto/dsl_service.proto");

    Ok(())
}
