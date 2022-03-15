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
pub extern "C" fn channel_send(channel_ptr: *mut Channel, ptr: *mut u8, len: u64) -> *mut u8 {
    assert!(!channel_ptr.is_null());
    assert!(!ptr.is_null());

    let channel: &Channel = unsafe { &*channel_ptr };

    match channel.send((ptr, len)) {
        Ok(_) => core::ptr::null_mut(),
        Err(SendError::Disconnected((nptr, nlen))) => {
            assert_eq!(nlen, len);
            nptr
        }
    }
}

#[no_mangle]
pub extern "C" fn channel_recv(channel_ptr: *mut Channel, out_len: &mut u64) -> *mut u8 {
    assert!(!channel_ptr.is_null());
    let channel: &Channel = unsafe { &*channel_ptr };

    match channel.recv() {
        Ok((ptr, len)) => {
            *out_len = len;
            ptr
        }
        Err(_) => core::ptr::null_mut(),
    }
}

#[no_mangle]
pub extern "C" fn channel_try_recv(channel_ptr: *mut Channel, out_len: &mut u64) -> *mut u8 {
    assert!(!channel_ptr.is_null());
    let channel: &Channel = unsafe { &*channel_ptr };

    match channel.try_recv() {
        Ok((ptr, len)) => {
            *out_len = len;
            ptr
        }
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
pub extern "C" fn drop_message_bytes(ptr: *mut u8, len: u64) {
    assert!(!ptr.is_null());
    let bytes = unsafe { core::slice::from_raw_parts_mut(ptr, usize::try_from(len).unwrap()) };
    let _v: Box<[u8]> = unsafe { Box::from_raw(bytes) };
}

#[no_mangle]
pub extern "C" fn slotSize() -> usize {
    core::mem::size_of::<Slot>()
}
