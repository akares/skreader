package skreader

type SkCommand string

const (
	SkCommandGetFirmwareVersion   SkCommand = "FV"  // Get Firmware Version
	SkCommandGetModelNumber       SkCommand = "MN"  // Get Model Name
	SkCommandGetStatus            SkCommand = "ST"  // Get Device Status
	SkCommandSetRemoteOn          SkCommand = "RT1" // Set Remote Mode On
	SkCommandSetRemoteOff         SkCommand = "RT0" // Set Remote Mode Off
	SkCommandSetFov               SkCommand = "AGw" // Set Field of View
	SkCommandSetMeasuringMode     SkCommand = "MMw" // Set Measuring Mode
	SkCommandSetExposureTime      SkCommand = "AMw" // Set Exposure Time
	SkCommandSetShutterSpeed      SkCommand = "SSw" // Set Shutter Speed
	SkCommandStartMeasuring       SkCommand = "RM0" // Start Measuring
	SkCommandGetMeasurementResult SkCommand = "NR"  // Get Measurement Result
)

type SkDeviceStatus int

const (
	SkDeviceStatusIdle SkDeviceStatus = iota
	SkDeviceStatusIdleOutMeas
	SkDeviceStatusBusyFlashStandby
	SkDeviceStatusBusyMeasuring
	SkDeviceStatusBusyInitializing
	SkDeviceStatusBusyDarkCalibration
	SkDeviceStatusErrorHw
)

type SkButtonStatus int

const (
	SkButtonStatusNone      SkButtonStatus = 0
	SkButtonStatusPower     SkButtonStatus = 1
	SkButtonStatusMeasuring SkButtonStatus = 2
	SkButtonStatusMemory    SkButtonStatus = 4
	SkButtonStatusMenu      SkButtonStatus = 8
	SkButtonStatusPanel     SkButtonStatus = 0x10
)

type SkRingStatus int

const (
	SkRingStatusUnpositioned SkRingStatus = iota
	SkRingStatusCal
	SkRingStatusLow
	SkRingStatusHigh
)

type SkRemoteStatus int

const (
	SkRemoteStatusOff SkRemoteStatus = iota
	SkRemoteStatusOn
)

type SkMeasuringMode int

const (
	SkMeasuringModeAmbient SkMeasuringMode = iota
	SkMeasuringModeCordlessFlash
	SkMeasuringModeCordFlash
)

type SkFieldOfView int

const (
	SkFieldOfView2Deg  SkFieldOfView = iota // 2°
	SkFieldOfView10Deg                      // 10°
)

type SkExposureTime int

const (
	SkExposureTimeAuto    SkExposureTime = iota
	SkExposureTime100Msec                // 0.1 s
	SkExposureTime1Sec                   // 1 s
)

type SkShutterSpeed string

const (
	SkShutterSpeed1Sec   SkShutterSpeed = "01" // 1 s
	SkShutterSpeed2Sec   SkShutterSpeed = "02" // 2 s
	SkShutterSpeed4Sec   SkShutterSpeed = "03" // 4 s
	SkShutterSpeed8Sec   SkShutterSpeed = "04" // 8 s
	SkShutterSpeed15Sec  SkShutterSpeed = "05" // 15 s
	SkShutterSpeed30Sec  SkShutterSpeed = "06" // 30 s
	SkShutterSpeed60Sec  SkShutterSpeed = "07" // 1/60 s
	SkShutterSpeed125Sec SkShutterSpeed = "08" // 1/125 s
	SkShutterSpeed250Sec SkShutterSpeed = "09" // 1/250 s
	SkShutterSpeed500Sec SkShutterSpeed = "10" // 1/500 s
)

type SkMeasuringMethod int

const (
	SkMeasuringMethodSingle SkMeasuringMethod = iota
	SkMeasuringMethodContinuous
)
