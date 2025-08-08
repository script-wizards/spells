package importer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/jmoiron/sqlx"
	"github.com/script-wizards/spells/internal/model"
	"gopkg.in/yaml.v3"
)

type FrontMatter struct {
	Type string   `yaml:"type"`
	Name string   `yaml:"name"`
	Tags []string `yaml:"tags"`
}

type Watcher struct {
	db      *sqlx.DB
	watcher *fsnotify.Watcher
	done    chan bool
}

func NewWatcher(db *sqlx.DB) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	return &Watcher{
		db:      db,
		watcher: watcher,
		done:    make(chan bool),
	}, nil
}

func (w *Watcher) Start(watchPath string) error {
	err := w.addMarkdownFiles(watchPath)
	if err != nil {
		return fmt.Errorf("failed to add markdown files: %w", err)
	}

	go func() {
		defer w.watcher.Close()
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				if w.shouldProcessEvent(event) {
					if err := w.processFile(event.Name); err != nil {
						log.Printf("Error processing file %s: %v", event.Name, err)
					}
				}
			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Watcher error: %v", err)
			case <-w.done:
				return
			}
		}
	}()

	return nil
}

func (w *Watcher) Stop() {
	close(w.done)
}

func (w *Watcher) addMarkdownFiles(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return w.watcher.Add(path)
		}

		if strings.HasSuffix(path, ".md") {
			dir := filepath.Dir(path)
			return w.watcher.Add(dir)
		}

		return nil
	})
}

func (w *Watcher) shouldProcessEvent(event fsnotify.Event) bool {
	return (event.Has(fsnotify.Write) || event.Has(fsnotify.Create)) &&
		strings.HasSuffix(event.Name, ".md")
}

func (w *Watcher) processFile(filename string) error {
	frontMatter, bodyText, err := w.parseFrontMatter(filename)
	if err != nil {
		return fmt.Errorf("failed to parse front matter: %w", err)
	}

	if frontMatter == nil || frontMatter.Type != "npc" {
		return nil
	}

	npc := w.convertToNPC(frontMatter, bodyText)
	return w.upsertNPC(npc)
}

func (w *Watcher) parseFrontMatter(filename string) (*FrontMatter, string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var allLines []string
	var inFrontMatter bool
	var frontMatterStarted bool
	var frontMatterLines []string
	var bodyLines []string

	for scanner.Scan() {
		line := scanner.Text()
		allLines = append(allLines, line)

		if line == "---" {
			if !frontMatterStarted {
				frontMatterStarted = true
				inFrontMatter = true
				continue
			} else if inFrontMatter {
				inFrontMatter = false
				continue
			}
		}

		if inFrontMatter {
			frontMatterLines = append(frontMatterLines, line)
		} else if frontMatterStarted && !inFrontMatter {
			bodyLines = append(bodyLines, line)
		} else if !frontMatterStarted {
			bodyLines = append(bodyLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, "", err
	}

	if len(frontMatterLines) == 0 {
		return nil, strings.Join(allLines, "\n"), nil
	}

	var fm FrontMatter
	yamlContent := strings.Join(frontMatterLines, "\n")
	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return nil, "", fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &fm, strings.TrimSpace(strings.Join(bodyLines, "\n")), nil
}

func (w *Watcher) convertToNPC(fm *FrontMatter, body string) *model.NPC {
	var description *string
	if body != "" {
		body = strings.TrimSpace(body)
		if body != "" {
			description = &body
		}
	}

	var tags *string
	if len(fm.Tags) > 0 {
		tagsJSON, _ := json.Marshal(fm.Tags)
		tagsStr := string(tagsJSON)
		tags = &tagsStr
	}

	return &model.NPC{
		Name:        fm.Name,
		Description: description,
		Status:      "neutral",
		Tags:        tags,
	}
}

func (w *Watcher) upsertNPC(npc *model.NPC) error {
	tx, err := w.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var existingNPC model.NPC
	query := `SELECT id, name, description, location, status, motivation, secrets, tags, 
			  last_mentioned, created_at FROM npcs WHERE name = ?`
	err = tx.Get(&existingNPC, query, npc.Name)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			if err := model.CreateNPC(tx, npc); err != nil {
				return fmt.Errorf("failed to create NPC: %w", err)
			}
		} else {
			return fmt.Errorf("failed to query existing NPC: %w", err)
		}
	} else {
		updateQuery := `UPDATE npcs SET description = ?, tags = ? WHERE id = ?`
		_, err = tx.Exec(updateQuery, npc.Description, npc.Tags, existingNPC.ID)
		if err != nil {
			return fmt.Errorf("failed to update NPC: %w", err)
		}
		npc.ID = existingNPC.ID
	}

	return tx.Commit()
}
