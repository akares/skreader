package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/akares/skreader"
)

//nolint:funlen,gocyclo
func main() {
	// Available commands
	help := flag.Bool("help", false, "Shows usage information")
	run := flag.Bool("run", false, "Runs a normal measurement and outputs the selected data")

	// Data to show
	showInfo := flag.Bool("info", false, "Shows info about the connected device")
	showAll := flag.Bool("all", false, "Shows all data from the connected device")
	showIlluminance := flag.Bool("Illuminance", false, "Shows illuminance data")
	showColorTemperature := flag.Bool("ColorTemperature", false, "Shows ColorTemperature data")
	showTristimulus := flag.Bool("Tristimulus", false, "Shows Tristimulus data")
	showCIE1931 := flag.Bool("CIE1931", false, "Shows CIE1931 data")
	showCIE1976 := flag.Bool("CIE1976", false, "Shows CIE1976 data")
	showDWL := flag.Bool("DWL", false, "Shows DWL data")
	show := flag.Bool("CRI", false, "Shows CRI data")
	showSpectra1nm := flag.Bool("Spectra1nm", false, "Shows Spectra1nm data")
	showSpectra5nm := flag.Bool("Spectra5nm", false, "Shows Spectra5nm data")

	// Shown by default if no other flag is set
	showLDi := flag.Bool("LDi", false, "Shows the most interesting data for LDs")

	flag.Parse()

	if *help || len(os.Args) == 1 {
		fmt.Println("Usage: skreader [options]")
		fmt.Println("Example: skreader -run -all")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()

		os.Exit(0)
	}

	if *run && len(os.Args) == 2 {
		*showLDi = true
	}

	if *showAll {
		*showInfo = true
		*showIlluminance = true
		*showColorTemperature = true
		*showTristimulus = true
		*showCIE1931 = true
		*showCIE1976 = true
		*showDWL = true
		*show = true
		*showSpectra1nm = true
		*showSpectra5nm = true
	}

	if *run {
		// Connect to SEKONIC device.
		sk, err := skreader.NewDeviceWithAdapter(&skreader.GousbAdapter{})
		if err != nil {
			panic(err)
		}
		defer sk.Close()

		// Get some basic info of the device.
		model, _ := sk.ModelName()
		fw, _ := sk.FirmwareVersion()

		// Get the current operational mode, knobs and buttons states of the device.
		st, err := sk.State()
		if err != nil {
			panic(err)
		}

		// Print the device info
		if *showInfo {
			fmt.Println(strconv.Quote(sk.String()))
			fmt.Println("Model:", strconv.Quote(model))
			fmt.Println("Firmware:", fw)
			fmt.Printf("State: %+v\n", st)
		}

		// Run one measurement.
		meas, err := sk.Measure()
		if err != nil {
			panic(err)
		}

		if *showIlluminance {
			fmt.Printf("------------\n")
			fmt.Printf("Illuminance:\n")
			fmt.Printf("LUX: %s\n", meas.Illuminance.Lux.Str)
			fmt.Printf("Fc: %s\n", meas.Illuminance.FootCandle)
		}

		if *showColorTemperature {
			fmt.Printf("------------\n")
			fmt.Printf("ColorTemperature:\n")
			fmt.Printf("CCT: %s\n", meas.ColorTemperature.Tcp)
			fmt.Printf("CCT DeltaUv: %s\n", meas.ColorTemperature.DeltaUv)
		}

		if *showTristimulus {
			fmt.Printf("------------\n")
			fmt.Printf("Tristimulus:\n")
			fmt.Printf("X: %s\n", meas.Tristimulus.X)
			fmt.Printf("Y: %s\n", meas.Tristimulus.Y)
			fmt.Printf("Z: %s\n", meas.Tristimulus.Z)
		}

		if *showCIE1931 {
			fmt.Printf("------------\n")
			fmt.Printf("CIE1931:\n")
			fmt.Printf("X: %s\n", meas.CIE1931.X)
			fmt.Printf("Y: %s\n", meas.CIE1931.Y)
		}

		if *showCIE1976 {
			fmt.Printf("------------\n")
			fmt.Printf("CIE1976:\n")
			fmt.Printf("Ud: %s\n", meas.CIE1976.Ud)
			fmt.Printf("Vd: %s\n", meas.CIE1976.Vd)
		}

		if *showDWL {
			fmt.Printf("------------\n")
			fmt.Printf("DominantWavelength:\n")
			fmt.Printf("Wavelength: %s\n", meas.DWL.Wavelength)
			fmt.Printf("ExcitationPurity: %s\n", meas.DWL.ExcitationPurity)
		}

		if *show {
			fmt.Printf("------------\n")
			fmt.Printf("CRI:\n")
			fmt.Printf("RA: %s\n", meas.ColorRenditionIndexes.Ra)
			for i := range meas.ColorRenditionIndexes.Ri {
				fmt.Printf("R%d: %s\n", i+1, meas.ColorRenditionIndexes.Ri[i])
			}
		}

		if *showSpectra1nm {
			fmt.Printf("------------\n")
			fmt.Printf("SpectralData 1nm:\n")
			for i := range meas.SpectralData1nm {
				// TODO: Missing one datapoint?
				wavelength := 380 + i
				fmt.Printf("%d,%f\n", wavelength, meas.SpectralData1nm[i].Val)
			}
		}

		if *showSpectra5nm {
			fmt.Printf("------------\n")
			fmt.Printf("SpectralData 5nm:\n")
			for i := range meas.SpectralData5nm {
				// TODO: Missing one datapoint?
				wavelength := 380 + (i * 5)
				fmt.Printf("%d,%f\n", wavelength, meas.SpectralData5nm[i].Val)
			}
		}

		if *showLDi {
			fmt.Printf("LUX: %s\n", meas.Illuminance.Lux.Str)
			fmt.Printf("CCT: %s\n", meas.ColorTemperature.Tcp)
			fmt.Printf("CCT DeltaUv: %s\n", meas.ColorTemperature.DeltaUv)
			fmt.Printf("RA: %s\n", meas.ColorRenditionIndexes.Ra)
			fmt.Printf("R9: %s\n", meas.ColorRenditionIndexes.Ri[8])
		}
	}
}
