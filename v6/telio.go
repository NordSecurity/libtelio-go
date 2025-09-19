
package telio

// #include <telio.h>
import "C"

import (
	"bytes"
	"fmt"
	"io"
	"unsafe"
	"encoding/binary"
	"errors"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
)



// This is needed, because as of go 1.24
// type RustBuffer C.RustBuffer cannot have methods,
// RustBuffer is treated as non-local type
type GoRustBuffer struct {
	inner C.RustBuffer
}

type RustBufferI interface {
	AsReader() *bytes.Reader
	Free()
	ToGoBytes() []byte
	Data() unsafe.Pointer
	Len() uint64
	Capacity() uint64
}

func RustBufferFromExternal(b RustBufferI) GoRustBuffer {
	return GoRustBuffer {
		inner: C.RustBuffer {
			capacity: C.uint64_t(b.Capacity()),
			len: C.uint64_t(b.Len()),
			data: (*C.uchar)(b.Data()),
		},
	}
}

func (cb GoRustBuffer) Capacity() uint64 {
	return uint64(cb.inner.capacity)
}

func (cb GoRustBuffer) Len() uint64 {
	return uint64(cb.inner.len)
}

func (cb GoRustBuffer) Data() unsafe.Pointer {
	return unsafe.Pointer(cb.inner.data)
}

func (cb GoRustBuffer) AsReader() *bytes.Reader {
	b := unsafe.Slice((*byte)(cb.inner.data), C.uint64_t(cb.inner.len))
	return bytes.NewReader(b)
}

func (cb GoRustBuffer) Free() {
	rustCall(func( status *C.RustCallStatus) bool {
		C.ffi_telio_rustbuffer_free(cb.inner, status)
		return false
	})
}

func (cb GoRustBuffer) ToGoBytes() []byte {
	return C.GoBytes(unsafe.Pointer(cb.inner.data), C.int(cb.inner.len))
}


func stringToRustBuffer(str string) C.RustBuffer {
	return bytesToRustBuffer([]byte(str))
}

func bytesToRustBuffer(b []byte) C.RustBuffer {
	if len(b) == 0 {
		return C.RustBuffer{}
	}
	// We can pass the pointer along here, as it is pinned
	// for the duration of this call
	foreign := C.ForeignBytes {
		len: C.int(len(b)),
		data: (*C.uchar)(unsafe.Pointer(&b[0])),
	}
	
	return rustCall(func( status *C.RustCallStatus) C.RustBuffer {
		return C.ffi_telio_rustbuffer_from_bytes(foreign, status)
	})
}


type BufLifter[GoType any] interface {
	Lift(value RustBufferI) GoType
}

type BufLowerer[GoType any] interface {
	Lower(value GoType) C.RustBuffer
}

type BufReader[GoType any] interface {
	Read(reader io.Reader) GoType
}

type BufWriter[GoType any] interface {
	Write(writer io.Writer, value GoType)
}

func LowerIntoRustBuffer[GoType any](bufWriter BufWriter[GoType], value GoType) C.RustBuffer {
	// This might be not the most efficient way but it does not require knowing allocation size
	// beforehand
	var buffer bytes.Buffer
	bufWriter.Write(&buffer, value)

	bytes, err := io.ReadAll(&buffer)
	if err != nil {
		panic(fmt.Errorf("reading written data: %w", err))
	}
	return bytesToRustBuffer(bytes)
}

func LiftFromRustBuffer[GoType any](bufReader BufReader[GoType], rbuf RustBufferI) GoType {
	defer rbuf.Free()
	reader := rbuf.AsReader()
	item := bufReader.Read(reader)
	if reader.Len() > 0 {
		// TODO: Remove this
		leftover, _ := io.ReadAll(reader)
		panic(fmt.Errorf("Junk remaining in buffer after lifting: %s", string(leftover)))
	}
	return item
}



func rustCallWithError[E any, U any](converter BufReader[*E], callback func(*C.RustCallStatus) U) (U, *E) {
	var status C.RustCallStatus
	returnValue := callback(&status)
	err := checkCallStatus(converter, status)
	return returnValue, err
}

func checkCallStatus[E any](converter BufReader[*E], status C.RustCallStatus) *E {
	switch status.code {
	case 0:
		return nil
	case 1:
		return LiftFromRustBuffer(converter, GoRustBuffer { inner: status.errorBuf })
	case 2:
		// when the rust code sees a panic, it tries to construct a rustBuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(GoRustBuffer { inner: status.errorBuf })))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		panic(fmt.Errorf("unknown status code: %d", status.code))
	}
}

func checkCallStatusUnknown(status C.RustCallStatus) error {
	switch status.code {
	case 0:
		return nil
	case 1:
		panic(fmt.Errorf("function not returning an error returned an error"))
	case 2:
		// when the rust code sees a panic, it tries to construct a C.RustBuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(GoRustBuffer {
				inner: status.errorBuf,
			})))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		return fmt.Errorf("unknown status code: %d", status.code)
	}
}

func rustCall[U any](callback func(*C.RustCallStatus) U) U {
	returnValue, err := rustCallWithError[error](nil, callback)
	if err != nil {
		panic(err)
	}
	return returnValue
}

type NativeError interface {
	AsError() error
}


func writeInt8(writer io.Writer, value int8) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint8(writer io.Writer, value uint8) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt16(writer io.Writer, value int16) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint16(writer io.Writer, value uint16) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt32(writer io.Writer, value int32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint32(writer io.Writer, value uint32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt64(writer io.Writer, value int64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint64(writer io.Writer, value uint64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeFloat32(writer io.Writer, value float32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeFloat64(writer io.Writer, value float64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}


func readInt8(reader io.Reader) int8 {
	var result int8
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint8(reader io.Reader) uint8 {
	var result uint8
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt16(reader io.Reader) int16 {
	var result int16
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint16(reader io.Reader) uint16 {
	var result uint16
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt32(reader io.Reader) int32 {
	var result int32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint32(reader io.Reader) uint32 {
	var result uint32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt64(reader io.Reader) int64 {
	var result int64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint64(reader io.Reader) uint64 {
	var result uint64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readFloat32(reader io.Reader) float32 {
	var result float32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readFloat64(reader io.Reader) float64 {
	var result float64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func init() {
        
        FfiConverterCallbackInterfaceTelioEventCbINSTANCE.register();
        FfiConverterCallbackInterfaceTelioLoggerCbINSTANCE.register();
        FfiConverterCallbackInterfaceTelioProtectCbINSTANCE.register();
        uniffiCheckChecksums()
}


func uniffiCheckChecksums() {
	// Get the bindings contract version from our ComponentInterface
	bindingsContractVersion := 26
	// Get the scaffolding contract version by calling the into the dylib
	scaffoldingContractVersion := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint32_t {
		return C.ffi_telio_uniffi_contract_version()
	})
	if bindingsContractVersion != int(scaffoldingContractVersion) {
		// If this happens try cleaning and rebuilding your project
		panic("telio: UniFFI contract version mismatch")
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_func_add_timestamps_to_logs()
	})
	if checksum != 10620 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_func_add_timestamps_to_logs: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_func_deserialize_feature_config()
	})
	if checksum != 11797 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_func_deserialize_feature_config: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_func_deserialize_meshnet_config()
	})
	if checksum != 53042 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_func_deserialize_meshnet_config: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_func_generate_public_key()
	})
	if checksum != 52233 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_func_generate_public_key: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_func_generate_secret_key()
	})
	if checksum != 5074 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_func_generate_secret_key: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_func_get_commit_sha()
	})
	if checksum != 39165 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_func_get_commit_sha: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_func_get_default_adapter()
	})
	if checksum != 47135 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_func_get_default_adapter: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_func_get_default_feature_config()
	})
	if checksum != 7045 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_func_get_default_feature_config: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_func_get_version_tag()
	})
	if checksum != 53700 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_func_get_version_tag: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_func_serialize_feature_config()
	})
	if checksum != 64562 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_func_serialize_feature_config: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_func_set_global_logger()
	})
	if checksum != 47236 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_func_set_global_logger: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_func_unset_global_logger()
	})
	if checksum != 32201 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_func_unset_global_logger: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_build()
	})
	if checksum != 46753 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_build: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_batching()
	})
	if checksum != 27812 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_batching: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_battery_saving_defaults()
	})
	if checksum != 10214 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_battery_saving_defaults: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_direct()
	})
	if checksum != 8489 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_direct: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_dynamic_wg_nt_control()
	})
	if checksum != 29236 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_dynamic_wg_nt_control: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_error_notification_service()
	})
	if checksum != 60879 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_error_notification_service: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_firewall_connection_reset()
	})
	if checksum != 63055 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_firewall_connection_reset: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_flush_events_on_stop_timeout_seconds()
	})
	if checksum != 48141 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_flush_events_on_stop_timeout_seconds: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_ipv6()
	})
	if checksum != 25251 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_ipv6: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_lana()
	})
	if checksum != 20972 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_lana: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_link_detection()
	})
	if checksum != 35122 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_link_detection: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_multicast()
	})
	if checksum != 10758 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_multicast: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_nicknames()
	})
	if checksum != 59848 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_nicknames: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_nurse()
	})
	if checksum != 24340 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_nurse: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_validate_keys()
	})
	if checksum != 10605 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_validate_keys: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_set_inter_thread_channel_size()
	})
	if checksum != 13344 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_set_inter_thread_channel_size: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_set_max_inter_thread_batched_pkts()
	})
	if checksum != 9508 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_set_max_inter_thread_batched_pkts: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_set_skt_buffer_size()
	})
	if checksum != 35161 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_set_skt_buffer_size: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_connect_to_exit_node()
	})
	if checksum != 5979 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_connect_to_exit_node: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_connect_to_exit_node_postquantum()
	})
	if checksum != 1113 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_connect_to_exit_node_postquantum: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_connect_to_exit_node_with_id()
	})
	if checksum != 16832 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_connect_to_exit_node_with_id: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_disable_magic_dns()
	})
	if checksum != 48202 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_disable_magic_dns: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_disconnect_from_exit_node()
	})
	if checksum != 36107 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_disconnect_from_exit_node: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_disconnect_from_exit_nodes()
	})
	if checksum != 56626 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_disconnect_from_exit_nodes: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_enable_magic_dns()
	})
	if checksum != 32172 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_enable_magic_dns: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_generate_stack_panic()
	})
	if checksum != 48978 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_generate_stack_panic: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_generate_thread_panic()
	})
	if checksum != 3906 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_generate_thread_panic: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_get_adapter_luid()
	})
	if checksum != 53187 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_get_adapter_luid: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_get_last_error()
	})
	if checksum != 1246 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_get_last_error: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_get_secret_key()
	})
	if checksum != 60553 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_get_secret_key: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_get_status_map()
	})
	if checksum != 52925 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_get_status_map: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_is_running()
	})
	if checksum != 5169 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_is_running: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_notify_network_change()
	})
	if checksum != 6052 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_notify_network_change: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_notify_sleep()
	})
	if checksum != 12814 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_notify_sleep: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_notify_wakeup()
	})
	if checksum != 58222 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_notify_wakeup: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_receive_ping()
	})
	if checksum != 36743 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_receive_ping: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_set_fwmark()
	})
	if checksum != 38777 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_set_fwmark: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_set_meshnet()
	})
	if checksum != 52858 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_set_meshnet: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_set_meshnet_off()
	})
	if checksum != 37791 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_set_meshnet_off: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_set_secret_key()
	})
	if checksum != 34182 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_set_secret_key: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_set_tun()
	})
	if checksum != 49747 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_set_tun: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_shutdown()
	})
	if checksum != 25927 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_shutdown: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_shutdown_hard()
	})
	if checksum != 6436 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_shutdown_hard: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_start()
	})
	if checksum != 39667 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_start: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_start_named()
	})
	if checksum != 29016 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_start_named: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_start_with_tun()
	})
	if checksum != 49772 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_start_with_tun: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_stop()
	})
	if checksum != 11709 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_stop: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_trigger_analytics_event()
	})
	if checksum != 2691 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_trigger_analytics_event: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_trigger_qos_collection()
	})
	if checksum != 54684 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_trigger_qos_collection: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_constructor_featuresdefaultsbuilder_new()
	})
	if checksum != 33447 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_constructor_featuresdefaultsbuilder_new: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_constructor_telio_new()
	})
	if checksum != 22327 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_constructor_telio_new: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_constructor_telio_new_with_protect()
	})
	if checksum != 57901 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_constructor_telio_new_with_protect: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telioeventcb_event()
	})
	if checksum != 57944 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telioeventcb_event: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_teliologgercb_log()
	})
	if checksum != 42728 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_teliologgercb_log: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telioprotectcb_protect()
	})
	if checksum != 52662 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telioprotectcb_protect: UniFFI API checksum mismatch")
	}
	}
}




type FfiConverterUint16 struct{}

var FfiConverterUint16INSTANCE = FfiConverterUint16{}

func (FfiConverterUint16) Lower(value uint16) C.uint16_t {
	return C.uint16_t(value)
}

func (FfiConverterUint16) Write(writer io.Writer, value uint16) {
	writeUint16(writer, value)
}

func (FfiConverterUint16) Lift(value C.uint16_t) uint16 {
	return uint16(value)
}

func (FfiConverterUint16) Read(reader io.Reader) uint16 {
	return readUint16(reader)
}

type FfiDestroyerUint16 struct {}

func (FfiDestroyerUint16) Destroy(_ uint16) {}


type FfiConverterUint32 struct{}

var FfiConverterUint32INSTANCE = FfiConverterUint32{}

func (FfiConverterUint32) Lower(value uint32) C.uint32_t {
	return C.uint32_t(value)
}

func (FfiConverterUint32) Write(writer io.Writer, value uint32) {
	writeUint32(writer, value)
}

func (FfiConverterUint32) Lift(value C.uint32_t) uint32 {
	return uint32(value)
}

func (FfiConverterUint32) Read(reader io.Reader) uint32 {
	return readUint32(reader)
}

type FfiDestroyerUint32 struct {}

func (FfiDestroyerUint32) Destroy(_ uint32) {}


type FfiConverterInt32 struct{}

var FfiConverterInt32INSTANCE = FfiConverterInt32{}

func (FfiConverterInt32) Lower(value int32) C.int32_t {
	return C.int32_t(value)
}

func (FfiConverterInt32) Write(writer io.Writer, value int32) {
	writeInt32(writer, value)
}

func (FfiConverterInt32) Lift(value C.int32_t) int32 {
	return int32(value)
}

func (FfiConverterInt32) Read(reader io.Reader) int32 {
	return readInt32(reader)
}

type FfiDestroyerInt32 struct {}

func (FfiDestroyerInt32) Destroy(_ int32) {}


type FfiConverterUint64 struct{}

var FfiConverterUint64INSTANCE = FfiConverterUint64{}

func (FfiConverterUint64) Lower(value uint64) C.uint64_t {
	return C.uint64_t(value)
}

func (FfiConverterUint64) Write(writer io.Writer, value uint64) {
	writeUint64(writer, value)
}

func (FfiConverterUint64) Lift(value C.uint64_t) uint64 {
	return uint64(value)
}

func (FfiConverterUint64) Read(reader io.Reader) uint64 {
	return readUint64(reader)
}

type FfiDestroyerUint64 struct {}

func (FfiDestroyerUint64) Destroy(_ uint64) {}


type FfiConverterBool struct{}

var FfiConverterBoolINSTANCE = FfiConverterBool{}

func (FfiConverterBool) Lower(value bool) C.int8_t {
	if value {
		return C.int8_t(1)
	}
	return C.int8_t(0)
}

func (FfiConverterBool) Write(writer io.Writer, value bool) {
	if value {
		writeInt8(writer, 1)
	} else {
		writeInt8(writer, 0)
	}
}

func (FfiConverterBool) Lift(value C.int8_t) bool {
	return value != 0
}

func (FfiConverterBool) Read(reader io.Reader) bool {
	return readInt8(reader) != 0
}

type FfiDestroyerBool struct {}

func (FfiDestroyerBool) Destroy(_ bool) {}


type FfiConverterString struct{}

var FfiConverterStringINSTANCE = FfiConverterString{}

func (FfiConverterString) Lift(rb RustBufferI) string {
	defer rb.Free()
	reader := rb.AsReader()
	b, err := io.ReadAll(reader)
	if err != nil {
		panic(fmt.Errorf("reading reader: %w", err))
	}
	return string(b)
}

func (FfiConverterString) Read(reader io.Reader) string {
	length := readInt32(reader)
	buffer := make([]byte, length)
	read_length, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		panic(err)
	}
	if read_length != int(length) {
		panic(fmt.Errorf("bad read length when reading string, expected %d, read %d", length, read_length))
	}
	return string(buffer)
}

func (FfiConverterString) Lower(value string) C.RustBuffer {
	return stringToRustBuffer(value)
}

func (FfiConverterString) Write(writer io.Writer, value string) {
	if len(value) > math.MaxInt32 {
		panic("String is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	write_length, err := io.WriteString(writer, value)
	if err != nil {
		panic(err)
	}
	if write_length != len(value) {
		panic(fmt.Errorf("bad write length when writing string, expected %d, written %d", len(value), write_length))
	}
}

type FfiDestroyerString struct {}

func (FfiDestroyerString) Destroy(_ string) {}



// Below is an implementation of synchronization requirements outlined in the link.
// https://github.com/mozilla/uniffi-rs/blob/0dc031132d9493ca812c3af6e7dd60ad2ea95bf0/uniffi_bindgen/src/bindings/kotlin/templates/ObjectRuntime.kt#L31

type FfiObject struct {
	pointer unsafe.Pointer
	callCounter atomic.Int64
	cloneFunction func(unsafe.Pointer, *C.RustCallStatus) unsafe.Pointer
	freeFunction func(unsafe.Pointer, *C.RustCallStatus)
	destroyed atomic.Bool
}

func newFfiObject(
	pointer unsafe.Pointer, 
	cloneFunction func(unsafe.Pointer, *C.RustCallStatus) unsafe.Pointer, 
	freeFunction func(unsafe.Pointer, *C.RustCallStatus),
) FfiObject {
	return FfiObject {
		pointer: pointer,
		cloneFunction: cloneFunction, 
		freeFunction: freeFunction,
	}
}

func (ffiObject *FfiObject)incrementPointer(debugName string) unsafe.Pointer {
	for {
		counter := ffiObject.callCounter.Load()
		if counter <= -1 {
			panic(fmt.Errorf("%v object has already been destroyed", debugName))
		}
		if counter == math.MaxInt64 {
			panic(fmt.Errorf("%v object call counter would overflow", debugName))
		}
		if ffiObject.callCounter.CompareAndSwap(counter, counter + 1) {
			break
		}
	}

	return rustCall(func(status *C.RustCallStatus) unsafe.Pointer {
		return ffiObject.cloneFunction(ffiObject.pointer, status)
	})
}

func (ffiObject *FfiObject)decrementPointer() {
	if ffiObject.callCounter.Add(-1) == -1 {
		ffiObject.freeRustArcPtr()
	}
}

func (ffiObject *FfiObject)destroy() {
	if ffiObject.destroyed.CompareAndSwap(false, true) {
		if ffiObject.callCounter.Add(-1) == -1 {
			ffiObject.freeRustArcPtr()
		}
	}
}

func (ffiObject *FfiObject)freeRustArcPtr() {
	rustCall(func(status *C.RustCallStatus) int32 {
		ffiObject.freeFunction(ffiObject.pointer, status)
		return 0
	})
}
// A [Features] builder that allows a simpler initialization of
// features with defaults comming from libtelio lib.
//
// !!! Should only be used then remote config is inaccessible !!!

type FeaturesDefaultsBuilderInterface interface {
	// Build final config
	Build() Features
	// Enable keepalive batching feature
	EnableBatching() *FeaturesDefaultsBuilder
	// Enable default wireguard timings, derp timings and other features for best battery performance
	EnableBatterySavingDefaults() *FeaturesDefaultsBuilder
	// Enable direct connections with defaults;
	EnableDirect() *FeaturesDefaultsBuilder
	// Enable dynamic WireGuard-NT control as per RFC LLT-0089
	EnableDynamicWgNtControl() *FeaturesDefaultsBuilder
	EnableErrorNotificationService() *FeaturesDefaultsBuilder
	// Enable firewall connection resets when NepTUN is used
	EnableFirewallConnectionReset() *FeaturesDefaultsBuilder
	// Enable blocking event flush with timout on stop with defaults
	EnableFlushEventsOnStopTimeoutSeconds() *FeaturesDefaultsBuilder
	// Enable IPv6 with defaults
	EnableIpv6() *FeaturesDefaultsBuilder
	// Enable lana, this requires input from apps
	EnableLana(eventPath string, isProd bool) *FeaturesDefaultsBuilder
	// Enable Link detection mechanism with defaults
	EnableLinkDetection() *FeaturesDefaultsBuilder
	// Eanable multicast with defaults
	EnableMulticast() *FeaturesDefaultsBuilder
	// Enable nicknames with defaults
	EnableNicknames() *FeaturesDefaultsBuilder
	// Enable nurse with defaults
	EnableNurse() *FeaturesDefaultsBuilder
	// Enable key valiation in set_config call with defaults
	EnableValidateKeys() *FeaturesDefaultsBuilder
	// Enable custom socket buffer sizes for NepTUN
	SetInterThreadChannelSize(interThreadChannelSize uint32) *FeaturesDefaultsBuilder
	// Enable custom socket buffer sizes for NepTUN
	SetMaxInterThreadBatchedPkts(maxInterThreadBatchedPkts uint32) *FeaturesDefaultsBuilder
	// Enable custom socket buffer sizes for NepTUN
	SetSktBufferSize(sktBufferSize uint32) *FeaturesDefaultsBuilder
}
// A [Features] builder that allows a simpler initialization of
// features with defaults comming from libtelio lib.
//
// !!! Should only be used then remote config is inaccessible !!!

type FeaturesDefaultsBuilder struct {
	ffiObject FfiObject
}
// Create a builder for Features with minimal defaults.
func NewFeaturesDefaultsBuilder() *FeaturesDefaultsBuilder {
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_constructor_featuresdefaultsbuilder_new(_uniffiStatus)
	}))
}




// Build final config
func (_self *FeaturesDefaultsBuilder) Build() Features {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer {
		inner: C.uniffi_telio_fn_method_featuresdefaultsbuilder_build(
		_pointer,_uniffiStatus),
	}
	}))
}

// Enable keepalive batching feature
func (_self *FeaturesDefaultsBuilder) EnableBatching() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_batching(
		_pointer,_uniffiStatus)
	}))
}

// Enable default wireguard timings, derp timings and other features for best battery performance
func (_self *FeaturesDefaultsBuilder) EnableBatterySavingDefaults() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_battery_saving_defaults(
		_pointer,_uniffiStatus)
	}))
}

// Enable direct connections with defaults;
func (_self *FeaturesDefaultsBuilder) EnableDirect() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_direct(
		_pointer,_uniffiStatus)
	}))
}

// Enable dynamic WireGuard-NT control as per RFC LLT-0089
func (_self *FeaturesDefaultsBuilder) EnableDynamicWgNtControl() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_dynamic_wg_nt_control(
		_pointer,_uniffiStatus)
	}))
}

func (_self *FeaturesDefaultsBuilder) EnableErrorNotificationService() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_error_notification_service(
		_pointer,_uniffiStatus)
	}))
}

// Enable firewall connection resets when NepTUN is used
func (_self *FeaturesDefaultsBuilder) EnableFirewallConnectionReset() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_firewall_connection_reset(
		_pointer,_uniffiStatus)
	}))
}

// Enable blocking event flush with timout on stop with defaults
func (_self *FeaturesDefaultsBuilder) EnableFlushEventsOnStopTimeoutSeconds() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_flush_events_on_stop_timeout_seconds(
		_pointer,_uniffiStatus)
	}))
}

// Enable IPv6 with defaults
func (_self *FeaturesDefaultsBuilder) EnableIpv6() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_ipv6(
		_pointer,_uniffiStatus)
	}))
}

// Enable lana, this requires input from apps
func (_self *FeaturesDefaultsBuilder) EnableLana(eventPath string, isProd bool) *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_lana(
		_pointer,FfiConverterStringINSTANCE.Lower(eventPath), FfiConverterBoolINSTANCE.Lower(isProd),_uniffiStatus)
	}))
}

// Enable Link detection mechanism with defaults
func (_self *FeaturesDefaultsBuilder) EnableLinkDetection() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_link_detection(
		_pointer,_uniffiStatus)
	}))
}

// Eanable multicast with defaults
func (_self *FeaturesDefaultsBuilder) EnableMulticast() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_multicast(
		_pointer,_uniffiStatus)
	}))
}

// Enable nicknames with defaults
func (_self *FeaturesDefaultsBuilder) EnableNicknames() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_nicknames(
		_pointer,_uniffiStatus)
	}))
}

// Enable nurse with defaults
func (_self *FeaturesDefaultsBuilder) EnableNurse() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_nurse(
		_pointer,_uniffiStatus)
	}))
}

// Enable key valiation in set_config call with defaults
func (_self *FeaturesDefaultsBuilder) EnableValidateKeys() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_validate_keys(
		_pointer,_uniffiStatus)
	}))
}

// Enable custom socket buffer sizes for NepTUN
func (_self *FeaturesDefaultsBuilder) SetInterThreadChannelSize(interThreadChannelSize uint32) *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_set_inter_thread_channel_size(
		_pointer,FfiConverterUint32INSTANCE.Lower(interThreadChannelSize),_uniffiStatus)
	}))
}

// Enable custom socket buffer sizes for NepTUN
func (_self *FeaturesDefaultsBuilder) SetMaxInterThreadBatchedPkts(maxInterThreadBatchedPkts uint32) *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_set_max_inter_thread_batched_pkts(
		_pointer,FfiConverterUint32INSTANCE.Lower(maxInterThreadBatchedPkts),_uniffiStatus)
	}))
}

// Enable custom socket buffer sizes for NepTUN
func (_self *FeaturesDefaultsBuilder) SetSktBufferSize(sktBufferSize uint32) *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_set_skt_buffer_size(
		_pointer,FfiConverterUint32INSTANCE.Lower(sktBufferSize),_uniffiStatus)
	}))
}
func (object *FeaturesDefaultsBuilder) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterFeaturesDefaultsBuilder struct {}

var FfiConverterFeaturesDefaultsBuilderINSTANCE = FfiConverterFeaturesDefaultsBuilder{}


func (c FfiConverterFeaturesDefaultsBuilder) Lift(pointer unsafe.Pointer) *FeaturesDefaultsBuilder {
	result := &FeaturesDefaultsBuilder {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) unsafe.Pointer {
				return C.uniffi_telio_fn_clone_featuresdefaultsbuilder(pointer, status)
			},
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_telio_fn_free_featuresdefaultsbuilder(pointer, status)
			},
		),
	}
	runtime.SetFinalizer(result, (*FeaturesDefaultsBuilder).Destroy)
	return result
}

func (c FfiConverterFeaturesDefaultsBuilder) Read(reader io.Reader) *FeaturesDefaultsBuilder {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterFeaturesDefaultsBuilder) Lower(value *FeaturesDefaultsBuilder) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer value.ffiObject.decrementPointer()
	return pointer
	
}

func (c FfiConverterFeaturesDefaultsBuilder) Write(writer io.Writer, value *FeaturesDefaultsBuilder) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerFeaturesDefaultsBuilder struct {}

func (_ FfiDestroyerFeaturesDefaultsBuilder) Destroy(value *FeaturesDefaultsBuilder) {
		value.Destroy()
}





type TelioInterface interface {
	// Wrapper for `telio_connect_to_exit_node_with_id` that doesn't take an identifier
	ConnectToExitNode(publicKey PublicKey, allowedIps *[]IpNet, endpoint *SocketAddr) error
	// Connects to the VPN exit node with post quantum tunnel
	//
	// Routing should be set by the user accordingly.
	//
	// # Parameters
	// - `identifier`: String that identifies the exit node, will be generated if null is passed.
	// - `public_key`: Base64 encoded WireGuard public key for an exit node.
	// - `allowed_ips`: Semicolon separated list of subnets which will be routed to the exit node.
	//                  Can be NULL, same as "0.0.0.0/0".
	// - `endpoint`: An endpoint to an exit node. Must contain a port.
	//
	// # Examples
	//
	// ```c
	// // Connects to VPN exit node.
	// telio.connect_to_exit_node_postquantum(
	//     "5e0009e1-75cf-4406-b9ce-0cbb4ea50366",
	//     "QKyApX/ewza7QEbC03Yt8t2ghu6nV5/rve/ZJvsecXo=",
	//     "0.0.0.0/0", // Equivalent
	//     "1.2.3.4:5678"
	// );
	//
	// // Connects to VPN exit node, with specified allowed_ips.
	// telio.connect_to_exit_node_postquantum(
	//     "5e0009e1-75cf-4406-b9ce-0cbb4ea50366",
	//     "QKyApX/ewza7QEbC03Yt8t2ghu6nV5/rve/ZJvsecXo=",
	//     "100.100.0.0/16;10.10.23.0/24",
	//     "1.2.3.4:5678"
	// );
	// ```

	ConnectToExitNodePostquantum(identifier *string, publicKey PublicKey, allowedIps *[]IpNet, endpoint SocketAddr) error
	// Connects to an exit node. (VPN if endpoint is not NULL, Peer if endpoint is NULL)
	//
	// Routing should be set by the user accordingly.
	//
	// # Parameters
	// - `identifier`: String that identifies the exit node, will be generated if null is passed.
	// - `public_key`: WireGuard public key for an exit node.
	// - `allowed_ips`: List of subnets which will be routed to the exit node.
	//                  Can be None, same as "0.0.0.0/0".
	// - `endpoint`: An endpoint to an exit node. Can be None, must contain a port.
	//
	// # Examples
	//
	// ```c
	// // Connects to VPN exit node.
	// telio.connect_to_exit_node_with_id(
	//     "5e0009e1-75cf-4406-b9ce-0cbb4ea50366",
	//     "QKyApX/ewza7QEbC03Yt8t2ghu6nV5/rve/ZJvsecXo=",
	//     "0.0.0.0/0", // Equivalent
	//     "1.2.3.4:5678"
	// );
	//
	// // Connects to VPN exit node, with specified allowed_ips.
	// telio.connect_to_exit_node_with_id(
	//     "5e0009e1-75cf-4406-b9ce-0cbb4ea50366",
	//     "QKyApX/ewza7QEbC03Yt8t2ghu6nV5/rve/ZJvsecXo=",
	//     "100.100.0.0/16;10.10.23.0/24",
	//     "1.2.3.4:5678"
	// );
	//
	// // Connect to exit peer via DERP
	// telio.connect_to_exit_node_with_id(
	//     "5e0009e1-75cf-4406-b9ce-0cbb4ea50366",
	//     "QKyApX/ewza7QEbC03Yt8t2ghu6nV5/rve/ZJvsecXo=",
	//     "0.0.0.0/0",
	//     NULL
	// );
	// ```

	ConnectToExitNodeWithId(identifier *string, publicKey PublicKey, allowedIps *[]IpNet, endpoint *SocketAddr) error
	// Disables magic DNS if it was enabled.
	DisableMagicDns() error
	// Disconnects from specified exit node.
	//
	// # Parameters
	// - `public_key`: WireGuard public key for exit node.

	DisconnectFromExitNode(publicKey PublicKey) error
	// Disconnects from all exit nodes with no parameters required.
	DisconnectFromExitNodes() error
	// Enables magic DNS if it was not enabled yet,
	//
	// Routing should be set by the user accordingly.
	//
	// # Parameters
	// - 'forward_servers': List of DNS servers to route the requests trough.
	//
	// # Examples
	//
	// ```c
	// // Enable magic dns with some forward servers
	// telio.enable_magic_dns("[\"1.1.1.1\", \"8.8.8.8\"]");
	//
	// // Enable magic dns with no forward server
	// telio.enable_magic_dns("[\"\"]");
	// ```
	EnableMagicDns(forwardServers []IpAddr) error
	// For testing only.
	GenerateStackPanic() error
	// For testing only.
	GenerateThreadPanic() error
	// get device luid.
	GetAdapterLuid() uint64
	// Get last error's message length, including trailing null
	GetLastError() string
	GetSecretKey() SecretKey
	GetStatusMap() []TelioNode
	IsRunning() (bool, error)
	// Notify telio with network state changes.
	//
	// # Parameters
	// - `network_info`: Json encoded network sate info.
	//                   Format to be decided, pass empty string for now.
	NotifyNetworkChange(networkInfo string) error
	// Notify telio when system goes to sleep.
	NotifySleep() error
	// Notify telio when system wakes up.
	NotifyWakeup() error
	ReceivePing() (string, error)
	// Sets fmark for started device.
	//
	// # Parameters
	// - `fwmark`: unsigned 32-bit integer

	SetFwmark(fwmark uint32) error
	// Enables meshnet if it is not enabled yet.
	// In case meshnet is enabled, this updates the peer map with the specified one.
	//
	// # Parameters
	// - `cfg`: Output of GET /v1/meshnet/machines/{machineIdentifier}/map

	SetMeshnet(cfg Config) error
	// Disables the meshnet functionality by closing all the connections.
	SetMeshnetOff() error
	// Sets private key for started device.
	//
	// If private_key is not set, device will never connect.
	//
	// # Parameters
	// - `private_key`: WireGuard private key.

	SetSecretKey(secretKey SecretKey) error
	// Sets the tunnel file descriptor
	//
	// # Parameters:
	// - `tun`: the file descriptor of the TUN interface

	SetTun(tun int32) error
	// Completely stop and uninit telio lib.
	Shutdown() error
	// Explicitly deallocate telio object and shutdown async rt.
	ShutdownHard() error
	// Start telio with specified adapter.
	//
	// Adapter will attempt to open its own tunnel.
	Start(secretKey SecretKey, adapter TelioAdapterType) error
	// Start telio with specified adapter and name.
	//
	// Adapter will attempt to open its own tunnel.
	StartNamed(secretKey SecretKey, adapter TelioAdapterType, name string) error
	// Start telio device with specified adapter and already open tunnel.
	//
	// Telio will take ownership of tunnel , and close it on stop.
	//
	// # Parameters
	// - `private_key`: base64 encoded private_key.
	// - `adapter`: Adapter type.
	// - `tun`: A valid filedescriptor to tun device.

	StartWithTun(secretKey SecretKey, adapter TelioAdapterType, tun int32) error
	// Stop telio device.
	Stop() error
	TriggerAnalyticsEvent() error
	TriggerQosCollection() error
}
type Telio struct {
	ffiObject FfiObject
}
// Create new telio library instance
// # Parameters
// - `events`:     Events callback
// - `features`:   JSON string of enabled features
func NewTelio(features Features, events TelioEventCb) (*Telio, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_constructor_telio_new(FfiConverterFeaturesINSTANCE.Lower(features), FfiConverterCallbackInterfaceTelioEventCbINSTANCE.Lower(events),_uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Telio
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTelioINSTANCE.Lift(_uniffiRV), nil
		}
}


// Create new telio library instance
// # Parameters
// - `events`:     Events callback
// - `features`:   JSON string of enabled features
// - `protect`:    Callback executed after exit-node connect (for VpnService::protectFromVpn())
func TelioNewWithProtect(features Features, events TelioEventCb, protect TelioProtectCb) (*Telio, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_constructor_telio_new_with_protect(FfiConverterFeaturesINSTANCE.Lower(features), FfiConverterCallbackInterfaceTelioEventCbINSTANCE.Lower(events), FfiConverterCallbackInterfaceTelioProtectCbINSTANCE.Lower(protect),_uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Telio
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTelioINSTANCE.Lift(_uniffiRV), nil
		}
}



// Wrapper for `telio_connect_to_exit_node_with_id` that doesn't take an identifier
func (_self *Telio) ConnectToExitNode(publicKey PublicKey, allowedIps *[]IpNet, endpoint *SocketAddr) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_connect_to_exit_node(
		_pointer,FfiConverterTypePublicKeyINSTANCE.Lower(publicKey), FfiConverterOptionalSequenceTypeIpNetINSTANCE.Lower(allowedIps), FfiConverterOptionalTypeSocketAddrINSTANCE.Lower(endpoint),_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

// Connects to the VPN exit node with post quantum tunnel
//
// Routing should be set by the user accordingly.
//
// # Parameters
// - `identifier`: String that identifies the exit node, will be generated if null is passed.
// - `public_key`: Base64 encoded WireGuard public key for an exit node.
// - `allowed_ips`: Semicolon separated list of subnets which will be routed to the exit node.
//                  Can be NULL, same as "0.0.0.0/0".
// - `endpoint`: An endpoint to an exit node. Must contain a port.
//
// # Examples
//
// ```c
// // Connects to VPN exit node.
// telio.connect_to_exit_node_postquantum(
//     "5e0009e1-75cf-4406-b9ce-0cbb4ea50366",
//     "QKyApX/ewza7QEbC03Yt8t2ghu6nV5/rve/ZJvsecXo=",
//     "0.0.0.0/0", // Equivalent
//     "1.2.3.4:5678"
// );
//
// // Connects to VPN exit node, with specified allowed_ips.
// telio.connect_to_exit_node_postquantum(
//     "5e0009e1-75cf-4406-b9ce-0cbb4ea50366",
//     "QKyApX/ewza7QEbC03Yt8t2ghu6nV5/rve/ZJvsecXo=",
//     "100.100.0.0/16;10.10.23.0/24",
//     "1.2.3.4:5678"
// );
// ```

func (_self *Telio) ConnectToExitNodePostquantum(identifier *string, publicKey PublicKey, allowedIps *[]IpNet, endpoint SocketAddr) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_connect_to_exit_node_postquantum(
		_pointer,FfiConverterOptionalStringINSTANCE.Lower(identifier), FfiConverterTypePublicKeyINSTANCE.Lower(publicKey), FfiConverterOptionalSequenceTypeIpNetINSTANCE.Lower(allowedIps), FfiConverterTypeSocketAddrINSTANCE.Lower(endpoint),_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

// Connects to an exit node. (VPN if endpoint is not NULL, Peer if endpoint is NULL)
//
// Routing should be set by the user accordingly.
//
// # Parameters
// - `identifier`: String that identifies the exit node, will be generated if null is passed.
// - `public_key`: WireGuard public key for an exit node.
// - `allowed_ips`: List of subnets which will be routed to the exit node.
//                  Can be None, same as "0.0.0.0/0".
// - `endpoint`: An endpoint to an exit node. Can be None, must contain a port.
//
// # Examples
//
// ```c
// // Connects to VPN exit node.
// telio.connect_to_exit_node_with_id(
//     "5e0009e1-75cf-4406-b9ce-0cbb4ea50366",
//     "QKyApX/ewza7QEbC03Yt8t2ghu6nV5/rve/ZJvsecXo=",
//     "0.0.0.0/0", // Equivalent
//     "1.2.3.4:5678"
// );
//
// // Connects to VPN exit node, with specified allowed_ips.
// telio.connect_to_exit_node_with_id(
//     "5e0009e1-75cf-4406-b9ce-0cbb4ea50366",
//     "QKyApX/ewza7QEbC03Yt8t2ghu6nV5/rve/ZJvsecXo=",
//     "100.100.0.0/16;10.10.23.0/24",
//     "1.2.3.4:5678"
// );
//
// // Connect to exit peer via DERP
// telio.connect_to_exit_node_with_id(
//     "5e0009e1-75cf-4406-b9ce-0cbb4ea50366",
//     "QKyApX/ewza7QEbC03Yt8t2ghu6nV5/rve/ZJvsecXo=",
//     "0.0.0.0/0",
//     NULL
// );
// ```

func (_self *Telio) ConnectToExitNodeWithId(identifier *string, publicKey PublicKey, allowedIps *[]IpNet, endpoint *SocketAddr) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_connect_to_exit_node_with_id(
		_pointer,FfiConverterOptionalStringINSTANCE.Lower(identifier), FfiConverterTypePublicKeyINSTANCE.Lower(publicKey), FfiConverterOptionalSequenceTypeIpNetINSTANCE.Lower(allowedIps), FfiConverterOptionalTypeSocketAddrINSTANCE.Lower(endpoint),_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

// Disables magic DNS if it was enabled.
func (_self *Telio) DisableMagicDns() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_disable_magic_dns(
		_pointer,_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

// Disconnects from specified exit node.
//
// # Parameters
// - `public_key`: WireGuard public key for exit node.

func (_self *Telio) DisconnectFromExitNode(publicKey PublicKey) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_disconnect_from_exit_node(
		_pointer,FfiConverterTypePublicKeyINSTANCE.Lower(publicKey),_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

// Disconnects from all exit nodes with no parameters required.
func (_self *Telio) DisconnectFromExitNodes() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_disconnect_from_exit_nodes(
		_pointer,_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

// Enables magic DNS if it was not enabled yet,
//
// Routing should be set by the user accordingly.
//
// # Parameters
// - 'forward_servers': List of DNS servers to route the requests trough.
//
// # Examples
//
// ```c
// // Enable magic dns with some forward servers
// telio.enable_magic_dns("[\"1.1.1.1\", \"8.8.8.8\"]");
//
// // Enable magic dns with no forward server
// telio.enable_magic_dns("[\"\"]");
// ```
func (_self *Telio) EnableMagicDns(forwardServers []IpAddr) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_enable_magic_dns(
		_pointer,FfiConverterSequenceTypeIpAddrINSTANCE.Lower(forwardServers),_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

// For testing only.
func (_self *Telio) GenerateStackPanic() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_generate_stack_panic(
		_pointer,_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

// For testing only.
func (_self *Telio) GenerateThreadPanic() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_generate_thread_panic(
		_pointer,_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

// get device luid.
func (_self *Telio) GetAdapterLuid() uint64 {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterUint64INSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint64_t {
		return C.uniffi_telio_fn_method_telio_get_adapter_luid(
		_pointer,_uniffiStatus)
	}))
}

// Get last error's message length, including trailing null
func (_self *Telio) GetLastError() string {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterStringINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer {
		inner: C.uniffi_telio_fn_method_telio_get_last_error(
		_pointer,_uniffiStatus),
	}
	}))
}

func (_self *Telio) GetSecretKey() SecretKey {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeSecretKeyINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer {
		inner: C.uniffi_telio_fn_method_telio_get_secret_key(
		_pointer,_uniffiStatus),
	}
	}))
}

func (_self *Telio) GetStatusMap() []TelioNode {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterSequenceTelioNodeINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer {
		inner: C.uniffi_telio_fn_method_telio_get_status_map(
		_pointer,_uniffiStatus),
	}
	}))
}

func (_self *Telio) IsRunning() (bool, error) {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_telio_fn_method_telio_is_running(
		_pointer,_uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue bool
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterBoolINSTANCE.Lift(_uniffiRV), nil
		}
}

// Notify telio with network state changes.
//
// # Parameters
// - `network_info`: Json encoded network sate info.
//                   Format to be decided, pass empty string for now.
func (_self *Telio) NotifyNetworkChange(networkInfo string) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_notify_network_change(
		_pointer,FfiConverterStringINSTANCE.Lower(networkInfo),_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

// Notify telio when system goes to sleep.
func (_self *Telio) NotifySleep() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_notify_sleep(
		_pointer,_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

// Notify telio when system wakes up.
func (_self *Telio) NotifyWakeup() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_notify_wakeup(
		_pointer,_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

func (_self *Telio) ReceivePing() (string, error) {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer {
		inner: C.uniffi_telio_fn_method_telio_receive_ping(
		_pointer,_uniffiStatus),
	}
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue string
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterStringINSTANCE.Lift(_uniffiRV), nil
		}
}

// Sets fmark for started device.
//
// # Parameters
// - `fwmark`: unsigned 32-bit integer

func (_self *Telio) SetFwmark(fwmark uint32) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_set_fwmark(
		_pointer,FfiConverterUint32INSTANCE.Lower(fwmark),_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

// Enables meshnet if it is not enabled yet.
// In case meshnet is enabled, this updates the peer map with the specified one.
//
// # Parameters
// - `cfg`: Output of GET /v1/meshnet/machines/{machineIdentifier}/map

func (_self *Telio) SetMeshnet(cfg Config) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_set_meshnet(
		_pointer,FfiConverterConfigINSTANCE.Lower(cfg),_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

// Disables the meshnet functionality by closing all the connections.
func (_self *Telio) SetMeshnetOff() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_set_meshnet_off(
		_pointer,_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

// Sets private key for started device.
//
// If private_key is not set, device will never connect.
//
// # Parameters
// - `private_key`: WireGuard private key.

func (_self *Telio) SetSecretKey(secretKey SecretKey) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_set_secret_key(
		_pointer,FfiConverterTypeSecretKeyINSTANCE.Lower(secretKey),_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

// Sets the tunnel file descriptor
//
// # Parameters:
// - `tun`: the file descriptor of the TUN interface

func (_self *Telio) SetTun(tun int32) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_set_tun(
		_pointer,FfiConverterInt32INSTANCE.Lower(tun),_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

// Completely stop and uninit telio lib.
func (_self *Telio) Shutdown() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_shutdown(
		_pointer,_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

// Explicitly deallocate telio object and shutdown async rt.
func (_self *Telio) ShutdownHard() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_shutdown_hard(
		_pointer,_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

// Start telio with specified adapter.
//
// Adapter will attempt to open its own tunnel.
func (_self *Telio) Start(secretKey SecretKey, adapter TelioAdapterType) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_start(
		_pointer,FfiConverterTypeSecretKeyINSTANCE.Lower(secretKey), FfiConverterTelioAdapterTypeINSTANCE.Lower(adapter),_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

// Start telio with specified adapter and name.
//
// Adapter will attempt to open its own tunnel.
func (_self *Telio) StartNamed(secretKey SecretKey, adapter TelioAdapterType, name string) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_start_named(
		_pointer,FfiConverterTypeSecretKeyINSTANCE.Lower(secretKey), FfiConverterTelioAdapterTypeINSTANCE.Lower(adapter), FfiConverterStringINSTANCE.Lower(name),_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

// Start telio device with specified adapter and already open tunnel.
//
// Telio will take ownership of tunnel , and close it on stop.
//
// # Parameters
// - `private_key`: base64 encoded private_key.
// - `adapter`: Adapter type.
// - `tun`: A valid filedescriptor to tun device.

func (_self *Telio) StartWithTun(secretKey SecretKey, adapter TelioAdapterType, tun int32) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_start_with_tun(
		_pointer,FfiConverterTypeSecretKeyINSTANCE.Lower(secretKey), FfiConverterTelioAdapterTypeINSTANCE.Lower(adapter), FfiConverterInt32INSTANCE.Lower(tun),_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

// Stop telio device.
func (_self *Telio) Stop() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_stop(
		_pointer,_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

func (_self *Telio) TriggerAnalyticsEvent() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_trigger_analytics_event(
		_pointer,_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}

func (_self *Telio) TriggerQosCollection() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_trigger_qos_collection(
		_pointer,_uniffiStatus)
		return false
	})
		return _uniffiErr.AsError()
}
func (object *Telio) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterTelio struct {}

var FfiConverterTelioINSTANCE = FfiConverterTelio{}


func (c FfiConverterTelio) Lift(pointer unsafe.Pointer) *Telio {
	result := &Telio {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) unsafe.Pointer {
				return C.uniffi_telio_fn_clone_telio(pointer, status)
			},
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_telio_fn_free_telio(pointer, status)
			},
		),
	}
	runtime.SetFinalizer(result, (*Telio).Destroy)
	return result
}

func (c FfiConverterTelio) Read(reader io.Reader) *Telio {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterTelio) Lower(value *Telio) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*Telio")
	defer value.ffiObject.decrementPointer()
	return pointer
	
}

func (c FfiConverterTelio) Write(writer io.Writer, value *Telio) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerTelio struct {}

func (_ FfiDestroyerTelio) Destroy(value *Telio) {
		value.Destroy()
}





// Rust representation of [meshnet map]
// A network map of all the Peers and the servers
type Config struct {
	// Description of the local peer
	This PeerBase
	// List of connected peers
	Peers *[]Peer
	// List of available derp servers
	DerpServers *[]Server
	// Dns configuration
	Dns *DnsConfig
}

func (r *Config) Destroy() {
		FfiDestroyerPeerBase{}.Destroy(r.This);
		FfiDestroyerOptionalSequencePeer{}.Destroy(r.Peers);
		FfiDestroyerOptionalSequenceServer{}.Destroy(r.DerpServers);
		FfiDestroyerOptionalDnsConfig{}.Destroy(r.Dns);
}

type FfiConverterConfig struct {}

var FfiConverterConfigINSTANCE = FfiConverterConfig{}

func (c FfiConverterConfig) Lift(rb RustBufferI) Config {
	return LiftFromRustBuffer[Config](c, rb)
}

func (c FfiConverterConfig) Read(reader io.Reader) Config {
	return Config {
			FfiConverterPeerBaseINSTANCE.Read(reader),
			FfiConverterOptionalSequencePeerINSTANCE.Read(reader),
			FfiConverterOptionalSequenceServerINSTANCE.Read(reader),
			FfiConverterOptionalDnsConfigINSTANCE.Read(reader),
	}
}

func (c FfiConverterConfig) Lower(value Config) C.RustBuffer {
	return LowerIntoRustBuffer[Config](c, value)
}

func (c FfiConverterConfig) Write(writer io.Writer, value Config) {
		FfiConverterPeerBaseINSTANCE.Write(writer, value.This);
		FfiConverterOptionalSequencePeerINSTANCE.Write(writer, value.Peers);
		FfiConverterOptionalSequenceServerINSTANCE.Write(writer, value.DerpServers);
		FfiConverterOptionalDnsConfigINSTANCE.Write(writer, value.Dns);
}

type FfiDestroyerConfig struct {}

func (_ FfiDestroyerConfig) Destroy(value Config) {
	value.Destroy()
}


// Representation of DNS configuration
type DnsConfig struct {
	// List of DNS servers
	DnsServers *[]IpAddr
}

func (r *DnsConfig) Destroy() {
		FfiDestroyerOptionalSequenceTypeIpAddr{}.Destroy(r.DnsServers);
}

type FfiConverterDnsConfig struct {}

var FfiConverterDnsConfigINSTANCE = FfiConverterDnsConfig{}

func (c FfiConverterDnsConfig) Lift(rb RustBufferI) DnsConfig {
	return LiftFromRustBuffer[DnsConfig](c, rb)
}

func (c FfiConverterDnsConfig) Read(reader io.Reader) DnsConfig {
	return DnsConfig {
			FfiConverterOptionalSequenceTypeIpAddrINSTANCE.Read(reader),
	}
}

func (c FfiConverterDnsConfig) Lower(value DnsConfig) C.RustBuffer {
	return LowerIntoRustBuffer[DnsConfig](c, value)
}

func (c FfiConverterDnsConfig) Write(writer io.Writer, value DnsConfig) {
		FfiConverterOptionalSequenceTypeIpAddrINSTANCE.Write(writer, value.DnsServers);
}

type FfiDestroyerDnsConfig struct {}

func (_ FfiDestroyerDnsConfig) Destroy(value DnsConfig) {
	value.Destroy()
}


// Error event. Used to inform the upper layer about errors in `libtelio`.
type ErrorEvent struct {
	// The level of the error
	Level ErrorLevel
	// The error code, used to denote the type of the error
	Code ErrorCode
	// A more descriptive text of the error
	Msg string
}

func (r *ErrorEvent) Destroy() {
		FfiDestroyerErrorLevel{}.Destroy(r.Level);
		FfiDestroyerErrorCode{}.Destroy(r.Code);
		FfiDestroyerString{}.Destroy(r.Msg);
}

type FfiConverterErrorEvent struct {}

var FfiConverterErrorEventINSTANCE = FfiConverterErrorEvent{}

func (c FfiConverterErrorEvent) Lift(rb RustBufferI) ErrorEvent {
	return LiftFromRustBuffer[ErrorEvent](c, rb)
}

func (c FfiConverterErrorEvent) Read(reader io.Reader) ErrorEvent {
	return ErrorEvent {
			FfiConverterErrorLevelINSTANCE.Read(reader),
			FfiConverterErrorCodeINSTANCE.Read(reader),
			FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterErrorEvent) Lower(value ErrorEvent) C.RustBuffer {
	return LowerIntoRustBuffer[ErrorEvent](c, value)
}

func (c FfiConverterErrorEvent) Write(writer io.Writer, value ErrorEvent) {
		FfiConverterErrorLevelINSTANCE.Write(writer, value.Level);
		FfiConverterErrorCodeINSTANCE.Write(writer, value.Code);
		FfiConverterStringINSTANCE.Write(writer, value.Msg);
}

type FfiDestroyerErrorEvent struct {}

func (_ FfiDestroyerErrorEvent) Destroy(value ErrorEvent) {
	value.Destroy()
}


type FeatureBatching struct {
	// direct connection threshold for batching
	DirectConnectionThreshold uint32
	// effective trigger period
	TriggerEffectiveDuration uint32
	// / cooldown after trigger was used
	TriggerCooldownDuration uint32
}

func (r *FeatureBatching) Destroy() {
		FfiDestroyerUint32{}.Destroy(r.DirectConnectionThreshold);
		FfiDestroyerUint32{}.Destroy(r.TriggerEffectiveDuration);
		FfiDestroyerUint32{}.Destroy(r.TriggerCooldownDuration);
}

type FfiConverterFeatureBatching struct {}

var FfiConverterFeatureBatchingINSTANCE = FfiConverterFeatureBatching{}

func (c FfiConverterFeatureBatching) Lift(rb RustBufferI) FeatureBatching {
	return LiftFromRustBuffer[FeatureBatching](c, rb)
}

func (c FfiConverterFeatureBatching) Read(reader io.Reader) FeatureBatching {
	return FeatureBatching {
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
	}
}

func (c FfiConverterFeatureBatching) Lower(value FeatureBatching) C.RustBuffer {
	return LowerIntoRustBuffer[FeatureBatching](c, value)
}

func (c FfiConverterFeatureBatching) Write(writer io.Writer, value FeatureBatching) {
		FfiConverterUint32INSTANCE.Write(writer, value.DirectConnectionThreshold);
		FfiConverterUint32INSTANCE.Write(writer, value.TriggerEffectiveDuration);
		FfiConverterUint32INSTANCE.Write(writer, value.TriggerCooldownDuration);
}

type FfiDestroyerFeatureBatching struct {}

func (_ FfiDestroyerFeatureBatching) Destroy(value FeatureBatching) {
	value.Destroy()
}


// Configure derp behaviour
type FeatureDerp struct {
	// Tcp keepalive set on derp server's side [default 15s]
	TcpKeepalive *uint32
	// Derp will send empty messages after this many seconds of not sending/receiving any data [default 60s]
	DerpKeepalive *uint32
	// Poll Keepalive: Application level keepalives meant to replace the TCP keepalives
	// They will reuse the derp_keepalive interval
	PollKeepalive *bool
	// Enable polling of remote peer states to reduce derp traffic
	EnablePolling *bool
	// Use Mozilla's root certificates instead of OS ones [default false]
	UseBuiltInRootCertificates bool
}

func (r *FeatureDerp) Destroy() {
		FfiDestroyerOptionalUint32{}.Destroy(r.TcpKeepalive);
		FfiDestroyerOptionalUint32{}.Destroy(r.DerpKeepalive);
		FfiDestroyerOptionalBool{}.Destroy(r.PollKeepalive);
		FfiDestroyerOptionalBool{}.Destroy(r.EnablePolling);
		FfiDestroyerBool{}.Destroy(r.UseBuiltInRootCertificates);
}

type FfiConverterFeatureDerp struct {}

var FfiConverterFeatureDerpINSTANCE = FfiConverterFeatureDerp{}

func (c FfiConverterFeatureDerp) Lift(rb RustBufferI) FeatureDerp {
	return LiftFromRustBuffer[FeatureDerp](c, rb)
}

func (c FfiConverterFeatureDerp) Read(reader io.Reader) FeatureDerp {
	return FeatureDerp {
			FfiConverterOptionalUint32INSTANCE.Read(reader),
			FfiConverterOptionalUint32INSTANCE.Read(reader),
			FfiConverterOptionalBoolINSTANCE.Read(reader),
			FfiConverterOptionalBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterFeatureDerp) Lower(value FeatureDerp) C.RustBuffer {
	return LowerIntoRustBuffer[FeatureDerp](c, value)
}

func (c FfiConverterFeatureDerp) Write(writer io.Writer, value FeatureDerp) {
		FfiConverterOptionalUint32INSTANCE.Write(writer, value.TcpKeepalive);
		FfiConverterOptionalUint32INSTANCE.Write(writer, value.DerpKeepalive);
		FfiConverterOptionalBoolINSTANCE.Write(writer, value.PollKeepalive);
		FfiConverterOptionalBoolINSTANCE.Write(writer, value.EnablePolling);
		FfiConverterBoolINSTANCE.Write(writer, value.UseBuiltInRootCertificates);
}

type FfiDestroyerFeatureDerp struct {}

func (_ FfiDestroyerFeatureDerp) Destroy(value FeatureDerp) {
	value.Destroy()
}


// Enable meshent direct connection
type FeatureDirect struct {
	// Endpoint providers [default all]
	Providers *EndpointProviders
	// Polling interval for endpoints [default 25s]
	EndpointIntervalSecs uint64
	// Configuration options for skipping unresponsive peers
	SkipUnresponsivePeers *FeatureSkipUnresponsivePeers
	// Parameters to optimize battery lifetime
	EndpointProvidersOptimization *FeatureEndpointProvidersOptimization
	// Configurable features for UPNP endpoint provider
	UpnpFeatures *FeatureUpnp
}

func (r *FeatureDirect) Destroy() {
		FfiDestroyerOptionalTypeEndpointProviders{}.Destroy(r.Providers);
		FfiDestroyerUint64{}.Destroy(r.EndpointIntervalSecs);
		FfiDestroyerOptionalFeatureSkipUnresponsivePeers{}.Destroy(r.SkipUnresponsivePeers);
		FfiDestroyerOptionalFeatureEndpointProvidersOptimization{}.Destroy(r.EndpointProvidersOptimization);
		FfiDestroyerOptionalFeatureUpnp{}.Destroy(r.UpnpFeatures);
}

type FfiConverterFeatureDirect struct {}

var FfiConverterFeatureDirectINSTANCE = FfiConverterFeatureDirect{}

func (c FfiConverterFeatureDirect) Lift(rb RustBufferI) FeatureDirect {
	return LiftFromRustBuffer[FeatureDirect](c, rb)
}

func (c FfiConverterFeatureDirect) Read(reader io.Reader) FeatureDirect {
	return FeatureDirect {
			FfiConverterOptionalTypeEndpointProvidersINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterOptionalFeatureSkipUnresponsivePeersINSTANCE.Read(reader),
			FfiConverterOptionalFeatureEndpointProvidersOptimizationINSTANCE.Read(reader),
			FfiConverterOptionalFeatureUpnpINSTANCE.Read(reader),
	}
}

func (c FfiConverterFeatureDirect) Lower(value FeatureDirect) C.RustBuffer {
	return LowerIntoRustBuffer[FeatureDirect](c, value)
}

func (c FfiConverterFeatureDirect) Write(writer io.Writer, value FeatureDirect) {
		FfiConverterOptionalTypeEndpointProvidersINSTANCE.Write(writer, value.Providers);
		FfiConverterUint64INSTANCE.Write(writer, value.EndpointIntervalSecs);
		FfiConverterOptionalFeatureSkipUnresponsivePeersINSTANCE.Write(writer, value.SkipUnresponsivePeers);
		FfiConverterOptionalFeatureEndpointProvidersOptimizationINSTANCE.Write(writer, value.EndpointProvidersOptimization);
		FfiConverterOptionalFeatureUpnpINSTANCE.Write(writer, value.UpnpFeatures);
}

type FfiDestroyerFeatureDirect struct {}

func (_ FfiDestroyerFeatureDirect) Destroy(value FeatureDirect) {
	value.Destroy()
}


// Feature configuration for DNS
type FeatureDns struct {
	// TTL for SOA record and for A and AAAA records [default 60s]
	TtlValue TtlValue
	// Configure options for exit dns [default None]
	ExitDns *FeatureExitDns
}

func (r *FeatureDns) Destroy() {
		FfiDestroyerTypeTtlValue{}.Destroy(r.TtlValue);
		FfiDestroyerOptionalFeatureExitDns{}.Destroy(r.ExitDns);
}

type FfiConverterFeatureDns struct {}

var FfiConverterFeatureDnsINSTANCE = FfiConverterFeatureDns{}

func (c FfiConverterFeatureDns) Lift(rb RustBufferI) FeatureDns {
	return LiftFromRustBuffer[FeatureDns](c, rb)
}

func (c FfiConverterFeatureDns) Read(reader io.Reader) FeatureDns {
	return FeatureDns {
			FfiConverterTypeTtlValueINSTANCE.Read(reader),
			FfiConverterOptionalFeatureExitDnsINSTANCE.Read(reader),
	}
}

func (c FfiConverterFeatureDns) Lower(value FeatureDns) C.RustBuffer {
	return LowerIntoRustBuffer[FeatureDns](c, value)
}

func (c FfiConverterFeatureDns) Write(writer io.Writer, value FeatureDns) {
		FfiConverterTypeTtlValueINSTANCE.Write(writer, value.TtlValue);
		FfiConverterOptionalFeatureExitDnsINSTANCE.Write(writer, value.ExitDns);
}

type FfiDestroyerFeatureDns struct {}

func (_ FfiDestroyerFeatureDns) Destroy(value FeatureDns) {
	value.Destroy()
}


// Control which battery optimizations are turned on
type FeatureEndpointProvidersOptimization struct {
	// Controls whether Stun endpoint provider should be turned off when there are no proxying peers
	OptimizeDirectUpgradeStun bool
	// Controls whether Upnp endpoint provider should be turned off when there are no proxying peers
	OptimizeDirectUpgradeUpnp bool
}

func (r *FeatureEndpointProvidersOptimization) Destroy() {
		FfiDestroyerBool{}.Destroy(r.OptimizeDirectUpgradeStun);
		FfiDestroyerBool{}.Destroy(r.OptimizeDirectUpgradeUpnp);
}

type FfiConverterFeatureEndpointProvidersOptimization struct {}

var FfiConverterFeatureEndpointProvidersOptimizationINSTANCE = FfiConverterFeatureEndpointProvidersOptimization{}

func (c FfiConverterFeatureEndpointProvidersOptimization) Lift(rb RustBufferI) FeatureEndpointProvidersOptimization {
	return LiftFromRustBuffer[FeatureEndpointProvidersOptimization](c, rb)
}

func (c FfiConverterFeatureEndpointProvidersOptimization) Read(reader io.Reader) FeatureEndpointProvidersOptimization {
	return FeatureEndpointProvidersOptimization {
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterFeatureEndpointProvidersOptimization) Lower(value FeatureEndpointProvidersOptimization) C.RustBuffer {
	return LowerIntoRustBuffer[FeatureEndpointProvidersOptimization](c, value)
}

func (c FfiConverterFeatureEndpointProvidersOptimization) Write(writer io.Writer, value FeatureEndpointProvidersOptimization) {
		FfiConverterBoolINSTANCE.Write(writer, value.OptimizeDirectUpgradeStun);
		FfiConverterBoolINSTANCE.Write(writer, value.OptimizeDirectUpgradeUpnp);
}

type FfiDestroyerFeatureEndpointProvidersOptimization struct {}

func (_ FfiDestroyerFeatureEndpointProvidersOptimization) Destroy(value FeatureEndpointProvidersOptimization) {
	value.Destroy()
}


// Configuration for the Error Notification Service
type FeatureErrorNotificationService struct {
	// Size of the internal queue of received and to-be-published vpn error notifications
	BufferSize uint32
	// Allow only post-quantum safe key exchange algorithm for the ENS HTTPS connection
	AllowOnlyPq bool
}

func (r *FeatureErrorNotificationService) Destroy() {
		FfiDestroyerUint32{}.Destroy(r.BufferSize);
		FfiDestroyerBool{}.Destroy(r.AllowOnlyPq);
}

type FfiConverterFeatureErrorNotificationService struct {}

var FfiConverterFeatureErrorNotificationServiceINSTANCE = FfiConverterFeatureErrorNotificationService{}

func (c FfiConverterFeatureErrorNotificationService) Lift(rb RustBufferI) FeatureErrorNotificationService {
	return LiftFromRustBuffer[FeatureErrorNotificationService](c, rb)
}

func (c FfiConverterFeatureErrorNotificationService) Read(reader io.Reader) FeatureErrorNotificationService {
	return FeatureErrorNotificationService {
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterFeatureErrorNotificationService) Lower(value FeatureErrorNotificationService) C.RustBuffer {
	return LowerIntoRustBuffer[FeatureErrorNotificationService](c, value)
}

func (c FfiConverterFeatureErrorNotificationService) Write(writer io.Writer, value FeatureErrorNotificationService) {
		FfiConverterUint32INSTANCE.Write(writer, value.BufferSize);
		FfiConverterBoolINSTANCE.Write(writer, value.AllowOnlyPq);
}

type FfiDestroyerFeatureErrorNotificationService struct {}

func (_ FfiDestroyerFeatureErrorNotificationService) Destroy(value FeatureErrorNotificationService) {
	value.Destroy()
}


// Configurable features for exit Dns
type FeatureExitDns struct {
	// Controls if it is allowed to reconfigure DNS peer when exit node is
	// (dis)connected.
	AutoSwitchDnsIps *bool
}

func (r *FeatureExitDns) Destroy() {
		FfiDestroyerOptionalBool{}.Destroy(r.AutoSwitchDnsIps);
}

type FfiConverterFeatureExitDns struct {}

var FfiConverterFeatureExitDnsINSTANCE = FfiConverterFeatureExitDns{}

func (c FfiConverterFeatureExitDns) Lift(rb RustBufferI) FeatureExitDns {
	return LiftFromRustBuffer[FeatureExitDns](c, rb)
}

func (c FfiConverterFeatureExitDns) Read(reader io.Reader) FeatureExitDns {
	return FeatureExitDns {
			FfiConverterOptionalBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterFeatureExitDns) Lower(value FeatureExitDns) C.RustBuffer {
	return LowerIntoRustBuffer[FeatureExitDns](c, value)
}

func (c FfiConverterFeatureExitDns) Write(writer io.Writer, value FeatureExitDns) {
		FfiConverterOptionalBoolINSTANCE.Write(writer, value.AutoSwitchDnsIps);
}

type FfiDestroyerFeatureExitDns struct {}

func (_ FfiDestroyerFeatureExitDns) Destroy(value FeatureExitDns) {
	value.Destroy()
}


// Feature config for firewall
type FeatureFirewall struct {
	// Turns on connection resets upon VPN server change
	NeptunResetConns bool
	// Turns on connection resets upon VPN server change (Deprecated alias for neptun_reset_conns)
	BoringtunResetConns bool
	// Ip range from RFC1918 to exclude from firewall blocking
	ExcludePrivateIpRange *Ipv4Net
	// Blackist for outgoing connections
	OutgoingBlacklist []FirewallBlacklistTuple
}

func (r *FeatureFirewall) Destroy() {
		FfiDestroyerBool{}.Destroy(r.NeptunResetConns);
		FfiDestroyerBool{}.Destroy(r.BoringtunResetConns);
		FfiDestroyerOptionalTypeIpv4Net{}.Destroy(r.ExcludePrivateIpRange);
		FfiDestroyerSequenceFirewallBlacklistTuple{}.Destroy(r.OutgoingBlacklist);
}

type FfiConverterFeatureFirewall struct {}

var FfiConverterFeatureFirewallINSTANCE = FfiConverterFeatureFirewall{}

func (c FfiConverterFeatureFirewall) Lift(rb RustBufferI) FeatureFirewall {
	return LiftFromRustBuffer[FeatureFirewall](c, rb)
}

func (c FfiConverterFeatureFirewall) Read(reader io.Reader) FeatureFirewall {
	return FeatureFirewall {
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterOptionalTypeIpv4NetINSTANCE.Read(reader),
			FfiConverterSequenceFirewallBlacklistTupleINSTANCE.Read(reader),
	}
}

func (c FfiConverterFeatureFirewall) Lower(value FeatureFirewall) C.RustBuffer {
	return LowerIntoRustBuffer[FeatureFirewall](c, value)
}

func (c FfiConverterFeatureFirewall) Write(writer io.Writer, value FeatureFirewall) {
		FfiConverterBoolINSTANCE.Write(writer, value.NeptunResetConns);
		FfiConverterBoolINSTANCE.Write(writer, value.BoringtunResetConns);
		FfiConverterOptionalTypeIpv4NetINSTANCE.Write(writer, value.ExcludePrivateIpRange);
		FfiConverterSequenceFirewallBlacklistTupleINSTANCE.Write(writer, value.OutgoingBlacklist);
}

type FfiDestroyerFeatureFirewall struct {}

func (_ FfiDestroyerFeatureFirewall) Destroy(value FeatureFirewall) {
	value.Destroy()
}


// Configurable features for Lana module
type FeatureLana struct {
	// Path of the file where events will be stored. If such file does not exist, it will be created, otherwise reused
	EventPath string
	// Whether the events should be sent to produciton or not
	Prod bool
}

func (r *FeatureLana) Destroy() {
		FfiDestroyerString{}.Destroy(r.EventPath);
		FfiDestroyerBool{}.Destroy(r.Prod);
}

type FfiConverterFeatureLana struct {}

var FfiConverterFeatureLanaINSTANCE = FfiConverterFeatureLana{}

func (c FfiConverterFeatureLana) Lift(rb RustBufferI) FeatureLana {
	return LiftFromRustBuffer[FeatureLana](c, rb)
}

func (c FfiConverterFeatureLana) Read(reader io.Reader) FeatureLana {
	return FeatureLana {
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterFeatureLana) Lower(value FeatureLana) C.RustBuffer {
	return LowerIntoRustBuffer[FeatureLana](c, value)
}

func (c FfiConverterFeatureLana) Write(writer io.Writer, value FeatureLana) {
		FfiConverterStringINSTANCE.Write(writer, value.EventPath);
		FfiConverterBoolINSTANCE.Write(writer, value.Prod);
}

type FfiDestroyerFeatureLana struct {}

func (_ FfiDestroyerFeatureLana) Destroy(value FeatureLana) {
	value.Destroy()
}


// Link detection mechanism
type FeatureLinkDetection struct {
	// Configurable rtt in seconds
	RttSeconds uint64
	// Check the link state before reporting it as down
	NoOfPings uint32
	// Use link detection for downgrade logic
	UseForDowngrade bool
}

func (r *FeatureLinkDetection) Destroy() {
		FfiDestroyerUint64{}.Destroy(r.RttSeconds);
		FfiDestroyerUint32{}.Destroy(r.NoOfPings);
		FfiDestroyerBool{}.Destroy(r.UseForDowngrade);
}

type FfiConverterFeatureLinkDetection struct {}

var FfiConverterFeatureLinkDetectionINSTANCE = FfiConverterFeatureLinkDetection{}

func (c FfiConverterFeatureLinkDetection) Lift(rb RustBufferI) FeatureLinkDetection {
	return LiftFromRustBuffer[FeatureLinkDetection](c, rb)
}

func (c FfiConverterFeatureLinkDetection) Read(reader io.Reader) FeatureLinkDetection {
	return FeatureLinkDetection {
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterFeatureLinkDetection) Lower(value FeatureLinkDetection) C.RustBuffer {
	return LowerIntoRustBuffer[FeatureLinkDetection](c, value)
}

func (c FfiConverterFeatureLinkDetection) Write(writer io.Writer, value FeatureLinkDetection) {
		FfiConverterUint64INSTANCE.Write(writer, value.RttSeconds);
		FfiConverterUint32INSTANCE.Write(writer, value.NoOfPings);
		FfiConverterBoolINSTANCE.Write(writer, value.UseForDowngrade);
}

type FfiDestroyerFeatureLinkDetection struct {}

func (_ FfiDestroyerFeatureLinkDetection) Destroy(value FeatureLinkDetection) {
	value.Destroy()
}


// Configurable features for Nurse module
type FeatureNurse struct {
	// Heartbeat interval in seconds. Default value is 3600.
	HeartbeatInterval uint64
	// Initial heartbeat interval in seconds. Default value is None.
	InitialHeartbeatInterval uint64
	// QoS configuration for Nurse
	Qos *FeatureQoS
	// Enable/disable Relay connection data
	EnableRelayConnData bool
	// Enable/disable NAT-traversal connections data
	EnableNatTraversalConnData bool
	// How long a session can exist before it is forcibly reported, in seconds. Default value is 24h.
	StateDurationCap uint64
}

func (r *FeatureNurse) Destroy() {
		FfiDestroyerUint64{}.Destroy(r.HeartbeatInterval);
		FfiDestroyerUint64{}.Destroy(r.InitialHeartbeatInterval);
		FfiDestroyerOptionalFeatureQoS{}.Destroy(r.Qos);
		FfiDestroyerBool{}.Destroy(r.EnableRelayConnData);
		FfiDestroyerBool{}.Destroy(r.EnableNatTraversalConnData);
		FfiDestroyerUint64{}.Destroy(r.StateDurationCap);
}

type FfiConverterFeatureNurse struct {}

var FfiConverterFeatureNurseINSTANCE = FfiConverterFeatureNurse{}

func (c FfiConverterFeatureNurse) Lift(rb RustBufferI) FeatureNurse {
	return LiftFromRustBuffer[FeatureNurse](c, rb)
}

func (c FfiConverterFeatureNurse) Read(reader io.Reader) FeatureNurse {
	return FeatureNurse {
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterOptionalFeatureQoSINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
	}
}

func (c FfiConverterFeatureNurse) Lower(value FeatureNurse) C.RustBuffer {
	return LowerIntoRustBuffer[FeatureNurse](c, value)
}

func (c FfiConverterFeatureNurse) Write(writer io.Writer, value FeatureNurse) {
		FfiConverterUint64INSTANCE.Write(writer, value.HeartbeatInterval);
		FfiConverterUint64INSTANCE.Write(writer, value.InitialHeartbeatInterval);
		FfiConverterOptionalFeatureQoSINSTANCE.Write(writer, value.Qos);
		FfiConverterBoolINSTANCE.Write(writer, value.EnableRelayConnData);
		FfiConverterBoolINSTANCE.Write(writer, value.EnableNatTraversalConnData);
		FfiConverterUint64INSTANCE.Write(writer, value.StateDurationCap);
}

type FfiDestroyerFeatureNurse struct {}

func (_ FfiDestroyerFeatureNurse) Destroy(value FeatureNurse) {
	value.Destroy()
}


// Enable wanted paths for telio
type FeaturePaths struct {
	// Enable paths in increasing priority: 0 is worse then 1 is worse then 2 ...
	// [PathType::Relay] always assumed as -1
	Priority []PathType
	// Force only one specific path to be used.
	Force *PathType
}

func (r *FeaturePaths) Destroy() {
		FfiDestroyerSequencePathType{}.Destroy(r.Priority);
		FfiDestroyerOptionalPathType{}.Destroy(r.Force);
}

type FfiConverterFeaturePaths struct {}

var FfiConverterFeaturePathsINSTANCE = FfiConverterFeaturePaths{}

func (c FfiConverterFeaturePaths) Lift(rb RustBufferI) FeaturePaths {
	return LiftFromRustBuffer[FeaturePaths](c, rb)
}

func (c FfiConverterFeaturePaths) Read(reader io.Reader) FeaturePaths {
	return FeaturePaths {
			FfiConverterSequencePathTypeINSTANCE.Read(reader),
			FfiConverterOptionalPathTypeINSTANCE.Read(reader),
	}
}

func (c FfiConverterFeaturePaths) Lower(value FeaturePaths) C.RustBuffer {
	return LowerIntoRustBuffer[FeaturePaths](c, value)
}

func (c FfiConverterFeaturePaths) Write(writer io.Writer, value FeaturePaths) {
		FfiConverterSequencePathTypeINSTANCE.Write(writer, value.Priority);
		FfiConverterOptionalPathTypeINSTANCE.Write(writer, value.Force);
}

type FfiDestroyerFeaturePaths struct {}

func (_ FfiDestroyerFeaturePaths) Destroy(value FeaturePaths) {
	value.Destroy()
}


// Configurable persistent keepalive periods for different types of peers
type FeaturePersistentKeepalive struct {
	// Persistent keepalive period given for VPN peers (in seconds) [default 15s]
	Vpn *uint32
	// Persistent keepalive period for direct peers (in seconds) [default 5s]
	Direct uint32
	// Persistent keepalive period for proxying peers (in seconds) [default 25s]
	Proxying *uint32
	// Persistent keepalive period for stun peers (in seconds) [default 25s]
	Stun *uint32
}

func (r *FeaturePersistentKeepalive) Destroy() {
		FfiDestroyerOptionalUint32{}.Destroy(r.Vpn);
		FfiDestroyerUint32{}.Destroy(r.Direct);
		FfiDestroyerOptionalUint32{}.Destroy(r.Proxying);
		FfiDestroyerOptionalUint32{}.Destroy(r.Stun);
}

type FfiConverterFeaturePersistentKeepalive struct {}

var FfiConverterFeaturePersistentKeepaliveINSTANCE = FfiConverterFeaturePersistentKeepalive{}

func (c FfiConverterFeaturePersistentKeepalive) Lift(rb RustBufferI) FeaturePersistentKeepalive {
	return LiftFromRustBuffer[FeaturePersistentKeepalive](c, rb)
}

func (c FfiConverterFeaturePersistentKeepalive) Read(reader io.Reader) FeaturePersistentKeepalive {
	return FeaturePersistentKeepalive {
			FfiConverterOptionalUint32INSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterOptionalUint32INSTANCE.Read(reader),
			FfiConverterOptionalUint32INSTANCE.Read(reader),
	}
}

func (c FfiConverterFeaturePersistentKeepalive) Lower(value FeaturePersistentKeepalive) C.RustBuffer {
	return LowerIntoRustBuffer[FeaturePersistentKeepalive](c, value)
}

func (c FfiConverterFeaturePersistentKeepalive) Write(writer io.Writer, value FeaturePersistentKeepalive) {
		FfiConverterOptionalUint32INSTANCE.Write(writer, value.Vpn);
		FfiConverterUint32INSTANCE.Write(writer, value.Direct);
		FfiConverterOptionalUint32INSTANCE.Write(writer, value.Proxying);
		FfiConverterOptionalUint32INSTANCE.Write(writer, value.Stun);
}

type FfiDestroyerFeaturePersistentKeepalive struct {}

func (_ FfiDestroyerFeaturePersistentKeepalive) Destroy(value FeaturePersistentKeepalive) {
	value.Destroy()
}


// Configurable WireGuard polling periods
type FeaturePolling struct {
	// Wireguard state polling period (in milliseconds) [default 1000ms]
	WireguardPollingPeriod uint32
	// Wireguard state polling period after state change (in milliseconds) [default 50ms]
	WireguardPollingPeriodAfterStateChange uint32
}

func (r *FeaturePolling) Destroy() {
		FfiDestroyerUint32{}.Destroy(r.WireguardPollingPeriod);
		FfiDestroyerUint32{}.Destroy(r.WireguardPollingPeriodAfterStateChange);
}

type FfiConverterFeaturePolling struct {}

var FfiConverterFeaturePollingINSTANCE = FfiConverterFeaturePolling{}

func (c FfiConverterFeaturePolling) Lift(rb RustBufferI) FeaturePolling {
	return LiftFromRustBuffer[FeaturePolling](c, rb)
}

func (c FfiConverterFeaturePolling) Read(reader io.Reader) FeaturePolling {
	return FeaturePolling {
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
	}
}

func (c FfiConverterFeaturePolling) Lower(value FeaturePolling) C.RustBuffer {
	return LowerIntoRustBuffer[FeaturePolling](c, value)
}

func (c FfiConverterFeaturePolling) Write(writer io.Writer, value FeaturePolling) {
		FfiConverterUint32INSTANCE.Write(writer, value.WireguardPollingPeriod);
		FfiConverterUint32INSTANCE.Write(writer, value.WireguardPollingPeriodAfterStateChange);
}

type FfiDestroyerFeaturePolling struct {}

func (_ FfiDestroyerFeaturePolling) Destroy(value FeaturePolling) {
	value.Destroy()
}


// Turns on post quantum VPN tunnel
type FeaturePostQuantumVpn struct {
	// Initial handshake retry interval in seconds
	HandshakeRetryIntervalS uint32
	// Rekey interval in seconds
	RekeyIntervalS uint32
	// Post-quantum protocol version
	Version uint32
}

func (r *FeaturePostQuantumVpn) Destroy() {
		FfiDestroyerUint32{}.Destroy(r.HandshakeRetryIntervalS);
		FfiDestroyerUint32{}.Destroy(r.RekeyIntervalS);
		FfiDestroyerUint32{}.Destroy(r.Version);
}

type FfiConverterFeaturePostQuantumVpn struct {}

var FfiConverterFeaturePostQuantumVpnINSTANCE = FfiConverterFeaturePostQuantumVpn{}

func (c FfiConverterFeaturePostQuantumVpn) Lift(rb RustBufferI) FeaturePostQuantumVpn {
	return LiftFromRustBuffer[FeaturePostQuantumVpn](c, rb)
}

func (c FfiConverterFeaturePostQuantumVpn) Read(reader io.Reader) FeaturePostQuantumVpn {
	return FeaturePostQuantumVpn {
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
	}
}

func (c FfiConverterFeaturePostQuantumVpn) Lower(value FeaturePostQuantumVpn) C.RustBuffer {
	return LowerIntoRustBuffer[FeaturePostQuantumVpn](c, value)
}

func (c FfiConverterFeaturePostQuantumVpn) Write(writer io.Writer, value FeaturePostQuantumVpn) {
		FfiConverterUint32INSTANCE.Write(writer, value.HandshakeRetryIntervalS);
		FfiConverterUint32INSTANCE.Write(writer, value.RekeyIntervalS);
		FfiConverterUint32INSTANCE.Write(writer, value.Version);
}

type FfiDestroyerFeaturePostQuantumVpn struct {}

func (_ FfiDestroyerFeaturePostQuantumVpn) Destroy(value FeaturePostQuantumVpn) {
	value.Destroy()
}


// QoS configuration options
type FeatureQoS struct {
	// How often to collect rtt data in seconds. Default value is 300.
	RttInterval uint64
	// Number of tries for each node. Default value is 3.
	RttTries uint32
	// Types of rtt analytics. Default is Ping.
	RttTypes []RttType
	// Number of buckets used for rtt and throughput. Default value is 5.
	Buckets uint32
}

func (r *FeatureQoS) Destroy() {
		FfiDestroyerUint64{}.Destroy(r.RttInterval);
		FfiDestroyerUint32{}.Destroy(r.RttTries);
		FfiDestroyerSequenceRttType{}.Destroy(r.RttTypes);
		FfiDestroyerUint32{}.Destroy(r.Buckets);
}

type FfiConverterFeatureQoS struct {}

var FfiConverterFeatureQoSINSTANCE = FfiConverterFeatureQoS{}

func (c FfiConverterFeatureQoS) Lift(rb RustBufferI) FeatureQoS {
	return LiftFromRustBuffer[FeatureQoS](c, rb)
}

func (c FfiConverterFeatureQoS) Read(reader io.Reader) FeatureQoS {
	return FeatureQoS {
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterSequenceRttTypeINSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
	}
}

func (c FfiConverterFeatureQoS) Lower(value FeatureQoS) C.RustBuffer {
	return LowerIntoRustBuffer[FeatureQoS](c, value)
}

func (c FfiConverterFeatureQoS) Write(writer io.Writer, value FeatureQoS) {
		FfiConverterUint64INSTANCE.Write(writer, value.RttInterval);
		FfiConverterUint32INSTANCE.Write(writer, value.RttTries);
		FfiConverterSequenceRttTypeINSTANCE.Write(writer, value.RttTypes);
		FfiConverterUint32INSTANCE.Write(writer, value.Buckets);
}

type FfiDestroyerFeatureQoS struct {}

func (_ FfiDestroyerFeatureQoS) Destroy(value FeatureQoS) {
	value.Destroy()
}


// Avoid sending periodic messages to peers with no traffic reported by wireguard
type FeatureSkipUnresponsivePeers struct {
	// Time after which peers is considered unresponsive if it didn't receive any packets
	NoRxThresholdSecs uint64
}

func (r *FeatureSkipUnresponsivePeers) Destroy() {
		FfiDestroyerUint64{}.Destroy(r.NoRxThresholdSecs);
}

type FfiConverterFeatureSkipUnresponsivePeers struct {}

var FfiConverterFeatureSkipUnresponsivePeersINSTANCE = FfiConverterFeatureSkipUnresponsivePeers{}

func (c FfiConverterFeatureSkipUnresponsivePeers) Lift(rb RustBufferI) FeatureSkipUnresponsivePeers {
	return LiftFromRustBuffer[FeatureSkipUnresponsivePeers](c, rb)
}

func (c FfiConverterFeatureSkipUnresponsivePeers) Read(reader io.Reader) FeatureSkipUnresponsivePeers {
	return FeatureSkipUnresponsivePeers {
			FfiConverterUint64INSTANCE.Read(reader),
	}
}

func (c FfiConverterFeatureSkipUnresponsivePeers) Lower(value FeatureSkipUnresponsivePeers) C.RustBuffer {
	return LowerIntoRustBuffer[FeatureSkipUnresponsivePeers](c, value)
}

func (c FfiConverterFeatureSkipUnresponsivePeers) Write(writer io.Writer, value FeatureSkipUnresponsivePeers) {
		FfiConverterUint64INSTANCE.Write(writer, value.NoRxThresholdSecs);
}

type FfiDestroyerFeatureSkipUnresponsivePeers struct {}

func (_ FfiDestroyerFeatureSkipUnresponsivePeers) Destroy(value FeatureSkipUnresponsivePeers) {
	value.Destroy()
}


// Configurable features for UPNP endpoint provider
type FeatureUpnp struct {
	// The upnp lease_duration parameter, in seconds. A value of 0 is infinite.
	LeaseDurationS uint32
}

func (r *FeatureUpnp) Destroy() {
		FfiDestroyerUint32{}.Destroy(r.LeaseDurationS);
}

type FfiConverterFeatureUpnp struct {}

var FfiConverterFeatureUpnpINSTANCE = FfiConverterFeatureUpnp{}

func (c FfiConverterFeatureUpnp) Lift(rb RustBufferI) FeatureUpnp {
	return LiftFromRustBuffer[FeatureUpnp](c, rb)
}

func (c FfiConverterFeatureUpnp) Read(reader io.Reader) FeatureUpnp {
	return FeatureUpnp {
			FfiConverterUint32INSTANCE.Read(reader),
	}
}

func (c FfiConverterFeatureUpnp) Lower(value FeatureUpnp) C.RustBuffer {
	return LowerIntoRustBuffer[FeatureUpnp](c, value)
}

func (c FfiConverterFeatureUpnp) Write(writer io.Writer, value FeatureUpnp) {
		FfiConverterUint32INSTANCE.Write(writer, value.LeaseDurationS);
}

type FfiDestroyerFeatureUpnp struct {}

func (_ FfiDestroyerFeatureUpnp) Destroy(value FeatureUpnp) {
	value.Destroy()
}


// Configurable features for Wireguard peers
type FeatureWireguard struct {
	// Configurable persistent keepalive periods for wireguard peers
	PersistentKeepalive FeaturePersistentKeepalive
	// Configurable WireGuard polling periods
	Polling FeaturePolling
	// Configurable up/down behavior of WireGuard-NT adapter. See RFC LLT-0089 for details
	EnableDynamicWgNtControl bool
	// Configurable socket buffer size for NepTUN
	SktBufferSize *uint32
	// Configurable socket buffer size for NepTUN
	InterThreadChannelSize *uint32
	// Configurable socket buffer size for NepTUN
	MaxInterThreadBatchedPkts *uint32
}

func (r *FeatureWireguard) Destroy() {
		FfiDestroyerFeaturePersistentKeepalive{}.Destroy(r.PersistentKeepalive);
		FfiDestroyerFeaturePolling{}.Destroy(r.Polling);
		FfiDestroyerBool{}.Destroy(r.EnableDynamicWgNtControl);
		FfiDestroyerOptionalUint32{}.Destroy(r.SktBufferSize);
		FfiDestroyerOptionalUint32{}.Destroy(r.InterThreadChannelSize);
		FfiDestroyerOptionalUint32{}.Destroy(r.MaxInterThreadBatchedPkts);
}

type FfiConverterFeatureWireguard struct {}

var FfiConverterFeatureWireguardINSTANCE = FfiConverterFeatureWireguard{}

func (c FfiConverterFeatureWireguard) Lift(rb RustBufferI) FeatureWireguard {
	return LiftFromRustBuffer[FeatureWireguard](c, rb)
}

func (c FfiConverterFeatureWireguard) Read(reader io.Reader) FeatureWireguard {
	return FeatureWireguard {
			FfiConverterFeaturePersistentKeepaliveINSTANCE.Read(reader),
			FfiConverterFeaturePollingINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterOptionalUint32INSTANCE.Read(reader),
			FfiConverterOptionalUint32INSTANCE.Read(reader),
			FfiConverterOptionalUint32INSTANCE.Read(reader),
	}
}

func (c FfiConverterFeatureWireguard) Lower(value FeatureWireguard) C.RustBuffer {
	return LowerIntoRustBuffer[FeatureWireguard](c, value)
}

func (c FfiConverterFeatureWireguard) Write(writer io.Writer, value FeatureWireguard) {
		FfiConverterFeaturePersistentKeepaliveINSTANCE.Write(writer, value.PersistentKeepalive);
		FfiConverterFeaturePollingINSTANCE.Write(writer, value.Polling);
		FfiConverterBoolINSTANCE.Write(writer, value.EnableDynamicWgNtControl);
		FfiConverterOptionalUint32INSTANCE.Write(writer, value.SktBufferSize);
		FfiConverterOptionalUint32INSTANCE.Write(writer, value.InterThreadChannelSize);
		FfiConverterOptionalUint32INSTANCE.Write(writer, value.MaxInterThreadBatchedPkts);
}

type FfiDestroyerFeatureWireguard struct {}

func (_ FfiDestroyerFeatureWireguard) Destroy(value FeatureWireguard) {
	value.Destroy()
}


// Encompasses all of the possible features that can be enabled
type Features struct {
	// Additional wireguard configuration
	Wireguard FeatureWireguard
	// Nurse features that can be configured for QoS
	Nurse *FeatureNurse
	// Event logging configurable features
	Lana *FeatureLana
	// Deprecated by direct since 4.0.0
	Paths *FeaturePaths
	// Configure options for direct WG connections
	Direct *FeatureDirect
	// Should only be set for macos sideload
	IsTestEnv *bool
	// Control if IP addresses and domains should be hidden in logs
	HideUserData bool
	// Control if thread IDs should be shown in the logs
	HideThreadId bool
	// Derp server specific configuration
	Derp *FeatureDerp
	// Flag to specify if keys should be validated
	ValidateKeys FeatureValidateKeys
	// IPv6 support
	Ipv6 bool
	// Nicknames support
	Nicknames bool
	// Feature config for firewall
	Firewall FeatureFirewall
	// If and for how long to flush events when stopping telio. Setting to Some(0) means waiting until all events have been flushed, regardless of how long it takes
	FlushEventsOnStopTimeoutSeconds *uint64
	// Link detection mechanism
	LinkDetection *FeatureLinkDetection
	// Feature configuration for DNS
	Dns FeatureDns
	// Post quantum VPN tunnel configuration
	PostQuantumVpn FeaturePostQuantumVpn
	// Multicast support
	Multicast bool
	// Batching
	Batching *FeatureBatching
	ErrorNotificationService *FeatureErrorNotificationService
}

func (r *Features) Destroy() {
		FfiDestroyerFeatureWireguard{}.Destroy(r.Wireguard);
		FfiDestroyerOptionalFeatureNurse{}.Destroy(r.Nurse);
		FfiDestroyerOptionalFeatureLana{}.Destroy(r.Lana);
		FfiDestroyerOptionalFeaturePaths{}.Destroy(r.Paths);
		FfiDestroyerOptionalFeatureDirect{}.Destroy(r.Direct);
		FfiDestroyerOptionalBool{}.Destroy(r.IsTestEnv);
		FfiDestroyerBool{}.Destroy(r.HideUserData);
		FfiDestroyerBool{}.Destroy(r.HideThreadId);
		FfiDestroyerOptionalFeatureDerp{}.Destroy(r.Derp);
		FfiDestroyerTypeFeatureValidateKeys{}.Destroy(r.ValidateKeys);
		FfiDestroyerBool{}.Destroy(r.Ipv6);
		FfiDestroyerBool{}.Destroy(r.Nicknames);
		FfiDestroyerFeatureFirewall{}.Destroy(r.Firewall);
		FfiDestroyerOptionalUint64{}.Destroy(r.FlushEventsOnStopTimeoutSeconds);
		FfiDestroyerOptionalFeatureLinkDetection{}.Destroy(r.LinkDetection);
		FfiDestroyerFeatureDns{}.Destroy(r.Dns);
		FfiDestroyerFeaturePostQuantumVpn{}.Destroy(r.PostQuantumVpn);
		FfiDestroyerBool{}.Destroy(r.Multicast);
		FfiDestroyerOptionalFeatureBatching{}.Destroy(r.Batching);
		FfiDestroyerOptionalFeatureErrorNotificationService{}.Destroy(r.ErrorNotificationService);
}

type FfiConverterFeatures struct {}

var FfiConverterFeaturesINSTANCE = FfiConverterFeatures{}

func (c FfiConverterFeatures) Lift(rb RustBufferI) Features {
	return LiftFromRustBuffer[Features](c, rb)
}

func (c FfiConverterFeatures) Read(reader io.Reader) Features {
	return Features {
			FfiConverterFeatureWireguardINSTANCE.Read(reader),
			FfiConverterOptionalFeatureNurseINSTANCE.Read(reader),
			FfiConverterOptionalFeatureLanaINSTANCE.Read(reader),
			FfiConverterOptionalFeaturePathsINSTANCE.Read(reader),
			FfiConverterOptionalFeatureDirectINSTANCE.Read(reader),
			FfiConverterOptionalBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterOptionalFeatureDerpINSTANCE.Read(reader),
			FfiConverterTypeFeatureValidateKeysINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterFeatureFirewallINSTANCE.Read(reader),
			FfiConverterOptionalUint64INSTANCE.Read(reader),
			FfiConverterOptionalFeatureLinkDetectionINSTANCE.Read(reader),
			FfiConverterFeatureDnsINSTANCE.Read(reader),
			FfiConverterFeaturePostQuantumVpnINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterOptionalFeatureBatchingINSTANCE.Read(reader),
			FfiConverterOptionalFeatureErrorNotificationServiceINSTANCE.Read(reader),
	}
}

func (c FfiConverterFeatures) Lower(value Features) C.RustBuffer {
	return LowerIntoRustBuffer[Features](c, value)
}

func (c FfiConverterFeatures) Write(writer io.Writer, value Features) {
		FfiConverterFeatureWireguardINSTANCE.Write(writer, value.Wireguard);
		FfiConverterOptionalFeatureNurseINSTANCE.Write(writer, value.Nurse);
		FfiConverterOptionalFeatureLanaINSTANCE.Write(writer, value.Lana);
		FfiConverterOptionalFeaturePathsINSTANCE.Write(writer, value.Paths);
		FfiConverterOptionalFeatureDirectINSTANCE.Write(writer, value.Direct);
		FfiConverterOptionalBoolINSTANCE.Write(writer, value.IsTestEnv);
		FfiConverterBoolINSTANCE.Write(writer, value.HideUserData);
		FfiConverterBoolINSTANCE.Write(writer, value.HideThreadId);
		FfiConverterOptionalFeatureDerpINSTANCE.Write(writer, value.Derp);
		FfiConverterTypeFeatureValidateKeysINSTANCE.Write(writer, value.ValidateKeys);
		FfiConverterBoolINSTANCE.Write(writer, value.Ipv6);
		FfiConverterBoolINSTANCE.Write(writer, value.Nicknames);
		FfiConverterFeatureFirewallINSTANCE.Write(writer, value.Firewall);
		FfiConverterOptionalUint64INSTANCE.Write(writer, value.FlushEventsOnStopTimeoutSeconds);
		FfiConverterOptionalFeatureLinkDetectionINSTANCE.Write(writer, value.LinkDetection);
		FfiConverterFeatureDnsINSTANCE.Write(writer, value.Dns);
		FfiConverterFeaturePostQuantumVpnINSTANCE.Write(writer, value.PostQuantumVpn);
		FfiConverterBoolINSTANCE.Write(writer, value.Multicast);
		FfiConverterOptionalFeatureBatchingINSTANCE.Write(writer, value.Batching);
		FfiConverterOptionalFeatureErrorNotificationServiceINSTANCE.Write(writer, value.ErrorNotificationService);
}

type FfiDestroyerFeatures struct {}

func (_ FfiDestroyerFeatures) Destroy(value Features) {
	value.Destroy()
}


// Tuple used to blacklist outgoing connections in Telio firewall
type FirewallBlacklistTuple struct {
	// Protocol of the packet to be blacklisted
	Protocol IpProtocol
	// Destination IP address of the packet
	Ip IpAddr
	// Destination port of the packet
	Port uint16
}

func (r *FirewallBlacklistTuple) Destroy() {
		FfiDestroyerIpProtocol{}.Destroy(r.Protocol);
		FfiDestroyerTypeIpAddr{}.Destroy(r.Ip);
		FfiDestroyerUint16{}.Destroy(r.Port);
}

type FfiConverterFirewallBlacklistTuple struct {}

var FfiConverterFirewallBlacklistTupleINSTANCE = FfiConverterFirewallBlacklistTuple{}

func (c FfiConverterFirewallBlacklistTuple) Lift(rb RustBufferI) FirewallBlacklistTuple {
	return LiftFromRustBuffer[FirewallBlacklistTuple](c, rb)
}

func (c FfiConverterFirewallBlacklistTuple) Read(reader io.Reader) FirewallBlacklistTuple {
	return FirewallBlacklistTuple {
			FfiConverterIpProtocolINSTANCE.Read(reader),
			FfiConverterTypeIpAddrINSTANCE.Read(reader),
			FfiConverterUint16INSTANCE.Read(reader),
	}
}

func (c FfiConverterFirewallBlacklistTuple) Lower(value FirewallBlacklistTuple) C.RustBuffer {
	return LowerIntoRustBuffer[FirewallBlacklistTuple](c, value)
}

func (c FfiConverterFirewallBlacklistTuple) Write(writer io.Writer, value FirewallBlacklistTuple) {
		FfiConverterIpProtocolINSTANCE.Write(writer, value.Protocol);
		FfiConverterTypeIpAddrINSTANCE.Write(writer, value.Ip);
		FfiConverterUint16INSTANCE.Write(writer, value.Port);
}

type FfiDestroyerFirewallBlacklistTuple struct {}

func (_ FfiDestroyerFirewallBlacklistTuple) Destroy(value FirewallBlacklistTuple) {
	value.Destroy()
}


// Description of a peer
type Peer struct {
	// The base object describing a peer
	Base PeerBase
	// The peer is local, when the flag is set
	IsLocal bool
	// Flag to control whether the peer allows incoming connections
	AllowIncomingConnections bool
	// Flag to control whether the Node allows routing through
	AllowPeerTrafficRouting bool
	// Flag to control whether the Node allows incoming local area access
	AllowPeerLocalNetworkAccess bool
	// Flag to control whether the peer allows incoming files
	AllowPeerSendFiles bool
	// Flag to control whether we allow multicast messages from the peer
	AllowMulticast bool
	// Flag to control whether the peer allows multicast messages from us
	PeerAllowsMulticast bool
}

func (r *Peer) Destroy() {
		FfiDestroyerPeerBase{}.Destroy(r.Base);
		FfiDestroyerBool{}.Destroy(r.IsLocal);
		FfiDestroyerBool{}.Destroy(r.AllowIncomingConnections);
		FfiDestroyerBool{}.Destroy(r.AllowPeerTrafficRouting);
		FfiDestroyerBool{}.Destroy(r.AllowPeerLocalNetworkAccess);
		FfiDestroyerBool{}.Destroy(r.AllowPeerSendFiles);
		FfiDestroyerBool{}.Destroy(r.AllowMulticast);
		FfiDestroyerBool{}.Destroy(r.PeerAllowsMulticast);
}

type FfiConverterPeer struct {}

var FfiConverterPeerINSTANCE = FfiConverterPeer{}

func (c FfiConverterPeer) Lift(rb RustBufferI) Peer {
	return LiftFromRustBuffer[Peer](c, rb)
}

func (c FfiConverterPeer) Read(reader io.Reader) Peer {
	return Peer {
			FfiConverterPeerBaseINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterPeer) Lower(value Peer) C.RustBuffer {
	return LowerIntoRustBuffer[Peer](c, value)
}

func (c FfiConverterPeer) Write(writer io.Writer, value Peer) {
		FfiConverterPeerBaseINSTANCE.Write(writer, value.Base);
		FfiConverterBoolINSTANCE.Write(writer, value.IsLocal);
		FfiConverterBoolINSTANCE.Write(writer, value.AllowIncomingConnections);
		FfiConverterBoolINSTANCE.Write(writer, value.AllowPeerTrafficRouting);
		FfiConverterBoolINSTANCE.Write(writer, value.AllowPeerLocalNetworkAccess);
		FfiConverterBoolINSTANCE.Write(writer, value.AllowPeerSendFiles);
		FfiConverterBoolINSTANCE.Write(writer, value.AllowMulticast);
		FfiConverterBoolINSTANCE.Write(writer, value.PeerAllowsMulticast);
}

type FfiDestroyerPeer struct {}

func (_ FfiDestroyerPeer) Destroy(value Peer) {
	value.Destroy()
}


// Characterstics describing a peer
type PeerBase struct {
	// 32-character identifier of the peer
	Identifier string
	// Public key of the peer
	PublicKey PublicKey
	// Hostname of the peer
	Hostname HiddenString
	// Ip address of peer
	IpAddresses *[]IpAddr
	// Nickname for the peer
	Nickname *HiddenString
}

func (r *PeerBase) Destroy() {
		FfiDestroyerString{}.Destroy(r.Identifier);
		FfiDestroyerTypePublicKey{}.Destroy(r.PublicKey);
		FfiDestroyerTypeHiddenString{}.Destroy(r.Hostname);
		FfiDestroyerOptionalSequenceTypeIpAddr{}.Destroy(r.IpAddresses);
		FfiDestroyerOptionalTypeHiddenString{}.Destroy(r.Nickname);
}

type FfiConverterPeerBase struct {}

var FfiConverterPeerBaseINSTANCE = FfiConverterPeerBase{}

func (c FfiConverterPeerBase) Lift(rb RustBufferI) PeerBase {
	return LiftFromRustBuffer[PeerBase](c, rb)
}

func (c FfiConverterPeerBase) Read(reader io.Reader) PeerBase {
	return PeerBase {
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterTypePublicKeyINSTANCE.Read(reader),
			FfiConverterTypeHiddenStringINSTANCE.Read(reader),
			FfiConverterOptionalSequenceTypeIpAddrINSTANCE.Read(reader),
			FfiConverterOptionalTypeHiddenStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterPeerBase) Lower(value PeerBase) C.RustBuffer {
	return LowerIntoRustBuffer[PeerBase](c, value)
}

func (c FfiConverterPeerBase) Write(writer io.Writer, value PeerBase) {
		FfiConverterStringINSTANCE.Write(writer, value.Identifier);
		FfiConverterTypePublicKeyINSTANCE.Write(writer, value.PublicKey);
		FfiConverterTypeHiddenStringINSTANCE.Write(writer, value.Hostname);
		FfiConverterOptionalSequenceTypeIpAddrINSTANCE.Write(writer, value.IpAddresses);
		FfiConverterOptionalTypeHiddenStringINSTANCE.Write(writer, value.Nickname);
}

type FfiDestroyerPeerBase struct {}

func (_ FfiDestroyerPeerBase) Destroy(value PeerBase) {
	value.Destroy()
}


// Representation of a server, which might be used
// both as a Relay server and Stun Server
type Server struct {
	// Server region code
	RegionCode string
	// Short name for the server
	Name string
	// Hostname of the server
	Hostname string
	// IP address of the server
	Ipv4 Ipv4Addr
	// Port on which server listens to relay requests
	RelayPort uint16
	// Port on which server listens to stun requests
	StunPort uint16
	// Port on which server listens for unencrypted stun requests
	StunPlaintextPort uint16
	// Server public key
	PublicKey PublicKey
	// Determines in which order the client tries to connect to the derp servers
	Weight uint32
	// When enabled the connection to servers is not encrypted
	UsePlainText bool
	// Status of the connection with the server
	ConnState RelayState
}

func (r *Server) Destroy() {
		FfiDestroyerString{}.Destroy(r.RegionCode);
		FfiDestroyerString{}.Destroy(r.Name);
		FfiDestroyerString{}.Destroy(r.Hostname);
		FfiDestroyerTypeIpv4Addr{}.Destroy(r.Ipv4);
		FfiDestroyerUint16{}.Destroy(r.RelayPort);
		FfiDestroyerUint16{}.Destroy(r.StunPort);
		FfiDestroyerUint16{}.Destroy(r.StunPlaintextPort);
		FfiDestroyerTypePublicKey{}.Destroy(r.PublicKey);
		FfiDestroyerUint32{}.Destroy(r.Weight);
		FfiDestroyerBool{}.Destroy(r.UsePlainText);
		FfiDestroyerRelayState{}.Destroy(r.ConnState);
}

type FfiConverterServer struct {}

var FfiConverterServerINSTANCE = FfiConverterServer{}

func (c FfiConverterServer) Lift(rb RustBufferI) Server {
	return LiftFromRustBuffer[Server](c, rb)
}

func (c FfiConverterServer) Read(reader io.Reader) Server {
	return Server {
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterTypeIpv4AddrINSTANCE.Read(reader),
			FfiConverterUint16INSTANCE.Read(reader),
			FfiConverterUint16INSTANCE.Read(reader),
			FfiConverterUint16INSTANCE.Read(reader),
			FfiConverterTypePublicKeyINSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterRelayStateINSTANCE.Read(reader),
	}
}

func (c FfiConverterServer) Lower(value Server) C.RustBuffer {
	return LowerIntoRustBuffer[Server](c, value)
}

func (c FfiConverterServer) Write(writer io.Writer, value Server) {
		FfiConverterStringINSTANCE.Write(writer, value.RegionCode);
		FfiConverterStringINSTANCE.Write(writer, value.Name);
		FfiConverterStringINSTANCE.Write(writer, value.Hostname);
		FfiConverterTypeIpv4AddrINSTANCE.Write(writer, value.Ipv4);
		FfiConverterUint16INSTANCE.Write(writer, value.RelayPort);
		FfiConverterUint16INSTANCE.Write(writer, value.StunPort);
		FfiConverterUint16INSTANCE.Write(writer, value.StunPlaintextPort);
		FfiConverterTypePublicKeyINSTANCE.Write(writer, value.PublicKey);
		FfiConverterUint32INSTANCE.Write(writer, value.Weight);
		FfiConverterBoolINSTANCE.Write(writer, value.UsePlainText);
		FfiConverterRelayStateINSTANCE.Write(writer, value.ConnState);
}

type FfiDestroyerServer struct {}

func (_ FfiDestroyerServer) Destroy(value Server) {
	value.Destroy()
}


// Description of a Node
type TelioNode struct {
	// An identifier for a node
	// Makes it possible to distinguish different nodes in the presence of key reuse
	Identifier string
	// Public key of the Node
	PublicKey PublicKey
	// Nickname for the peer
	Nickname *string
	// State of the node (Connecting, connected, or disconnected)
	State NodeState
	// Link state hint (Down, Up)
	LinkState *LinkState
	// Is the node exit node
	IsExit bool
	// Is the node is a vpn server.
	IsVpn bool
	// IP addresses of the node
	IpAddresses []IpAddr
	// List of IP's which can connect to the node
	AllowedIps []IpNet
	// Endpoint used by node
	Endpoint *SocketAddr
	// Hostname of the node
	Hostname *string
	// Flag to control whether the Node allows incoming connections
	AllowIncomingConnections bool
	// Flag to control whether the Node allows routing through
	AllowPeerTrafficRouting bool
	// Flag to control whether the Node allows incoming local area access
	AllowPeerLocalNetworkAccess bool
	// Flag to control whether the Node allows incoming files
	AllowPeerSendFiles bool
	// Connection type in the network mesh (through Relay or hole punched directly)
	Path PathType
	// Flag to control whether we allow multicast messages from the Node
	AllowMulticast bool
	// Flag to control whether the Node allows multicast messages from us
	PeerAllowsMulticast bool
	// Configuration for the Error Notification Service
	VpnConnectionError *VpnConnectionError
}

func (r *TelioNode) Destroy() {
		FfiDestroyerString{}.Destroy(r.Identifier);
		FfiDestroyerTypePublicKey{}.Destroy(r.PublicKey);
		FfiDestroyerOptionalString{}.Destroy(r.Nickname);
		FfiDestroyerNodeState{}.Destroy(r.State);
		FfiDestroyerOptionalLinkState{}.Destroy(r.LinkState);
		FfiDestroyerBool{}.Destroy(r.IsExit);
		FfiDestroyerBool{}.Destroy(r.IsVpn);
		FfiDestroyerSequenceTypeIpAddr{}.Destroy(r.IpAddresses);
		FfiDestroyerSequenceTypeIpNet{}.Destroy(r.AllowedIps);
		FfiDestroyerOptionalTypeSocketAddr{}.Destroy(r.Endpoint);
		FfiDestroyerOptionalString{}.Destroy(r.Hostname);
		FfiDestroyerBool{}.Destroy(r.AllowIncomingConnections);
		FfiDestroyerBool{}.Destroy(r.AllowPeerTrafficRouting);
		FfiDestroyerBool{}.Destroy(r.AllowPeerLocalNetworkAccess);
		FfiDestroyerBool{}.Destroy(r.AllowPeerSendFiles);
		FfiDestroyerPathType{}.Destroy(r.Path);
		FfiDestroyerBool{}.Destroy(r.AllowMulticast);
		FfiDestroyerBool{}.Destroy(r.PeerAllowsMulticast);
		FfiDestroyerOptionalVpnConnectionError{}.Destroy(r.VpnConnectionError);
}

type FfiConverterTelioNode struct {}

var FfiConverterTelioNodeINSTANCE = FfiConverterTelioNode{}

func (c FfiConverterTelioNode) Lift(rb RustBufferI) TelioNode {
	return LiftFromRustBuffer[TelioNode](c, rb)
}

func (c FfiConverterTelioNode) Read(reader io.Reader) TelioNode {
	return TelioNode {
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterTypePublicKeyINSTANCE.Read(reader),
			FfiConverterOptionalStringINSTANCE.Read(reader),
			FfiConverterNodeStateINSTANCE.Read(reader),
			FfiConverterOptionalLinkStateINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterSequenceTypeIpAddrINSTANCE.Read(reader),
			FfiConverterSequenceTypeIpNetINSTANCE.Read(reader),
			FfiConverterOptionalTypeSocketAddrINSTANCE.Read(reader),
			FfiConverterOptionalStringINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterPathTypeINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterOptionalVpnConnectionErrorINSTANCE.Read(reader),
	}
}

func (c FfiConverterTelioNode) Lower(value TelioNode) C.RustBuffer {
	return LowerIntoRustBuffer[TelioNode](c, value)
}

func (c FfiConverterTelioNode) Write(writer io.Writer, value TelioNode) {
		FfiConverterStringINSTANCE.Write(writer, value.Identifier);
		FfiConverterTypePublicKeyINSTANCE.Write(writer, value.PublicKey);
		FfiConverterOptionalStringINSTANCE.Write(writer, value.Nickname);
		FfiConverterNodeStateINSTANCE.Write(writer, value.State);
		FfiConverterOptionalLinkStateINSTANCE.Write(writer, value.LinkState);
		FfiConverterBoolINSTANCE.Write(writer, value.IsExit);
		FfiConverterBoolINSTANCE.Write(writer, value.IsVpn);
		FfiConverterSequenceTypeIpAddrINSTANCE.Write(writer, value.IpAddresses);
		FfiConverterSequenceTypeIpNetINSTANCE.Write(writer, value.AllowedIps);
		FfiConverterOptionalTypeSocketAddrINSTANCE.Write(writer, value.Endpoint);
		FfiConverterOptionalStringINSTANCE.Write(writer, value.Hostname);
		FfiConverterBoolINSTANCE.Write(writer, value.AllowIncomingConnections);
		FfiConverterBoolINSTANCE.Write(writer, value.AllowPeerTrafficRouting);
		FfiConverterBoolINSTANCE.Write(writer, value.AllowPeerLocalNetworkAccess);
		FfiConverterBoolINSTANCE.Write(writer, value.AllowPeerSendFiles);
		FfiConverterPathTypeINSTANCE.Write(writer, value.Path);
		FfiConverterBoolINSTANCE.Write(writer, value.AllowMulticast);
		FfiConverterBoolINSTANCE.Write(writer, value.PeerAllowsMulticast);
		FfiConverterOptionalVpnConnectionErrorINSTANCE.Write(writer, value.VpnConnectionError);
}

type FfiDestroyerTelioNode struct {}

func (_ FfiDestroyerTelioNode) Destroy(value TelioNode) {
	value.Destroy()
}



// Available Endpoint Providers for meshnet direct connections
type EndpointProvider uint

const (
	// Use local interface ips as possible endpoints
	EndpointProviderLocal EndpointProvider = 1
	// Use stun and wg-stun results as possible endpoints
	EndpointProviderStun EndpointProvider = 2
	// Use IGD and upnp to generate endpoints
	EndpointProviderUpnp EndpointProvider = 3
)

type FfiConverterEndpointProvider struct {}

var FfiConverterEndpointProviderINSTANCE = FfiConverterEndpointProvider{}

func (c FfiConverterEndpointProvider) Lift(rb RustBufferI) EndpointProvider {
	return LiftFromRustBuffer[EndpointProvider](c, rb)
}

func (c FfiConverterEndpointProvider) Lower(value EndpointProvider) C.RustBuffer {
	return LowerIntoRustBuffer[EndpointProvider](c, value)
}
func (FfiConverterEndpointProvider) Read(reader io.Reader) EndpointProvider {
	id := readInt32(reader)
	return EndpointProvider(id)
}

func (FfiConverterEndpointProvider) Write(writer io.Writer, value EndpointProvider) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerEndpointProvider struct {}

func (_ FfiDestroyerEndpointProvider) Destroy(value EndpointProvider) {
}




// Error code. Common error code representation (for statistics).
type ErrorCode uint

const (
	// There is no error in the execution
	ErrorCodeNoError ErrorCode = 1
	// The error type is unknown
	ErrorCodeUnknown ErrorCode = 2
)

type FfiConverterErrorCode struct {}

var FfiConverterErrorCodeINSTANCE = FfiConverterErrorCode{}

func (c FfiConverterErrorCode) Lift(rb RustBufferI) ErrorCode {
	return LiftFromRustBuffer[ErrorCode](c, rb)
}

func (c FfiConverterErrorCode) Lower(value ErrorCode) C.RustBuffer {
	return LowerIntoRustBuffer[ErrorCode](c, value)
}
func (FfiConverterErrorCode) Read(reader io.Reader) ErrorCode {
	id := readInt32(reader)
	return ErrorCode(id)
}

func (FfiConverterErrorCode) Write(writer io.Writer, value ErrorCode) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerErrorCode struct {}

func (_ FfiDestroyerErrorCode) Destroy(value ErrorCode) {
}




// Error levels. Used for app to decide what to do with `telio` device when error happens.
type ErrorLevel uint

const (
	// The error level is critical (highest priority)
	ErrorLevelCritical ErrorLevel = 1
	// The error level is severe
	ErrorLevelSevere ErrorLevel = 2
	// The error is a warning
	ErrorLevelWarning ErrorLevel = 3
	// The error is of the lowest priority
	ErrorLevelNotice ErrorLevel = 4
)

type FfiConverterErrorLevel struct {}

var FfiConverterErrorLevelINSTANCE = FfiConverterErrorLevel{}

func (c FfiConverterErrorLevel) Lift(rb RustBufferI) ErrorLevel {
	return LiftFromRustBuffer[ErrorLevel](c, rb)
}

func (c FfiConverterErrorLevel) Lower(value ErrorLevel) C.RustBuffer {
	return LowerIntoRustBuffer[ErrorLevel](c, value)
}
func (FfiConverterErrorLevel) Read(reader io.Reader) ErrorLevel {
	id := readInt32(reader)
	return ErrorLevel(id)
}

func (FfiConverterErrorLevel) Write(writer io.Writer, value ErrorLevel) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerErrorLevel struct {}

func (_ FfiDestroyerErrorLevel) Destroy(value ErrorLevel) {
}




// Main object of `Event`. See `Event::new()` for init options.
type Event interface {
	Destroy()
}
// Used to report events related to the Relay
type EventRelay struct {
	Body Server
}

func (e EventRelay) Destroy() {
		FfiDestroyerServer{}.Destroy(e.Body);
}
// Used to report events related to the Node
type EventNode struct {
	Body TelioNode
}

func (e EventNode) Destroy() {
		FfiDestroyerTelioNode{}.Destroy(e.Body);
}
// Initialize an Error type event.
// Used to inform errors to the upper layers of libtelio
type EventError struct {
	Body ErrorEvent
}

func (e EventError) Destroy() {
		FfiDestroyerErrorEvent{}.Destroy(e.Body);
}

type FfiConverterEvent struct {}

var FfiConverterEventINSTANCE = FfiConverterEvent{}

func (c FfiConverterEvent) Lift(rb RustBufferI) Event {
	return LiftFromRustBuffer[Event](c, rb)
}

func (c FfiConverterEvent) Lower(value Event) C.RustBuffer {
	return LowerIntoRustBuffer[Event](c, value)
}
func (FfiConverterEvent) Read(reader io.Reader) Event {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return EventRelay{
				FfiConverterServerINSTANCE.Read(reader),
			};
		case 2:
			return EventNode{
				FfiConverterTelioNodeINSTANCE.Read(reader),
			};
		case 3:
			return EventError{
				FfiConverterErrorEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterEvent.Read()", id));
	}
}

func (FfiConverterEvent) Write(writer io.Writer, value Event) {
	switch variant_value := value.(type) {
		case EventRelay:
			writeInt32(writer, 1)
			FfiConverterServerINSTANCE.Write(writer, variant_value.Body)
		case EventNode:
			writeInt32(writer, 2)
			FfiConverterTelioNodeINSTANCE.Write(writer, variant_value.Body)
		case EventError:
			writeInt32(writer, 3)
			FfiConverterErrorEventINSTANCE.Write(writer, variant_value.Body)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterEvent.Write", value))
	}
}

type FfiDestroyerEvent struct {}

func (_ FfiDestroyerEvent) Destroy(value Event) {
	value.Destroy()
}




// Next layer protocol for IP packet
type IpProtocol uint

const (
	// UDP protocol
	IpProtocolUdp IpProtocol = 1
	// TCP protocol
	IpProtocolTcp IpProtocol = 2
)

type FfiConverterIpProtocol struct {}

var FfiConverterIpProtocolINSTANCE = FfiConverterIpProtocol{}

func (c FfiConverterIpProtocol) Lift(rb RustBufferI) IpProtocol {
	return LiftFromRustBuffer[IpProtocol](c, rb)
}

func (c FfiConverterIpProtocol) Lower(value IpProtocol) C.RustBuffer {
	return LowerIntoRustBuffer[IpProtocol](c, value)
}
func (FfiConverterIpProtocol) Read(reader io.Reader) IpProtocol {
	id := readInt32(reader)
	return IpProtocol(id)
}

func (FfiConverterIpProtocol) Write(writer io.Writer, value IpProtocol) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerIpProtocol struct {}

func (_ FfiDestroyerIpProtocol) Destroy(value IpProtocol) {
}




// Link state hint
type LinkState uint

const (
	LinkStateDown LinkState = 1
	LinkStateUp LinkState = 2
)

type FfiConverterLinkState struct {}

var FfiConverterLinkStateINSTANCE = FfiConverterLinkState{}

func (c FfiConverterLinkState) Lift(rb RustBufferI) LinkState {
	return LiftFromRustBuffer[LinkState](c, rb)
}

func (c FfiConverterLinkState) Lower(value LinkState) C.RustBuffer {
	return LowerIntoRustBuffer[LinkState](c, value)
}
func (FfiConverterLinkState) Read(reader io.Reader) LinkState {
	id := readInt32(reader)
	return LinkState(id)
}

func (FfiConverterLinkState) Write(writer io.Writer, value LinkState) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerLinkState struct {}

func (_ FfiDestroyerLinkState) Destroy(value LinkState) {
}




// Available NAT types
type NatType uint

const (
	// UDP is always blocked.
	NatTypeUdpBlocked NatType = 1
	// No NAT, public IP, no firewall.
	NatTypeOpenInternet NatType = 2
	// No NAT, public IP, but symmetric UDP firewall.
	NatTypeSymmetricUdpFirewall NatType = 3
	// A full cone NAT is one where all requests from the same internal IP address and port are
	// mapped to the same external IP address and port. Furthermore, any external host can send
	// a packet to the internal host, by sending a packet to the mapped external address.
	NatTypeFullCone NatType = 4
	// A restricted cone NAT is one where all requests from the same internal IP address and
	// port are mapped to the same external IP address and port. Unlike a full cone NAT, an external
	// host (with IP address X) can send a packet to the internal host only if the internal host
	// had previously sent a packet to IP address X.
	NatTypeRestrictedCone NatType = 5
	// A port restricted cone NAT is like a restricted cone NAT, but the restriction
	// includes port numbers. Specifically, an external host can send a packet, with source IP
	// address X and source port P, to the internal host only if the internal host had previously
	// sent a packet to IP address X and port P.
	NatTypePortRestrictedCone NatType = 6
	// A symmetric NAT is one where all requests from the same internal IP address and port,
	// to a specific destination IP address and port, are mapped to the same external IP address and
	// port.  If the same host sends a packet with the same source address and port, but to
	// a different destination, a different mapping is used. Furthermore, only the external host that
	// receives a packet can send a UDP packet back to the internal host.
	NatTypeSymmetric NatType = 7
	// Unknown
	NatTypeUnknown NatType = 8
)

type FfiConverterNatType struct {}

var FfiConverterNatTypeINSTANCE = FfiConverterNatType{}

func (c FfiConverterNatType) Lift(rb RustBufferI) NatType {
	return LiftFromRustBuffer[NatType](c, rb)
}

func (c FfiConverterNatType) Lower(value NatType) C.RustBuffer {
	return LowerIntoRustBuffer[NatType](c, value)
}
func (FfiConverterNatType) Read(reader io.Reader) NatType {
	id := readInt32(reader)
	return NatType(id)
}

func (FfiConverterNatType) Write(writer io.Writer, value NatType) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerNatType struct {}

func (_ FfiDestroyerNatType) Destroy(value NatType) {
}




// Connection state of the node
type NodeState uint

const (
	// Node is disconnected
	NodeStateDisconnected NodeState = 1
	// Trying to connect to the Node
	NodeStateConnecting NodeState = 2
	// Node is connected
	NodeStateConnected NodeState = 3
)

type FfiConverterNodeState struct {}

var FfiConverterNodeStateINSTANCE = FfiConverterNodeState{}

func (c FfiConverterNodeState) Lift(rb RustBufferI) NodeState {
	return LiftFromRustBuffer[NodeState](c, rb)
}

func (c FfiConverterNodeState) Lower(value NodeState) C.RustBuffer {
	return LowerIntoRustBuffer[NodeState](c, value)
}
func (FfiConverterNodeState) Read(reader io.Reader) NodeState {
	id := readInt32(reader)
	return NodeState(id)
}

func (FfiConverterNodeState) Write(writer io.Writer, value NodeState) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerNodeState struct {}

func (_ FfiDestroyerNodeState) Destroy(value NodeState) {
}




// Mesh connection path type
type PathType uint

const (
	// Nodes connected via a middle-man relay
	PathTypeRelay PathType = 1
	// Nodes connected directly via WG
	PathTypeDirect PathType = 2
)

type FfiConverterPathType struct {}

var FfiConverterPathTypeINSTANCE = FfiConverterPathType{}

func (c FfiConverterPathType) Lift(rb RustBufferI) PathType {
	return LiftFromRustBuffer[PathType](c, rb)
}

func (c FfiConverterPathType) Lower(value PathType) C.RustBuffer {
	return LowerIntoRustBuffer[PathType](c, value)
}
func (FfiConverterPathType) Read(reader io.Reader) PathType {
	id := readInt32(reader)
	return PathType(id)
}

func (FfiConverterPathType) Write(writer io.Writer, value PathType) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerPathType struct {}

func (_ FfiDestroyerPathType) Destroy(value PathType) {
}




// The currrent state of our connection to derp server
type RelayState uint

const (
	// Disconnected from the Derp server
	RelayStateDisconnected RelayState = 1
	// Connecting to the Derp server
	RelayStateConnecting RelayState = 2
	// Connected to the Derp server
	RelayStateConnected RelayState = 3
)

type FfiConverterRelayState struct {}

var FfiConverterRelayStateINSTANCE = FfiConverterRelayState{}

func (c FfiConverterRelayState) Lift(rb RustBufferI) RelayState {
	return LiftFromRustBuffer[RelayState](c, rb)
}

func (c FfiConverterRelayState) Lower(value RelayState) C.RustBuffer {
	return LowerIntoRustBuffer[RelayState](c, value)
}
func (FfiConverterRelayState) Read(reader io.Reader) RelayState {
	id := readInt32(reader)
	return RelayState(id)
}

func (FfiConverterRelayState) Write(writer io.Writer, value RelayState) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerRelayState struct {}

func (_ FfiDestroyerRelayState) Destroy(value RelayState) {
}




// Available ways to calculate RTT
type RttType uint

const (
	// Simple ping request
	RttTypePing RttType = 1
)

type FfiConverterRttType struct {}

var FfiConverterRttTypeINSTANCE = FfiConverterRttType{}

func (c FfiConverterRttType) Lift(rb RustBufferI) RttType {
	return LiftFromRustBuffer[RttType](c, rb)
}

func (c FfiConverterRttType) Lower(value RttType) C.RustBuffer {
	return LowerIntoRustBuffer[RttType](c, value)
}
func (FfiConverterRttType) Read(reader io.Reader) RttType {
	id := readInt32(reader)
	return RttType(id)
}

func (FfiConverterRttType) Write(writer io.Writer, value RttType) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerRttType struct {}

func (_ FfiDestroyerRttType) Destroy(value RttType) {
}




// Possible adapters.
type TelioAdapterType uint

const (
	// Userland rust implementation.
	TelioAdapterTypeNepTun TelioAdapterType = 1
	// Userland rust implementation. (Deprecated alias for NepTUN).
	TelioAdapterTypeBoringTun TelioAdapterType = 2
	// Linux in-kernel WireGuard implementation
	TelioAdapterTypeLinuxNativeTun TelioAdapterType = 3
	// WindowsNativeWireguardNt implementation
	TelioAdapterTypeWindowsNativeTun TelioAdapterType = 4
)

type FfiConverterTelioAdapterType struct {}

var FfiConverterTelioAdapterTypeINSTANCE = FfiConverterTelioAdapterType{}

func (c FfiConverterTelioAdapterType) Lift(rb RustBufferI) TelioAdapterType {
	return LiftFromRustBuffer[TelioAdapterType](c, rb)
}

func (c FfiConverterTelioAdapterType) Lower(value TelioAdapterType) C.RustBuffer {
	return LowerIntoRustBuffer[TelioAdapterType](c, value)
}
func (FfiConverterTelioAdapterType) Read(reader io.Reader) TelioAdapterType {
	id := readInt32(reader)
	return TelioAdapterType(id)
}

func (FfiConverterTelioAdapterType) Write(writer io.Writer, value TelioAdapterType) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTelioAdapterType struct {}

func (_ FfiDestroyerTelioAdapterType) Destroy(value TelioAdapterType) {
}


type TelioError struct {
	err error
}

// Convience method to turn *TelioError into error
// Avoiding treating nil pointer as non nil error interface
func (err *TelioError) AsError() error {
	if err == nil {
		return nil
	} else {
		return err
	}
}

func (err TelioError) Error() string {
	return fmt.Sprintf("TelioError: %s", err.err.Error())
}

func (err TelioError) Unwrap() error {
	return err.err
}

// Err* are used for checking error type with `errors.Is`
var ErrTelioErrorUnknownError = fmt.Errorf("TelioErrorUnknownError")
var ErrTelioErrorInvalidKey = fmt.Errorf("TelioErrorInvalidKey")
var ErrTelioErrorBadConfig = fmt.Errorf("TelioErrorBadConfig")
var ErrTelioErrorLockError = fmt.Errorf("TelioErrorLockError")
var ErrTelioErrorInvalidString = fmt.Errorf("TelioErrorInvalidString")
var ErrTelioErrorAlreadyStarted = fmt.Errorf("TelioErrorAlreadyStarted")
var ErrTelioErrorNotStarted = fmt.Errorf("TelioErrorNotStarted")

// Variant structs
type TelioErrorUnknownError struct {
	Inner string
}
func NewTelioErrorUnknownError(
	inner string,
) *TelioError {
	return &TelioError { err: &TelioErrorUnknownError {
			Inner: inner,} }
}

func (e TelioErrorUnknownError) destroy() {
		FfiDestroyerString{}.Destroy(e.Inner)
}


func (err TelioErrorUnknownError) Error() string {
	return fmt.Sprint("UnknownError",
		": ",
		
		"Inner=",
		err.Inner,
	)
}

func (self TelioErrorUnknownError) Is(target error) bool {
	return target == ErrTelioErrorUnknownError
}
type TelioErrorInvalidKey struct {
}
func NewTelioErrorInvalidKey(
) *TelioError {
	return &TelioError { err: &TelioErrorInvalidKey {} }
}

func (e TelioErrorInvalidKey) destroy() {
}


func (err TelioErrorInvalidKey) Error() string {
	return fmt.Sprint("InvalidKey",
		
	)
}

func (self TelioErrorInvalidKey) Is(target error) bool {
	return target == ErrTelioErrorInvalidKey
}
type TelioErrorBadConfig struct {
}
func NewTelioErrorBadConfig(
) *TelioError {
	return &TelioError { err: &TelioErrorBadConfig {} }
}

func (e TelioErrorBadConfig) destroy() {
}


func (err TelioErrorBadConfig) Error() string {
	return fmt.Sprint("BadConfig",
		
	)
}

func (self TelioErrorBadConfig) Is(target error) bool {
	return target == ErrTelioErrorBadConfig
}
type TelioErrorLockError struct {
}
func NewTelioErrorLockError(
) *TelioError {
	return &TelioError { err: &TelioErrorLockError {} }
}

func (e TelioErrorLockError) destroy() {
}


func (err TelioErrorLockError) Error() string {
	return fmt.Sprint("LockError",
		
	)
}

func (self TelioErrorLockError) Is(target error) bool {
	return target == ErrTelioErrorLockError
}
type TelioErrorInvalidString struct {
}
func NewTelioErrorInvalidString(
) *TelioError {
	return &TelioError { err: &TelioErrorInvalidString {} }
}

func (e TelioErrorInvalidString) destroy() {
}


func (err TelioErrorInvalidString) Error() string {
	return fmt.Sprint("InvalidString",
		
	)
}

func (self TelioErrorInvalidString) Is(target error) bool {
	return target == ErrTelioErrorInvalidString
}
type TelioErrorAlreadyStarted struct {
}
func NewTelioErrorAlreadyStarted(
) *TelioError {
	return &TelioError { err: &TelioErrorAlreadyStarted {} }
}

func (e TelioErrorAlreadyStarted) destroy() {
}


func (err TelioErrorAlreadyStarted) Error() string {
	return fmt.Sprint("AlreadyStarted",
		
	)
}

func (self TelioErrorAlreadyStarted) Is(target error) bool {
	return target == ErrTelioErrorAlreadyStarted
}
type TelioErrorNotStarted struct {
}
func NewTelioErrorNotStarted(
) *TelioError {
	return &TelioError { err: &TelioErrorNotStarted {} }
}

func (e TelioErrorNotStarted) destroy() {
}


func (err TelioErrorNotStarted) Error() string {
	return fmt.Sprint("NotStarted",
		
	)
}

func (self TelioErrorNotStarted) Is(target error) bool {
	return target == ErrTelioErrorNotStarted
}

type FfiConverterTelioError struct{}

var FfiConverterTelioErrorINSTANCE = FfiConverterTelioError{}

func (c FfiConverterTelioError) Lift(eb RustBufferI) *TelioError {
	return LiftFromRustBuffer[*TelioError](c, eb)
}

func (c FfiConverterTelioError) Lower(value *TelioError) C.RustBuffer {
	return LowerIntoRustBuffer[*TelioError](c, value)
}

func (c FfiConverterTelioError) Read(reader io.Reader) *TelioError {
	errorID := readUint32(reader)

	switch errorID {
	case 1:
		return &TelioError{ &TelioErrorUnknownError{
			Inner: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 2:
		return &TelioError{ &TelioErrorInvalidKey{
		}}
	case 3:
		return &TelioError{ &TelioErrorBadConfig{
		}}
	case 4:
		return &TelioError{ &TelioErrorLockError{
		}}
	case 5:
		return &TelioError{ &TelioErrorInvalidString{
		}}
	case 6:
		return &TelioError{ &TelioErrorAlreadyStarted{
		}}
	case 7:
		return &TelioError{ &TelioErrorNotStarted{
		}}
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterTelioError.Read()", errorID))
	}
}

func (c FfiConverterTelioError) Write(writer io.Writer, value *TelioError) {
	switch variantValue := value.err.(type) {
		case *TelioErrorUnknownError:
			writeInt32(writer, 1)
			FfiConverterStringINSTANCE.Write(writer, variantValue.Inner)
		case *TelioErrorInvalidKey:
			writeInt32(writer, 2)
		case *TelioErrorBadConfig:
			writeInt32(writer, 3)
		case *TelioErrorLockError:
			writeInt32(writer, 4)
		case *TelioErrorInvalidString:
			writeInt32(writer, 5)
		case *TelioErrorAlreadyStarted:
			writeInt32(writer, 6)
		case *TelioErrorNotStarted:
			writeInt32(writer, 7)
		default:
			_ = variantValue
			panic(fmt.Sprintf("invalid error value `%v` in FfiConverterTelioError.Write", value))
	}
}

type FfiDestroyerTelioError struct {}

func (_ FfiDestroyerTelioError) Destroy(value *TelioError) {
	switch variantValue := value.err.(type) {
		case TelioErrorUnknownError:
			variantValue.destroy()
		case TelioErrorInvalidKey:
			variantValue.destroy()
		case TelioErrorBadConfig:
			variantValue.destroy()
		case TelioErrorLockError:
			variantValue.destroy()
		case TelioErrorInvalidString:
			variantValue.destroy()
		case TelioErrorAlreadyStarted:
			variantValue.destroy()
		case TelioErrorNotStarted:
			variantValue.destroy()
		default:
			_ = variantValue
			panic(fmt.Sprintf("invalid error value `%v` in FfiDestroyerTelioError.Destroy", value))
	}
}




// Possible log levels.
type TelioLogLevel uint

const (
	TelioLogLevelError TelioLogLevel = 1
	TelioLogLevelWarning TelioLogLevel = 2
	TelioLogLevelInfo TelioLogLevel = 3
	TelioLogLevelDebug TelioLogLevel = 4
	TelioLogLevelTrace TelioLogLevel = 5
)

type FfiConverterTelioLogLevel struct {}

var FfiConverterTelioLogLevelINSTANCE = FfiConverterTelioLogLevel{}

func (c FfiConverterTelioLogLevel) Lift(rb RustBufferI) TelioLogLevel {
	return LiftFromRustBuffer[TelioLogLevel](c, rb)
}

func (c FfiConverterTelioLogLevel) Lower(value TelioLogLevel) C.RustBuffer {
	return LowerIntoRustBuffer[TelioLogLevel](c, value)
}
func (FfiConverterTelioLogLevel) Read(reader io.Reader) TelioLogLevel {
	id := readInt32(reader)
	return TelioLogLevel(id)
}

func (FfiConverterTelioLogLevel) Write(writer io.Writer, value TelioLogLevel) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTelioLogLevel struct {}

func (_ FfiDestroyerTelioLogLevel) Destroy(value TelioLogLevel) {
}




// Possible VPN errors received from the Error Notification Service
type VpnConnectionError uint

const (
	// Unknown error
	VpnConnectionErrorUnknown VpnConnectionError = 1
	// Connection limit reached
	VpnConnectionErrorConnectionLimitReached VpnConnectionError = 2
	// Server will undergo maintenance in the near future
	VpnConnectionErrorServerMaintenance VpnConnectionError = 3
	// Authentication failed
	VpnConnectionErrorUnauthenticated VpnConnectionError = 4
	// There is a newer connection to this VPN server
	VpnConnectionErrorSuperseded VpnConnectionError = 5
)

type FfiConverterVpnConnectionError struct {}

var FfiConverterVpnConnectionErrorINSTANCE = FfiConverterVpnConnectionError{}

func (c FfiConverterVpnConnectionError) Lift(rb RustBufferI) VpnConnectionError {
	return LiftFromRustBuffer[VpnConnectionError](c, rb)
}

func (c FfiConverterVpnConnectionError) Lower(value VpnConnectionError) C.RustBuffer {
	return LowerIntoRustBuffer[VpnConnectionError](c, value)
}
func (FfiConverterVpnConnectionError) Read(reader io.Reader) VpnConnectionError {
	id := readInt32(reader)
	return VpnConnectionError(id)
}

func (FfiConverterVpnConnectionError) Write(writer io.Writer, value VpnConnectionError) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerVpnConnectionError struct {}

func (_ FfiDestroyerVpnConnectionError) Destroy(value VpnConnectionError) {
}



type TelioEventCb interface {
	
	Event(payload Event) error
	
}


type FfiConverterCallbackInterfaceTelioEventCb struct {
	handleMap *concurrentHandleMap[TelioEventCb]
}

var FfiConverterCallbackInterfaceTelioEventCbINSTANCE = FfiConverterCallbackInterfaceTelioEventCb {
	handleMap: newConcurrentHandleMap[TelioEventCb](),
}

func (c FfiConverterCallbackInterfaceTelioEventCb) Lift(handle uint64) TelioEventCb {
	val, ok := c.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}
	return val
}

func (c FfiConverterCallbackInterfaceTelioEventCb) Read(reader io.Reader) TelioEventCb {
	return c.Lift(readUint64(reader))
}

func (c FfiConverterCallbackInterfaceTelioEventCb) Lower(value TelioEventCb) C.uint64_t {
	return C.uint64_t(c.handleMap.insert(value))
}

func (c FfiConverterCallbackInterfaceTelioEventCb) Write(writer io.Writer, value TelioEventCb) {
	writeUint64(writer, uint64(c.Lower(value)))
}

type FfiDestroyerCallbackInterfaceTelioEventCb struct {}

func (FfiDestroyerCallbackInterfaceTelioEventCb) Destroy(value TelioEventCb) {}

type uniffiCallbackResult C.int8_t

const (
	uniffiIdxCallbackFree               uniffiCallbackResult = 0
	uniffiCallbackResultSuccess         uniffiCallbackResult = 0
	uniffiCallbackResultError           uniffiCallbackResult = 1
	uniffiCallbackUnexpectedResultError uniffiCallbackResult = 2
	uniffiCallbackCancelled             uniffiCallbackResult = 3
)


type concurrentHandleMap[T any] struct {
	handles       map[uint64]T
	currentHandle uint64
	lock          sync.RWMutex
}

func newConcurrentHandleMap[T any]() *concurrentHandleMap[T] {
	return &concurrentHandleMap[T]{
		handles:  map[uint64]T{},
	}
}

func (cm *concurrentHandleMap[T]) insert(obj T) uint64 {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	cm.currentHandle = cm.currentHandle + 1
	cm.handles[cm.currentHandle] = obj
	return cm.currentHandle
}

func (cm *concurrentHandleMap[T]) remove(handle uint64) {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	delete(cm.handles, handle)
}

func (cm *concurrentHandleMap[T]) tryGet(handle uint64) (T, bool) {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	val, ok := cm.handles[handle]
	return val, ok
}

//export telio_cgo_dispatchCallbackInterfaceTelioEventCbMethod0
func telio_cgo_dispatchCallbackInterfaceTelioEventCbMethod0(uniffiHandle C.uint64_t,payload C.RustBuffer,uniffiOutReturn *C.void,callStatus *C.RustCallStatus,) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterCallbackInterfaceTelioEventCbINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}
	
	

	 err :=
    uniffiObj.Event(
        FfiConverterEventINSTANCE.Lift(GoRustBuffer {
		inner: payload,
	}),
    )
	
    
	if err != nil {
		var actualError *TelioError
		if errors.As(err, &actualError) {
			*callStatus = C.RustCallStatus {
				code: C.int8_t(uniffiCallbackResultError),
				errorBuf: FfiConverterTelioErrorINSTANCE.Lower(actualError),
			}
		} else {
			*callStatus = C.RustCallStatus {
				code: C.int8_t(uniffiCallbackUnexpectedResultError),
			}
		}
		return
	}


	
}

var UniffiVTableCallbackInterfaceTelioEventCbINSTANCE = C.UniffiVTableCallbackInterfaceTelioEventCb {
	event: (C.UniffiCallbackInterfaceTelioEventCbMethod0)(C.telio_cgo_dispatchCallbackInterfaceTelioEventCbMethod0),

	uniffiFree: (C.UniffiCallbackInterfaceFree)(C.telio_cgo_dispatchCallbackInterfaceTelioEventCbFree),
}

//export telio_cgo_dispatchCallbackInterfaceTelioEventCbFree
func telio_cgo_dispatchCallbackInterfaceTelioEventCbFree(handle C.uint64_t) {
	FfiConverterCallbackInterfaceTelioEventCbINSTANCE.handleMap.remove(uint64(handle))
}

func (c FfiConverterCallbackInterfaceTelioEventCb) register() {
	C.uniffi_telio_fn_init_callback_vtable_telioeventcb(&UniffiVTableCallbackInterfaceTelioEventCbINSTANCE)
}



type TelioLoggerCb interface {
	
	Log(logLevel TelioLogLevel, payload string) error
	
}


type FfiConverterCallbackInterfaceTelioLoggerCb struct {
	handleMap *concurrentHandleMap[TelioLoggerCb]
}

var FfiConverterCallbackInterfaceTelioLoggerCbINSTANCE = FfiConverterCallbackInterfaceTelioLoggerCb {
	handleMap: newConcurrentHandleMap[TelioLoggerCb](),
}

func (c FfiConverterCallbackInterfaceTelioLoggerCb) Lift(handle uint64) TelioLoggerCb {
	val, ok := c.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}
	return val
}

func (c FfiConverterCallbackInterfaceTelioLoggerCb) Read(reader io.Reader) TelioLoggerCb {
	return c.Lift(readUint64(reader))
}

func (c FfiConverterCallbackInterfaceTelioLoggerCb) Lower(value TelioLoggerCb) C.uint64_t {
	return C.uint64_t(c.handleMap.insert(value))
}

func (c FfiConverterCallbackInterfaceTelioLoggerCb) Write(writer io.Writer, value TelioLoggerCb) {
	writeUint64(writer, uint64(c.Lower(value)))
}

type FfiDestroyerCallbackInterfaceTelioLoggerCb struct {}

func (FfiDestroyerCallbackInterfaceTelioLoggerCb) Destroy(value TelioLoggerCb) {}



//export telio_cgo_dispatchCallbackInterfaceTelioLoggerCbMethod0
func telio_cgo_dispatchCallbackInterfaceTelioLoggerCbMethod0(uniffiHandle C.uint64_t,logLevel C.RustBuffer,payload C.RustBuffer,uniffiOutReturn *C.void,callStatus *C.RustCallStatus,) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterCallbackInterfaceTelioLoggerCbINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}
	
	

	 err :=
    uniffiObj.Log(
        FfiConverterTelioLogLevelINSTANCE.Lift(GoRustBuffer {
		inner: logLevel,
	}),
        FfiConverterStringINSTANCE.Lift(GoRustBuffer {
		inner: payload,
	}),
    )
	
    
	if err != nil {
		var actualError *TelioError
		if errors.As(err, &actualError) {
			*callStatus = C.RustCallStatus {
				code: C.int8_t(uniffiCallbackResultError),
				errorBuf: FfiConverterTelioErrorINSTANCE.Lower(actualError),
			}
		} else {
			*callStatus = C.RustCallStatus {
				code: C.int8_t(uniffiCallbackUnexpectedResultError),
			}
		}
		return
	}


	
}

var UniffiVTableCallbackInterfaceTelioLoggerCbINSTANCE = C.UniffiVTableCallbackInterfaceTelioLoggerCb {
	log: (C.UniffiCallbackInterfaceTelioLoggerCbMethod0)(C.telio_cgo_dispatchCallbackInterfaceTelioLoggerCbMethod0),

	uniffiFree: (C.UniffiCallbackInterfaceFree)(C.telio_cgo_dispatchCallbackInterfaceTelioLoggerCbFree),
}

//export telio_cgo_dispatchCallbackInterfaceTelioLoggerCbFree
func telio_cgo_dispatchCallbackInterfaceTelioLoggerCbFree(handle C.uint64_t) {
	FfiConverterCallbackInterfaceTelioLoggerCbINSTANCE.handleMap.remove(uint64(handle))
}

func (c FfiConverterCallbackInterfaceTelioLoggerCb) register() {
	C.uniffi_telio_fn_init_callback_vtable_teliologgercb(&UniffiVTableCallbackInterfaceTelioLoggerCbINSTANCE)
}



type TelioProtectCb interface {
	
	Protect(socketId int32) error
	
}


type FfiConverterCallbackInterfaceTelioProtectCb struct {
	handleMap *concurrentHandleMap[TelioProtectCb]
}

var FfiConverterCallbackInterfaceTelioProtectCbINSTANCE = FfiConverterCallbackInterfaceTelioProtectCb {
	handleMap: newConcurrentHandleMap[TelioProtectCb](),
}

func (c FfiConverterCallbackInterfaceTelioProtectCb) Lift(handle uint64) TelioProtectCb {
	val, ok := c.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}
	return val
}

func (c FfiConverterCallbackInterfaceTelioProtectCb) Read(reader io.Reader) TelioProtectCb {
	return c.Lift(readUint64(reader))
}

func (c FfiConverterCallbackInterfaceTelioProtectCb) Lower(value TelioProtectCb) C.uint64_t {
	return C.uint64_t(c.handleMap.insert(value))
}

func (c FfiConverterCallbackInterfaceTelioProtectCb) Write(writer io.Writer, value TelioProtectCb) {
	writeUint64(writer, uint64(c.Lower(value)))
}

type FfiDestroyerCallbackInterfaceTelioProtectCb struct {}

func (FfiDestroyerCallbackInterfaceTelioProtectCb) Destroy(value TelioProtectCb) {}



//export telio_cgo_dispatchCallbackInterfaceTelioProtectCbMethod0
func telio_cgo_dispatchCallbackInterfaceTelioProtectCbMethod0(uniffiHandle C.uint64_t,socketId C.int32_t,uniffiOutReturn *C.void,callStatus *C.RustCallStatus,) {
	handle := uint64(uniffiHandle)
	uniffiObj, ok := FfiConverterCallbackInterfaceTelioProtectCbINSTANCE.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}
	
	

	 err :=
    uniffiObj.Protect(
        FfiConverterInt32INSTANCE.Lift(socketId),
    )
	
    
	if err != nil {
		var actualError *TelioError
		if errors.As(err, &actualError) {
			*callStatus = C.RustCallStatus {
				code: C.int8_t(uniffiCallbackResultError),
				errorBuf: FfiConverterTelioErrorINSTANCE.Lower(actualError),
			}
		} else {
			*callStatus = C.RustCallStatus {
				code: C.int8_t(uniffiCallbackUnexpectedResultError),
			}
		}
		return
	}


	
}

var UniffiVTableCallbackInterfaceTelioProtectCbINSTANCE = C.UniffiVTableCallbackInterfaceTelioProtectCb {
	protect: (C.UniffiCallbackInterfaceTelioProtectCbMethod0)(C.telio_cgo_dispatchCallbackInterfaceTelioProtectCbMethod0),

	uniffiFree: (C.UniffiCallbackInterfaceFree)(C.telio_cgo_dispatchCallbackInterfaceTelioProtectCbFree),
}

//export telio_cgo_dispatchCallbackInterfaceTelioProtectCbFree
func telio_cgo_dispatchCallbackInterfaceTelioProtectCbFree(handle C.uint64_t) {
	FfiConverterCallbackInterfaceTelioProtectCbINSTANCE.handleMap.remove(uint64(handle))
}

func (c FfiConverterCallbackInterfaceTelioProtectCb) register() {
	C.uniffi_telio_fn_init_callback_vtable_telioprotectcb(&UniffiVTableCallbackInterfaceTelioProtectCbINSTANCE)
}




type FfiConverterOptionalUint32 struct{}

var FfiConverterOptionalUint32INSTANCE = FfiConverterOptionalUint32{}

func (c FfiConverterOptionalUint32) Lift(rb RustBufferI) *uint32 {
	return LiftFromRustBuffer[*uint32](c, rb)
}

func (_ FfiConverterOptionalUint32) Read(reader io.Reader) *uint32 {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterUint32INSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalUint32) Lower(value *uint32) C.RustBuffer {
	return LowerIntoRustBuffer[*uint32](c, value)
}

func (_ FfiConverterOptionalUint32) Write(writer io.Writer, value *uint32) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterUint32INSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalUint32 struct {}

func (_ FfiDestroyerOptionalUint32) Destroy(value *uint32) {
	if value != nil {
		FfiDestroyerUint32{}.Destroy(*value)
	}
}



type FfiConverterOptionalUint64 struct{}

var FfiConverterOptionalUint64INSTANCE = FfiConverterOptionalUint64{}

func (c FfiConverterOptionalUint64) Lift(rb RustBufferI) *uint64 {
	return LiftFromRustBuffer[*uint64](c, rb)
}

func (_ FfiConverterOptionalUint64) Read(reader io.Reader) *uint64 {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterUint64INSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalUint64) Lower(value *uint64) C.RustBuffer {
	return LowerIntoRustBuffer[*uint64](c, value)
}

func (_ FfiConverterOptionalUint64) Write(writer io.Writer, value *uint64) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterUint64INSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalUint64 struct {}

func (_ FfiDestroyerOptionalUint64) Destroy(value *uint64) {
	if value != nil {
		FfiDestroyerUint64{}.Destroy(*value)
	}
}



type FfiConverterOptionalBool struct{}

var FfiConverterOptionalBoolINSTANCE = FfiConverterOptionalBool{}

func (c FfiConverterOptionalBool) Lift(rb RustBufferI) *bool {
	return LiftFromRustBuffer[*bool](c, rb)
}

func (_ FfiConverterOptionalBool) Read(reader io.Reader) *bool {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterBoolINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalBool) Lower(value *bool) C.RustBuffer {
	return LowerIntoRustBuffer[*bool](c, value)
}

func (_ FfiConverterOptionalBool) Write(writer io.Writer, value *bool) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterBoolINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalBool struct {}

func (_ FfiDestroyerOptionalBool) Destroy(value *bool) {
	if value != nil {
		FfiDestroyerBool{}.Destroy(*value)
	}
}



type FfiConverterOptionalString struct{}

var FfiConverterOptionalStringINSTANCE = FfiConverterOptionalString{}

func (c FfiConverterOptionalString) Lift(rb RustBufferI) *string {
	return LiftFromRustBuffer[*string](c, rb)
}

func (_ FfiConverterOptionalString) Read(reader io.Reader) *string {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterStringINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalString) Lower(value *string) C.RustBuffer {
	return LowerIntoRustBuffer[*string](c, value)
}

func (_ FfiConverterOptionalString) Write(writer io.Writer, value *string) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterStringINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalString struct {}

func (_ FfiDestroyerOptionalString) Destroy(value *string) {
	if value != nil {
		FfiDestroyerString{}.Destroy(*value)
	}
}



type FfiConverterOptionalDnsConfig struct{}

var FfiConverterOptionalDnsConfigINSTANCE = FfiConverterOptionalDnsConfig{}

func (c FfiConverterOptionalDnsConfig) Lift(rb RustBufferI) *DnsConfig {
	return LiftFromRustBuffer[*DnsConfig](c, rb)
}

func (_ FfiConverterOptionalDnsConfig) Read(reader io.Reader) *DnsConfig {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterDnsConfigINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalDnsConfig) Lower(value *DnsConfig) C.RustBuffer {
	return LowerIntoRustBuffer[*DnsConfig](c, value)
}

func (_ FfiConverterOptionalDnsConfig) Write(writer io.Writer, value *DnsConfig) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterDnsConfigINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalDnsConfig struct {}

func (_ FfiDestroyerOptionalDnsConfig) Destroy(value *DnsConfig) {
	if value != nil {
		FfiDestroyerDnsConfig{}.Destroy(*value)
	}
}



type FfiConverterOptionalFeatureBatching struct{}

var FfiConverterOptionalFeatureBatchingINSTANCE = FfiConverterOptionalFeatureBatching{}

func (c FfiConverterOptionalFeatureBatching) Lift(rb RustBufferI) *FeatureBatching {
	return LiftFromRustBuffer[*FeatureBatching](c, rb)
}

func (_ FfiConverterOptionalFeatureBatching) Read(reader io.Reader) *FeatureBatching {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterFeatureBatchingINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalFeatureBatching) Lower(value *FeatureBatching) C.RustBuffer {
	return LowerIntoRustBuffer[*FeatureBatching](c, value)
}

func (_ FfiConverterOptionalFeatureBatching) Write(writer io.Writer, value *FeatureBatching) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterFeatureBatchingINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalFeatureBatching struct {}

func (_ FfiDestroyerOptionalFeatureBatching) Destroy(value *FeatureBatching) {
	if value != nil {
		FfiDestroyerFeatureBatching{}.Destroy(*value)
	}
}



type FfiConverterOptionalFeatureDerp struct{}

var FfiConverterOptionalFeatureDerpINSTANCE = FfiConverterOptionalFeatureDerp{}

func (c FfiConverterOptionalFeatureDerp) Lift(rb RustBufferI) *FeatureDerp {
	return LiftFromRustBuffer[*FeatureDerp](c, rb)
}

func (_ FfiConverterOptionalFeatureDerp) Read(reader io.Reader) *FeatureDerp {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterFeatureDerpINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalFeatureDerp) Lower(value *FeatureDerp) C.RustBuffer {
	return LowerIntoRustBuffer[*FeatureDerp](c, value)
}

func (_ FfiConverterOptionalFeatureDerp) Write(writer io.Writer, value *FeatureDerp) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterFeatureDerpINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalFeatureDerp struct {}

func (_ FfiDestroyerOptionalFeatureDerp) Destroy(value *FeatureDerp) {
	if value != nil {
		FfiDestroyerFeatureDerp{}.Destroy(*value)
	}
}



type FfiConverterOptionalFeatureDirect struct{}

var FfiConverterOptionalFeatureDirectINSTANCE = FfiConverterOptionalFeatureDirect{}

func (c FfiConverterOptionalFeatureDirect) Lift(rb RustBufferI) *FeatureDirect {
	return LiftFromRustBuffer[*FeatureDirect](c, rb)
}

func (_ FfiConverterOptionalFeatureDirect) Read(reader io.Reader) *FeatureDirect {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterFeatureDirectINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalFeatureDirect) Lower(value *FeatureDirect) C.RustBuffer {
	return LowerIntoRustBuffer[*FeatureDirect](c, value)
}

func (_ FfiConverterOptionalFeatureDirect) Write(writer io.Writer, value *FeatureDirect) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterFeatureDirectINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalFeatureDirect struct {}

func (_ FfiDestroyerOptionalFeatureDirect) Destroy(value *FeatureDirect) {
	if value != nil {
		FfiDestroyerFeatureDirect{}.Destroy(*value)
	}
}



type FfiConverterOptionalFeatureEndpointProvidersOptimization struct{}

var FfiConverterOptionalFeatureEndpointProvidersOptimizationINSTANCE = FfiConverterOptionalFeatureEndpointProvidersOptimization{}

func (c FfiConverterOptionalFeatureEndpointProvidersOptimization) Lift(rb RustBufferI) *FeatureEndpointProvidersOptimization {
	return LiftFromRustBuffer[*FeatureEndpointProvidersOptimization](c, rb)
}

func (_ FfiConverterOptionalFeatureEndpointProvidersOptimization) Read(reader io.Reader) *FeatureEndpointProvidersOptimization {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterFeatureEndpointProvidersOptimizationINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalFeatureEndpointProvidersOptimization) Lower(value *FeatureEndpointProvidersOptimization) C.RustBuffer {
	return LowerIntoRustBuffer[*FeatureEndpointProvidersOptimization](c, value)
}

func (_ FfiConverterOptionalFeatureEndpointProvidersOptimization) Write(writer io.Writer, value *FeatureEndpointProvidersOptimization) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterFeatureEndpointProvidersOptimizationINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalFeatureEndpointProvidersOptimization struct {}

func (_ FfiDestroyerOptionalFeatureEndpointProvidersOptimization) Destroy(value *FeatureEndpointProvidersOptimization) {
	if value != nil {
		FfiDestroyerFeatureEndpointProvidersOptimization{}.Destroy(*value)
	}
}



type FfiConverterOptionalFeatureErrorNotificationService struct{}

var FfiConverterOptionalFeatureErrorNotificationServiceINSTANCE = FfiConverterOptionalFeatureErrorNotificationService{}

func (c FfiConverterOptionalFeatureErrorNotificationService) Lift(rb RustBufferI) *FeatureErrorNotificationService {
	return LiftFromRustBuffer[*FeatureErrorNotificationService](c, rb)
}

func (_ FfiConverterOptionalFeatureErrorNotificationService) Read(reader io.Reader) *FeatureErrorNotificationService {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterFeatureErrorNotificationServiceINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalFeatureErrorNotificationService) Lower(value *FeatureErrorNotificationService) C.RustBuffer {
	return LowerIntoRustBuffer[*FeatureErrorNotificationService](c, value)
}

func (_ FfiConverterOptionalFeatureErrorNotificationService) Write(writer io.Writer, value *FeatureErrorNotificationService) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterFeatureErrorNotificationServiceINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalFeatureErrorNotificationService struct {}

func (_ FfiDestroyerOptionalFeatureErrorNotificationService) Destroy(value *FeatureErrorNotificationService) {
	if value != nil {
		FfiDestroyerFeatureErrorNotificationService{}.Destroy(*value)
	}
}



type FfiConverterOptionalFeatureExitDns struct{}

var FfiConverterOptionalFeatureExitDnsINSTANCE = FfiConverterOptionalFeatureExitDns{}

func (c FfiConverterOptionalFeatureExitDns) Lift(rb RustBufferI) *FeatureExitDns {
	return LiftFromRustBuffer[*FeatureExitDns](c, rb)
}

func (_ FfiConverterOptionalFeatureExitDns) Read(reader io.Reader) *FeatureExitDns {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterFeatureExitDnsINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalFeatureExitDns) Lower(value *FeatureExitDns) C.RustBuffer {
	return LowerIntoRustBuffer[*FeatureExitDns](c, value)
}

func (_ FfiConverterOptionalFeatureExitDns) Write(writer io.Writer, value *FeatureExitDns) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterFeatureExitDnsINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalFeatureExitDns struct {}

func (_ FfiDestroyerOptionalFeatureExitDns) Destroy(value *FeatureExitDns) {
	if value != nil {
		FfiDestroyerFeatureExitDns{}.Destroy(*value)
	}
}



type FfiConverterOptionalFeatureLana struct{}

var FfiConverterOptionalFeatureLanaINSTANCE = FfiConverterOptionalFeatureLana{}

func (c FfiConverterOptionalFeatureLana) Lift(rb RustBufferI) *FeatureLana {
	return LiftFromRustBuffer[*FeatureLana](c, rb)
}

func (_ FfiConverterOptionalFeatureLana) Read(reader io.Reader) *FeatureLana {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterFeatureLanaINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalFeatureLana) Lower(value *FeatureLana) C.RustBuffer {
	return LowerIntoRustBuffer[*FeatureLana](c, value)
}

func (_ FfiConverterOptionalFeatureLana) Write(writer io.Writer, value *FeatureLana) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterFeatureLanaINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalFeatureLana struct {}

func (_ FfiDestroyerOptionalFeatureLana) Destroy(value *FeatureLana) {
	if value != nil {
		FfiDestroyerFeatureLana{}.Destroy(*value)
	}
}



type FfiConverterOptionalFeatureLinkDetection struct{}

var FfiConverterOptionalFeatureLinkDetectionINSTANCE = FfiConverterOptionalFeatureLinkDetection{}

func (c FfiConverterOptionalFeatureLinkDetection) Lift(rb RustBufferI) *FeatureLinkDetection {
	return LiftFromRustBuffer[*FeatureLinkDetection](c, rb)
}

func (_ FfiConverterOptionalFeatureLinkDetection) Read(reader io.Reader) *FeatureLinkDetection {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterFeatureLinkDetectionINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalFeatureLinkDetection) Lower(value *FeatureLinkDetection) C.RustBuffer {
	return LowerIntoRustBuffer[*FeatureLinkDetection](c, value)
}

func (_ FfiConverterOptionalFeatureLinkDetection) Write(writer io.Writer, value *FeatureLinkDetection) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterFeatureLinkDetectionINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalFeatureLinkDetection struct {}

func (_ FfiDestroyerOptionalFeatureLinkDetection) Destroy(value *FeatureLinkDetection) {
	if value != nil {
		FfiDestroyerFeatureLinkDetection{}.Destroy(*value)
	}
}



type FfiConverterOptionalFeatureNurse struct{}

var FfiConverterOptionalFeatureNurseINSTANCE = FfiConverterOptionalFeatureNurse{}

func (c FfiConverterOptionalFeatureNurse) Lift(rb RustBufferI) *FeatureNurse {
	return LiftFromRustBuffer[*FeatureNurse](c, rb)
}

func (_ FfiConverterOptionalFeatureNurse) Read(reader io.Reader) *FeatureNurse {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterFeatureNurseINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalFeatureNurse) Lower(value *FeatureNurse) C.RustBuffer {
	return LowerIntoRustBuffer[*FeatureNurse](c, value)
}

func (_ FfiConverterOptionalFeatureNurse) Write(writer io.Writer, value *FeatureNurse) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterFeatureNurseINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalFeatureNurse struct {}

func (_ FfiDestroyerOptionalFeatureNurse) Destroy(value *FeatureNurse) {
	if value != nil {
		FfiDestroyerFeatureNurse{}.Destroy(*value)
	}
}



type FfiConverterOptionalFeaturePaths struct{}

var FfiConverterOptionalFeaturePathsINSTANCE = FfiConverterOptionalFeaturePaths{}

func (c FfiConverterOptionalFeaturePaths) Lift(rb RustBufferI) *FeaturePaths {
	return LiftFromRustBuffer[*FeaturePaths](c, rb)
}

func (_ FfiConverterOptionalFeaturePaths) Read(reader io.Reader) *FeaturePaths {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterFeaturePathsINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalFeaturePaths) Lower(value *FeaturePaths) C.RustBuffer {
	return LowerIntoRustBuffer[*FeaturePaths](c, value)
}

func (_ FfiConverterOptionalFeaturePaths) Write(writer io.Writer, value *FeaturePaths) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterFeaturePathsINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalFeaturePaths struct {}

func (_ FfiDestroyerOptionalFeaturePaths) Destroy(value *FeaturePaths) {
	if value != nil {
		FfiDestroyerFeaturePaths{}.Destroy(*value)
	}
}



type FfiConverterOptionalFeatureQoS struct{}

var FfiConverterOptionalFeatureQoSINSTANCE = FfiConverterOptionalFeatureQoS{}

func (c FfiConverterOptionalFeatureQoS) Lift(rb RustBufferI) *FeatureQoS {
	return LiftFromRustBuffer[*FeatureQoS](c, rb)
}

func (_ FfiConverterOptionalFeatureQoS) Read(reader io.Reader) *FeatureQoS {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterFeatureQoSINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalFeatureQoS) Lower(value *FeatureQoS) C.RustBuffer {
	return LowerIntoRustBuffer[*FeatureQoS](c, value)
}

func (_ FfiConverterOptionalFeatureQoS) Write(writer io.Writer, value *FeatureQoS) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterFeatureQoSINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalFeatureQoS struct {}

func (_ FfiDestroyerOptionalFeatureQoS) Destroy(value *FeatureQoS) {
	if value != nil {
		FfiDestroyerFeatureQoS{}.Destroy(*value)
	}
}



type FfiConverterOptionalFeatureSkipUnresponsivePeers struct{}

var FfiConverterOptionalFeatureSkipUnresponsivePeersINSTANCE = FfiConverterOptionalFeatureSkipUnresponsivePeers{}

func (c FfiConverterOptionalFeatureSkipUnresponsivePeers) Lift(rb RustBufferI) *FeatureSkipUnresponsivePeers {
	return LiftFromRustBuffer[*FeatureSkipUnresponsivePeers](c, rb)
}

func (_ FfiConverterOptionalFeatureSkipUnresponsivePeers) Read(reader io.Reader) *FeatureSkipUnresponsivePeers {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterFeatureSkipUnresponsivePeersINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalFeatureSkipUnresponsivePeers) Lower(value *FeatureSkipUnresponsivePeers) C.RustBuffer {
	return LowerIntoRustBuffer[*FeatureSkipUnresponsivePeers](c, value)
}

func (_ FfiConverterOptionalFeatureSkipUnresponsivePeers) Write(writer io.Writer, value *FeatureSkipUnresponsivePeers) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterFeatureSkipUnresponsivePeersINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalFeatureSkipUnresponsivePeers struct {}

func (_ FfiDestroyerOptionalFeatureSkipUnresponsivePeers) Destroy(value *FeatureSkipUnresponsivePeers) {
	if value != nil {
		FfiDestroyerFeatureSkipUnresponsivePeers{}.Destroy(*value)
	}
}



type FfiConverterOptionalFeatureUpnp struct{}

var FfiConverterOptionalFeatureUpnpINSTANCE = FfiConverterOptionalFeatureUpnp{}

func (c FfiConverterOptionalFeatureUpnp) Lift(rb RustBufferI) *FeatureUpnp {
	return LiftFromRustBuffer[*FeatureUpnp](c, rb)
}

func (_ FfiConverterOptionalFeatureUpnp) Read(reader io.Reader) *FeatureUpnp {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterFeatureUpnpINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalFeatureUpnp) Lower(value *FeatureUpnp) C.RustBuffer {
	return LowerIntoRustBuffer[*FeatureUpnp](c, value)
}

func (_ FfiConverterOptionalFeatureUpnp) Write(writer io.Writer, value *FeatureUpnp) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterFeatureUpnpINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalFeatureUpnp struct {}

func (_ FfiDestroyerOptionalFeatureUpnp) Destroy(value *FeatureUpnp) {
	if value != nil {
		FfiDestroyerFeatureUpnp{}.Destroy(*value)
	}
}



type FfiConverterOptionalLinkState struct{}

var FfiConverterOptionalLinkStateINSTANCE = FfiConverterOptionalLinkState{}

func (c FfiConverterOptionalLinkState) Lift(rb RustBufferI) *LinkState {
	return LiftFromRustBuffer[*LinkState](c, rb)
}

func (_ FfiConverterOptionalLinkState) Read(reader io.Reader) *LinkState {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterLinkStateINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalLinkState) Lower(value *LinkState) C.RustBuffer {
	return LowerIntoRustBuffer[*LinkState](c, value)
}

func (_ FfiConverterOptionalLinkState) Write(writer io.Writer, value *LinkState) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterLinkStateINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalLinkState struct {}

func (_ FfiDestroyerOptionalLinkState) Destroy(value *LinkState) {
	if value != nil {
		FfiDestroyerLinkState{}.Destroy(*value)
	}
}



type FfiConverterOptionalPathType struct{}

var FfiConverterOptionalPathTypeINSTANCE = FfiConverterOptionalPathType{}

func (c FfiConverterOptionalPathType) Lift(rb RustBufferI) *PathType {
	return LiftFromRustBuffer[*PathType](c, rb)
}

func (_ FfiConverterOptionalPathType) Read(reader io.Reader) *PathType {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterPathTypeINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalPathType) Lower(value *PathType) C.RustBuffer {
	return LowerIntoRustBuffer[*PathType](c, value)
}

func (_ FfiConverterOptionalPathType) Write(writer io.Writer, value *PathType) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterPathTypeINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalPathType struct {}

func (_ FfiDestroyerOptionalPathType) Destroy(value *PathType) {
	if value != nil {
		FfiDestroyerPathType{}.Destroy(*value)
	}
}



type FfiConverterOptionalVpnConnectionError struct{}

var FfiConverterOptionalVpnConnectionErrorINSTANCE = FfiConverterOptionalVpnConnectionError{}

func (c FfiConverterOptionalVpnConnectionError) Lift(rb RustBufferI) *VpnConnectionError {
	return LiftFromRustBuffer[*VpnConnectionError](c, rb)
}

func (_ FfiConverterOptionalVpnConnectionError) Read(reader io.Reader) *VpnConnectionError {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterVpnConnectionErrorINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalVpnConnectionError) Lower(value *VpnConnectionError) C.RustBuffer {
	return LowerIntoRustBuffer[*VpnConnectionError](c, value)
}

func (_ FfiConverterOptionalVpnConnectionError) Write(writer io.Writer, value *VpnConnectionError) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterVpnConnectionErrorINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalVpnConnectionError struct {}

func (_ FfiDestroyerOptionalVpnConnectionError) Destroy(value *VpnConnectionError) {
	if value != nil {
		FfiDestroyerVpnConnectionError{}.Destroy(*value)
	}
}



type FfiConverterOptionalSequencePeer struct{}

var FfiConverterOptionalSequencePeerINSTANCE = FfiConverterOptionalSequencePeer{}

func (c FfiConverterOptionalSequencePeer) Lift(rb RustBufferI) *[]Peer {
	return LiftFromRustBuffer[*[]Peer](c, rb)
}

func (_ FfiConverterOptionalSequencePeer) Read(reader io.Reader) *[]Peer {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterSequencePeerINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalSequencePeer) Lower(value *[]Peer) C.RustBuffer {
	return LowerIntoRustBuffer[*[]Peer](c, value)
}

func (_ FfiConverterOptionalSequencePeer) Write(writer io.Writer, value *[]Peer) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterSequencePeerINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalSequencePeer struct {}

func (_ FfiDestroyerOptionalSequencePeer) Destroy(value *[]Peer) {
	if value != nil {
		FfiDestroyerSequencePeer{}.Destroy(*value)
	}
}



type FfiConverterOptionalSequenceServer struct{}

var FfiConverterOptionalSequenceServerINSTANCE = FfiConverterOptionalSequenceServer{}

func (c FfiConverterOptionalSequenceServer) Lift(rb RustBufferI) *[]Server {
	return LiftFromRustBuffer[*[]Server](c, rb)
}

func (_ FfiConverterOptionalSequenceServer) Read(reader io.Reader) *[]Server {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterSequenceServerINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalSequenceServer) Lower(value *[]Server) C.RustBuffer {
	return LowerIntoRustBuffer[*[]Server](c, value)
}

func (_ FfiConverterOptionalSequenceServer) Write(writer io.Writer, value *[]Server) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterSequenceServerINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalSequenceServer struct {}

func (_ FfiDestroyerOptionalSequenceServer) Destroy(value *[]Server) {
	if value != nil {
		FfiDestroyerSequenceServer{}.Destroy(*value)
	}
}



type FfiConverterOptionalSequenceTypeIpAddr struct{}

var FfiConverterOptionalSequenceTypeIpAddrINSTANCE = FfiConverterOptionalSequenceTypeIpAddr{}

func (c FfiConverterOptionalSequenceTypeIpAddr) Lift(rb RustBufferI) *[]IpAddr {
	return LiftFromRustBuffer[*[]IpAddr](c, rb)
}

func (_ FfiConverterOptionalSequenceTypeIpAddr) Read(reader io.Reader) *[]IpAddr {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterSequenceTypeIpAddrINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalSequenceTypeIpAddr) Lower(value *[]IpAddr) C.RustBuffer {
	return LowerIntoRustBuffer[*[]IpAddr](c, value)
}

func (_ FfiConverterOptionalSequenceTypeIpAddr) Write(writer io.Writer, value *[]IpAddr) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterSequenceTypeIpAddrINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalSequenceTypeIpAddr struct {}

func (_ FfiDestroyerOptionalSequenceTypeIpAddr) Destroy(value *[]IpAddr) {
	if value != nil {
		FfiDestroyerSequenceTypeIpAddr{}.Destroy(*value)
	}
}



type FfiConverterOptionalSequenceTypeIpNet struct{}

var FfiConverterOptionalSequenceTypeIpNetINSTANCE = FfiConverterOptionalSequenceTypeIpNet{}

func (c FfiConverterOptionalSequenceTypeIpNet) Lift(rb RustBufferI) *[]IpNet {
	return LiftFromRustBuffer[*[]IpNet](c, rb)
}

func (_ FfiConverterOptionalSequenceTypeIpNet) Read(reader io.Reader) *[]IpNet {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterSequenceTypeIpNetINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalSequenceTypeIpNet) Lower(value *[]IpNet) C.RustBuffer {
	return LowerIntoRustBuffer[*[]IpNet](c, value)
}

func (_ FfiConverterOptionalSequenceTypeIpNet) Write(writer io.Writer, value *[]IpNet) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterSequenceTypeIpNetINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalSequenceTypeIpNet struct {}

func (_ FfiDestroyerOptionalSequenceTypeIpNet) Destroy(value *[]IpNet) {
	if value != nil {
		FfiDestroyerSequenceTypeIpNet{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypeEndpointProviders struct{}

var FfiConverterOptionalTypeEndpointProvidersINSTANCE = FfiConverterOptionalTypeEndpointProviders{}

func (c FfiConverterOptionalTypeEndpointProviders) Lift(rb RustBufferI) *EndpointProviders {
	return LiftFromRustBuffer[*EndpointProviders](c, rb)
}

func (_ FfiConverterOptionalTypeEndpointProviders) Read(reader io.Reader) *EndpointProviders {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeEndpointProvidersINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeEndpointProviders) Lower(value *EndpointProviders) C.RustBuffer {
	return LowerIntoRustBuffer[*EndpointProviders](c, value)
}

func (_ FfiConverterOptionalTypeEndpointProviders) Write(writer io.Writer, value *EndpointProviders) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeEndpointProvidersINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeEndpointProviders struct {}

func (_ FfiDestroyerOptionalTypeEndpointProviders) Destroy(value *EndpointProviders) {
	if value != nil {
		FfiDestroyerTypeEndpointProviders{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypeHiddenString struct{}

var FfiConverterOptionalTypeHiddenStringINSTANCE = FfiConverterOptionalTypeHiddenString{}

func (c FfiConverterOptionalTypeHiddenString) Lift(rb RustBufferI) *HiddenString {
	return LiftFromRustBuffer[*HiddenString](c, rb)
}

func (_ FfiConverterOptionalTypeHiddenString) Read(reader io.Reader) *HiddenString {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeHiddenStringINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeHiddenString) Lower(value *HiddenString) C.RustBuffer {
	return LowerIntoRustBuffer[*HiddenString](c, value)
}

func (_ FfiConverterOptionalTypeHiddenString) Write(writer io.Writer, value *HiddenString) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeHiddenStringINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeHiddenString struct {}

func (_ FfiDestroyerOptionalTypeHiddenString) Destroy(value *HiddenString) {
	if value != nil {
		FfiDestroyerTypeHiddenString{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypeIpv4Net struct{}

var FfiConverterOptionalTypeIpv4NetINSTANCE = FfiConverterOptionalTypeIpv4Net{}

func (c FfiConverterOptionalTypeIpv4Net) Lift(rb RustBufferI) *Ipv4Net {
	return LiftFromRustBuffer[*Ipv4Net](c, rb)
}

func (_ FfiConverterOptionalTypeIpv4Net) Read(reader io.Reader) *Ipv4Net {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeIpv4NetINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeIpv4Net) Lower(value *Ipv4Net) C.RustBuffer {
	return LowerIntoRustBuffer[*Ipv4Net](c, value)
}

func (_ FfiConverterOptionalTypeIpv4Net) Write(writer io.Writer, value *Ipv4Net) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeIpv4NetINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeIpv4Net struct {}

func (_ FfiDestroyerOptionalTypeIpv4Net) Destroy(value *Ipv4Net) {
	if value != nil {
		FfiDestroyerTypeIpv4Net{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypeSocketAddr struct{}

var FfiConverterOptionalTypeSocketAddrINSTANCE = FfiConverterOptionalTypeSocketAddr{}

func (c FfiConverterOptionalTypeSocketAddr) Lift(rb RustBufferI) *SocketAddr {
	return LiftFromRustBuffer[*SocketAddr](c, rb)
}

func (_ FfiConverterOptionalTypeSocketAddr) Read(reader io.Reader) *SocketAddr {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeSocketAddrINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeSocketAddr) Lower(value *SocketAddr) C.RustBuffer {
	return LowerIntoRustBuffer[*SocketAddr](c, value)
}

func (_ FfiConverterOptionalTypeSocketAddr) Write(writer io.Writer, value *SocketAddr) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeSocketAddrINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeSocketAddr struct {}

func (_ FfiDestroyerOptionalTypeSocketAddr) Destroy(value *SocketAddr) {
	if value != nil {
		FfiDestroyerTypeSocketAddr{}.Destroy(*value)
	}
}



type FfiConverterSequenceFirewallBlacklistTuple struct{}

var FfiConverterSequenceFirewallBlacklistTupleINSTANCE = FfiConverterSequenceFirewallBlacklistTuple{}

func (c FfiConverterSequenceFirewallBlacklistTuple) Lift(rb RustBufferI) []FirewallBlacklistTuple {
	return LiftFromRustBuffer[[]FirewallBlacklistTuple](c, rb)
}

func (c FfiConverterSequenceFirewallBlacklistTuple) Read(reader io.Reader) []FirewallBlacklistTuple {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]FirewallBlacklistTuple, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterFirewallBlacklistTupleINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceFirewallBlacklistTuple) Lower(value []FirewallBlacklistTuple) C.RustBuffer {
	return LowerIntoRustBuffer[[]FirewallBlacklistTuple](c, value)
}

func (c FfiConverterSequenceFirewallBlacklistTuple) Write(writer io.Writer, value []FirewallBlacklistTuple) {
	if len(value) > math.MaxInt32 {
		panic("[]FirewallBlacklistTuple is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterFirewallBlacklistTupleINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceFirewallBlacklistTuple struct {}

func (FfiDestroyerSequenceFirewallBlacklistTuple) Destroy(sequence []FirewallBlacklistTuple) {
	for _, value := range sequence {
		FfiDestroyerFirewallBlacklistTuple{}.Destroy(value)	
	}
}



type FfiConverterSequencePeer struct{}

var FfiConverterSequencePeerINSTANCE = FfiConverterSequencePeer{}

func (c FfiConverterSequencePeer) Lift(rb RustBufferI) []Peer {
	return LiftFromRustBuffer[[]Peer](c, rb)
}

func (c FfiConverterSequencePeer) Read(reader io.Reader) []Peer {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]Peer, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterPeerINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequencePeer) Lower(value []Peer) C.RustBuffer {
	return LowerIntoRustBuffer[[]Peer](c, value)
}

func (c FfiConverterSequencePeer) Write(writer io.Writer, value []Peer) {
	if len(value) > math.MaxInt32 {
		panic("[]Peer is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterPeerINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequencePeer struct {}

func (FfiDestroyerSequencePeer) Destroy(sequence []Peer) {
	for _, value := range sequence {
		FfiDestroyerPeer{}.Destroy(value)	
	}
}



type FfiConverterSequenceServer struct{}

var FfiConverterSequenceServerINSTANCE = FfiConverterSequenceServer{}

func (c FfiConverterSequenceServer) Lift(rb RustBufferI) []Server {
	return LiftFromRustBuffer[[]Server](c, rb)
}

func (c FfiConverterSequenceServer) Read(reader io.Reader) []Server {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]Server, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterServerINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceServer) Lower(value []Server) C.RustBuffer {
	return LowerIntoRustBuffer[[]Server](c, value)
}

func (c FfiConverterSequenceServer) Write(writer io.Writer, value []Server) {
	if len(value) > math.MaxInt32 {
		panic("[]Server is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterServerINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceServer struct {}

func (FfiDestroyerSequenceServer) Destroy(sequence []Server) {
	for _, value := range sequence {
		FfiDestroyerServer{}.Destroy(value)	
	}
}



type FfiConverterSequenceTelioNode struct{}

var FfiConverterSequenceTelioNodeINSTANCE = FfiConverterSequenceTelioNode{}

func (c FfiConverterSequenceTelioNode) Lift(rb RustBufferI) []TelioNode {
	return LiftFromRustBuffer[[]TelioNode](c, rb)
}

func (c FfiConverterSequenceTelioNode) Read(reader io.Reader) []TelioNode {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]TelioNode, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTelioNodeINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTelioNode) Lower(value []TelioNode) C.RustBuffer {
	return LowerIntoRustBuffer[[]TelioNode](c, value)
}

func (c FfiConverterSequenceTelioNode) Write(writer io.Writer, value []TelioNode) {
	if len(value) > math.MaxInt32 {
		panic("[]TelioNode is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTelioNodeINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTelioNode struct {}

func (FfiDestroyerSequenceTelioNode) Destroy(sequence []TelioNode) {
	for _, value := range sequence {
		FfiDestroyerTelioNode{}.Destroy(value)	
	}
}



type FfiConverterSequenceEndpointProvider struct{}

var FfiConverterSequenceEndpointProviderINSTANCE = FfiConverterSequenceEndpointProvider{}

func (c FfiConverterSequenceEndpointProvider) Lift(rb RustBufferI) []EndpointProvider {
	return LiftFromRustBuffer[[]EndpointProvider](c, rb)
}

func (c FfiConverterSequenceEndpointProvider) Read(reader io.Reader) []EndpointProvider {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]EndpointProvider, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterEndpointProviderINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceEndpointProvider) Lower(value []EndpointProvider) C.RustBuffer {
	return LowerIntoRustBuffer[[]EndpointProvider](c, value)
}

func (c FfiConverterSequenceEndpointProvider) Write(writer io.Writer, value []EndpointProvider) {
	if len(value) > math.MaxInt32 {
		panic("[]EndpointProvider is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterEndpointProviderINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceEndpointProvider struct {}

func (FfiDestroyerSequenceEndpointProvider) Destroy(sequence []EndpointProvider) {
	for _, value := range sequence {
		FfiDestroyerEndpointProvider{}.Destroy(value)	
	}
}



type FfiConverterSequencePathType struct{}

var FfiConverterSequencePathTypeINSTANCE = FfiConverterSequencePathType{}

func (c FfiConverterSequencePathType) Lift(rb RustBufferI) []PathType {
	return LiftFromRustBuffer[[]PathType](c, rb)
}

func (c FfiConverterSequencePathType) Read(reader io.Reader) []PathType {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]PathType, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterPathTypeINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequencePathType) Lower(value []PathType) C.RustBuffer {
	return LowerIntoRustBuffer[[]PathType](c, value)
}

func (c FfiConverterSequencePathType) Write(writer io.Writer, value []PathType) {
	if len(value) > math.MaxInt32 {
		panic("[]PathType is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterPathTypeINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequencePathType struct {}

func (FfiDestroyerSequencePathType) Destroy(sequence []PathType) {
	for _, value := range sequence {
		FfiDestroyerPathType{}.Destroy(value)	
	}
}



type FfiConverterSequenceRttType struct{}

var FfiConverterSequenceRttTypeINSTANCE = FfiConverterSequenceRttType{}

func (c FfiConverterSequenceRttType) Lift(rb RustBufferI) []RttType {
	return LiftFromRustBuffer[[]RttType](c, rb)
}

func (c FfiConverterSequenceRttType) Read(reader io.Reader) []RttType {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]RttType, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterRttTypeINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceRttType) Lower(value []RttType) C.RustBuffer {
	return LowerIntoRustBuffer[[]RttType](c, value)
}

func (c FfiConverterSequenceRttType) Write(writer io.Writer, value []RttType) {
	if len(value) > math.MaxInt32 {
		panic("[]RttType is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterRttTypeINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceRttType struct {}

func (FfiDestroyerSequenceRttType) Destroy(sequence []RttType) {
	for _, value := range sequence {
		FfiDestroyerRttType{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeIpAddr struct{}

var FfiConverterSequenceTypeIpAddrINSTANCE = FfiConverterSequenceTypeIpAddr{}

func (c FfiConverterSequenceTypeIpAddr) Lift(rb RustBufferI) []IpAddr {
	return LiftFromRustBuffer[[]IpAddr](c, rb)
}

func (c FfiConverterSequenceTypeIpAddr) Read(reader io.Reader) []IpAddr {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]IpAddr, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeIpAddrINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeIpAddr) Lower(value []IpAddr) C.RustBuffer {
	return LowerIntoRustBuffer[[]IpAddr](c, value)
}

func (c FfiConverterSequenceTypeIpAddr) Write(writer io.Writer, value []IpAddr) {
	if len(value) > math.MaxInt32 {
		panic("[]IpAddr is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeIpAddrINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeIpAddr struct {}

func (FfiDestroyerSequenceTypeIpAddr) Destroy(sequence []IpAddr) {
	for _, value := range sequence {
		FfiDestroyerTypeIpAddr{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeIpNet struct{}

var FfiConverterSequenceTypeIpNetINSTANCE = FfiConverterSequenceTypeIpNet{}

func (c FfiConverterSequenceTypeIpNet) Lift(rb RustBufferI) []IpNet {
	return LiftFromRustBuffer[[]IpNet](c, rb)
}

func (c FfiConverterSequenceTypeIpNet) Read(reader io.Reader) []IpNet {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]IpNet, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeIpNetINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeIpNet) Lower(value []IpNet) C.RustBuffer {
	return LowerIntoRustBuffer[[]IpNet](c, value)
}

func (c FfiConverterSequenceTypeIpNet) Write(writer io.Writer, value []IpNet) {
	if len(value) > math.MaxInt32 {
		panic("[]IpNet is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeIpNetINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeIpNet struct {}

func (FfiDestroyerSequenceTypeIpNet) Destroy(sequence []IpNet) {
	for _, value := range sequence {
		FfiDestroyerTypeIpNet{}.Destroy(value)	
	}
}


/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type EndpointProviders = []EndpointProvider
type FfiConverterTypeEndpointProviders = FfiConverterSequenceEndpointProvider
type FfiDestroyerTypeEndpointProviders = FfiDestroyerSequenceEndpointProvider
var FfiConverterTypeEndpointProvidersINSTANCE = FfiConverterSequenceEndpointProvider{}


/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type FeatureValidateKeys = bool
type FfiConverterTypeFeatureValidateKeys = FfiConverterBool
type FfiDestroyerTypeFeatureValidateKeys = FfiDestroyerBool
var FfiConverterTypeFeatureValidateKeysINSTANCE = FfiConverterBool{}


/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type HiddenString = string
type FfiConverterTypeHiddenString = FfiConverterString
type FfiDestroyerTypeHiddenString = FfiDestroyerString
var FfiConverterTypeHiddenStringINSTANCE = FfiConverterString{}


/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type IpAddr = string
type FfiConverterTypeIpAddr = FfiConverterString
type FfiDestroyerTypeIpAddr = FfiDestroyerString
var FfiConverterTypeIpAddrINSTANCE = FfiConverterString{}


/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type IpNet = string
type FfiConverterTypeIpNet = FfiConverterString
type FfiDestroyerTypeIpNet = FfiDestroyerString
var FfiConverterTypeIpNetINSTANCE = FfiConverterString{}


/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type Ipv4Addr = string
type FfiConverterTypeIpv4Addr = FfiConverterString
type FfiDestroyerTypeIpv4Addr = FfiDestroyerString
var FfiConverterTypeIpv4AddrINSTANCE = FfiConverterString{}


/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type Ipv4Net = string
type FfiConverterTypeIpv4Net = FfiConverterString
type FfiDestroyerTypeIpv4Net = FfiDestroyerString
var FfiConverterTypeIpv4NetINSTANCE = FfiConverterString{}


/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type PublicKey = string
type FfiConverterTypePublicKey = FfiConverterString
type FfiDestroyerTypePublicKey = FfiDestroyerString
var FfiConverterTypePublicKeyINSTANCE = FfiConverterString{}


/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type SecretKey = string
type FfiConverterTypeSecretKey = FfiConverterString
type FfiDestroyerTypeSecretKey = FfiDestroyerString
var FfiConverterTypeSecretKeyINSTANCE = FfiConverterString{}


/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type SocketAddr = string
type FfiConverterTypeSocketAddr = FfiConverterString
type FfiDestroyerTypeSocketAddr = FfiDestroyerString
var FfiConverterTypeSocketAddrINSTANCE = FfiConverterString{}


/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type TtlValue = uint32
type FfiConverterTypeTtlValue = FfiConverterUint32
type FfiDestroyerTypeTtlValue = FfiDestroyerUint32
var FfiConverterTypeTtlValueINSTANCE = FfiConverterUint32{}

// For testing only - embeds timestamps into generated logs
func AddTimestampsToLogs()  {
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_func_add_timestamps_to_logs(_uniffiStatus)
		return false
	})
}

// Utility function to create a `Features` object from a json-string
// Passing an empty string will return the default feature config
func DeserializeFeatureConfig(fstr string) (Features, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer {
		inner: C.uniffi_telio_fn_func_deserialize_feature_config(FfiConverterStringINSTANCE.Lower(fstr),_uniffiStatus),
	}
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue Features
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterFeaturesINSTANCE.Lift(_uniffiRV), nil
		}
}

// Utility function to create a `Config` object from a json-string
func DeserializeMeshnetConfig(cfgStr string) (Config, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer {
		inner: C.uniffi_telio_fn_func_deserialize_meshnet_config(FfiConverterStringINSTANCE.Lower(cfgStr),_uniffiStatus),
	}
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue Config
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterConfigINSTANCE.Lift(_uniffiRV), nil
		}
}

// Get the public key that corresponds to a given private key.
func GeneratePublicKey(secretKey SecretKey) PublicKey {
	return FfiConverterTypePublicKeyINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer {
		inner: C.uniffi_telio_fn_func_generate_public_key(FfiConverterTypeSecretKeyINSTANCE.Lower(secretKey),_uniffiStatus),
	}
	}))
}

// Generate a new secret key.
func GenerateSecretKey() SecretKey {
	return FfiConverterTypeSecretKeyINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer {
		inner: C.uniffi_telio_fn_func_generate_secret_key(_uniffiStatus),
	}
	}))
}

// Get current commit sha.
func GetCommitSha() string {
	return FfiConverterStringINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer {
		inner: C.uniffi_telio_fn_func_get_commit_sha(_uniffiStatus),
	}
	}))
}

// Get default recommended adapter type for platform.
func GetDefaultAdapter() TelioAdapterType {
	return FfiConverterTelioAdapterTypeINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer {
		inner: C.uniffi_telio_fn_func_get_default_adapter(_uniffiStatus),
	}
	}))
}

// Utility function to get the default feature config
func GetDefaultFeatureConfig() Features {
	return FfiConverterFeaturesINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer {
		inner: C.uniffi_telio_fn_func_get_default_feature_config(_uniffiStatus),
	}
	}))
}

// Get current version tag.
func GetVersionTag() string {
	return FfiConverterStringINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer {
		inner: C.uniffi_telio_fn_func_get_version_tag(_uniffiStatus),
	}
	}))
}

// Utility function to create a json-string from a Features instance
func SerializeFeatureConfig(features Features) (string, error) {
	_uniffiRV, _uniffiErr := rustCallWithError[TelioError](FfiConverterTelioError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return GoRustBuffer {
		inner: C.uniffi_telio_fn_func_serialize_feature_config(FfiConverterFeaturesINSTANCE.Lower(features),_uniffiStatus),
	}
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue string
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterStringINSTANCE.Lift(_uniffiRV), nil
		}
}

// Set the global logger.
// # Parameters
// - `log_level`: Max log level to log.
// - `logger`: Callback to handle logging events.
func SetGlobalLogger(logLevel TelioLogLevel, logger TelioLoggerCb)  {
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_func_set_global_logger(FfiConverterTelioLogLevelINSTANCE.Lower(logLevel), FfiConverterCallbackInterfaceTelioLoggerCbINSTANCE.Lower(logger),_uniffiStatus)
		return false
	})
}

// Unset the global logger.
// After this call finishes, previously registered logger will not be called.
func UnsetGlobalLogger()  {
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_func_unset_global_logger(_uniffiStatus)
		return false
	})
}

