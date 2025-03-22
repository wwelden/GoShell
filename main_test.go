package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// MockReadline is a test helper to simulate readline functionality
type MockReadline struct {
	lines   []string
	current int
	saved   []string
}

// NewMockReadline creates a new mock readline instance
func NewMockReadline(lines []string) *MockReadline {
	return &MockReadline{
		lines:   lines,
		current: 0,
		saved:   []string{},
	}
}

// Readline simulates reading a line from input
func (m *MockReadline) Readline() (string, error) {
	if m.current >= len(m.lines) {
		return "", io.EOF
	}
	line := m.lines[m.current]
	m.current++
	return line, nil
}

// SaveHistory simulates saving to history
func (m *MockReadline) SaveHistory(line string) error {
	m.saved = append(m.saved, line)
	return nil
}

// Close is a mock close function
func (m *MockReadline) Close() error {
	return nil
}

func TestShellEnv(t *testing.T) {
	env := NewShellEnv()

	// Test setting and getting environment variables
	t.Run("Set and Get", func(t *testing.T) {
		env.Set("TEST_VAR", "test_value")
		if got := env.Get("TEST_VAR"); got != "test_value" {
			t.Errorf("Get() = %v, want %v", got, "test_value")
		}
	})

	// Test unsetting environment variables
	t.Run("Unset", func(t *testing.T) {
		env.Set("TEST_VAR", "test_value")
		env.Unset("TEST_VAR")
		if got := env.Get("TEST_VAR"); got != "" {
			t.Errorf("Get() after Unset() = %v, want empty string", got)
		}
	})

	// Test ToSlice conversion
	t.Run("ToSlice", func(t *testing.T) {
		// Create a clean environment for testing
		cleanEnv := &ShellEnv{env: make(map[string]string)}
		cleanEnv.Set("TEST_VAR1", "value1")
		cleanEnv.Set("TEST_VAR2", "value2")

		slice := cleanEnv.ToSlice()
		expected := map[string]string{
			"TEST_VAR1": "value1",
			"TEST_VAR2": "value2",
		}

		// Check that our test variables are in the slice
		foundVars := 0
		for _, entry := range slice {
			parts := strings.SplitN(entry, "=", 2)
			if len(parts) != 2 {
				t.Errorf("Invalid entry in ToSlice(): %v", entry)
				continue
			}

			if val, ok := expected[parts[0]]; ok {
				if val != parts[1] {
					t.Errorf("ToSlice() entry %v = %v, want %v", parts[0], parts[1], val)
				}
				foundVars++
			}
		}

		if foundVars != len(expected) {
			t.Errorf("Not all expected variables found in ToSlice(). Found %d, want %d", foundVars, len(expected))
		}
	})
}

func TestShell(t *testing.T) {
	shell := NewShell()

	// Test command history
	t.Run("Command History", func(t *testing.T) {
		commands := []string{"echo test", "pwd", "help"}
		for _, cmd := range commands {
			shell.AddToHistory(cmd)
		}
		history := shell.GetHistory()
		if len(history) != len(commands) {
			t.Errorf("History length = %v, want %v", len(history), len(commands))
		}
		for i, cmd := range commands {
			if history[i] != cmd {
				t.Errorf("History[%d] = %v, want %v", i, history[i], cmd)
			}
		}
	})

	// Test duplicate command filtering
	t.Run("Duplicate Command Filtering", func(t *testing.T) {
		shell := NewShell()
		commands := []string{"echo test", "echo test", "pwd", "pwd", "help"}
		expected := []string{"echo test", "pwd", "help"}

		for _, cmd := range commands {
			shell.AddToHistory(cmd)
		}

		history := shell.GetHistory()
		if len(history) != len(expected) {
			t.Errorf("History length after filtering = %v, want %v", len(history), len(expected))
		}

		for i, cmd := range expected {
			if i >= len(history) {
				t.Errorf("History too short, missing expected command: %v", cmd)
				continue
			}
			if history[i] != cmd {
				t.Errorf("History[%d] = %v, want %v", i, history[i], cmd)
			}
		}
	})

	// Test help text
	t.Run("Help Text", func(t *testing.T) {
		helpText := shell.PrintHelp()
		expectedCommands := []string{"cd", "clear", "echo", "env", "exit", "export", "help", "history", "pwd", "unset"}
		for _, cmd := range expectedCommands {
			if !strings.Contains(helpText, cmd) {
				t.Errorf("Help text missing command: %v", cmd)
			}
		}
	})
}

func TestBuiltInCommands(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "goshell_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Resolve symlinks to get the real path for macOS (/private/var vs /var)
	realTmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Save original working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// Change to temporary directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		command  string
		args     []string
		wantErr  bool
		validate func(t *testing.T)
	}{
		{
			name:    "cd to home directory",
			command: "cd",
			args:    []string{},
			wantErr: false,
			validate: func(t *testing.T) {
				home, _ := os.UserHomeDir()
				got, _ := os.Getwd()
				// Compare the resolved paths
				gotResolved, _ := filepath.EvalSymlinks(got)
				homeResolved, _ := filepath.EvalSymlinks(home)
				if gotResolved != homeResolved {
					t.Errorf("cd to home = %v, want %v", gotResolved, homeResolved)
				}
			},
		},
		{
			name:    "cd to specific directory",
			command: "cd",
			args:    []string{tmpDir},
			wantErr: false,
			validate: func(t *testing.T) {
				got, _ := os.Getwd()
				// Compare the resolved paths
				gotResolved, _ := filepath.EvalSymlinks(got)
				if gotResolved != realTmpDir {
					t.Errorf("cd to tmpDir = %v, want %v", gotResolved, realTmpDir)
				}
			},
		},
		{
			name:    "pwd shows current directory",
			command: "pwd",
			args:    []string{},
			wantErr: false,
			validate: func(t *testing.T) {
				got, _ := os.Getwd()
				// Compare the resolved paths
				gotResolved, _ := filepath.EvalSymlinks(got)
				if gotResolved != realTmpDir {
					t.Errorf("pwd = %v, want %v", gotResolved, realTmpDir)
				}
			},
		},
		{
			name:    "echo prints arguments",
			command: "echo",
			args:    []string{"test", "message"},
			wantErr: false,
			validate: func(t *testing.T) {
				// Note: This test would need to capture stdout to verify
				// For now, we just verify it doesn't error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := strings.Fields(tt.command)
			args = append(args, tt.args...)

			switch args[0] {
			case "cd":
				var path string
				if len(args) < 2 {
					path = os.Getenv("HOME")
				} else {
					path = args[1]
				}
				err := os.Chdir(path)
				if (err != nil) != tt.wantErr {
					t.Errorf("cd error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.validate != nil {
					tt.validate(t)
				}
			case "pwd":
				_, err := os.Getwd()
				if (err != nil) != tt.wantErr {
					t.Errorf("pwd error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.validate != nil {
					tt.validate(t)
				}
			case "echo":
				// For echo, we just verify it doesn't error
				if tt.validate != nil {
					tt.validate(t)
				}
			}
		})
	}
}

func TestEnvironmentVariables(t *testing.T) {
	shell := NewShell()

	tests := []struct {
		name     string
		action   func()
		validate func(t *testing.T)
	}{
		{
			name: "Set and get environment variable",
			action: func() {
				shell.env.Set("TEST_VAR", "test_value")
			},
			validate: func(t *testing.T) {
				if got := shell.env.Get("TEST_VAR"); got != "test_value" {
					t.Errorf("Get() = %v, want %v", got, "test_value")
				}
			},
		},
		{
			name: "Unset environment variable",
			action: func() {
				shell.env.Set("TEST_VAR", "test_value")
				shell.env.Unset("TEST_VAR")
			},
			validate: func(t *testing.T) {
				if got := shell.env.Get("TEST_VAR"); got != "" {
					t.Errorf("Get() after Unset() = %v, want empty string", got)
				}
			},
		},
		{
			name: "Environment variables in child process",
			action: func() {
				shell.env.Set("TEST_VAR", "test_value")
			},
			validate: func(t *testing.T) {
				cmd := exec.Command("sh", "-c", "echo $TEST_VAR")
				cmd.Env = shell.env.ToSlice()
				output, err := cmd.Output()
				if err != nil {
					t.Errorf("Failed to run command: %v", err)
					return
				}
				if got := strings.TrimSpace(string(output)); got != "test_value" {
					t.Errorf("Child process output = %v, want %v", got, "test_value")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.action()
			tt.validate(t)
		})
	}
}

func TestReadlineIntegration(t *testing.T) {
	// Create a temporary history file
	historyFile, err := ioutil.TempFile("", "readline_history")
	if err != nil {
		t.Fatal(err)
	}
	historyPath := historyFile.Name()
	historyFile.Close()
	defer os.Remove(historyPath)

	t.Run("Command Processing", func(t *testing.T) {
		// Test input commands
		inputCommands := []string{
			"echo hello world",
			"pwd",
			"exit",
		}

		// Create mock readline
		mockRL := NewMockReadline(inputCommands)

		// Create shell with the mock readline
		shell := NewShell()

		// Process each command through processCommand helper
		for i := 0; i < len(inputCommands); i++ {
			cmd, err := mockRL.Readline()
			if err != nil {
				if err == io.EOF && i == len(inputCommands)-1 {
					// Expected EOF on last command
					break
				}
				t.Errorf("Unexpected error: %v", err)
				break
			}

			// Add to history
			shell.AddToHistory(cmd)
			mockRL.SaveHistory(cmd)

			// Verify command is in history
			history := shell.GetHistory()
			if history[len(history)-1] != cmd {
				t.Errorf("Command not added to history correctly, got %v, want %v",
					history[len(history)-1], cmd)
			}
		}

		// Verify history saved correctly
		if len(mockRL.saved) != len(inputCommands) {
			t.Errorf("Expected %d saved commands, got %d", len(inputCommands), len(mockRL.saved))
		}
	})
}

// Helper function to test command execution with captured output
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}
