package domain

type ListNode struct {
	Val        Key
	Key        Key
	Next, Prev *ListNode
}

type FreqListNode struct {
	*ListNode
	Freq       int
	Next, Prev *FreqListNode
}
