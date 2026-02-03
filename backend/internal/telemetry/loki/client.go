// Package loki provides a client to push log entries to Grafana Loki.
package loki

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// PushRequest is the Loki push API request body (v1).
type PushRequest struct {
	Streams []Stream `json:"streams"`
}

// Stream is a single stream with labels and log entries.
type Stream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"` // each entry is [timestamp_ns, log_line]
}

// labelSanitize replaces characters that are invalid in Loki label names/values.
// Loki labels: name must match [a-zA-Z_:][a-zA-Z0-9_:]*, value can be any string but we avoid problematic chars.
var labelSanitize = regexp.MustCompile(`[^a-zA-Z0-9_\-:]`)

// eventFields is used to parse only the fields we need from a TelemetryEvent JSON for labels and timestamp.
type eventFields struct {
	OrgID     string `json:"orgId"`
	EventType string `json:"eventType"`
	Source    string `json:"source"`
	CreatedAt string `json:"createdAt"` // RFC3339 from protobuf timestamp
}

// PushEventJSON parses the telemetry event JSON (Kafka message value), extracts timestamp and labels, and pushes to Loki.
// If parsing fails, the raw line is pushed with current time and no extra labels.
func PushEventJSON(ctx context.Context, baseURL string, rawJSON []byte) error {
	line := string(rawJSON)
	labels := map[string]string{}
	ts := time.Now().UTC()
	var fields eventFields
	if err := json.Unmarshal(rawJSON, &fields); err == nil {
		if fields.OrgID != "" {
			labels["org_id"] = fields.OrgID
		}
		if fields.EventType != "" {
			labels["event_type"] = fields.EventType
		}
		if fields.Source != "" {
			labels["source"] = fields.Source
		}
		if fields.CreatedAt != "" {
			if t, err := time.Parse(time.RFC3339Nano, fields.CreatedAt); err == nil {
				ts = t
			} else if t, err := time.Parse(time.RFC3339, fields.CreatedAt); err == nil {
				ts = t
			}
		}
	}
	return PushEvent(ctx, baseURL, ts, line, labels)
}

// PushEvent sends a single log line to Loki at the given base URL (e.g. http://localhost:3100).
// timestamp is the event time; line is the log line (e.g. JSON). labels are added to the stream (e.g. job=ztcp, org_id=...).
// Returns an error if the HTTP request fails or Loki returns non-2xx.
func PushEvent(ctx context.Context, baseURL string, timestamp time.Time, line string, labels map[string]string) error {
	if baseURL == "" {
		return fmt.Errorf("loki: base URL is empty")
	}
	ns := timestamp.UnixNano()
	streamLabels := make(map[string]string, len(labels)+1)
	streamLabels["job"] = "ztcp"
	for k, v := range labels {
		sanitized := labelSanitize.ReplaceAllString(strings.TrimSpace(v), "_")
		if sanitized != "" {
			streamLabels[k] = sanitized
		}
	}
	body := PushRequest{
		Streams: []Stream{{
			Stream: streamLabels,
			Values: [][]string{{fmt.Sprintf("%d", ns), line}},
		}},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	url := strings.TrimSuffix(baseURL, "/") + "/loki/api/v1/push"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("loki: push returned %s", resp.Status)
	}
	return nil
}
