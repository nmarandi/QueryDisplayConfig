package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

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
	x int32
	y int32
}

type DISPLAYCONFIG_RATIONAL struct {
	Numerator   uint32
	Denominator uint32
}

type DISPLAYCONFIG_VIDEO_SIGNAL_INFO struct {
	pixelRate        uint64
	hSyncFreq        DISPLAYCONFIG_RATIONAL
	vSyncFreq        DISPLAYCONFIG_RATIONAL
	activeSize       POINTL
	totalSize        POINTL
	videoStandard    uint32
	scanLineOrdering uint32
}

type DISPLAYCONFIG_TARGET_MODE struct {
	targetVideoSignalInfo DISPLAYCONFIG_VIDEO_SIGNAL_INFO
}

type DISPLAYCONFIG_MODE_INFO struct {
	infoType  uint32
	id        uint32
	adapterId LUID
	modeInfo  DISPLAYCONFIG_TARGET_MODE
}

func main() {
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
		if modeInfo[i].modeInfo.targetVideoSignalInfo.vSyncFreq.Denominator != 0 {
			freq = modeInfo[i].modeInfo.targetVideoSignalInfo.vSyncFreq.Numerator / modeInfo[i].modeInfo.targetVideoSignalInfo.vSyncFreq.Denominator
		} else {
			freq = 0
		}
		fmt.Printf("Mode %d: %d x %d @ %d Hz\n", i, modeInfo[i].modeInfo.targetVideoSignalInfo.activeSize.x, modeInfo[i].modeInfo.targetVideoSignalInfo.activeSize.y, freq)
	}
}
