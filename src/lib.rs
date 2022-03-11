pub mod channel;
pub use channel::*;

#[no_mangle]
pub extern "C" fn new_channel(cap: u32) -> *mut Channel {
    Box::into_raw(Box::new(Channel::with_capacity(cap)))
}

#[no_mangle]
pub extern "C" fn drop_channel(channel_ptr: *mut Channel) {
    assert!(!channel_ptr.is_null());
    let _channel = unsafe { Box::from_raw(channel_ptr) };
}

#[no_mangle]
pub extern "C" fn channel_try_recv(channel_ptr: *mut Channel) -> *mut Message {
    assert!(!channel_ptr.is_null());
    let channel = unsafe { Box::from_raw(channel_ptr) };
    match channel.try_recv() {
        Ok(msg) => Box::into_raw(Box::new(dbg!(msg))),
        Err(_) => core::ptr::null_mut(),
    }
}

#[no_mangle]
pub extern "C" fn new_message_bytes(ptr: *const u8, len: u64) -> *mut u8 {
    assert!(!ptr.is_null());

    let bytes = unsafe { core::slice::from_raw_parts(ptr, usize::try_from(len).unwrap()) };

    Box::into_raw(bytes.to_vec().into_boxed_slice()).cast()
}

#[no_mangle]
pub extern "C" fn new_message(ptr: *const u8, len: u64) -> *mut Message {
    assert!(!ptr.is_null());

    let bytes = unsafe { core::slice::from_raw_parts(ptr, usize::try_from(len).unwrap()) };

    Box::into_raw(Box::new(Message::from_bytes(bytes))).cast()
}

#[no_mangle]
pub extern "C" fn drop_message_bytes(ptr: *mut u8, len: u64) {
    assert!(!ptr.is_null());
    let bytes = unsafe { core::slice::from_raw_parts_mut(ptr, usize::try_from(len).unwrap()) };
    let _v: Box<[u8]> = unsafe { Box::from_raw(bytes) };
}

#[no_mangle]
pub extern "C" fn drop_message(msg: *mut Message) {
    assert!(!msg.is_null());
    let _msg = unsafe { Box::from_raw(msg) };
}

#[no_mangle]
pub extern "C" fn slotSize() -> usize {
    core::mem::size_of::<Slot>()
}
