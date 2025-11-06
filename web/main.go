//go:build js

package main

import (
	"fmt"
	"runtime"
	"structs"
	"syscall/js"
	"unsafe"

	"github.com/cogentcore/webgpu/wgpu"
)

//go:wasmimport wgpu_native wgpuGetVersion
func wgpuGetVersion() uint32

//go:wasmimport wgpu_native wgpuSetLogLevel
func wgpuSetLogLevel(level wgpu.LogLevel)

//go:wasmimport wgpu_native wgpuSetLogCallback
func wgpuSetLogCallback(callback uint32, userdata unsafe.Pointer)

//go:wasmimport wgpu_native wgpuCreateInstance
func wgpuCreateInstance(desc unsafe.Pointer) uint32

//go:wasmimport wgpu_native wgpuInstanceRelease
func wgpuInstanceRelease(instance uint32)

//go:wasmimport wgpu_native wgpuInstanceRequestAdapter
func wgpuInstanceRequestAdapter(instance uint32, options uint32, callbackInfo uint32) uint64

//go:wasmexport logCallback
func logCallback(level, str, userdata uint32) {
	type WGPUStringView struct {
		Content uint32
		Length  uint32
	}

	strView := readValueFromWGPU[WGPUStringView](str)
	text := readMemoryFromWGPU(strView.Content, strView.Length)

	fmt.Println("[WGPU]", string(text))
}

type WGPUChainedStruct uint32
type WGPUCallbackMode uint32

type WGPURequestAdapterCallback uint32

type WGPURequestAdapterCallbackInfo struct {
	_                          structs.HostLayout
	NextInChain                WGPUChainedStruct
	Mode                       WGPUCallbackMode //
	WGPURequestAdapterCallback uint32
	Userdata1                  uint32
	Userdata2                  uint32
}

//go:wasmexport requestAdapterCallback
func requestAdapterCallback(status, adapter, message, userdata1, userdata2 uint32) {
	fmt.Println("Callback", status, adapter, message, userdata1, userdata2)
}

func readMemoryFromWGPU(offset, len uint32) []byte {
	dest := make([]byte, len)
	buf := memWGPU.Invoke(offset, len)
	js.CopyBytesToGo(dest, buf)
	return dest
}

func readValueFromWGPU[T any](offset uint32) *T {
	var tZero T

	n := uint32(unsafe.Sizeof(tZero))
	buf := readMemoryFromWGPU(offset, n)
	return (*T)(unsafe.Pointer(&buf[0]))
}

func memoryToWGPU(offset uint32, buf []byte) {
	dest := memWGPU.Invoke(offset, len(buf))
	js.CopyBytesToJS(dest, buf)
}

func valueToWGPU[T any](offset uint32, value *T) {
	defer runtime.KeepAlive(value)

	n := unsafe.Sizeof(value)
	buf := unsafe.Slice((*byte)(unsafe.Pointer(value)), n)
	memoryToWGPU(offset, buf)
}

var memWGPU = js.Global().Call("eval", `
	(addr, len) => new Uint8Array(wgpu.exports.memory.buffer, addr, len)
`)

func main() {
	// dirty hack to get a place in the table. Hopefully,  no one
	// will ever call $table0[1000]
	js.Global().Call("eval", `
		wgpu.exports.__indirect_function_table.set(1000, wasm.instance.exports.logCallback)
		wgpu.exports.__indirect_function_table.set(1001, wasm.instance.exports.requestAdapterCallback)
	`)

	fmt.Println(wgpuGetVersion())
	wgpuSetLogLevel(wgpu.LogLevelTrace)
	wgpuSetLogCallback(1000, nil)

	instance := wgpuCreateInstance(nil)
	defer wgpuInstanceRelease(instance)

	fmt.Println("Instance:", instance)

	opts := WGPURequestAdapterCallbackInfo{
		NextInChain:                0,
		Mode:                       0,
		WGPURequestAdapterCallback: 1001,
		Userdata1:                  0x1111,
		Userdata2:                  0x2222,
	}

	valueToWGPU(10000, &opts)

	wgpuInstanceRequestAdapter(instance, 0, 10000)
}
