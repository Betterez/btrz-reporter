package reporter

import (
	"bufio"
	"errors"
	"io"
	"os"
	"regexp"
	"strconv"
)

const (
	versionFileName = "/etc/os-release"
	memoryFileName  = "/proc/meminfo"
)

// MemoryUsage - shows the usage in linux
type MemoryUsage struct {
	availableMemory int64
	freeMemory      int64
	totalMemory     int64
	osVersion       float64
}

// GetTotalMemory - return total memory
func (info *MemoryUsage) GetTotalMemory() float64 {
	return float64(info.totalMemory)
}

// GetUsedMemory - return used memory
func (info *MemoryUsage) GetUsedMemory() float64 {
	return info.GetTotalMemory() - info.GetFreeMemory()
}

// GetFreeMemory - returns the available memory (not the actual free one)
func (info *MemoryUsage) GetFreeMemory() float64 {
	if info.osVersion > 16 {
		return float64(info.availableMemory)
	}
	return float64(info.freeMemory)
}

// GetUsedMemoryPercentage - return the used memory percentage
func (info *MemoryUsage) GetUsedMemoryPercentage() float64 {
	return float64(info.GetUsedMemory() * 100 / info.GetTotalMemory())
}

// GetOSVersion - returns the detected os version
func (info *MemoryUsage) GetOSVersion() float64 {
	return info.osVersion
}

func loadOSVersion() (float64, error) {
	osVersion := 0.0
	if _, err := os.Stat(versionFileName); err != nil {
		return 0, errors.New("Unknown os")
	}
	exp, err := regexp.Compile(`VERSION_ID=\"([\d]+.[\d]+)\"`)
	if err != nil {
		return 0, err
	}
	data, err := os.Open(versionFileName)
	if err != nil {
		return osVersion, err
	}
	reader := bufio.NewReader(data)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return osVersion, err
		}
		results := exp.FindAllStringSubmatch(line, -1)
		if results != nil {
			osVersion, err = strconv.ParseFloat(results[0][1], 64)
			if err != nil {
				return osVersion, err
			}
			break
		}
	}
	return osVersion, nil
}

// LoadMemoryValue - load memory values
func LoadMemoryValue() (*MemoryUsage, error) {
	osVersion, err := loadOSVersion()
	if err != nil {
		return nil, err
	}
	result := &MemoryUsage{osVersion: osVersion}
	memFileLocation := "/proc/meminfo"
	exp, err := regexp.Compile(`(Mem[\w]+:)[\s]+([\d]+)`)
	if err != nil {
		return nil, err
	}
	data, err := os.Open(memFileLocation)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(data)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		results := exp.FindAllStringSubmatch(line, -1)
		if results != nil {
			if results[0][1] == "MemTotal:" {
				result.totalMemory, err = strconv.ParseInt(results[0][2], 10, 64)
				if err != nil {
					return nil, err
				}
			}
			if results[0][1] == "MemFree:" {
				result.freeMemory, err = strconv.ParseInt(results[0][2], 10, 64)
				if err != nil {
					return nil, err
				}
			}
			if results[0][1] == "MemAvailable:" {
				result.availableMemory, err = strconv.ParseInt(results[0][2], 10, 64)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return result, nil
}
