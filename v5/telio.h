

// This file was autogenerated by some hot garbage in the `uniffi` crate.
// Trust me, you don't want to mess with it!



#include <stdbool.h>
#include <stdint.h>

// The following structs are used to implement the lowest level
// of the FFI, and thus useful to multiple uniffied crates.
// We ensure they are declared exactly once, with a header guard, UNIFFI_SHARED_H.
#ifdef UNIFFI_SHARED_H
	// We also try to prevent mixing versions of shared uniffi header structs.
	// If you add anything to the #else block, you must increment the version suffix in UNIFFI_SHARED_HEADER_V6
	#ifndef UNIFFI_SHARED_HEADER_V6
		#error Combining helper code from multiple versions of uniffi is not supported
	#endif // ndef UNIFFI_SHARED_HEADER_V6
#else
#define UNIFFI_SHARED_H
#define UNIFFI_SHARED_HEADER_V6
// ⚠️ Attention: If you change this #else block (ending in `#endif // def UNIFFI_SHARED_H`) you *must* ⚠️
// ⚠️ increment the version suffix in all instances of UNIFFI_SHARED_HEADER_V6 in this file.           ⚠️

typedef struct RustBuffer {
	int32_t capacity;
	int32_t len;
	uint8_t *data;
} RustBuffer;

typedef int32_t (*ForeignCallback)(uint64_t, int32_t, uint8_t *, int32_t, RustBuffer *);

// Task defined in Rust that Go executes
typedef void (*RustTaskCallback)(const void *, int8_t);

// Callback to execute Rust tasks using a Go routine
//
// Args:
//   executor: ForeignExecutor lowered into a uint64_t value
//   delay: Delay in MS
//   task: RustTaskCallback to call
//   task_data: data to pass the task callback
typedef int8_t (*ForeignExecutorCallback)(uint64_t, uint32_t, RustTaskCallback, void *);

typedef struct ForeignBytes {
	int32_t len;
	const uint8_t *data;
} ForeignBytes;

// Error definitions
typedef struct RustCallStatus {
	int8_t code;
	RustBuffer errorBuf;
} RustCallStatus;

// Continuation callback for UniFFI Futures
typedef void (*RustFutureContinuation)(void * , int8_t);

// ⚠️ Attention: If you change this #else block (ending in `#endif // def UNIFFI_SHARED_H`) you *must* ⚠️
// ⚠️ increment the version suffix in all instances of UNIFFI_SHARED_HEADER_V6 in this file.           ⚠️
#endif // def UNIFFI_SHARED_H

// Needed because we can't execute the callback directly from go.
void cgo_rust_task_callback_bridge_telio(RustTaskCallback, const void *, int8_t);

int8_t uniffiForeignExecutorCallbacktelio(uint64_t, uint32_t, RustTaskCallback, void*);

void uniffiFutureContinuationCallbacktelio(void*, int8_t);

void uniffi_telio_fn_free_telio(
	void* ptr,
	RustCallStatus* out_status
);

void* uniffi_telio_fn_constructor_telio_new(
	RustBuffer features,
	uint64_t events,
	RustCallStatus* out_status
);

void* uniffi_telio_fn_constructor_telio_new_with_protect(
	RustBuffer features,
	uint64_t events,
	uint64_t protect,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_connect_to_exit_node(
	void* ptr,
	RustBuffer public_key,
	RustBuffer allowed_ips,
	RustBuffer endpoint,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_connect_to_exit_node_postquantum(
	void* ptr,
	RustBuffer identifier,
	RustBuffer public_key,
	RustBuffer allowed_ips,
	RustBuffer endpoint,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_connect_to_exit_node_with_id(
	void* ptr,
	RustBuffer identifier,
	RustBuffer public_key,
	RustBuffer allowed_ips,
	RustBuffer endpoint,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_disable_magic_dns(
	void* ptr,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_disconnect_from_exit_node(
	void* ptr,
	RustBuffer public_key,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_disconnect_from_exit_nodes(
	void* ptr,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_enable_magic_dns(
	void* ptr,
	RustBuffer forward_servers,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_generate_stack_panic(
	void* ptr,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_generate_thread_panic(
	void* ptr,
	RustCallStatus* out_status
);

uint64_t uniffi_telio_fn_method_telio_get_adapter_luid(
	void* ptr,
	RustCallStatus* out_status
);

RustBuffer uniffi_telio_fn_method_telio_get_last_error(
	void* ptr,
	RustCallStatus* out_status
);

RustBuffer uniffi_telio_fn_method_telio_get_nat(
	void* ptr,
	RustBuffer ip,
	uint16_t port,
	RustCallStatus* out_status
);

RustBuffer uniffi_telio_fn_method_telio_get_secret_key(
	void* ptr,
	RustCallStatus* out_status
);

RustBuffer uniffi_telio_fn_method_telio_get_status_map(
	void* ptr,
	RustCallStatus* out_status
);

int8_t uniffi_telio_fn_method_telio_is_running(
	void* ptr,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_notify_network_change(
	void* ptr,
	RustBuffer network_info,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_notify_sleep(
	void* ptr,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_notify_wakeup(
	void* ptr,
	RustCallStatus* out_status
);

uint32_t uniffi_telio_fn_method_telio_probe_pmtu(
	void* ptr,
	RustBuffer host,
	RustCallStatus* out_status
);

RustBuffer uniffi_telio_fn_method_telio_receive_ping(
	void* ptr,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_set_fwmark(
	void* ptr,
	uint32_t fwmark,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_set_meshnet(
	void* ptr,
	RustBuffer cfg,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_set_meshnet_off(
	void* ptr,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_set_secret_key(
	void* ptr,
	RustBuffer secret_key,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_shutdown(
	void* ptr,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_shutdown_hard(
	void* ptr,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_start(
	void* ptr,
	RustBuffer secret_key,
	RustBuffer adapter,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_start_named(
	void* ptr,
	RustBuffer secret_key,
	RustBuffer adapter,
	RustBuffer name,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_start_with_tun(
	void* ptr,
	RustBuffer secret_key,
	RustBuffer adapter,
	int32_t tun,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_stop(
	void* ptr,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_trigger_analytics_event(
	void* ptr,
	RustCallStatus* out_status
);

void uniffi_telio_fn_method_telio_trigger_qos_collection(
	void* ptr,
	RustCallStatus* out_status
);

void uniffi_telio_fn_init_callback_telioeventcb(
	ForeignCallback callback_stub,
	RustCallStatus* out_status
);

void uniffi_telio_fn_init_callback_teliologgercb(
	ForeignCallback callback_stub,
	RustCallStatus* out_status
);

void uniffi_telio_fn_init_callback_telioprotectcb(
	ForeignCallback callback_stub,
	RustCallStatus* out_status
);

RustBuffer uniffi_telio_fn_func_deserialize_feature_config(
	RustBuffer fstr,
	RustCallStatus* out_status
);

RustBuffer uniffi_telio_fn_func_deserialize_meshnet_config(
	RustBuffer cfg_str,
	RustCallStatus* out_status
);

RustBuffer uniffi_telio_fn_func_generate_public_key(
	RustBuffer secret_key,
	RustCallStatus* out_status
);

RustBuffer uniffi_telio_fn_func_generate_secret_key(
	RustCallStatus* out_status
);

RustBuffer uniffi_telio_fn_func_get_commit_sha(
	RustCallStatus* out_status
);

RustBuffer uniffi_telio_fn_func_get_default_adapter(
	RustCallStatus* out_status
);

RustBuffer uniffi_telio_fn_func_get_default_feature_config(
	RustCallStatus* out_status
);

RustBuffer uniffi_telio_fn_func_get_version_tag(
	RustCallStatus* out_status
);

void uniffi_telio_fn_func_set_global_logger(
	RustBuffer log_level,
	uint64_t logger,
	RustCallStatus* out_status
);

RustBuffer ffi_telio_rustbuffer_alloc(
	int32_t size,
	RustCallStatus* out_status
);

RustBuffer ffi_telio_rustbuffer_from_bytes(
	ForeignBytes bytes,
	RustCallStatus* out_status
);

void ffi_telio_rustbuffer_free(
	RustBuffer buf,
	RustCallStatus* out_status
);

RustBuffer ffi_telio_rustbuffer_reserve(
	RustBuffer buf,
	int32_t additional,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_continuation_callback_set(
	RustFutureContinuation callback,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_poll_u8(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_cancel_u8(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_free_u8(
	void* handle,
	RustCallStatus* out_status
);

uint8_t ffi_telio_rust_future_complete_u8(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_poll_i8(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_cancel_i8(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_free_i8(
	void* handle,
	RustCallStatus* out_status
);

int8_t ffi_telio_rust_future_complete_i8(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_poll_u16(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_cancel_u16(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_free_u16(
	void* handle,
	RustCallStatus* out_status
);

uint16_t ffi_telio_rust_future_complete_u16(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_poll_i16(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_cancel_i16(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_free_i16(
	void* handle,
	RustCallStatus* out_status
);

int16_t ffi_telio_rust_future_complete_i16(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_poll_u32(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_cancel_u32(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_free_u32(
	void* handle,
	RustCallStatus* out_status
);

uint32_t ffi_telio_rust_future_complete_u32(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_poll_i32(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_cancel_i32(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_free_i32(
	void* handle,
	RustCallStatus* out_status
);

int32_t ffi_telio_rust_future_complete_i32(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_poll_u64(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_cancel_u64(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_free_u64(
	void* handle,
	RustCallStatus* out_status
);

uint64_t ffi_telio_rust_future_complete_u64(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_poll_i64(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_cancel_i64(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_free_i64(
	void* handle,
	RustCallStatus* out_status
);

int64_t ffi_telio_rust_future_complete_i64(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_poll_f32(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_cancel_f32(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_free_f32(
	void* handle,
	RustCallStatus* out_status
);

float ffi_telio_rust_future_complete_f32(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_poll_f64(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_cancel_f64(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_free_f64(
	void* handle,
	RustCallStatus* out_status
);

double ffi_telio_rust_future_complete_f64(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_poll_pointer(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_cancel_pointer(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_free_pointer(
	void* handle,
	RustCallStatus* out_status
);

void* ffi_telio_rust_future_complete_pointer(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_poll_rust_buffer(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_cancel_rust_buffer(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_free_rust_buffer(
	void* handle,
	RustCallStatus* out_status
);

RustBuffer ffi_telio_rust_future_complete_rust_buffer(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_poll_void(
	void* handle,
	void* uniffi_callback,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_cancel_void(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_free_void(
	void* handle,
	RustCallStatus* out_status
);

void ffi_telio_rust_future_complete_void(
	void* handle,
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_func_deserialize_feature_config(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_func_deserialize_meshnet_config(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_func_generate_public_key(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_func_generate_secret_key(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_func_get_commit_sha(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_func_get_default_adapter(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_func_get_default_feature_config(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_func_get_version_tag(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_func_set_global_logger(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_connect_to_exit_node(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_connect_to_exit_node_postquantum(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_connect_to_exit_node_with_id(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_disable_magic_dns(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_disconnect_from_exit_node(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_disconnect_from_exit_nodes(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_enable_magic_dns(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_generate_stack_panic(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_generate_thread_panic(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_get_adapter_luid(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_get_last_error(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_get_nat(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_get_secret_key(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_get_status_map(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_is_running(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_notify_network_change(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_notify_sleep(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_notify_wakeup(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_probe_pmtu(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_receive_ping(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_set_fwmark(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_set_meshnet(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_set_meshnet_off(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_set_secret_key(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_shutdown(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_shutdown_hard(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_start(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_start_named(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_start_with_tun(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_stop(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_trigger_analytics_event(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telio_trigger_qos_collection(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_constructor_telio_new(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_constructor_telio_new_with_protect(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telioeventcb_event(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_teliologgercb_log(
	RustCallStatus* out_status
);

uint16_t uniffi_telio_checksum_method_telioprotectcb_protect(
	RustCallStatus* out_status
);

uint32_t ffi_telio_uniffi_contract_version(
	RustCallStatus* out_status
);


int32_t telio_cgo_TelioEventCb(uint64_t, int32_t, uint8_t *, int32_t, RustBuffer *);
int32_t telio_cgo_TelioLoggerCb(uint64_t, int32_t, uint8_t *, int32_t, RustBuffer *);
int32_t telio_cgo_TelioProtectCb(uint64_t, int32_t, uint8_t *, int32_t, RustBuffer *);

