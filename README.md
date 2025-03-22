# GoShell

GoShell is a lightweight, feature-rich shell implemented in Go. It provides essential shell functionality with built-in commands, environment variable support, command history with arrow key navigation, pipe support, and visually enhanced file listings.

## Features

- **Core Shell Functionality**
  - Command execution with argument support
  - Pipe operator (`|`) for connecting commands
  - Command history with persistent storage
  - Arrow key navigation (up/down to browse history, left/right to edit)

- **Environment Variables**
  - View environment variables with `env` or `export`
  - Set environment variables using `export KEY=VALUE`
  - Remove environment variables using `unset KEY`
  - Environment inheritance for child processes

- **Built-in Commands**
  - `cd [dir]` - Change directory (defaults to HOME)
  - `clear` - Clear the terminal screen
  - `echo [args...]` - Print arguments to standard output
  - `env` - Display all environment variables
  - `exit` - Exit the shell
  - `export [KEY=VALUE]` - Set or display environment variables
  - `help` - Show available commands and descriptions
  - `history` - Show command history
  - `ls [dir]` - List directory contents with colorized output and file type icons
  - `pwd` - Print working directory
  - `unset KEY` - Remove an environment variable

- **Enhanced File Listings**
  - Colorized output for different file types
  - Emoji icons for visual file type identification:
    - 📁 Directories
    - 🔗 Symbolic links
    - 💽 Device files
    - 📊 Named pipes
    - 🔌 Sockets
    - ⚙️ Executable files
    - 📄 Text files (txt, md, log, csv)
    - 📕 PDF documents
    - 📘 Word documents (doc, docx, odt)
    - 📗 Spreadsheets (xls, xlsx, ods)
    - 📙 Presentations (ppt, pptx, odp)
    - 🖼️ Images (jpg, png, gif, etc.)
    - 🎵 Audio files (mp3, wav, etc.)
    - 🎬 Video files (mp4, avi, etc.)
    - 🔹 Go files
    - 🐍 Python files
    - 🟨 JavaScript/TypeScript files
    - 🌐 HTML files
    - 🎨 CSS files
    - 🔶 C/C++ files
    - ☕ Java files
    - 💎 Ruby files
    - 💲 Shell scripts
    - 📦 Archive files (zip, tar, etc.)
    - 🔧 Configuration files (json, yaml, etc.)

## Installation

### Prerequisites

- Go 1.16 or higher
- Git

### Building from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/goshell.git
   cd goshell
   ```

2. Install dependencies:
   ```bash
   go get github.com/chzyer/readline
   ```

3. Build the executable:
   ```bash
   go build -o goshell
   ```

4. Run the shell:
   ```bash
   ./goshell
   ```

## Usage Examples

Start the shell:
```bash
./goshell
```

Basic command execution:
```bash
goshell> ls -la
```

Using the enhanced directory listing:
```bash
goshell> ls
```

Using pipes:
```bash
goshell> ls -la | grep .go
```

Setting environment variables:
```bash
goshell> export MY_VAR=hello
goshell> echo $MY_VAR
hello
```

Navigating command history:
- Press the up arrow key to see previous commands
- Press the down arrow key to see more recent commands

## Development

### Running Tests

Run the test suite:
```bash
go test -v
```

### Project Structure

- `main.go` - Main shell implementation
- `main_test.go` - Test suite

## License

[MIT License](LICENSE)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request


## Future Enhancements

- Autocomplete commands
- Support for scripting
- Tab completion for file and directory names
- More built-in commands and utilities
- Git?
