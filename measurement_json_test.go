package skreader_test

import (
	"testing"
	"time"

	"github.com/akares/skreader"
)

//nolint:gocyclo
func TestMeasurementJSON(t *testing.T) {
	now := time.Now()

	for _, tt := range []struct {
		name      string
		measName  string
		measNote  string
		testdata  []byte
		wantRange skreader.ValueRange
		wantErr   bool
	}{
		{
			name:      "range ok",
			measName:  "test",
			measNote:  "test",
			testdata:  skreader.Testdata,
			wantRange: skreader.RangeOk,
			wantErr:   false,
		},
		{
			name:      "range under",
			measName:  "test",
			measNote:  "test",
			testdata:  skreader.TestdataUnder,
			wantRange: skreader.RangeUnder,
			wantErr:   false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m, err := skreader.NewMeasurementFromBytes(tt.testdata)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			mjs := skreader.NewJSONMeasurement(m, tt.measName, tt.measNote, now)

			if mjs.Name != tt.measName {
				t.Errorf("Name = %v, want %v", mjs.Name, tt.measName)
			}
			if mjs.Note != tt.measNote {
				t.Errorf("Note = %v, want %v", mjs.Note, tt.measNote)
			}

			if mjs.Tristimulus.X != m.Tristimulus.X.Val {
				t.Errorf("Tristimulus.X = %v, want %v", mjs.Tristimulus.X, m.Tristimulus.X.Val)
			}
			if mjs.Tristimulus.Y != m.Tristimulus.Y.Val {
				t.Errorf("Tristimulus.Y = %v, want %v", mjs.Tristimulus.Y, m.Tristimulus.Y.Val)
			}
			if mjs.Tristimulus.Z != m.Tristimulus.Z.Val {
				t.Errorf("Tristimulus.Z = %v, want %v", mjs.Tristimulus.Z, m.Tristimulus.Z.Val)
			}
			if mjs.ColorTemperature.CCT != m.ColorTemperature.Tcp.Val {
				t.Errorf("ColorTemperature.CCT = %v, want %v", mjs.ColorTemperature.CCT, m.ColorTemperature.Tcp.Val)
			}
			if mjs.ColorTemperature.DeltaUv != m.ColorTemperature.DeltaUv.Val {
				t.Errorf("ColorTemperature.DeltaUv = %v, want %v", mjs.ColorTemperature.DeltaUv, m.ColorTemperature.DeltaUv.Val)
			}
			if mjs.Illuminance.Fc != m.Illuminance.FootCandle.Val {
				t.Errorf("Illuminance.Fc = %v, want %v", mjs.Illuminance.Fc, m.Illuminance.FootCandle.Val)
			}
			if mjs.Illuminance.LUX != m.Illuminance.Lux.Val {
				t.Errorf("Illuminance.LUX = %v, want %v", mjs.Illuminance.LUX, m.Illuminance.Lux.Val)
			}
			if mjs.CIE1931.X != m.CIE1931.X.Val {
				t.Errorf("CIE1931.X = %v, want %v", mjs.CIE1931.X, m.CIE1931.X.Val)
			}
			if mjs.CIE1931.Y != m.CIE1931.Y.Val {
				t.Errorf("CIE1931.Y = %v, want %v", mjs.CIE1931.Y, m.CIE1931.Y.Val)
			}
			if mjs.CIE1976.Ud != m.CIE1976.Ud.Val {
				t.Errorf("CIE1976.Ud = %v, want %v", mjs.CIE1976.Ud, m.CIE1976.Ud.Val)
			}
			if mjs.CIE1976.Vd != m.CIE1976.Vd.Val {
				t.Errorf("CIE1976.Vd = %v, want %v", mjs.CIE1976.Vd, m.CIE1976.Vd.Val)
			}
			if mjs.CRI.Ra != m.ColorRenditionIndexes.Ra.Val {
				t.Errorf("CRI.RA = %v, want %v", mjs.CRI.Ra, m.ColorRenditionIndexes.Ra.Val)
			}
			if len(mjs.CRI.Ri) != len(m.ColorRenditionIndexes.Ri) {
				t.Errorf("CRI.Ri = %v, want %v", mjs.CRI.Ri, m.ColorRenditionIndexes.Ri)
			}
			if mjs.DWL.Wavelength != m.DWL.Wavelength.Val {
				t.Errorf("DWL.Wavelength = %v, want %v", mjs.DWL.Wavelength, m.DWL.Wavelength.Val)
			}
			if mjs.DWL.ExcitationPurity != m.DWL.ExcitationPurity.Val {
				t.Errorf("DWL.ExcitationPurity = %v, want %v", mjs.DWL.ExcitationPurity, m.DWL.ExcitationPurity.Val)
			}
			if mjs.SpectralData[0].Range.Type != "1nm" {
				t.Errorf("SpectralData[0].Range.Type = %v, want %v", mjs.SpectralData[0].Range.Type, "1nm")
			}
			if mjs.SpectralData[1].Range.Type != "5nm" {
				t.Errorf("SpectralData[1].Range.Type = %v, want %v", mjs.SpectralData[0].Range.Type, "5nm")
			}
		})
	}
}
