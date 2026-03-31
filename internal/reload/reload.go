package reload

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Reloader handles debounced nginx reload.
type Reloader struct {
	command  string
	debounce time.Duration
	logger   *slog.Logger
	mu       sync.Mutex
	timer    *time.Timer
}

// New creates a Reloader. If debounce is 0, Trigger() calls reload immediately.
func New(command string, debounce time.Duration, logger *slog.Logger) *Reloader {
	return &Reloader{
		command:  command,
		debounce: debounce,
		logger:   logger,
	}
}

// Trigger schedules a debounced reload.
func (r *Reloader) Trigger() {
	if r.debounce == 0 {
		if err := r.exec(); err != nil {
			r.logger.Error("reload failed", "error", err)
		}
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.timer != nil {
		r.timer.Stop()
	}
	r.timer = time.AfterFunc(r.debounce, func() {
		if err := r.exec(); err != nil {
			r.logger.Error("reload failed", "error", err)
		}
	})
}

// ReloadNow executes the reload command immediately.
func (r *Reloader) ReloadNow() error {
	r.mu.Lock()
	if r.timer != nil {
		r.timer.Stop()
		r.timer = nil
	}
	r.mu.Unlock()

	return r.exec()
}

// Stop cancels any pending debounced reload.
func (r *Reloader) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.timer != nil {
		r.timer.Stop()
		r.timer = nil
	}
}

func (r *Reloader) exec() error {
	parts := strings.Fields(r.command)
	if len(parts) == 0 {
		return fmt.Errorf("empty reload command")
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("reload command failed: %w: %s", err, string(output))
	}
	r.logger.Info("nginx reloaded", "command", r.command)
	return nil
}
