package state

import (
	"encoding/json"
	"os"
	"spuderman/pkg/utils"
	"sync"
)

type State struct {
	CompletedHosts map[string]bool `json:"completed_hosts"`
}

type Manager struct {
	path  string
	state State
	mu    sync.Mutex
}

func NewManager(path string) (*Manager, error) {
	m := &Manager{
		path: path,
		state: State{
			CompletedHosts: make(map[string]bool),
		},
	}

	// Try to load existing
	if _, err := os.Stat(path); err == nil {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(content, &m.state); err != nil {
			// If corrupt, warn and start fresh? Or error?
			utils.LogError("Failed to parse state file %s: %v. Starting fresh.", path, err)
			m.state.CompletedHosts = make(map[string]bool)
		}
	}

	return m, nil
}

func (m *Manager) IsCompleted(host string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state.CompletedHosts[host]
}

func (m *Manager) MarkCompleted(host string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.state.CompletedHosts[host] = true
	m.save()
}

func (m *Manager) save() {
	// Must be called with lock held
	content, err := json.MarshalIndent(m.state, "", "  ")
	if err == nil {
		os.WriteFile(m.path, content, 0644)
	} else {
		utils.LogError("Failed to save state: %v", err)
	}
}
