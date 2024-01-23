package skreader_test

import (
	"testing"

	"github.com/akares/skreader"
)

func TestMeasurementRange(t *testing.T) {
	for _, tt := range []struct {
		name      string
		testdata  []byte
		wantRange skreader.ValueRange
		wantErr   bool
	}{
		{
			name:      "range ok",
			testdata:  skreader.Testdata,
			wantRange: skreader.RangeOk,
			wantErr:   false,
		},
		{
			name:      "range under",
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
			if m.Illuminance.Lux.Range != tt.wantRange {
				t.Errorf("got %v, want %v", m.Illuminance.Lux.Range, tt.wantRange)
			}
		})
	}
}
