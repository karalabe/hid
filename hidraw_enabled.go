// hid - Gopher Interface Devices (USB HID)
// Copyright (c) 2017 Péter Szilágyi. All rights reserved.
//
// This file is released under the 3-clause BSD license. 

// +build linux,hidraw,cgo

package hid

/*
#cgo CFLAGS: -I./hidapi/hidapi

#cgo linux CFLAGS: -I./libusb/libusb -DDEFAULT_VISIBILITY="" -DOS_LINUX -D_GNU_SOURCE -DPOLL_NFDS_TYPE=int
#cgo linux,!android LDFLAGS: -lrt
#cgo linux LDFLAGS: -ludev


	#include <sys/poll.h>
	#include "os/threads_posix.c"
	#include "os/poll_posix.c"

	#include "os/linux_usbfs.c"
	#include "os/linux_netlink.c"

	#include "core.c"
	#include "descriptor.c"
	#include "hotplug.c"
	#include "io.c"
	#include "strerror.c"
	#include "sync.c"

	#include "hidapi/linux/hid.c"

	static uint32_t
	get_bytes (uint8_t * rpt, size_t len, size_t num_bytes, size_t cur)
	{
	  // Return if there aren't enough bytes.
	  if (cur + num_bytes >= len)
	    return 0;

	  if (num_bytes == 0)
	    return 0;
	  else if (num_bytes == 1)
	    {
	      return rpt[cur + 1];
	    }
	  else if (num_bytes == 2)
	    {
	      return (rpt[cur + 2] * 256 + rpt[cur + 1]);
	    }
	  else
	    return 0;
	}

	static int
	get_usage (uint8_t * report_descriptor, size_t size,
		   unsigned short *usage_page, unsigned short *usage)
	{
	  size_t i = 0;
	  int size_code;
	  int data_len, key_size;
	  int usage_found = 0, usage_page_found = 0;

	  while (i < size)
	    {
	      int key = report_descriptor[i];
	      int key_cmd = key & 0xfc;

	      if ((key & 0xf0) == 0xf0)
		{
		  fprintf (stderr, "invalid data received.\n");
		  return -1;
		}
	      else
		{
		  size_code = key & 0x3;
		  switch (size_code)
		    {
		    case 0:
		    case 1:
		    case 2:
		      data_len = size_code;
		      break;
		    case 3:
		      data_len = 4;
		      break;
		    default:
		      // Can't ever happen since size_code is & 0x3
		      data_len = 0;
		      break;
		    };
		  key_size = 1;
		}

	      if (key_cmd == 0x4)
		{
		  *usage_page = get_bytes (report_descriptor, size, data_len, i);
		  usage_page_found = 1;
		}
	      if (key_cmd == 0x8)
		{
		  *usage = get_bytes (report_descriptor, size, data_len, i);
		  usage_found = 1;
		}

	      if (usage_page_found && usage_found)
		return 0;		// success

	      i += data_len + key_size;
	    }

	  return -1;			// failure
	}



	static int
	get_usages (struct hid_device_info *dev, unsigned short *usage_page,
		    unsigned short *usage)
	{
	  int res, desc_size;
	  int ret = -1;
	  struct hidraw_report_descriptor rpt_desc;
	  int handle = open (dev->path, O_RDWR);
	  if (handle > 0)
	    {
	      memset (&rpt_desc, 0, sizeof (rpt_desc));
	      res = ioctl (handle, HIDIOCGRDESCSIZE, &desc_size);
	      if (res >= 0)
		{
		  rpt_desc.size = desc_size;
		  res = ioctl (handle, HIDIOCGRDESC, &rpt_desc);
		  if (res >= 0)
		    {
		      res =
			get_usage (rpt_desc.value, rpt_desc.size, usage_page, usage);
		      if (res >= 0)
			{
			  ret = 0;
			}
		    }
		}
	      close (handle);
	    }
	  return ret;
	}


*/
import "C"
import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

func init() {
	// Initialize the HIDAPI library
	C.hid_init()
}

// Supported returns whether this platform is supported by the HID library or not.
// The goal of this method is to allow programatically handling platforms that do
// not support USB HID and not having to fall back to build constraints.
func Supported() bool {
	return true
}

// Enumerate returns a list of all the HID devices attached to the system which
// match the vendor and product id:
//  - If the vendor id is set to 0 then any vendor matches.
//  - If the product id is set to 0 then any product matches.
//  - If the vendor and product id are both 0, all HID devices are returned.
func Enumerate(vendorID uint16, productID uint16) []DeviceInfo {
	fmt.Println("Trying to enumerate devices with hidraw")
	// Gather all device infos and ensure they are freed before returning
	head := C.hid_enumerate(C.ushort(vendorID), C.ushort(productID))
	if head == nil {
		return nil
	}
	defer C.hid_free_enumeration(head)

	// Iterate the list and retrieve the device details
	var infos []DeviceInfo
	for ; head != nil; head = head.next {

		usage := C.ushort(0)
		usagePage := C.ushort(0)
		ok := C.get_usages(head, &usagePage, &usage)
		if ok != 0 {
			// TODO: figure out what to do if that failed.
		}

		info := DeviceInfo{
			Path:      C.GoString(head.path),
			VendorID:  uint16(head.vendor_id),
			ProductID: uint16(head.product_id),
			Release:   uint16(head.release_number),
			UsagePage: uint16(usagePage),
			Usage:     uint16(usage),
			Interface: int(head.interface_number),
		}
		if head.serial_number != nil {
			info.Serial, _ = wcharTToString(head.serial_number)
		}
		if head.product_string != nil {
			info.Product, _ = wcharTToString(head.product_string)
		}
		if head.manufacturer_string != nil {
			info.Manufacturer, _ = wcharTToString(head.manufacturer_string)
		}
		infos = append(infos, info)
	}
	return infos
}

// Open connects to an HID device by its path name.
func (info DeviceInfo) Open() (*Device, error) {
	path := C.CString(info.Path)
	defer C.free(unsafe.Pointer(path))

	device := C.hid_open_path(path)
	if device == nil {
		return nil, fmt.Errorf("hidapi: failed to open device: %s", C.GoString(path))
	}
	return &Device{
		DeviceInfo: info,
		device:     device,
	}, nil
}

// Device is a live HID USB connected device handle.
type Device struct {
	DeviceInfo // Embed the infos for easier access

	device *C.hid_device // Low level HID device to communicate through
	lock   sync.Mutex
}

// Close releases the HID USB device handle.
func (dev *Device) Close() {
	dev.lock.Lock()
	defer dev.lock.Unlock()

	if dev.device != nil {
		C.hid_close(dev.device)
		dev.device = nil
	}
}

// Write sends an output report to a HID device.
//
// Write will send the data on the first OUT endpoint, if one exists. If it does
// not, it will send the data through the Control Endpoint (Endpoint 0).
func (dev *Device) Write(b []byte) (int, error) {
	// Abort if nothing to write
	if len(b) == 0 {
		return 0, nil
	}
	// Abort if device closed in between
	dev.lock.Lock()
	device := dev.device
	dev.lock.Unlock()

	if device == nil {
		return 0, ErrDeviceClosed
	}
	// Prepend a HID report ID on Windows, other OSes don't need it
	var report []byte
	if runtime.GOOS == "windows" {
		report = append([]byte{0x00}, b...)
	} else {
		report = b
	}
	// Execute the write operation
	written := int(C.hid_write(device, (*C.uchar)(&report[0]), C.size_t(len(report))))
	if written == -1 {
		// If the write failed, verify if closed or other error
		dev.lock.Lock()
		device = dev.device
		dev.lock.Unlock()

		if device == nil {
			return 0, ErrDeviceClosed
		}
		// Device not closed, some other error occurred
		message := C.hid_error(device)
		if message == nil {
			return 0, errors.New("hidapi: unknown failure")
		}
		failure, _ := wcharTToString(message)
		return 0, errors.New("hidapi: " + failure)
	}
	return written, nil
}

// Read retrieves an input report from a HID device.
func (dev *Device) Read(b []byte) (int, error) {
	// Aborth if nothing to read
	if len(b) == 0 {
		return 0, nil
	}
	// Abort if device closed in between
	dev.lock.Lock()
	device := dev.device
	dev.lock.Unlock()

	if device == nil {
		return 0, ErrDeviceClosed
	}
	// Execute the read operation
	read := int(C.hid_read(device, (*C.uchar)(&b[0]), C.size_t(len(b))))
	if read == -1 {
		// If the read failed, verify if closed or other error
		dev.lock.Lock()
		device = dev.device
		dev.lock.Unlock()

		if device == nil {
			return 0, ErrDeviceClosed
		}
		// Device not closed, some other error occurred
		message := C.hid_error(device)
		if message == nil {
			return 0, errors.New("hidapi: unknown failure")
		}
		failure, _ := wcharTToString(message)
		return 0, errors.New("hidapi: " + failure)
	}
	return read, nil
}
