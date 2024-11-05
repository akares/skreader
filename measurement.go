package skreader

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Measurement represents a measurement data from SEKONIC device.
// Data format is based on original C-7000 SDK from SEKONIC (distributed only as Windows DLL).
type Measurement struct {
	Tristimulus      TristimulusValue        // Tristimulus values in XYZ color space
	ColorTemperature ColorTemperatureValue   // Correlated Color Temperature
	Illuminance      IlluminanceValue        // Illuminance
	CIE1931          CIE1931Value            // CIE 1931 (x, y, z) chromaticity coordinates
	CIE1976          CIE1976Value            // CIE 1976 (u', v') chromaticity coordinates
	DWL              DominantWavelengthValue // Dominant Wavelength
	PPFD             DecimalValue            // Photosynthetic Photon Flux Density

	ColorRenditionIndexes ColorRenditionIndexesValue // Color Rendition Indexes

	SpectralData5nm [81]DecimalValue  // Spectral Data (5nm)
	SpectralData1nm [401]DecimalValue // Spectral Data (1nm)
	PeakWavelength  int               // Peak Wavelength (380...780nm)
}

const (
	MeasurementDataValidSize = 2380 // tested on C-7000, C-800, C-700
)

// ValueRange indicates if a measurement result value is within/over/under limits.
type ValueRange int

const (
	RangeOk ValueRange = iota
	RangeUnder
	RangeOver
)

func (r ValueRange) String() string {
	switch r {
	case RangeOk:
		return "Ok"
	case RangeUnder:
		return "Under"
	case RangeOver:
		return "Over"
	default:
		return "Unknown"
	}
}

// DecimalValue represents a decimal value with string representation and validity indicator.
type DecimalValue struct {
	Val   float64    // value
	Str   string     // raw string representation
	Range ValueRange // value validity indicator
}

// String returns string representation of the DecimalValue instance.
// If the value is out of limits, the Range string representation is returned.
func (v DecimalValue) String() string {
	if v.Range != RangeOk {
		return v.Range.String()
	}

	return v.Str
}

// TristimulusValue represents a tristimulus value in XYZ color space.
type TristimulusValue struct {
	X DecimalValue
	Y DecimalValue
	Z DecimalValue
}

// ColorTemperatureValue represents a correlated color temperature value in Kelvin.
type ColorTemperatureValue struct {
	//nolint:all
	Tcp     DecimalValue // Correlated Color Temperature
	DeltaUv DecimalValue // Deviation from the Planckian locus
}

// IlluminanceValue represents an illuminance value in Lux and foot-candle units.
type IlluminanceValue struct {
	Lux        DecimalValue
	FootCandle DecimalValue
}

// CIE1931Value represents a CIE 1931 (x, y, z) chromaticity coordinates.
type CIE1931Value struct {
	X DecimalValue
	Y DecimalValue
	Z DecimalValue
}

// CIE1976Value represents a CIE 1976 (u', v') chromaticity coordinates.
type CIE1976Value struct {
	Ud DecimalValue
	Vd DecimalValue
}

// DominantWavelengthValue represents a dominant wavelength value in nm and excitation purity in %.
type DominantWavelengthValue struct {
	Wavelength       DecimalValue
	ExcitationPurity DecimalValue
}

// ColorRenditionIndexesValue represents a color rendition indexes Ra and Ri.
type ColorRenditionIndexesValue struct {
	Ra DecimalValue
	Ri [15]DecimalValue
}

// TODO: Not implemented here but available for C-7000 FW > 25 extended measurement data:
// TM30, SSI, TLCI

// NewMeasurementFromBytes creates a new Measurement instance from the given raw
// binary response from SEKONIC device.
// Note: currently only ambient measuring mode results are supported.
//
//nolint:exhaustruct,funlen,gomnd,gocyclo
func NewMeasurementFromBytes(data []byte) (*Measurement, error) {
	if len(data) < MeasurementDataValidSize {
		return nil, fmt.Errorf("invalid measurement data size: %d < %d bytes", len(data), MeasurementDataValidSize)
	}

	// Parse binary data to struct.
	//
	// Data offsets and sizes are based on SEKONIC USB data packet layout which
	// seems to be stable between various devices.
	//
	// Magic numbers for limits and precisions are based on original C-7000 SDK.

	m := &Measurement{}

	// Color temperature and deviation from the Planckian locus
	m.ColorTemperature.Tcp = toDecimalValue(parseFloat32(data, 50), 1563, 100000, 0)
	m.ColorTemperature.DeltaUv = toDecimalValue(parseFloat32(data, 55), -0.1, 0.1, 4)
	if m.ColorTemperature.DeltaUv.Range != RangeOk { // limit the CCT value (C-800 returns Tcp=50000 value instead of "Over" as C-7000 does)
		m.ColorTemperature.Tcp.Range = m.ColorTemperature.DeltaUv.Range
	}

	// Illuminance values in Lux and foot-candle units
	m.Illuminance.Lux = parseLuxToDecimalValue(data, 271, 100, 200000)
	m.Illuminance.FootCandle = parseLuxToDecimalValue(data, 276, 0.093000002205371857, 18580.607421875)

	// Tristimulus values in XYZ color space
	m.Tristimulus.X = toDecimalValue(parseFloat64(data, 281), 0, 1000000, 4)
	m.Tristimulus.Y = toDecimalValue(parseFloat64(data, 290), 0, 1000000, 4)
	m.Tristimulus.Z = toDecimalValue(parseFloat64(data, 299), 0, 1000000, 4)

	// CIE1931 (x, y, z) chromaticity coordinates
	m.CIE1931.X = toDecimalValue(parseFloat32(data, 308), 0, 1, 4)
	m.CIE1931.Y = toDecimalValue(parseFloat32(data, 313), 0, 1, 4)
	if m.CIE1931.X.Range != RangeOk {
		m.CIE1931.Z.Range = m.CIE1931.X.Range
	} else if m.CIE1931.Y.Range != RangeOk {
		m.CIE1931.Z.Range = m.CIE1931.Y.Range
	} else {
		m.CIE1931.Z = toDecimalValue(1.0-m.CIE1931.X.Val-m.CIE1931.Y.Val, 0, 1, 4)
	}

	// CIE1976 (u', v') chromaticity coordinates
	m.CIE1976.Ud = toDecimalValue(parseFloat32(data, 328), 0, 1, 4)
	m.CIE1976.Vd = toDecimalValue(parseFloat32(data, 333), 0, 1, 4)

	// Dominant Wavelength
	m.DWL.Wavelength = toDecimalValue(parseFloat32(data, 338), -780, 780, 0)
	m.DWL.ExcitationPurity = toDecimalValue(parseFloat32(data, 343), 0, 100, 1)

	// CRI (Ra, Ri)
	m.ColorRenditionIndexes.Ra = toDecimalValue(parseFloat32(data, 348), -100, 100, 1)
	for i := range m.ColorRenditionIndexes.Ri {
		m.ColorRenditionIndexes.Ri[i] = toDecimalValue(parseFloat32(data, 353+i*5), -100, 100, 1)
	}

	// Boundaries check

	if m.Illuminance.Lux.Range == RangeUnder {
		for i := range m.SpectralData5nm {
			m.SpectralData5nm[i].Range = RangeUnder
		}
		for i := range m.SpectralData1nm {
			m.SpectralData1nm[i].Range = RangeUnder
		}
	} else if m.Illuminance.Lux.Range == RangeOver {
		for i := range m.SpectralData5nm {
			m.SpectralData5nm[i].Range = RangeOver
			m.SpectralData5nm[i].Val = 9999.9
		}
		for i := range m.SpectralData1nm {
			m.SpectralData1nm[i].Range = RangeOver
			m.SpectralData1nm[i].Val = 9999.9
		}
	} else {
		for i := range m.SpectralData5nm {
			m.SpectralData5nm[i] = toDecimalValue(parseFloat32(data, 428+i*4), 0, 9999.9, 8)
		}
		m.PeakWavelength = 380
		maxval := m.SpectralData1nm[0].Val
		for i := range m.SpectralData1nm {
			m.SpectralData1nm[i] = toDecimalValue(parseFloat32(data, 753+i*4), 0, 9999.9, 8)
			if m.SpectralData1nm[i].Val > 0 && m.SpectralData1nm[i].Val > maxval {
				maxval = m.SpectralData1nm[i].Val
				m.PeakWavelength = 380 + i
			}
		}
	}

	m.PPFD = toDecimalValue(parseFloat32(data, 2376), 0, 9999.9, 1)

	// Boundaries extra check

	if m.Illuminance.Lux.Range == RangeOk && m.Illuminance.Lux.Val < 5 {
		m.ColorTemperature.Tcp.Range = RangeUnder
		m.ColorTemperature.DeltaUv.Range = RangeUnder
		m.CIE1931.X.Range = RangeUnder
		m.CIE1931.Y.Range = RangeUnder
		m.CIE1931.Z.Range = RangeUnder
		m.CIE1976.Ud.Range = RangeUnder
		m.CIE1976.Vd.Range = RangeUnder
		m.DWL.Wavelength.Range = RangeUnder
		m.DWL.ExcitationPurity.Range = RangeUnder
		m.ColorRenditionIndexes.Ra.Range = RangeUnder
		for i := range m.ColorRenditionIndexes.Ri {
			m.ColorRenditionIndexes.Ri[i].Range = RangeUnder
		}
	}

	if m.ColorTemperature.Tcp.Range != RangeOk {
		m.ColorTemperature.DeltaUv.Range = m.ColorTemperature.Tcp.Range
		m.ColorRenditionIndexes.Ra.Range = m.ColorTemperature.Tcp.Range
		for i := range m.ColorRenditionIndexes.Ri {
			m.ColorRenditionIndexes.Ri[i].Range = m.ColorTemperature.Tcp.Range
		}
	}

	return m, nil
}

// String returns limited string representation of the Measurement instance.
// Used mostly for debugging.
func (m *Measurement) String() string {
	return fmt.Sprintf("Lux=%s x=%s y=%s CCT=%s", m.Illuminance.Lux.Str, m.CIE1931.X.Str, m.CIE1931.Y.Str, m.ColorTemperature.Tcp.Str)
}

// Repr returns full string representation of the Measurement instance.
// Used mostly for debugging.
func (m *Measurement) Repr() string {
	return fmt.Sprintf("%+v", *m)
}

// parseFloat32 parses a float32 value from the given data slice.
func parseFloat32(data []byte, offset int) float64 {
	return float64(math.Float32frombits(binary.BigEndian.Uint32(data[offset : offset+4])))
}

// parseFloat64 parses a float64 value from the given data slice.
func parseFloat64(data []byte, offset int) float64 {
	return math.Float64frombits(binary.BigEndian.Uint64(data[offset : offset+8]))
}

// parseLux parses a float32 value from the given data slice and returns DecimalValue.
// It's like a parseFloat32 but with a more specific precision calc related to Lux measurement.
// Magic numbers are based on original C-7000 SDK from SEKONIC.
//
//nolint:gomnd
func parseLuxToDecimalValue(data []byte, offset int, lowRange, highRange float64) DecimalValue {
	val := parseFloat32(data, offset)

	if val < 9.9499998092651367 {
		val = round(val, 2)
	} else if val < 99.949996948242188 {
		val = round(val, 1)
	} else if val < 999.5 {
		val = round(val, 0)
	} else if val < 9995.0 {
		val = round(val/10.0, 0) * 10.0
	} else if val < 99950.0 {
		val = round(val/100.0, 0) * 100.0
	} else {
		val = round(val/1000.0, 0) * 1000.0
	}

	var precision int
	if val < 100 {
		precision = 1
	}

	return toDecimalValue(val, lowRange, highRange, precision)
}

// toDecimalValue converts the given float64 value to DecimalValue instance.
// String representation is rounded to the given precision.
func toDecimalValue(val, lowRange, highRange float64, precision int) DecimalValue {
	v := val
	s := fmt.Sprintf("%.*f", precision, v)
	r := RangeOk

	if v < lowRange {
		r = RangeUnder
		s = r.String()
	} else if v > highRange {
		r = RangeOver
		s = r.String()
	}

	return DecimalValue{
		Val:   v,
		Str:   s,
		Range: r,
	}
}

func round(val float64, precision int) float64 {
	return math.Round(val*(math.Pow10(precision))) / math.Pow10(precision)
}
