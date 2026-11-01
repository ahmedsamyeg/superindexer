package connection

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/usenet/client"
)

// ConnectionManager manages connections to NNTP servers
type ConnectionManager struct {
	// Configuration
	maxConnectionsPerServer int
	timeout                 time.Duration
	maxRetries              int
	
	// Connection pools
	pools map[int64]*ConnectionPool
	
	mu sync.RWMutex
}

// ConnectionPool holds connections to a specific NNTP server
type ConnectionPool struct {
	serverID       int64
	connections    chan *client.NNTPClient
	maxConnections int
	mu             sync.Mutex
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(maxConnectionsPerServer int, timeout time.Duration, maxRetries int) *ConnectionManager {
	return &ConnectionManager{
		maxConnectionsPerServer: maxConnectionsPerServer,
		timeout:                 timeout,
		maxRetries:              maxRetries,
		pools:                   make(map[int64]*ConnectionPool),
	}
}

// GetConnection retrieves a connection from the pool or creates one if needed
func (cm *ConnectionManager) GetConnection(ctx context.Context, server *models.NNTPServer) (*client.NNTPClient, error) {
	cm.mu.RLock()
	pool, exists := cm.pools[server.ID]
	cm.mu.RUnlock()

	if !exists {
		cm.mu.Lock()
		// Double-check to avoid race conditions
		if pool, exists = cm.pools[server.ID]; !exists {
			pool = &ConnectionPool{
				serverID:       server.ID,
				connections:    make(chan *client.NNTPClient, cm.maxConnectionsPerServer),
				maxConnections: cm.maxConnectionsPerServer,
			}
			cm.pools[server.ID] = pool
		}
		cm.mu.Unlock()
	}

	// Try to get an existing connection from the pool
	select {
	case conn := <-pool.connections:
		// Validate that the connection is still valid
		if conn.IsConnected() {
			return conn, nil
		}
		// If connection is not valid, fall through to create new one
	default:
		// Pool is empty, check if we can create a new connection
	}

	// Create a new client and connect to the server
	client := client.NewNNTPClient(server, cm.timeout, cm.maxRetries)
	err := client.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NNTP server %s: %w", server.Hostname, err)
	}

	return client, nil
}

// ReleaseConnection returns a connection back to the pool
func (cm *ConnectionManager) ReleaseConnection(serverID int64, client *client.NNTPClient) {
	cm.mu.RLock()
	pool, exists := cm.pools[serverID]
	cm.mu.RUnlock()

	if !exists {
		return
	}

	// Try to put the connection back into the pool
	select {
	case pool.connections <- client:
		// Connection returned to pool
	default:
		// Pool is full, close the connection
		client.Disconnect()
	}
}

// CloseServerConnections closes all connections for a specific server
func (cm *ConnectionManager) CloseServerConnections(serverID int64) {
	cm.mu.RLock()
	pool, exists := cm.pools[serverID]
	cm.mu.RUnlock()

	if !exists {
		return
	}

	// Close all connections in the pool
	close(pool.connections)
	
	cm.mu.Lock()
	delete(cm.pools, serverID)
	cm.mu.Unlock()
}

// CloseAll closes all connections and cleans up the manager
func (cm *ConnectionManager) CloseAll() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for serverID, pool := range cm.pools {
		close(pool.connections)
		delete(cm.pools, serverID)
	}
}

// GetServerConnectionCount returns the number of active connections for a server
func (cm *ConnectionManager) GetServerConnectionCount(serverID int64) int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	pool, exists := cm.pools[serverID]
	if !exists {
		return 0
	}

	return len(pool.connections)
}