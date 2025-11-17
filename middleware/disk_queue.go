package middleware

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
)

// DiskQueue provides persistent queue storage for telemetry data when database is unavailable
type DiskQueue struct {
	basePath    string
	segmentSize  int64 // Maximum size per segment file
	currentSize  int64
	currentFile  *os.File
	mu          sync.RWMutex
	closed      bool
}

// QueueEntry represents a queued entry
type QueueEntry struct {
	Timestamp time.Time                 `json:"timestamp"`
	Log       *model.HeimdallRequestLog `json:"log"`
}

// NewDiskQueue creates a new disk queue
func NewDiskQueue(basePath string) *DiskQueue {
	if basePath == "" {
		basePath = "/tmp/heimdall_queue"
	}
	
	queue := &DiskQueue{
		basePath:   basePath,
		segmentSize: 100 * 1024 * 1024, // 100MB per segment
	}
	
	// Ensure directory exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		logger.SysLog(fmt.Sprintf("Failed to create disk queue directory: %v", err))
		return nil
	}
	
	// Initialize or recover queue
	queue.recover()
	
	return queue
}

// recover recovers entries from disk queue
func (dq *DiskQueue) recover() {
	files, err := filepath.Glob(filepath.Join(dq.basePath, "*.queue"))
	if err != nil {
		logger.SysLog(fmt.Sprintf("Failed to read queue files: %v", err))
		return
	}
	
	if len(files) == 0 {
		logger.SysLog("No queue files found, starting fresh")
		return
	}
	
	// Find the most recent file to continue from
	var latestFile string
	var latestTime time.Time
	
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		if info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
			latestFile = file
		}
	}
	
	if latestFile != "" {
		logger.SysLog(fmt.Sprintf("Recovering from queue file: %s", latestFile))
		dq.currentFile, err = os.OpenFile(latestFile, os.O_APPEND|os.O_RDWR, 0644)
		if err != nil {
			logger.SysLog(fmt.Sprintf("Failed to open queue file: %v", err))
			return
		}
		
		// Get current file size
		if stat, err := dq.currentFile.Stat(); err == nil {
			dq.currentSize = stat.Size()
		}
	}
}

// Enqueue adds an entry to the disk queue
func (dq *DiskQueue) Enqueue(log *model.HeimdallRequestLog) error {
	if dq == nil || dq.closed {
		return fmt.Errorf("disk queue is not available")
	}
	
	dq.mu.Lock()
	defer dq.mu.Unlock()
	
	entry := QueueEntry{
		Timestamp: time.Now(),
		Log:       log,
	}
	
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal queue entry: %w", err)
	}
	
	// Check if we need to rotate the file
	if dq.currentSize > dq.segmentSize || dq.currentFile == nil {
		if err := dq.rotateFile(); err != nil {
			return fmt.Errorf("failed to rotate queue file: %w", err)
		}
	}
	
	// Write to file with newline separator
	data = append(data, '\n')
	n, err := dq.currentFile.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to queue file: %w", err)
	}
	
	dq.currentSize += int64(n)
	
	// Sync to disk for durability
	if err := dq.currentFile.Sync(); err != nil {
		logger.SysLog(fmt.Sprintf("Failed to sync queue file: %v", err))
	}
	
	return nil
}

// DequeueBatch retrieves a batch of entries from the disk queue
func (dq *DiskQueue) DequeueBatch(batchSize int) ([]*model.HeimdallRequestLog, error) {
	if dq == nil || dq.closed {
		return nil, fmt.Errorf("disk queue is not available")
	}
	
	dq.mu.Lock()
	defer dq.mu.Unlock()
	
	if dq.currentFile == nil {
		return nil, nil
	}
	
	var entries []*model.HeimdallRequestLog
	decoder := json.NewDecoder(dq.currentFile)
	
	// Read file from beginning
	if _, err := dq.currentFile.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to seek to beginning of queue file: %w", err)
	}
	
	// Read entries line by line
	for i := 0; i < batchSize; i++ {
		var entry QueueEntry
		if err := decoder.Decode(&entry); err != nil {
			if err.Error() == "EOF" {
				break
			}
			logger.SysLog(fmt.Sprintf("Failed to decode queue entry: %v", err))
			continue
		}
		
		// Skip old entries (older than 24 hours)
		if time.Since(entry.Timestamp) > 24*time.Hour {
			continue
		}
		
		entries = append(entries, entry.Log)
	}
	
	// If we read some entries, truncate the file
	if len(entries) > 0 {
		if err := dq.truncateAndRewrite(decoder); err != nil {
			return entries, fmt.Errorf("failed to truncate queue file: %w", err)
		}
	}
	
	return entries, nil
}

// rotateFile creates a new queue file
func (dq *DiskQueue) rotateFile() error {
	// Close current file
	if dq.currentFile != nil {
		dq.currentFile.Close()
	}
	
	// Create new file with timestamp
	filename := fmt.Sprintf("heimdall_queue_%s.queue", time.Now().Format("20060102_150405"))
	filepath := filepath.Join(dq.basePath, filename)
	
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	
	dq.currentFile = file
	dq.currentSize = 0
	
	return nil
}

// truncateAndRewrite truncates the file and rewrites remaining entries
func (dq *DiskQueue) truncateAndRewrite(decoder *json.Decoder) error {
	// Create temporary file for remaining entries
	tempFile, err := os.CreateTemp(dq.basePath, "temp_queue_*.tmp")
	if err != nil {
		return err
	}
	defer tempFile.Close()
	
	// Read remaining entries and write to temp file
	remainingCount := 0
	for {
		var entry QueueEntry
		if err := decoder.Decode(&entry); err != nil {
			if err.Error() == "EOF" {
				break
			}
			continue
		}
		
		// Skip old entries
		if time.Since(entry.Timestamp) > 24*time.Hour {
			continue
		}
		
		data, err := json.Marshal(entry)
		if err != nil {
			continue
		}
		data = append(data, '\n')
		
		if _, err := tempFile.Write(data); err != nil {
			return err
		}
		remainingCount++
	}
	
	// Sync temp file
	if err := tempFile.Sync(); err != nil {
		return err
	}
	
	// Get current file path
	currentPath := dq.currentFile.Name()
	
	// Close current file
	dq.currentFile.Close()
	
	// Replace current file with temp file
	if err := os.Rename(tempFile.Name(), currentPath); err != nil {
		return err
	}
	
	// Reopen the file
	dq.currentFile, err = os.OpenFile(currentPath, os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	
	// Update current size
	if stat, err := dq.currentFile.Stat(); err == nil {
		dq.currentSize = stat.Size()
	}
	
	logger.SysLog(fmt.Sprintf("Truncated queue file, %d entries remaining", remainingCount))
	
	return nil
}

// Size returns the current size of the queue
func (dq *DiskQueue) Size() int64 {
	if dq == nil || dq.closed {
		return 0
	}
	
	dq.mu.RLock()
	defer dq.mu.RUnlock()
	
	return dq.currentSize
}

// Close closes the disk queue
func (dq *DiskQueue) Close() error {
	if dq == nil || dq.closed {
		return nil
	}
	
	dq.mu.Lock()
	defer dq.mu.Unlock()
	
	dq.closed = true
	
	if dq.currentFile != nil {
		return dq.currentFile.Close()
	}
	
	return nil
}

// Cleanup removes old queue files
func (dq *DiskQueue) Cleanup() error {
	if dq == nil {
		return nil
	}
	
	files, err := filepath.Glob(filepath.Join(dq.basePath, "*.queue"))
	if err != nil {
		return err
	}
	
	cutoff := time.Now().Add(-24 * time.Hour) // Remove files older than 24 hours
	
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		
		if info.ModTime().Before(cutoff) {
			if err := os.Remove(file); err != nil {
				logger.SysLog(fmt.Sprintf("Failed to remove old queue file %s: %v", file, err))
			} else {
				logger.SysLog(fmt.Sprintf("Removed old queue file: %s", file))
			}
		}
	}
	
	return nil
}

// GetQueueStats returns statistics about the disk queue
func (dq *DiskQueue) GetQueueStats() map[string]interface{} {
	if dq == nil {
		return map[string]interface{}{"available": false}
	}
	
	dq.mu.RLock()
	defer dq.mu.RUnlock()
	
	stats := map[string]interface{}{
		"available":     !dq.closed,
		"current_size":  dq.currentSize,
		"segment_size":  dq.segmentSize,
	}
	
	// Count files
	files, err := filepath.Glob(filepath.Join(dq.basePath, "*.queue"))
	if err == nil {
		stats["file_count"] = len(files)
	}
	
	return stats
}
