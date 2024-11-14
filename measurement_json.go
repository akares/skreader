package skreader

import "time"

// TODO: Add tests

// MeasurementJSON is a struct that represents a JSON object that can be used to
// serialize a Measurement object.
type MeasurementJSON struct {
	Name             string               `json:"Name"`
	Note             string               `json:"Note"`
	Timestamp        int64                `json:"Timestamp"`
	Illuminance      IlluminanceJSON      `json:"Illuminance"`
	ColorTemperature ColorTemperatureJSON `json:"ColorTemperature"`
	Tristimulus      TristimulusJSON      `json:"Tristimulus"`
	CIE1931          CIE1931JSON          `json:"CIE1931"`
	CIE1976          CIE1976JSON          `json:"CIE1976"`
	DWL              DWLJSON              `json:"DWL"`
	CRI              CRIJSON              `json:"CRI"`
	SpectralData     []SpectralDataJSON   `json:"SpectralData"`
}

type IlluminanceJSON struct {
	LUX float64 `json:"LUX"`
	Fc  float64 `json:"Fc"`
}

type ColorTemperatureJSON struct {
	CCT        float64 `json:"CCT"`
	CCTDeltaUV float64 `json:"CCT DeltaUV"`
}

type TristimulusJSON struct {
	X float64 `json:"X"`
	Y float64 `json:"Y"`
	Z float64 `json:"Z"`
}

type CIE1931JSON struct {
	X float64 `json:"X"`
	Y float64 `json:"Y"`
}

type CIE1976JSON struct {
	Ud float64 `json:"Ud"`
	Vd float64 `json:"Vd"`
}

type DWLJSON struct {
	Wavelength       float64 `json:"Wavelength"`
	ExcitationPurity float64 `json:"ExcitationPurity"`
}

type CRIJSON struct {
	RA float64   `json:"RA"`
	Ri []float64 `json:"Ri"`
}

type SpectralDataJSON struct {
	Range  SpectralDataRangeJSON `json:"Range"`
	Values []float64             `json:"Values"`
}

type SpectralDataRangeJSON struct {
	Type    string `json:"Type"`
	StartNm int    `json:"StartNm"`
	EndNm   int    `json:"EndNm"`
	StepNm  int    `json:"StepNm"`
}

func NewFromMeasurement(meas *Measurement, measName, measNote string, measTime time.Time) MeasurementJSON {
	res := MeasurementJSON{
		Name:      measName,
		Note:      measNote,
		Timestamp: measTime.Unix(),
		Illuminance: IlluminanceJSON{
			LUX: meas.Illuminance.Lux.Val,
			Fc:  meas.Illuminance.FootCandle.Val,
		},
		ColorTemperature: ColorTemperatureJSON{
			CCT:        meas.ColorTemperature.Tcp.Val,
			CCTDeltaUV: meas.ColorTemperature.DeltaUv.Val,
		},
		Tristimulus: TristimulusJSON{
			X: meas.Tristimulus.X.Val,
			Y: meas.Tristimulus.Y.Val,
			Z: meas.Tristimulus.Z.Val,
		},
		CIE1931: CIE1931JSON{
			X: meas.CIE1931.X.Val,
			Y: meas.CIE1931.Y.Val,
		},
		CIE1976: CIE1976JSON{
			Ud: meas.CIE1976.Ud.Val,
			Vd: meas.CIE1976.Vd.Val,
		},
		DWL: DWLJSON{
			Wavelength:       meas.DWL.Wavelength.Val,
			ExcitationPurity: meas.DWL.ExcitationPurity.Val,
		},
		CRI: CRIJSON{
			RA: meas.ColorRenditionIndexes.Ra.Val,
			Ri: make([]float64, len(meas.ColorRenditionIndexes.Ri)), // populated later
		},
		SpectralData: []SpectralDataJSON{}, // populated later
	}

	// Populate Ri
	for i, val := range meas.ColorRenditionIndexes.Ri {
		res.CRI.Ri[i] = val.Val
	}

	// Populate 1nm
	spectralData1nm := SpectralDataJSON{
		Range: SpectralDataRangeJSON{
			Type:    "1nm",
			StartNm: 380,
			EndNm:   780,
			StepNm:  1,
		},
		Values: make([]float64, len(meas.SpectralData1nm)),
	}
	for i, val := range &meas.SpectralData1nm {
		spectralData1nm.Values[i] = val.Val
	}

	// Populate 5nm
	spectralData5nm := SpectralDataJSON{
		Range: SpectralDataRangeJSON{
			Type:    "5nm",
			StartNm: 380,
			EndNm:   780,
			StepNm:  5,
		},
		Values: make([]float64, len(meas.SpectralData5nm)),
	}
	for i, val := range &meas.SpectralData5nm {
		spectralData5nm.Values[i] = val.Val
	}

	res.SpectralData = []SpectralDataJSON{spectralData1nm, spectralData5nm}

	return res
}
