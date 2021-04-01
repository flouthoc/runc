// +build linux

package fscommon

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/opencontainers/runc/libcontainer/cgroups"
)

var (
	ErrNotValidFormat = errors.New("line is not a valid key value format")

	// Deprecated: use cgroups.OpenFile instead.
	OpenFile = cgroups.OpenFile
	// Deprecated: use cgroups.ReadFile instead.
	ReadFile = cgroups.ReadFile
	// Deprecated: use cgroups.WriteFile instead.
	WriteFile = cgroups.WriteFile
)

// ParseUint converts a string to an uint64 integer.
// Negative values are returned at zero as, due to kernel bugs,
// some of the memory cgroup stats can be negative.
func ParseUint(s string, base, bitSize int) (uint64, error) {
	value, err := strconv.ParseUint(s, base, bitSize)
	if err != nil {
		intValue, intErr := strconv.ParseInt(s, base, bitSize)
		// 1. Handle negative values greater than MinInt64 (and)
		// 2. Handle negative values lesser than MinInt64
		if intErr == nil && intValue < 0 {
			return 0, nil
		} else if intErr != nil && intErr.(*strconv.NumError).Err == strconv.ErrRange && intValue < 0 {
			return 0, nil
		}

		return value, err
	}

	return value, nil
}

// ParseKeyValue parses a space-separated "name value" kind of cgroup
// parameter and returns its key as a string, and its value as uint64
// (ParseUint is used to convert the value). For example,
// "io_service_bytes 1234" will be returned as "io_service_bytes", 1234.
func ParseKeyValue(t string) (string, uint64, error) {
	parts := strings.SplitN(t, " ", 3)
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("line %q is not in key value format", t)
	}

	value, err := ParseUint(parts[1], 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("unable to convert to uint64: %v", err)
	}

	return parts[0], value, nil
}

// GetValueByKey reads a key-value pairs from the specified cgroup file,
// and returns a value of the specified key. ParseUint is used for value
// conversion.
func GetValueByKey(path, file, key string) (uint64, error) {
	content, err := cgroups.ReadFile(path, file)
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		arr := strings.Split(line, " ")
		if len(arr) == 2 && arr[0] == key {
			return ParseUint(arr[1], 10, 64)
		}
	}

	return 0, nil
}

// GetCgroupParamUint reads a single uint64 value from the specified cgroup file.
// If the value read is "max", the math.MaxUint64 is returned.
func GetCgroupParamUint(path, file string) (uint64, error) {
	contents, err := GetCgroupParamString(path, file)
	if err != nil {
		return 0, err
	}
	contents = strings.TrimSpace(contents)
	if contents == "max" {
		return math.MaxUint64, nil
	}

	res, err := ParseUint(contents, 10, 64)
	if err != nil {
		return res, fmt.Errorf("unable to parse file %q", path+"/"+file)
	}
	return res, nil
}

// GetCgroupParamInt reads a single int64 value from specified cgroup file.
// If the value read is "max", the math.MaxInt64 is returned.
func GetCgroupParamInt(path, file string) (int64, error) {
	contents, err := cgroups.ReadFile(path, file)
	if err != nil {
		return 0, err
	}
	contents = strings.TrimSpace(contents)
	if contents == "max" {
		return math.MaxInt64, nil
	}

	res, err := strconv.ParseInt(contents, 10, 64)
	if err != nil {
		return res, fmt.Errorf("unable to parse %q as a int from Cgroup file %q", contents, path+"/"+file)
	}
	return res, nil
}

// GetCgroupParamString reads a string from the specified cgroup file.
func GetCgroupParamString(path, file string) (string, error) {
	contents, err := cgroups.ReadFile(path, file)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(contents), nil
}

// ParseRdmaKV: parse raw string to RdmaEntry.
func parseRdmaKV(raw string, entry *cgroups.RdmaEntry) {
	var value uint64
	var err error

	parts := strings.SplitN(raw, "=", 3)
	if len(parts) == 2 {
		if parts[1] == "max" {
			value = math.MaxUint32
		} else {
			value, err = strconv.ParseUint(parts[1], 10, 32)
			if err != nil {
				return
			}
		}
		if parts[0] == "hca_handle" {
			entry.HcaHandles = uint32(value)
		} else if parts[0] == "hca_object" {
			entry.HcaObjects = uint32(value)
		}
	}
}

// ConvertRdmaEntry: Converts array of rawstrings to RdmaEntries
func ConvertRdmaEntry(strEntries []string) []cgroups.RdmaEntry {
	rdmaEntries := make([]cgroups.RdmaEntry, len(strEntries))
	for i := range strEntries {
		parts := strings.SplitN(strEntries[i], " ", 4)
		if len(parts) == 3 {
			entry := new(cgroups.RdmaEntry)
			entry.Device = parts[0]
			parseRdmaKV(parts[1], entry)
			parseRdmaKV(parts[2], entry)

			rdmaEntries = append(rdmaEntries, *entry)
		}
	}
	return rdmaEntries
}
