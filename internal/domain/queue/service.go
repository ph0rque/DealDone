package queue

// Service defines the interface for queue management operations
type Service interface {
	// Queue management
	Start() error
	Stop() error
	EnqueueDocument(dealName, documentPath, documentName string, priority ProcessingPriority, metadata map[string]interface{}) (*QueueItem, error)
	DequeueDocument() (*QueueItem, error)
	GetQueueStatus() (*QueueStatus, error)
	QueryQueue(filter *QueueFilter) ([]*QueueItem, error)
	UpdateItemStatus(itemID string, status QueueItemStatus, metadata map[string]interface{}) error
	RemoveItem(itemID string) error

	// Queue persistence
	SaveQueue() error
	LoadQueue() error
}
