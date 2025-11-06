package middleware

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/QuantumNous/new-api/model"
)

func TestNewDiskQueue(t *testing.T) {
	tempDir := t.TempDir()
	
	queue := NewDiskQueue(tempDir)
	require.NotNil(t, queue)
	assert.False(t, queue.closed)
	assert.Equal(t, int64(0), queue.currentSize)
	assert.Equal(t, int64(100*1024*1024), queue.segmentSize)
	
	// Check if directory was created
	_, err := os.Stat(tempDir)
	assert.NoError(t, err)
}

func TestDiskQueue_Enqueue(t *testing.T) {
	tempDir := t.TempDir()
	queue := NewDiskQueue(tempDir)
	require.NotNil(t, queue)
	
	// Create a test log entry
	log := &model.HeimdallRequestLog{
		RequestId:     "test-123",
		NormalizedURL: "/test",
		HTTPMethod:    "GET",
		HTTPStatus:    200,
		LatencyMs:     100,
		ClientIP:      "192.168.1.1",
	}
	
	// Enqueue the log
	err := queue.Enqueue(log)
	assert.NoError(t, err)
	assert.True(t, queue.currentSize > 0)
	
	// Close queue
	err = queue.Close()
	assert.NoError(t, err)
	assert.True(t, queue.closed)
}

func TestDiskQueue_EnqueueMultiple(t *testing.T) {
	tempDir := t.TempDir()
	queue := NewDiskQueue(tempDir)
	require.NotNil(t, queue)
	
	// Enqueue multiple entries
	for i := 0; i < 10; i++ {
		log := &model.HeimdallRequestLog{
			RequestId:     "test-123",
			NormalizedURL: "/test",
			HTTPMethod:    "GET",
			HTTPStatus:    200,
			LatencyMs:     int64(i * 10),
			ClientIP:      "192.168.1.1",
		}
		
		err := queue.Enqueue(log)
		assert.NoError(t, err)
	}
	
	assert.True(t, queue.currentSize > 0)
	
	queue.Close()
}

func TestDiskQueue_DequeueBatch(t *testing.T) {
	tempDir := t.TempDir()
	queue := NewDiskQueue(tempDir)
	require.NotNil(t, queue)
	
	// Enqueue some entries
	entries := make([]*model.HeimdallRequestLog, 5)
	for i := 0; i < 5; i++ {
		entries[i] = &model.HeimdallRequestLog{
			RequestId:     "test-123",
			NormalizedURL: "/test",
			HTTPMethod:    "GET",
			HTTPStatus:    200,
			LatencyMs:     int64(i * 10),
			ClientIP:      "192.168.1.1",
		}
		
		err := queue.Enqueue(entries[i])
		assert.NoError(t, err)
	}
	
	// Dequeue entries
	dequeueEntries, err := queue.DequeueBatch(3)
	assert.NoError(t, err)
	assert.Len(t, dequeueEntries, 3)
	
	// Verify entries
	for i, entry := range dequeueEntries {
		assert.Equal(t, "test-123", entry.RequestId)
		assert.Equal(t, "/test", entry.NormalizedURL)
		assert.Equal(t, "GET", entry.HTTPMethod)
		assert.Equal(t, 200, entry.HTTPStatus)
		assert.Equal(t, int64(i*10), entry.LatencyMs)
		assert.Equal(t, "192.168.1.1", entry.ClientIP)
	}
	
	// Dequeue remaining entries
	remainingEntries, err := queue.DequeueBatch(10)
	assert.NoError(t, err)
	assert.Len(t, remainingEntries, 2)
	
	queue.Close()
}

func TestDiskQueue_DequeueBatch_Empty(t *testing.T) {
	tempDir := t.TempDir()
	queue := NewDiskQueue(tempDir)
	require.NotNil(t, queue)
	
	// Try to dequeue from empty queue
	entries, err := queue.DequeueBatch(10)
	assert.NoError(t, err)
	assert.Len(t, entries, 0)
	
	queue.Close()
}

func TestDiskQueue_RotateFile(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create a queue with small segment size for testing
	queue := &DiskQueue{
		basePath:   tempDir,
		segmentSize: 100, // Very small segment size
	}
	
	// Initialize queue
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	// Create initial file
	err := queue.rotateFile()
	assert.NoError(t, err)
	assert.NotNil(t, queue.currentFile)
	assert.Equal(t, int64(0), queue.currentSize)
	
	// Get the file path
	initialPath := queue.currentFile.Name()
	
	// Rotate again
	err = queue.rotateFile()
	assert.NoError(t, err)
	
	// Check that a new file was created
	newPath := queue.currentFile.Name()
	assert.NotEqual(t, initialPath, newPath)
	
	// Close files
	queue.currentFile.Close()
}

func TestDiskQueue_Cleanup(t *testing.T) {
	tempDir := t.TempDir()
	queue := NewDiskQueue(tempDir)
	require.NotNil(t, queue)
	
	// Create some old queue files
	oldFile1 := filepath.Join(tempDir, "old_queue_20200101_000000.queue")
	oldFile2 := filepath.Join(tempDir, "old_queue_20200102_000000.queue")
	newFile := filepath.Join(tempDir, "heimdall_queue_20231201_120000.queue")
	
	// Create files with different timestamps
	_, err := os.Create(oldFile1)
	assert.NoError(t, err)
	_, err = os.Create(oldFile2)
	assert.NoError(t, err)
	_, err = os.Create(newFile)
	assert.NoError(t, err)
	
	// Set old file timestamps
	oldTime := time.Now().Add(-25 * time.Hour) // 25 hours ago
	os.Chtimes(oldFile1, oldTime, oldTime)
	os.Chtimes(oldFile2, oldTime, oldTime)
	
	// Run cleanup
	err = queue.Cleanup()
	assert.NoError(t, err)
	
	// Check that old files were removed and new file remains
	_, err = os.Stat(oldFile1)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(oldFile2)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(newFile)
	assert.NoError(t, err)
	
	queue.Close()
}

func TestDiskQueue_GetQueueStats(t *testing.T) {
	tempDir := t.TempDir()
	queue := NewDiskQueue(tempDir)
	require.NotNil(t, queue)
	
	// Get initial stats
	stats := queue.GetQueueStats()
	assert.True(t, stats["available"].(bool))
	assert.Equal(t, int64(0), stats["current_size"])
	assert.Equal(t, int64(100*1024*1024), stats["segment_size"])
	assert.Equal(t, 0, stats["file_count"].(int))
	
	// Enqueue an entry
	log := &model.HeimdallRequestLog{
		RequestId:     "test-123",
		NormalizedURL: "/test",
		HTTPMethod:    "GET",
		HTTPStatus:    200,
		LatencyMs:     100,
		ClientIP:      "192.168.1.1",
	}
	
	err := queue.Enqueue(log)
	assert.NoError(t, err)
	
	// Get updated stats
	stats = queue.GetQueueStats()
	assert.True(t, stats["available"].(bool))
	assert.True(t, stats["current_size"].(int64) > 0)
	assert.Equal(t, int64(100*1024*1024), stats["segment_size"])
	assert.Equal(t, 1, stats["file_count"].(int))
	
	queue.Close()
}

func TestDiskQueue_NilQueue(t *testing.T) {
	var queue *DiskQueue
	
	// Test methods on nil queue
	err := queue.Enqueue(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available")
	
	entries, err := queue.DequeueBatch(10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available")
	assert.Nil(t, entries)
	
	assert.Equal(t, int64(0), queue.Size())
	
	err = queue.Close()
	assert.NoError(t, err)
	
	stats := queue.GetQueueStats()
	assert.False(t, stats["available"].(bool))
}

func TestDiskQueue_ClosedQueue(t *testing.T) {
	tempDir := t.TempDir()
	queue := NewDiskQueue(tempDir)
	require.NotNil(t, queue)
	
	// Close the queue
	err := queue.Close()
	assert.NoError(t, err)
	
	// Try to enqueue after closing
	log := &model.HeimdallRequestLog{
		RequestId: "test-123",
	}
	
	err = queue.Enqueue(log)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available")
	
	// Try to dequeue after closing
	entries, err := queue.DequeueBatch(10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available")
	assert.Nil(t, entries)
}

func TestQueueEntry_Marshal(t *testing.T) {
	log := &model.HeimdallRequestLog{
		RequestId:     "test-123",
		NormalizedURL: "/test",
		HTTPMethod:    "GET",
		HTTPStatus:    200,
		LatencyMs:     100,
		ClientIP:      "192.168.1.1",
	}
	
	entry := QueueEntry{
		Timestamp: time.Now().UTC(),
		Log:       log,
	}
	
	// Test marshaling
	data, err := json.Marshal(entry)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	
	// Test unmarshaling
	var unmarshaled QueueEntry
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	
	assert.Equal(t, entry.RequestId, unmarshaled.RequestId)
	assert.Equal(t, entry.NormalizedURL, unmarshaled.NormalizedURL)
	assert.Equal(t, entry.HTTPMethod, unmarshaled.HTTPMethod)
	assert.Equal(t, entry.HTTPStatus, unmarshaled.HTTPStatus)
	assert.Equal(t, entry.LatencyMs, unmarshaled.LatencyMs)
	assert.Equal(t, entry.ClientIP, unmarshaled.ClientIP)
}

func TestDiskQueue_Recover(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create an existing queue file
	existingFile := filepath.Join(tempDir, "heimdall_queue_20231201_120000.queue")
	file, err := os.Create(existingFile)
	assert.NoError(t, err)
	
	// Write some test data
	entry := QueueEntry{
		Timestamp: time.Now(),
		Log: &model.HeimdallRequestLog{
			RequestId:     "test-recover",
			NormalizedURL: "/test",
			HTTPMethod:    "GET",
			HTTPStatus:    200,
			LatencyMs:     100,
			ClientIP:      "192.168.1.1",
		},
	}
	
	data, err := json.Marshal(entry)
	assert.NoError(t, err)
	data = append(data, '\n')
	
	_, err = file.Write(data)
	assert.NoError(t, err)
	file.Close()
	
	// Create queue and let it recover
	queue := NewDiskQueue(tempDir)
	require.NotNil(t, queue)
	
	// Should have recovered from existing file
	assert.NotNil(t, queue.currentFile)
	
	// Should be able to dequeue the entry
	entries, err := queue.DequeueBatch(1)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	
	if len(entries) > 0 {
		assert.Equal(t, "test-recover", entries[0].RequestId)
	}
	
	queue.Close()
}

// Benchmark tests
func BenchmarkDiskQueue_Enqueue(b *testing.B) {
	tempDir := b.TempDir()
	queue := NewDiskQueue(tempDir)
	
	log := &model.HeimdallRequestLog{
		RequestId:     "test-123",
		NormalizedURL: "/test",
		HTTPMethod:    "GET",
		HTTPStatus:    200,
		LatencyMs:     100,
		ClientIP:      "192.168.1.1",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.Enqueue(log)
	}
	
	queue.Close()
}

func BenchmarkDiskQueue_DequeueBatch(b *testing.B) {
	tempDir := b.TempDir()
	queue := NewDiskQueue(tempDir)
	
	// Pre-populate queue
	for i := 0; i < 1000; i++ {
		log := &model.HeimdallRequestLog{
			RequestId:     "test-123",
			NormalizedURL: "/test",
			HTTPMethod:    "GET",
			HTTPStatus:    200,
			LatencyMs:     int64(i),
			ClientIP:      "192.168.1.1",
		}
		queue.Enqueue(log)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.DequeueBatch(10)
	}
	
	queue.Close()
}
