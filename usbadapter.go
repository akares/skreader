package skreader

import "io"

// UsbAdapter is an interface that used to abstract USB communication.
// It is used by SEKONIC Device handler to communicate with device.
// It is not intended to be used by end user exept of initial injecting as dependency to Device handler.
// It is not mandatory to make it thread safe, it is (hopefully) done by Device handler.
type UsbAdapter interface {
	UsbAdapterOpenerCloser
	UsbAdapterReaderWriter
	UsbAdapterDescriber
}

type UsbAdapterOpenerCloser interface {
	Open() error // Open must ensure that device is in correct state for communication or return an error.
	io.Closer    // Close must release all previously allocated resources.
}

type UsbAdapterReaderWriter interface {
	io.Reader // Read must read raw binary data from device or an error.
	io.Writer // Write must ensure the raw binary data is correctly sent to device or return an error.
}

type UsbAdapterDescriber interface {
	Manufacturer() (string, error) // Manufacturer must return device manufacturer name or empty string or an error.
	Product() (string, error)      // Product must return device product name or empty string or an error.
}
