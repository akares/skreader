package skreader

import "time"

// TODO: Add tests

// MeasurementJSON is a struct that represents a JSON object that can be used to
// serialize a Measurement object.
type MeasurementJSON struct {
	Name             string                `json:"Name"`
	Note             string                `json:"Note"`
	Timestamp        int64                 `json:"Timestamp"`
	Illuminance      IlluminanceJSON       `json:"Illuminance"`
	ColorTemperature ColorTemperatureJSON  `json:"ColorTemperature"`
	Tristimulus      TristimulusJSON       `json:"Tristimulus"`
	CIE1931          CIE1931JSON           `json:"CIE1931"`
	CIE1976          CIE1976JSON           `json:"CIE1976"`
	DWL              DWLJSON               `json:"DWL"`
	CRI              CRIJSON               `json:"CRI"`
	WaveLengths      []WaveLengthGroupJSON `json:"WaveLengths"`
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
	RA float64     `json:"RA"`
	Ri []CRIRiJSON `json:"Ri"`
}

type CRIRiJSON struct {
	Ri    int     `json:"Ri"`
	Value float64 `json:"value"`
}

type WaveLengthGroupJSON struct {
	Type  string     `json:"type"`
	Waves []WaveJSON `json:"waves"`
}

type WaveJSON struct {
	Nm    int     `json:"Nm"`
	Value float64 `json:"value"`
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
			Ri: []CRIRiJSON{},
		},
		WaveLengths: []WaveLengthGroupJSON{
			{Type: "1nm", Waves: []WaveJSON{}},
			{Type: "5nm", Waves: []WaveJSON{}},
		},
	}

	// Populate Ri
	for i, val := range meas.ColorRenditionIndexes.Ri {
		res.CRI.Ri = append(res.CRI.Ri, CRIRiJSON{
			Ri:    i + 1,
			Value: val.Val,
		})
	}

	// Populate 1nm
	for i, val := range &meas.SpectralData1nm {
		res.WaveLengths[0].Waves = append(res.WaveLengths[0].Waves, WaveJSON{
			Nm:    380 + i,
			Value: val.Val,
		})
	}

	// Populate 5nm
	for i, val := range &meas.SpectralData5nm {
		res.WaveLengths[1].Waves = append(res.WaveLengths[1].Waves, WaveJSON{
			Nm:    380 + (i * 5),
			Value: val.Val,
		})
	}

	return res
}
