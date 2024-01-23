package skreader

import (
	"errors"
	"fmt"

	"github.com/google/gousb"
)

const (
	IDVendor  = 0x0A41
	IDProduct = 0x7003

	EndpointNumOut = 0x02
	EndpointNumIn  = 0x81
)

var _ UsbAdapter = (*GousbAdapter)(nil) // assert it implements UsbAdapter interface

// GousbAdapter implements UsbAdapter interface using `gousb` library.
// The gousb package is a Google's Go-like wrapper around `libusb` library which is required to be
// installed on the target system. See more: https://github.com/libusb/libusb/wiki
type GousbAdapter struct {
	ctx      *gousb.Context
	dev      *gousb.Device
	intf     *gousb.Interface
	intfDone func()
	epIn     *gousb.InEndpoint
	epOut    *gousb.OutEndpoint
}

// Expose all used gousb functions as variables to be able to mock them in tests.
var (
	GousbNewContext = func() *gousb.Context { //nolint:all
		return gousb.NewContext()
	}
	GousbOpenDeviceWithVIDPID = func(ctx *gousb.Context, vid gousb.ID, pid gousb.ID) (*gousb.Device, error) {
		return ctx.OpenDeviceWithVIDPID(vid, pid)
	}
	GousbDefaultInterface = func(d *gousb.Device) (*gousb.Interface, func(), error) {
		return d.DefaultInterface()
	}
	GousbInEndpoint = func(i *gousb.Interface, epNum int) (*gousb.InEndpoint, error) {
		return i.InEndpoint(epNum)
	}
	GousbOutEndpoint = func(i *gousb.Interface, epNum int) (*gousb.OutEndpoint, error) {
		return i.OutEndpoint(epNum)
	}
	GousbManufacturer = func(d *gousb.Device) (string, error) {
		return d.Manufacturer()
	}
	GousbProduct = func(d *gousb.Device) (string, error) {
		return d.Product()
	}
	GousbRead = func(ep *gousb.InEndpoint, buf []byte) (int, error) {
		return ep.Read(buf)
	}
	GousbWrite = func(ep *gousb.OutEndpoint, buf []byte) (int, error) {
		return ep.Write(buf)
	}
)

func (u *GousbAdapter) Open() (err error) {
	u.ctx = GousbNewContext()

	u.dev, err = GousbOpenDeviceWithVIDPID(u.ctx, IDVendor, IDProduct)
	if err != nil {
		return fmt.Errorf("could not open a device: %v", err)
	}
	if u.dev == nil {
		return errors.New("could not open a device, is it connected?")
	}

	u.intf, u.intfDone, err = GousbDefaultInterface(u.dev)
	if err != nil {
		return fmt.Errorf("could not get default interface: %v", err)
	}

	u.epIn, err = GousbInEndpoint(u.intf, EndpointNumIn)
	if err != nil {
		return fmt.Errorf("could not get IN endpoint: %v", err)
	}

	u.epOut, err = GousbOutEndpoint(u.intf, EndpointNumOut)
	if err != nil {
		return fmt.Errorf("could not get OUT endpoint: %v", err)
	}

	return nil
}

func (u *GousbAdapter) Close() (err error) {
	if u.intfDone != nil {
		u.intfDone()
	}
	if u.dev != nil {
		err = u.dev.Close()
	}
	if u.ctx != nil {
		err = u.ctx.Close()
	}

	return err
}

func (u *GousbAdapter) Read(buf []byte) (int, error) {
	return GousbRead(u.epIn, buf)
}

func (u *GousbAdapter) Write(buf []byte) (int, error) {
	return GousbWrite(u.epOut, buf)
}

func (u *GousbAdapter) Manufacturer() (string, error) {
	manufacturer, err := GousbManufacturer(u.dev)
	if err != nil {
		return "", fmt.Errorf("could not read Manufacturer: %v", err)
	}

	return manufacturer, nil
}

func (u *GousbAdapter) Product() (string, error) {
	product, err := GousbProduct(u.dev)
	if err != nil {
		return "", fmt.Errorf("could not read Product: %v", err)
	}

	return product, nil
}
