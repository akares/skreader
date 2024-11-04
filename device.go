package skreader

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

const (
	WaitConnTimeoutDefault = time.Duration(5) * time.Second       // how long to wait for device to start measuring
	WaitMeasTimeoutDefault = time.Duration(20) * time.Second      // how long to wait for device to end measuring
	WaitPollFreqDefault    = time.Duration(50) * time.Millisecond // how often to poll device for status

	ReadBufSize = MeasurementDataValidSize // MeasurementDataValidSize bytes is currently the maximum size of the data that can be sent by the device
)

var (
	SkResponseOK = []byte{6, 48} // ACK response from device when command executed successfully
)

// Device represents SEKONIC device handler.
type Device struct {
	adapter UsbAdapter

	Manufacturer string
	Product      string

	MeasurementConfig DeviceMeasurementConfig // Currently supported only by C-7000

	mu sync.Mutex
}

// DeviceState represents device current operational mode, knobs and buttons states.
type DeviceState struct {
	Status SkDeviceStatus
	Remote SkRemoteStatus
	Button SkButtonStatus
	Ring   SkRingStatus
}

// DeviceMeasurementConfig represents configuration used for measurement.
type DeviceMeasurementConfig struct {
	MeasuringMode SkMeasuringMode // only ambient is supported yet
	FieldOfView   SkFieldOfView
	ExposureTime  SkExposureTime
	ShutterSpeed  SkShutterSpeed
}

// NewDeviceWithAdapter creates SEKONIC device handler using provided UsbAdapter concrete implementation.
// Running this function will ensure that device is connected and will read its minimal USB device info.
//
// Sending commands to device is done by calling methods on returned Device struct.
//
// Command execution is synchronized, so only one command can be executed at a time
// and (hopefully) is safe for concurrent use by multiple goroutines within one Device instance.
// Note that even if commands sent in parallel will be executed in sequence, it still can put the device
// into incorrect logical state.
//
// After using device, Close method must be called to release all allocated resources.
// Until Close is called, device may not be available for other processes.
func NewDeviceWithAdapter(adapter UsbAdapter) (*Device, error) {
	if adapter == nil {
		return nil, fmt.Errorf("adapter is nil")
	}

	err := adapter.Open()
	if err != nil {
		return nil, err
	}

	manufacturer, err := adapter.Manufacturer()
	if err != nil {
		return nil, err
	}

	product, err := adapter.Product()
	if err != nil {
		return nil, err
	}

	// Default measurement configuration with recommended values.
	defaultConfig := DeviceMeasurementConfig{
		MeasuringMode: SkMeasuringModeAmbient,
		FieldOfView:   SkFieldOfView2Deg,
		ExposureTime:  SkExposureTimeAuto,
		ShutterSpeed:  SkShutterSpeed125Sec,
	}

	return &Device{ //nolint:exhaustruct
		adapter:           adapter,
		Manufacturer:      manufacturer,
		Product:           product,
		MeasurementConfig: defaultConfig,
	}, nil
}

// String returns device readble name. It tries to use Manufacturer and Product, but if any of them
// is empty, it uses "SEKONIC" dummy name.
func (d *Device) String() string {
	if d.Product != "" && d.Manufacturer != "" {
		return fmt.Sprintf("%s %s", d.Manufacturer, d.Product)
	}
	if d.Manufacturer != "" {
		return d.Manufacturer
	}
	if d.Product != "" {
		return d.Product
	}

	return "SEKONIC"
}

// Measure performs one measurement and returns result.
func (d *Device) Measure() (*Measurement, error) {
	err := d.WaitReady(WaitConnTimeoutDefault, WaitPollFreqDefault)
	if err != nil {
		return nil, err
	}

	err = d.SetRemoteOn()
	if err != nil {
		return nil, err
	}
	defer func() { _ = d.SetRemoteOff() }()

	err = d.SetMeasurementConfiguration()
	if err != nil {
		return nil, err
	}

	err = d.StartMeasuring()
	if err != nil {
		return nil, err
	}

	err = d.WaitReady(WaitMeasTimeoutDefault, WaitPollFreqDefault)
	if err != nil {
		return nil, err
	}

	return d.MeasurementResult()
}



// WaitReady waits for device to be ready for next measurement.
// It polls device state every step duration until idle status is reached or timeout duration is reached.
// If device state is not valid for measurement, error is returned.
// If timeout is reached, error is returned.
func (d *Device) WaitReady(duration, step time.Duration) error {
	timeout := time.After(duration)
	ticker := time.NewTicker(step)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			st, e := d.State()
			if e != nil {
				continue // ignore status read error, will repeat in next tick
			}
			if st.Ring != SkRingStatusLow {
				return fmt.Errorf("ring is not set to low position")
			}
			if st.Button == SkButtonStatusMeasuring {
				return fmt.Errorf("measuring button is pressed")
			}
			if st.Status == SkDeviceStatusIdle ||
				st.Status == SkDeviceStatusIdleOutMeas {
				return nil // waiting succeeded
			}
		case <-timeout:
			return fmt.Errorf("timeout waiting for device to end measuring (%s)", duration)
		}
	}
}

// MeasurementResult requests device measurement result data.
func (d *Device) MeasurementResult() (*Measurement, error) {
	// Response data example:
	// NR@@@ + data
	data, err := d.execCommand(SkCommandGetMeasurementResult, 0, 0)
	if err != nil {
		return nil, err
	}

	return NewMeasurementFromBytes(data)
}

// ModelName requests device model name.
func (d *Device) ModelName() (string, error) {
	// Response data example (chars):
	// MN@@@C-800\x00\x00\x00\x00\x00
	//      ^ Model Name chars start at pos 5, end randomly with bunch of trailing null bytes
	const (
		cmd     = SkCommandGetModelNumber
		datapos = 5
		datalen = 0
	)
	data, err := d.execCommand(cmd, datapos, datalen) // -> "C-800\x00\x00\x00\x00\x00"
	if err != nil {
		return "", err
	}

	return toString(data), nil // -> "C-800"
}

// FirmwareVersion requests device main firmware version.
func (d *Device) FirmwareVersion() (int, error) {
	// Response data example (chars):
	// FV@@@20,C36E,27,7881,11,B216,14,50CC,17,74EC
	//              27 <- main FW version chars are at pos 13 and 14 (used for feature detection)
	//                      11 <-   14 <- other FW versions (not used by this pkg)
	const (
		cmd     = SkCommandGetFirmwareVersion
		datapos = 13
		datalen = 2
	)
	data, err := d.execCommand(cmd, datapos, datalen) // -> "27"
	if err != nil {
		return 0, err
	}
	ver, err := toInt(data)
	if err != nil {
		return 0, fmt.Errorf("%s command error: invalid response: %v: %s", cmd, data, err)
	}

	return ver, nil // -> 27
}

// State requests device current operational mode, knobs and buttons states.
func (d *Device) State() (*DeviceState, error) {
	// Response data example (chars)
	// ST@@@
	// Response data example (bytes):
	// [83  84 64  64  64]
	//  S   T  st1 st2 key
	const (
		cmd     = SkCommandGetStatus
		datapos = 2
		datalen = 3
	)
	data, err := d.execCommand(cmd, datapos, datalen) // -> [st1 st2 key]
	if err != nil {
		return nil, err
	}

	st1 := data[0]
	st2 := data[1]
	key := data[2]

	var (
		status SkDeviceStatus
		remote SkRemoteStatus
		button SkButtonStatus
		ring   SkRingStatus
	)

	// A bit of bit shenanigans fun

	if (st1 & 0x10) != 0 {
		status = SkDeviceStatusErrorHw
	} else if (st1 & 1) != 0 {
		if (st2 & 1) != 0 {
			status = SkDeviceStatusBusyInitializing
		} else if (st2 & 4) != 0 {
			status = SkDeviceStatusBusyDarkCalibration
		} else if (st2 & 0x10) != 0 {
			status = SkDeviceStatusBusyFlashStandby
		} else if (st2 & 8) != 0 {
			status = SkDeviceStatusBusyMeasuring
		}
	} else if (st1 & 8) != 0 {
		status = SkDeviceStatusIdleOutMeas
	} else {
		status = SkDeviceStatusIdle
	}
	if (st1 & 2) != 0 {
		remote = SkRemoteStatusOn
	}
	button = SkButtonStatus(key & 0x1F)    //nolint:gomnd
	ring = SkRingStatus((key & 0x60) >> 5) //nolint:gomnd

	// Fun is over

	return &DeviceState{
		Status: status,
		Remote: remote,
		Button: button,
		Ring:   ring,
	}, nil
}

// SetMeasurementConfiguration sends measurement configuration options to device.
func (d *Device) SetMeasurementConfiguration() error {
	if !d.SupportsMeasurementConfiguration() {
		return nil
	}

	const tag = "set measurement configuration"

	setMeasuringMode := fmt.Sprintf("%s,%d", SkCommandSetMeasuringMode, d.MeasurementConfig.MeasuringMode)
	if _, err := d.execCommand(SkCommand(setMeasuringMode), 0, 0); err != nil {
		return fmt.Errorf("%s: set measurement mode error: %s", tag, err)
	}

	setShutterSpeed := fmt.Sprintf("%s,0,%s", SkCommandSetShutterSpeed, d.MeasurementConfig.ShutterSpeed)
	if _, err := d.execCommand(SkCommand(setShutterSpeed), 0, 0); err != nil {
		return fmt.Errorf("%s: set shutter speed error: %s", tag, err)
	}

	if !d.SupportsExtendedMeasurementConfiguration() {
		return nil
	}

	setFov := fmt.Sprintf("%s,%d", SkCommandSetFov, d.MeasurementConfig.FieldOfView)
	if _, err := d.execCommand(SkCommand(setFov), 0, 0); err != nil {
		return fmt.Errorf("%s: set field of view error: %s", tag, err)
	}

	setExposureTime := fmt.Sprintf("%s,%d", SkCommandSetExposureTime, d.MeasurementConfig.ExposureTime)
	if _, err := d.execCommand(SkCommand(setExposureTime), 0, 0); err != nil {
		return fmt.Errorf("%s: set exposure time error: %s", tag, err)
	}

	return nil
}

// SetRemoteOn sets device to remote control mode.
// In this mode, device is ready to receive remote commands.
func (d *Device) SetRemoteOn() error {
	_, err := d.execCommand(SkCommandSetRemoteOn, 0, 0)

	return err
}

// SetRemoteOff sets device back to normal control mode.
// In this mode, device is ready to be used manually.
func (d *Device) SetRemoteOff() error {
	_, err := d.execCommand(SkCommandSetRemoteOff, 0, 0)

	return err
}

// StartMeasuring sends command to device to start measuring.
func (d *Device) StartMeasuring() error {
	_, err := d.execCommand(SkCommandStartMeasuring, 0, 0)

	return err
}

// Close releases all allocated resources.
// After calling this method, device handle is no longer usable.
// There is no separate Open method, because it is done implicitly by NewFromAdapter.
func (d *Device) Close() error {
	return d.adapter.Close()
}

// execCommand sends SkCommand to device and reads response. Parameters datapos and datalen are used to extract
// only necessary bytes from response buffer. If datalen is 0, whole response data is returned.
func (d *Device) execCommand(cmd SkCommand, datapos, datalen int) ([]byte, error) {
	// Ensure only one command at a time
	d.mu.Lock()
	defer d.mu.Unlock()

	cmdbytes := []byte(cmd)

	// Send command to device
	err := d.write(cmdbytes)
	if err != nil {
		return nil, err
	}

	// Read acknowledge response
	data, err := d.read()
	if err != nil {
		return nil, err
	}

	// Check acknowledge response is OK
	if !bytes.Equal(data, SkResponseOK) {
		return nil, fmt.Errorf("not OK response: %v", data)
	}

	// Read main response
	data, err = d.read()
	if err != nil {
		return nil, err
	}

	// Boundaries check
	if len(data) < datapos+datalen {
		return nil, fmt.Errorf("%s command error: invalid response length %d (min is %d)", cmd, len(data), datapos+datalen)
	}

	// Check response is the current command sent response.
	// Compare first 2 chars only because in response some commands may vary third char.
	if !bytes.Equal(data[0:2], cmdbytes[0:2]) {
		return nil, fmt.Errorf("not command response: %v (%s)", data[0:len(cmdbytes)], string(data[0:len(cmdbytes)]))
	}

	if datalen == 0 {
		return data[datapos:], nil // return whole length of response data
	}

	return data[datapos : datapos+datalen], nil // return only requested length of response data
}

// read reads raw binary data from device.
func (d *Device) read() (buf []byte, err error) {
	buf = make([]byte, ReadBufSize)

	n, err := d.adapter.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("IN endpoint returned an error: %s", err)
	}
	if n == 0 {
		return nil, errors.New("IN endpoint returned 0 bytes")
	}

	return buf[:n], nil
}

// write sends raw binary data to device.
func (d *Device) write(buf []byte) error {
	n, err := d.adapter.Write(buf)
	if err != nil {
		return fmt.Errorf("OUT endpoint returned an error: %v", err)
	}
	if n < len(buf) {
		return fmt.Errorf("OUT endpoint wrote %d bytes only, which is less than data size %d bytes", n, len(buf))
	}

	return nil
}

// SupportsMeasurementConfiguration reports whether device supports
// measurement configuration.
// Tested C-700, C-800, C-7000 models.
func (d *Device) SupportsMeasurementConfiguration() bool {
	model, _ := d.ModelName()

	return model == "C-7000" || model == "C-800"
}

// SupportsExtendedMeasurementConfiguration reports whether device supports
// extended measurement configuration.
// Tested C-700, C-800, C-7000 models.
func (d *Device) SupportsExtendedMeasurementConfiguration() bool {
	model, _ := d.ModelName()
	fw, _ := d.FirmwareVersion()

	return model == "C-7000" && fw > 25
}

// toString converts byte slice to string, ignoring everything after first null byte.
// If there is no null byte, whole slice is converted to string.
// This is used to convert device response data to string when it is required.
// Example: []byte{65, 66, 67, 0, 0, 0} -> "ABC"
func toString(data []byte) string {
	end := len(data)
	if nullPos := bytes.IndexByte(data[:end], 0); nullPos >= 0 {
		if nullPos < end {
			end = nullPos
		}
	}

	return string(data[:end])
}

// toInt converts byte slice to int.
// Even though it is simple wrapper around strconv.Atoi, it is used to make code more readable.
func toInt(data []byte) (int, error) {
	return strconv.Atoi(string(data))
}
