package cache

import "cache/internal/domain"

// DoubleLinkedList represents a doubly linked list data structure.
type DoubleLinkedList struct {
	Head, Tail *domain.ListNode // Head and Tail represent references to the first and last nodes of the list, respectively.
}
