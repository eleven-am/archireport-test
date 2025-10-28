package graphql

import (
	"context"
	"sync"

	"github.com/eleven-am/enclave/ent"
)

type notificationBroker struct {
	mu          sync.RWMutex
	subscribers map[int]map[chan *ent.Notification]struct{}
}

func newNotificationBroker() *notificationBroker {
	return &notificationBroker{
		subscribers: make(map[int]map[chan *ent.Notification]struct{}),
	}
}

func recipientIDFromNotification(notification *ent.Notification) int {
	if notification == nil {
		return 0
	}
	if notification.Edges.Recipient != nil {
		return notification.Edges.Recipient.ID
	}
	if recipient, err := notification.Edges.RecipientOrErr(); err == nil && recipient != nil {
		return recipient.ID
	}
	return 0
}

func (b *notificationBroker) Subscribe(userID int) (<-chan *ent.Notification, func()) {
	ch := make(chan *ent.Notification, 1)
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.subscribers[userID]; !ok {
		b.subscribers[userID] = make(map[chan *ent.Notification]struct{})
	}
	b.subscribers[userID][ch] = struct{}{}
	return ch, func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if subs, ok := b.subscribers[userID]; ok {
			if _, exists := subs[ch]; exists {
				delete(subs, ch)
				close(ch)
				if len(subs) == 0 {
					delete(b.subscribers, userID)
				}
			}
		}
	}
}

func (b *notificationBroker) Publish(_ context.Context, notification *ent.Notification) {
	if notification == nil {
		return
	}
	userID := recipientIDFromNotification(notification)
	if userID == 0 {
		return
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	subs := b.subscribers[userID]
	for ch := range subs {
		select {
		case ch <- notification:
		default:
		}
	}
}
