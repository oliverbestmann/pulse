package pulse

import (
	"os"
	"runtime"
	"strings"

	"github.com/cogentcore/webgpu/wgpu"
)

var forceFallbackAdapter = os.Getenv("WGPU_FORCE_FALLBACK_ADAPTER") == "1"

func init() {
	runtime.LockOSThread()

	switch strings.ToUpper(os.Getenv("WGPU_LOG_LEVEL")) {
	case "OFF":
		wgpu.SetLogLevel(wgpu.LogLevelOff)
	case "ERROR":
		wgpu.SetLogLevel(wgpu.LogLevelError)
	case "WARN":
		wgpu.SetLogLevel(wgpu.LogLevelWarn)
	case "INFO":
		wgpu.SetLogLevel(wgpu.LogLevelInfo)
	case "DEBUG":
		wgpu.SetLogLevel(wgpu.LogLevelDebug)
	case "TRACE":
		wgpu.SetLogLevel(wgpu.LogLevelTrace)
	}
}

// Context encapsulates the low level state of the webgpu context,
// this includes the Device, Surface and active Adapter
type Context struct {
	*wgpu.Device
	*wgpu.Queue
	Surface *wgpu.Surface
	Adapter *wgpu.Adapter
}

func New(sd *wgpu.SurfaceDescriptor) (st *Context, err error) {
	defer func() {
		if err != nil && st != nil {
			st.Release()
			st = nil
		}
	}()

	st = &Context{}

	// create the webgpu instance
	instance := wgpu.CreateInstance(nil)
	defer instance.Release()

	// create a Surface based on the window
	st.Surface = instance.CreateSurface(sd)

	// create an adapter that can render to the Surface
	st.Adapter, err = instance.RequestAdapter(&wgpu.RequestAdapterOptions{
		ForceFallbackAdapter: forceFallbackAdapter,
		CompatibleSurface:    st.Surface,
	})

	if err != nil {
		return
	}

	// get a Device with the default settings
	st.Device, err = st.Adapter.RequestDevice(nil)
	if err != nil {
		return
	}

	st.Queue = st.Device.GetQueue()

	return st, nil
}

func (d *Context) Release() {
	if d.Queue != nil {
		d.Queue.Release()
		d.Queue = nil
	}

	if d.Device != nil {
		d.Device.Release()
		d.Device = nil
	}

	if d.Adapter != nil {
		d.Adapter.Release()
		d.Adapter = nil
	}

	if d.Surface != nil {
		d.Surface.Release()
		d.Surface = nil
	}
}
