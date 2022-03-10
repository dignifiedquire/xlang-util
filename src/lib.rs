pub mod channel;
pub use channel::*;

#[no_mangle]
pub extern "C" fn new_channel(cap: u32) -> *mut Channel {
    Box::into_raw(Box::new(Channel::with_capacity(cap)))
}

#[no_mangle]
pub extern "C" fn drop_channel(channel_ptr: *mut Channel) {
    assert!(!channel_ptr.is_null());
    let channel = unsafe { Box::from_raw(channel_ptr) };
    drop(channel)
}

#[no_mangle]
pub extern "C" fn new_message_bytes(ptr: *const u8, len: u64) -> *mut u8 {
    assert!(!ptr.is_null());

    let bytes = unsafe { core::slice::from_raw_parts(ptr, usize::try_from(len).unwrap()) };

    Box::into_raw(bytes.to_vec().into_boxed_slice()) as *mut _
}

#[no_mangle]
pub extern "C" fn drop_message(msg: *mut Message) {
    assert!(!msg.is_null());
    let _msg = unsafe { Box::from_raw(msg) };
}
