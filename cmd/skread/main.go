package main

import (
	"fmt"
	"strconv"

	"github.com/akares/skreader"
)

func main() {
	sk, err := skreader.NewDeviceWithAdapter(&skreader.GousbAdapter{})
	if err != nil {
		panic(err)
	}
	defer sk.Close()

	model, _ := sk.ModelName()
	fw, _ := sk.FirmwareVersion()

	fmt.Println(strconv.Quote(sk.String()))
	fmt.Println("MN:", strconv.Quote(model))
	fmt.Println("FW:", fw)

	st, err := sk.State()
	if err != nil {
		panic(err)
	}
	fmt.Printf("ST: %+v\n", st)

	meas, err := sk.Measure()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Meas: %s\n", meas.Repr())
	fmt.Printf("Meas: %s\n", meas.String())
}
