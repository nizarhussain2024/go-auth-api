package main

import (
	"sync"
	"time"
)

type AuditLog struct {
	UserID    string
	Action    string
	Resource  string
	IPAddress string
	Timestamp time.Time
	Details   map[string]interface{}
}

type AuditLogger struct {
	mu    sync.RWMutex
	logs  []*AuditLog
	maxLogs int
}

var auditLogger = &AuditLogger{
	logs:    make([]*AuditLog, 0),
	maxLogs: 1000,
}

func (al *AuditLogger) Log(userID, action, resource, ipAddress string, details map[string]interface{}) {
	al.mu.Lock()
	defer al.mu.Unlock()
	
	log := &AuditLog{
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		IPAddress: ipAddress,
		Timestamp: time.Now(),
		Details:   details,
	}
	
	al.logs = append(al.logs, log)
	
	// Keep only last maxLogs entries
	if len(al.logs) > al.maxLogs {
		al.logs = al.logs[len(al.logs)-al.maxLogs:]
	}
}

func (al *AuditLogger) GetLogs(userID string, limit int) []*AuditLog {
	al.mu.RLock()
	defer al.mu.RUnlock()
	
	var filtered []*AuditLog
	for i := len(al.logs) - 1; i >= 0 && len(filtered) < limit; i-- {
		if al.logs[i].UserID == userID {
			filtered = append(filtered, al.logs[i])
		}
	}
	
	return filtered
}


