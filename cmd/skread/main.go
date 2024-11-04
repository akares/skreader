package main

import (
	"fmt"
	"strconv"
	"flag"
	"os"
	"github.com/akares/skreader"
)

func main() {

	// Help
	help := flag.Bool("help", false, "Display help information")

	// Mode
	run := flag.Bool("run", false, "Runs a normal measurement and outputs the selected data")

	// Device info
	info := flag.Bool("info", false, "Shows info about the connected device")

	// Data to show and later save
	all := flag.Bool("all", false, "Shows all data from the connected device")
	illuminance := flag.Bool("illuminance", false, "Shows illuminance data")
	ColorTemperature := flag.Bool("ColorTemperature", false, "Shows ColorTemperature data")
	Tristimulus := flag.Bool("Tristimulus", false, "Shows Tristimulus data")
	CIE1931 := flag.Bool("CIE1931", false, "Shows CIE1931 data")
	CIE1976 := flag.Bool("CIE1976", false, "Shows CIE1976 data")
	DWL := flag.Bool("DWL", false, "Shows DWL data")
	CRI := flag.Bool("CRI", false, "Shows CRI data")
	Spectra1nm := flag.Bool("Spectra1nm", false, "Shows Spectra1nm data")
	Spectra5nm := flag.Bool("Spectra5nm", false, "Shows Spectra5nm data")

	// Special
	LDi := flag.Bool("LDi", false, "Shows the most interesting data for LDs")



	// Parse all flags
	flag.Parse()



	// Help
	if *help {
		fmt.Println("Usage: go run main.go [options]")
		fmt.Println("Example: go run main.go --run --all")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Set the rest to true if all is set.
	if *all {
		*info = true
		*illuminance = true
		*ColorTemperature = true
		*Tristimulus = true
		*CIE1931 = true
		*CIE1976 = true
		*DWL = true
		*CRI = true
		*Spectra1nm = true
		*Spectra5nm = true
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


			// print the device info
			if *info {
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




			if *illuminance {
				
				fmt.Printf("------------\n")
				fmt.Printf("Illuminance:\n")
				fmt.Printf("LUX: %s\n", meas.Illuminance.Lux.Str)
				fmt.Printf("Fc: %s\n", meas.Illuminance.FootCandle)

			}


			
			if *ColorTemperature {

				fmt.Printf("------------\n")
				fmt.Printf("ColorTemperature:\n")
				fmt.Printf("CCT: %s\n", meas.ColorTemperature.Tcp)
				fmt.Printf("CCT DeltaUv: %s\n", meas.ColorTemperature.DeltaUv)

			}


			
			if *Tristimulus {
				fmt.Printf("------------\n")
				fmt.Printf("Tristimulus:\n")
				fmt.Printf("X: %s\n", meas.Tristimulus.X)
				fmt.Printf("Y: %s\n", meas.Tristimulus.Y)
				fmt.Printf("Z: %s\n", meas.Tristimulus.Z)
			}




			if *CIE1931 {
				fmt.Printf("------------\n")
				fmt.Printf("CIE1931:\n")
				fmt.Printf("X: %s\n", meas.CIE1931.X)
				fmt.Printf("Y: %s\n", meas.CIE1931.Y)
			}



			if *CIE1976 {
				fmt.Printf("------------\n")
				fmt.Printf("CIE1976:\n")
				fmt.Printf("Ud: %s\n", meas.CIE1976.Ud)
				fmt.Printf("Vd: %s\n", meas.CIE1976.Vd)
			}


			if *DWL{

				fmt.Printf("------------\n")
				fmt.Printf("DominantWavelength:\n")
				fmt.Printf("Wavelength: %s\n", meas.DWL.Wavelength)
				fmt.Printf("ExcitationPurity: %s\n", meas.DWL.ExcitationPurity)

			}

			

			if *CRI {
				fmt.Printf("------------\n")
				fmt.Printf("CRI:\n")
				fmt.Printf("RA: %s\n", meas.ColorRenditionIndexes.Ra)

				for i := range meas.ColorRenditionIndexes.Ri {
					fmt.Printf("R%d: %s\n", i+1, meas.ColorRenditionIndexes.Ri[i])
				}
			}

			


			if *Spectra1nm {
				fmt.Printf("------------\n")
				fmt.Printf("SpectralData 1nm:\n")
				for i := range meas.SpectralData1nm {
					//Missing one datapoint?
					var wavelength int = 380+i
			
					fmt.Printf("%d,%f\n", wavelength, meas.SpectralData1nm[i].Val)
					
				}
			}



			if *Spectra5nm {
				fmt.Printf("------------\n")
				fmt.Printf("SpectralData 5nm:\n")
				for i := range meas.SpectralData5nm {
					//Missing one datapoint?
					var wavelength int = 380+(i*5)
			
					fmt.Printf("%d,%f\n", wavelength, meas.SpectralData5nm[i].Val)
					
				}
			}
		

			if *LDi {

				fmt.Printf("LUX: %s\n", meas.Illuminance.Lux.Str)
				fmt.Printf("CCT: %s\n", meas.ColorTemperature.Tcp)
				fmt.Printf("CCT DeltaUv: %s\n", meas.ColorTemperature.DeltaUv)
				fmt.Printf("RA: %s\n", meas.ColorRenditionIndexes.Ra)
				fmt.Printf("R9: %s\n", meas.ColorRenditionIndexes.Ri[8])

			}



		}



}
