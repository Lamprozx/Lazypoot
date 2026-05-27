package app

import "lazypoot/types"

type NavStack struct {
	stack []types.Screen
}

func newNavStack() NavStack {
	return NavStack{stack: []types.Screen{types.ScreenMainMenu}}
}

func (n *NavStack) Current() types.Screen {
	if len(n.stack) == 0 {
		return types.ScreenMainMenu
	}
	return n.stack[len(n.stack)-1]
}

func (n *NavStack) Push(s types.Screen) {
	n.stack = append(n.stack, s)
}

func (n *NavStack) Pop() bool {
	if len(n.stack) <= 1 {
		return false
	}
	n.stack = n.stack[:len(n.stack)-1]
	return true
}

func (n *NavStack) Reset() {
	n.stack = []types.Screen{types.ScreenMainMenu}
}

func clampCursor(index, max int) int {
	if max <= 0 {
		return 0
	}
	if index < 0 {
		return 0
	}
	if index >= max {
		return max - 1
	}
	return index
}
