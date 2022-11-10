package filter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mbrt/gmailctl/internal/engine/parser"
)

// There's no documented limit on filter size on Gmail, but this educated guess
// is better than nothing.
const defaultSizeLimit = 20

// FromRules translates rules into entries that map directly into Gmail filters.
func FromRules(rs []parser.Rule) (Filters, error) {
	return FromRulesWithLimit(rs, defaultSizeLimit)
}

// FromRulesWithLimit translates rules into entries that map directly into
// Gmail, but uses a custom size limit.
func FromRulesWithLimit(rs []parser.Rule, sizeLimit int) (Filters, error) {
	res := Filters{}
	for i, rule := range rs {
		filters, err := FromRule(rule, sizeLimit)
		if err != nil {
			return res, fmt.Errorf("generating rule #%d: %w", i, err)
		}
		res = append(res, filters...)
	}
	return res, nil
}

// FromRule translates a rule into entries that map directly into Gmail filters.
func FromRule(rule parser.Rule, sizeLimit int) (Filters, error) {
	var crits []Criteria
	for _, c := range splitCriteria(rule.Criteria, sizeLimit) {
		criteria, err := GenerateCriteria(c)
		if err != nil {
			return nil, fmt.Errorf("generating criteria: %w", err)
		}
		crits = append(crits, criteria)
	}

	actions, err := generateActions(rule.Actions)
	if err != nil {
		return nil, fmt.Errorf("generating actions: %w", err)
	}

	return combineCriteriaWithActions(crits, actions), nil
}

// GenerateCriteria translates a rule criteria into an entry that maps
// directly into Gmail filters.
func GenerateCriteria(crit parser.CriteriaAST) (Criteria, error) {
	if node, ok := crit.(*parser.Node); ok {
		return generateNode(node)
	}
	if leaf, ok := crit.(*parser.Leaf); ok {
		return generateLeaf(leaf)
	}
	return Criteria{}, errors.New("found unknown criteria node")
}

func generateNode(node *parser.Node) (Criteria, error) {
	switch node.Operation {
	case parser.OperationOr:
		query := ""
		for _, child := range node.Children {
			cq, err := generateCriteriaAsString(child)
			if err != nil {
				return Criteria{}, err
			}
			query = joinQueries(query, cq)
		}
		return Criteria{
			Query: fmt.Sprintf("{%s}", query),
		}, nil

	case parser.OperationAnd:
		res := Criteria{}
		for _, child := range node.Children {
			crit, err := GenerateCriteria(child)
			if err != nil {
				return res, err
			}
			res = joinCriteria(res, crit)

		}
		return res, nil

	case parser.OperationNot:
		if ln := len(node.Children); ln != 1 {
			return Criteria{}, fmt.Errorf("after 'not' got %d children, expected 1", ln)
		}
		cq, err := generateCriteriaAsString(node.Children[0])
		return Criteria{
			Query: fmt.Sprintf("-%s", cq),
		}, err
	}

	return Criteria{}, fmt.Errorf("unknown node operation %d", node.Operation)
}

func generateLeaf(leaf *parser.Leaf) (Criteria, error) {
	needEscape := leaf.Function != parser.FunctionQuery && !leaf.IsRaw
	query := joinStrings(needEscape, leaf.Args...)
	if len(leaf.Args) > 1 {
		var err error
		if query, err = groupWithOperation(query, leaf.Grouping); err != nil {
			return Criteria{}, err
		}
	}

	switch leaf.Function {
	case parser.FunctionFrom:
		return Criteria{
			From: query,
		}, nil
	case parser.FunctionTo:
		return Criteria{
			To: query,
		}, nil
	case parser.FunctionSubject:
		return Criteria{
			Subject: query,
		}, nil
	case parser.FunctionCc:
		return Criteria{
			Query: fmt.Sprintf("cc:%s", query),
		}, nil
	case parser.FunctionBcc:
		return Criteria{
			Query: fmt.Sprintf("bcc:%s", query),
		}, nil
	case parser.FunctionReplyTo:
		return Criteria{
			Query: fmt.Sprintf("replyto:%s", query),
		}, nil
	case parser.FunctionList:
		return Criteria{
			Query: fmt.Sprintf("list:%s", query),
		}, nil
	case parser.FunctionHas, parser.FunctionQuery:
		return Criteria{
			Query: query,
		}, nil
	default:
		return Criteria{}, fmt.Errorf("unknown function type %d", leaf.Function)
	}
}

func generateCriteriaAsString(crit parser.CriteriaAST) (string, error) {
	if node, ok := crit.(*parser.Node); ok {
		return generateNodeAsString(node)
	}
	if leaf, ok := crit.(*parser.Leaf); ok {
		return generateLeafAsString(leaf)
	}
	return "", errors.New("found unknown criteria node")
}

func generateNodeAsString(node *parser.Node) (string, error) {
	query := ""
	for _, child := range node.Children {
		cq, err := generateCriteriaAsString(child)
		if err != nil {
			return "", err
		}
		query = joinQueries(query, cq)
	}
	return groupWithOperation(query, node.Operation)
}

func generateLeafAsString(leaf *parser.Leaf) (string, error) {
	needEscape := leaf.Function != parser.FunctionQuery && !leaf.IsRaw
	query := joinStrings(needEscape, leaf.Args...)
	if len(leaf.Args) > 1 {
		var err error
		if query, err = groupWithOperation(query, leaf.Grouping); err != nil {
			return "", err
		}
	}

	switch leaf.Function {
	case parser.FunctionHas, parser.FunctionQuery:
		return query, nil
	default:
		return fmt.Sprintf("%v:%s", leaf.Function, query), nil
	}
}

func groupWithOperation(query string, op parser.OperationType) (string, error) {
	switch op {
	case parser.OperationOr:
		return fmt.Sprintf("{%s}", query), nil

	case parser.OperationAnd:
		return fmt.Sprintf("(%s)", query), nil

	case parser.OperationNot:
		return fmt.Sprintf("-%s", query), nil
	default:
		return "", fmt.Errorf("unknown node operation %d", op)
	}
}

func joinCriteria(c1, c2 Criteria) Criteria {
	return Criteria{
		From:    joinQueries(c1.From, c2.From),
		To:      joinQueries(c1.To, c2.To),
		Subject: joinQueries(c1.Subject, c2.Subject),
		Query:   joinQueries(c1.Query, c2.Query),
	}
}

func joinQueries(f1, f2 string) string {
	// No need to escape queries because they are either logical operations
	// or functions.
	if f1 == "" {
		return f2
	}
	if f2 == "" {
		return f1
	}
	return fmt.Sprintf("%s %s", f1, f2)
}

func joinStrings(escape bool, a ...string) string {
	if escape {
		return joinQuoted(a...)
	}
	return strings.Join(a, " ")
}

func joinQuoted(a ...string) string {
	return strings.Join(quoteStrings(a...), " ")
}

func quoteStrings(a ...string) []string {
	res := make([]string, len(a))
	for i, s := range a {
		res[i] = quote(s)
	}
	return res
}

func quote(a string) string {
	if strings.ContainsAny(a, " \t{}()") {
		return fmt.Sprintf(`"%s"`, a)
	}
	// We need to quote the plus sign, _unless_ it's within a full email
	// address. This is necessary because "foo+bar" is considered like
	// "foo OR bar", but "foo+bar@gmail.com" is not.
	if strings.Contains(a, "+") && !strings.Contains(a, "@") {
		return fmt.Sprintf(`"%s"`, a)
	}
	return a
}

func splitCriteria(tree parser.CriteriaAST, limit int) []parser.CriteriaAST {
	var res []parser.CriteriaAST
	for _, c := range splitRootOr(tree) {
		res = append(res, splitBigCriteria(c, limit)...)
	}
	return res
}

type splitVisitor struct {
	limit int
	res   []parser.CriteriaAST
}

func (v *splitVisitor) VisitNode(n *parser.Node) {
	rem := n.Children
	for len(rem) > v.limit {
		v.res = append(v.res, &parser.Node{
			Operation: n.Operation,
			Children:  rem[:v.limit],
		})
		rem = rem[v.limit:]
	}
	// Add the last chunk.
	v.res = append(v.res, &parser.Node{
		Operation: n.Operation,
		Children:  rem,
	})
}

func (v *splitVisitor) VisitLeaf(n *parser.Leaf) {
	rem := n.Args
	for len(rem) > v.limit {
		v.res = append(v.res, &parser.Leaf{
			Function: n.Function,
			Grouping: n.Grouping,
			IsRaw:    n.IsRaw,
			Args:     rem[:v.limit],
		})
		rem = rem[v.limit:]
	}
	// Add the last chunk.
	v.res = append(v.res, &parser.Leaf{
		Function: n.Function,
		Grouping: n.Grouping,
		IsRaw:    n.IsRaw,
		Args:     rem,
	})
}

func splitBigCriteria(tree parser.CriteriaAST, limit int) []parser.CriteriaAST {
	// Gmail filters have a size limit, after which they will silently
	// fail to be applied. To counter that we try to split up filters
	// that are too big.
	if size := countNodes(tree); size < limit {
		// We don't bother with small filters.
		return []parser.CriteriaAST{tree}
	}
	if tree.RootOperation() == parser.OperationOr {
		// If the root operation is OR, we can split.
		sv := splitVisitor{limit: limit}
		tree.AcceptVisitor(&sv)
		return sv.res
	}
	if tree.RootOperation() == parser.OperationAnd {
		// We still have hope to split this and have set of smaller filters.
		// We just need to find the biggest child with an OR at the root and
		// split it this way:
		// ({a, b, c} d) => ({a, b} d), (c d)
		return splitNestedAnd(tree, limit)
	}

	// Nothing we can do about this :(
	return []parser.CriteriaAST{tree}
}

func splitNestedAnd(root parser.CriteriaAST, limit int) []parser.CriteriaAST {
	n, ok := root.(*parser.Node)
	if !ok {
		// There's no nesting, just a single function, so we can't do anything.
		return []parser.CriteriaAST{root}
	}

	// Find the biggest child with the form {a, b, c, d}
	maxChildren := 0
	childID := -1
	for i, c := range n.Children {
		if count := countNodes(c); count > maxChildren && c.RootOperation() == parser.OperationOr {
			childID = i
			maxChildren = count
		}
	}
	if childID < 0 {
		// We couldn't find any child with an OR as root operation.
		return []parser.CriteriaAST{root}
	}
	bigChild := n.Children[childID]

	// Split it up.
	// We want to respect the limit for each new split up filter, but not go
	// over the top and split it completely.
	siblingsSize := countNodes(root) - maxChildren
	newLimit := limit - siblingsSize
	if newLimit < 1 {
		// We have no hope of respecting the limit, but let's do our best
		// and split up the biggest child completely
		newLimit = 1
	}
	sv := splitVisitor{limit: newLimit}
	bigChild.AcceptVisitor(&sv)

	// Combine the split up child with all the rest of the children.
	//
	// Take all the children except the one split up.
	var siblings []parser.CriteriaAST
	for i, c := range n.Children {
		if i == childID {
			continue
		}
		siblings = append(siblings, c)
	}
	// Combine every element of the split up child with all the siblings.
	var res []parser.CriteriaAST
	for _, c := range sv.res {
		res = append(res, &parser.Node{
			Operation: parser.OperationAnd,
			Children: append(
				[]parser.CriteriaAST{c}, clone(siblings)...,
			),
		})
	}

	return res
}

func clone(tl []parser.CriteriaAST) []parser.CriteriaAST {
	var res []parser.CriteriaAST
	for _, n := range tl {
		res = append(res, n.Clone())
	}
	return res
}

type countVisitor struct{ res int }

func (v *countVisitor) VisitNode(n *parser.Node) {
	for _, c := range n.Children {
		c.AcceptVisitor(v)
	}
	v.res++
}
func (v *countVisitor) VisitLeaf(n *parser.Leaf) {
	// Note that in case of l.IsRaw, this number will be imprecise, as
	// there will be multiple operands in the same expression.
	v.res += len(n.Args)
}

func countNodes(tree parser.CriteriaAST) int {
	cv := countVisitor{}
	tree.AcceptVisitor(&cv)
	return cv.res
}

func splitRootOr(tree parser.CriteriaAST) []parser.CriteriaAST {
	// Since Gmail filters are all applied when they match, we can reduce
	// the size of a rule and make it more readable by splitting a single
	// rule where wee have an OR as the top-level operation, with a set of
	// rules, each a child of the original OR.
	//
	// Example: or(from:a to:b list:c) => archive
	// can be rewritten with 3 rules:
	// - from:a => archive
	// - to:b => archive
	// - list:c => archive
	root, ok := tree.(*parser.Node)
	if !ok || root.Operation != parser.OperationOr {
		return []parser.CriteriaAST{tree}
	}
	return root.Children
}

func generateActions(actions parser.Actions) ([]Actions, error) {
	res := []Actions{
		{
			Archive:          actions.Archive,
			Delete:           actions.Delete,
			MarkImportant:    fromOptionalBool(actions.MarkImportant, true),
			MarkNotImportant: fromOptionalBool(actions.MarkImportant, false),
			MarkRead:         actions.MarkRead,
			Category:         actions.Category,
			MarkNotSpam:      fromOptionalBool(actions.MarkSpam, false),
			Star:             actions.Star,
			Forward:          actions.Forward,
		},
	}

	if fromOptionalBool(actions.MarkSpam, true) {
		return nil, errors.New("gmail filters don't allow one to send messages to spam directly")
	}

	if len(actions.Labels) == 0 {
		return res, nil
	}
	// Since every action can contain a single lable only, we might need to
	// produce multiple actions.
	//
	// The first label can stay in the first action
	res[0].AddLabel = actions.Labels[0]

	// The rest of the labels need a separate action
	for _, label := range actions.Labels[1:] {
		res = append(res, Actions{AddLabel: label})
	}

	return res, nil
}

// fromOptionalBool returns the value of the given option if present,
// reversing its value if positive is false.
func fromOptionalBool(opt *bool, positive bool) bool {
	if opt == nil {
		return false
	}
	return *opt == positive
}

func combineCriteriaWithActions(criteria []Criteria, actions []Actions) Filters {
	// We have to make a Cartesian product of criteria and actions
	var res Filters

	for _, c := range criteria {
		for _, a := range actions {
			res = append(res, Filter{
				Criteria: c,
				Action:   a,
			})
		}
	}

	return res
}
