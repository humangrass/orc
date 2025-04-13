package entities

import (
	linuxproc "github.com/c9s/goprocinfo/linux"
	"log"
)

type Stats struct {
	MemStats  *linuxproc.MemInfo
	DiskStats *linuxproc.Disk
	CPUStats  *linuxproc.CPUStat
	LoadStats *linuxproc.LoadAvg
	TaskCount int
}

func (s *Stats) MemTotalKb() uint64 {
	return s.MemStats.MemTotal
}

func (s *Stats) MemAvailableKb() uint64 {
	return s.MemStats.MemAvailable
}

func (s *Stats) MemUsedKb() uint64 {
	return s.MemTotalKb() - s.MemAvailableKb()
}

func (s *Stats) MemUsedPercent() uint64 {
	return s.MemAvailableKb() / s.MemTotalKb()
}

func (s *Stats) DiskTotal() uint64 {
	return s.DiskStats.All
}

func (s *Stats) DiskFree() uint64 {
	return s.DiskStats.Free
}

func (s *Stats) DiskUsed() uint64 {
	return s.DiskStats.Used
}

func (s *Stats) CPUsage() float64 {
	idle := s.CPUStats.Idle + s.CPUStats.IOWait
	nonIdle := s.CPUStats.User + s.CPUStats.Nice + s.CPUStats.System + s.CPUStats.IRQ + s.CPUStats.SoftIRQ + s.CPUStats.Steal
	total := idle + nonIdle

	if total == 0 {
		return 0.0
	}
	return (float64(total) / float64(idle)) / float64(total)
}

func GetStats() *Stats {
	return &Stats{
		MemStats:  GetMemoryInfo(),
		DiskStats: GetDiskInfo(),
		CPUStats:  GetCPUStats(),
		LoadStats: GetLoadAvg(),
	}
}

func GetMemoryInfo() *linuxproc.MemInfo {
	memStats, err := linuxproc.ReadMemInfo("/proc/meminfo")
	if err != nil {
		log.Println("Error reading /proc/meminfo:", err)
		return nil
	}

	return memStats
}

func GetDiskInfo() *linuxproc.Disk {
	diskStats, err := linuxproc.ReadDisk("/")
	if err != nil {
		log.Println("Error reading from '/':", err)
		return nil
	}

	return diskStats
}

func GetCPUStats() *linuxproc.CPUStat {
	stats, err := linuxproc.ReadStat("/proc/stat")
	if err != nil {
		log.Println("Error reading /proc/stat:", err)
		return nil
	}

	return &stats.CPUStatAll
}

func GetLoadAvg() *linuxproc.LoadAvg {
	loadAvg, err := linuxproc.ReadLoadAvg("/proc/loadavg")
	if err != nil {
		log.Println("Error reading /proc/loadavg:", err)
		return nil
	}

	return loadAvg
}
