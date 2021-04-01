// +build linux

package fscommon

import (
	"bufio"
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/opencontainers/runc/libcontainer/cgroups"
	"golang.org/x/sys/unix"
)

// parseRdmaKV: parses raw string to RdmaEntry.
func parseRdmaKV(raw string, entry *cgroups.RdmaEntry) error {
	var value uint64
	var err error

	parts := strings.SplitN(raw, "=", 3)

	if len(parts) == 2 {
		if parts[1] == "max" {
			value = math.MaxUint64
		} else {
			value, err = strconv.ParseUint(parts[1], 10, 32)
			if err != nil {
				return err
			}
		}
		if parts[0] == "hca_handle" {
			entry.HcaHandles = uint32(value)
		} else if parts[0] == "hca_object" {
			entry.HcaObjects = uint32(value)
		}
	} else {
		return errors.New("Unable to parse RDMA entry")
	}

	return nil
}

// ReadRDMAEntries: Reads and converts array of rawstrings to RdmaEntries from file
func ReadRDMAEntries(dir, file string) ([]cgroups.RdmaEntry, error) {
	rdmaEntries := make([]cgroups.RdmaEntry, 0)
	fd, err := cgroups.OpenFile(dir, file, unix.O_RDONLY)
	if err != nil {
		return nil, err
	}
	defer fd.Close() //nolint:errorlint
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		strEntry := scanner.Text()
		parts := strings.SplitN(strEntry, " ", 4)
		if len(parts) == 3 {
			entry := new(cgroups.RdmaEntry)
			entry.Device = parts[0]
			err = parseRdmaKV(parts[1], entry)
			if err != nil {
				continue
			}
			err = parseRdmaKV(parts[2], entry)
			if err != nil {
				continue
			}

			rdmaEntries = append(rdmaEntries, *entry)
		}
	}
	return rdmaEntries, scanner.Err()
}
