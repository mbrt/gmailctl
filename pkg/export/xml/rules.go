package xml

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/mbrt/gmailfilter/pkg/config"
)

// Property values
const (
	PropertyFrom          = "from"
	PropertyTo            = "to"
	PropertySubject       = "subject"
	PropertyHas           = "hasTheWord"
	PropertyMarkImportant = "shouldAlwaysMarkAsImportant"
	PropertyApplyLabel    = "label"
	PropertyApplyCategory = "smartLabelToApply"
	PropertyDelete        = "shouldTrash"
	PropertyArchive       = "shouldArchive"
	PropertyMarkRead      = "shouldMarkAsRead"
)

// SmartLabel values
const (
	SmartLabelPersonal     = "personal"
	SmartLabelGroup        = "group"
	SmartLabelNotification = "notification"
	SmartLabelPromo        = "promo"
	SmartLabelSocial       = "social"
)

// Entry is a Gmail filter
type Entry []Property

// Property is a property of a Gmail filter, as specified by its XML format
type Property struct {
	Name  string
	Value string
}

// GenerateRules translates a config into entries that map directly into Gmail filters
func GenerateRules(cfg config.Config) ([]Entry, error) {
	res := []Entry{}
	for i, rule := range cfg.Rules {
		entries, err := generateRule(rule, cfg.Consts)
		if err != nil {
			return res, errors.Wrap(err, fmt.Sprintf("error generating rule #%d", i))
		}
		res = append(res, entries...)
	}
	return res, nil
}

func generateRule(rule config.Rule, consts config.Consts) ([]Entry, error) {
	filters, err := generateFilters(rule.Filters, consts)
	if err != nil {
		return nil, errors.Wrap(err, "error generating filters")
	}
	if len(filters) == 0 {
		return nil, errors.New("at least one filter has to be specified")
	}
	actions, err := generateActions(rule.Actions)
	if err != nil {
		return nil, errors.Wrap(err, "error generating actions")
	}
	if len(actions) == 0 {
		return nil, errors.New("at least one action has to be specified")
	}
	return combineFiltersActions(filters, actions), nil
}

func generateFilters(filters config.Filters, consts config.Consts) ([]Property, error) {
	res := []Property{}
	// simple filters
	mf, err := generateMatchFilters(filters.MatchFilters)
	if err != nil {
		return nil, errors.Wrap(err, "error generating match filters")
	}
	res = append(res, mf...)

	// simple filters with consts
	resolved, err := resolveFiltersConsts(filters.Consts.MatchFilters, consts)
	if err != nil {
		return nil, errors.Wrap(err, "error resolving consts in filter")
	}
	mf, err = generateMatchFilters(resolved)
	if err != nil {
		return nil, errors.Wrap(err, "error generating const match filters")
	}
	res = append(res, mf...)

	// negated filters
	mf, err = generateNegatedFilters(filters.Not)
	if err != nil {
		return nil, errors.Wrap(err, "error generating negated filters")
	}
	res = append(res, mf...)

	// negated filters with consts
	resolved, err = resolveFiltersConsts(filters.Consts.Not, consts)
	if err != nil {
		return nil, errors.Wrap(err, "error resolving consts in filter")
	}
	mf, err = generateNegatedFilters(resolved)
	if err != nil {
		return nil, errors.Wrap(err, "error generating negated filters")
	}
	res = append(res, mf...)

	return res, nil
}

func generateMatchFilters(filters config.MatchFilters) ([]Property, error) {
	res := []Property{}
	if len(filters.From) > 0 {
		p := Property{PropertyFrom, joinOR(filters.From)}
		res = append(res, p)
	}
	if len(filters.To) > 0 {
		p := Property{PropertyTo, joinOR(filters.To)}
		res = append(res, p)
	}
	if len(filters.Subject) > 0 {
		p := Property{PropertySubject, joinOR(filters.Subject)}
		res = append(res, p)
	}
	if len(filters.Has) > 0 {
		p := Property{PropertyHas, joinOR(filters.Has)}
		res = append(res, p)
	}
	return res, nil
}

func generateNegatedFilters(filters config.MatchFilters) ([]Property, error) {
	clauses := []string{}
	if len(filters.From) > 0 {
		c := fmt.Sprintf("-{from:%s}", joinOR(filters.From))
		clauses = append(clauses, c)
	}
	if len(filters.To) > 0 {
		c := fmt.Sprintf("-{to:%s}", joinOR(filters.To))
		clauses = append(clauses, c)
	}
	if len(filters.Subject) > 0 {
		c := fmt.Sprintf("-{subject:%s}", joinOR(filters.Subject))
		clauses = append(clauses, c)
	}
	if len(filters.Has) > 0 {
		c := fmt.Sprintf("-%s", joinOR(filters.Has))
		clauses = append(clauses, c)
	}

	if len(clauses) == 0 {
		return nil, nil
	}

	res := Property{PropertyHas, strings.Join(clauses, " ")}
	return []Property{res}, nil
}

func generateActions(actions config.Actions) ([]Property, error) {
	res := []Property{}
	if actions.Archive {
		res = append(res, Property{PropertyArchive, "true"})
	}
	if actions.Delete {
		res = append(res, Property{PropertyDelete, "true"})
	}
	if actions.MarkImportant {
		res = append(res, Property{PropertyMarkImportant, "true"})
	}
	if actions.MarkRead {
		res = append(res, Property{PropertyMarkRead, "true"})
	}
	if len(actions.Category) > 0 {
		cat, err := categoryToSmartLabel(actions.Category)
		if err != nil {
			return nil, err
		}
		res = append(res, Property{PropertyApplyCategory, cat})
	}
	for _, label := range actions.Labels {
		res = append(res, Property{PropertyApplyLabel, label})
	}
	return res, nil
}

func resolveFiltersConsts(mf config.MatchFilters, consts config.Consts) (config.MatchFilters, error) {
	from, err := resolveConsts(mf.From, consts)
	if err != nil {
		return mf, errors.Wrap(err, "error in resolving 'from' clause")
	}
	to, err := resolveConsts(mf.To, consts)
	if err != nil {
		return mf, errors.Wrap(err, "error in resolving 'to' clause")
	}
	sub, err := resolveConsts(mf.Subject, consts)
	if err != nil {
		return mf, errors.Wrap(err, "error in resolving 'subject' clause")
	}
	has, err := resolveConsts(mf.Has, consts)
	if err != nil {
		return mf, errors.Wrap(err, "error in resolving 'has' clause")
	}
	res := config.MatchFilters{
		From:    from,
		To:      to,
		Subject: sub,
		Has:     has,
	}
	return res, nil
}

func resolveConsts(a []string, consts config.Consts) ([]string, error) {
	res := []string{}
	for _, s := range a {
		resolved, ok := consts[s]
		if !ok {
			return nil, fmt.Errorf("failed to resolve const '%s'", s)
		}
		res = append(res, resolved.Values...)
	}
	return res, nil
}

func categoryToSmartLabel(cat config.Category) (string, error) {
	var smartl string
	switch cat {
	case config.CategoryPersonal:
		smartl = SmartLabelPersonal
	case config.CategorySocial:
		smartl = SmartLabelSocial
	case config.CategoryUpdates:
		smartl = SmartLabelNotification
	case config.CategoryForums:
		smartl = SmartLabelGroup
	case config.CategoryPromotions:
		smartl = SmartLabelPromo
	default:
		// TODO: move this to config package
		possib := []string{
			string(config.CategoryPersonal),
			string(config.CategorySocial),
			string(config.CategoryUpdates),
			string(config.CategoryForums),
			string(config.CategoryPromotions),
		}
		return "", fmt.Errorf("unrecognized category '%s' (possible values: %s)",
			cat, strings.Join(possib, ", "))
	}
	return fmt.Sprintf("^smartlabel_%s", smartl), nil
}

func joinOR(a []string) string {
	if len(a) == 0 {
		return ""
	}
	if len(a) == 1 {
		return quote(a[0])
	}
	return fmt.Sprintf("{%s}", strings.Join(quoteStrings(a), " "))
}

func quoteStrings(a []string) []string {
	res := make([]string, len(a))
	for i, s := range a {
		res[i] = quote(s)
	}
	return res
}

func quote(a string) string {
	if strings.ContainsRune(a, ' ') {
		return fmt.Sprintf(`"%s"`, a)
	}
	return a
}

func combineFiltersActions(filters []Property, actions []Property) []Entry {
	// Since only one label is allowed in the exported entries,
	// we have to create a new entry for each label and use the same filters for each of them
	res := []Entry{}
	curr := copyPropertiesToEntry(filters)
	countLabels := 0
	for _, action := range actions {
		if action.Name == PropertyApplyLabel {
			countLabels++
			if countLabels > 1 {
				// Append the current entry and start with a fresh one
				res = append(res, curr)
				curr = copyPropertiesToEntry(filters)
			}
			countLabels = 1
		}
		curr = append(curr, action)
	}
	// Append the last entry
	res = append(res, curr)

	return res
}

func copyPropertiesToEntry(p []Property) Entry {
	cp := make([]Property, len(p))
	copy(cp, p)
	return Entry(cp)
}
