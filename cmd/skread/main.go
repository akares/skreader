package main

import (
	"fmt"
	"strconv"

	"github.com/akares/skreader"
)

func main() {
	// Connect to SEKONIC device.
	sk, err := skreader.NewDeviceWithAdapter(&skreader.GousbAdapter{})
	if err != nil {
		panic(err)
	}
	defer sk.Close()

	// Get some basic info of the device.
	model, _ := sk.ModelName()
	fw, _ := sk.FirmwareVersion()

	fmt.Println(strconv.Quote(sk.String()))
	fmt.Println("MN:", strconv.Quote(model))
	fmt.Println("FW:", fw)

	// Get the current operational mode, knobs and buttons states of the device.
	st, err := sk.State()
	if err != nil {
		panic(err)
	}
	fmt.Printf("ST: %+v\n", st)

	// Run one measurement.
	meas, err := sk.Measure()
	if err != nil {
		panic(err)
	}

	// Print the measurement result in various vays.
	fmt.Printf("Meas: %s\n", meas.Repr())
	fmt.Printf("Meas: %s\n", meas.String())
	fmt.Printf(
		"Lux=%s x=%s y=%s CCT=%s\n",
		meas.Illuminance.Lux.Str,
		meas.CIE1931.X.Str,
		meas.CIE1931.Y.Str,
		meas.ColorTemperature.Tcp.Str,
	)
}
