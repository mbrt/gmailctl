package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/mbrt/gmailctl/internal/engine/config/v1alpha3"
	"github.com/mbrt/gmailctl/internal/engine/filter"
	"github.com/mbrt/gmailctl/internal/engine/label"
	"github.com/mbrt/gmailctl/internal/engine/rimport"
)

// MergeStrategy defines how to resolve conflicts automatically
type MergeStrategy string

const (
	StrategyPrompt MergeStrategy = ""       // Default: prompt for each conflict
	StrategyLocal  MergeStrategy = "local"  // Auto-keep local on conflict
	StrategyGmail  MergeStrategy = "gmail"  // Auto-take Gmail on conflict
)

// ValidStrategies returns list of valid strategy values for help text
func ValidStrategies() []string {
	return []string{"local", "gmail"}
}

// MergeResult contains the result of analyzing filters for merge
type MergeResult struct {
	HasNewFilters bool
	HasConflicts  bool
	NewRules      []v1alpha3.Rule
	NewLabels     []v1alpha3.Label
}

// analyzeForMerge compares local and Gmail filters/labels and returns what needs to be added.
// This does NOT modify any files - it just analyzes and returns new rules/labels.
func analyzeForMerge(localFilters filter.Filters, localLabels label.Labels,
	gmailFilters filter.Filters, gmailLabels label.Labels,
	strategy MergeStrategy, interactive bool) (MergeResult, error) {

	result := MergeResult{}

	// 1. Categorize filters
	categories := filter.CategorizeMerge(localFilters, gmailFilters)

	// 2. Find new labels (gmail-only labels) - do this early for summary
	localLabelMap := make(map[string]bool)
	for _, l := range localLabels {
		localLabelMap[l.Name] = true
	}
	for _, l := range gmailLabels {
		if !localLabelMap[l.Name] {
			var color *v1alpha3.LabelColor
			if l.Color != nil {
				color = &v1alpha3.LabelColor{
					Background: l.Color.Background,
					Text:       l.Color.Text,
				}
			}
			result.NewLabels = append(result.NewLabels, v1alpha3.Label{
				Name:  l.Name,
				Color: color,
			})
		}
	}

	// 3. Report merge summary
	fmt.Println("=== Merge Analysis ===")
	fmt.Println()
	fmt.Println("Filters:")
	if len(categories.Matched) > 0 {
		fmt.Printf("  Unchanged: %d filter(s) already in sync\n", len(categories.Matched))
	}
	if len(categories.LocalOnly) > 0 {
		fmt.Printf("  Local-only: %d filter(s) (keeping as-is)\n", len(categories.LocalOnly))
	}
	if len(categories.GmailOnly) > 0 {
		fmt.Printf("  Gmail-only: %d new filter(s) to add\n", len(categories.GmailOnly))
	}
	if len(categories.Conflicts) > 0 {
		fmt.Printf("  Conflicts: %d filter(s) need manual resolution\n", len(categories.Conflicts))
	}

	fmt.Println()
	fmt.Println("Labels:")
	fmt.Printf("  Local: %d, Gmail: %d\n", len(localLabels), len(gmailLabels))
	if len(result.NewLabels) > 0 {
		fmt.Printf("  New from Gmail: %d label(s) to add\n", len(result.NewLabels))
	} else {
		fmt.Println("  No new labels to add")
	}
	fmt.Println()

	// 4. Handle conflicts - these require manual resolution
	if len(categories.Conflicts) > 0 {
		result.HasConflicts = true
		fmt.Println("=== Filter Conflicts (require manual editing) ===")
		fmt.Println()

		for i, conflict := range categories.Conflicts {
			fmt.Printf("--- Conflict %d/%d ---\n", i+1, len(categories.Conflicts))
			diff := filter.NewMinimalFiltersDiff(
				filter.Filters{conflict.Gmail},
				filter.Filters{conflict.Local},
				false, // debugInfo
				3,     // contextLines
				true,  // colorize
			)
			fmt.Println(diff.String())
		}

		if !interactive && strategy == StrategyPrompt {
			return result, fmt.Errorf(
				"%d conflict(s) require manual resolution.\n"+
					"Edit your local config to match Gmail, or use 'gmailctl apply' to push local changes.",
				len(categories.Conflicts))
		}

		fmt.Println("Note: Conflicts above require manual editing in your config file.")
		fmt.Println()
	}

	// 5. Check if there's anything to add
	if len(categories.GmailOnly) == 0 && len(result.NewLabels) == 0 {
		fmt.Println("Nothing new to add from Gmail.")
		return result, nil
	}

	// 6. Convert Gmail-only filters to rules
	if len(categories.GmailOnly) > 0 {
		result.HasNewFilters = true
		newRules, err := filtersToRules(categories.GmailOnly)
		if err != nil {
			return result, fmt.Errorf("converting new filters to rules: %w", err)
		}
		result.NewRules = newRules
	}

	return result, nil
}

// rulesToJsonnet converts rules to Jsonnet text format for display
func rulesToJsonnet(rules []v1alpha3.Rule) (string, error) {
	if len(rules) == 0 {
		return "", nil
	}

	var buf bytes.Buffer
	for i, rule := range rules {
		// Convert rule to JSON
		b, err := json.MarshalIndent(rule, "    ", "  ")
		if err != nil {
			return "", err
		}

		// Make JSON look like Jsonnet (remove quotes from keys)
		text := string(b)
		keyRe := regexp.MustCompile(`"([a-zA-Z][a-zA-Z0-9]*)":`)
		text = keyRe.ReplaceAllString(text, "$1:")

		buf.WriteString("    ")
		buf.WriteString(text)
		if i < len(rules)-1 {
			buf.WriteString(",")
		}
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

// labelsToJsonnet converts labels to Jsonnet text format for display
func labelsToJsonnet(labels []v1alpha3.Label) (string, error) {
	if len(labels) == 0 {
		return "", nil
	}

	var buf bytes.Buffer
	for i, lbl := range labels {
		b, err := json.MarshalIndent(lbl, "    ", "  ")
		if err != nil {
			return "", err
		}

		text := string(b)
		keyRe := regexp.MustCompile(`"([a-zA-Z][a-zA-Z0-9]*)":`)
		text = keyRe.ReplaceAllString(text, "$1:")

		buf.WriteString("    ")
		buf.WriteString(text)
		if i < len(labels)-1 {
			buf.WriteString(",")
		}
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

// mergeTimestamp returns the current timestamp for merge comments
func mergeTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// appendRulesToFile appends new rules and labels to a Jsonnet file.
// It does text-based insertion to preserve the original file structure.
func appendRulesToFile(originalContent string, newRules []v1alpha3.Rule, newLabels []v1alpha3.Label) (string, error) {
	if len(newRules) == 0 && len(newLabels) == 0 {
		return originalContent, nil
	}

	result := originalContent

	// Append new labels if any
	if len(newLabels) > 0 {
		labelsText, err := labelsToJsonnet(newLabels)
		if err != nil {
			return "", fmt.Errorf("converting labels to jsonnet: %w", err)
		}

		// Find the labels array closing bracket
		// Look for "labels:" or "labels :" and find its closing ]
		labelsIdx := findArrayStart(result, "labels")
		if labelsIdx != -1 {
			// Find the matching ] for this array
			bracketIdx := findMatchingBracket(result, labelsIdx)
			if bracketIdx != -1 {
				// Check if there's content before the bracket (need comma)
				beforeBracket := strings.TrimRight(result[:bracketIdx], " \t\n")
				needsComma := len(beforeBracket) > 0 && !strings.HasSuffix(beforeBracket, "[") && !strings.HasSuffix(beforeBracket, ",")

				// Build insertion
				var insertion strings.Builder
				if needsComma {
					insertion.WriteString(",")
				}
				insertion.WriteString("\n")
				insertion.WriteString(fmt.Sprintf("    // --- New labels from Gmail (added by gmailctl pull --merge @ %s) ---\n", mergeTimestamp()))
				insertion.WriteString(labelsText)

				// Insert before the closing bracket
				result = result[:bracketIdx] + insertion.String() + result[bracketIdx:]
			}
		}
	}

	// Append new rules if any
	if len(newRules) > 0 {
		rulesText, err := rulesToJsonnet(newRules)
		if err != nil {
			return "", fmt.Errorf("converting rules to jsonnet: %w", err)
		}

		// Find the rules array closing bracket
		rulesIdx := findArrayStart(result, "rules")
		if rulesIdx != -1 {
			// Find the matching ] for this array
			bracketIdx := findMatchingBracket(result, rulesIdx)
			if bracketIdx != -1 {
				// Check if there's content before the bracket (need comma)
				beforeBracket := strings.TrimRight(result[:bracketIdx], " \t\n")
				needsComma := len(beforeBracket) > 0 && !strings.HasSuffix(beforeBracket, "[") && !strings.HasSuffix(beforeBracket, ",")

				// Build insertion
				var insertion strings.Builder
				if needsComma {
					insertion.WriteString(",")
				}
				insertion.WriteString("\n")
				insertion.WriteString(fmt.Sprintf("    // --- New rules from Gmail (added by gmailctl pull --merge @ %s) ---\n", mergeTimestamp()))
				insertion.WriteString(rulesText)

				// Insert before the closing bracket
				result = result[:bracketIdx] + insertion.String() + result[bracketIdx:]
			}
		} else {
			// Fallback: find last ] before last } (original heuristic)
			lastBrace := strings.LastIndex(result, "}")
			if lastBrace == -1 {
				return "", fmt.Errorf("cannot find closing brace in config file")
			}

			searchArea := result[:lastBrace]
			lastBracket := strings.LastIndex(searchArea, "]")
			if lastBracket == -1 {
				return "", fmt.Errorf("cannot find rules array closing bracket in config file")
			}

			beforeBracket := strings.TrimRight(result[:lastBracket], " \t\n")
			needsComma := len(beforeBracket) > 0 && !strings.HasSuffix(beforeBracket, "[") && !strings.HasSuffix(beforeBracket, ",")

			var insertion strings.Builder
			if needsComma {
				insertion.WriteString(",")
			}
			insertion.WriteString("\n")
			insertion.WriteString(fmt.Sprintf("    // --- New rules from Gmail (added by gmailctl pull --merge @ %s) ---\n", mergeTimestamp()))
			insertion.WriteString(rulesText)

			result = result[:lastBracket] + insertion.String() + result[lastBracket:]
		}
	}

	return result, nil
}

// findArrayStart finds the position of '[' after "key:" at the top level of the main config object.
// It skips any "key:" that appears inside local variable declarations.
func findArrayStart(content, key string) int {
	// First, find the start of the main config object
	// The main object is the last top-level { in the file (after all local declarations)
	mainObjStart := findMainObjectStart(content)
	if mainObjStart == -1 {
		return -1
	}

	// Now search for "key:" only within the main object, at the first nesting level
	// We need to find "key:" that's directly inside the main object, not nested deeper
	searchArea := content[mainObjStart:]

	// Try both "key:" and "key :"
	patterns := []string{key + ":", key + " :"}
	for _, pattern := range patterns {
		// Find all occurrences and check if they're at the right nesting level
		offset := 0
		for {
			idx := strings.Index(searchArea[offset:], pattern)
			if idx == -1 {
				break
			}
			actualIdx := mainObjStart + offset + idx

			// Check if this is at nesting level 1 (directly inside main object)
			if isAtNestingLevel(content, mainObjStart, actualIdx, 1) {
				// Find the [ after this
				afterKey := content[actualIdx+len(pattern):]
				bracketOffset := strings.Index(afterKey, "[")
				if bracketOffset != -1 {
					return actualIdx + len(pattern) + bracketOffset
				}
			}
			offset += idx + len(pattern)
		}
	}
	return -1
}

// findMainObjectStart finds the start of the main config object (the { that's not inside a local declaration)
func findMainObjectStart(content string) int {
	// The main object is typically the last top-level {
	// We need to skip past all "local ... = ..." declarations

	// Find all top-level braces and return the last one that starts an object
	depth := 0
	inString := false
	escaped := false
	lastTopLevelBrace := -1

	for i := 0; i < len(content); i++ {
		c := content[i]

		if escaped {
			escaped = false
			continue
		}

		if c == '\\' && inString {
			escaped = true
			continue
		}

		if c == '"' {
			inString = !inString
			continue
		}

		if inString {
			continue
		}

		switch c {
		case '{':
			if depth == 0 {
				lastTopLevelBrace = i
			}
			depth++
		case '}':
			depth--
		case '[':
			depth++
		case ']':
			depth--
		}
	}

	return lastTopLevelBrace
}

// isAtNestingLevel checks if the position is at the specified nesting level relative to mainObjStart
func isAtNestingLevel(content string, mainObjStart, pos, targetLevel int) bool {
	depth := 0
	inString := false
	escaped := false

	for i := mainObjStart; i < pos && i < len(content); i++ {
		c := content[i]

		if escaped {
			escaped = false
			continue
		}

		if c == '\\' && inString {
			escaped = true
			continue
		}

		if c == '"' {
			inString = !inString
			continue
		}

		if inString {
			continue
		}

		switch c {
		case '{', '[':
			depth++
		case '}', ']':
			depth--
		}
	}

	return depth == targetLevel
}

// findMatchingBracket finds the ] that matches the [ at the given position
func findMatchingBracket(content string, openBracketIdx int) int {
	if openBracketIdx >= len(content) || content[openBracketIdx] != '[' {
		return -1
	}

	depth := 0
	inString := false
	escaped := false

	for i := openBracketIdx; i < len(content); i++ {
		c := content[i]

		if escaped {
			escaped = false
			continue
		}

		if c == '\\' && inString {
			escaped = true
			continue
		}

		if c == '"' {
			inString = !inString
			continue
		}

		if inString {
			continue
		}

		if c == '[' {
			depth++
		} else if c == ']' {
			depth--
			if depth == 0 {
				return i
			}
		}
	}

	return -1
}

// filtersToRules converts internal filter representation to v1alpha3 rules.
// This is extracted from rimport.Import to allow preserving local config structure.
func filtersToRules(fs filter.Filters) ([]v1alpha3.Rule, error) {
	var rules []v1alpha3.Rule
	for i, f := range fs {
		r, err := rimport.FilterToRule(f)
		if err != nil {
			return nil, fmt.Errorf("converting filter #%d: %w", i, err)
		}
		rules = append(rules, r)
	}
	return rules, nil
}

// labelsToConfig converts internal label representation to v1alpha3 labels.
func labelsToConfig(ls label.Labels) []v1alpha3.Label {
	var labels []v1alpha3.Label
	for _, l := range ls {
		var color *v1alpha3.LabelColor
		if l.Color != nil {
			color = &v1alpha3.LabelColor{
				Background: l.Color.Background,
				Text:       l.Color.Text,
			}
		}
		labels = append(labels, v1alpha3.Label{
			Name:  l.Name,
			Color: color,
		})
	}
	return labels
}

