package hostinfo

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func GetMemory() (*Memory, error) {

	output, err := exec.Command("awk", "/^MemTotal:/{t=$2} /^MemFree:/{f=$2} /^Buffers:/{b=$2} /^Cached:/{c=$2} /^SwapTotal:/{st=$2} /^SwapFree:/{sf=$2} /^SwapCached:/{sc=$2} END{cache=b+c; used=t-f-cache; su=st-sf; print t; print used; print f; print cache; print st; print su; print sc}", "/proc/meminfo").Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) != 7 {
		fmt.Println("GetMemInfo", "解析内存信息失败", lines)
		return nil, fmt.Errorf("解析内存信息失败")
	}

	// 字符串转数字
	total, err := strconv.ParseInt(strings.TrimSpace(lines[0]), 10, 64)
	if err != nil {
		fmt.Println("GetMemInfo", "解析内存total信息失败", lines[0], err)
		return nil, fmt.Errorf("解析内存total信息失败")
	}
	used, err := strconv.ParseInt(strings.TrimSpace(lines[1]), 10, 64)
	if err != nil {
		fmt.Println("GetMemInfo", "解析内存used信息失败", lines[1], err)
		return nil, fmt.Errorf("解析内存used信息失败")
	}
	free, err := strconv.ParseInt(strings.TrimSpace(lines[2]), 10, 64)
	if err != nil {
		fmt.Println("GetMemInfo", "解析内存free信息失败", lines[2], err)
		return nil, fmt.Errorf("解析内存free信息失败")
	}
	cache, err := strconv.ParseInt(strings.TrimSpace(lines[3]), 10, 64)
	if err != nil {
		fmt.Println("GetMemInfo", "解析内存cache信息失败", lines[3], err)
		return nil, fmt.Errorf("解析内存cache信息失败")
	}

	swapTotal, err := strconv.ParseInt(strings.TrimSpace(lines[4]), 10, 64)
	if err != nil {
		fmt.Println("GetMemInfo", "解析swapTotal信息失败", lines[4], err)
		return nil, fmt.Errorf("解析swapTotal信息失败")
	}
	swapUsed, err := strconv.ParseInt(strings.TrimSpace(lines[5]), 10, 64)
	if err != nil {
		fmt.Println("GetMemInfo", "解析swapFree信息失败", lines[5], err)
		return nil, fmt.Errorf("解析swapFree信息失败")
	}
	swapCache, err := strconv.ParseInt(strings.TrimSpace(lines[6]), 10, 64)
	if err != nil {
		fmt.Println("GetMemInfo", "解析swapUsed信息失败", lines[6], err)
		return nil, fmt.Errorf("解析swapUsed信息失败")
	}

	memInfo := Memory{
		Total:     total * 1024,
		Used:      used * 1024,
		Free:      free * 1024,
		Cache:     cache * 1024,
		SwapTotal: swapTotal * 1024,
		SwapFree:  (swapTotal - swapUsed) * 1024,
		SwapUsed:  swapUsed * 1024,
		SwapCache: swapCache * 1024,
	}

	return &memInfo, nil
}
