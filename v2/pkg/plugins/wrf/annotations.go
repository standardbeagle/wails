package wrf

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Annotation represents a WRF annotation
type Annotation struct {
	Type   string            // query, mutation, subscription, event
	Params map[string]string // Parameters like cache, invalidate
}

// ParseAnnotations extracts WRF annotations from comments
func ParseAnnotations(comments []string) []Annotation {
	var annotations []Annotation

	// Pattern: @wrf:type(param: "value", param2: value)
	pattern := regexp.MustCompile(`@wrf:(\w+)(?:\((.*?)\))?`)

	for _, comment := range comments {
		matches := pattern.FindStringSubmatch(comment)
		if len(matches) < 2 {
			continue
		}

		ann := Annotation{
			Type:   matches[1],
			Params: make(map[string]string),
		}

		// Parse parameters if present
		if len(matches) > 2 && matches[2] != "" {
			params := parseParameters(matches[2])
			ann.Params = params
		}

		annotations = append(annotations, ann)
	}

	return annotations
}

// parseParameters parses annotation parameters
func parseParameters(paramStr string) map[string]string {
	params := make(map[string]string)

	// Simple parser for key: value pairs
	parts := strings.Split(paramStr, ",")
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), ":", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.Trim(strings.TrimSpace(kv[1]), `"'`)
			params[key] = value
		}
	}

	return params
}

// Common annotation types
const (
	AnnQuery        = "query"
	AnnMutation     = "mutation"
	AnnSubscription = "subscription"
	AnnEvent        = "event"
)

// QueryAnnotation extracts query-specific parameters
type QueryAnnotation struct {
	Cache      time.Duration `json:"cache"`
	Events     []string      `json:"events"`
	StaleTime  time.Duration `json:"staleTime"`
	RefetchOn  []string      `json:"refetchOn"`
	Enabled    bool          `json:"enabled"`
}

// MutationAnnotation extracts mutation-specific parameters
type MutationAnnotation struct {
	Optimistic  bool     `json:"optimistic"`
	Invalidate  []string `json:"invalidate"`
	Events      []string `json:"events"`
	OnSuccess   string   `json:"onSuccess"`
	RetryCount  int      `json:"retryCount"`
}

// SubscriptionAnnotation extracts subscription-specific parameters
type SubscriptionAnnotation struct {
	EventName string `json:"eventName"`
	AutoStart bool   `json:"autoStart"`
	Buffer    int    `json:"buffer"`
}

// EventAnnotation extracts event-specific parameters
type EventAnnotation struct {
	EventName string `json:"eventName"`
	Global    bool   `json:"global"`
	Typed     bool   `json:"typed"`
}

// ExtractQueryAnnotation extracts query annotation from parsed annotations
func ExtractQueryAnnotation(annotations []Annotation) *QueryAnnotation {
	for _, ann := range annotations {
		if ann.Type == AnnQuery {
			qa := &QueryAnnotation{
				Cache:     5 * time.Minute, // default
				StaleTime: 1 * time.Minute, // default
				Enabled:   true,            // default
			}

			if cache, ok := ann.Params["cache"]; ok {
				if d, err := parseDuration(cache); err == nil {
					qa.Cache = d
				}
			}

			if staleTime, ok := ann.Params["staleTime"]; ok {
				if d, err := parseDuration(staleTime); err == nil {
					qa.StaleTime = d
				}
			}

			if events, ok := ann.Params["events"]; ok {
				qa.Events = parseStringArray(events)
			}

			if refetchOn, ok := ann.Params["refetchOn"]; ok {
				qa.RefetchOn = parseStringArray(refetchOn)
			}

			if enabled, ok := ann.Params["enabled"]; ok {
				qa.Enabled = parseBool(enabled)
			}

			return qa
		}
	}
	return nil
}

// ExtractMutationAnnotation extracts mutation annotation from parsed annotations
func ExtractMutationAnnotation(annotations []Annotation) *MutationAnnotation {
	for _, ann := range annotations {
		if ann.Type == AnnMutation {
			ma := &MutationAnnotation{
				Optimistic: false, // default
				RetryCount: 3,     // default
			}

			if optimistic, ok := ann.Params["optimistic"]; ok {
				ma.Optimistic = parseBool(optimistic)
			}

			if invalidate, ok := ann.Params["invalidate"]; ok {
				ma.Invalidate = parseStringArray(invalidate)
			}

			if events, ok := ann.Params["events"]; ok {
				ma.Events = parseStringArray(events)
			}

			if onSuccess, ok := ann.Params["onSuccess"]; ok {
				ma.OnSuccess = onSuccess
			}

			if retryCount, ok := ann.Params["retryCount"]; ok {
				if count, err := strconv.Atoi(retryCount); err == nil {
					ma.RetryCount = count
				}
			}

			return ma
		}
	}
	return nil
}

// ExtractSubscriptionAnnotation extracts subscription annotation from parsed annotations
func ExtractSubscriptionAnnotation(annotations []Annotation) *SubscriptionAnnotation {
	for _, ann := range annotations {
		if ann.Type == AnnSubscription {
			sa := &SubscriptionAnnotation{
				AutoStart: true, // default
				Buffer:    10,   // default
			}

			if eventName, ok := ann.Params["eventName"]; ok {
				sa.EventName = eventName
			}

			if autoStart, ok := ann.Params["autoStart"]; ok {
				sa.AutoStart = parseBool(autoStart)
			}

			if buffer, ok := ann.Params["buffer"]; ok {
				if buf, err := strconv.Atoi(buffer); err == nil {
					sa.Buffer = buf
				}
			}

			return sa
		}
	}
	return nil
}

// ExtractEventAnnotation extracts event annotation from parsed annotations
func ExtractEventAnnotation(annotations []Annotation) *EventAnnotation {
	for _, ann := range annotations {
		if ann.Type == AnnEvent {
			ea := &EventAnnotation{
				Global: false, // default
				Typed:  true,  // default
			}

			if eventName, ok := ann.Params["eventName"]; ok {
				ea.EventName = eventName
			}

			if global, ok := ann.Params["global"]; ok {
				ea.Global = parseBool(global)
			}

			if typed, ok := ann.Params["typed"]; ok {
				ea.Typed = parseBool(typed)
			}

			return ea
		}
	}
	return nil
}

// Helper functions for parsing parameters

func parseDuration(s string) (time.Duration, error) {
	// Handle numeric values as seconds
	if d, err := strconv.Atoi(s); err == nil {
		return time.Duration(d) * time.Second, nil
	}
	
	// Handle duration strings like "5m", "1h"
	return time.ParseDuration(s)
}

func parseStringArray(s string) []string {
	if s == "" {
		return nil
	}
	
	// Handle comma-separated values
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	
	return result
}

func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "yes"
}