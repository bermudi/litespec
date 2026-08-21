package main

import (
	"fmt"

	"github.com/bermudi/litespec/v2/internal"
)

// Response holds all output representations for a command.
type Response struct {
	Full        any    // Full JSON output
	Minimal     any    // Minimal JSON output
	Text        string // Human-readable text output
	MinimalText string // Minimal text output (tab-separated)
}

// Render picks the right output format based on flags.
func Render(resp Response, asJSON, asMinimal bool) error {
	if asJSON {
		v := resp.Full
		if asMinimal && resp.Minimal != nil {
			v = resp.Minimal
		}
		data, err := internal.MarshalJSON(v)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}
	if asMinimal {
		fmt.Println(resp.MinimalText)
		return nil
	}
	fmt.Print(resp.Text)
	return nil
}
