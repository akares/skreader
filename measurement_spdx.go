package skreader

import "time"

// SPDX tm2714 documentation
// https://colour.readthedocs.io/en/v0.3.10/_modules/colour/io/ies_tm2714.html

// Online calculators
// https://www.ies.org/standards/standards-toolbox/tm-30-spectral-calculator/
// https://luox.app

type SPDXHeader struct {
	Description string `xml:"Description"`
	Comments    string `xml:"Comments"`
	Date        string `xml:"Report_date"`
}

type SPDXWavelengths struct {
	SpectralData []SpectralData `xml:"SpectralData"`
}

type SpectralData struct {
	Wavelength float64 `xml:"wavelength,attr"`
	Value      float64 `xml:",chardata"`
}

func Header(measName, measNote string, measTime time.Time) SPDXHeader {

	formattedTime := measTime.Format("2006-01-02T15:04:05")

	header := SPDXHeader{
		Description: measName,
		Comments:    measNote,
		Date:        formattedTime,
	}
	return header
}

func NewSpdxMeasurement(meas *Measurement) SPDXWavelengths {

	var wavelengths SPDXWavelengths

	var spectra SpectralData

	for i, val := range &meas.SpectralData1nm {
		wl := float64(i) + 380
		value := float64(val.Val)

		spectra = SpectralData{
			Wavelength: wl,
			Value:      value,
		}

		wavelengths.SpectralData = append(wavelengths.SpectralData, spectra)
	}

	return wavelengths
}
