# Implementation Plan: Location Movement System

## Overview

This plan implements a Location navigation layer on top of the existing LocationArea system. The implementation adds a `visit` command, scopes `map`/`mapb`/`explore` to the visited Location, and introduces a dynamic colored prompt reflecting navigation context. Tasks are ordered so each step builds on the previous, with tests placed close to the code they verify.

Several foundational pieces are already in place: the `Location` struct, `GetLocation` API function, updated `Config` struct, `truncateName`, and `buildPrompt`. The remaining work focuses on the command implementations and REPL integration.

## Tasks

- [x] 1. Implement `commandVisit` and register it
  - [x] 1.1 Implement `commandVisit` function in `main.go`
    - Validate args: if `len(args) != 1`, print "You must provide a location name to visit" and return nil
    - Call `pokeapi.GetLocation(args[0], cache)`
    - On error: print "Location '<name>' not found", leave Config unchanged, return nil
    - On success: set `cfg.CurrentLocation` to the result, set `cfg.CurrentArea = ""`, set `cfg.ScopedPageIndex = 0`
    - Display: location name, region name, and area count (len of areas list)
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.8_

  - [x] 1.2 Register `visit` command in `commandRegistry` in `main.go`
    - Key: "visit"
    - Name: "visit"
    - Description: "Visit a location to explore its areas"
    - Callback: `commandVisit`
    - _Requirements: 6.1, 6.2_

  - [x] 1.3 Write property test for visit state transition
    - **Property 1: Visit State Transition**
    - **Validates: Requirements 1.1, 1.5, 2.8**
    - Generate random valid Location structs, simulate successful visit, assert CurrentLocation equals new data, CurrentArea is empty, ScopedPageIndex is 0

  - [x] 1.4 Write property test for visit output format
    - **Property 2: Visit Output Contains Required Information**
    - **Validates: Requirements 1.2**
    - Generate random location name, region, and area list, verify output contains all three pieces of information

  - [x] 1.5 Write property test for failed visit preserving state
    - **Property 3: Failed Visit Preserves State**
    - **Validates: Requirements 1.3**
    - Generate random initial Config state, simulate API failure, assert all Config fields remain unchanged

- [x] 2. Modify `map`/`mapb` for scoped pagination
  - [x] 2.1 Rewrite `commandMap` in `main.go` for scoped behavior
    - If `cfg.CurrentLocation == nil`: print "No location set. Use 'visit <name>' to visit a location first", return nil
    - If `len(cfg.CurrentLocation.Areas) == 0`: print "This location has no areas", return nil
    - Calculate page start: `cfg.ScopedPageIndex * 20`
    - If start >= len(areas): print "You're on the last page!", return nil
    - Calculate page end: `min((cfg.ScopedPageIndex+1)*20, len(areas))`
    - Print each area name in the slice `[start:end]`
    - Increment `cfg.ScopedPageIndex`
    - _Requirements: 2.1, 2.3, 2.5, 2.6, 2.7_

  - [x] 2.2 Rewrite `commandMapBack` in `main.go` for scoped behavior
    - If `cfg.CurrentLocation == nil`: print "No location set. Use 'visit <name>' to visit a location first", return nil
    - If `len(cfg.CurrentLocation.Areas) == 0`: print "This location has no areas", return nil
    - If `cfg.ScopedPageIndex <= 0`: print "You're on the first page!", return nil
    - Decrement `cfg.ScopedPageIndex`
    - Calculate page start: `cfg.ScopedPageIndex * 20`, end: `min((cfg.ScopedPageIndex+1)*20, len(areas))`
    - Print each area name in the slice `[start:end]`
    - _Requirements: 2.2, 2.4, 2.5, 2.6, 2.7_

  - [x] 2.3 Write property test for pagination correctness
    - **Property 4: Pagination Correctness**
    - **Validates: Requirements 2.1, 2.2, 2.6**
    - Use `rapid` to generate area lists of 0-100 items and random page indices, verify map returns correct slice and never exceeds 20 elements

- [ ] 3. Modify `commandExplore` for scoped validation
  - [~] 3.1 Rewrite `commandExplore` in `main.go` for scoped behavior
    - If `cfg.CurrentLocation == nil`: print "No location set. Use 'visit <name>' to visit a location first", return nil
    - If `len(args) != 1`: print "You must provide a location area to explore", return nil
    - Validate area name against `cfg.CurrentLocation.Areas`: if not found, print "Area '<name>' is not part of <location>. Use 'map' to see available areas", return nil (no API call)
    - Call `pokeapi.GetLocationAreaDetails(args[0], cache)`
    - On error: print "Error fetching area details: <error>", leave `cfg.CurrentArea` unchanged, return nil
    - On success: set `cfg.CurrentArea = args[0]`, print "Pokemons in <area>:" followed by each encounter name prefixed with "  - "
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

  - [~] 3.2 Write property test for area validation rejection
    - **Property 5: Area Validation Rejects Invalid Areas**
    - **Validates: Requirements 3.2**
    - Generate random Location and random area names NOT in the areas list, verify explore returns error and no API call is made

  - [~] 3.3 Write property test for successful explore updating state
    - **Property 6: Successful Explore Updates Active Area**
    - **Validates: Requirements 3.1, 3.4**
    - Generate random Location with non-empty areas, pick a valid area, verify Config.CurrentArea equals that area name after success

  - [~] 3.4 Write property test for failed explore preserving state
    - **Property 7: Failed Explore Preserves State**
    - **Validates: Requirements 3.5**
    - Generate random initial Config, simulate API failure after validation passes, verify CurrentArea unchanged

- [~] 4. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 5. Integrate dynamic prompt into REPL loop
  - [~] 5.1 Update `startREPL` in `repl.go` to use `buildPrompt`
    - Replace the hardcoded `"Pokedex > "` string in `reader.ReadLine(...)` with a call to `buildPrompt(cfg)`
    - Ensure prompt is recomputed each iteration so it reflects state changes from visit/explore commands
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

  - [~] 5.2 Write property test for prompt building
    - **Property 8: Prompt Building**
    - **Validates: Requirements 4.1, 4.2, 4.3**
    - Use `rapid` to generate random Config states (nil location, location set, area set) and verify the prompt format rules hold and always ends with "Pokedex > "

  - [~] 5.3 Write property test for name truncation
    - **Property 9: Name Truncation**
    - **Validates: Requirements 4.6**
    - Use `rapid` to generate random strings of varying lengths and verify: strings > 30 chars produce exactly 30 chars ending in "...", strings ≤ 30 return unchanged

- [~] 6. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties using the `rapid` library
- Unit tests validate specific examples and edge cases
- The existing `commandMap`/`commandMapBack` global pagination (using Next/Previous URLs) is replaced by scoped in-memory pagination — the `cfg.Next`/`cfg.Prev` fields may be removed or left unused after this feature
- The `Location` struct, `GetLocation` function, `Config` fields, `truncateName`, and `buildPrompt` are already implemented

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1"] },
    { "id": 1, "tasks": ["1.2", "1.3", "1.4", "1.5"] },
    { "id": 2, "tasks": ["2.1", "2.2", "3.1"] },
    { "id": 3, "tasks": ["2.3", "3.2", "3.3", "3.4"] },
    { "id": 4, "tasks": ["5.1"] },
    { "id": 5, "tasks": ["5.2", "5.3"] }
  ]
}
```
