fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Compile the shared protobuf definition from the Go API directory
    tonic_build::configure()
        .build_server(true)
        .build_client(false)
        .extern_path(".google.protobuf.Timestamp", "::prost_types::Timestamp")
        .compile_protos(&["../../api/proto/dsl_service.proto"], &["../../api/proto"])?;

    println!("cargo:rerun-if-changed=../../api/proto/dsl_service.proto");

    Ok(())
}
