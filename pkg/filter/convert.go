package filter

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/mbrt/gmailctl/pkg/parser"
)

// FromRules translates rules into entries that map directly into Gmail filters.
func FromRules(rs []parser.Rule) (Filters, error) {
	res := Filters{}
	for i, rule := range rs {
		filters, err := fromRule(rule)
		if err != nil {
			return res, errors.Wrap(err, fmt.Sprintf("error generating rule #%d", i))
		}
		res = append(res, filters...)
	}
	return res, nil
}

func fromRule(rule parser.Rule) ([]Filter, error) {
	criteria, err := generateCriteria(rule.Criteria)
	if err != nil {
		return nil, errors.Wrap(err, "error generating criteria")
	}
	actions, err := generateActions(rule.Actions)
	if err != nil {
		return nil, errors.Wrap(err, "error generating actions")
	}
	return combineCriteriaWithActions(criteria, actions), nil
}

func generateCriteria(crit parser.CriteriaAST) (Criteria, error) {
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
			crit, err := generateCriteria(child)
			if err != nil {
				return res, err
			}
			res = joinCriteria(res, crit)

		}
		return res, nil

	case parser.OperationNot:
		if ln := len(node.Children); ln != 1 {
			return Criteria{}, errors.Errorf("after 'not' got %d children, expected 1", ln)
		}
		cq, err := generateCriteriaAsString(node.Children[0])
		return Criteria{
			Query: fmt.Sprintf("-%s", cq),
		}, err
	}

	return Criteria{}, errors.Errorf("unknown node operation %d", node.Operation)
}

func generateLeaf(leaf *parser.Leaf) (Criteria, error) {
	needEscape := leaf.Function != parser.FunctionQuery
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
	case parser.FunctionList:
		return Criteria{
			Query: fmt.Sprintf("list:%s", query),
		}, nil
	case parser.FunctionHas, parser.FunctionQuery:
		return Criteria{
			Query: query,
		}, nil
	default:
		return Criteria{}, errors.Errorf("unknown function type %d", leaf.Function)
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
	needEscape := leaf.Function != parser.FunctionQuery
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
		return "", errors.Errorf("unknown node operation %d", op)
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
		return joinEscaped(a...)
	}
	return strings.Join(a, " ")
}

func joinEscaped(a ...string) string {
	return strings.Join(escapeStrings(a...), " ")
}

func escapeStrings(a ...string) []string {
	res := make([]string, len(a))
	for i, s := range a {
		res[i] = escape(s)
	}
	return res
}

func escape(a string) string {
	if strings.ContainsAny(a, " \t{}()") {
		return fmt.Sprintf(`"%s"`, a)
	}
	return a
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
		},
	}

	if fromOptionalBool(actions.MarkSpam, true) {
		return nil, errors.New("Gmail filters don't allow to send messages to spam directly")
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

func combineCriteriaWithActions(criteria Criteria, actions []Actions) Filters {
	// We have to duplicate the criteria for all the given actions
	res := make(Filters, len(actions))
	for i, action := range actions {
		res[i] = Filter{
			Criteria: criteria,
			Action:   action,
		}
	}
	return res
}
