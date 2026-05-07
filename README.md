# ⚡ openbattery

[![CI](https://github.com/przybylku/openbattery/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/przybylku/openbattery/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/przybylku/openbattery)](go.mod)
[![Latest Release](https://img.shields.io/github/v/release/przybylku/openbattery?sort=semver)](https://github.com/przybylku/openbattery/releases)
[![License](https://img.shields.io/github/license/przybylku/openbattery)](LICENSE)

A beautiful, responsive **macOS** terminal battery monitor built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

Watch your battery level, power draw, discharge estimates, and health stats in real time — right from your terminal.

<img width="800" alt="openbattery dashboard" src="https://github.com/przybylku/openbattery/assets/screenshot.png">

## Features

- **Live dashboard** — battery percentage with gradient progress bar, power draw (watts), remaining time, and charge estimates
- **History sparkline** — ▁▂▃▄▅▆▇█ chart of battery level over time with Y-axis labels and time range
- **Smart estimates** — remaining time calculated from your average power consumption (last hour / today)
- **Battery health** — current max capacity vs design capacity, cycle count
- **Two tabs** — `1` for dashboard, `2` for history, `r` to refresh, `q` to quit
- **Responsive layout** — adapts seamlessly to any terminal size
- **Persistent history** — saves up to 500 samples to `~/.openbattery/history.json`
- **Zero config** — just run it, no setup needed

## Installation

```bash
go install github.com/przybylku/openbattery@latest
```

Or clone and build:

```bash
git clone https://github.com/przybylku/openbattery.git
cd openbattery
go build -o openbattery .
./openbattery
```

Requires **macOS** (uses `pmset` and `ioreg` system utilities).

## Data Sources

`openbattery` reads from two macOS system commands:

- **`pmset -g batt`** — battery percentage, charge status, Apple's time estimate
- **`ioreg -rn AppleSmartBattery`** — amperage (mA), voltage (mV), capacities (mAh), cycle count

All data is processed into a clean TUI with color-coded indicators:

| Condition | Color |
|-----------|-------|
| Battery ≥ 50% | Green |
| Battery ≥ 20% | Yellow |
| Battery < 20% | Red |
| Charging | Cyan |
| Charged | Green |

## History

Every 10 seconds, a data point is appended to `~/.openbattery/history.json` (keeps last 500 entries). The history tab renders a sparkline of the last 60 samples with the most recent 10 displayed in a table.

## License

MIT
