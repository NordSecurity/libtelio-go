
package telio

// #include <telio.h>
import "C"

import (
	"bytes"
	"fmt"
	"io"
	"unsafe"
	"encoding/binary"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
)



type RustBuffer = C.RustBuffer

type RustBufferI interface {
	AsReader() *bytes.Reader
	Free()
	ToGoBytes() []byte
	Data() unsafe.Pointer
	Len() int
	Capacity() int
}

func RustBufferFromExternal(b RustBufferI) RustBuffer {
	return RustBuffer {
		capacity: C.int(b.Capacity()),
		len: C.int(b.Len()),
		data: (*C.uchar)(b.Data()),
	}
}

func (cb RustBuffer) Capacity() int {
	return int(cb.capacity)
}

func (cb RustBuffer) Len() int {
	return int(cb.len)
}

func (cb RustBuffer) Data() unsafe.Pointer {
	return unsafe.Pointer(cb.data)
}

func (cb RustBuffer) AsReader() *bytes.Reader {
	b := unsafe.Slice((*byte)(cb.data), C.int(cb.len))
	return bytes.NewReader(b)
}

func (cb RustBuffer) Free() {
	rustCall(func( status *C.RustCallStatus) bool {
		C.ffi_telio_rustbuffer_free(cb, status)
		return false
	})
}

func (cb RustBuffer) ToGoBytes() []byte {
	return C.GoBytes(unsafe.Pointer(cb.data), C.int(cb.len))
}


func stringToRustBuffer(str string) RustBuffer {
	return bytesToRustBuffer([]byte(str))
}

func bytesToRustBuffer(b []byte) RustBuffer {
	if len(b) == 0 {
		return RustBuffer{}
	}
	// We can pass the pointer along here, as it is pinned
	// for the duration of this call
	foreign := C.ForeignBytes {
		len: C.int(len(b)),
		data: (*C.uchar)(unsafe.Pointer(&b[0])),
	}
	
	return rustCall(func( status *C.RustCallStatus) RustBuffer {
		return C.ffi_telio_rustbuffer_from_bytes(foreign, status)
	})
}



type BufLifter[GoType any] interface {
	Lift(value RustBufferI) GoType
}

type BufLowerer[GoType any] interface {
	Lower(value GoType) RustBuffer
}

type FfiConverter[GoType any, FfiType any] interface {
	Lift(value FfiType) GoType
	Lower(value GoType) FfiType
}

type BufReader[GoType any] interface {
	Read(reader io.Reader) GoType
}

type BufWriter[GoType any] interface {
	Write(writer io.Writer, value GoType)
}

type FfiRustBufConverter[GoType any, FfiType any] interface {
	FfiConverter[GoType, FfiType]
	BufReader[GoType]
}

func LowerIntoRustBuffer[GoType any](bufWriter BufWriter[GoType], value GoType) RustBuffer {
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



func rustCallWithError[U any](converter BufLifter[error], callback func(*C.RustCallStatus) U) (U, error) {
	var status C.RustCallStatus
	returnValue := callback(&status)
	err := checkCallStatus(converter, status)

	return returnValue, err
}

func checkCallStatus(converter BufLifter[error], status C.RustCallStatus) error {
	switch status.code {
	case 0:
		return nil
	case 1:
		return converter.Lift(status.errorBuf)
	case 2:
		// when the rust code sees a panic, it tries to construct a rustbuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(status.errorBuf)))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		return fmt.Errorf("unknown status code: %d", status.code)
	}
}

func checkCallStatusUnknown(status C.RustCallStatus) error {
	switch status.code {
	case 0:
		return nil
	case 1:
		panic(fmt.Errorf("function not returning an error returned an error"))
	case 2:
		// when the rust code sees a panic, it tries to construct a rustbuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(status.errorBuf)))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		return fmt.Errorf("unknown status code: %d", status.code)
	}
}

func rustCall[U any](callback func(*C.RustCallStatus) U) U {
	returnValue, err := rustCallWithError(nil, callback)
	if err != nil {
		panic(err)
	}
	return returnValue
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
        
        (&FfiConverterCallbackInterfaceTelioEventCb{}).register();
        (&FfiConverterCallbackInterfaceTelioLoggerCb{}).register();
        (&FfiConverterCallbackInterfaceTelioProtectCb{}).register();
        uniffiCheckChecksums()
}


func uniffiCheckChecksums() {
	// Get the bindings contract version from our ComponentInterface
	bindingsContractVersion := 24
	// Get the scaffolding contract version by calling the into the dylib
	scaffoldingContractVersion := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint32_t {
		return C.ffi_telio_uniffi_contract_version(uniffiStatus)
	})
	if bindingsContractVersion != int(scaffoldingContractVersion) {
		// If this happens try cleaning and rebuilding your project
		panic("telio: UniFFI contract version mismatch")
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_func_deserialize_feature_config(uniffiStatus)
	})
	if checksum != 61040 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_func_deserialize_feature_config: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_func_deserialize_meshnet_config(uniffiStatus)
	})
	if checksum != 5696 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_func_deserialize_meshnet_config: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_func_generate_public_key(uniffiStatus)
	})
	if checksum != 39651 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_func_generate_public_key: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_func_generate_secret_key(uniffiStatus)
	})
	if checksum != 44282 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_func_generate_secret_key: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_func_get_commit_sha(uniffiStatus)
	})
	if checksum != 39165 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_func_get_commit_sha: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_func_get_default_adapter(uniffiStatus)
	})
	if checksum != 64813 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_func_get_default_adapter: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_func_get_default_feature_config(uniffiStatus)
	})
	if checksum != 52439 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_func_get_default_feature_config: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_func_get_version_tag(uniffiStatus)
	})
	if checksum != 53700 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_func_get_version_tag: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_func_set_global_logger(uniffiStatus)
	})
	if checksum != 25683 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_func_set_global_logger: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_build(uniffiStatus)
	})
	if checksum != 18842 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_build: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_batching(uniffiStatus)
	})
	if checksum != 27812 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_batching: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_battery_saving_defaults(uniffiStatus)
	})
	if checksum != 10214 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_battery_saving_defaults: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_direct(uniffiStatus)
	})
	if checksum != 8489 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_direct: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_firewall_connection_reset(uniffiStatus)
	})
	if checksum != 63055 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_firewall_connection_reset: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_flush_events_on_stop_timeout_seconds(uniffiStatus)
	})
	if checksum != 48141 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_flush_events_on_stop_timeout_seconds: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_ipv6(uniffiStatus)
	})
	if checksum != 25251 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_ipv6: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_lana(uniffiStatus)
	})
	if checksum != 20972 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_lana: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_link_detection(uniffiStatus)
	})
	if checksum != 35122 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_link_detection: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_multicast(uniffiStatus)
	})
	if checksum != 10758 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_multicast: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_nicknames(uniffiStatus)
	})
	if checksum != 59848 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_nicknames: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_nurse(uniffiStatus)
	})
	if checksum != 24340 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_nurse: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_pmtu_discovery(uniffiStatus)
	})
	if checksum != 39164 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_pmtu_discovery: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_validate_keys(uniffiStatus)
	})
	if checksum != 10605 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_featuresdefaultsbuilder_enable_validate_keys: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_connect_to_exit_node(uniffiStatus)
	})
	if checksum != 62657 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_connect_to_exit_node: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_connect_to_exit_node_postquantum(uniffiStatus)
	})
	if checksum != 49599 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_connect_to_exit_node_postquantum: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_connect_to_exit_node_with_id(uniffiStatus)
	})
	if checksum != 42608 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_connect_to_exit_node_with_id: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_disable_magic_dns(uniffiStatus)
	})
	if checksum != 30932 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_disable_magic_dns: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_disconnect_from_exit_node(uniffiStatus)
	})
	if checksum != 4404 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_disconnect_from_exit_node: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_disconnect_from_exit_nodes(uniffiStatus)
	})
	if checksum != 33222 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_disconnect_from_exit_nodes: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_enable_magic_dns(uniffiStatus)
	})
	if checksum != 31729 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_enable_magic_dns: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_generate_stack_panic(uniffiStatus)
	})
	if checksum != 15629 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_generate_stack_panic: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_generate_thread_panic(uniffiStatus)
	})
	if checksum != 37036 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_generate_thread_panic: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_get_adapter_luid(uniffiStatus)
	})
	if checksum != 53187 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_get_adapter_luid: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_get_last_error(uniffiStatus)
	})
	if checksum != 1246 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_get_last_error: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_get_nat(uniffiStatus)
	})
	if checksum != 11642 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_get_nat: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_get_secret_key(uniffiStatus)
	})
	if checksum != 35090 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_get_secret_key: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_get_status_map(uniffiStatus)
	})
	if checksum != 58739 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_get_status_map: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_is_running(uniffiStatus)
	})
	if checksum != 44343 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_is_running: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_notify_network_change(uniffiStatus)
	})
	if checksum != 56036 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_notify_network_change: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_notify_sleep(uniffiStatus)
	})
	if checksum != 46674 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_notify_sleep: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_notify_wakeup(uniffiStatus)
	})
	if checksum != 23592 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_notify_wakeup: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_probe_pmtu(uniffiStatus)
	})
	if checksum != 28113 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_probe_pmtu: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_receive_ping(uniffiStatus)
	})
	if checksum != 127 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_receive_ping: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_set_fwmark(uniffiStatus)
	})
	if checksum != 52541 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_set_fwmark: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_set_meshnet(uniffiStatus)
	})
	if checksum != 21583 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_set_meshnet: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_set_meshnet_off(uniffiStatus)
	})
	if checksum != 32794 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_set_meshnet_off: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_set_secret_key(uniffiStatus)
	})
	if checksum != 4273 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_set_secret_key: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_shutdown(uniffiStatus)
	})
	if checksum != 25385 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_shutdown: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_shutdown_hard(uniffiStatus)
	})
	if checksum != 46450 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_shutdown_hard: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_start(uniffiStatus)
	})
	if checksum != 30743 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_start: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_start_named(uniffiStatus)
	})
	if checksum != 50320 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_start_named: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_start_with_tun(uniffiStatus)
	})
	if checksum != 57601 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_start_with_tun: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_stop(uniffiStatus)
	})
	if checksum != 44700 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_stop: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_trigger_analytics_event(uniffiStatus)
	})
	if checksum != 54857 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_trigger_analytics_event: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telio_trigger_qos_collection(uniffiStatus)
	})
	if checksum != 37519 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telio_trigger_qos_collection: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_constructor_featuresdefaultsbuilder_new(uniffiStatus)
	})
	if checksum != 9604 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_constructor_featuresdefaultsbuilder_new: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_constructor_telio_new(uniffiStatus)
	})
	if checksum != 21500 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_constructor_telio_new: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_constructor_telio_new_with_protect(uniffiStatus)
	})
	if checksum != 35715 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_constructor_telio_new_with_protect: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telioeventcb_event(uniffiStatus)
	})
	if checksum != 10177 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_telioeventcb_event: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_teliologgercb_log(uniffiStatus)
	})
	if checksum != 46379 {
		// If this happens try cleaning and rebuilding your project
		panic("telio: uniffi_telio_checksum_method_teliologgercb_log: UniFFI API checksum mismatch")
	}
	}
	{
	checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_telio_checksum_method_telioprotectcb_protect(uniffiStatus)
	})
	if checksum != 63197 {
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
	if err != nil {
		panic(err)
	}
	if read_length != int(length) {
		panic(fmt.Errorf("bad read length when reading string, expected %d, read %d", length, read_length))
	}
	return string(buffer)
}

func (FfiConverterString) Lower(value string) RustBuffer {
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
	freeFunction func(unsafe.Pointer, *C.RustCallStatus)
	destroyed atomic.Bool
}

func newFfiObject(pointer unsafe.Pointer, freeFunction func(unsafe.Pointer, *C.RustCallStatus)) FfiObject {
	return FfiObject {
		pointer: pointer,
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

	return ffiObject.pointer
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

type FeaturesDefaultsBuilder struct {
	ffiObject FfiObject
}
// Create a builder for Features with minimal defaults.
func NewFeaturesDefaultsBuilder() *FeaturesDefaultsBuilder {
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_constructor_featuresdefaultsbuilder_new( _uniffiStatus)
	}))
}




// Build final config
func (_self *FeaturesDefaultsBuilder)Build() Features {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeFeaturesINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_build(
		_pointer, _uniffiStatus)
	}))
}


// Enable keepalive batching feature
func (_self *FeaturesDefaultsBuilder)EnableBatching() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_batching(
		_pointer, _uniffiStatus)
	}))
}


// Enable default wireguard timings, derp timings and other features for best battery performance
func (_self *FeaturesDefaultsBuilder)EnableBatterySavingDefaults() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_battery_saving_defaults(
		_pointer, _uniffiStatus)
	}))
}


// Enable direct connections with defaults;
func (_self *FeaturesDefaultsBuilder)EnableDirect() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_direct(
		_pointer, _uniffiStatus)
	}))
}


// Enable firewall connection resets when boringtun is used
func (_self *FeaturesDefaultsBuilder)EnableFirewallConnectionReset() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_firewall_connection_reset(
		_pointer, _uniffiStatus)
	}))
}


// Enable blocking event flush with timout on stop with defaults
func (_self *FeaturesDefaultsBuilder)EnableFlushEventsOnStopTimeoutSeconds() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_flush_events_on_stop_timeout_seconds(
		_pointer, _uniffiStatus)
	}))
}


// Enable IPv6 with defaults
func (_self *FeaturesDefaultsBuilder)EnableIpv6() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_ipv6(
		_pointer, _uniffiStatus)
	}))
}


// Enable lana, this requires input from apps
func (_self *FeaturesDefaultsBuilder)EnableLana(eventPath string, isProd bool) *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_lana(
		_pointer,FfiConverterStringINSTANCE.Lower(eventPath), FfiConverterBoolINSTANCE.Lower(isProd), _uniffiStatus)
	}))
}


// Enable Link detection mechanism with defaults
func (_self *FeaturesDefaultsBuilder)EnableLinkDetection() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_link_detection(
		_pointer, _uniffiStatus)
	}))
}


// Eanable multicast with defaults
func (_self *FeaturesDefaultsBuilder)EnableMulticast() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_multicast(
		_pointer, _uniffiStatus)
	}))
}


// Enable nicknames with defaults
func (_self *FeaturesDefaultsBuilder)EnableNicknames() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_nicknames(
		_pointer, _uniffiStatus)
	}))
}


// Enable nurse with defaults
func (_self *FeaturesDefaultsBuilder)EnableNurse() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_nurse(
		_pointer, _uniffiStatus)
	}))
}


// Enable PMTU discovery with defaults;
func (_self *FeaturesDefaultsBuilder)EnablePmtuDiscovery() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_pmtu_discovery(
		_pointer, _uniffiStatus)
	}))
}


// Enable key valiation in set_config call with defaults
func (_self *FeaturesDefaultsBuilder)EnableValidateKeys() *FeaturesDefaultsBuilder {
	_pointer := _self.ffiObject.incrementPointer("*FeaturesDefaultsBuilder")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterFeaturesDefaultsBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_method_featuresdefaultsbuilder_enable_validate_keys(
		_pointer, _uniffiStatus)
	}))
}



func (object *FeaturesDefaultsBuilder)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterFeaturesDefaultsBuilder struct {}

var FfiConverterFeaturesDefaultsBuilderINSTANCE = FfiConverterFeaturesDefaultsBuilder{}

func (c FfiConverterFeaturesDefaultsBuilder) Lift(pointer unsafe.Pointer) *FeaturesDefaultsBuilder {
	result := &FeaturesDefaultsBuilder {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_telio_fn_free_featuresdefaultsbuilder(pointer, status)
		}),
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


type Telio struct {
	ffiObject FfiObject
}
// Create new telio library instance
// # Parameters
// - `events`:     Events callback
// - `features`:   JSON string of enabled features
func NewTelio(features Features, events TelioEventCb) (*Telio, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_constructor_telio_new(FfiConverterTypeFeaturesINSTANCE.Lower(features), FfiConverterCallbackInterfaceTelioEventCbINSTANCE.Lower(events), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Telio
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTelioINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


// Create new telio library instance
// # Parameters
// - `events`:     Events callback
// - `features`:   JSON string of enabled features
// - `protect`:    Callback executed after exit-node connect (for VpnService::protectFromVpn())
func TelioNewWithProtect(features Features, events TelioEventCb, protect TelioProtectCb) (*Telio, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_telio_fn_constructor_telio_new_with_protect(FfiConverterTypeFeaturesINSTANCE.Lower(features), FfiConverterCallbackInterfaceTelioEventCbINSTANCE.Lower(events), FfiConverterCallbackInterfaceTelioProtectCbINSTANCE.Lower(protect), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue *Telio
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTelioINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}



// Wrapper for `telio_connect_to_exit_node_with_id` that doesn't take an identifier
func (_self *Telio)ConnectToExitNode(publicKey PublicKey, allowedIps *[]IpNet, endpoint *SocketAddr) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_connect_to_exit_node(
		_pointer,FfiConverterTypePublicKeyINSTANCE.Lower(publicKey), FfiConverterOptionalSequenceTypeIpNetINSTANCE.Lower(allowedIps), FfiConverterOptionalTypeSocketAddrINSTANCE.Lower(endpoint), _uniffiStatus)
		return false
	})
		return _uniffiErr
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

func (_self *Telio)ConnectToExitNodePostquantum(identifier *string, publicKey PublicKey, allowedIps *[]IpNet, endpoint SocketAddr) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_connect_to_exit_node_postquantum(
		_pointer,FfiConverterOptionalStringINSTANCE.Lower(identifier), FfiConverterTypePublicKeyINSTANCE.Lower(publicKey), FfiConverterOptionalSequenceTypeIpNetINSTANCE.Lower(allowedIps), FfiConverterTypeSocketAddrINSTANCE.Lower(endpoint), _uniffiStatus)
		return false
	})
		return _uniffiErr
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

func (_self *Telio)ConnectToExitNodeWithId(identifier *string, publicKey PublicKey, allowedIps *[]IpNet, endpoint *SocketAddr) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_connect_to_exit_node_with_id(
		_pointer,FfiConverterOptionalStringINSTANCE.Lower(identifier), FfiConverterTypePublicKeyINSTANCE.Lower(publicKey), FfiConverterOptionalSequenceTypeIpNetINSTANCE.Lower(allowedIps), FfiConverterOptionalTypeSocketAddrINSTANCE.Lower(endpoint), _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// Disables magic DNS if it was enabled.
func (_self *Telio)DisableMagicDns() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_disable_magic_dns(
		_pointer, _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// Disconnects from specified exit node.
//
// # Parameters
// - `public_key`: WireGuard public key for exit node.

func (_self *Telio)DisconnectFromExitNode(publicKey PublicKey) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_disconnect_from_exit_node(
		_pointer,FfiConverterTypePublicKeyINSTANCE.Lower(publicKey), _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// Disconnects from all exit nodes with no parameters required.
func (_self *Telio)DisconnectFromExitNodes() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_disconnect_from_exit_nodes(
		_pointer, _uniffiStatus)
		return false
	})
		return _uniffiErr
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
func (_self *Telio)EnableMagicDns(forwardServers []IpAddr) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_enable_magic_dns(
		_pointer,FfiConverterSequenceTypeIpAddrINSTANCE.Lower(forwardServers), _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// For testing only.
func (_self *Telio)GenerateStackPanic() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_generate_stack_panic(
		_pointer, _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// For testing only.
func (_self *Telio)GenerateThreadPanic() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_generate_thread_panic(
		_pointer, _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// get device luid.
func (_self *Telio)GetAdapterLuid() uint64 {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterUint64INSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint64_t {
		return C.uniffi_telio_fn_method_telio_get_adapter_luid(
		_pointer, _uniffiStatus)
	}))
}


// Get last error's message length, including trailing null
func (_self *Telio)GetLastError() string {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterStringINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_telio_fn_method_telio_get_last_error(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Telio)GetNat(ip string, port uint16) (NatType, error) {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_telio_fn_method_telio_get_nat(
		_pointer,FfiConverterStringINSTANCE.Lower(ip), FfiConverterUint16INSTANCE.Lower(port), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue NatType
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTypeNatTypeINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *Telio)GetSecretKey() SecretKey {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeSecretKeyINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_telio_fn_method_telio_get_secret_key(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Telio)GetStatusMap() []TelioNode {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterSequenceTypeTelioNodeINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_telio_fn_method_telio_get_status_map(
		_pointer, _uniffiStatus)
	}))
}


func (_self *Telio)IsRunning() (bool, error) {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_telio_fn_method_telio_is_running(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue bool
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterBoolINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


// Notify telio with network state changes.
//
// # Parameters
// - `network_info`: Json encoded network sate info.
//                   Format to be decided, pass empty string for now.
func (_self *Telio)NotifyNetworkChange(networkInfo string) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_notify_network_change(
		_pointer,FfiConverterStringINSTANCE.Lower(networkInfo), _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// Notify telio when system goes to sleep.
func (_self *Telio)NotifySleep() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_notify_sleep(
		_pointer, _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// Notify telio when system wakes up.
func (_self *Telio)NotifyWakeup() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_notify_wakeup(
		_pointer, _uniffiStatus)
		return false
	})
		return _uniffiErr
}


func (_self *Telio)ProbePmtu(host string) (uint32, error) {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) C.uint32_t {
		return C.uniffi_telio_fn_method_telio_probe_pmtu(
		_pointer,FfiConverterStringINSTANCE.Lower(host), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue uint32
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterUint32INSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


func (_self *Telio)ReceivePing() (string, error) {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_telio_fn_method_telio_receive_ping(
		_pointer, _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue string
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterStringINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}


// Sets fmark for started device.
//
// # Parameters
// - `fwmark`: unsigned 32-bit integer

func (_self *Telio)SetFwmark(fwmark uint32) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_set_fwmark(
		_pointer,FfiConverterUint32INSTANCE.Lower(fwmark), _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// Enables meshnet if it is not enabled yet.
// In case meshnet is enabled, this updates the peer map with the specified one.
//
// # Parameters
// - `cfg`: Output of GET /v1/meshnet/machines/{machineIdentifier}/map

func (_self *Telio)SetMeshnet(cfg Config) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_set_meshnet(
		_pointer,FfiConverterTypeConfigINSTANCE.Lower(cfg), _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// Disables the meshnet functionality by closing all the connections.
func (_self *Telio)SetMeshnetOff() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_set_meshnet_off(
		_pointer, _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// Sets private key for started device.
//
// If private_key is not set, device will never connect.
//
// # Parameters
// - `private_key`: WireGuard private key.

func (_self *Telio)SetSecretKey(secretKey SecretKey) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_set_secret_key(
		_pointer,FfiConverterTypeSecretKeyINSTANCE.Lower(secretKey), _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// Completely stop and uninit telio lib.
func (_self *Telio)Shutdown() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_shutdown(
		_pointer, _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// Explicitly deallocate telio object and shutdown async rt.
func (_self *Telio)ShutdownHard() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_shutdown_hard(
		_pointer, _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// Start telio with specified adapter.
//
// Adapter will attempt to open its own tunnel.
func (_self *Telio)Start(secretKey SecretKey, adapter TelioAdapterType) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_start(
		_pointer,FfiConverterTypeSecretKeyINSTANCE.Lower(secretKey), FfiConverterTypeTelioAdapterTypeINSTANCE.Lower(adapter), _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// Start telio with specified adapter and name.
//
// Adapter will attempt to open its own tunnel.
func (_self *Telio)StartNamed(secretKey SecretKey, adapter TelioAdapterType, name string) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_start_named(
		_pointer,FfiConverterTypeSecretKeyINSTANCE.Lower(secretKey), FfiConverterTypeTelioAdapterTypeINSTANCE.Lower(adapter), FfiConverterStringINSTANCE.Lower(name), _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// Start telio device with specified adapter and already open tunnel.
//
// Telio will take ownership of tunnel , and close it on stop.
//
// # Parameters
// - `private_key`: base64 encoded private_key.
// - `adapter`: Adapter type.
// - `tun`: A valid filedescriptor to tun device.

func (_self *Telio)StartWithTun(secretKey SecretKey, adapter TelioAdapterType, tun int32) error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_start_with_tun(
		_pointer,FfiConverterTypeSecretKeyINSTANCE.Lower(secretKey), FfiConverterTypeTelioAdapterTypeINSTANCE.Lower(adapter), FfiConverterInt32INSTANCE.Lower(tun), _uniffiStatus)
		return false
	})
		return _uniffiErr
}


// Stop telio device.
func (_self *Telio)Stop() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_stop(
		_pointer, _uniffiStatus)
		return false
	})
		return _uniffiErr
}


func (_self *Telio)TriggerAnalyticsEvent() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_trigger_analytics_event(
		_pointer, _uniffiStatus)
		return false
	})
		return _uniffiErr
}


func (_self *Telio)TriggerQosCollection() error {
	_pointer := _self.ffiObject.incrementPointer("*Telio")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_method_telio_trigger_qos_collection(
		_pointer, _uniffiStatus)
		return false
	})
		return _uniffiErr
}



func (object *Telio)Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterTelio struct {}

var FfiConverterTelioINSTANCE = FfiConverterTelio{}

func (c FfiConverterTelio) Lift(pointer unsafe.Pointer) *Telio {
	result := &Telio {
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_telio_fn_free_telio(pointer, status)
		}),
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
		FfiDestroyerTypePeerBase{}.Destroy(r.This);
		FfiDestroyerOptionalSequenceTypePeer{}.Destroy(r.Peers);
		FfiDestroyerOptionalSequenceTypeServer{}.Destroy(r.DerpServers);
		FfiDestroyerOptionalTypeDnsConfig{}.Destroy(r.Dns);
}

type FfiConverterTypeConfig struct {}

var FfiConverterTypeConfigINSTANCE = FfiConverterTypeConfig{}

func (c FfiConverterTypeConfig) Lift(rb RustBufferI) Config {
	return LiftFromRustBuffer[Config](c, rb)
}

func (c FfiConverterTypeConfig) Read(reader io.Reader) Config {
	return Config {
			FfiConverterTypePeerBaseINSTANCE.Read(reader),
			FfiConverterOptionalSequenceTypePeerINSTANCE.Read(reader),
			FfiConverterOptionalSequenceTypeServerINSTANCE.Read(reader),
			FfiConverterOptionalTypeDnsConfigINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeConfig) Lower(value Config) RustBuffer {
	return LowerIntoRustBuffer[Config](c, value)
}

func (c FfiConverterTypeConfig) Write(writer io.Writer, value Config) {
		FfiConverterTypePeerBaseINSTANCE.Write(writer, value.This);
		FfiConverterOptionalSequenceTypePeerINSTANCE.Write(writer, value.Peers);
		FfiConverterOptionalSequenceTypeServerINSTANCE.Write(writer, value.DerpServers);
		FfiConverterOptionalTypeDnsConfigINSTANCE.Write(writer, value.Dns);
}

type FfiDestroyerTypeConfig struct {}

func (_ FfiDestroyerTypeConfig) Destroy(value Config) {
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

type FfiConverterTypeDnsConfig struct {}

var FfiConverterTypeDnsConfigINSTANCE = FfiConverterTypeDnsConfig{}

func (c FfiConverterTypeDnsConfig) Lift(rb RustBufferI) DnsConfig {
	return LiftFromRustBuffer[DnsConfig](c, rb)
}

func (c FfiConverterTypeDnsConfig) Read(reader io.Reader) DnsConfig {
	return DnsConfig {
			FfiConverterOptionalSequenceTypeIpAddrINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeDnsConfig) Lower(value DnsConfig) RustBuffer {
	return LowerIntoRustBuffer[DnsConfig](c, value)
}

func (c FfiConverterTypeDnsConfig) Write(writer io.Writer, value DnsConfig) {
		FfiConverterOptionalSequenceTypeIpAddrINSTANCE.Write(writer, value.DnsServers);
}

type FfiDestroyerTypeDnsConfig struct {}

func (_ FfiDestroyerTypeDnsConfig) Destroy(value DnsConfig) {
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
		FfiDestroyerTypeErrorLevel{}.Destroy(r.Level);
		FfiDestroyerTypeErrorCode{}.Destroy(r.Code);
		FfiDestroyerString{}.Destroy(r.Msg);
}

type FfiConverterTypeErrorEvent struct {}

var FfiConverterTypeErrorEventINSTANCE = FfiConverterTypeErrorEvent{}

func (c FfiConverterTypeErrorEvent) Lift(rb RustBufferI) ErrorEvent {
	return LiftFromRustBuffer[ErrorEvent](c, rb)
}

func (c FfiConverterTypeErrorEvent) Read(reader io.Reader) ErrorEvent {
	return ErrorEvent {
			FfiConverterTypeErrorLevelINSTANCE.Read(reader),
			FfiConverterTypeErrorCodeINSTANCE.Read(reader),
			FfiConverterStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeErrorEvent) Lower(value ErrorEvent) RustBuffer {
	return LowerIntoRustBuffer[ErrorEvent](c, value)
}

func (c FfiConverterTypeErrorEvent) Write(writer io.Writer, value ErrorEvent) {
		FfiConverterTypeErrorLevelINSTANCE.Write(writer, value.Level);
		FfiConverterTypeErrorCodeINSTANCE.Write(writer, value.Code);
		FfiConverterStringINSTANCE.Write(writer, value.Msg);
}

type FfiDestroyerTypeErrorEvent struct {}

func (_ FfiDestroyerTypeErrorEvent) Destroy(value ErrorEvent) {
	value.Destroy()
}


type FeatureBatching struct {
	// direct connection threshold for batching
	DirectConnectionThreshold uint32
}

func (r *FeatureBatching) Destroy() {
		FfiDestroyerUint32{}.Destroy(r.DirectConnectionThreshold);
}

type FfiConverterTypeFeatureBatching struct {}

var FfiConverterTypeFeatureBatchingINSTANCE = FfiConverterTypeFeatureBatching{}

func (c FfiConverterTypeFeatureBatching) Lift(rb RustBufferI) FeatureBatching {
	return LiftFromRustBuffer[FeatureBatching](c, rb)
}

func (c FfiConverterTypeFeatureBatching) Read(reader io.Reader) FeatureBatching {
	return FeatureBatching {
			FfiConverterUint32INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFeatureBatching) Lower(value FeatureBatching) RustBuffer {
	return LowerIntoRustBuffer[FeatureBatching](c, value)
}

func (c FfiConverterTypeFeatureBatching) Write(writer io.Writer, value FeatureBatching) {
		FfiConverterUint32INSTANCE.Write(writer, value.DirectConnectionThreshold);
}

type FfiDestroyerTypeFeatureBatching struct {}

func (_ FfiDestroyerTypeFeatureBatching) Destroy(value FeatureBatching) {
	value.Destroy()
}


// Configure derp behaviour
type FeatureDerp struct {
	// Tcp keepalive set on derp server's side [default 15s]
	TcpKeepalive *uint32
	// Derp will send empty messages after this many seconds of not sending/receiving any data [default 60s]
	DerpKeepalive *uint32
	// Enable polling of remote peer states to reduce derp traffic
	EnablePolling *bool
	// Use Mozilla's root certificates instead of OS ones [default false]
	UseBuiltInRootCertificates bool
}

func (r *FeatureDerp) Destroy() {
		FfiDestroyerOptionalUint32{}.Destroy(r.TcpKeepalive);
		FfiDestroyerOptionalUint32{}.Destroy(r.DerpKeepalive);
		FfiDestroyerOptionalBool{}.Destroy(r.EnablePolling);
		FfiDestroyerBool{}.Destroy(r.UseBuiltInRootCertificates);
}

type FfiConverterTypeFeatureDerp struct {}

var FfiConverterTypeFeatureDerpINSTANCE = FfiConverterTypeFeatureDerp{}

func (c FfiConverterTypeFeatureDerp) Lift(rb RustBufferI) FeatureDerp {
	return LiftFromRustBuffer[FeatureDerp](c, rb)
}

func (c FfiConverterTypeFeatureDerp) Read(reader io.Reader) FeatureDerp {
	return FeatureDerp {
			FfiConverterOptionalUint32INSTANCE.Read(reader),
			FfiConverterOptionalUint32INSTANCE.Read(reader),
			FfiConverterOptionalBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFeatureDerp) Lower(value FeatureDerp) RustBuffer {
	return LowerIntoRustBuffer[FeatureDerp](c, value)
}

func (c FfiConverterTypeFeatureDerp) Write(writer io.Writer, value FeatureDerp) {
		FfiConverterOptionalUint32INSTANCE.Write(writer, value.TcpKeepalive);
		FfiConverterOptionalUint32INSTANCE.Write(writer, value.DerpKeepalive);
		FfiConverterOptionalBoolINSTANCE.Write(writer, value.EnablePolling);
		FfiConverterBoolINSTANCE.Write(writer, value.UseBuiltInRootCertificates);
}

type FfiDestroyerTypeFeatureDerp struct {}

func (_ FfiDestroyerTypeFeatureDerp) Destroy(value FeatureDerp) {
	value.Destroy()
}


// Enable meshent direct connection
type FeatureDirect struct {
	// Endpoint providers [default all]
	Providers *EndpointProviders
	// Polling interval for endpoints [default 10s]
	EndpointIntervalSecs uint64
	// Configuration options for skipping unresponsive peers
	SkipUnresponsivePeers *FeatureSkipUnresponsivePeers
	// Parameters to optimize battery lifetime
	EndpointProvidersOptimization *FeatureEndpointProvidersOptimization
}

func (r *FeatureDirect) Destroy() {
		FfiDestroyerOptionalTypeEndpointProviders{}.Destroy(r.Providers);
		FfiDestroyerUint64{}.Destroy(r.EndpointIntervalSecs);
		FfiDestroyerOptionalTypeFeatureSkipUnresponsivePeers{}.Destroy(r.SkipUnresponsivePeers);
		FfiDestroyerOptionalTypeFeatureEndpointProvidersOptimization{}.Destroy(r.EndpointProvidersOptimization);
}

type FfiConverterTypeFeatureDirect struct {}

var FfiConverterTypeFeatureDirectINSTANCE = FfiConverterTypeFeatureDirect{}

func (c FfiConverterTypeFeatureDirect) Lift(rb RustBufferI) FeatureDirect {
	return LiftFromRustBuffer[FeatureDirect](c, rb)
}

func (c FfiConverterTypeFeatureDirect) Read(reader io.Reader) FeatureDirect {
	return FeatureDirect {
			FfiConverterOptionalTypeEndpointProvidersINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterOptionalTypeFeatureSkipUnresponsivePeersINSTANCE.Read(reader),
			FfiConverterOptionalTypeFeatureEndpointProvidersOptimizationINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFeatureDirect) Lower(value FeatureDirect) RustBuffer {
	return LowerIntoRustBuffer[FeatureDirect](c, value)
}

func (c FfiConverterTypeFeatureDirect) Write(writer io.Writer, value FeatureDirect) {
		FfiConverterOptionalTypeEndpointProvidersINSTANCE.Write(writer, value.Providers);
		FfiConverterUint64INSTANCE.Write(writer, value.EndpointIntervalSecs);
		FfiConverterOptionalTypeFeatureSkipUnresponsivePeersINSTANCE.Write(writer, value.SkipUnresponsivePeers);
		FfiConverterOptionalTypeFeatureEndpointProvidersOptimizationINSTANCE.Write(writer, value.EndpointProvidersOptimization);
}

type FfiDestroyerTypeFeatureDirect struct {}

func (_ FfiDestroyerTypeFeatureDirect) Destroy(value FeatureDirect) {
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
		FfiDestroyerOptionalTypeFeatureExitDns{}.Destroy(r.ExitDns);
}

type FfiConverterTypeFeatureDns struct {}

var FfiConverterTypeFeatureDnsINSTANCE = FfiConverterTypeFeatureDns{}

func (c FfiConverterTypeFeatureDns) Lift(rb RustBufferI) FeatureDns {
	return LiftFromRustBuffer[FeatureDns](c, rb)
}

func (c FfiConverterTypeFeatureDns) Read(reader io.Reader) FeatureDns {
	return FeatureDns {
			FfiConverterTypeTtlValueINSTANCE.Read(reader),
			FfiConverterOptionalTypeFeatureExitDnsINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFeatureDns) Lower(value FeatureDns) RustBuffer {
	return LowerIntoRustBuffer[FeatureDns](c, value)
}

func (c FfiConverterTypeFeatureDns) Write(writer io.Writer, value FeatureDns) {
		FfiConverterTypeTtlValueINSTANCE.Write(writer, value.TtlValue);
		FfiConverterOptionalTypeFeatureExitDnsINSTANCE.Write(writer, value.ExitDns);
}

type FfiDestroyerTypeFeatureDns struct {}

func (_ FfiDestroyerTypeFeatureDns) Destroy(value FeatureDns) {
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

type FfiConverterTypeFeatureEndpointProvidersOptimization struct {}

var FfiConverterTypeFeatureEndpointProvidersOptimizationINSTANCE = FfiConverterTypeFeatureEndpointProvidersOptimization{}

func (c FfiConverterTypeFeatureEndpointProvidersOptimization) Lift(rb RustBufferI) FeatureEndpointProvidersOptimization {
	return LiftFromRustBuffer[FeatureEndpointProvidersOptimization](c, rb)
}

func (c FfiConverterTypeFeatureEndpointProvidersOptimization) Read(reader io.Reader) FeatureEndpointProvidersOptimization {
	return FeatureEndpointProvidersOptimization {
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFeatureEndpointProvidersOptimization) Lower(value FeatureEndpointProvidersOptimization) RustBuffer {
	return LowerIntoRustBuffer[FeatureEndpointProvidersOptimization](c, value)
}

func (c FfiConverterTypeFeatureEndpointProvidersOptimization) Write(writer io.Writer, value FeatureEndpointProvidersOptimization) {
		FfiConverterBoolINSTANCE.Write(writer, value.OptimizeDirectUpgradeStun);
		FfiConverterBoolINSTANCE.Write(writer, value.OptimizeDirectUpgradeUpnp);
}

type FfiDestroyerTypeFeatureEndpointProvidersOptimization struct {}

func (_ FfiDestroyerTypeFeatureEndpointProvidersOptimization) Destroy(value FeatureEndpointProvidersOptimization) {
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

type FfiConverterTypeFeatureExitDns struct {}

var FfiConverterTypeFeatureExitDnsINSTANCE = FfiConverterTypeFeatureExitDns{}

func (c FfiConverterTypeFeatureExitDns) Lift(rb RustBufferI) FeatureExitDns {
	return LiftFromRustBuffer[FeatureExitDns](c, rb)
}

func (c FfiConverterTypeFeatureExitDns) Read(reader io.Reader) FeatureExitDns {
	return FeatureExitDns {
			FfiConverterOptionalBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFeatureExitDns) Lower(value FeatureExitDns) RustBuffer {
	return LowerIntoRustBuffer[FeatureExitDns](c, value)
}

func (c FfiConverterTypeFeatureExitDns) Write(writer io.Writer, value FeatureExitDns) {
		FfiConverterOptionalBoolINSTANCE.Write(writer, value.AutoSwitchDnsIps);
}

type FfiDestroyerTypeFeatureExitDns struct {}

func (_ FfiDestroyerTypeFeatureExitDns) Destroy(value FeatureExitDns) {
	value.Destroy()
}


// Feature config for firewall
type FeatureFirewall struct {
	// Turns on connection resets upon VPN server change
	BoringtunResetConns bool
}

func (r *FeatureFirewall) Destroy() {
		FfiDestroyerBool{}.Destroy(r.BoringtunResetConns);
}

type FfiConverterTypeFeatureFirewall struct {}

var FfiConverterTypeFeatureFirewallINSTANCE = FfiConverterTypeFeatureFirewall{}

func (c FfiConverterTypeFeatureFirewall) Lift(rb RustBufferI) FeatureFirewall {
	return LiftFromRustBuffer[FeatureFirewall](c, rb)
}

func (c FfiConverterTypeFeatureFirewall) Read(reader io.Reader) FeatureFirewall {
	return FeatureFirewall {
			FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFeatureFirewall) Lower(value FeatureFirewall) RustBuffer {
	return LowerIntoRustBuffer[FeatureFirewall](c, value)
}

func (c FfiConverterTypeFeatureFirewall) Write(writer io.Writer, value FeatureFirewall) {
		FfiConverterBoolINSTANCE.Write(writer, value.BoringtunResetConns);
}

type FfiDestroyerTypeFeatureFirewall struct {}

func (_ FfiDestroyerTypeFeatureFirewall) Destroy(value FeatureFirewall) {
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

type FfiConverterTypeFeatureLana struct {}

var FfiConverterTypeFeatureLanaINSTANCE = FfiConverterTypeFeatureLana{}

func (c FfiConverterTypeFeatureLana) Lift(rb RustBufferI) FeatureLana {
	return LiftFromRustBuffer[FeatureLana](c, rb)
}

func (c FfiConverterTypeFeatureLana) Read(reader io.Reader) FeatureLana {
	return FeatureLana {
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFeatureLana) Lower(value FeatureLana) RustBuffer {
	return LowerIntoRustBuffer[FeatureLana](c, value)
}

func (c FfiConverterTypeFeatureLana) Write(writer io.Writer, value FeatureLana) {
		FfiConverterStringINSTANCE.Write(writer, value.EventPath);
		FfiConverterBoolINSTANCE.Write(writer, value.Prod);
}

type FfiDestroyerTypeFeatureLana struct {}

func (_ FfiDestroyerTypeFeatureLana) Destroy(value FeatureLana) {
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

type FfiConverterTypeFeatureLinkDetection struct {}

var FfiConverterTypeFeatureLinkDetectionINSTANCE = FfiConverterTypeFeatureLinkDetection{}

func (c FfiConverterTypeFeatureLinkDetection) Lift(rb RustBufferI) FeatureLinkDetection {
	return LiftFromRustBuffer[FeatureLinkDetection](c, rb)
}

func (c FfiConverterTypeFeatureLinkDetection) Read(reader io.Reader) FeatureLinkDetection {
	return FeatureLinkDetection {
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFeatureLinkDetection) Lower(value FeatureLinkDetection) RustBuffer {
	return LowerIntoRustBuffer[FeatureLinkDetection](c, value)
}

func (c FfiConverterTypeFeatureLinkDetection) Write(writer io.Writer, value FeatureLinkDetection) {
		FfiConverterUint64INSTANCE.Write(writer, value.RttSeconds);
		FfiConverterUint32INSTANCE.Write(writer, value.NoOfPings);
		FfiConverterBoolINSTANCE.Write(writer, value.UseForDowngrade);
}

type FfiDestroyerTypeFeatureLinkDetection struct {}

func (_ FfiDestroyerTypeFeatureLinkDetection) Destroy(value FeatureLinkDetection) {
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
	// Enable/disable collecting nat type
	EnableNatTypeCollection bool
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
		FfiDestroyerOptionalTypeFeatureQoS{}.Destroy(r.Qos);
		FfiDestroyerBool{}.Destroy(r.EnableNatTypeCollection);
		FfiDestroyerBool{}.Destroy(r.EnableRelayConnData);
		FfiDestroyerBool{}.Destroy(r.EnableNatTraversalConnData);
		FfiDestroyerUint64{}.Destroy(r.StateDurationCap);
}

type FfiConverterTypeFeatureNurse struct {}

var FfiConverterTypeFeatureNurseINSTANCE = FfiConverterTypeFeatureNurse{}

func (c FfiConverterTypeFeatureNurse) Lift(rb RustBufferI) FeatureNurse {
	return LiftFromRustBuffer[FeatureNurse](c, rb)
}

func (c FfiConverterTypeFeatureNurse) Read(reader io.Reader) FeatureNurse {
	return FeatureNurse {
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterOptionalTypeFeatureQoSINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFeatureNurse) Lower(value FeatureNurse) RustBuffer {
	return LowerIntoRustBuffer[FeatureNurse](c, value)
}

func (c FfiConverterTypeFeatureNurse) Write(writer io.Writer, value FeatureNurse) {
		FfiConverterUint64INSTANCE.Write(writer, value.HeartbeatInterval);
		FfiConverterUint64INSTANCE.Write(writer, value.InitialHeartbeatInterval);
		FfiConverterOptionalTypeFeatureQoSINSTANCE.Write(writer, value.Qos);
		FfiConverterBoolINSTANCE.Write(writer, value.EnableNatTypeCollection);
		FfiConverterBoolINSTANCE.Write(writer, value.EnableRelayConnData);
		FfiConverterBoolINSTANCE.Write(writer, value.EnableNatTraversalConnData);
		FfiConverterUint64INSTANCE.Write(writer, value.StateDurationCap);
}

type FfiDestroyerTypeFeatureNurse struct {}

func (_ FfiDestroyerTypeFeatureNurse) Destroy(value FeatureNurse) {
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
		FfiDestroyerSequenceTypePathType{}.Destroy(r.Priority);
		FfiDestroyerOptionalTypePathType{}.Destroy(r.Force);
}

type FfiConverterTypeFeaturePaths struct {}

var FfiConverterTypeFeaturePathsINSTANCE = FfiConverterTypeFeaturePaths{}

func (c FfiConverterTypeFeaturePaths) Lift(rb RustBufferI) FeaturePaths {
	return LiftFromRustBuffer[FeaturePaths](c, rb)
}

func (c FfiConverterTypeFeaturePaths) Read(reader io.Reader) FeaturePaths {
	return FeaturePaths {
			FfiConverterSequenceTypePathTypeINSTANCE.Read(reader),
			FfiConverterOptionalTypePathTypeINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFeaturePaths) Lower(value FeaturePaths) RustBuffer {
	return LowerIntoRustBuffer[FeaturePaths](c, value)
}

func (c FfiConverterTypeFeaturePaths) Write(writer io.Writer, value FeaturePaths) {
		FfiConverterSequenceTypePathTypeINSTANCE.Write(writer, value.Priority);
		FfiConverterOptionalTypePathTypeINSTANCE.Write(writer, value.Force);
}

type FfiDestroyerTypeFeaturePaths struct {}

func (_ FfiDestroyerTypeFeaturePaths) Destroy(value FeaturePaths) {
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

type FfiConverterTypeFeaturePersistentKeepalive struct {}

var FfiConverterTypeFeaturePersistentKeepaliveINSTANCE = FfiConverterTypeFeaturePersistentKeepalive{}

func (c FfiConverterTypeFeaturePersistentKeepalive) Lift(rb RustBufferI) FeaturePersistentKeepalive {
	return LiftFromRustBuffer[FeaturePersistentKeepalive](c, rb)
}

func (c FfiConverterTypeFeaturePersistentKeepalive) Read(reader io.Reader) FeaturePersistentKeepalive {
	return FeaturePersistentKeepalive {
			FfiConverterOptionalUint32INSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterOptionalUint32INSTANCE.Read(reader),
			FfiConverterOptionalUint32INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFeaturePersistentKeepalive) Lower(value FeaturePersistentKeepalive) RustBuffer {
	return LowerIntoRustBuffer[FeaturePersistentKeepalive](c, value)
}

func (c FfiConverterTypeFeaturePersistentKeepalive) Write(writer io.Writer, value FeaturePersistentKeepalive) {
		FfiConverterOptionalUint32INSTANCE.Write(writer, value.Vpn);
		FfiConverterUint32INSTANCE.Write(writer, value.Direct);
		FfiConverterOptionalUint32INSTANCE.Write(writer, value.Proxying);
		FfiConverterOptionalUint32INSTANCE.Write(writer, value.Stun);
}

type FfiDestroyerTypeFeaturePersistentKeepalive struct {}

func (_ FfiDestroyerTypeFeaturePersistentKeepalive) Destroy(value FeaturePersistentKeepalive) {
	value.Destroy()
}


// PMTU discovery configuration for VPN connection
type FeaturePmtuDiscovery struct {
	// A timeout for wait for the ICMP response packet
	ResponseWaitTimeoutS uint32
}

func (r *FeaturePmtuDiscovery) Destroy() {
		FfiDestroyerUint32{}.Destroy(r.ResponseWaitTimeoutS);
}

type FfiConverterTypeFeaturePmtuDiscovery struct {}

var FfiConverterTypeFeaturePmtuDiscoveryINSTANCE = FfiConverterTypeFeaturePmtuDiscovery{}

func (c FfiConverterTypeFeaturePmtuDiscovery) Lift(rb RustBufferI) FeaturePmtuDiscovery {
	return LiftFromRustBuffer[FeaturePmtuDiscovery](c, rb)
}

func (c FfiConverterTypeFeaturePmtuDiscovery) Read(reader io.Reader) FeaturePmtuDiscovery {
	return FeaturePmtuDiscovery {
			FfiConverterUint32INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFeaturePmtuDiscovery) Lower(value FeaturePmtuDiscovery) RustBuffer {
	return LowerIntoRustBuffer[FeaturePmtuDiscovery](c, value)
}

func (c FfiConverterTypeFeaturePmtuDiscovery) Write(writer io.Writer, value FeaturePmtuDiscovery) {
		FfiConverterUint32INSTANCE.Write(writer, value.ResponseWaitTimeoutS);
}

type FfiDestroyerTypeFeaturePmtuDiscovery struct {}

func (_ FfiDestroyerTypeFeaturePmtuDiscovery) Destroy(value FeaturePmtuDiscovery) {
	value.Destroy()
}


// Turns on post quantum VPN tunnel
type FeaturePostQuantumVpn struct {
	// Initial handshake retry interval in seconds
	HandshakeRetryIntervalS uint32
	// Rekey interval in seconds
	RekeyIntervalS uint32
}

func (r *FeaturePostQuantumVpn) Destroy() {
		FfiDestroyerUint32{}.Destroy(r.HandshakeRetryIntervalS);
		FfiDestroyerUint32{}.Destroy(r.RekeyIntervalS);
}

type FfiConverterTypeFeaturePostQuantumVPN struct {}

var FfiConverterTypeFeaturePostQuantumVPNINSTANCE = FfiConverterTypeFeaturePostQuantumVPN{}

func (c FfiConverterTypeFeaturePostQuantumVPN) Lift(rb RustBufferI) FeaturePostQuantumVpn {
	return LiftFromRustBuffer[FeaturePostQuantumVpn](c, rb)
}

func (c FfiConverterTypeFeaturePostQuantumVPN) Read(reader io.Reader) FeaturePostQuantumVpn {
	return FeaturePostQuantumVpn {
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFeaturePostQuantumVPN) Lower(value FeaturePostQuantumVpn) RustBuffer {
	return LowerIntoRustBuffer[FeaturePostQuantumVpn](c, value)
}

func (c FfiConverterTypeFeaturePostQuantumVPN) Write(writer io.Writer, value FeaturePostQuantumVpn) {
		FfiConverterUint32INSTANCE.Write(writer, value.HandshakeRetryIntervalS);
		FfiConverterUint32INSTANCE.Write(writer, value.RekeyIntervalS);
}

type FfiDestroyerTypeFeaturePostQuantumVpn struct {}

func (_ FfiDestroyerTypeFeaturePostQuantumVpn) Destroy(value FeaturePostQuantumVpn) {
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
		FfiDestroyerSequenceTypeRttType{}.Destroy(r.RttTypes);
		FfiDestroyerUint32{}.Destroy(r.Buckets);
}

type FfiConverterTypeFeatureQoS struct {}

var FfiConverterTypeFeatureQoSINSTANCE = FfiConverterTypeFeatureQoS{}

func (c FfiConverterTypeFeatureQoS) Lift(rb RustBufferI) FeatureQoS {
	return LiftFromRustBuffer[FeatureQoS](c, rb)
}

func (c FfiConverterTypeFeatureQoS) Read(reader io.Reader) FeatureQoS {
	return FeatureQoS {
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterSequenceTypeRttTypeINSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFeatureQoS) Lower(value FeatureQoS) RustBuffer {
	return LowerIntoRustBuffer[FeatureQoS](c, value)
}

func (c FfiConverterTypeFeatureQoS) Write(writer io.Writer, value FeatureQoS) {
		FfiConverterUint64INSTANCE.Write(writer, value.RttInterval);
		FfiConverterUint32INSTANCE.Write(writer, value.RttTries);
		FfiConverterSequenceTypeRttTypeINSTANCE.Write(writer, value.RttTypes);
		FfiConverterUint32INSTANCE.Write(writer, value.Buckets);
}

type FfiDestroyerTypeFeatureQoS struct {}

func (_ FfiDestroyerTypeFeatureQoS) Destroy(value FeatureQoS) {
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

type FfiConverterTypeFeatureSkipUnresponsivePeers struct {}

var FfiConverterTypeFeatureSkipUnresponsivePeersINSTANCE = FfiConverterTypeFeatureSkipUnresponsivePeers{}

func (c FfiConverterTypeFeatureSkipUnresponsivePeers) Lift(rb RustBufferI) FeatureSkipUnresponsivePeers {
	return LiftFromRustBuffer[FeatureSkipUnresponsivePeers](c, rb)
}

func (c FfiConverterTypeFeatureSkipUnresponsivePeers) Read(reader io.Reader) FeatureSkipUnresponsivePeers {
	return FeatureSkipUnresponsivePeers {
			FfiConverterUint64INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFeatureSkipUnresponsivePeers) Lower(value FeatureSkipUnresponsivePeers) RustBuffer {
	return LowerIntoRustBuffer[FeatureSkipUnresponsivePeers](c, value)
}

func (c FfiConverterTypeFeatureSkipUnresponsivePeers) Write(writer io.Writer, value FeatureSkipUnresponsivePeers) {
		FfiConverterUint64INSTANCE.Write(writer, value.NoRxThresholdSecs);
}

type FfiDestroyerTypeFeatureSkipUnresponsivePeers struct {}

func (_ FfiDestroyerTypeFeatureSkipUnresponsivePeers) Destroy(value FeatureSkipUnresponsivePeers) {
	value.Destroy()
}


// Configurable features for Wireguard peers
type FeatureWireguard struct {
	// Configurable persistent keepalive periods for wireguard peers
	PersistentKeepalive FeaturePersistentKeepalive
}

func (r *FeatureWireguard) Destroy() {
		FfiDestroyerTypeFeaturePersistentKeepalive{}.Destroy(r.PersistentKeepalive);
}

type FfiConverterTypeFeatureWireguard struct {}

var FfiConverterTypeFeatureWireguardINSTANCE = FfiConverterTypeFeatureWireguard{}

func (c FfiConverterTypeFeatureWireguard) Lift(rb RustBufferI) FeatureWireguard {
	return LiftFromRustBuffer[FeatureWireguard](c, rb)
}

func (c FfiConverterTypeFeatureWireguard) Read(reader io.Reader) FeatureWireguard {
	return FeatureWireguard {
			FfiConverterTypeFeaturePersistentKeepaliveINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFeatureWireguard) Lower(value FeatureWireguard) RustBuffer {
	return LowerIntoRustBuffer[FeatureWireguard](c, value)
}

func (c FfiConverterTypeFeatureWireguard) Write(writer io.Writer, value FeatureWireguard) {
		FfiConverterTypeFeaturePersistentKeepaliveINSTANCE.Write(writer, value.PersistentKeepalive);
}

type FfiDestroyerTypeFeatureWireguard struct {}

func (_ FfiDestroyerTypeFeatureWireguard) Destroy(value FeatureWireguard) {
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
	// Controll if IP addresses should be hidden in logs
	HideIps bool
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
	// Enable PMTU discovery, enabled by default
	PmtuDiscovery *FeaturePmtuDiscovery
	// Multicast support
	Multicast bool
	// Batching
	Batching *FeatureBatching
}

func (r *Features) Destroy() {
		FfiDestroyerTypeFeatureWireguard{}.Destroy(r.Wireguard);
		FfiDestroyerOptionalTypeFeatureNurse{}.Destroy(r.Nurse);
		FfiDestroyerOptionalTypeFeatureLana{}.Destroy(r.Lana);
		FfiDestroyerOptionalTypeFeaturePaths{}.Destroy(r.Paths);
		FfiDestroyerOptionalTypeFeatureDirect{}.Destroy(r.Direct);
		FfiDestroyerOptionalBool{}.Destroy(r.IsTestEnv);
		FfiDestroyerBool{}.Destroy(r.HideIps);
		FfiDestroyerOptionalTypeFeatureDerp{}.Destroy(r.Derp);
		FfiDestroyerTypeFeatureValidateKeys{}.Destroy(r.ValidateKeys);
		FfiDestroyerBool{}.Destroy(r.Ipv6);
		FfiDestroyerBool{}.Destroy(r.Nicknames);
		FfiDestroyerTypeFeatureFirewall{}.Destroy(r.Firewall);
		FfiDestroyerOptionalUint64{}.Destroy(r.FlushEventsOnStopTimeoutSeconds);
		FfiDestroyerOptionalTypeFeatureLinkDetection{}.Destroy(r.LinkDetection);
		FfiDestroyerTypeFeatureDns{}.Destroy(r.Dns);
		FfiDestroyerTypeFeaturePostQuantumVpn{}.Destroy(r.PostQuantumVpn);
		FfiDestroyerOptionalTypeFeaturePmtuDiscovery{}.Destroy(r.PmtuDiscovery);
		FfiDestroyerBool{}.Destroy(r.Multicast);
		FfiDestroyerOptionalTypeFeatureBatching{}.Destroy(r.Batching);
}

type FfiConverterTypeFeatures struct {}

var FfiConverterTypeFeaturesINSTANCE = FfiConverterTypeFeatures{}

func (c FfiConverterTypeFeatures) Lift(rb RustBufferI) Features {
	return LiftFromRustBuffer[Features](c, rb)
}

func (c FfiConverterTypeFeatures) Read(reader io.Reader) Features {
	return Features {
			FfiConverterTypeFeatureWireguardINSTANCE.Read(reader),
			FfiConverterOptionalTypeFeatureNurseINSTANCE.Read(reader),
			FfiConverterOptionalTypeFeatureLanaINSTANCE.Read(reader),
			FfiConverterOptionalTypeFeaturePathsINSTANCE.Read(reader),
			FfiConverterOptionalTypeFeatureDirectINSTANCE.Read(reader),
			FfiConverterOptionalBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterOptionalTypeFeatureDerpINSTANCE.Read(reader),
			FfiConverterTypeFeatureValidateKeysINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterTypeFeatureFirewallINSTANCE.Read(reader),
			FfiConverterOptionalUint64INSTANCE.Read(reader),
			FfiConverterOptionalTypeFeatureLinkDetectionINSTANCE.Read(reader),
			FfiConverterTypeFeatureDnsINSTANCE.Read(reader),
			FfiConverterTypeFeaturePostQuantumVPNINSTANCE.Read(reader),
			FfiConverterOptionalTypeFeaturePmtuDiscoveryINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterOptionalTypeFeatureBatchingINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeFeatures) Lower(value Features) RustBuffer {
	return LowerIntoRustBuffer[Features](c, value)
}

func (c FfiConverterTypeFeatures) Write(writer io.Writer, value Features) {
		FfiConverterTypeFeatureWireguardINSTANCE.Write(writer, value.Wireguard);
		FfiConverterOptionalTypeFeatureNurseINSTANCE.Write(writer, value.Nurse);
		FfiConverterOptionalTypeFeatureLanaINSTANCE.Write(writer, value.Lana);
		FfiConverterOptionalTypeFeaturePathsINSTANCE.Write(writer, value.Paths);
		FfiConverterOptionalTypeFeatureDirectINSTANCE.Write(writer, value.Direct);
		FfiConverterOptionalBoolINSTANCE.Write(writer, value.IsTestEnv);
		FfiConverterBoolINSTANCE.Write(writer, value.HideIps);
		FfiConverterOptionalTypeFeatureDerpINSTANCE.Write(writer, value.Derp);
		FfiConverterTypeFeatureValidateKeysINSTANCE.Write(writer, value.ValidateKeys);
		FfiConverterBoolINSTANCE.Write(writer, value.Ipv6);
		FfiConverterBoolINSTANCE.Write(writer, value.Nicknames);
		FfiConverterTypeFeatureFirewallINSTANCE.Write(writer, value.Firewall);
		FfiConverterOptionalUint64INSTANCE.Write(writer, value.FlushEventsOnStopTimeoutSeconds);
		FfiConverterOptionalTypeFeatureLinkDetectionINSTANCE.Write(writer, value.LinkDetection);
		FfiConverterTypeFeatureDnsINSTANCE.Write(writer, value.Dns);
		FfiConverterTypeFeaturePostQuantumVPNINSTANCE.Write(writer, value.PostQuantumVpn);
		FfiConverterOptionalTypeFeaturePmtuDiscoveryINSTANCE.Write(writer, value.PmtuDiscovery);
		FfiConverterBoolINSTANCE.Write(writer, value.Multicast);
		FfiConverterOptionalTypeFeatureBatchingINSTANCE.Write(writer, value.Batching);
}

type FfiDestroyerTypeFeatures struct {}

func (_ FfiDestroyerTypeFeatures) Destroy(value Features) {
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
	// Flag to control whether the peer allows incoming files
	AllowPeerSendFiles bool
	// Flag to control whether we allow multicast messages from the peer
	AllowMulticast bool
	// Flag to control whether the peer allows multicast messages from us
	PeerAllowsMulticast bool
}

func (r *Peer) Destroy() {
		FfiDestroyerTypePeerBase{}.Destroy(r.Base);
		FfiDestroyerBool{}.Destroy(r.IsLocal);
		FfiDestroyerBool{}.Destroy(r.AllowIncomingConnections);
		FfiDestroyerBool{}.Destroy(r.AllowPeerSendFiles);
		FfiDestroyerBool{}.Destroy(r.AllowMulticast);
		FfiDestroyerBool{}.Destroy(r.PeerAllowsMulticast);
}

type FfiConverterTypePeer struct {}

var FfiConverterTypePeerINSTANCE = FfiConverterTypePeer{}

func (c FfiConverterTypePeer) Lift(rb RustBufferI) Peer {
	return LiftFromRustBuffer[Peer](c, rb)
}

func (c FfiConverterTypePeer) Read(reader io.Reader) Peer {
	return Peer {
			FfiConverterTypePeerBaseINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypePeer) Lower(value Peer) RustBuffer {
	return LowerIntoRustBuffer[Peer](c, value)
}

func (c FfiConverterTypePeer) Write(writer io.Writer, value Peer) {
		FfiConverterTypePeerBaseINSTANCE.Write(writer, value.Base);
		FfiConverterBoolINSTANCE.Write(writer, value.IsLocal);
		FfiConverterBoolINSTANCE.Write(writer, value.AllowIncomingConnections);
		FfiConverterBoolINSTANCE.Write(writer, value.AllowPeerSendFiles);
		FfiConverterBoolINSTANCE.Write(writer, value.AllowMulticast);
		FfiConverterBoolINSTANCE.Write(writer, value.PeerAllowsMulticast);
}

type FfiDestroyerTypePeer struct {}

func (_ FfiDestroyerTypePeer) Destroy(value Peer) {
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

type FfiConverterTypePeerBase struct {}

var FfiConverterTypePeerBaseINSTANCE = FfiConverterTypePeerBase{}

func (c FfiConverterTypePeerBase) Lift(rb RustBufferI) PeerBase {
	return LiftFromRustBuffer[PeerBase](c, rb)
}

func (c FfiConverterTypePeerBase) Read(reader io.Reader) PeerBase {
	return PeerBase {
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterTypePublicKeyINSTANCE.Read(reader),
			FfiConverterTypeHiddenStringINSTANCE.Read(reader),
			FfiConverterOptionalSequenceTypeIpAddrINSTANCE.Read(reader),
			FfiConverterOptionalTypeHiddenStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypePeerBase) Lower(value PeerBase) RustBuffer {
	return LowerIntoRustBuffer[PeerBase](c, value)
}

func (c FfiConverterTypePeerBase) Write(writer io.Writer, value PeerBase) {
		FfiConverterStringINSTANCE.Write(writer, value.Identifier);
		FfiConverterTypePublicKeyINSTANCE.Write(writer, value.PublicKey);
		FfiConverterTypeHiddenStringINSTANCE.Write(writer, value.Hostname);
		FfiConverterOptionalSequenceTypeIpAddrINSTANCE.Write(writer, value.IpAddresses);
		FfiConverterOptionalTypeHiddenStringINSTANCE.Write(writer, value.Nickname);
}

type FfiDestroyerTypePeerBase struct {}

func (_ FfiDestroyerTypePeerBase) Destroy(value PeerBase) {
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
		FfiDestroyerTypeRelayState{}.Destroy(r.ConnState);
}

type FfiConverterTypeServer struct {}

var FfiConverterTypeServerINSTANCE = FfiConverterTypeServer{}

func (c FfiConverterTypeServer) Lift(rb RustBufferI) Server {
	return LiftFromRustBuffer[Server](c, rb)
}

func (c FfiConverterTypeServer) Read(reader io.Reader) Server {
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
			FfiConverterTypeRelayStateINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeServer) Lower(value Server) RustBuffer {
	return LowerIntoRustBuffer[Server](c, value)
}

func (c FfiConverterTypeServer) Write(writer io.Writer, value Server) {
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
		FfiConverterTypeRelayStateINSTANCE.Write(writer, value.ConnState);
}

type FfiDestroyerTypeServer struct {}

func (_ FfiDestroyerTypeServer) Destroy(value Server) {
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
	// Flag to control whether the Node allows incoming files
	AllowPeerSendFiles bool
	// Connection type in the network mesh (through Relay or hole punched directly)
	Path PathType
	// Flag to control whether we allow multicast messages from the Node
	AllowMulticast bool
	// Flag to control whether the Node allows multicast messages from us
	PeerAllowsMulticast bool
}

func (r *TelioNode) Destroy() {
		FfiDestroyerString{}.Destroy(r.Identifier);
		FfiDestroyerTypePublicKey{}.Destroy(r.PublicKey);
		FfiDestroyerOptionalString{}.Destroy(r.Nickname);
		FfiDestroyerTypeNodeState{}.Destroy(r.State);
		FfiDestroyerOptionalTypeLinkState{}.Destroy(r.LinkState);
		FfiDestroyerBool{}.Destroy(r.IsExit);
		FfiDestroyerBool{}.Destroy(r.IsVpn);
		FfiDestroyerSequenceTypeIpAddr{}.Destroy(r.IpAddresses);
		FfiDestroyerSequenceTypeIpNet{}.Destroy(r.AllowedIps);
		FfiDestroyerOptionalTypeSocketAddr{}.Destroy(r.Endpoint);
		FfiDestroyerOptionalString{}.Destroy(r.Hostname);
		FfiDestroyerBool{}.Destroy(r.AllowIncomingConnections);
		FfiDestroyerBool{}.Destroy(r.AllowPeerSendFiles);
		FfiDestroyerTypePathType{}.Destroy(r.Path);
		FfiDestroyerBool{}.Destroy(r.AllowMulticast);
		FfiDestroyerBool{}.Destroy(r.PeerAllowsMulticast);
}

type FfiConverterTypeTelioNode struct {}

var FfiConverterTypeTelioNodeINSTANCE = FfiConverterTypeTelioNode{}

func (c FfiConverterTypeTelioNode) Lift(rb RustBufferI) TelioNode {
	return LiftFromRustBuffer[TelioNode](c, rb)
}

func (c FfiConverterTypeTelioNode) Read(reader io.Reader) TelioNode {
	return TelioNode {
			FfiConverterStringINSTANCE.Read(reader),
			FfiConverterTypePublicKeyINSTANCE.Read(reader),
			FfiConverterOptionalStringINSTANCE.Read(reader),
			FfiConverterTypeNodeStateINSTANCE.Read(reader),
			FfiConverterOptionalTypeLinkStateINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterSequenceTypeIpAddrINSTANCE.Read(reader),
			FfiConverterSequenceTypeIpNetINSTANCE.Read(reader),
			FfiConverterOptionalTypeSocketAddrINSTANCE.Read(reader),
			FfiConverterOptionalStringINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterTypePathTypeINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeTelioNode) Lower(value TelioNode) RustBuffer {
	return LowerIntoRustBuffer[TelioNode](c, value)
}

func (c FfiConverterTypeTelioNode) Write(writer io.Writer, value TelioNode) {
		FfiConverterStringINSTANCE.Write(writer, value.Identifier);
		FfiConverterTypePublicKeyINSTANCE.Write(writer, value.PublicKey);
		FfiConverterOptionalStringINSTANCE.Write(writer, value.Nickname);
		FfiConverterTypeNodeStateINSTANCE.Write(writer, value.State);
		FfiConverterOptionalTypeLinkStateINSTANCE.Write(writer, value.LinkState);
		FfiConverterBoolINSTANCE.Write(writer, value.IsExit);
		FfiConverterBoolINSTANCE.Write(writer, value.IsVpn);
		FfiConverterSequenceTypeIpAddrINSTANCE.Write(writer, value.IpAddresses);
		FfiConverterSequenceTypeIpNetINSTANCE.Write(writer, value.AllowedIps);
		FfiConverterOptionalTypeSocketAddrINSTANCE.Write(writer, value.Endpoint);
		FfiConverterOptionalStringINSTANCE.Write(writer, value.Hostname);
		FfiConverterBoolINSTANCE.Write(writer, value.AllowIncomingConnections);
		FfiConverterBoolINSTANCE.Write(writer, value.AllowPeerSendFiles);
		FfiConverterTypePathTypeINSTANCE.Write(writer, value.Path);
		FfiConverterBoolINSTANCE.Write(writer, value.AllowMulticast);
		FfiConverterBoolINSTANCE.Write(writer, value.PeerAllowsMulticast);
}

type FfiDestroyerTypeTelioNode struct {}

func (_ FfiDestroyerTypeTelioNode) Destroy(value TelioNode) {
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

type FfiConverterTypeEndpointProvider struct {}

var FfiConverterTypeEndpointProviderINSTANCE = FfiConverterTypeEndpointProvider{}

func (c FfiConverterTypeEndpointProvider) Lift(rb RustBufferI) EndpointProvider {
	return LiftFromRustBuffer[EndpointProvider](c, rb)
}

func (c FfiConverterTypeEndpointProvider) Lower(value EndpointProvider) RustBuffer {
	return LowerIntoRustBuffer[EndpointProvider](c, value)
}
func (FfiConverterTypeEndpointProvider) Read(reader io.Reader) EndpointProvider {
	id := readInt32(reader)
	return EndpointProvider(id)
}

func (FfiConverterTypeEndpointProvider) Write(writer io.Writer, value EndpointProvider) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeEndpointProvider struct {}

func (_ FfiDestroyerTypeEndpointProvider) Destroy(value EndpointProvider) {
}




// Error code. Common error code representation (for statistics).
type ErrorCode uint

const (
	// There is no error in the execution
	ErrorCodeNoError ErrorCode = 1
	// The error type is unknown
	ErrorCodeUnknown ErrorCode = 2
)

type FfiConverterTypeErrorCode struct {}

var FfiConverterTypeErrorCodeINSTANCE = FfiConverterTypeErrorCode{}

func (c FfiConverterTypeErrorCode) Lift(rb RustBufferI) ErrorCode {
	return LiftFromRustBuffer[ErrorCode](c, rb)
}

func (c FfiConverterTypeErrorCode) Lower(value ErrorCode) RustBuffer {
	return LowerIntoRustBuffer[ErrorCode](c, value)
}
func (FfiConverterTypeErrorCode) Read(reader io.Reader) ErrorCode {
	id := readInt32(reader)
	return ErrorCode(id)
}

func (FfiConverterTypeErrorCode) Write(writer io.Writer, value ErrorCode) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeErrorCode struct {}

func (_ FfiDestroyerTypeErrorCode) Destroy(value ErrorCode) {
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

type FfiConverterTypeErrorLevel struct {}

var FfiConverterTypeErrorLevelINSTANCE = FfiConverterTypeErrorLevel{}

func (c FfiConverterTypeErrorLevel) Lift(rb RustBufferI) ErrorLevel {
	return LiftFromRustBuffer[ErrorLevel](c, rb)
}

func (c FfiConverterTypeErrorLevel) Lower(value ErrorLevel) RustBuffer {
	return LowerIntoRustBuffer[ErrorLevel](c, value)
}
func (FfiConverterTypeErrorLevel) Read(reader io.Reader) ErrorLevel {
	id := readInt32(reader)
	return ErrorLevel(id)
}

func (FfiConverterTypeErrorLevel) Write(writer io.Writer, value ErrorLevel) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeErrorLevel struct {}

func (_ FfiDestroyerTypeErrorLevel) Destroy(value ErrorLevel) {
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
		FfiDestroyerTypeServer{}.Destroy(e.Body);
}
// Used to report events related to the Node
type EventNode struct {
	Body TelioNode
}

func (e EventNode) Destroy() {
		FfiDestroyerTypeTelioNode{}.Destroy(e.Body);
}
// Initialize an Error type event.
// Used to inform errors to the upper layers of libtelio
type EventError struct {
	Body ErrorEvent
}

func (e EventError) Destroy() {
		FfiDestroyerTypeErrorEvent{}.Destroy(e.Body);
}

type FfiConverterTypeEvent struct {}

var FfiConverterTypeEventINSTANCE = FfiConverterTypeEvent{}

func (c FfiConverterTypeEvent) Lift(rb RustBufferI) Event {
	return LiftFromRustBuffer[Event](c, rb)
}

func (c FfiConverterTypeEvent) Lower(value Event) RustBuffer {
	return LowerIntoRustBuffer[Event](c, value)
}
func (FfiConverterTypeEvent) Read(reader io.Reader) Event {
	id := readInt32(reader)
	switch (id) {
		case 1:
			return EventRelay{
				FfiConverterTypeServerINSTANCE.Read(reader),
			};
		case 2:
			return EventNode{
				FfiConverterTypeTelioNodeINSTANCE.Read(reader),
			};
		case 3:
			return EventError{
				FfiConverterTypeErrorEventINSTANCE.Read(reader),
			};
		default:
			panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeEvent.Read()", id));
	}
}

func (FfiConverterTypeEvent) Write(writer io.Writer, value Event) {
	switch variant_value := value.(type) {
		case EventRelay:
			writeInt32(writer, 1)
			FfiConverterTypeServerINSTANCE.Write(writer, variant_value.Body)
		case EventNode:
			writeInt32(writer, 2)
			FfiConverterTypeTelioNodeINSTANCE.Write(writer, variant_value.Body)
		case EventError:
			writeInt32(writer, 3)
			FfiConverterTypeErrorEventINSTANCE.Write(writer, variant_value.Body)
		default:
			_ = variant_value
			panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeEvent.Write", value))
	}
}

type FfiDestroyerTypeEvent struct {}

func (_ FfiDestroyerTypeEvent) Destroy(value Event) {
	value.Destroy()
}




// Link state hint
type LinkState uint

const (
	LinkStateDown LinkState = 1
	LinkStateUp LinkState = 2
)

type FfiConverterTypeLinkState struct {}

var FfiConverterTypeLinkStateINSTANCE = FfiConverterTypeLinkState{}

func (c FfiConverterTypeLinkState) Lift(rb RustBufferI) LinkState {
	return LiftFromRustBuffer[LinkState](c, rb)
}

func (c FfiConverterTypeLinkState) Lower(value LinkState) RustBuffer {
	return LowerIntoRustBuffer[LinkState](c, value)
}
func (FfiConverterTypeLinkState) Read(reader io.Reader) LinkState {
	id := readInt32(reader)
	return LinkState(id)
}

func (FfiConverterTypeLinkState) Write(writer io.Writer, value LinkState) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeLinkState struct {}

func (_ FfiDestroyerTypeLinkState) Destroy(value LinkState) {
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

type FfiConverterTypeNatType struct {}

var FfiConverterTypeNatTypeINSTANCE = FfiConverterTypeNatType{}

func (c FfiConverterTypeNatType) Lift(rb RustBufferI) NatType {
	return LiftFromRustBuffer[NatType](c, rb)
}

func (c FfiConverterTypeNatType) Lower(value NatType) RustBuffer {
	return LowerIntoRustBuffer[NatType](c, value)
}
func (FfiConverterTypeNatType) Read(reader io.Reader) NatType {
	id := readInt32(reader)
	return NatType(id)
}

func (FfiConverterTypeNatType) Write(writer io.Writer, value NatType) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeNatType struct {}

func (_ FfiDestroyerTypeNatType) Destroy(value NatType) {
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

type FfiConverterTypeNodeState struct {}

var FfiConverterTypeNodeStateINSTANCE = FfiConverterTypeNodeState{}

func (c FfiConverterTypeNodeState) Lift(rb RustBufferI) NodeState {
	return LiftFromRustBuffer[NodeState](c, rb)
}

func (c FfiConverterTypeNodeState) Lower(value NodeState) RustBuffer {
	return LowerIntoRustBuffer[NodeState](c, value)
}
func (FfiConverterTypeNodeState) Read(reader io.Reader) NodeState {
	id := readInt32(reader)
	return NodeState(id)
}

func (FfiConverterTypeNodeState) Write(writer io.Writer, value NodeState) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeNodeState struct {}

func (_ FfiDestroyerTypeNodeState) Destroy(value NodeState) {
}




// Mesh connection path type
type PathType uint

const (
	// Nodes connected via a middle-man relay
	PathTypeRelay PathType = 1
	// Nodes connected directly via WG
	PathTypeDirect PathType = 2
)

type FfiConverterTypePathType struct {}

var FfiConverterTypePathTypeINSTANCE = FfiConverterTypePathType{}

func (c FfiConverterTypePathType) Lift(rb RustBufferI) PathType {
	return LiftFromRustBuffer[PathType](c, rb)
}

func (c FfiConverterTypePathType) Lower(value PathType) RustBuffer {
	return LowerIntoRustBuffer[PathType](c, value)
}
func (FfiConverterTypePathType) Read(reader io.Reader) PathType {
	id := readInt32(reader)
	return PathType(id)
}

func (FfiConverterTypePathType) Write(writer io.Writer, value PathType) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypePathType struct {}

func (_ FfiDestroyerTypePathType) Destroy(value PathType) {
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

type FfiConverterTypeRelayState struct {}

var FfiConverterTypeRelayStateINSTANCE = FfiConverterTypeRelayState{}

func (c FfiConverterTypeRelayState) Lift(rb RustBufferI) RelayState {
	return LiftFromRustBuffer[RelayState](c, rb)
}

func (c FfiConverterTypeRelayState) Lower(value RelayState) RustBuffer {
	return LowerIntoRustBuffer[RelayState](c, value)
}
func (FfiConverterTypeRelayState) Read(reader io.Reader) RelayState {
	id := readInt32(reader)
	return RelayState(id)
}

func (FfiConverterTypeRelayState) Write(writer io.Writer, value RelayState) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeRelayState struct {}

func (_ FfiDestroyerTypeRelayState) Destroy(value RelayState) {
}




// Available ways to calculate RTT
type RttType uint

const (
	// Simple ping request
	RttTypePing RttType = 1
)

type FfiConverterTypeRttType struct {}

var FfiConverterTypeRttTypeINSTANCE = FfiConverterTypeRttType{}

func (c FfiConverterTypeRttType) Lift(rb RustBufferI) RttType {
	return LiftFromRustBuffer[RttType](c, rb)
}

func (c FfiConverterTypeRttType) Lower(value RttType) RustBuffer {
	return LowerIntoRustBuffer[RttType](c, value)
}
func (FfiConverterTypeRttType) Read(reader io.Reader) RttType {
	id := readInt32(reader)
	return RttType(id)
}

func (FfiConverterTypeRttType) Write(writer io.Writer, value RttType) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeRttType struct {}

func (_ FfiDestroyerTypeRttType) Destroy(value RttType) {
}




// Possible adapters.
type TelioAdapterType uint

const (
	// Userland rust implementation.
	TelioAdapterTypeBoringTun TelioAdapterType = 1
	// Linux in-kernel WireGuard implementation
	TelioAdapterTypeLinuxNativeTun TelioAdapterType = 2
	// WireguardGo implementation
	TelioAdapterTypeWireguardGoTun TelioAdapterType = 3
	// WindowsNativeWireguardNt implementation
	TelioAdapterTypeWindowsNativeTun TelioAdapterType = 4
)

type FfiConverterTypeTelioAdapterType struct {}

var FfiConverterTypeTelioAdapterTypeINSTANCE = FfiConverterTypeTelioAdapterType{}

func (c FfiConverterTypeTelioAdapterType) Lift(rb RustBufferI) TelioAdapterType {
	return LiftFromRustBuffer[TelioAdapterType](c, rb)
}

func (c FfiConverterTypeTelioAdapterType) Lower(value TelioAdapterType) RustBuffer {
	return LowerIntoRustBuffer[TelioAdapterType](c, value)
}
func (FfiConverterTypeTelioAdapterType) Read(reader io.Reader) TelioAdapterType {
	id := readInt32(reader)
	return TelioAdapterType(id)
}

func (FfiConverterTypeTelioAdapterType) Write(writer io.Writer, value TelioAdapterType) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeTelioAdapterType struct {}

func (_ FfiDestroyerTypeTelioAdapterType) Destroy(value TelioAdapterType) {
}


type TelioError struct {
	err error
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
	return &TelioError{
		err: &TelioErrorUnknownError{
			Inner: inner,
		},
	}
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
	return &TelioError{
		err: &TelioErrorInvalidKey{
		},
	}
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
	return &TelioError{
		err: &TelioErrorBadConfig{
		},
	}
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
	return &TelioError{
		err: &TelioErrorLockError{
		},
	}
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
	return &TelioError{
		err: &TelioErrorInvalidString{
		},
	}
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
	return &TelioError{
		err: &TelioErrorAlreadyStarted{
		},
	}
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
	return &TelioError{
		err: &TelioErrorNotStarted{
		},
	}
}

func (err TelioErrorNotStarted) Error() string {
	return fmt.Sprint("NotStarted",
		
	)
}

func (self TelioErrorNotStarted) Is(target error) bool {
	return target == ErrTelioErrorNotStarted
}

type FfiConverterTypeTelioError struct{}

var FfiConverterTypeTelioErrorINSTANCE = FfiConverterTypeTelioError{}

func (c FfiConverterTypeTelioError) Lift(eb RustBufferI) error {
	return LiftFromRustBuffer[error](c, eb)
}

func (c FfiConverterTypeTelioError) Lower(value *TelioError) RustBuffer {
	return LowerIntoRustBuffer[*TelioError](c, value)
}

func (c FfiConverterTypeTelioError) Read(reader io.Reader) error {
	errorID := readUint32(reader)

	switch errorID {
	case 1:
		return &TelioError{&TelioErrorUnknownError{
			Inner: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 2:
		return &TelioError{&TelioErrorInvalidKey{
		}}
	case 3:
		return &TelioError{&TelioErrorBadConfig{
		}}
	case 4:
		return &TelioError{&TelioErrorLockError{
		}}
	case 5:
		return &TelioError{&TelioErrorInvalidString{
		}}
	case 6:
		return &TelioError{&TelioErrorAlreadyStarted{
		}}
	case 7:
		return &TelioError{&TelioErrorNotStarted{
		}}
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterTypeTelioError.Read()", errorID))
	}
}

func (c FfiConverterTypeTelioError) Write(writer io.Writer, value *TelioError) {
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
			panic(fmt.Sprintf("invalid error value `%v` in FfiConverterTypeTelioError.Write", value))
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

type FfiConverterTypeTelioLogLevel struct {}

var FfiConverterTypeTelioLogLevelINSTANCE = FfiConverterTypeTelioLogLevel{}

func (c FfiConverterTypeTelioLogLevel) Lift(rb RustBufferI) TelioLogLevel {
	return LiftFromRustBuffer[TelioLogLevel](c, rb)
}

func (c FfiConverterTypeTelioLogLevel) Lower(value TelioLogLevel) RustBuffer {
	return LowerIntoRustBuffer[TelioLogLevel](c, value)
}
func (FfiConverterTypeTelioLogLevel) Read(reader io.Reader) TelioLogLevel {
	id := readInt32(reader)
	return TelioLogLevel(id)
}

func (FfiConverterTypeTelioLogLevel) Write(writer io.Writer, value TelioLogLevel) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeTelioLogLevel struct {}

func (_ FfiDestroyerTypeTelioLogLevel) Destroy(value TelioLogLevel) {
}




type uniffiCallbackResult C.int32_t

const (
	uniffiIdxCallbackFree               uniffiCallbackResult = 0
	uniffiCallbackResultSuccess         uniffiCallbackResult = 0
	uniffiCallbackResultError           uniffiCallbackResult = 1
	uniffiCallbackUnexpectedResultError uniffiCallbackResult = 2
	uniffiCallbackCancelled             uniffiCallbackResult = 3
)


type concurrentHandleMap[T any] struct {
	leftMap       map[uint64]*T
	rightMap      map[*T]uint64
	currentHandle uint64
	lock          sync.RWMutex
}

func newConcurrentHandleMap[T any]() *concurrentHandleMap[T] {
	return &concurrentHandleMap[T]{
		leftMap:  map[uint64]*T{},
		rightMap: map[*T]uint64{},
	}
}

func (cm *concurrentHandleMap[T]) insert(obj *T) uint64 {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	if existingHandle, ok := cm.rightMap[obj]; ok {
		return existingHandle
	}
	cm.currentHandle = cm.currentHandle + 1
	cm.leftMap[cm.currentHandle] = obj
	cm.rightMap[obj] = cm.currentHandle
	return cm.currentHandle
}

func (cm *concurrentHandleMap[T]) remove(handle uint64) bool {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	if val, ok := cm.leftMap[handle]; ok {
		delete(cm.leftMap, handle)
		delete(cm.rightMap, val)
	}
	return false
}

func (cm *concurrentHandleMap[T]) tryGet(handle uint64) (*T, bool) {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	val, ok := cm.leftMap[handle]
	return val, ok
}

type FfiConverterCallbackInterface[CallbackInterface any] struct {
	handleMap *concurrentHandleMap[CallbackInterface]
}

func (c *FfiConverterCallbackInterface[CallbackInterface]) drop(handle uint64) RustBuffer {
	c.handleMap.remove(handle)
	return RustBuffer{}
}

func (c *FfiConverterCallbackInterface[CallbackInterface]) Lift(handle uint64) CallbackInterface {
	val, ok := c.handleMap.tryGet(handle)
	if !ok {
		panic(fmt.Errorf("no callback in handle map: %d", handle))
	}
	return *val
}

func (c *FfiConverterCallbackInterface[CallbackInterface]) Read(reader io.Reader) CallbackInterface {
	return c.Lift(readUint64(reader))
}

func (c *FfiConverterCallbackInterface[CallbackInterface]) Lower(value CallbackInterface) C.uint64_t {
	return C.uint64_t(c.handleMap.insert(&value))
}

func (c *FfiConverterCallbackInterface[CallbackInterface]) Write(writer io.Writer, value CallbackInterface) {
	writeUint64(writer, uint64(c.Lower(value)))
}
type TelioEventCb interface {
	
	Event(payload Event) *TelioError
	
}

// foreignCallbackCallbackInterfaceTelioEventCb cannot be callable be a compiled function at a same time
type foreignCallbackCallbackInterfaceTelioEventCb struct {}

//export telio_cgo_TelioEventCb
func telio_cgo_TelioEventCb(handle C.uint64_t, method C.int32_t, argsPtr *C.uint8_t, argsLen C.int32_t, outBuf *C.RustBuffer) C.int32_t {
	cb := FfiConverterCallbackInterfaceTelioEventCbINSTANCE.Lift(uint64(handle));
	switch method {
	case 0:
		// 0 means Rust is done with the callback, and the callback
		// can be dropped by the foreign language.
		*outBuf = FfiConverterCallbackInterfaceTelioEventCbINSTANCE.drop(uint64(handle))
		// See docs of ForeignCallback in `uniffi/src/ffi/foreigncallbacks.rs`
		return C.int32_t(uniffiIdxCallbackFree)

	case 1:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceTelioEventCb{}.InvokeEvent(cb, args, outBuf);
		return C.int32_t(result)
	
	default:
		// This should never happen, because an out of bounds method index won't
		// ever be used. Once we can catch errors, we should return an InternalException.
		// https://github.com/mozilla/uniffi-rs/issues/351
		return C.int32_t(uniffiCallbackUnexpectedResultError)
	}
}

func (foreignCallbackCallbackInterfaceTelioEventCb) InvokeEvent (callback TelioEventCb, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	err :=callback.Event(FfiConverterTypeEventINSTANCE.Read(reader));

        if err != nil {
		// The only way to bypass an unexpected error is to bypass pointer to an empty
		// instance of the error
		if err.err == nil {
			return uniffiCallbackUnexpectedResultError
		}
		*outBuf = LowerIntoRustBuffer[*TelioError](FfiConverterTypeTelioErrorINSTANCE, err)
		return uniffiCallbackResultError
	}
	return uniffiCallbackResultSuccess
}


type FfiConverterCallbackInterfaceTelioEventCb struct {
	FfiConverterCallbackInterface[TelioEventCb]
}

var FfiConverterCallbackInterfaceTelioEventCbINSTANCE = &FfiConverterCallbackInterfaceTelioEventCb {
	FfiConverterCallbackInterface: FfiConverterCallbackInterface[TelioEventCb]{
		handleMap: newConcurrentHandleMap[TelioEventCb](),
	},
}

// This is a static function because only 1 instance is supported for registering
func (c *FfiConverterCallbackInterfaceTelioEventCb) register() {
	rustCall(func(status *C.RustCallStatus) int32 {
		C.uniffi_telio_fn_init_callback_telioeventcb(C.ForeignCallback(C.telio_cgo_TelioEventCb), status)
		return 0
	})
}

type FfiDestroyerCallbackInterfaceTelioEventCb struct {}

func (FfiDestroyerCallbackInterfaceTelioEventCb) Destroy(value TelioEventCb) {
}





type TelioLoggerCb interface {
	
	Log(logLevel TelioLogLevel, payload string) *TelioError
	
}

// foreignCallbackCallbackInterfaceTelioLoggerCb cannot be callable be a compiled function at a same time
type foreignCallbackCallbackInterfaceTelioLoggerCb struct {}

//export telio_cgo_TelioLoggerCb
func telio_cgo_TelioLoggerCb(handle C.uint64_t, method C.int32_t, argsPtr *C.uint8_t, argsLen C.int32_t, outBuf *C.RustBuffer) C.int32_t {
	cb := FfiConverterCallbackInterfaceTelioLoggerCbINSTANCE.Lift(uint64(handle));
	switch method {
	case 0:
		// 0 means Rust is done with the callback, and the callback
		// can be dropped by the foreign language.
		*outBuf = FfiConverterCallbackInterfaceTelioLoggerCbINSTANCE.drop(uint64(handle))
		// See docs of ForeignCallback in `uniffi/src/ffi/foreigncallbacks.rs`
		return C.int32_t(uniffiIdxCallbackFree)

	case 1:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceTelioLoggerCb{}.InvokeLog(cb, args, outBuf);
		return C.int32_t(result)
	
	default:
		// This should never happen, because an out of bounds method index won't
		// ever be used. Once we can catch errors, we should return an InternalException.
		// https://github.com/mozilla/uniffi-rs/issues/351
		return C.int32_t(uniffiCallbackUnexpectedResultError)
	}
}

func (foreignCallbackCallbackInterfaceTelioLoggerCb) InvokeLog (callback TelioLoggerCb, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	err :=callback.Log(FfiConverterTypeTelioLogLevelINSTANCE.Read(reader), FfiConverterStringINSTANCE.Read(reader));

        if err != nil {
		// The only way to bypass an unexpected error is to bypass pointer to an empty
		// instance of the error
		if err.err == nil {
			return uniffiCallbackUnexpectedResultError
		}
		*outBuf = LowerIntoRustBuffer[*TelioError](FfiConverterTypeTelioErrorINSTANCE, err)
		return uniffiCallbackResultError
	}
	return uniffiCallbackResultSuccess
}


type FfiConverterCallbackInterfaceTelioLoggerCb struct {
	FfiConverterCallbackInterface[TelioLoggerCb]
}

var FfiConverterCallbackInterfaceTelioLoggerCbINSTANCE = &FfiConverterCallbackInterfaceTelioLoggerCb {
	FfiConverterCallbackInterface: FfiConverterCallbackInterface[TelioLoggerCb]{
		handleMap: newConcurrentHandleMap[TelioLoggerCb](),
	},
}

// This is a static function because only 1 instance is supported for registering
func (c *FfiConverterCallbackInterfaceTelioLoggerCb) register() {
	rustCall(func(status *C.RustCallStatus) int32 {
		C.uniffi_telio_fn_init_callback_teliologgercb(C.ForeignCallback(C.telio_cgo_TelioLoggerCb), status)
		return 0
	})
}

type FfiDestroyerCallbackInterfaceTelioLoggerCb struct {}

func (FfiDestroyerCallbackInterfaceTelioLoggerCb) Destroy(value TelioLoggerCb) {
}





type TelioProtectCb interface {
	
	Protect(socketId int32) *TelioError
	
}

// foreignCallbackCallbackInterfaceTelioProtectCb cannot be callable be a compiled function at a same time
type foreignCallbackCallbackInterfaceTelioProtectCb struct {}

//export telio_cgo_TelioProtectCb
func telio_cgo_TelioProtectCb(handle C.uint64_t, method C.int32_t, argsPtr *C.uint8_t, argsLen C.int32_t, outBuf *C.RustBuffer) C.int32_t {
	cb := FfiConverterCallbackInterfaceTelioProtectCbINSTANCE.Lift(uint64(handle));
	switch method {
	case 0:
		// 0 means Rust is done with the callback, and the callback
		// can be dropped by the foreign language.
		*outBuf = FfiConverterCallbackInterfaceTelioProtectCbINSTANCE.drop(uint64(handle))
		// See docs of ForeignCallback in `uniffi/src/ffi/foreigncallbacks.rs`
		return C.int32_t(uniffiIdxCallbackFree)

	case 1:
		var result uniffiCallbackResult
		args := unsafe.Slice((*byte)(argsPtr), argsLen)
		result = foreignCallbackCallbackInterfaceTelioProtectCb{}.InvokeProtect(cb, args, outBuf);
		return C.int32_t(result)
	
	default:
		// This should never happen, because an out of bounds method index won't
		// ever be used. Once we can catch errors, we should return an InternalException.
		// https://github.com/mozilla/uniffi-rs/issues/351
		return C.int32_t(uniffiCallbackUnexpectedResultError)
	}
}

func (foreignCallbackCallbackInterfaceTelioProtectCb) InvokeProtect (callback TelioProtectCb, args []byte, outBuf *C.RustBuffer) uniffiCallbackResult {
	reader := bytes.NewReader(args)
	err :=callback.Protect(FfiConverterInt32INSTANCE.Read(reader));

        if err != nil {
		// The only way to bypass an unexpected error is to bypass pointer to an empty
		// instance of the error
		if err.err == nil {
			return uniffiCallbackUnexpectedResultError
		}
		*outBuf = LowerIntoRustBuffer[*TelioError](FfiConverterTypeTelioErrorINSTANCE, err)
		return uniffiCallbackResultError
	}
	return uniffiCallbackResultSuccess
}


type FfiConverterCallbackInterfaceTelioProtectCb struct {
	FfiConverterCallbackInterface[TelioProtectCb]
}

var FfiConverterCallbackInterfaceTelioProtectCbINSTANCE = &FfiConverterCallbackInterfaceTelioProtectCb {
	FfiConverterCallbackInterface: FfiConverterCallbackInterface[TelioProtectCb]{
		handleMap: newConcurrentHandleMap[TelioProtectCb](),
	},
}

// This is a static function because only 1 instance is supported for registering
func (c *FfiConverterCallbackInterfaceTelioProtectCb) register() {
	rustCall(func(status *C.RustCallStatus) int32 {
		C.uniffi_telio_fn_init_callback_telioprotectcb(C.ForeignCallback(C.telio_cgo_TelioProtectCb), status)
		return 0
	})
}

type FfiDestroyerCallbackInterfaceTelioProtectCb struct {}

func (FfiDestroyerCallbackInterfaceTelioProtectCb) Destroy(value TelioProtectCb) {
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

func (c FfiConverterOptionalUint32) Lower(value *uint32) RustBuffer {
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

func (c FfiConverterOptionalUint64) Lower(value *uint64) RustBuffer {
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

func (c FfiConverterOptionalBool) Lower(value *bool) RustBuffer {
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

func (c FfiConverterOptionalString) Lower(value *string) RustBuffer {
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



type FfiConverterOptionalTypeDnsConfig struct{}

var FfiConverterOptionalTypeDnsConfigINSTANCE = FfiConverterOptionalTypeDnsConfig{}

func (c FfiConverterOptionalTypeDnsConfig) Lift(rb RustBufferI) *DnsConfig {
	return LiftFromRustBuffer[*DnsConfig](c, rb)
}

func (_ FfiConverterOptionalTypeDnsConfig) Read(reader io.Reader) *DnsConfig {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeDnsConfigINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeDnsConfig) Lower(value *DnsConfig) RustBuffer {
	return LowerIntoRustBuffer[*DnsConfig](c, value)
}

func (_ FfiConverterOptionalTypeDnsConfig) Write(writer io.Writer, value *DnsConfig) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeDnsConfigINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeDnsConfig struct {}

func (_ FfiDestroyerOptionalTypeDnsConfig) Destroy(value *DnsConfig) {
	if value != nil {
		FfiDestroyerTypeDnsConfig{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypeFeatureBatching struct{}

var FfiConverterOptionalTypeFeatureBatchingINSTANCE = FfiConverterOptionalTypeFeatureBatching{}

func (c FfiConverterOptionalTypeFeatureBatching) Lift(rb RustBufferI) *FeatureBatching {
	return LiftFromRustBuffer[*FeatureBatching](c, rb)
}

func (_ FfiConverterOptionalTypeFeatureBatching) Read(reader io.Reader) *FeatureBatching {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeFeatureBatchingINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeFeatureBatching) Lower(value *FeatureBatching) RustBuffer {
	return LowerIntoRustBuffer[*FeatureBatching](c, value)
}

func (_ FfiConverterOptionalTypeFeatureBatching) Write(writer io.Writer, value *FeatureBatching) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeFeatureBatchingINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeFeatureBatching struct {}

func (_ FfiDestroyerOptionalTypeFeatureBatching) Destroy(value *FeatureBatching) {
	if value != nil {
		FfiDestroyerTypeFeatureBatching{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypeFeatureDerp struct{}

var FfiConverterOptionalTypeFeatureDerpINSTANCE = FfiConverterOptionalTypeFeatureDerp{}

func (c FfiConverterOptionalTypeFeatureDerp) Lift(rb RustBufferI) *FeatureDerp {
	return LiftFromRustBuffer[*FeatureDerp](c, rb)
}

func (_ FfiConverterOptionalTypeFeatureDerp) Read(reader io.Reader) *FeatureDerp {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeFeatureDerpINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeFeatureDerp) Lower(value *FeatureDerp) RustBuffer {
	return LowerIntoRustBuffer[*FeatureDerp](c, value)
}

func (_ FfiConverterOptionalTypeFeatureDerp) Write(writer io.Writer, value *FeatureDerp) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeFeatureDerpINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeFeatureDerp struct {}

func (_ FfiDestroyerOptionalTypeFeatureDerp) Destroy(value *FeatureDerp) {
	if value != nil {
		FfiDestroyerTypeFeatureDerp{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypeFeatureDirect struct{}

var FfiConverterOptionalTypeFeatureDirectINSTANCE = FfiConverterOptionalTypeFeatureDirect{}

func (c FfiConverterOptionalTypeFeatureDirect) Lift(rb RustBufferI) *FeatureDirect {
	return LiftFromRustBuffer[*FeatureDirect](c, rb)
}

func (_ FfiConverterOptionalTypeFeatureDirect) Read(reader io.Reader) *FeatureDirect {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeFeatureDirectINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeFeatureDirect) Lower(value *FeatureDirect) RustBuffer {
	return LowerIntoRustBuffer[*FeatureDirect](c, value)
}

func (_ FfiConverterOptionalTypeFeatureDirect) Write(writer io.Writer, value *FeatureDirect) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeFeatureDirectINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeFeatureDirect struct {}

func (_ FfiDestroyerOptionalTypeFeatureDirect) Destroy(value *FeatureDirect) {
	if value != nil {
		FfiDestroyerTypeFeatureDirect{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypeFeatureEndpointProvidersOptimization struct{}

var FfiConverterOptionalTypeFeatureEndpointProvidersOptimizationINSTANCE = FfiConverterOptionalTypeFeatureEndpointProvidersOptimization{}

func (c FfiConverterOptionalTypeFeatureEndpointProvidersOptimization) Lift(rb RustBufferI) *FeatureEndpointProvidersOptimization {
	return LiftFromRustBuffer[*FeatureEndpointProvidersOptimization](c, rb)
}

func (_ FfiConverterOptionalTypeFeatureEndpointProvidersOptimization) Read(reader io.Reader) *FeatureEndpointProvidersOptimization {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeFeatureEndpointProvidersOptimizationINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeFeatureEndpointProvidersOptimization) Lower(value *FeatureEndpointProvidersOptimization) RustBuffer {
	return LowerIntoRustBuffer[*FeatureEndpointProvidersOptimization](c, value)
}

func (_ FfiConverterOptionalTypeFeatureEndpointProvidersOptimization) Write(writer io.Writer, value *FeatureEndpointProvidersOptimization) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeFeatureEndpointProvidersOptimizationINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeFeatureEndpointProvidersOptimization struct {}

func (_ FfiDestroyerOptionalTypeFeatureEndpointProvidersOptimization) Destroy(value *FeatureEndpointProvidersOptimization) {
	if value != nil {
		FfiDestroyerTypeFeatureEndpointProvidersOptimization{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypeFeatureExitDns struct{}

var FfiConverterOptionalTypeFeatureExitDnsINSTANCE = FfiConverterOptionalTypeFeatureExitDns{}

func (c FfiConverterOptionalTypeFeatureExitDns) Lift(rb RustBufferI) *FeatureExitDns {
	return LiftFromRustBuffer[*FeatureExitDns](c, rb)
}

func (_ FfiConverterOptionalTypeFeatureExitDns) Read(reader io.Reader) *FeatureExitDns {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeFeatureExitDnsINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeFeatureExitDns) Lower(value *FeatureExitDns) RustBuffer {
	return LowerIntoRustBuffer[*FeatureExitDns](c, value)
}

func (_ FfiConverterOptionalTypeFeatureExitDns) Write(writer io.Writer, value *FeatureExitDns) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeFeatureExitDnsINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeFeatureExitDns struct {}

func (_ FfiDestroyerOptionalTypeFeatureExitDns) Destroy(value *FeatureExitDns) {
	if value != nil {
		FfiDestroyerTypeFeatureExitDns{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypeFeatureLana struct{}

var FfiConverterOptionalTypeFeatureLanaINSTANCE = FfiConverterOptionalTypeFeatureLana{}

func (c FfiConverterOptionalTypeFeatureLana) Lift(rb RustBufferI) *FeatureLana {
	return LiftFromRustBuffer[*FeatureLana](c, rb)
}

func (_ FfiConverterOptionalTypeFeatureLana) Read(reader io.Reader) *FeatureLana {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeFeatureLanaINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeFeatureLana) Lower(value *FeatureLana) RustBuffer {
	return LowerIntoRustBuffer[*FeatureLana](c, value)
}

func (_ FfiConverterOptionalTypeFeatureLana) Write(writer io.Writer, value *FeatureLana) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeFeatureLanaINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeFeatureLana struct {}

func (_ FfiDestroyerOptionalTypeFeatureLana) Destroy(value *FeatureLana) {
	if value != nil {
		FfiDestroyerTypeFeatureLana{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypeFeatureLinkDetection struct{}

var FfiConverterOptionalTypeFeatureLinkDetectionINSTANCE = FfiConverterOptionalTypeFeatureLinkDetection{}

func (c FfiConverterOptionalTypeFeatureLinkDetection) Lift(rb RustBufferI) *FeatureLinkDetection {
	return LiftFromRustBuffer[*FeatureLinkDetection](c, rb)
}

func (_ FfiConverterOptionalTypeFeatureLinkDetection) Read(reader io.Reader) *FeatureLinkDetection {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeFeatureLinkDetectionINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeFeatureLinkDetection) Lower(value *FeatureLinkDetection) RustBuffer {
	return LowerIntoRustBuffer[*FeatureLinkDetection](c, value)
}

func (_ FfiConverterOptionalTypeFeatureLinkDetection) Write(writer io.Writer, value *FeatureLinkDetection) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeFeatureLinkDetectionINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeFeatureLinkDetection struct {}

func (_ FfiDestroyerOptionalTypeFeatureLinkDetection) Destroy(value *FeatureLinkDetection) {
	if value != nil {
		FfiDestroyerTypeFeatureLinkDetection{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypeFeatureNurse struct{}

var FfiConverterOptionalTypeFeatureNurseINSTANCE = FfiConverterOptionalTypeFeatureNurse{}

func (c FfiConverterOptionalTypeFeatureNurse) Lift(rb RustBufferI) *FeatureNurse {
	return LiftFromRustBuffer[*FeatureNurse](c, rb)
}

func (_ FfiConverterOptionalTypeFeatureNurse) Read(reader io.Reader) *FeatureNurse {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeFeatureNurseINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeFeatureNurse) Lower(value *FeatureNurse) RustBuffer {
	return LowerIntoRustBuffer[*FeatureNurse](c, value)
}

func (_ FfiConverterOptionalTypeFeatureNurse) Write(writer io.Writer, value *FeatureNurse) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeFeatureNurseINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeFeatureNurse struct {}

func (_ FfiDestroyerOptionalTypeFeatureNurse) Destroy(value *FeatureNurse) {
	if value != nil {
		FfiDestroyerTypeFeatureNurse{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypeFeaturePaths struct{}

var FfiConverterOptionalTypeFeaturePathsINSTANCE = FfiConverterOptionalTypeFeaturePaths{}

func (c FfiConverterOptionalTypeFeaturePaths) Lift(rb RustBufferI) *FeaturePaths {
	return LiftFromRustBuffer[*FeaturePaths](c, rb)
}

func (_ FfiConverterOptionalTypeFeaturePaths) Read(reader io.Reader) *FeaturePaths {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeFeaturePathsINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeFeaturePaths) Lower(value *FeaturePaths) RustBuffer {
	return LowerIntoRustBuffer[*FeaturePaths](c, value)
}

func (_ FfiConverterOptionalTypeFeaturePaths) Write(writer io.Writer, value *FeaturePaths) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeFeaturePathsINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeFeaturePaths struct {}

func (_ FfiDestroyerOptionalTypeFeaturePaths) Destroy(value *FeaturePaths) {
	if value != nil {
		FfiDestroyerTypeFeaturePaths{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypeFeaturePmtuDiscovery struct{}

var FfiConverterOptionalTypeFeaturePmtuDiscoveryINSTANCE = FfiConverterOptionalTypeFeaturePmtuDiscovery{}

func (c FfiConverterOptionalTypeFeaturePmtuDiscovery) Lift(rb RustBufferI) *FeaturePmtuDiscovery {
	return LiftFromRustBuffer[*FeaturePmtuDiscovery](c, rb)
}

func (_ FfiConverterOptionalTypeFeaturePmtuDiscovery) Read(reader io.Reader) *FeaturePmtuDiscovery {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeFeaturePmtuDiscoveryINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeFeaturePmtuDiscovery) Lower(value *FeaturePmtuDiscovery) RustBuffer {
	return LowerIntoRustBuffer[*FeaturePmtuDiscovery](c, value)
}

func (_ FfiConverterOptionalTypeFeaturePmtuDiscovery) Write(writer io.Writer, value *FeaturePmtuDiscovery) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeFeaturePmtuDiscoveryINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeFeaturePmtuDiscovery struct {}

func (_ FfiDestroyerOptionalTypeFeaturePmtuDiscovery) Destroy(value *FeaturePmtuDiscovery) {
	if value != nil {
		FfiDestroyerTypeFeaturePmtuDiscovery{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypeFeatureQoS struct{}

var FfiConverterOptionalTypeFeatureQoSINSTANCE = FfiConverterOptionalTypeFeatureQoS{}

func (c FfiConverterOptionalTypeFeatureQoS) Lift(rb RustBufferI) *FeatureQoS {
	return LiftFromRustBuffer[*FeatureQoS](c, rb)
}

func (_ FfiConverterOptionalTypeFeatureQoS) Read(reader io.Reader) *FeatureQoS {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeFeatureQoSINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeFeatureQoS) Lower(value *FeatureQoS) RustBuffer {
	return LowerIntoRustBuffer[*FeatureQoS](c, value)
}

func (_ FfiConverterOptionalTypeFeatureQoS) Write(writer io.Writer, value *FeatureQoS) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeFeatureQoSINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeFeatureQoS struct {}

func (_ FfiDestroyerOptionalTypeFeatureQoS) Destroy(value *FeatureQoS) {
	if value != nil {
		FfiDestroyerTypeFeatureQoS{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypeFeatureSkipUnresponsivePeers struct{}

var FfiConverterOptionalTypeFeatureSkipUnresponsivePeersINSTANCE = FfiConverterOptionalTypeFeatureSkipUnresponsivePeers{}

func (c FfiConverterOptionalTypeFeatureSkipUnresponsivePeers) Lift(rb RustBufferI) *FeatureSkipUnresponsivePeers {
	return LiftFromRustBuffer[*FeatureSkipUnresponsivePeers](c, rb)
}

func (_ FfiConverterOptionalTypeFeatureSkipUnresponsivePeers) Read(reader io.Reader) *FeatureSkipUnresponsivePeers {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeFeatureSkipUnresponsivePeersINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeFeatureSkipUnresponsivePeers) Lower(value *FeatureSkipUnresponsivePeers) RustBuffer {
	return LowerIntoRustBuffer[*FeatureSkipUnresponsivePeers](c, value)
}

func (_ FfiConverterOptionalTypeFeatureSkipUnresponsivePeers) Write(writer io.Writer, value *FeatureSkipUnresponsivePeers) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeFeatureSkipUnresponsivePeersINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeFeatureSkipUnresponsivePeers struct {}

func (_ FfiDestroyerOptionalTypeFeatureSkipUnresponsivePeers) Destroy(value *FeatureSkipUnresponsivePeers) {
	if value != nil {
		FfiDestroyerTypeFeatureSkipUnresponsivePeers{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypeLinkState struct{}

var FfiConverterOptionalTypeLinkStateINSTANCE = FfiConverterOptionalTypeLinkState{}

func (c FfiConverterOptionalTypeLinkState) Lift(rb RustBufferI) *LinkState {
	return LiftFromRustBuffer[*LinkState](c, rb)
}

func (_ FfiConverterOptionalTypeLinkState) Read(reader io.Reader) *LinkState {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeLinkStateINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeLinkState) Lower(value *LinkState) RustBuffer {
	return LowerIntoRustBuffer[*LinkState](c, value)
}

func (_ FfiConverterOptionalTypeLinkState) Write(writer io.Writer, value *LinkState) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeLinkStateINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeLinkState struct {}

func (_ FfiDestroyerOptionalTypeLinkState) Destroy(value *LinkState) {
	if value != nil {
		FfiDestroyerTypeLinkState{}.Destroy(*value)
	}
}



type FfiConverterOptionalTypePathType struct{}

var FfiConverterOptionalTypePathTypeINSTANCE = FfiConverterOptionalTypePathType{}

func (c FfiConverterOptionalTypePathType) Lift(rb RustBufferI) *PathType {
	return LiftFromRustBuffer[*PathType](c, rb)
}

func (_ FfiConverterOptionalTypePathType) Read(reader io.Reader) *PathType {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypePathTypeINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypePathType) Lower(value *PathType) RustBuffer {
	return LowerIntoRustBuffer[*PathType](c, value)
}

func (_ FfiConverterOptionalTypePathType) Write(writer io.Writer, value *PathType) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypePathTypeINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypePathType struct {}

func (_ FfiDestroyerOptionalTypePathType) Destroy(value *PathType) {
	if value != nil {
		FfiDestroyerTypePathType{}.Destroy(*value)
	}
}



type FfiConverterOptionalSequenceTypePeer struct{}

var FfiConverterOptionalSequenceTypePeerINSTANCE = FfiConverterOptionalSequenceTypePeer{}

func (c FfiConverterOptionalSequenceTypePeer) Lift(rb RustBufferI) *[]Peer {
	return LiftFromRustBuffer[*[]Peer](c, rb)
}

func (_ FfiConverterOptionalSequenceTypePeer) Read(reader io.Reader) *[]Peer {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterSequenceTypePeerINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalSequenceTypePeer) Lower(value *[]Peer) RustBuffer {
	return LowerIntoRustBuffer[*[]Peer](c, value)
}

func (_ FfiConverterOptionalSequenceTypePeer) Write(writer io.Writer, value *[]Peer) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterSequenceTypePeerINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalSequenceTypePeer struct {}

func (_ FfiDestroyerOptionalSequenceTypePeer) Destroy(value *[]Peer) {
	if value != nil {
		FfiDestroyerSequenceTypePeer{}.Destroy(*value)
	}
}



type FfiConverterOptionalSequenceTypeServer struct{}

var FfiConverterOptionalSequenceTypeServerINSTANCE = FfiConverterOptionalSequenceTypeServer{}

func (c FfiConverterOptionalSequenceTypeServer) Lift(rb RustBufferI) *[]Server {
	return LiftFromRustBuffer[*[]Server](c, rb)
}

func (_ FfiConverterOptionalSequenceTypeServer) Read(reader io.Reader) *[]Server {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterSequenceTypeServerINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalSequenceTypeServer) Lower(value *[]Server) RustBuffer {
	return LowerIntoRustBuffer[*[]Server](c, value)
}

func (_ FfiConverterOptionalSequenceTypeServer) Write(writer io.Writer, value *[]Server) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterSequenceTypeServerINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalSequenceTypeServer struct {}

func (_ FfiDestroyerOptionalSequenceTypeServer) Destroy(value *[]Server) {
	if value != nil {
		FfiDestroyerSequenceTypeServer{}.Destroy(*value)
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

func (c FfiConverterOptionalSequenceTypeIpAddr) Lower(value *[]IpAddr) RustBuffer {
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

func (c FfiConverterOptionalSequenceTypeIpNet) Lower(value *[]IpNet) RustBuffer {
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

func (c FfiConverterOptionalTypeEndpointProviders) Lower(value *EndpointProviders) RustBuffer {
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

func (c FfiConverterOptionalTypeHiddenString) Lower(value *HiddenString) RustBuffer {
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

func (c FfiConverterOptionalTypeSocketAddr) Lower(value *SocketAddr) RustBuffer {
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



type FfiConverterSequenceTypePeer struct{}

var FfiConverterSequenceTypePeerINSTANCE = FfiConverterSequenceTypePeer{}

func (c FfiConverterSequenceTypePeer) Lift(rb RustBufferI) []Peer {
	return LiftFromRustBuffer[[]Peer](c, rb)
}

func (c FfiConverterSequenceTypePeer) Read(reader io.Reader) []Peer {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]Peer, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypePeerINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypePeer) Lower(value []Peer) RustBuffer {
	return LowerIntoRustBuffer[[]Peer](c, value)
}

func (c FfiConverterSequenceTypePeer) Write(writer io.Writer, value []Peer) {
	if len(value) > math.MaxInt32 {
		panic("[]Peer is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypePeerINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypePeer struct {}

func (FfiDestroyerSequenceTypePeer) Destroy(sequence []Peer) {
	for _, value := range sequence {
		FfiDestroyerTypePeer{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeServer struct{}

var FfiConverterSequenceTypeServerINSTANCE = FfiConverterSequenceTypeServer{}

func (c FfiConverterSequenceTypeServer) Lift(rb RustBufferI) []Server {
	return LiftFromRustBuffer[[]Server](c, rb)
}

func (c FfiConverterSequenceTypeServer) Read(reader io.Reader) []Server {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]Server, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeServerINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeServer) Lower(value []Server) RustBuffer {
	return LowerIntoRustBuffer[[]Server](c, value)
}

func (c FfiConverterSequenceTypeServer) Write(writer io.Writer, value []Server) {
	if len(value) > math.MaxInt32 {
		panic("[]Server is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeServerINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeServer struct {}

func (FfiDestroyerSequenceTypeServer) Destroy(sequence []Server) {
	for _, value := range sequence {
		FfiDestroyerTypeServer{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeTelioNode struct{}

var FfiConverterSequenceTypeTelioNodeINSTANCE = FfiConverterSequenceTypeTelioNode{}

func (c FfiConverterSequenceTypeTelioNode) Lift(rb RustBufferI) []TelioNode {
	return LiftFromRustBuffer[[]TelioNode](c, rb)
}

func (c FfiConverterSequenceTypeTelioNode) Read(reader io.Reader) []TelioNode {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]TelioNode, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeTelioNodeINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeTelioNode) Lower(value []TelioNode) RustBuffer {
	return LowerIntoRustBuffer[[]TelioNode](c, value)
}

func (c FfiConverterSequenceTypeTelioNode) Write(writer io.Writer, value []TelioNode) {
	if len(value) > math.MaxInt32 {
		panic("[]TelioNode is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeTelioNodeINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeTelioNode struct {}

func (FfiDestroyerSequenceTypeTelioNode) Destroy(sequence []TelioNode) {
	for _, value := range sequence {
		FfiDestroyerTypeTelioNode{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeEndpointProvider struct{}

var FfiConverterSequenceTypeEndpointProviderINSTANCE = FfiConverterSequenceTypeEndpointProvider{}

func (c FfiConverterSequenceTypeEndpointProvider) Lift(rb RustBufferI) []EndpointProvider {
	return LiftFromRustBuffer[[]EndpointProvider](c, rb)
}

func (c FfiConverterSequenceTypeEndpointProvider) Read(reader io.Reader) []EndpointProvider {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]EndpointProvider, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeEndpointProviderINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeEndpointProvider) Lower(value []EndpointProvider) RustBuffer {
	return LowerIntoRustBuffer[[]EndpointProvider](c, value)
}

func (c FfiConverterSequenceTypeEndpointProvider) Write(writer io.Writer, value []EndpointProvider) {
	if len(value) > math.MaxInt32 {
		panic("[]EndpointProvider is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeEndpointProviderINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeEndpointProvider struct {}

func (FfiDestroyerSequenceTypeEndpointProvider) Destroy(sequence []EndpointProvider) {
	for _, value := range sequence {
		FfiDestroyerTypeEndpointProvider{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypePathType struct{}

var FfiConverterSequenceTypePathTypeINSTANCE = FfiConverterSequenceTypePathType{}

func (c FfiConverterSequenceTypePathType) Lift(rb RustBufferI) []PathType {
	return LiftFromRustBuffer[[]PathType](c, rb)
}

func (c FfiConverterSequenceTypePathType) Read(reader io.Reader) []PathType {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]PathType, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypePathTypeINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypePathType) Lower(value []PathType) RustBuffer {
	return LowerIntoRustBuffer[[]PathType](c, value)
}

func (c FfiConverterSequenceTypePathType) Write(writer io.Writer, value []PathType) {
	if len(value) > math.MaxInt32 {
		panic("[]PathType is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypePathTypeINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypePathType struct {}

func (FfiDestroyerSequenceTypePathType) Destroy(sequence []PathType) {
	for _, value := range sequence {
		FfiDestroyerTypePathType{}.Destroy(value)	
	}
}



type FfiConverterSequenceTypeRttType struct{}

var FfiConverterSequenceTypeRttTypeINSTANCE = FfiConverterSequenceTypeRttType{}

func (c FfiConverterSequenceTypeRttType) Lift(rb RustBufferI) []RttType {
	return LiftFromRustBuffer[[]RttType](c, rb)
}

func (c FfiConverterSequenceTypeRttType) Read(reader io.Reader) []RttType {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]RttType, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeRttTypeINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeRttType) Lower(value []RttType) RustBuffer {
	return LowerIntoRustBuffer[[]RttType](c, value)
}

func (c FfiConverterSequenceTypeRttType) Write(writer io.Writer, value []RttType) {
	if len(value) > math.MaxInt32 {
		panic("[]RttType is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeRttTypeINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeRttType struct {}

func (FfiDestroyerSequenceTypeRttType) Destroy(sequence []RttType) {
	for _, value := range sequence {
		FfiDestroyerTypeRttType{}.Destroy(value)	
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

func (c FfiConverterSequenceTypeIpAddr) Lower(value []IpAddr) RustBuffer {
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

func (c FfiConverterSequenceTypeIpNet) Lower(value []IpNet) RustBuffer {
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
type FfiConverterTypeEndpointProviders = FfiConverterSequenceTypeEndpointProvider
type FfiDestroyerTypeEndpointProviders = FfiDestroyerSequenceTypeEndpointProvider
var FfiConverterTypeEndpointProvidersINSTANCE = FfiConverterSequenceTypeEndpointProvider{}


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

// Utility function to create a `Features` object from a json-string
// Passing an empty string will return the default feature config
func DeserializeFeatureConfig(fstr string) (Features, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_telio_fn_func_deserialize_feature_config(FfiConverterStringINSTANCE.Lower(fstr), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue Features
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTypeFeaturesINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

// Utility function to create a `Config` object from a json-string
func DeserializeMeshnetConfig(cfgStr string) (Config, error) {
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeTelioError{},func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_telio_fn_func_deserialize_meshnet_config(FfiConverterStringINSTANCE.Lower(cfgStr), _uniffiStatus)
	})
		if _uniffiErr != nil {
			var _uniffiDefaultValue Config
			return _uniffiDefaultValue, _uniffiErr
		} else {
			return FfiConverterTypeConfigINSTANCE.Lift(_uniffiRV), _uniffiErr
		}
}

// Get the public key that corresponds to a given private key.
func GeneratePublicKey(secretKey SecretKey) PublicKey {
	return FfiConverterTypePublicKeyINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_telio_fn_func_generate_public_key(FfiConverterTypeSecretKeyINSTANCE.Lower(secretKey), _uniffiStatus)
	}))
}

// Generate a new secret key.
func GenerateSecretKey() SecretKey {
	return FfiConverterTypeSecretKeyINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_telio_fn_func_generate_secret_key( _uniffiStatus)
	}))
}

// Get current commit sha.
func GetCommitSha() string {
	return FfiConverterStringINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_telio_fn_func_get_commit_sha( _uniffiStatus)
	}))
}

// Get default recommended adapter type for platform.
func GetDefaultAdapter() TelioAdapterType {
	return FfiConverterTypeTelioAdapterTypeINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_telio_fn_func_get_default_adapter( _uniffiStatus)
	}))
}

// Utility function to get the default feature config
func GetDefaultFeatureConfig() Features {
	return FfiConverterTypeFeaturesINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_telio_fn_func_get_default_feature_config( _uniffiStatus)
	}))
}

// Get current version tag.
func GetVersionTag() string {
	return FfiConverterStringINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_telio_fn_func_get_version_tag( _uniffiStatus)
	}))
}

// Set the global logger.
// # Parameters
// - `log_level`: Max log level to log.
// - `logger`: Callback to handle logging events.
func SetGlobalLogger(logLevel TelioLogLevel, logger TelioLoggerCb)  {
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_telio_fn_func_set_global_logger(FfiConverterTypeTelioLogLevelINSTANCE.Lower(logLevel), FfiConverterCallbackInterfaceTelioLoggerCbINSTANCE.Lower(logger), _uniffiStatus)
		return false
	})
}

