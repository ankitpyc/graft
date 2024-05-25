package domain

// ListNode represents a node in a linked list.
type ListNode struct {
	Val  string    // Val represents the value stored in the node.
	Key  string    // Key represents the key associated with the value.
	Next *ListNode // Next points to the next node in the linked list.
	Prev *ListNode // Prev points to the previous node in the linked list.
}

// FreqListNode represents a node in a frequency list, extending ListNode with frequency information.
type FreqListNode struct {
	*ListNode               // Embeds ListNode for basic linked list functionality.
	Freq      int           // Freq represents the frequency of the associated key.
	Next      *FreqListNode // Next points to the next node in the frequency list.
	Prev      *FreqListNode // Prev points to the previous node in the frequency list.
}
