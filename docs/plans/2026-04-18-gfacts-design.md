# gfacts Design Document

**Date:** 2026-04-18
**Status:** Draft

## Motivation

Port the core capabilities of Puppet Facter to Go as a library (`gfacts`) for embedding in other Go infrastructure projects. The goal is not 1:1 API compatibility with Ruby Facter, but equivalent fact coverage for Linux and macOS with a clean Go-native API.

Primary use case: `import "github.com/<owner>/gfacts"` in Go infrastructure tooling to gather hardware, software, and capability facts about a node.

## Scope

### In scope

- Linux and macOS support
- Pure Go (no CGO), exec/shell as last resort
- Core fact categories: os, kernel, networking, processors, memory, hardware, disks, mountpoints, uptime, cloud, virtual, ssh
- External facts: static files (JSON, TXT) and executables in `/etc/gfacts/gfacts.d/`
- Programmatic fact registration
- CLI binary (`gfacts`)

### Out of scope

- Windows, AIX, Solaris, BSD
- FFI / CGO bindings
- Identity and timezone facts
- YAML or HOCON output/input
- Ruby custom fact compatibility
- Puppet integration

## Public API

```go
package gfacts

// Gather returns all facts as a flat map with dotted keys.
// Example keys: "os.name", "networking.interfaces.eth0.ip"
func Gather() (map[string]any, error)

// GatherWithOptions returns all facts using the provided options.
func GatherWithOptions(opts Options) (map[string]any, error)

// Value returns a single fact by dotted key.
// Prefix lookups return a sub-map (e.g., Value("os") returns all os.* facts).
func Value(key string) (any, error)

// Register adds a static programmatic fact.
func Register(key string, value any)

// RegisterFunc adds a dynamically resolved programmatic fact.
func RegisterFunc(key string, fn FactFunc)

type FactFunc func() (any, error)

type Options struct {
    ExternalDirs []string      // Directories to scan for external facts.
                               // Default: ["/etc/gfacts/gfacts.d"]
    NoExternal   bool          // Skip all external facts.
    NoCustomExec bool          // Load static files but skip executables.
    ExecTimeout  time.Duration // Timeout for executable facts. Default: 30s.
}
```

### Fact key conventions

- All keys are lowercase, dotted: `os.name`, `memory.system.total`
- Dynamic segments allowed: `networking.interfaces.eth0.ip`, `disks.sda.size`
- `Value("prefix")` returns a `map[string]any` sub-map for all facts under that prefix
- `Value("os.name")` returns the leaf value directly

## Project Structure

```
gfacts/
├── gfacts.go              # Public API
├── registry.go            # Fact registry, resolution orchestration
├── facts/
│   ├── collector.go       # Collector interface
│   ├── os_linux.go        # os.* facts (Linux)
│   ├── os_darwin.go       # os.* facts (macOS)
│   ├── kernel_linux.go
│   ├── kernel_darwin.go
│   ├── networking_linux.go
│   ├── networking_darwin.go
│   ├── processors_linux.go
│   ├── processors_darwin.go
│   ├── memory_linux.go
│   ├── memory_darwin.go
│   ├── hardware_linux.go
│   ├── hardware_darwin.go
│   ├── disks_linux.go
│   ├── disks_darwin.go
│   ├── mountpoints_linux.go
│   ├── mountpoints_darwin.go
│   ├── uptime_linux.go
│   ├── uptime_darwin.go
│   ├── cloud.go           # Cross-platform (HTTP metadata)
│   ├── virtual_linux.go
│   ├── virtual_darwin.go
│   ├── ssh.go             # Cross-platform (read key files)
│   └── platform.go        # Shared helpers
├── external/
│   ├── loader.go          # Directory scanner, dispatcher
│   ├── text.go            # key=value parser
│   ├── json.go            # JSON flattener
│   └── exec.go            # Executable runner
├── cmd/
│   └── gfacts/
│       └── main.go        # CLI binary
└── go.mod
```

Platform-specific code uses Go build tags (`//go:build linux`, `//go:build darwin`).

## Fact Collector Interface

```go
// facts/collector.go
type Collector interface {
    Collect() (map[string]any, error)
}
```

Each collector returns a flat `map[string]any` with dotted keys prefixed by its category (e.g., `os.name`, `memory.system.total`). The registry merges all collector results into a single map.

All collectors run in parallel via `sync.WaitGroup`. Errors are logged but do not fail the overall `Gather()` call — best-effort resolution, return what succeeded.

## Data Sources by Category

### Linux

| Category | Primary data source |
|----------|-------------------|
| os | `/etc/os-release`, `uname` syscall |
| kernel | `uname` syscall |
| networking | `/sys/class/net/`, `net.Interfaces()`, `/proc/net/route`, `/etc/resolv.conf` |
| processors | `/proc/cpuinfo` |
| memory | `/proc/meminfo` |
| hardware | `/sys/class/dmi/id/*` |
| disks | `/sys/block/`, `/proc/partitions` |
| mountpoints | `/proc/mounts`, `syscall.Statfs` |
| uptime | `/proc/uptime` |
| cloud | `/sys/class/dmi/id/` for hints, HTTP to `169.254.169.254` |
| virtual | `/sys/class/dmi/id/product_name`, `/proc/cpuinfo` flags, cgroup files |
| ssh | `/etc/ssh/ssh_host_*_key.pub` |

### macOS

| Category | Primary data source |
|----------|-------------------|
| os | `uname` syscall, `sysctl kern.osproductversion` |
| kernel | `uname` syscall |
| networking | `net.Interfaces()`, route syscall messages, `/etc/resolv.conf` |
| processors | `sysctl hw.ncpu`, `machdep.cpu.*` |
| memory | `sysctl hw.memsize`, `vm.swapusage` |
| hardware | `sysctl hw.model`, `system_profiler`, `ioreg` (cached with 30-day TTL) |
| disks | `syscall.Statfs` for volumes; `diskutil list -plist` via exec for physical disks |
| mountpoints | `syscall.Statfs` |
| uptime | `sysctl kern.boottime` (raw bytes, manual struct unpacking) |
| cloud | HTTP to `169.254.169.254` |
| virtual | `sysctl kern.hv_vmm_present`, `hw.model` |
| ssh | `/etc/ssh/ssh_host_*_key.pub` |

### Known exec requirements

These are cases where shelling out is necessary because no pure Go alternative exists:

- **macOS disks**: `diskutil list -plist` for physical disk enumeration
- **macOS hardware**: `system_profiler SPHardwareDataType` and `ioreg -r -c IOPlatformDevice` for hardware profile (model name, description, serial, etc.) — cached at `~/.cache/gfacts/hardware_profile.json` with a 30-day TTL

## External Facts

### Directory

Default: `/etc/gfacts/gfacts.d/`

Configurable via `Options.ExternalDirs`. Multiple directories supported; all are scanned.

### Static files

**`.txt`** — one `key=value` per line. Blank lines and `#` comment lines are ignored.

```
datacenter=us-east-1
environment=production
```

**`.json`** — nested JSON is flattened to dotted keys.

```json
{"app": {"version": "1.2.3", "tier": "frontend"}}
```

Produces: `app.version = "1.2.3"`, `app.tier = "frontend"`.

### Executable facts

- Any file with the execute bit set
- Run with configurable timeout (default 30s)
- Stdout parsed as JSON first; falls back to key=value line parsing
- Non-zero exit code: fact is skipped, logged at **warn** level
- Stderr: logged at **warn** level

### Precedence (highest wins)

1. Programmatic facts (`Register` / `RegisterFunc`)
2. External executable facts
3. External static facts (JSON, then TXT)
4. Core built-in facts

## CLI

```
$ gfacts                        # All facts, JSON (default)
$ gfacts os                     # All os.* facts, JSON (prefix match)
$ gfacts os.name                # Bare value, no JSON wrapper
$ gfacts os networking.ip       # Filtered JSON with both
$ gfacts --plain                # All facts, key=value format
$ gfacts --plain os.name        # Bare value, no key
$ gfacts --debug                # Debug logging to stderr
$ gfacts --no-external          # Skip external facts
$ gfacts --external-dir /path   # Additional external directory
```

**Output formats:**
- JSON (default)
- Plain key=value (`--plain`)

**Exit codes:**
- `0` — success
- `1` — general error
- `2` — requested fact not found

## Known Edge Cases & Difficulties

### 1. macOS sysctl struct values

`syscall.SysctlRaw` returns raw bytes for struct-typed sysctl values (e.g., `kern.boottime` is a `timeval`). Requires manual byte unpacking with awareness of endianness and struct padding. Use `encoding/binary` with `binary.LittleEndian` (both supported Mac architectures are little-endian).

### 2. Primary IP detection

Both `net.Interfaces()` and the resolver give you all IPs including Docker bridges, veth pairs, loopback. Determining `networking.ip` (the primary IP) requires finding the default route:
- Linux: parse `/proc/net/route` for the `0.0.0.0` destination
- macOS: parse route socket messages or use `net.InterfaceAddrs()` heuristics

### 3. Cloud detection timeouts

HTTP requests to cloud metadata endpoints (`169.254.169.254`) must have short timeouts (1-2 seconds). On non-cloud instances these will always time out. Run cloud detection in parallel with other collectors so it doesn't block overall resolution.

### 4. Container detection (cgroup v1 vs v2)

- cgroup v1: check `/proc/1/cgroup` for `/docker/`, `/lxc/`, etc.
- cgroup v2: different format in `/proc/self/mountinfo`
- Also check: `/.dockerenv`, `/run/.containerenv` (podman)
- Container types: Docker, podman, LXC, systemd-nspawn

### 5. `/proc/cpuinfo` architecture variance

ARM and x86 have different field names. `model name` does not exist on all architectures. Handle missing fields gracefully with zero values or "unknown".

### 6. JSON flattening for external facts

Nested JSON becomes dotted keys. Arrays are handled as: `key.0`, `key.1`, etc. Null values are preserved as `nil` in the map.

### 7. Fact resolution errors

Individual collector failures must not fail the entire `Gather()` call. Log the error, skip that collector's facts, return everything else. The caller gets a partial map rather than an error for transient issues (e.g., permission denied on one `/sys` file).

## Future Considerations

These are explicitly out of scope for v1 but the design should not prevent them:

- **Additional platforms**: BSD, Solaris. Build tags make this additive — new `_freebsd.go` files, no changes to existing code.
- **Fact caching with TTL**: Some facts (cloud metadata, hardware) change rarely. A caching layer could avoid re-resolution.
- **Structured typed API**: Typed accessors like `gfacts.OS()` returning a struct, built on top of the map API.
- **gRPC / HTTP server mode**: Expose facts over the network for agent-based architectures.

## Dependencies

Targeting minimal external dependencies:

- Go standard library (primary)
- `golang.org/x/sys` — extended syscall support if `syscall` package is insufficient
- No other required dependencies anticipated

## Testing Strategy

- **Unit tests per collector**: Mock `/proc`, `/sys` file contents via test fixtures or `os.DirFS`
- **Integration tests**: Run on actual Linux and macOS (CI matrix)
- **External fact tests**: Temp directories with test fixtures
- **CLI tests**: Capture stdout, verify JSON and plain output
- **Cross-reference**: Compare output against Ruby Facter on same system for key facts during development
