package parser

const maxSimplifyPasses = 4

// Logical operations.
const (
	OperationNone Operation = iota
	OperationAND
	OperationOR
	OperationNot
)

// Operation is the type of logical operator.
type Operation int

// Functions.
const (
	FunctionNone FunctionType = iota
	FunctionFrom
	FunctionTo
	FunctionCc
	FunctionSubject
	FunctionList
	FunctionHas
)

// FunctionType is the type of a function.
type FunctionType int

// CriteriaAST is the abstract syntax tree of a filter criteria.
type CriteriaAST interface{}

// Node is an AST node with children nodes. It can only be a logical operator.
type Node struct {
	Operation Operation
	Children  []CriteriaAST
}

// Leaf is an AST node with no children.
//
// If the function has multiple arguments, they are grouped together with a
// logical operator. For example: from:{a b} has two arguments grouped with
// an OR and from:(a b) is grouped with an AND.
type Leaf struct {
	Function FunctionType
	Grouping Operation
	Args     []string
}

// SimplifyCriteria applies multiple simplifications to a criteria.
func SimplifyCriteria(tree CriteriaAST) ([]CriteriaAST, error) {
	// Do not start from zero, otherwise we would be done immediately
	changes := 1
	for i := 0; changes > 0 && i < maxSimplifyPasses; i++ {
		changes = logicalGrouping(tree)
		changes += functionsGrouping(tree)
		var c int
		tree, c = removeRedundancy(tree)
		changes += c
	}
	return splitRootOr(tree), nil
}

func logicalGrouping(tree CriteriaAST) int {
	root, ok := tree.(*Node)
	if !ok {
		// Leaves don't apply
		return 0
	}

	// Recurse to children first.
	count := 0
	for _, child := range root.Children {
		count += logicalGrouping(child)
	}

	// The not operator does not apply.
	if root.Operation == OperationNot {
		return count
	}

	// Try to find child nodes with my same operation and squash them.
	//
	// Example:
	// and(foo, and(bar, baz), quax) => and(foo, bar, baz, quax)
	newChildren := []CriteriaAST{}
	for _, child := range root.Children {
		childNode, ok := child.(*Node)
		if !ok || childNode.Operation != root.Operation {
			// Nothing to simplify here
			newChildren = append(newChildren, child)
			continue
		}
		// The operation of the child is the same, get rid of it
		// and add its children here.
		newChildren = append(newChildren, childNode.Children...)
		count++
	}
	root.Children = newChildren

	return count
}

func functionsGrouping(tree CriteriaAST) int {
	root, ok := tree.(*Node)
	if !ok {
		// Leaves don't apply
		return 0
	}

	// Recurse to children first.
	count := 0
	for _, child := range root.Children {
		count += functionsGrouping(child)
	}

	// If there's only one child, then there's no need to proceed.
	if len(root.Children) <= 1 {
		return count
	}

	// Group leaf nodes together and squash them.
	//
	// Example:
	// and(foo:x bar:y foo:z) => and(foo:(x z) bar:z)
	newChildren := []CriteriaAST{}
	functions := map[FunctionType][]string{}
	for _, child := range root.Children {
		leaf, ok := child.(*Leaf)
		if !ok || (len(leaf.Args) > 1 && leaf.Grouping != root.Operation) {
			// Non-leaves and leaves grouped by a different operator have to
			// stay as-is.
			newChildren = append(newChildren, child)
			continue
		}
		functions[leaf.Function] = append(functions[leaf.Function], leaf.Args...)
	}

	// Re-construct the grouped children
	for ft, args := range functions {
		newChildren = append(newChildren, &Leaf{
			Function: ft,
			Grouping: root.Operation,
			Args:     args,
		})
		count++
	}

	root.Children = newChildren

	return count
}

func removeRedundancy(tree CriteriaAST) (CriteriaAST, int) {
	root, ok := tree.(*Node)
	if !ok {
		// Leaves don't apply
		return tree, 0
	}

	// Recurse to children first.
	count := 0
	newChildren := []CriteriaAST{}
	for _, child := range root.Children {
		newChild, c := removeRedundancy(child)
		count += c
		newChildren = append(newChildren, newChild)
	}
	root.Children = newChildren

	if root.Operation == OperationNot {
		newRoot, c := simplifyNot(root)
		return newRoot, count + c
	}

	// All good, this operator is useful
	if len(root.Children) != 1 {
		// Side note: this catches also the case where we have no children
		// which means the tree is invalid, but at least we don't crash.
		return root, count
	}

	// A single child means this operator is not useful
	//
	// Example:
	// or(a) => a
	return root.Children[0], count + 1
}

func simplifyNot(root *Node) (CriteriaAST, int) {
	// If the child is another 'not', we can get rid of both.
	if len(root.Children) != 1 {
		// Something is wrong here: let's just return the tree as is
		return root, 0
	}

	child, ok := root.Children[0].(*Node)
	if !ok || child.Operation != OperationNot {
		return root, 0
	}

	if len(child.Children) != 1 {
		// Something is wrong here: let's just return the tree as is
		return root, 0
	}

	return child.Children[0], 1
}

func splitRootOr(tree CriteriaAST) []CriteriaAST {
	root, ok := tree.(*Node)
	if !ok || root.Operation != OperationOR {
		return []CriteriaAST{tree}
	}
	return root.Children
}
