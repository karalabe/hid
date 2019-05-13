// hid - Gopher Interface Devices (USB HID)
// Copyright (c) 2017 Péter Szilágyi. All rights reserved.
//
// This file is released under the 3-clause BSD license. Note however that Linux
// support depends on libusb, released under GNU LGPL 2.1 or later.

// +build !linux,!darwin,!windows ios !cgo

package hid

// Supported returns whether this platform is supported by the HID library or not.
// The goal of this method is to allow programatically handling platforms that do
// not support USB HID and not having to fall back to build constraints.
func Supported() bool {
	return false
}

// Enumerate returns a list of all the HID devices attached to the system which
// match the vendor and product id. On platforms that this file implements the
// function is a noop and returns an empty list always.
func Enumerate(vendorID uint16, productID uint16) []DeviceInfo {
	return nil
}

// HidDevice is a live HID USB connected device handle. On platforms that this file
// implements the type lacks the actual HID device and all methods are noop.
type HidDevice struct {
	HidDeviceInfo // Embed the infos for easier access
}

// Open connects to an HID device by its path name. On platforms that this file
// implements the method just returns an error.
func (info HidDeviceInfo) Open() (*Device, error) {
	return nil, ErrUnsupportedPlatform
}

// Close releases the HID USB device handle. On platforms that this file implements
// the method is just a noop.
func (dev *HidDevice) Close() error { return nil }

// Write sends an output report to a HID device. On platforms that this file
// implements the method just returns an error.
func (dev *HidDevice) Write(b []byte) (int, error) {
	return 0, ErrUnsupportedPlatform
}

// GenericDeviceHandle represents a libusb device_handle struct
type GenericDeviceHandle *C.struct_libusb_device_handle

// GenericLibUsbDevice represents a libusb device struct
type GenericLibUsbDevice *C.struct_libusb_device

// Read retrieves an input report from a HID device. On platforms that this file
// implements the method just returns an error.
func (dev *HidDevice) Read(b []byte) (int, error) {
	return 0, ErrUnsupportedPlatform
}

// GenericDeviceOpen is a helper function to call the C version of open.
func GenericDeviceOpen(dev GenericLibUsbDevice) (GenericDeviceHandle, error) {
	return nil, ErrUnsupportedPlatform
}

// GenericDeviceClose is a helper function to close a libusb device
func GenericDeviceClose(handle GenericDeviceHandle) {
}

// InterruptTransfer is a helpler function for libusb's interrupt transfer function
func InterruptTransfer(handle GenericDeviceHandle, endpoint uint8, data []byte, timeout uint) ([]byte, error) {
	return data[:0], ErrUnsupportedPlatform
}
