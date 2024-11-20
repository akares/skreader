package skreader_test

import (
	"testing"
	"time"

	"github.com/akares/skreader"
)

func TestMeasurementSPDX(t *testing.T) {
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

			spdxHeader := skreader.NewSPDXHeader(tt.measName, tt.measNote, now)

			if spdxHeader.Comments != tt.measName {
				t.Errorf("Comments = %v, want %v", spdxHeader.Comments, tt.measName)
			}
			if spdxHeader.Description != tt.measName {
				t.Errorf("Description = %v, want %v", spdxHeader.Description, tt.measNote)
			}

			spdxSpectralDistribution := skreader.NewSPDXSpectralDistribution(m)

			if spdxSpectralDistribution.SpectralData[0].Value != m.SpectralData1nm[0].Val {
				t.Errorf("SpectralData[0].Value = %v, want %v", spdxSpectralDistribution.SpectralData[0].Value, m.SpectralData1nm[0].Val)
			}
			if spdxSpectralDistribution.SpectralData[400].Value != m.SpectralData1nm[400].Val {
				t.Errorf("SpectralData[400].Value = %v, want %v", spdxSpectralDistribution.SpectralData[400].Value, m.SpectralData1nm[400].Val)
			}
		})
	}
}
