package parser

import "sort"

const maxSimplifyPasses = 4

// Logical operations.
const (
	OperationNone OperationType = iota
	OperationAnd
	OperationOr
	OperationNot
)

// OperationType is the type of logical operator.
type OperationType int

func (t OperationType) String() string {
	switch t {
	case OperationNone:
		return "<none>"
	case OperationAnd:
		return "and"
	case OperationOr:
		return "or"
	default:
		return "<unknown>"
	}
}

// Functions.
const (
	FunctionNone FunctionType = iota
	FunctionFrom
	FunctionTo
	FunctionCc
	FunctionBcc
	FunctionReplyTo
	FunctionSubject
	FunctionList
	FunctionHas
	FunctionQuery
)

// FunctionType is the type of a function.
type FunctionType int

func (f FunctionType) String() string {
	switch f {
	case FunctionNone:
		return "<none>"
	case FunctionFrom:
		return "from"
	case FunctionTo:
		return "to"
	case FunctionCc:
		return "cc"
	case FunctionBcc:
		return "bcc"
	case FunctionReplyTo:
		return "replyto"
	case FunctionSubject:
		return "subject"
	case FunctionList:
		return "list"
	case FunctionHas:
		return "has"
	case FunctionQuery:
		return "query"
	default:
		return "<unknown>"
	}
}

// CriteriaAST is the abstract syntax tree of a filter criteria.
type CriteriaAST interface {
	// RootOperation returns the operation performed by the root node,
	// or the grouping, if the root is a leaf.
	RootOperation() OperationType
	// RootOperation returns the function performed by the root node,
	// if any.
	RootFunction() FunctionType
	// IsLeaf returns true if the root node is a leaf.
	IsLeaf() bool
	// AcceptVisitor implements the visitor pattern.
	AcceptVisitor(v Visitor)
	// Clone returns a deep copy of the tree.
	Clone() CriteriaAST
}

// Node is an AST node with children nodes. It can only be a logical operator.
type Node struct {
	Operation OperationType
	Children  []CriteriaAST
}

// RootOperation returns the operation performed by the root node.
func (n *Node) RootOperation() OperationType {
	return n.Operation
}

// RootFunction will always return 'none'.
func (n *Node) RootFunction() FunctionType {
	return FunctionNone
}

// IsLeaf will always return false.
func (n *Node) IsLeaf() bool {
	return false
}

// AcceptVisitor implements the visitor pattern.
func (n *Node) AcceptVisitor(v Visitor) {
	v.VisitNode(n)
}

// Clone returns a deep copy of the tree.
func (n *Node) Clone() CriteriaAST {
	var children []CriteriaAST
	for _, c := range n.Children {
		children = append(children, c.Clone())
	}
	return &Node{
		Operation: n.Operation,
		Children:  children,
	}
}

// Leaf is an AST node with no children.
//
// If the function has multiple arguments, they are grouped together with a
// logical operator. For example: from:{a b} has two arguments grouped with
// an OR and from:(a b) is grouped with an AND.
type Leaf struct {
	Function FunctionType
	Grouping OperationType
	Args     []string
	IsRaw    bool
}

// RootOperation returns the grouping of the leaf.
func (n *Leaf) RootOperation() OperationType {
	return n.Grouping
}

// RootFunction returns the function of the leaf.
func (n *Leaf) RootFunction() FunctionType {
	return n.Function
}

// IsLeaf will always return true.
func (n *Leaf) IsLeaf() bool {
	return true
}

// AcceptVisitor implements the visitor pattern.
func (n *Leaf) AcceptVisitor(v Visitor) {
	v.VisitLeaf(n)
}

// Clone returns a deep copy of the leaf node.
func (n *Leaf) Clone() CriteriaAST {
	return &Leaf{
		Function: n.Function,
		Grouping: n.Grouping,
		Args:     n.Args,
		IsRaw:    n.IsRaw,
	}
}

// Visitor implements the visitor pattern for CriteriaAST.
type Visitor interface {
	VisitNode(n *Node)
	VisitLeaf(n *Leaf)
}

// SimplifyCriteria applies multiple simplifications to a criteria.
func SimplifyCriteria(tree CriteriaAST) (CriteriaAST, error) {
	res, err := simplify(tree)
	// We use maps, so the resulting tree is non-deterministic.
	// To fix that we sort the trees.
	sortTree(res)
	return res, err
}

func simplify(tree CriteriaAST) (CriteriaAST, error) {
	changes := 1 // Avoid stopping before the first round

	// We want to apply the passes multiple times, because one
	// simplification can unlock another. We can block when no
	// further progress can be made.
	for i := 0; changes > 0 && i < maxSimplifyPasses; i++ {
		changes = logicalGrouping(tree)
		changes += functionsGrouping(tree)
		newTree, c := removeRedundancy(tree)
		changes += c
		tree = newTree
	}

	return tree, nil
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
	rawFunctions := map[FunctionType]bool{}
	for _, child := range root.Children {
		leaf, ok := child.(*Leaf)
		if !ok || (len(leaf.Args) > 1 && leaf.Grouping != root.Operation) {
			// Non-leaves and leaves grouped by a different operator have to
			// stay as-is.
			newChildren = append(newChildren, child)
			continue
		}
		functions[leaf.Function] = append(functions[leaf.Function], leaf.Args...)
		// When grouping preserve the 'raw' modifier.
		if leaf.IsRaw {
			rawFunctions[leaf.Function] = true
		}
	}

	// Re-construct the grouped children
	for ft, args := range functions {
		_, raw := rawFunctions[ft]
		newChildren = append(newChildren, &Leaf{
			Function: ft,
			Grouping: root.Operation,
			Args:     args,
			IsRaw:    raw,
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

func sortTreeNodes(nodes []CriteriaAST) {
	// Recurse to individual nodes first
	for _, child := range nodes {
		sortTree(child)
	}

	sort.Slice(nodes, func(i, j int) bool {
		// ordering will be:
		// - leaves in grouping and function order, then
		// - nodes in operation order
		ni, nj := nodes[i], nodes[j]
		if ni.IsLeaf() != nj.IsLeaf() {
			return ni.IsLeaf()
		}
		// They are both either nodes or leaves,
		// we can just compare first the operation and then the function
		if ni.RootOperation() != nj.RootOperation() {
			return ni.RootOperation() < nj.RootOperation()
		}
		return ni.RootFunction() < nj.RootFunction()
	})
}

func sortTree(tree CriteriaAST) {
	if root, ok := tree.(*Node); ok {
		// Sort children recursively
		sortTreeNodes(root.Children)
	}
}
