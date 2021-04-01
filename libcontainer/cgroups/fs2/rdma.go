// +build linux

package fs2

import (
	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/opencontainers/runc/libcontainer/cgroups/fscommon"
	"github.com/opencontainers/runc/libcontainer/configs"
)

func setRdma(path string, r *configs.Resources) error {
	return fscommon.RdmaSet(path, r)
}

func statRdma(path string, stats *cgroups.Stats) error {
	return fscommon.RdmaGetStats(path, stats)
}
