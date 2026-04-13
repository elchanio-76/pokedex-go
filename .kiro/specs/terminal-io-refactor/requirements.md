# Requirements Document

## Introduction

The Pokedex CLI application currently has a `LineReader` interface that abstracts terminal input, with two implementations: a simple `bufio.Scanner`-based reader and a raw-mode `termReader` using `golang.org/x/term`. However, output is not abstracted — command callbacks and the REPL loop use `fmt.Println`/`fmt.Printf` directly, which breaks in raw terminal mode because raw mode requires `\r\n` line endings instead of bare `\n`. This refactoring unifies input and output behind a single `Terminal` interface so that all I/O works correctly regardless of the underlying terminal mode.

## Glossary

- **Terminal**: The unified interface that provides both input (ReadLine) and output (Print, Println) methods for the REPL and command callbacks.
- **Line_Reader**: The `bufio.Scanner`-based implementation of Terminal for cooked (normal) terminal mode.
- **Term_Reader**: The `golang.org/x/term`-based implementation of Terminal for raw terminal mode.
- **REPL**: The Read-Eval-Print Loop that reads user commands, dispatches them, and displays results.
- **Command_Callback**: A function registered in the command registry that executes a CLI command and produces output.
- **Raw_Mode**: Terminal mode where input is not line-buffered and output requires explicit `\r\n` for newlines.
- **Cooked_Mode**: Normal terminal mode where the OS handles line editing and `\n` is sufficient for newlines.

## Requirements

### Requirement 1: Unified Terminal Interface

**User Story:** As a developer, I want a single interface that covers both input and output, so that all terminal I/O goes through one abstraction regardless of terminal mode.

#### Acceptance Criteria

1. THE Terminal interface SHALL provide a ReadLine method that accepts a prompt string and returns the user input and an error.
2. THE Terminal interface SHALL provide a Print method that accepts a format string and variadic arguments.
3. THE Terminal interface SHALL provide a Println method that accepts variadic arguments.
4. THE Terminal interface SHALL provide a Close method that returns an error.
5. THE Terminal interface SHALL provide a SetHistory method that accepts a string slice (reserved for future use).

### Requirement 2: Line_Reader Implementation (Cooked Mode)

**User Story:** As a developer, I want the bufio-based implementation to satisfy the Terminal interface, so that the application works in normal terminal mode.

#### Acceptance Criteria

1. THE Line_Reader SHALL implement the Terminal interface.
2. WHEN ReadLine is called, THE Line_Reader SHALL print the prompt to standard output and read one line from standard input.
3. WHEN Print is called, THE Line_Reader SHALL write the formatted string to standard output using fmt.Printf semantics.
4. WHEN Println is called, THE Line_Reader SHALL write the arguments to standard output using fmt.Println semantics.
5. WHEN Close is called, THE Line_Reader SHALL return nil without performing cleanup.
6. WHEN SetHistory is called, THE Line_Reader SHALL accept the argument without performing any action.

### Requirement 3: Term_Reader Implementation (Raw Mode)

**User Story:** As a developer, I want the raw-mode implementation to satisfy the Terminal interface with correct newline handling, so that output displays properly in raw terminal mode.

#### Acceptance Criteria

1. THE Term_Reader SHALL implement the Terminal interface.
2. WHEN Print is called with a format string containing newline characters, THE Term_Reader SHALL replace each bare `\n` with `\r\n` before writing to the terminal.
3. WHEN Println is called, THE Term_Reader SHALL replace each bare `\n` in the formatted output with `\r\n` and append `\r\n` as the line terminator.
4. WHEN ReadLine is called, THE Term_Reader SHALL display the prompt, read user input character by character, echo typed characters, handle backspace, and return the completed line after the user presses Enter.
5. WHEN ReadLine detects Enter, THE Term_Reader SHALL write `\r\n` and return the accumulated input without re-printing the prompt.
6. WHEN Close is called, THE Term_Reader SHALL restore the terminal to its previous state using the saved terminal state.
7. WHEN SetHistory is called, THE Term_Reader SHALL accept the argument without performing any action.

### Requirement 4: REPL Uses Terminal Interface for Output

**User Story:** As a developer, I want the REPL loop to use the Terminal interface for all output, so that output displays correctly in both terminal modes.

#### Acceptance Criteria

1. WHEN an unknown command is entered, THE REPL SHALL use Terminal.Print to display the error message instead of calling fmt.Printf directly.
2. WHEN a command callback returns an error, THE REPL SHALL use Terminal.Println to display the error instead of calling fmt.Println directly.

### Requirement 5: Command Callbacks Use Terminal Interface for Output

**User Story:** As a developer, I want all command callbacks to receive and use the Terminal interface for output, so that command output renders correctly in both terminal modes.

#### Acceptance Criteria

1. THE Command_Callback function signature SHALL accept the Terminal interface as a parameter.
2. WHEN commandHelp is called, THE Command_Callback SHALL use Terminal.Print to display the help text.
3. WHEN commandMap is called, THE Command_Callback SHALL use Terminal.Print and Terminal.Println to display location areas.
4. WHEN commandMapBack is called, THE Command_Callback SHALL use Terminal.Print and Terminal.Println to display location areas.
5. WHEN commandExplore is called, THE Command_Callback SHALL use Terminal.Print to display pokemon encounters.
6. WHEN commandCatch is called, THE Command_Callback SHALL use Terminal.Print to display catch results.
7. WHEN commandInspect is called, THE Command_Callback SHALL use Terminal.Print to display pokemon details.
8. WHEN commandPokedex is called, THE Command_Callback SHALL use Terminal.Println and Terminal.Print to display the pokedex contents.
9. WHEN commandExit is called, THE Command_Callback SHALL use Terminal.Println to display the goodbye message before exiting.

### Requirement 6: No Direct fmt Output Calls in REPL or Commands

**User Story:** As a developer, I want to ensure no direct `fmt.Println` or `fmt.Printf` calls remain in the REPL or command callbacks, so that all output is guaranteed to go through the Terminal abstraction.

#### Acceptance Criteria

1. THE REPL module (repl.go) SHALL contain zero direct calls to fmt.Println or fmt.Printf for user-visible output.
2. THE command callbacks in main.go SHALL contain zero direct calls to fmt.Println or fmt.Printf for user-visible output.
