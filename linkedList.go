package main

type Node struct {
	prev *Node
	next *Node
	data interface{}
}

type List struct {
	head *Node
	tail *Node
}

func (l *List) Insert(d interface{}) {
	newNode := &Node{data: d, prev: nil, next: nil}
	if l.head == nil {
		l.head = newNode
		l.tail = newNode
	} else {
		l.head.prev = newNode
		newNode.next = l.head
		l.head = newNode
	}
}

func (l *List) Pop() interface{} {
	data := l.head.data

	l.head = l.head.next
	l.head.next.prev = nil

	return data
}