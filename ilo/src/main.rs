use uuid::Uuid;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let root_uuid = Uuid::new_v5(&Uuid::NAMESPACE_DNS, b"zat.is");
    print!("{}", root_uuid);
    Ok(())
}
