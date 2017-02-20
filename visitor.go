package jsonoscope

type (
	// Visitor defines methods that are invoked during the depth-first traversal
	// of a JSON object.
	Visitor interface {
		// Enter is invoked when a node in the JSON tree is first reached. It is
		// supplied with the node's path and type.
		Enter(path string, token Token)

		// Enter is invoked on a node when its subtree (if any) has been fully
		// traversed. Invocations of Enter are matched one-to-one with calls
		// to Exit, and exit is provided the same token and path as Enter was.
		//
		// Exit is additionally provided with the node's signature, or a
		// semantically-deterministic hash of its contents.
		Exit(path string, token Token, signature []byte)
	}

	// CustomVisitor is a Visitor implementation whose functionality
	// can be dynamically configured for convenience.
	CustomVisitor struct {
		OnEnter func(path string, token Token)
		OnExit  func(path string, token Token, signature []byte)
	}

	// CountingVisitor tallies the number of visited nodes in a JSON tree.
	CountingVisitor int
)

// Enter implements the Visitor interface by invoking the OnEnter callback,
// if one exists. Otherwise this method has no effect.
func (cv CustomVisitor) Enter(path string, token Token) {
	if cv.OnEnter != nil {
		cv.OnEnter(path, token)
	}
}

// Exit implements the Visitor interface by invoking the OnExit callback,
// if one exists. Otherwise this method has no effect.
func (cv CustomVisitor) Exit(path string, token Token, signature []byte) {
	if cv.OnExit != nil {
		cv.OnExit(path, token, signature)
	}
}

// Enter implements the Visitor interface. It has no effect.
func (cv *CountingVisitor) Enter(path string, token Token) {
	// no implementation
}

// Exit implements the Visitor interface. It increments the counter by one.
func (cv *CountingVisitor) Exit(path string, token Token, signature []byte) {
	*cv = *cv + 1
}

// Nodes returns the count of total visited nodes.
func (cv *CountingVisitor) Nodes() int {
	return int(*cv)
}
