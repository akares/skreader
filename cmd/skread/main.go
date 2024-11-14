package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"encoding/json"

	"github.com/akares/skreader"
	"github.com/urfave/cli/v2"
)

const (
	name        = "skreader"
	version     = "0.3.0"
	description = "command line tool for SEKONIC spectrometers remote control"
)

type JSONResponse struct {
	Device       string                     `json:"Device"`
	Model        string                     `json:"Model"`
	Firmware     string                     `json:"Firmware"`
	Status       string                     `json:"Status"`
	Remote       string                     `json:"Remote"`
	Button       string                     `json:"Button"`
	Ring         string                     `json:"Ring"`
	Measurements []skreader.MeasurementJSON `json:"measurements"`
}

func skConnect() (*skreader.Device, error) {
	sk, err := skreader.NewDeviceWithAdapter(&skreader.GousbAdapter{})
	if err != nil {
		return nil, err
	}

	return sk, nil
}

func infoCmd(c *cli.Context) error {
	if c.Bool("fake-device") {
		fmt.Println("Fake device")

		return nil
	}

	sk, err := skConnect()
	if err != nil {
		return err
	}
	defer sk.Close()

	st, err := sk.State()
	if err != nil {
		return err
	}

	model, _ := sk.ModelName()
	fw, _ := sk.FirmwareVersion()

	fmt.Println("Device:", sk.String())
	fmt.Println("Model:", model)
	fmt.Println("Firmware:", fw)
	fmt.Println("Status:", st.Status)
	fmt.Println("Remote:", st.Remote)
	fmt.Println("Button:", st.Button, st.Ring)
	fmt.Println("Ring:", st.Ring)

	return nil
}

func jsonCmd(c *cli.Context) error {
	var meas *skreader.Measurement
	var err error

	var response JSONResponse
	if c.Bool("fake-device") {
		meas, err = skreader.NewMeasurementFromBytes(skreader.Testdata)
		if err != nil {
			return err
		}
		response = JSONResponse{
			Device:       "fake-device",
			Model:        "n/a",
			Firmware:     "n/a",
			Status:       "n/a",
			Remote:       "n/a",
			Button:       "n/a",
			Ring:         "n/a",
			Measurements: []skreader.MeasurementJSON{}, // populated later
		}
	} else {
		var sk *skreader.Device
		sk, err = skConnect()
		if err != nil {
			return err
		}
		defer sk.Close()

		meas, err = sk.Measure()
		if err != nil {
			return err
		}

		var st *skreader.DeviceState
		st, err = sk.State()
		if err != nil {
			return err
		}

		model, _ := sk.ModelName()
		fw, _ := sk.FirmwareVersion()

		response = JSONResponse{
			Device:       sk.String(),
			Model:        model,
			Firmware:     fmt.Sprintf("%v", fw),
			Status:       fmt.Sprintf("%v", st.Status),
			Remote:       fmt.Sprintf("%v", st.Remote),
			Button:       fmt.Sprintf("%v", st.Button),
			Ring:         fmt.Sprintf("%v", st.Ring),
			Measurements: []skreader.MeasurementJSON{}, // populated later
		}
	}

	measName := c.String("name")
	measNote := c.String("note")
	measTime := time.Now()

	measJSON := skreader.NewFromMeasurement(meas, measName, measNote, measTime)
	response.Measurements = append(response.Measurements, measJSON)

	file, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)

		return err
	}

	fmt.Println(string(file))

	return nil
}

//nolint:gocyclo,funlen
func measureCmd(c *cli.Context) error {
	var meas *skreader.Measurement
	var err error

	if c.Bool("fake-device") {
		meas, err = skreader.NewMeasurementFromBytes(skreader.Testdata)
		if err != nil {
			return err
		}
	} else {
		var sk *skreader.Device
		sk, err = skConnect()
		if err != nil {
			return err
		}
		defer sk.Close()

		meas, err = sk.Measure()
		if err != nil {
			return err
		}
	}

	verbose := c.Bool("verbose")

	showIlluminance := c.Bool("illuminance") || c.Bool("all") || c.Bool("simple")
	showColorTemperature := c.Bool("color-temperature") || c.Bool("all") || c.Bool("simple")
	showTristimulus := c.Bool("tristimulus") || c.Bool("all") || c.Bool("simple")
	showCIE1931 := c.Bool("cie1931") || c.Bool("all") || c.Bool("simple")
	showCIE1976 := c.Bool("cie1976") || c.Bool("all") || c.Bool("simple")
	showDWL := c.Bool("dwl") || c.Bool("all") || c.Bool("simple")

	showCRI := c.Bool("cri") || c.Bool("all")
	showSpectra1nm := c.Bool("spectra1nm") || c.Bool("all")
	showSpectra5nm := c.Bool("spectra5nm") || c.Bool("all")

	// Shown by default if no other flag is set
	showLDi := c.Bool("ldi") || c.Bool("all") || (!showIlluminance && !showColorTemperature && !showTristimulus && !showCIE1931 && !showCIE1976 && !showDWL && !showCRI && !showSpectra1nm && !showSpectra5nm)

	if showIlluminance {
		if verbose {
			fmt.Println("------------")
			fmt.Println("Illuminance:")
		}
		fmt.Println("LUX:", meas.Illuminance.Lux.Str)
		fmt.Println("Fc:", meas.Illuminance.FootCandle)
	}

	if showColorTemperature {
		if verbose {
			fmt.Println("------------")
			fmt.Println("ColorTemperature:")
		}
		fmt.Println("CCT:", meas.ColorTemperature.Tcp)
		fmt.Println("CCT DeltaUv:", meas.ColorTemperature.DeltaUv)
	}

	if showTristimulus {
		if verbose {
			fmt.Println("------------")
			fmt.Println("Tristimulus:")
		}
		fmt.Println("X:", meas.Tristimulus.X)
		fmt.Println("Y:", meas.Tristimulus.Y)
		fmt.Println("Z:", meas.Tristimulus.Z)
	}

	if showCIE1931 {
		if verbose {
			fmt.Println("------------")
			fmt.Println("CIE1931:")
		}
		fmt.Println("X:", meas.CIE1931.X)
		fmt.Println("Y:", meas.CIE1931.Y)
	}

	if showCIE1976 {
		if verbose {
			fmt.Println("CIE1976:")
			fmt.Println("------------")
		}
		fmt.Println("Ud:", meas.CIE1976.Ud)
		fmt.Println("Vd:", meas.CIE1976.Vd)
	}

	if showDWL {
		if verbose {
			fmt.Println("------------")
			fmt.Println("DominantWavelength:")
		}
		fmt.Println("DominantWavelength:", meas.DWL.Wavelength)
		fmt.Println("ExcitationPurity:", meas.DWL.ExcitationPurity)
	}

	if showCRI {
		if verbose {
			fmt.Println("------------")
			fmt.Println("CRI:")
		}
		fmt.Println("RA:", meas.ColorRenditionIndexes.Ra)
		for i := range meas.ColorRenditionIndexes.Ri {
			fmt.Printf("R%d: %s\n", i+1, meas.ColorRenditionIndexes.Ri[i])
		}
	}

	if showSpectra1nm {
		if verbose {
			fmt.Println("------------")
			fmt.Println("SpectralData 1nm:")
		}
		for i := range meas.SpectralData1nm {
			wavelength := 380 + i
			fmt.Printf("%d,%f\n", wavelength, meas.SpectralData1nm[i].Val)
		}
	}

	if showSpectra5nm {
		if verbose {
			fmt.Println("------------")
			fmt.Println("SpectralData 5nm:")
		}
		for i := range meas.SpectralData5nm {
			wavelength := 380 + (i * 5)
			fmt.Printf("%d,%f\n", wavelength, meas.SpectralData5nm[i].Val)
		}
	}

	if showLDi {
		if verbose {
			fmt.Println("------------")
		}
		fmt.Println("LUX:", meas.Illuminance.Lux.Str)
		fmt.Println("CCT:", meas.ColorTemperature.Tcp)
		fmt.Println("CCT DeltaUv:", meas.ColorTemperature.DeltaUv)
		fmt.Println("RA:", meas.ColorRenditionIndexes.Ra)
		fmt.Println("R9:", meas.ColorRenditionIndexes.Ri[8])
	}

	return nil
}

//nolint:exhaustruct,funlen
func main() {
	app := &cli.App{
		Name:                   name,
		Version:                version,
		Usage:                  description,
		Suggest:                true,
		EnableBashCompletion:   true,
		UseShortOptionHandling: true,
		Commands: []*cli.Command{
			{
				Name:   "info",
				Usage:  "Shows info about the connected device.",
				Action: infoCmd,
			},
			{
				Name:   "json",
				Usage:  "outputs all data as json",
				Action: jsonCmd,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Aliases:  []string{"na"},
						Usage:    "Measurement name",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "note",
						Aliases:  []string{"no"},
						Usage:    "Measurement note",
						Required: true,
					},
				},
			},
			{
				Name:   "measure",
				Usage:  "Runs one measurement and outputs the selected data.",
				Action: measureCmd,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "ldi",
						Aliases: []string{"l"},
						Usage:   "include the most interesting data for LDs",
					},
					&cli.BoolFlag{
						Name:    "all",
						Aliases: []string{"a"},
						Usage:   "include all measurement data",
					},
					&cli.BoolFlag{
						Name:    "simple",
						Aliases: []string{"s"},
						Usage:   "include all simple measurement data (excluding spectra and CRI)",
					},
					&cli.BoolFlag{
						Name:    "illuminance",
						Aliases: []string{"ill", "i"},
						Usage:   "include illuminance values in Lux and foot-candle units",
					},
					&cli.BoolFlag{
						Name:    "color-temperature",
						Aliases: []string{"cct", "c"},
						Usage:   "include color temperature values in Kelvin and delta-uv units",
					},
					&cli.BoolFlag{
						Name:    "tristimulus",
						Aliases: []string{"tri", "t"},
						Usage:   "include tristimulus values in XYZ color space",
					},
					&cli.BoolFlag{
						Name:    "cie1931",
						Aliases: []string{"xy", "x"},
						Usage:   "include CIE1931 (x, y) chromaticity coordinates",
					},
					&cli.BoolFlag{
						Name:    "cie1976",
						Aliases: []string{"uv", "u"},
						Usage:   "include CIE1976 (u', v') chromaticity coordinates",
					},
					&cli.BoolFlag{
						Name:    "dwl",
						Aliases: []string{"d"},
						Usage:   "include dominant wavelength value",
					},
					&cli.BoolFlag{
						Name:    "cri",
						Aliases: []string{"r"},
						Usage:   "include CRI (Ra, Ri) values",
					},
					&cli.BoolFlag{
						Name:    "spectra1nm",
						Aliases: []string{"1mm", "1"},
						Usage:   "include spectral data for 1nm wavelength",
					},
					&cli.BoolFlag{
						Name:    "spectra5nm",
						Aliases: []string{"5mm", "5"},
						Usage:   "include spectral data for 5nm wavelength",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "print more messages",
					},
				},
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "fake-device",
				Aliases: []string{"fake", "f"},
				Usage:   "use fake device for testing",
			},
		},
	}

	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print only the version",
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
