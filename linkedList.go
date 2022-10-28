package main

type Node struct {
	prev *Node
	next *Node
}

type List struct {
	head *Node
	tail *Node
}