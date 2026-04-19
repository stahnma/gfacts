package facts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// CloudCollector detects cloud provider and gathers instance metadata.
type CloudCollector struct{}

func (c *CloudCollector) Collect() (map[string]any, error) {
	result := make(map[string]any)

	provider := detectProvider()
	if provider == "" {
		return result, nil
	}

	result["cloud.provider"] = provider

	switch provider {
	case "aws":
		c.collectAWS(result)
	case "gcp":
		c.collectGCP(result)
	case "azure":
		c.collectAzure(result)
	}

	return result, nil
}

func detectProvider() string {
	// Check DMI hints on Linux.
	if data, err := os.ReadFile("/sys/class/dmi/id/board_asset_tag"); err == nil {
		tag := strings.TrimSpace(string(data))
		if strings.Contains(tag, "Amazon") || strings.Contains(tag, "aws") {
			return "aws"
		}
	}
	if data, err := os.ReadFile("/sys/class/dmi/id/sys_vendor"); err == nil {
		vendor := strings.TrimSpace(string(data))
		switch {
		case strings.Contains(vendor, "Amazon"):
			return "aws"
		case strings.Contains(vendor, "Google"):
			return "gcp"
		case strings.Contains(vendor, "Microsoft"):
			return "azure"
		}
	}
	if data, err := os.ReadFile("/sys/class/dmi/id/chassis_asset_tag"); err == nil {
		tag := strings.TrimSpace(string(data))
		if strings.Contains(tag, "7783-7084-3265-9085-8269-3286-77") {
			return "azure"
		}
	}

	// Fall back to metadata endpoint probing.
	if probeURL("http://169.254.169.254/latest/meta-data/", nil) {
		return "aws"
	}
	if probeURL("http://metadata.google.internal/computeMetadata/v1/", map[string]string{"Metadata-Flavor": "Google"}) {
		return "gcp"
	}
	if probeURL("http://169.254.169.254/metadata/instance?api-version=2021-02-01", map[string]string{"Metadata": "true"}) {
		return "azure"
	}

	return ""
}

func probeURL(url string, headers map[string]string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}

func (c *CloudCollector) collectAWS(result map[string]any) {
	endpoints := map[string]string{
		"cloud.aws.instance_id":    "/latest/meta-data/instance-id",
		"cloud.aws.instance_type":  "/latest/meta-data/instance-type",
		"cloud.aws.ami_id":         "/latest/meta-data/ami-id",
		"cloud.aws.region":         "/latest/meta-data/placement/region",
		"cloud.aws.az":             "/latest/meta-data/placement/availability-zone",
		"cloud.aws.local_hostname": "/latest/meta-data/local-hostname",
		"cloud.aws.local_ipv4":     "/latest/meta-data/local-ipv4",
		"cloud.aws.public_ipv4":    "/latest/meta-data/public-ipv4",
	}
	base := "http://169.254.169.254"
	for key, path := range endpoints {
		if val := fetchMeta(base+path, nil); val != "" {
			result[key] = val
		}
	}
}

func (c *CloudCollector) collectGCP(result map[string]any) {
	headers := map[string]string{"Metadata-Flavor": "Google"}
	base := "http://metadata.google.internal/computeMetadata/v1"
	endpoints := map[string]string{
		"cloud.gcp.instance_id":   "/instance/id",
		"cloud.gcp.machine_type":  "/instance/machine-type",
		"cloud.gcp.zone":          "/instance/zone",
		"cloud.gcp.project_id":    "/project/project-id",
		"cloud.gcp.hostname":      "/instance/hostname",
	}
	for key, path := range endpoints {
		if val := fetchMeta(base+path, headers); val != "" {
			result[key] = val
		}
	}
}

func (c *CloudCollector) collectAzure(result map[string]any) {
	headers := map[string]string{"Metadata": "true"}
	url := "http://169.254.169.254/metadata/instance?api-version=2021-02-01&format=json"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var meta map[string]any
	if err := json.Unmarshal(body, &meta); err != nil {
		return
	}

	if compute, ok := meta["compute"].(map[string]any); ok {
		fields := map[string]string{
			"cloud.azure.vm_id":       "vmId",
			"cloud.azure.vm_size":     "vmSize",
			"cloud.azure.location":    "location",
			"cloud.azure.resource_group": "resourceGroupName",
			"cloud.azure.subscription_id": "subscriptionId",
		}
		for factKey, metaKey := range fields {
			if v, ok := compute[metaKey]; ok {
				result[factKey] = fmt.Sprintf("%v", v)
			}
		}
	}
}

func fetchMeta(url string, headers map[string]string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return ""
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(body))
}
