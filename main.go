package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chzyer/readline"
)

// ANSI color codes
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Red       = "\033[31m"
	Green     = "\033[32m"
	Yellow    = "\033[33m"
	Blue      = "\033[34m"
	Magenta   = "\033[35m"
	Cyan      = "\033[36m"
	White     = "\033[37m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"
)

// ShellEnv stores the shell's environment variables
type ShellEnv struct {
	env map[string]string
}

// NewShellEnv creates a new shell environment with system environment variables
func NewShellEnv() *ShellEnv {
	env := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			env[pair[0]] = pair[1]
		}
	}

	// Set default LS_COLORS for colorized output
	if _, exists := env["LS_COLORS"]; !exists {
		env["LS_COLORS"] = "di=1;34:ln=1;36:so=1;35:pi=1;33:ex=1;32:bd=1;33:cd=1;33:su=1;31:sg=1;31:tw=1;34:ow=1;34"
	}

	return &ShellEnv{env: env}
}

// Set sets an environment variable
func (se *ShellEnv) Set(key, value string) {
	se.env[key] = value
}

// Get retrieves an environment variable
func (se *ShellEnv) Get(key string) string {
	return se.env[key]
}

// Unset removes an environment variable
func (se *ShellEnv) Unset(key string) {
	delete(se.env, key)
}

// ToSlice converts the environment map to a slice of "KEY=VALUE" strings
func (se *ShellEnv) ToSlice() []string {
	var result []string
	for k, v := range se.env {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

// Shell represents the shell state
type Shell struct {
	env     *ShellEnv
	history []string
}

// NewShell creates a new shell instance
func NewShell() *Shell {
	return &Shell{
		env:     NewShellEnv(),
		history: make([]string, 0),
	}
}

// AddToHistory adds a command to the shell's history
func (s *Shell) AddToHistory(cmd string) {
	// Don't add empty commands or duplicates of the last command
	if cmd == "" || (len(s.history) > 0 && s.history[len(s.history)-1] == cmd) {
		return
	}
	s.history = append(s.history, cmd)
}

// GetHistory returns the command history
func (s *Shell) GetHistory() []string {
	return s.history
}

// PrintHelp prints available commands and their descriptions
func (s *Shell) PrintHelp() string {
	helpText := `Available commands:
  cd [dir]          Change directory (default: HOME)
  clear             Clear the screen
  echo [args...]    Print arguments
  env               Display environment variables
  exit              Exit the shell
  export [KEY=VALUE] Set environment variables
  help              Show this help message
  history           Show command history
  ls [dir]          List directory contents with colorized output
  pwd               Print working directory
  unset KEY         Remove environment variable`
	fmt.Println(helpText)
	return helpText
}

// ColorizedLS implements a colorized directory listing
func (s *Shell) ColorizedLS(dir string) error {
	// If no directory is provided, use the current directory
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	// Read directory contents
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Sort entries (directories first, then files)
	sort.Slice(entries, func(i, j int) bool {
		iIsDir := entries[i].IsDir()
		jIsDir := entries[j].IsDir()
		if iIsDir && !jIsDir {
			return true
		}
		if !iIsDir && jIsDir {
			return false
		}
		return entries[i].Name() < entries[j].Name()
	})

	// Create a slice to store formatted entry names
	var formattedEntries []string
	maxWidth := 0

	// Format entries with appropriate colors and emoji icons
	for _, entry := range entries {
		var color string
		var icon string
		name := entry.Name()

		// Get file info
		info, err := entry.Info()
		if err != nil {
			// If we can't get info, just add without color or icon
			formattedEntries = append(formattedEntries, name)
			continue
		}

		// Determine icon and color based on file type, extension, and permissions
		switch {
		case entry.IsDir():
			name = name + "/" // Add trailing slash for directories
			color = Bold + Blue
			icon = "📁 " // Folder icon
		case info.Mode()&fs.ModeSymlink != 0:
			color = Bold + Cyan
			icon = "🔗 " // Link icon
		case info.Mode()&fs.ModeDevice != 0:
			color = Bold + Yellow
			icon = "💽 " // Device icon
		case info.Mode()&fs.ModeNamedPipe != 0:
			color = Bold + Yellow
			icon = "📊 " // Pipe icon
		case info.Mode()&fs.ModeSocket != 0:
			color = Bold + Magenta
			icon = "🔌 " // Socket icon
		case info.Mode()&0111 != 0:
			color = Bold + Green
			icon = "⚙️  " // Executable icon
		default:
			// Choose icon based on file extension
			ext := strings.ToLower(filepath.Ext(name))
			switch ext {
			case ".txt", ".md", ".log", ".csv":
				icon = "📄 " // Text file
				color = White
			case ".pdf":
				icon = "📕 " // Document
				color = Red
			case ".doc", ".docx", ".odt":
				icon = "📘 " // Word document
				color = Blue
			case ".xls", ".xlsx", ".ods":
				icon = "📗 " // Spreadsheet
				color = Green
			case ".ppt", ".pptx", ".odp":
				icon = "📙 " // Presentation
				color = Yellow
			case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg":
				icon = "🖼️  " // Image
				color = Magenta
			case ".mp3", ".wav", ".flac", ".ogg", ".m4a":
				icon = "🎵 " // Audio
				color = Cyan
			case ".mp4", ".avi", ".mkv", ".mov", ".wmv":
				icon = "🎬 " // Video
				color = Yellow
			case ".zip", ".tar", ".gz", ".rar", ".7z":
				icon = "📦 " // Archive
				color = Red
			case ".go":
				icon = "🔹 " // Go files
				color = Cyan
			case ".py":
				icon = "🐍 " // Python files
				color = Yellow
			case ".js", ".ts":
				icon = "🟨 " // JavaScript/TypeScript
				color = Yellow
			case ".html", ".htm":
				icon = "🌐 " // HTML
				color = Bold + Red
			case ".css":
				icon = "🎨 " // CSS
				color = Bold + Magenta
			case ".c", ".cpp", ".h", ".hpp":
				icon = "🔶 " // C/C++
				color = Blue
			case ".java":
				icon = "☕ " // Java
				color = Red
			case ".sh", ".bash", ".zsh":
				icon = "💲 " // Shell scripts
				color = Green
			case ".rb":
				icon = "💎 " // Ruby
				color = Red
			case ".json", ".yaml", ".yml", ".toml", ".xml":
				icon = "🔧 " // Config files
				color = Yellow
			default:
				icon = "📄 " // Default file icon
				color = Reset
			}
		}

		// Add colored name with icon to our entries list
		formattedName := fmt.Sprintf("%s%s%s%s%s", color, icon, name, Reset, "")
		formattedEntries = append(formattedEntries, formattedName)

		// Track the maximum width for columnar output
		// Account for emoji (typically 2 chars wide) + space + name length
		displayWidth := len(name) + 3 // +3 for emoji and space
		if displayWidth > maxWidth {
			maxWidth = displayWidth
		}
	}

	// Print entries in a grid-like format
	termWidth := 80 // Default terminal width
	if ws, err := getTerminalSize(); err == nil {
		termWidth = ws.Col
	}

	// Calculate columns based on terminal width and max filename width
	// Add 2 for some padding between columns
	colWidth := maxWidth + 2
	numCols := termWidth / colWidth
	if numCols < 1 {
		numCols = 1
	}

	// Print entries in rows and columns
	for i, entry := range formattedEntries {
		// Print the entry with padding
		fmt.Print(entry)

		// Add appropriate spacing for columnar output
		if (i+1)%numCols != 0 && i < len(formattedEntries)-1 {
			// Print spaces to fill the column
			// We need to account for the invisible ANSI color codes and emoji width
			paddingWidth := colWidth - len(stripANSI(entry))
			// Emojis typically take 2 character positions in terminal
			// We need to adjust for this to maintain proper alignment
			if strings.Contains(entry, "📁") || strings.Contains(entry, "🔗") ||
				strings.Contains(entry, "📄") || strings.Contains(entry, "🖼️") {
				paddingWidth += 1
			}
			fmt.Print(strings.Repeat(" ", paddingWidth))
		} else {
			// End of row or last entry
			fmt.Println()
		}
	}

	// Ensure a newline at the end if needed
	if len(formattedEntries)%numCols != 0 {
		fmt.Println()
	}

	return nil
}

// stripANSI removes ANSI escape codes from a string
func stripANSI(str string) string {
	// Regular expression to match ANSI escape codes: \x1b\[[0-9;]*[a-zA-Z]
	// For simplicity and performance, we'll just remove the specific color codes we use
	result := str
	for _, code := range []string{
		Reset, Bold, Red, Green, Yellow, Blue, Magenta, Cyan, White,
		BgRed, BgGreen, BgYellow, BgBlue, BgMagenta, BgCyan, BgWhite,
		Bold + Red, Bold + Green, Bold + Yellow, Bold + Blue, Bold + Magenta, Bold + Cyan, Bold + White,
	} {
		result = strings.ReplaceAll(result, code, "")
	}
	return result
}

// TermSize represents terminal dimensions
type TermSize struct {
	Row, Col int
}

// getTerminalSize attempts to get the dimensions of the terminal
func getTerminalSize() (TermSize, error) {
	// Default size in case we can't detect
	defaultSize := TermSize{Row: 24, Col: 80}

	// Try to get terminal size using stty
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return defaultSize, err
	}

	// Parse the output
	parts := strings.Split(strings.TrimSpace(string(out)), " ")
	if len(parts) != 2 {
		return defaultSize, fmt.Errorf("unexpected format from stty")
	}

	// Convert to integers
	row, err := parseInt(parts[0])
	if err != nil {
		return defaultSize, err
	}

	col, err := parseInt(parts[1])
	if err != nil {
		return defaultSize, err
	}

	return TermSize{Row: row, Col: col}, nil
}

// parseInt parses a string to int with error handling
func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

func main() {
	shell := NewShell()

	// Configure readline
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "goshell> ",
		HistoryFile:     "/tmp/goshell_history",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing readline: %v\n", err)
		os.Exit(1)
	}
	defer rl.Close()

	for {
		// Read input using readline (supports arrow keys for history)
		input, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				continue
			} else if err == io.EOF {
				fmt.Println("Goodbye!")
				return
			}
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			continue
		}

		// Trim whitespace
		input = strings.TrimSpace(input)

		// Skip empty commands
		if input == "" {
			continue
		}

		// Add command to history
		shell.AddToHistory(input)
		rl.SaveHistory(input)

		// Handle built-in commands before piping logic
		args := strings.Fields(input)
		if len(args) == 0 {
			continue
		}

		switch args[0] {
		case "cd":
			var path string
			if len(args) < 2 {
				path = os.Getenv("HOME")
			} else {
				path = args[1]
			}
			if err := os.Chdir(path); err != nil {
				fmt.Fprintln(os.Stderr, "Error changing directory:", err)
			}
			continue

		case "clear":
			cmd := exec.Command("clear")
			cmd.Stdout = os.Stdout
			cmd.Run()
			continue

		case "echo":
			// Join all arguments with spaces and print
			fmt.Println(strings.Join(args[1:], " "))
			continue

		case "env":
			// Print all environment variables
			for _, env := range shell.env.ToSlice() {
				fmt.Println(env)
			}
			continue

		case "export":
			if len(args) < 2 {
				// Print all environment variables
				for _, env := range shell.env.ToSlice() {
					fmt.Println(env)
				}
				continue
			}
			// Handle export KEY=VALUE
			for _, arg := range args[1:] {
				parts := strings.SplitN(arg, "=", 2)
				if len(parts) == 2 {
					shell.env.Set(parts[0], parts[1])
				} else {
					fmt.Fprintf(os.Stderr, "Invalid export syntax: %s\n", arg)
				}
			}
			continue

		case "exit":
			fmt.Println("Goodbye!")
			return

		case "help":
			shell.PrintHelp()
			continue

		case "history":
			for i, cmd := range shell.GetHistory() {
				fmt.Printf("%d  %s\n", i+1, cmd)
			}
			continue

		case "ls":
			// Check if we should use the built-in colorized ls or system ls
			if len(args) > 1 && (args[1] == "--help" || args[1] == "-l") {
				// For complex ls commands, fall back to system ls with color
				systemArgs := append([]string{"--color=auto"}, args[1:]...)
				cmd := exec.Command("ls", systemArgs...)
				cmd.Env = shell.env.ToSlice()
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Run()
			} else {
				// Use our built-in colorized ls for simple directory listings
				var dir string
				if len(args) > 1 {
					dir = args[1]
				} else {
					dir = "."
				}
				if err := shell.ColorizedLS(dir); err != nil {
					fmt.Fprintln(os.Stderr, "Error listing directory:", err)
				}
			}
			continue

		case "pwd":
			dir, err := os.Getwd()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error getting working directory:", err)
			} else {
				fmt.Println(dir)
			}
			continue

		case "unset":
			if len(args) < 2 {
				fmt.Fprintln(os.Stderr, "Usage: unset KEY")
				continue
			}
			shell.env.Unset(args[1])
			continue
		}

		// If the command includes a pipe, handle piping logic
		if strings.Contains(input, "|") {
			// Split commands by "|"
			commands := strings.Split(input, "|")
			var cmds []*exec.Cmd
			var pipes []*os.File

			// Create a command for each segment
			for _, cmdStr := range commands {
				parts := strings.Fields(strings.TrimSpace(cmdStr))
				if len(parts) == 0 {
					continue
				}

				// Handle 'ls' specially to ensure colors are enabled
				if parts[0] == "ls" {
					parts = append([]string{"ls", "--color=auto"}, parts[1:]...)
				}

				cmd := exec.Command(parts[0], parts[1:]...)
				cmd.Env = shell.env.ToSlice()
				cmds = append(cmds, cmd)
			}

			// Link them with pipes
			for i := 0; i < len(cmds)-1; i++ {
				r, w, err := os.Pipe()
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error creating pipe:", err)
					break
				}
				cmds[i].Stdout = w
				cmds[i].Stderr = os.Stderr
				cmds[i+1].Stdin = r
				pipes = append(pipes, r, w)
			}

			// If first cmd doesn't have an input yet, use stdin
			if cmds[0].Stdin == nil {
				cmds[0].Stdin = os.Stdin
			}
			// If last cmd doesn't have an output yet, use stdout
			if cmds[len(cmds)-1].Stdout == nil {
				cmds[len(cmds)-1].Stdout = os.Stdout
			}
			cmds[len(cmds)-1].Stderr = os.Stderr

			// Start each command
			for _, c := range cmds {
				if err := c.Start(); err != nil {
					fmt.Fprintln(os.Stderr, "Error starting command:", err)
				}
			}

			// Close all pipe ends in the parent
			for _, p := range pipes {
				p.Close()
			}

			// Wait for each command to finish
			for _, c := range cmds {
				if err := c.Wait(); err != nil {
					fmt.Fprintln(os.Stderr, "Error waiting for command:", err)
				}
			}
			continue
		}

		// Special handling for ls to ensure colors are enabled
		if args[0] == "ls" {
			// Create a new args slice with --color=auto inserted
			colorArgs := []string{"--color=auto"}
			colorArgs = append(colorArgs, args[1:]...)

			// Execute ls with color
			cmd := exec.Command("ls", colorArgs...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Env = shell.env.ToSlice()

			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
			}
			continue
		}

		// For non-piped external commands, execute normally
		command := args[0]
		args = args[1:]
		cmd := exec.Command(command, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = shell.env.ToSlice()

		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		}
	}
}
