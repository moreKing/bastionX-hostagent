package hostinfo

import (
	"fmt"
	"testing"
)

func TestGetSystemInfo(t *testing.T) {
	systemInfo, err := GetSystemInfo()
	if err != nil {
		t.Errorf("GetSystemInfo() error = %v", err)
		return
	}
	fmt.Printf("%#v\n", systemInfo)
}
