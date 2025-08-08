# Spells DM Toolkit — TODO Checklist

> Mark each task with **`[x]`** when done. ID codes correspond to the detailed blueprint.

## 1 — Repository Bootstrap

- [x] **R1**  `go mod init`, `.gitignore`, MIT license
- [x] **R2**  Makefile: `build`, `test`, `lint` targets
- [x] **R3**  GitHub Actions: Go 1.22 CI matrix, fail on `go vet` warnings

## 2 — Configuration Layer

- [ ] **C1**  Path resolver: `$XDG_CONFIG_HOME/spells` fallback to `~/.config/spells`
- [ ] **C2**  Load + merge global YAML into struct with defaults test

## 3 — Database Core

- [ ] **D1**  Embed `migrations/*.sql` via `embed.FS`
- [ ] **D2**  `Open()` helper returning `*sqlx.DB`, enable WAL, busy timeout
- [ ] **D3**  `migrate.Up()` idempotence test

## 4 — Domain Models & CRUD

- [ ] **M1**  `sessions` DAO with Create / Get / AdvanceTurn + unit tests
- [ ] **M2**  `npcs` DAO with search stub

## 5 — CLI Skeleton

- [ ] **L1**  Cobra root `spells`, global `--config` flag
- [ ] **L2**  `init` sub‑command: create empty DB + default config

## 6 — TUI Foundation

- [ ] **T1**  Bubbletea program skeleton with placeholder panes, clean exit

## 7 — Session Engine

- [ ] **E1**  Turn‑advance service emitting `TurnAdvanced` event + unit test

## 8 — Fuzzy Search

- [ ] **S1**  Trigram index builder & query function with ranked output

## 9 — Initiative Tracker

- [ ] **I1**  `initiative_order` DAO + tests
- [ ] **I2**  TUI pane rendering list from DAO mock

## 10 — Oracle System

- [ ] **O1**  PEG parser for table syntax; parse smoke test
- [ ] **O2**  Dice roller `Roll("2d6+1")` deterministic seed test
- [ ] **O3**  Resolve single‑level table with mock data test

## 11 — File‑Watch Import

- [ ] **W1**  fsnotify watcher on `*.md`, debounce 250 ms
- [ ] **W2**  Front‑matter YAML → NPC struct converter unit test

## 12 — Content Generation Commands

- [ ] **G1**  `gen name --type tavern` command returning pseudo‑random name

## 13 — Multi‑Process State Sharing

- [ ] **S2**  Concurrent write test: two processes advance turn without corruption

## 14 — Packaging & Release

- [ ] **P1**  `go build -ldflags "-s -w"` produce <10 MB binary test

## 15 — Performance & Fuzz Tests

- [ ] **PF1**  Benchmark: search 1 000 NPCs must <100 ms

---

### Optional Stretch Items

- [ ] Static builds for additional OS/ARCH
- [ ] Homebrew tap formula
- [ ] GitHub Release Action producing checksums & artifacts
