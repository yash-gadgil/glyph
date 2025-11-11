fn main() -> Result<(), Box<dyn std::error::Error>> {
    use std::path::Path;

    let manifest_dir = Path::new(env!("CARGO_MANIFEST_DIR"));
    let repo_root = manifest_dir
        .parent()
        .and_then(|p| p.parent())
        .expect("crate must be at services/<crate>");

    let proto_dir = repo_root.join("proto");
    let proto_file = proto_dir.join("order_book").join("order_book.proto");

    tonic_build::configure()
        .build_client(false)
        .build_server(true)
        .compile_protos(&[proto_file.as_path()], &[proto_dir.as_path()])?;

    Ok(())
}
