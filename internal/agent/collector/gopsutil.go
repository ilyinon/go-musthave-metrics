package collector

import (
	"fmt"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// Gopsutil collects system metrics using gopsutil library.
// It returns memory statistics and per-CPU utilization values.
func Gopsutil() model.RuntimeMetrics {
	g := model.RuntimeMetrics{}

	if vm, err := mem.VirtualMemory(); err == nil {
		g["TotalMemory"] = float64(vm.Total)
		g["FreeMemory"] = float64(vm.Free)
	}

	if cpuPercents, err := cpu.Percent(0, true); err == nil {
		for i, v := range cpuPercents {
			g[fmt.Sprintf("CPUutilization%d", i+1)] = v
		}
	}

	return g
}
