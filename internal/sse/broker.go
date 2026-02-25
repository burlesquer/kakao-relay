package sse

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	redisclient "github.com/openclaw/relay-server-go/internal/redis"
)

const (
	HeartbeatInterval = 30 * time.Second
)

type Event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Client struct {
	AccountID string
	Events    chan Event
	Done      chan struct{}
}

type Broker struct {
	redis   *redisclient.Client
	clients map[string]map[*Client]bool // accountID -> set of clients
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewBroker(redisClient *redisclient.Client) *Broker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Broker{
		redis:   redisClient,
		clients: make(map[string]map[*Client]bool),
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (b *Broker) Subscribe(accountID string) *Client {
	client := &Client{
		AccountID: accountID,
		Events:    make(chan Event, 100),
		Done:      make(chan struct{}),
	}

	b.mu.Lock()
	if b.clients[accountID] == nil {
		b.clients[accountID] = make(map[*Client]bool)
		go b.subscribeToRedis(accountID)
	}
	b.clients[accountID][client] = true
	clientCount := len(b.clients[accountID])
	b.mu.Unlock()

	log.Info().
		Str("accountId", accountID).
		Int("clientCount", clientCount).
		Msg("sse client subscribed")

	return client
}

func (b *Broker) Unsubscribe(client *Client) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if clients, ok := b.clients[client.AccountID]; ok {
		delete(clients, client)
		close(client.Done)

		if len(clients) == 0 {
			delete(b.clients, client.AccountID)
		}

		log.Info().
			Str("accountId", client.AccountID).
			Int("clientCount", len(clients)).
			Msg("sse client unsubscribed")
	}
}

func (b *Broker) Publish(ctx context.Context, accountID string, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	channel := redisclient.MessageChannel(accountID)
	return b.redis.Publish(ctx, channel, data).Err()
}

func (b *Broker) subscribeToRedis(accountID string) {
	channel := redisclient.MessageChannel(accountID)
	pubsub := b.redis.Subscribe(b.ctx, channel)
	defer pubsub.Close()

	log.Debug().
		Str("accountId", accountID).
		Str("channel", channel).
		Msg("redis pubsub subscribed")

	ch := pubsub.Channel()

	for {
		select {
		case <-b.ctx.Done():
			return

		case msg, ok := <-ch:
			if !ok {
				return
			}

			var event Event
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				log.Error().Err(err).Msg("failed to unmarshal event")
				continue
			}

			b.broadcast(accountID, event)
		}
	}
}

func (b *Broker) broadcast(accountID string, event Event) {
	b.mu.RLock()
	clients := b.clients[accountID]
	b.mu.RUnlock()

	for client := range clients {
		select {
		case client.Events <- event:
		default:
			log.Warn().
				Str("accountId", accountID).
				Msg("client event buffer full, dropping event")
		}
	}
}

func (b *Broker) Close() {
	b.cancel()

	b.mu.Lock()
	defer b.mu.Unlock()

	for _, clients := range b.clients {
		for client := range clients {
			close(client.Done)
		}
	}
	b.clients = make(map[string]map[*Client]bool)
}

func (b *Broker) ClientCount(accountID string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients[accountID])
}

func (b *Broker) TotalClients() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	total := 0
	for _, clients := range b.clients {
		total += len(clients)
	}
	return total
}
