package parser

func and(children ...CriteriaAST) *Node {
	return &Node{
		Operation: OperationAnd,
		Children:  children,
	}
}

func or(children ...CriteriaAST) *Node {
	return &Node{
		Operation: OperationOr,
		Children:  children,
	}
}

func not(child CriteriaAST) *Node {
	return &Node{
		Operation: OperationNot,
		Children:  []CriteriaAST{child},
	}
}

func fn(ftype FunctionType, op OperationType, args ...string) *Leaf {
	return &Leaf{
		Function: ftype,
		Grouping: op,
		Args:     args,
	}
}

func fn1(ftype FunctionType, arg string) *Leaf {
	return fn(ftype, OperationNone, arg)
}
