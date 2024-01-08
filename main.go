package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"syscall"
	"unsafe"
)

// #cgo LDFLAGS: -luser32
// #include <windows.h>
// #include <winuser.h>
import "C"

const (
	QDC_ALL_PATHS                  = 0x00000001
	QDC_ONLY_ACTIVE_PATHS          = 0x00000002
	QDC_DATABASE_CURRENT           = 0x00000004
	QDC_VIRTUAL_MODE_AWARE         = 0x00000010
	QDC_INCLUDE_HMD                = 0x00000020
	QDC_VIRTUAL_REFRESH_RATE_AWARE = 0x00000040
)

var (
	user32                      = syscall.NewLazyDLL("user32.dll")
	QueryDisplayConfig          = user32.NewProc("QueryDisplayConfig")
	GetDisplayConfigBufferSizes = user32.NewProc("GetDisplayConfigBufferSizes")
)

type LUID struct {
	lowPart  uint32
	highPart int32
}

type DISPLAYCONFIG_PATH_SOURCE_INFO struct {
	adapterId   LUID
	id          uint32
	modeInfoIdx uint32
	statusFlags uint32
}

type DISPLAYCONFIG_PATH_TARGET_INFO struct {
	adapterId        LUID
	id               uint32
	modeInfoIdx      uint32
	outputTechnology uint32
	rotation         uint32
	scaling          uint32
	refreshRate      DISPLAYCONFIG_RATIONAL
	scanLineOrdering uint32
	targetAvailable  int32
	statusFlags      uint32
}

type DISPLAYCONFIG_PATH_INFO struct {
	sourceInfo DISPLAYCONFIG_PATH_SOURCE_INFO
	targetInfo DISPLAYCONFIG_PATH_TARGET_INFO
	flags      uint32
}

type POINTL struct {
	X int32
	Y int32
}

type DISPLAYCONFIG_RATIONAL struct {
	Numerator   uint32
	Denominator uint32
}

type DISPLAYCONFIG_VIDEO_SIGNAL_INFO struct {
	PixelRate        uint64
	HSyncFreq        DISPLAYCONFIG_RATIONAL
	VSyncFreq        DISPLAYCONFIG_RATIONAL
	ActiveSize       POINTL
	TotalSize        POINTL
	VideoStandard    uint32
	ScanLineOrdering uint32
}

type DISPLAYCONFIG_TARGET_MODE struct {
	TargetVideoSignalInfo DISPLAYCONFIG_VIDEO_SIGNAL_INFO
}

type DISPLAYCONFIG_MODE_INFO struct {
	infoType  uint32
	id        uint32
	adapterId LUID
	modeInfo  DISPLAYCONFIG_TARGET_MODE
}

func queryDisplayConfigCGO() {
	var pathCount C.UINT32
	var modeInfoCount C.UINT32
	C.GetDisplayConfigBufferSizes(C.QDC_ONLY_ACTIVE_PATHS|C.QDC_VIRTUAL_MODE_AWARE, &pathCount, &modeInfoCount)
	fmt.Println("pathCount:", pathCount, "modeInfoCount:", modeInfoCount)
	// create a buffer of the required size for the path and mode info
	pathInfo := make([]C.DISPLAYCONFIG_PATH_INFO, pathCount)
	modeInfo := make([]C.DISPLAYCONFIG_MODE_INFO, modeInfoCount)
	C.QueryDisplayConfig(C.QDC_ONLY_ACTIVE_PATHS|C.QDC_VIRTUAL_MODE_AWARE, &pathCount, &pathInfo[0], &modeInfoCount, &modeInfo[0], nil)
	for i := 0; i < int(pathCount); i++ {
		// print all the paths informations and status flags
		fmt.Printf("Path %d: %d -> %d\n", i, pathInfo[i].sourceInfo.id, pathInfo[i].targetInfo.id)
		fmt.Printf("  Source: AdapterID: %d, ID: %d, ModeInfoIdx: %+v, StatusFlags: %d %d\n",
			pathInfo[i].sourceInfo.adapterId, pathInfo[i].sourceInfo.id, binary.LittleEndian.Uint32(pathInfo[i].sourceInfo.anon0[:]), pathInfo[i].sourceInfo.statusFlags, pathInfo[i].targetInfo.statusFlags)
	}
	for i := 0; i < int(modeInfoCount); i++ {
		freq := uint32(0)
		var targetMode DISPLAYCONFIG_TARGET_MODE
		binary.Read(bytes.NewReader(modeInfo[i].anon0[:]), binary.LittleEndian, &targetMode)
		if targetMode.TargetVideoSignalInfo.VSyncFreq.Denominator != 0 {
			freq = targetMode.TargetVideoSignalInfo.VSyncFreq.Numerator / targetMode.TargetVideoSignalInfo.VSyncFreq.Denominator
		} else {
			freq = 0
		}
		fmt.Printf("Mode %d: %d x %d @ %d Hz\n", i, targetMode.TargetVideoSignalInfo.ActiveSize.X, targetMode.TargetVideoSignalInfo.ActiveSize.Y, freq)
	}
}

func queryDisplayConfigSyscall() {
	pathCount := uint32(0)
	modeInfoCount := uint32(0)
	r, _, err := GetDisplayConfigBufferSizes.Call(QDC_ONLY_ACTIVE_PATHS|QDC_VIRTUAL_MODE_AWARE, uintptr(unsafe.Pointer(&pathCount)), uintptr(unsafe.Pointer(&modeInfoCount)))
	if r != 0 {
		fmt.Println("GetDisplayConfigBufferSizes failed:", err)
	}
	// create a buffer of the required size for the path and mode info
	pathInfo := make([]DISPLAYCONFIG_PATH_INFO, pathCount)
	modeInfo := make([]DISPLAYCONFIG_MODE_INFO, modeInfoCount)
	r, _, err = QueryDisplayConfig.Call(QDC_ONLY_ACTIVE_PATHS|QDC_VIRTUAL_MODE_AWARE, uintptr(unsafe.Pointer(&pathCount)),
		uintptr(unsafe.Pointer(&pathInfo[0])), uintptr(unsafe.Pointer(&modeInfoCount)), uintptr(unsafe.Pointer(&modeInfo[0])), 0)
	if r != 0 {
		fmt.Println("QueryDisplayConfig failed:", err)
	}
	for i := uint32(0); i < pathCount; i++ {

		// print all the paths informations and status flags
		fmt.Printf("Path %d: %d -> %d\n", i, pathInfo[i].sourceInfo.id, pathInfo[i].targetInfo.id)
		fmt.Printf("  Source: AdapterID: %d, ID: %d, ModeInfoIdx: %d, StatusFlags: %d %d\n",
			pathInfo[i].sourceInfo.adapterId, pathInfo[i].sourceInfo.id, pathInfo[i].sourceInfo.modeInfoIdx, pathInfo[i].sourceInfo.statusFlags, pathInfo[i].targetInfo.statusFlags)
	}
	for i := uint32(0); i < modeInfoCount; i++ {
		freq := uint32(0)
		if modeInfo[i].modeInfo.TargetVideoSignalInfo.VSyncFreq.Denominator != 0 {
			freq = modeInfo[i].modeInfo.TargetVideoSignalInfo.VSyncFreq.Numerator / modeInfo[i].modeInfo.TargetVideoSignalInfo.VSyncFreq.Denominator
		} else {
			freq = 0
		}
		fmt.Printf("Mode %d: %d x %d @ %d Hz\n", i, modeInfo[i].modeInfo.TargetVideoSignalInfo.ActiveSize.X, modeInfo[i].modeInfo.TargetVideoSignalInfo.ActiveSize.Y, freq)
	}
}

func main() {
	fmt.Println("-----QueryDisplayConfig using CGO-----")
	queryDisplayConfigCGO()
	fmt.Println("-----QueryDisplayConfig using syscall-----")
	queryDisplayConfigSyscall()
}
