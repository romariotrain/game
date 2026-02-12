package sim

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// RunCLI checks os.Args for --simulate and runs the simulator.
// Returns true if --simulate was found (and handled), false otherwise.
func RunCLI() bool {
	args := os.Args[1:]
	if len(args) == 0 {
		return false
	}

	if args[0] != "--simulate" {
		return false
	}

	seed := time.Now().UnixNano()

	// Parse optional seed/run count
	if len(args) >= 2 {
		if n, err := strconv.ParseInt(args[1], 10, 64); err == nil {
			seed = n
		}
	}

	fmt.Println(FullReport(seed))
	return true
}

// RunCLIAutoTune runs binary-search enemy auto-tuning.
// Usage:
//
//	--simulate-tune [seed] [runsPerEval] [iterations]
//	--simulate-autotune [seed] [runsPerEval] [iterations]
//
// The first prints only tuning summary.
// The second prints tuning summary + full simulation report on tuned enemies.
func RunCLIAutoTune() bool {
	args := os.Args[1:]
	if len(args) == 0 {
		return false
	}
	if args[0] != "--simulate-tune" && args[0] != "--simulate-autotune" {
		return false
	}

	fullReport := args[0] == "--simulate-autotune"

	seed := time.Now().UnixNano()
	if len(args) >= 2 {
		if n, err := strconv.ParseInt(args[1], 10, 64); err == nil {
			seed = n
		}
	}

	opts := DefaultAutoTuneOptions(seed)
	if len(args) >= 3 {
		if n, err := strconv.Atoi(args[2]); err == nil && n > 0 {
			opts.RunsPerEval = n
		}
	}
	if len(args) >= 4 {
		if n, err := strconv.Atoi(args[3]); err == nil && n > 0 {
			opts.Iterations = n
		}
	}

	base := GetPresetEnemies()
	tuned, results := AutoTuneEnemies(base, opts)
	fmt.Print(AutoTuneSummary(results))

	if fullReport {
		fmt.Println(FullReportWithEnemies(seed, tuned))
	}
	return true
}

// RunCLICompact runs a compact simulation for a specific archetype.
// Usage: --simulate-compact <archetype-index> <days> <runs>
func RunCLICompact() bool {
	args := os.Args[1:]
	if len(args) == 0 || args[0] != "--simulate-compact" {
		return false
	}

	archIdx := 0 // default: Balanced
	days := 90
	runs := 10

	if len(args) >= 2 {
		if n, err := strconv.Atoi(args[1]); err == nil && n >= 0 && n < 5 {
			archIdx = n
		}
	}
	if len(args) >= 3 {
		if n, err := strconv.Atoi(args[2]); err == nil && n > 0 {
			days = n
		}
	}
	if len(args) >= 4 {
		if n, err := strconv.Atoi(args[3]); err == nil && n > 0 {
			runs = n
		}
	}

	archetypes := DefaultArchetypes()
	cfg := SimConfig{
		Days:      days,
		Seed:      time.Now().UnixNano(),
		Archetype: archetypes[archIdx],
		Enemies:   GetPresetEnemies(),
	}

	fmt.Println(CompactTable(cfg, runs))
	return true
}
