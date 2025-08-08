# Spells: DM Toolkit – Implementation Blueprint

## Blueprint — Implementation Spine

1. **Repository bootstrap**
   - Go module, dependency pinning, Justfile, CI workflow scaffold.
2. **Configuration layer**
   - XDG-compatible config resolver, YAML parser, defaults.
3. **Database core**
   - Embedded SQL migrations, connection pool, health check, transaction helper.
4. **Domain models + CRUD**
   - Sessions → NPCs → Places → Initiative → Timers/Events, each with unit tests.
5. **CLI skeleton**
   - Cobra root, sub-command registry: `init`, `track`, `oracle`, `gen`, `timer`, `debug`.
6. **TUI foundation**
   - Bubbletea program shell, model/update/view wiring, static panes.
7. **Session engine**
   - Turn counter, timer queue, trigger evaluator; integrated tests.
8. **Fuzzy search**
   - Trigram index against NPC/Place tables; ranked query function.
9. **Initiative tracker**
   - DB CRUD + live TUI pane; HP editing, sort order.
10. **Oracle system**
    - Perchance parser, dice roller, nested resolution; CLI + cache table.
11. **File-watch import**
    - fsnotify watcher, Markdown front-matter parser, hot-reload into DB.
12. **Content generation commands**
    - Name/rumor/NPC generators with deterministic seeding for testability.
13. **Multi-process state sharing**
    - SQLite WAL tuning, optimistic concurrency, conflict resolution tests.
14. **Packaging & release**
    - Static builds, Homebrew tap formula, GitHub Actions release workflow.
15. **Perf & fuzz tests**
    - Large-campaign benchmarks, go-fuzz table parser fuzzing.

---

## Milestone → Task Breakdown (right‑sized)

| Milestone  | Task ID | Atomic Task                                                              |
| ---------- | ------- | ------------------------------------------------------------------------ |
| Repo       | R1      | `go mod init`, `.gitignore`, MIT license                                 |
| Repo       | R2      | Justfile: `build`, `test`, `lint` targets                                |
| Repo       | R3      | GitHub Actions: Go 1.24 CI matrix, fail on `go vet` warnings             |
| Config     | C1      | Path resolver: `$XDG_CONFIG_HOME/spells` fallback to `~/.config/spells`  |
| Config     | C2      | Load + merge global YAML into struct with defaults test                  |
| DB         | D1      | Embed `migrations/*.sql` via `embed.FS`                                  |
| DB         | D2      | `Open()` helper returning `*sqlx.DB`, WAL, busy timeout                  |
| DB         | D3      | `migrate.Up()` idempotence test                                          |
| Model      | M1      | `sessions` DAO with Create/Get/AdvanceTurn methods + unit tests          |
| Model      | M2      | `npcs` DAO with search stub                                              |
| CLI        | L1      | Cobra root `spells`, global `--config` flag                              |
| CLI        | L2      | `init` subcommand: create empty DB + default config                      |
| TUI        | T1      | Bubbletea program skeleton printing “TODO” panes, exit cleanly           |
| Engine     | E1      | Turn-advance service emitting `TurnAdvanced` event, unit test            |
| Search     | S1      | Trigram index builder for strings, query function returning ranked slice |
| Initiative | I1      | `initiative_order` DAO + tests                                           |
| Initiative | I2      | TUI pane rendering static list from DAO mock                             |
| Oracle     | O1      | PEG parser for table syntax; parse smoke test                            |
| Oracle     | O2      | Dice roller `Roll("2d6+1")` with deterministic seed test                 |
| Oracle     | O3      | Resolve single-level table with mock data test                           |
| Watcher    | W1      | fsnotify watcher on `*.md`, debounce 250 ms                              |
| Watcher    | W2      | Front-matter YAML → NPC struct converter unit test                       |
| Gen        | G1      | `gen name --type tavern` command returning pseudo-random name            |
| State      | S2      | Concurrent write test: two processes advance turn without corruption     |
| Release    | P1      | `go build -ldflags "-s -w"` produce <10 MB binary test                   |
| Perf       | PF1     | Benchmark: search 1 000 NPCs must <100 ms                                |

All tasks are 30–90 min work units with explicit tests; no hidden dependency leaps.

---

## Code‑Generation Prompt Chain

> **Usage**: Feed each prompt verbatim to your code‑generation LLM. Wait for green tests before sending the next.

### Prompt 01 — Repository Bootstrap

```text
You are a Go code‑gen agent. Create initial repo skeleton.

Files:
- go.mod with module path `github.com/script-wizards/spells`
- .gitignore for Go, macOS, Linux
- LICENSE (CC0)
- Justfile with targets: build (go build ./cmd/spells), test (go test ./...), lint (go vet ./...)
- cmd/spells/main.go: Cobra root cmd “spells”, flags --config (string), --version
- internal/version/version.go: package version, var Version = "dev"
- cmd/spells/root_test.go: exec `spells --version` via `os/exec` expecting "dev\n"

CI:
- .github/workflows/ci.yml: Go 1.24, run just lint && just test

Return complete file contents. Do not include compiled output. Tests must pass `go test ./...`.
```

### Prompt 02 — Configuration Loader

```text
Extend repo.

Add:
- internal/config/config.go with type Config{ TorchDuration int `yaml:"torch_duration_turns"` }
- DefaultConfig() returning Config{TorchDuration:10}
- func Load(path string) (Config, error):
    * if path empty, resolve $XDG_CONFIG_HOME/spells/config.yaml else use supplied
    * merge defaults → override with file values
    * create dirs + empty file if not exist
- internal/config/config_test.go covering:
    * default fallback when no file
    * override TorchDuration:6 from temp YAML file

No changes elsewhere. All tests green.
```

### Prompt 03 — Migration Framework

```text
Add DB layer.

Files:
- internal/db/migrate.go with `RunMigrations(db *sqlx.DB, fs embed.FS) error`
- embed folder migrations/0001_init.sql containing `CREATE TABLE schema_version (id int);`
- internal/db/connection.go with Open(path string) (*sqlx.DB, error):
    * sqlite3, WAL on, busy_timeout=5000ms
    * call RunMigrations
- internal/db/connection_test.go:
    * temp dir DB, verify schema_version table exists after Open()

Use github.com/jmoiron/sqlx and modernc.org/sqlite (CGO-free). Tests must pass.
```

### Prompt 04 — `init` Sub‑Command

```text
Implement CLI `spells init`.

Behavior:
- Accept `--path` (default "./campaign.db")
- Call db.Open(), create default config YAML alongside DB
- Print "initialized" on success
Tests:
- cmd/init_test.go launches `spells init --path ./tmp/test.db`, asserts file exists and exit 0.

Revise Cobra setup to register subcommand. All prior tests remain green.
```

### Prompt 05 — Session Model

```text
Add sessions DAO.

Files:
- internal/model/session.go with type Session{ID int64, CurrentTurn int64}
- Methods: Create(tx), Get(db, id), AdvanceTurn(tx, delta) (updates CurrentTurn)
- internal/model/session_test.go:
    * create session, advance 3 turns, verify value

SQL migration: migrations/0002_sessions.sql creating table per spec.

Ensure migrations auto-apply. All tests green.
```

### Prompt 06 — Bubbletea Shell

```text
Add TUI skeleton.

Files:
- internal/tui/model.go implementing bubbletea.Model with Init(Update)(View) showing placeholder panes.
- cmd/spells/track.go Cobra cmd `track` that opens DB, starts tea.NewProgram(model).
- internal/tui/model_test.go: smoke test Init→Update(tea.KeyMsg{Type:tea.KeyCtrlC}) returns tea.Quit.

Dependency: github.com/charmbracelet/bubbletea v0.25. All tests green, CLI `spells track` launches without panic.
```

### Prompt 07 — Turn Engine

```text
Turn engine service.

Add internal/engine/turn.go:
- type Engine struct{DB *sqlx.DB}
- func (e *Engine) Advance(delta int64) error:
    * within tx: session.AdvanceTurn(...)
    * emit log line "TURN_ADVANCED n→n+delta"

Unit test: advance twice, assert final turn count.

Wire into TUI: spacebar advances 1 turn, rerender turn count.

All tests green.
```

### Prompt 08 — Trigram Fuzzy Search

```text
Search package.

Files:
- internal/search/trigram.go with BuildIndex([]string) Index and Query(idx, q string, limit int) []Match{Value,Score}
- Simple trigram cosine similarity.

Unit test: index ["kobold","goblin"], query "kobol" returns "kobold" rank 0>.

No production wiring yet.
```

### Prompt 09 — NPC Model + Search Wire‑up

```text
Add NPC table + DAO.

Migration 0003_npcs.sql per spec (id,name,description,...).

Model methods: CreateNPC, SearchNPC(search.Index, query, limit).

TUI: “/” key enters search mode; display top 5 NPC names from DB via trigram search.

Tests:
- dao test for insert/find
- search integration test: create 3 NPCs, query returns expected.

All tests green.
```

### Prompt 10 — Initiative Tracker Pane

```text
Add initiative tables (encounter, initiative_order) migration 0004.

DAO: AddCombatant, ListActiveBySort.

TUI: new pane shows ordered list with HP; "i" enters add-combatant modal (stub).

Unit test covers DAO ordering.

All tests green.
```

### Prompt 11 — Dice Roller

```text
Utility.

internal/dice/dice.go with Roll(expr string, rng *rand.Rand) (total int, breakdown []int, err).

Support NdM+K syntax.

Tests: "2d6+1" seeded rng returns deterministic slice [3,4]+1 total 8.

No wiring yet.
```

### Prompt 12 — Oracle Parser + Resolver

```text
Parser.

internal/oracle/parser.go using pigeon/peg to parse `{option1|option2}`, `[table]`, dice.

Resolver resolves single level using in-memory map[string]string table.

Unit tests cover:
- Parse string with nested refs
- Resolve with dice roller stub

Add CLI `spells oracle "1d4 rat"` returning JSON dump. Tests for CLI.

All tests green.
```

### Prompt 13 — File Watcher Import

```text
Watcher.

internal/importer/watch.go start fsnotify watcher on `*.md`, on change parse front-matter:

```

---

## type: npc name: Thorg tags: [orc,warlord]

Body text...

```

Convert to NPC, upsert.

Tests: temp dir, write file, await channel event, verify DB row.

Wire watcher into `track` command on startup.

All tests green.
```

### Prompt 14 — Concurrency Safeguard

```text
Add optimistic concurrency.

internal/db/tx.go wraps Exec with retry on SQLITE_BUSY up to 3 attempts with backoff.

Integration test: two goroutines advance turns concurrently 1000×, final turn == 2000.

All tests green.
```

### Prompt 15 — Release Pipeline

```text
Add GitHub release action.

- tag push triggers build for linux/amd64, darwin/arm64 with `-ldflags "-X internal/version.Version=$TAG"`
- upload artifacts, generate SHA256.

No tests; CI must pass.

End wiring: update --version flag to print internal/version.Version.
```
