use uuid::Uuid;
use zerocopy::{
    AsBytes, FromBytes, LittleEndian, U16, U32, U64, Unaligned, LayoutVerified, BigEndian,
};

#[derive(FromBytes, AsBytes, Unaligned)]
#[repr(C)]
struct UKey {
    a: U64<BigEndian>,
    b: U64<BigEndian>,
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let root_uuid = Uuid::new_v5(&Uuid::NAMESPACE_DNS, b"zat.is");
    print!("{}", root_uuid);
    Ok(())
}
