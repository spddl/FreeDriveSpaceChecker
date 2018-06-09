// +build windows

package disk

import (
	"os"
	"syscall"
	"unsafe"
)

// Space returns total and free bytes available in a directory, e.g. `C:\`.
// It returns free space available to the user (including quota limitations),
// so it can be lower than the free space of the disk.
// https://github.com/StalkR/goircbot/blob/master/lib/disk/space_windows.go
func Space(path string) (total, free uint64, err error) {
	kernel32, err := syscall.LoadLibrary("Kernel32.dll")
	if err != nil {
		return
	}
	defer syscall.FreeLibrary(kernel32)
	GetDiskFreeSpaceEx, err := syscall.GetProcAddress(syscall.Handle(kernel32), "GetDiskFreeSpaceExW")
	if err != nil {
		return
	}
	lpFreeBytesAvailable := uint64(0)
	lpTotalNumberOfBytes := uint64(0)
	lpTotalNumberOfFreeBytes := uint64(0)
	r1, _, e1 := syscall.Syscall6(uintptr(GetDiskFreeSpaceEx), 4,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(path))),
		uintptr(unsafe.Pointer(&lpFreeBytesAvailable)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfBytes)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfFreeBytes)), 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
		return
	}
	total = uint64(lpTotalNumberOfBytes)
	free = uint64(lpFreeBytesAvailable)
	return
}

// Getdrives gibt alle Laufwerke zurück
// https://stackoverflow.com/a/23129242
func Getdrives() (r []string) {
	for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		_, err := os.Open(string(drive) + ":\\")
		if err == nil {
			r = append(r, string(drive))
		}
	}
	return
}
