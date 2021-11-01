package v1alpha1

import "fmt"

// ResolveConsts returns a copy of the config with all the constants
// replaced in the filters.
func ResolveConsts(c Config) (Config, error) {
	// Don't touch the original, copy the rules
	var rules []Rule
	for i, r := range c.Rules {
		f, err := resolveFilters(r.Filters, c.Consts)
		if err != nil {
			return c, fmt.Errorf("in rule #%d: %w", i, err)
		}
		rules = append(rules, Rule{f, r.Actions})
	}

	c.Rules = rules
	// Get rid of the constants
	c.Consts = Consts{}
	return c, nil
}

func resolveFilters(f Filters, consts Consts) (Filters, error) {
	var res Filters

	// Resolve the consts
	cm, err := resolveFiltersConsts(f.Consts.MatchFilters, consts)
	if err != nil {
		return res, err
	}
	ncm, err := resolveFiltersConsts(f.Consts.Not, consts)
	if err != nil {
		return res, err
	}

	// Join the non const configuration with the resolved one
	res.MatchFilters = joinMatchFilters(f.MatchFilters, cm)
	res.Not = joinMatchFilters(f.Not, ncm)

	return res, nil
}

func resolveFiltersConsts(mf MatchFilters, consts Consts) (MatchFilters, error) {
	from, err := resolveConsts(mf.From, consts)
	if err != nil {
		return mf, fmt.Errorf("resolving 'from' clause: %w", err)
	}
	to, err := resolveConsts(mf.To, consts)
	if err != nil {
		return mf, fmt.Errorf("resolving 'to' clause: %w", err)
	}
	cc, err := resolveConsts(mf.Cc, consts)
	if err != nil {
		return mf, fmt.Errorf("resolving 'cc' clause: %w", err)
	}
	sub, err := resolveConsts(mf.Subject, consts)
	if err != nil {
		return mf, fmt.Errorf("resolving 'subject' clause: %w", err)
	}
	has, err := resolveConsts(mf.Has, consts)
	if err != nil {
		return mf, fmt.Errorf("resolving 'has' clause: %w", err)
	}
	list, err := resolveConsts(mf.List, consts)
	if err != nil {
		return mf, fmt.Errorf("resolving 'list' clause: %w", err)
	}
	res := MatchFilters{
		From:    from,
		To:      to,
		Cc:      cc,
		Subject: sub,
		Has:     has,
		List:    list,
	}
	return res, nil
}

func resolveConsts(a []string, consts Consts) ([]string, error) {
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

func joinMatchFilters(f1, f2 MatchFilters) MatchFilters {
	res := MatchFilters{}
	res.From = joinFilter(f1.From, f2.From)
	res.To = joinFilter(f1.To, f2.To)
	res.Cc = joinFilter(f1.Cc, f2.Cc)
	res.Subject = joinFilter(f1.Subject, f2.Subject)
	res.Has = joinFilter(f1.Has, f2.Has)
	res.List = joinFilter(f1.List, f2.List)
	return res
}

func joinFilter(f1, f2 []string) []string {
	res := []string{}
	res = append(res, f1...)
	res = append(res, f2...)
	return res
}
