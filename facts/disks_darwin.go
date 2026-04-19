//go:build darwin

package facts

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
)

// DisksCollector gathers disk facts on macOS.
// This requires exec'ing diskutil as there is no pure Go way to enumerate
// physical disks on macOS without CGO (IOKit).
type DisksCollector struct{}

func (d *DisksCollector) Collect() (map[string]any, error) {
	result := make(map[string]any)

	out, err := exec.Command("diskutil", "list", "-plist").Output()
	if err != nil {
		slog.Debug("diskutil exec failed", "error", err)
		return result, nil
	}

	disks := parseDiskutilPlist(out)
	for _, disk := range disks {
		name := disk.id
		if name == "" {
			continue
		}
		prefix := fmt.Sprintf("disks.%s", name)
		if disk.size > 0 {
			result[prefix+".size_bytes"] = disk.size
			result[prefix+".size"] = humanBytes(disk.size)
		}
		if disk.content != "" {
			result[prefix+".type"] = strings.ToLower(disk.content)
		}
	}

	return result, nil
}

type diskInfo struct {
	id      string
	size    uint64
	content string
}

// parseDiskutilPlist parses the XML plist from diskutil list -plist.
// It extracts AllDisksAndPartitions entries using basic XML parsing
// to avoid external dependencies.
func parseDiskutilPlist(data []byte) []diskInfo {
	// The plist has a top-level dict with key "AllDisksAndPartitions"
	// followed by an array of dicts. Each dict has DeviceIdentifier, Size, Content keys.
	type plistDict struct {
		keys   []string
		values []string
	}

	var result []diskInfo

	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	var inArray bool
	var currentKey string
	var current diskInfo
	var inDict int

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "array":
				if currentKey == "AllDisksAndPartitions" {
					inArray = true
				}
			case "dict":
				if inArray {
					inDict++
					if inDict == 1 {
						current = diskInfo{}
					}
				}
			case "key":
				// Read the key text
				if tok2, err := decoder.Token(); err == nil {
					if cd, ok := tok2.(xml.CharData); ok {
						currentKey = string(cd)
					}
				}
			case "integer":
				if inArray && inDict == 1 {
					if tok2, err := decoder.Token(); err == nil {
						if cd, ok := tok2.(xml.CharData); ok {
							if currentKey == "Size" {
								if v, err := strconv.ParseUint(string(cd), 10, 64); err == nil {
									current.size = v
								}
							}
						}
					}
				}
			case "string":
				if inArray && inDict == 1 {
					if tok2, err := decoder.Token(); err == nil {
						if cd, ok := tok2.(xml.CharData); ok {
							switch currentKey {
							case "DeviceIdentifier":
								current.id = string(cd)
							case "Content":
								current.content = string(cd)
							}
						}
					}
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "dict":
				if inArray {
					if inDict == 1 && current.id != "" {
						result = append(result, current)
					}
					inDict--
				}
			case "array":
				if inArray {
					inArray = false
				}
			}
		}
	}

	return result
}
