# Implementation Plan: Terminal I/O Refactor

## Overview

Rename `LineReader` → `Terminal`, add `\r\n` translation in `termReader`, thread the `Terminal` interface through the REPL and all command callbacks, and eliminate all direct `fmt.Print*` calls from `repl.go` and `main.go`.

## Tasks

- [ ] 1. Rename LineReader interface to Terminal and update constructors
  - [x] 1.1 Rename `LineReader` interface to `Terminal` in `internal/input/input.go`
    - Rename the interface from `LineReader` to `Terminal`
    - Change `NewLineReader()` return type from `LineReader` to `Terminal`
    - Change `NewTerminalReader()` return type from `*termReader` to `Terminal`
    - Replace `interface{}` with `any` in all method signatures
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 3.1_

  - [x] 1.2 Add `\n` → `\r\n` translation in `termReader.Print` and `termReader.Println`
    - Extract or inline the normalize-then-translate pattern: `strings.ReplaceAll(s, "\r\n", "\n")` then `strings.ReplaceAll(s, "\n", "\r\n")`
    - Apply translation in `termReader.Print` after `fmt.Sprintf`
    - Apply translation in `termReader.Println` after `fmt.Sprintln`
    - Add `"strings"` import to `internal/input/input.go`
    - _Requirements: 3.2, 3.3_

  - [ ]* 1.3 Write property tests for newline translation
    - Create `internal/input/input_test.go`
    - Export or extract the translation logic into a testable function (e.g., `TranslateCRLF`)
    - **Property 1: Newline translation eliminates bare line feeds** — for any arbitrary string, after translation no bare `\n` exists without a preceding `\r`
    - **Validates: Requirements 3.2, 3.3**
    - **Property 2: Newline translation is idempotent** — applying translation twice produces the same result as once
    - **Validates: Requirements 3.2, 3.3**
    - Use `testing/quick` or `pgregory.net/rapid` with at least 100 iterations

- [x] 2. Checkpoint - Verify interface rename and translation
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 3. Thread Terminal into REPL and update command callback signature
  - [x] 3.1 Update `startREPL` to accept `input.Terminal` and use it for output
    - Change `startREPL` signature from `reader input.LineReader` to `reader input.Terminal`
    - Replace `fmt.Printf("Unknown command %s\r\n", words[0])` with `reader.Print("Unknown command %s\n", words[0])`
    - Replace `fmt.Println(err)` with `reader.Println(err)`
    - Remove unused `"fmt"` import from `repl.go` if no longer needed
    - _Requirements: 4.1, 4.2, 6.1_

  - [x] 3.2 Update `cliCommand` callback signature to accept `input.Terminal`
    - Change `callback` field type from `func(cfg *Config, args []string) error` to `func(cfg *Config, args []string, t input.Terminal) error`
    - Update the callback invocation in `startREPL` to pass `reader` as the third argument
    - _Requirements: 5.1_

  - [x] 3.3 Update all command callbacks in `main.go` to accept and use `input.Terminal`
    - Add `t input.Terminal` parameter to every command function: `commandExit`, `commandHelp`, `commandMap`, `commandMapBack`, `commandExplore`, `commandCatch`, `commandInspect`, `commandPokedex`
    - Replace every `fmt.Println(...)` with `t.Println(...)` and every `fmt.Printf(...)` with `t.Print(...)` in each callback
    - Remove `"fmt"` import from `main.go` if no longer needed (keep it if `fmt.Errorf` is still used)
    - _Requirements: 5.2, 5.3, 5.4, 5.5, 5.6, 5.7, 5.8, 5.9, 6.2_

  - [x] 3.4 Update `main()` to use `input.Terminal` type
    - Change `reader` variable type annotation (if any) to `input.Terminal`
    - Remove `defer reader.Close()` from `main()` since `startREPL` already defers Close
    - Ensure `startREPL(&cfg, reader)` compiles with the new signature
    - _Requirements: 1.1, 4.1_

- [x] 4. Checkpoint - Verify full compilation and existing tests
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 5. Final verification
  - [x] 5.1 Verify no direct `fmt.Print*` calls remain in REPL or command output paths
    - Grep `repl.go` and command callbacks in `main.go` for `fmt.Println`, `fmt.Printf`, `fmt.Print` — only `fmt.Errorf` should remain
    - _Requirements: 6.1, 6.2_

  - [x] 5.2 Write unit tests for REPL and command callback output routing
    - Create a mock `Terminal` in `repl_test.go` that records Print/Println calls
    - Test that unknown command triggers `Print` with "Unknown command" message
    - Test that callback error triggers `Println` with the error
    - Test that `commandHelp` calls `Print` with help text
    - _Requirements: 4.1, 4.2, 5.2_

- [ ] 6. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.
