package skreader

// Assert FakeusbAdapter implements UsbDevice adapter interface
var _ UsbAdapter = (*FakeusbAdapter)(nil)

// FakeusbAdapter implements UsbAdapter interface for testing purposes.
// It does nothing but allows to mock USB device responses by setting up desired values in struct fields.
type FakeusbAdapter struct {
	OpenResponse         error
	CloseResponse        error
	WriteResponse        error
	ReadResponse         []FakeusbAdapterReadResponse
	ReadResponseIndex    int
	ManufacturerResponse FakeusbAdapterManufacturerResponse
	ProductResponse      FakeusbAdapterProductResponse
}

type FakeusbAdapterReadResponse struct {
	Data []byte
	Err  error
}

type FakeusbAdapterManufacturerResponse struct {
	Val string
	Err error
}

type FakeusbAdapterProductResponse struct {
	Val string
	Err error
}

func (f *FakeusbAdapter) Open() (err error) {
	return f.OpenResponse
}

func (f *FakeusbAdapter) Close() (err error) {
	return f.CloseResponse
}

func (f *FakeusbAdapter) Read(buf []byte) (int, error) {
	r := f.ReadResponse[f.ReadResponseIndex]
	n := copy(buf, r.Data)
	f.ReadResponseIndex++

	return n, r.Err
}

func (f *FakeusbAdapter) Write(_ []byte) (int, error) {
	return 2, f.WriteResponse
}

func (f *FakeusbAdapter) Manufacturer() (string, error) {
	return f.ManufacturerResponse.Val, f.ManufacturerResponse.Err
}

func (f *FakeusbAdapter) Product() (string, error) {
	return f.ProductResponse.Val, f.ProductResponse.Err
}
