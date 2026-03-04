package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Vault manages the append-only JSONL log files for sync/backup.
type Vault struct {
	dir      string
	deviceID string
	mu       sync.Mutex
}

// New creates or opens a Vault at the given directory.
func New(dir string) (*Vault, error) {
	if err := os.MkdirAll(filepath.Join(dir, "log"), 0755); err != nil {
		return nil, fmt.Errorf("vault: failed to create log dir: %w", err)
	}
	deviceID, err := getOrCreateDeviceID(dir)
	if err != nil {
		return nil, err
	}
	return &Vault{dir: dir, deviceID: deviceID}, nil
}

// Append writes an operation to the current device's monthly log file.
// Errors are non-fatal: the caller should log but not fail the request.
func (v *Vault) Append(op OpType, entityType EntityType, entityID string, data any) error {
	entry := Entry{
		ID:       uuid.New().String(),
		Op:       op,
		Type:     entityType,
		EntityID: entityID,
		Data:     data,
		TS:       time.Now().UTC().Format(time.RFC3339Nano),
	}

	line, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("vault: failed to marshal entry: %w", err)
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	f, err := os.OpenFile(v.currentLogFile(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("vault: failed to open log file: %w", err)
	}
	defer f.Close()

	_, err = f.Write(append(line, '\n'))
	return err
}

// Dir returns the vault root directory.
func (v *Vault) Dir() string {
	return v.dir
}

// DeviceID returns the current device's identifier.
func (v *Vault) DeviceID() string {
	return v.deviceID
}

// currentLogFile returns the path to this device's current monthly log file.
func (v *Vault) currentLogFile() string {
	month := time.Now().UTC().Format("2006-01")
	return filepath.Join(v.dir, "log", fmt.Sprintf("%s-%s.jsonl", v.deviceID, month))
}

func getOrCreateDeviceID(dir string) (string, error) {
	idFile := filepath.Join(dir, "device_id")
	data, err := os.ReadFile(idFile)
	if err == nil {
		id := strings.TrimSpace(string(data))
		if id != "" {
			return id, nil
		}
	}

	hostname, _ := os.Hostname()
	id := fmt.Sprintf("%s-%s", hostname, uuid.New().String()[:8])
	if err := os.WriteFile(idFile, []byte(id), 0644); err != nil {
		return "", fmt.Errorf("vault: failed to write device_id: %w", err)
	}
	return id, nil
}
