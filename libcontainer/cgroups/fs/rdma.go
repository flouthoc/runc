// +build linux

package fs

import (
	"strconv"
	"strings"

	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/opencontainers/runc/libcontainer/cgroups/fscommon"
	"github.com/opencontainers/runc/libcontainer/configs"
)

type RdmaGroup struct {
}

func (s *RdmaGroup) Name() string {
	return "rdma"
}

func (s *RdmaGroup) Apply(path string, d *cgroupData) error {
	return join(path, d.pid)
}

func createCmdString(device string, limits configs.LinuxRdma) string {
	cmdString := device
	if limits.HcaHandles != nil {
		cmdString += " hca_handle=" + strconv.FormatUint(uint64(*limits.HcaHandles), 10)
	}
	if limits.HcaObjects != nil {
		cmdString += " hca_object=" + strconv.FormatUint(uint64(*limits.HcaObjects), 10)
	}
	return cmdString
}

func (s *RdmaGroup) Set(path string, r *configs.Resources) error {
	for device, limits := range r.Rdma {
		if err := fscommon.WriteFile(path, "rdma.max", createCmdString(device, limits)); err != nil {
			return err
		}
	}
	return nil
}

func (s *RdmaGroup) GetStats(path string, stats *cgroups.Stats) error {
	if !cgroups.PathExists(path) {
		return nil
	}
	currentData, err := fscommon.ReadFile(path, "rdma.current")
	if err != nil {
		return err
	}
	currentPerDevices := strings.Split(currentData, "\n")
	maxData, err := fscommon.ReadFile(path, "rdma.max")
	if err != nil {
		return err
	}
	maxPerDevices := strings.Split(maxData, "\n")
	// If device got removed between reading two files, ignore returning
	// stats.
	if len(currentPerDevices) != len(maxPerDevices) {
		return nil
	}
	currentEntries := fscommon.ConvertRdmaEntry(currentPerDevices)
	maxEntries := fscommon.ConvertRdmaEntry(maxPerDevices)

	stats.RdmaStats = cgroups.RdmaStats{
		RdmaLimit:   maxEntries,
		RdmaCurrent: currentEntries,
	}

	return nil
}
