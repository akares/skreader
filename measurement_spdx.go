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

type SPDXSpectralDataPoint struct {
	Wavelength float64 `xml:"wavelength,attr"`
	Value      float64 `xml:",chardata"`
}

type SPDXSpectralDistribution struct {
	SpectralData []SPDXSpectralDataPoint `xml:"SpectralData"`
}

func NewSPDXHeader(measName, measNote string, measTime time.Time) SPDXHeader {
	formattedTime := measTime.Format("2006-01-02T15:04:05")

	header := SPDXHeader{
		Description: measName,
		Comments:    measNote,
		Date:        formattedTime,
	}

	return header
}

func NewSPDXSpectralDistribution(meas *Measurement) SPDXSpectralDistribution {
	spectralData := make([]SPDXSpectralDataPoint, len(meas.SpectralData1nm))
	for i, val := range &meas.SpectralData1nm {
		spectralData[i] = SPDXSpectralDataPoint{
			Wavelength: float64(i) + 380,
			Value:      val.Val,
		}
	}

	return SPDXSpectralDistribution{
		SpectralData: spectralData,
	}
}
