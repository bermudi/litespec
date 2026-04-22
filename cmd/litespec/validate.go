package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bermudi/litespec/internal"
)

func cmdValidate(args []string) error {
	if hasHelpFlag(args) {
		printValidateHelp()
		return nil
	}
	if err := checkUnknownFlags(args, map[string]bool{"--all": true, "--changes": true, "--specs": true, "--decisions": true, "--strict": true, "--json": true, "--type": true}); err != nil {
		return err
	}

	var positional string
	var flagAll, flagChanges, flagSpecs, flagDecisions, strict, asJSON bool
	var typeFilter string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--all":
			flagAll = true
		case "--changes":
			flagChanges = true
		case "--specs":
			flagSpecs = true
		case "--decisions":
			flagDecisions = true
		case "--strict":
			strict = true
		case jsonFlag:
			asJSON = true
		case "--type":
			if i+1 >= len(args) {
				return fmt.Errorf("--type requires a value (change or spec)")
			}
			typeFilter = args[i+1]
			i++
		default:
			if !strings.HasPrefix(args[i], "-") && positional == "" {
				positional = args[i]
			}
		}
	}

	hasBulk := flagAll || flagChanges || flagSpecs || flagDecisions

	if positional != "" && hasBulk {
		return fmt.Errorf("positional name and bulk flags (--all, --changes, --specs, --decisions) are mutually exclusive")
	}

	if typeFilter != "" && positional == "" {
		return fmt.Errorf("--type requires a positional name")
	}

	if typeFilter != "" && hasBulk {
		return fmt.Errorf("--type cannot be used with bulk flags")
	}

	if typeFilter != "" && typeFilter != "change" && typeFilter != "spec" && typeFilter != "decision" {
		return fmt.Errorf("--type must be 'change', 'spec', or 'decision', got %q", typeFilter)
	}

	root, err := internal.FindProjectRoot()
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(root, internal.ProjectDirName)); err != nil {
		return fmt.Errorf("not a litespec project. Run 'litespec init' first")
	}

	var result *internal.ValidationResult

	if positional != "" {
		changeList, _ := internal.ListChanges(root)
		specList, _ := internal.ListSpecs(root)
		changeNames := make([]string, len(changeList))
		for i, c := range changeList {
			changeNames[i] = c.Name
		}
		specNames := make([]string, len(specList))
		for i, s := range specList {
			specNames[i] = s.Name
		}
		isChange := contains(changeNames, positional)
		isSpec := contains(specNames, positional)
		isDecision := false
		decisionMatch, _ := internal.FindDecisionBySlug(root, positional)
		if decisionMatch != nil {
			isDecision = true
		}

		if typeFilter == "change" {
			isSpec = false
			isDecision = false
		} else if typeFilter == "spec" {
			isChange = false
			isDecision = false
		} else if typeFilter == "decision" {
			isChange = false
			isSpec = false
		}

		matches := 0
		if isChange {
			matches++
		}
		if isSpec {
			matches++
		}
		if isDecision {
			matches++
		}

		if matches > 1 {
			return fmt.Errorf("%q is ambiguous — matches multiple artifact types. Use --type change, --type spec, or --type decision", positional)
		}

		if matches == 0 {
			return fmt.Errorf("%q not found as a change, spec, or decision", positional)
		}

		if isChange {
			result, err = internal.ValidateChange(root, positional)
		} else if isSpec {
			result, err = internal.ValidateSpec(root, positional)
		} else {
			result, err = internal.ValidateDecision(root, positional)
		}
	} else {
		// Mutual exclusion: --decisions is exclusive with --changes and --specs
		if flagDecisions && (flagChanges || flagSpecs) {
			return fmt.Errorf("--decisions cannot be combined with --changes or --specs (use --all to validate everything)")
		}

		if flagDecisions {
			result, err = internal.ValidateDecisions(root)
			if err != nil {
				return err
			}
			if strict && len(result.Warnings) > 0 {
				result.Valid = false
			}
		} else {
			validateSpecs := flagSpecs || flagAll || (!flagChanges && !flagSpecs && !flagDecisions)
			validateChanges := flagChanges || flagAll || (!flagChanges && !flagSpecs && !flagDecisions)

			if validateSpecs && validateChanges {
				result, err = internal.ValidateAll(root, strict)
			} else if validateSpecs {
				result, err = internal.ValidateSpecs(root)
			} else {
			changes, listErr := internal.ListChanges(root)
			if listErr != nil {
				return listErr
			}
			result = &internal.ValidationResult{Valid: true}
			for _, ci := range changes {
				changeResult, changeErr := internal.ValidateChange(root, ci.Name)
				if changeErr != nil {
					return changeErr
				}
				result.Errors = append(result.Errors, changeResult.Errors...)
				result.Warnings = append(result.Warnings, changeResult.Warnings...)
				result.ChangesCount += changeResult.ChangesCount
				result.CapabilitiesCount += changeResult.CapabilitiesCount
				result.RequirementsCount += changeResult.RequirementsCount
				result.ScenariosCount += changeResult.ScenariosCount
			}

			depMap, depErr := internal.LoadDepMap(root)
			if depErr == nil {
				cycles := internal.DetectCycles(depMap)
				for _, cycle := range cycles {
					path := strings.Join(cycle, " -> ")
					result.Errors = append(result.Errors, internal.ValidationIssue{
						Severity: internal.SeverityError,
						Message:  fmt.Sprintf("dependency cycle detected: %s", path),
					})
				}

				overlaps := internal.DetectOverlaps(root, changes, depMap)
				result.Warnings = append(result.Warnings, overlaps...)
			}

			result.Valid = len(result.Errors) == 0
				if strict && len(result.Warnings) > 0 {
					result.Valid = false
				}
			}
		}
	}

	if err != nil {
		return err
	}

	if asJSON {
		out := internal.BuildValidationResultJSON(result)
		data, err := internal.MarshalJSON(out)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		if !result.Valid || (strict && len(result.Warnings) > 0) {
			return fmt.Errorf("validation failed")
		}
		return nil
	}

	for _, issue := range result.Errors {
		fmt.Printf("ERROR  %s: %s\n", issue.File, issue.Message)
	}
	for _, issue := range result.Warnings {
		fmt.Printf("WARN   %s: %s\n", issue.File, issue.Message)
	}

	if strict && len(result.Warnings) > 0 {
		return fmt.Errorf("validation failed")
	}

	if !result.Valid {
		return fmt.Errorf("validation failed")
	}

	fmt.Printf("ok: %d %s, %d %s, %d %s, %d %s\n",
		result.ChangesCount, pluralize("change", result.ChangesCount),
		result.CapabilitiesCount, pluralize("capability", result.CapabilitiesCount),
		result.RequirementsCount, pluralize("requirement", result.RequirementsCount),
		result.ScenariosCount, pluralize("scenario", result.ScenariosCount))
	return nil
}
