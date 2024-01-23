package skreader_test

import (
	"errors"
	"testing"

	sekonic "github.com/akares/skreader"
)

var testSKResponseOK = []byte{6, 48}

func TestNewDeviceWithAdapter(t *testing.T) {
	//nolint:exhaustruct
	for _, tt := range []struct {
		name    string
		adapter sekonic.UsbAdapter
		wantErr bool
	}{
		{
			name:    "nil adapter",
			adapter: nil,
			wantErr: true,
		},
		{
			name:    "valid adapter",
			adapter: &sekonic.FakeusbAdapter{},
			wantErr: false,
		},
		{
			name: "open error",
			adapter: &sekonic.FakeusbAdapter{
				OpenResponse: errors.New("open error"),
			},
			wantErr: true,
		},
		{
			name: "manufacturer error",
			adapter: &sekonic.FakeusbAdapter{
				ManufacturerResponse: sekonic.FakeusbAdapterManufacturerResponse{
					Err: errors.New("manufacturer error"),
				},
			},
			wantErr: true,
		},
		{
			name: "product error",
			adapter: &sekonic.FakeusbAdapter{
				ProductResponse: sekonic.FakeusbAdapterProductResponse{
					Err: errors.New("product error"),
				},
			},
			wantErr: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sekonic.NewDeviceWithAdapter(tt.adapter)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDeviceWithAdapter() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
		})
	}
}

func TestString(t *testing.T) {
	//nolint:exhaustruct
	for _, tt := range []struct {
		name    string
		adapter sekonic.UsbAdapter
		want    string
	}{
		{
			name:    "empty default",
			adapter: &sekonic.FakeusbAdapter{},
			want:    "SEKONIC",
		},
		{
			name: "full",
			adapter: &sekonic.FakeusbAdapter{
				ManufacturerResponse: sekonic.FakeusbAdapterManufacturerResponse{
					Val: "FakeManufacturer",
				},
				ProductResponse: sekonic.FakeusbAdapterProductResponse{
					Val: "FakeProduct",
				},
			},
			want: "FakeManufacturer FakeProduct",
		},
		{
			name: "only manufacturer",
			adapter: &sekonic.FakeusbAdapter{
				ManufacturerResponse: sekonic.FakeusbAdapterManufacturerResponse{
					Val: "FakeManufacturer",
				},
			},
			want: "FakeManufacturer",
		},
		{
			name: "only product",
			adapter: &sekonic.FakeusbAdapter{
				ProductResponse: sekonic.FakeusbAdapterProductResponse{
					Val: "FakeProduct",
				},
			},
			want: "FakeProduct",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			d, _ := sekonic.NewDeviceWithAdapter(tt.adapter)

			if got := d.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClose(t *testing.T) {
	//nolint:exhaustruct
	for _, tt := range []struct {
		name    string
		adapter sekonic.UsbAdapter
		wantErr bool
	}{
		{
			name: "valid adapter",
			adapter: &sekonic.FakeusbAdapter{
				CloseResponse: nil,
			},
			wantErr: false,
		},
		{
			name: "close error",
			adapter: &sekonic.FakeusbAdapter{
				CloseResponse: errors.New("close error"),
			},
			wantErr: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			d, _ := sekonic.NewDeviceWithAdapter(tt.adapter)

			if err := d.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestModelName(t *testing.T) { //nolint:funlen
	//nolint:exhaustruct
	for _, tt := range []struct {
		name    string
		adapter sekonic.UsbAdapter
		want    string
		wantErr bool
	}{
		{
			name: "valid adapter",
			adapter: &sekonic.FakeusbAdapter{
				ReadResponse: []sekonic.FakeusbAdapterReadResponse{
					{
						Data: testSKResponseOK,
						Err:  nil,
					},
					{
						Data: []byte("MN@@@C-800\x00\x00\x00\x00\x00"),
						Err:  nil,
					},
				},
			},
			want:    "C-800",
			wantErr: false,
		},
		{
			name: "invalid ACK",
			adapter: &sekonic.FakeusbAdapter{
				ReadResponse: []sekonic.FakeusbAdapterReadResponse{
					{
						Data: []byte{12, 34},
						Err:  nil,
					},
					{
						Data: []byte("MN@@@C-800\x00\x00\x00\x00\x00"),
						Err:  nil,
					},
				},
			},
			want:    "C-800",
			wantErr: true,
		},
		{
			name: "invalid command ACK",
			adapter: &sekonic.FakeusbAdapter{
				ReadResponse: []sekonic.FakeusbAdapterReadResponse{
					{
						Data: testSKResponseOK,
						Err:  nil,
					},
					{
						Data: []byte("XX@@@C-800\x00\x00\x00\x00\x00"),
						Err:  nil,
					},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "empty data",
			adapter: &sekonic.FakeusbAdapter{
				ReadResponse: []sekonic.FakeusbAdapterReadResponse{
					{
						Data: testSKResponseOK,
						Err:  nil,
					},
					{
						Data: []byte(""),
						Err:  nil,
					},
				},
			},
			want:    "C-800",
			wantErr: true,
		},
		{
			name: "shortest data",
			adapter: &sekonic.FakeusbAdapter{
				ReadResponse: []sekonic.FakeusbAdapterReadResponse{
					{
						Data: testSKResponseOK,
						Err:  nil,
					},
					{
						Data: []byte("MN@@@"),
						Err:  nil,
					},
				},
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "too short data",
			adapter: &sekonic.FakeusbAdapter{
				ReadResponse: []sekonic.FakeusbAdapterReadResponse{
					{
						Data: testSKResponseOK,
						Err:  nil,
					},
					{
						Data: []byte("MN@@"),
						Err:  nil,
					},
				},
			},
			want:    "C-800",
			wantErr: true,
		},
		{
			name: "read error ACK",
			adapter: &sekonic.FakeusbAdapter{
				ReadResponse: []sekonic.FakeusbAdapterReadResponse{
					{
						Data: nil,
						Err:  errors.New("read error"),
					},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "read error data",
			adapter: &sekonic.FakeusbAdapter{
				ReadResponse: []sekonic.FakeusbAdapterReadResponse{
					{
						Data: testSKResponseOK,
						Err:  nil,
					},
					{
						Data: nil,
						Err:  errors.New("read error"),
					},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "write error",
			adapter: &sekonic.FakeusbAdapter{
				WriteResponse: errors.New("write error"),
				ReadResponse: []sekonic.FakeusbAdapterReadResponse{
					{
						Data: testSKResponseOK,
						Err:  nil,
					},
					{
						Data: []byte("MN@@@C-800\x00\x00\x00\x00\x00"),
						Err:  nil,
					},
				},
			},
			want:    "",
			wantErr: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			d, _ := sekonic.NewDeviceWithAdapter(tt.adapter)

			if got, err := d.ModelName(); (err != nil) != tt.wantErr {
				t.Errorf("ModelName() = %s, want %v, error = %v, wantErr %v", got, tt.want, err, tt.wantErr)
			}
		})
	}
}

func TestFirmwareVersion(t *testing.T) { //nolint:funlen
	//nolint:exhaustruct
	for _, tt := range []struct {
		name    string
		adapter sekonic.UsbAdapter
		want    int
		wantErr bool
	}{
		{
			name: "valid adapter",
			adapter: &sekonic.FakeusbAdapter{
				ReadResponse: []sekonic.FakeusbAdapterReadResponse{
					{
						Data: testSKResponseOK,
						Err:  nil,
					},
					{
						Data: []byte("FV@@@20,C36E,27,7881,11,B216,14,50CC,17,74EC"),
						Err:  nil,
					},
				},
			},
			want:    27,
			wantErr: false,
		},
		{
			name: "invalid ACK",
			adapter: &sekonic.FakeusbAdapter{
				ReadResponse: []sekonic.FakeusbAdapterReadResponse{
					{
						Data: []byte{12, 34},
						Err:  nil,
					},
					{
						Data: []byte("FV@@@20,C36E,27,7881,11,B216,14,50CC,17,74EC"),
						Err:  nil,
					},
				},
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "invalid command ACK",
			adapter: &sekonic.FakeusbAdapter{
				ReadResponse: []sekonic.FakeusbAdapterReadResponse{
					{
						Data: testSKResponseOK,
						Err:  nil,
					},
					{
						Data: []byte("XX@@@20,C36E,27,7881,11,B216,14,50CC,17,74EC"),
						Err:  nil,
					},
				},
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "empty data",
			adapter: &sekonic.FakeusbAdapter{
				ReadResponse: []sekonic.FakeusbAdapterReadResponse{
					{
						Data: testSKResponseOK,
						Err:  nil,
					},
					{
						Data: []byte(""),
						Err:  nil,
					},
				},
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "short data",
			adapter: &sekonic.FakeusbAdapter{
				ReadResponse: []sekonic.FakeusbAdapterReadResponse{
					{
						Data: testSKResponseOK,
						Err:  nil,
					},
					{
						Data: []byte("FW@@@"),
						Err:  nil,
					},
				},
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "invalid data",
			adapter: &sekonic.FakeusbAdapter{
				ReadResponse: []sekonic.FakeusbAdapterReadResponse{
					{
						Data: testSKResponseOK,
						Err:  nil,
					},
					{
						Data: []byte("FV@@@20,C36E,FF,7881,11,B216,14,50CC,17,74EC"),
						Err:  nil,
					},
				},
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "read error ACK",
			adapter: &sekonic.FakeusbAdapter{
				ReadResponse: []sekonic.FakeusbAdapterReadResponse{
					{
						Data: nil,
						Err:  errors.New("read error"),
					},
				},
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "read error data",
			adapter: &sekonic.FakeusbAdapter{
				ReadResponse: []sekonic.FakeusbAdapterReadResponse{
					{
						Data: testSKResponseOK,
						Err:  nil,
					},
					{
						Data: nil,
						Err:  errors.New("read error"),
					},
				},
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "write error",
			adapter: &sekonic.FakeusbAdapter{
				WriteResponse: errors.New("write error"),
				ReadResponse: []sekonic.FakeusbAdapterReadResponse{
					{
						Data: testSKResponseOK,
						Err:  nil,
					},
					{
						Data: []byte("FV@@@20,C36E,27,7881,11,B216,14,50CC,17,74EC"),
						Err:  nil,
					},
				},
			},
			want:    0,
			wantErr: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			d, _ := sekonic.NewDeviceWithAdapter(tt.adapter)

			if got, err := d.FirmwareVersion(); (err != nil) != tt.wantErr {
				t.Errorf("FirmwareVersion() = %d, want %d, error = %v, wantErr %v", got, tt.want, err, tt.wantErr)
			}
		})
	}
}
