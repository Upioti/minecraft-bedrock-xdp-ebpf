package analytics

import (
	xdp "bedrock-xdp/xdp_utils"

	"github.com/cilium/ebpf"
)

//Counter reset logic and Map Lookups should probably be reworked but what do you expect, this is open source.

var (
	udpPacketMap   *ebpf.Map // udp_pass_pps
	otherPacketMap *ebpf.Map // other_pass_pps
)

func StartPPS(Collection *ebpf.Collection) {
	// Fetch BPF maps generated by the kernel program
	udpPacketMap = xdp.GetMap("udp_pass_pps", Collection)
	otherPacketMap = xdp.GetMap("other_pass_pps", Collection)
}

// ResetPassPPS resets pass packet counters after stats are read.
func ResetPassPPS() {
	var resetCount uint64 = 0
	var totalKey uint32 = 0

	// Reset totals (index 0) each second
	if udpPacketMap != nil {
		udpPacketMap.Update(&totalKey, &resetCount, ebpf.UpdateAny)
	}
	if otherPacketMap != nil {
		otherPacketMap.Update(&totalKey, &resetCount, ebpf.UpdateAny)
	}

	// Clear per-IP buckets if any
	resetPacketCountMap(udpPacketMap)
	resetPacketCountMap(otherPacketMap)
}

// GetTotalPPS returns aggregated PPS for the requested class.
// "udp" -> UDP, "other" -> all other traffic, default -> sum of both.
func GetTotalPPS(protocol string) uint64 {
	switch protocol {
	case "udp":
		return getMapTotalCount(udpPacketMap) / uint64(StatIntervalSec)
	case "other":
		return getMapTotalCount(otherPacketMap) / uint64(StatIntervalSec)
	default:
		return (getMapTotalCount(udpPacketMap) + getMapTotalCount(otherPacketMap)) / uint64(StatIntervalSec)
	}
}
