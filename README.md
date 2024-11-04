# skreader

[![Test Go code](https://github.com/akares/skreader/actions/workflows/test.yml/badge.svg)](https://github.com/akares/skreader/actions/workflows/test.yml) [![Lint Go code](https://github.com/akares/skreader/actions/workflows/lint.yml/badge.svg)](https://github.com/akares/skreader/actions/workflows/lint.yml)

Golang library and example command line tool for SEKONIC spectrometers remote control.

Based on original C# SDK for Windows from SEKONIC.

<img src="doc/Sekonic-C-7000.jpg" width="640" alt="Sekonic C-7000" />

## Supported (tested) models

-   Sekonic C-700
-   Sekonic C-800
-   Sekonic C-800-U (US Version. Basically the same as C-800)
-   Sekonic C-7000 (supports extended measurement configuration: FOV and Exposure Time)

## Supported (tested) platforms

-   Darwin
-   Windows
-   Linux

## Known limitations

TM-30 and TCLI only works on C-7000 firmware accordingly to SEKONIC SDK manual page 11?

## Dependencies

Default implementation uses [gousb](https://github.com/google/gousb) wrapper for the libusb library.

You must have [libusb-1.0](https://github.com/libusb/libusb/wiki) installed on your target system to be able to communicate with USB devices.

Installation for different platforms is covered in
[gousb documentation](https://github.com/google/gousb/blob/master/README.md#dependencies).

_Alternatively_ you can provide custom USB implementation with [simple interface](usbadapter.go) close to io.Reader. See the default [gousb based implementation](gousb_adapter.go) for reference.

## SDK usage

See the [skread](cmd/skread/main.go) command implementation.

## MacOS install example


Use homebrew to install go.
go on MacOS will take care of dependencies when running this script for the first time.

```
homebrew install go
```

Connect your device.
Then simply run example below.



## Run example

```
go run ./cmd/skread --run --all
```

```
go run ./cmd/skread --help
```

## License

This project is licensed under the terms of the MIT license.

## Legal notices

All product names, logos, and brands are property of their respective owners. All company, product and service names used in this package are for identification purposes only. Use of these names, logos, and brands does not imply endorsement.

-   SEKONIC is a registered trademark of SEKONIC CORPORATION.
-   Google is a registered trademark of Google LLC.
-   Windows is a registered trademark of Microsoft Corporation.
-   Linux is the registered trademark of Linus Torvalds.
