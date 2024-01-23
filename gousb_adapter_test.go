package skreader_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/gousb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/akares/skreader"
)

type GousbMock struct {
	mock.Mock
}

func (m *GousbMock) NewContext() *gousb.Context {
	args := m.Called()

	return args.Get(0).(*gousb.Context)
}

func (m *GousbMock) OpenDeviceWithVIDPID(ctx *gousb.Context, vid, pid gousb.ID) (*gousb.Device, error) {
	args := m.Called(ctx, vid, pid)
	d := args.Get(0)
	if d == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*gousb.Device), args.Error(1)
}

func (m *GousbMock) DefaultInterface(d *gousb.Device) (*gousb.Interface, func(), error) {
	args := m.Called(d)

	return args.Get(0).(*gousb.Interface), args.Get(1).(func()), args.Error(2)
}

func (m *GousbMock) InEndpoint(i *gousb.Interface, epNum int) (*gousb.InEndpoint, error) {
	args := m.Called(i, epNum)

	return args.Get(0).(*gousb.InEndpoint), args.Error(1)
}

func (m *GousbMock) OutEndpoint(i *gousb.Interface, epNum int) (*gousb.OutEndpoint, error) {
	args := m.Called(i, epNum)

	return args.Get(0).(*gousb.OutEndpoint), args.Error(1)
}

func (m *GousbMock) Manufacturer(d *gousb.Device) (string, error) {
	args := m.Called(d)

	return args.String(0), args.Error(1)
}

func (m *GousbMock) Product(d *gousb.Device) (string, error) {
	args := m.Called(d)

	return args.String(0), args.Error(1)
}

func (m *GousbMock) Read(ep *gousb.InEndpoint, buf []byte) (int, error) {
	args := m.Called(ep, buf)

	return args.Int(0), args.Error(1)
}

func (m *GousbMock) Write(ep *gousb.OutEndpoint, buf []byte) (int, error) {
	args := m.Called(ep, buf)

	return args.Int(0), args.Error(1)
}

func setupTest() (m *GousbMock, tearDown func()) {
	m = &GousbMock{} //nolint:exhaustruct

	m.On("NewContext").Return(&gousb.Context{}, nil)
	m.On("OpenDeviceWithVIDPID", mock.Anything, mock.Anything, mock.Anything).Return(&gousb.Device{}, nil) //nolint:exhaustruct
	m.On("DefaultInterface", mock.Anything).Return(&gousb.Interface{}, func() {}, nil)                     //nolint:exhaustruct
	m.On("InEndpoint", mock.Anything, mock.Anything).Return(&gousb.InEndpoint{}, nil)
	m.On("OutEndpoint", mock.Anything, mock.Anything).Return(&gousb.OutEndpoint{}, nil)
	m.On("Manufacturer", mock.Anything).Return("TheManufacturer", nil)
	m.On("Product", mock.Anything).Return("TheProduct", nil)

	skreader.GousbNewContext = m.NewContext
	skreader.GousbOpenDeviceWithVIDPID = m.OpenDeviceWithVIDPID
	skreader.GousbDefaultInterface = m.DefaultInterface
	skreader.GousbDefaultInterface = m.DefaultInterface
	skreader.GousbInEndpoint = m.InEndpoint
	skreader.GousbOutEndpoint = m.OutEndpoint
	skreader.GousbManufacturer = m.Manufacturer
	skreader.GousbProduct = m.Product
	skreader.GousbRead = m.Read
	skreader.GousbWrite = m.Write

	tearDown = func() {}

	return m, tearDown
}

//nolint:funlen
func TestGousbAdapterHappyPath(t *testing.T) {
	m, tearDown := setupTest()
	defer tearDown()

	sk, err := skreader.NewDeviceWithAdapter(&skreader.GousbAdapter{})
	assert.Nil(t, err, "NewDeviceWithAdapter() error")
	assert.NotNil(t, sk, "NewDeviceWithAdapter() is nil")

	m.On("Write", mock.Anything, mock.Anything).Return(10, nil)

	//
	// Test ModelName()
	//

	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(15, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte("MN@@@C-800\x00\x00\x00\x00\x00")) //nolint:gocritic
	})
	mn, err := sk.ModelName()
	assert.Nil(t, err, "ModelName() error")
	assert.Equal(t, "C-800", mn, "ModelName() invalid")

	//
	// Test SupportsMeasurementConfiguration()
	//

	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(15, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte("MN@@@C-12345\x00\x00\x00\x00\x00")) //nolint:gocritic
	})
	assert.False(t, sk.SupportsMeasurementConfiguration(), "SupportsMeasurementConfiguration() invalid")

	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(15, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte("MN@@@C-800\x00\x00\x00\x00\x00")) //nolint:gocritic
	})
	assert.True(t, sk.SupportsMeasurementConfiguration(), "SupportsMeasurementConfiguration() invalid")

	//
	// Test SupportsExtendedMeasurementConfiguration()
	//

	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(15, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte("MN@@@C-800\x00\x00\x00\x00\x00")) //nolint:gocritic
	})
	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(15, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte("FV@@@20,C36E,27,7881,11,B216,14,50CC,17,74EC")) //nolint:gocritic
	})
	assert.False(t, sk.SupportsExtendedMeasurementConfiguration(), "SupportsExtendedMeasurementConfiguration() invalid")

	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(15, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte("MN@@@C-7000\x00\x00\x00\x00\x00")) //nolint:gocritic
	})
	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(15, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte("FV@@@20,C36E,27,7881,11,B216,14,50CC,17,74EC")) //nolint:gocritic
	})
	assert.True(t, sk.SupportsExtendedMeasurementConfiguration(), "SupportsExtendedMeasurementConfiguration() invalid")

	//
	// Test FirmwareVersion()
	//

	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(15, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte("FV@@@20,C36E,27,7881,11,B216,14,50CC,17,74EC")) //nolint:gocritic
	})
	fw, err := sk.FirmwareVersion()
	assert.Nil(t, err, "FirmwareVersion() error")
	assert.Equal(t, 27, fw, "FirmwareVersion() invalid")

	//
	// Test SetRemoteOn()
	//

	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte("RT")) //nolint:gocritic
	})
	err = sk.SetRemoteOn()
	assert.Nil(t, err, "SetRemoteOn() error")

	//
	// Test SetRemoteOff()
	//

	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte("RT")) //nolint:gocritic
	})
	err = sk.SetRemoteOff()
	assert.Nil(t, err, "SetRemoteOff() error")

	//
	// Test State()
	//

	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(5, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte("ST@@@")) //nolint:gocritic
	})
	_, err = sk.State()
	assert.Nil(t, err, "State() error")

	//
	// Test WaitReady()
	//

	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(5, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte("ST@@@")) //nolint:gocritic
	})
	err = sk.WaitReady(time.Duration(10)*time.Millisecond, time.Duration(1)*time.Millisecond)
	assert.Nil(t, err, "WaitReady() error")

	//
	// Test WaitReady() Ring Pos err
	//

	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(5, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{83, 84, 255, 255, 0})
	})
	err = sk.WaitReady(time.Duration(10)*time.Millisecond, time.Duration(1)*time.Millisecond)
	assert.NotNil(t, err, "WaitReady() error")

	//
	// Test WaitReady() Status err 1
	//

	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(5, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{83, 84, 255, 0, 0})
	})
	err = sk.WaitReady(time.Duration(10)*time.Millisecond, time.Duration(1)*time.Millisecond)
	assert.NotNil(t, err, "WaitReady() error")

	//
	// Test WaitReady() Status err 2
	//

	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(5, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{83, 84, 8, 255, 255})
	})
	err = sk.WaitReady(time.Duration(10)*time.Millisecond, time.Duration(1)*time.Millisecond)
	assert.NotNil(t, err, "WaitReady() error")

	//
	// SetMeasurementConfiguration
	//

	// ModelNumber OK
	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(15, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte("MN@@@C-7000\x00\x00\x00\x00\x00")) //nolint:gocritic
	})

	// MeasuringMode OK
	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte("MM")) //nolint:gocritic
	})
	// ShutterSpeed OK
	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte("SS")) //nolint:gocritic
	})

	// ModelNumber OK
	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(15, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte("MN@@@C-7000\x00\x00\x00\x00\x00")) //nolint:gocritic
	})
	// FirmwareVersion OK
	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(15, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte("FV@@@20,C36E,27,7881,11,B216,14,50CC,17,74EC")) //nolint:gocritic
	})

	// SetFov OK
	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte("AG")) //nolint:gocritic
	})
	// SetExposureTime OK
	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte("AM")) //nolint:gocritic
	})

	err = sk.SetMeasurementConfiguration()
	assert.Nil(t, err, "SetMeasurementConfiguration() error")

	//
	// Test StartMeasuring()
	//

	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte("RM")) //nolint:gocritic
	})
	err = sk.StartMeasuring()
	assert.Nil(t, err, "StartMeasuring() error")

	//
	// Test MeasurementResult()
	//

	// Invalid data size
	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(1234, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, skreader.Testdata)
	})
	_, err = sk.MeasurementResult()
	assert.NotNil(t, err, "MeasurementResult() error")

	// Correct data
	m.On("Read", mock.Anything, mock.Anything).Return(2, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	m.On("Read", mock.Anything, mock.Anything).Return(2380, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, skreader.Testdata)
	})
	_, err = sk.MeasurementResult()
	assert.Nil(t, err, "MeasurementResult() error")

	err = sk.Close()
	assert.Nil(t, err, "Close() error")
}

func TestGousbAdapterReadErrors(t *testing.T) { //nolint:funlen
	m, tearDown := setupTest()
	defer tearDown()

	sk, err := skreader.NewDeviceWithAdapter(&skreader.GousbAdapter{})
	assert.Nil(t, err, "NewDeviceWithAdapter() error")
	assert.NotNil(t, sk, "NewDeviceWithAdapter() is nil")

	m.On("Write", mock.Anything, mock.Anything).Return(2, nil)

	m.On("Read", mock.Anything, mock.Anything).Return(2, errors.New("read error")).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	mn, err := sk.ModelName()
	assert.NotNil(t, err, "ModelName() error")
	assert.Equal(t, "", mn, "ModelName() invalid")

	m.On("Read", mock.Anything, mock.Anything).Return(2, errors.New("read error")).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{6, 48})
	})
	fw, err := sk.FirmwareVersion()
	assert.NotNil(t, err, "FirmwareVersion() error")
	assert.Equal(t, 0, fw, "FirmwareVersion() invalid")

	m.On("Manufacturer", mock.Anything).Unset()
	m.On("Manufacturer", mock.Anything).Return("", errors.New("read error")).Once()

	sk, err = skreader.NewDeviceWithAdapter(&skreader.GousbAdapter{})
	assert.NotNil(t, err, "NewDeviceWithAdapter() error")
	assert.Nil(t, sk, "NewDeviceWithAdapter() is not nil")

	m.On("Manufacturer", mock.Anything).Return("TheManufacturer", nil)

	m.On("Product", mock.Anything).Unset()
	m.On("Product", mock.Anything).Return("", errors.New("read error")).Once()

	sk, err = skreader.NewDeviceWithAdapter(&skreader.GousbAdapter{})
	assert.NotNil(t, err, "NewDeviceWithAdapter() error")
	assert.Nil(t, sk, "NewDeviceWithAdapter() is not nil")

	m.On("OpenDeviceWithVIDPID", mock.Anything, mock.Anything, mock.Anything).Unset()
	m.On("OpenDeviceWithVIDPID", mock.Anything, mock.Anything, mock.Anything).Return(&gousb.Device{}, errors.New("init error")).Once() //nolint:exhaustruct

	sk, err = skreader.NewDeviceWithAdapter(&skreader.GousbAdapter{})
	assert.NotNil(t, err, "NewDeviceWithAdapter() error")
	assert.Nil(t, sk, "NewDeviceWithAdapter() is not nil")

	m.On("OpenDeviceWithVIDPID", mock.Anything, mock.Anything, mock.Anything).Return(&gousb.Device{}, nil) //nolint:exhaustruct

	m.On("OpenDeviceWithVIDPID", mock.Anything, mock.Anything, mock.Anything).Unset()
	m.On("OpenDeviceWithVIDPID", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Once()

	sk, err = skreader.NewDeviceWithAdapter(&skreader.GousbAdapter{})
	assert.NotNil(t, err, "NewDeviceWithAdapter() error")
	assert.Nil(t, sk, "NewDeviceWithAdapter() is not nil")

	m.On("OpenDeviceWithVIDPID", mock.Anything, mock.Anything, mock.Anything).Return(&gousb.Device{}, nil) //nolint:exhaustruct

	m.On("DefaultInterface", mock.Anything).Unset()
	m.On("DefaultInterface", mock.Anything).Return(&gousb.Interface{}, func() {}, errors.New("init error")).Once() //nolint:exhaustruct

	sk, err = skreader.NewDeviceWithAdapter(&skreader.GousbAdapter{})
	assert.NotNil(t, err, "NewDeviceWithAdapter() error")
	assert.Nil(t, sk, "NewDeviceWithAdapter() is not nil")

	m.On("DefaultInterface", mock.Anything).Return(&gousb.Interface{}, func() {}, nil) //nolint:exhaustruct

	m.On("InEndpoint", mock.Anything, mock.Anything).Unset()
	m.On("InEndpoint", mock.Anything, mock.Anything).Return(&gousb.InEndpoint{}, errors.New("init error")).Once()

	sk, err = skreader.NewDeviceWithAdapter(&skreader.GousbAdapter{})
	assert.NotNil(t, err, "NewDeviceWithAdapter() error")
	assert.Nil(t, sk, "NewDeviceWithAdapter() is not nil")

	m.On("InEndpoint", mock.Anything, mock.Anything).Return(&gousb.InEndpoint{}, nil)

	m.On("OutEndpoint", mock.Anything, mock.Anything).Unset()
	m.On("OutEndpoint", mock.Anything, mock.Anything).Return(&gousb.OutEndpoint{}, errors.New("init error")).Once()

	sk, err = skreader.NewDeviceWithAdapter(&skreader.GousbAdapter{})
	assert.NotNil(t, err, "NewDeviceWithAdapter() error")
	assert.Nil(t, sk, "NewDeviceWithAdapter() is not nil")

	m.On("OutEndpoint", mock.Anything, mock.Anything).Return(&gousb.OutEndpoint{}, nil)
}

func TestGousbAdapterReadNoData(t *testing.T) {
	m, tearDown := setupTest()
	defer tearDown()

	sk, err := skreader.NewDeviceWithAdapter(&skreader.GousbAdapter{})
	assert.Nil(t, err, "NewDeviceWithAdapter() error")
	assert.NotNil(t, sk, "NewDeviceWithAdapter() is nil")

	m.On("Write", mock.Anything, mock.Anything).Return(2, nil)

	m.On("Read", mock.Anything, mock.Anything).Return(0, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{})
	})
	mn, err := sk.ModelName()
	assert.NotNil(t, err, "ModelName() error")
	assert.Equal(t, "", mn, "ModelName() invalid")

	m.On("Read", mock.Anything, mock.Anything).Return(0, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{})
	})
	fw, err := sk.FirmwareVersion()
	assert.NotNil(t, err, "FirmwareVersion() error")
	assert.Equal(t, 0, fw, "FirmwareVersion() invalid")
}

func TestGousbAdapterReadDataOverflow(t *testing.T) {
	m, tearDown := setupTest()
	defer tearDown()

	sk, err := skreader.NewDeviceWithAdapter(&skreader.GousbAdapter{})
	assert.Nil(t, err, "NewDeviceWithAdapter() error")
	assert.NotNil(t, sk, "NewDeviceWithAdapter() is nil")

	m.On("Write", mock.Anything, mock.Anything).Return(2, nil)

	m.On("Read", mock.Anything, mock.Anything).Return(100, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{})
	})
	mn, err := sk.ModelName()
	assert.NotNil(t, err, "ModelName() error")
	assert.Equal(t, "", mn, "ModelName() invalid")

	m.On("Read", mock.Anything, mock.Anything).Return(100, nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(1).([]byte)
		copy(buf, []byte{})
	})
	fw, err := sk.FirmwareVersion()
	assert.NotNil(t, err, "FirmwareVersion() error")
	assert.Equal(t, 0, fw, "FirmwareVersion() invalid")
}

func TestGousbAdapterWriteErrors(t *testing.T) {
	m, tearDown := setupTest()
	defer tearDown()

	sk, err := skreader.NewDeviceWithAdapter(&skreader.GousbAdapter{})
	assert.Nil(t, err, "NewDeviceWithAdapter() error")
	assert.NotNil(t, sk, "NewDeviceWithAdapter() is nil")

	m.On("Write", mock.Anything, mock.Anything).Return(0, errors.New("write error")).Once()
	mn, err := sk.ModelName()
	assert.NotNil(t, err, "ModelName() error")
	assert.Equal(t, "", mn, "ModelName() invalid")

	m.On("Write", mock.Anything, mock.Anything).Return(2, errors.New("write error")).Once()
	fw, err := sk.FirmwareVersion()
	assert.NotNil(t, err, "FirmwareVersion() error")
	assert.Equal(t, 0, fw, "FirmwareVersion() invalid")
}

func TestGousbAdapterWriteSizeErrors(t *testing.T) {
	m, tearDown := setupTest()
	defer tearDown()

	sk, err := skreader.NewDeviceWithAdapter(&skreader.GousbAdapter{})
	assert.Nil(t, err, "NewDeviceWithAdapter() error")
	assert.NotNil(t, sk, "NewDeviceWithAdapter() is nil")

	m.On("Write", mock.Anything, mock.Anything).Return(0, nil)
	mn, err := sk.ModelName()
	assert.NotNil(t, err, "ModelName() error")
	assert.Equal(t, "", mn, "ModelName() invalid")

	m.On("Write", mock.Anything, mock.Anything).Return(0, nil).Once()
	fw, err := sk.FirmwareVersion()
	assert.NotNil(t, err, "FirmwareVersion() error")
	assert.Equal(t, 0, fw, "FirmwareVersion() invalid")
}
