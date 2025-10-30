package cli

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"solecode/pkg/config"
	"solecode/pkg/database"

	"github.com/spf13/cobra"
)

var (
	cfgFile       string
	migrationsDir string
)

var rootCmd = &cobra.Command{
	Use:   "userapi",
	Short: "User Management API",
	Long:  "A REST API for user management with MySQL and Redis",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		runMigrations("up")
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "migrate-down",
	Short: "Rollback database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		runMigrations("down")
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "migrate-status",
	Short: "Show migration status",
	Run: func(cmd *cobra.Command, args []string) {
		showMigrationStatus()
	},
}

var migrateCreateCmd = &cobra.Command{
	Use:   "migrate-create [name]",
	Short: "Create new migration files",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		createMigration(args[0])
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show application version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("User Management API v1.0.0")
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "conf/conf.yaml", "config file path")
	rootCmd.PersistentFlags().StringVar(&migrationsDir, "migrations-dir", "docs/migrations", "migrations directory")

	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(migrateDownCmd)
	rootCmd.AddCommand(migrateStatusCmd)
	rootCmd.AddCommand(migrateCreateCmd)
	rootCmd.AddCommand(versionCmd)
}

// Migration represents a database migration
type Migration struct {
	Version string
	Name    string
	UpSQL   string
	DownSQL string
}

func runMigrations(direction string) {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewMySQLDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Ensure migration log table exists
	if err := createMigrationLogTable(db); err != nil {
		log.Fatalf("Failed to create migration log table: %v", err)
	}

	migrations, err := loadMigrations()
	if err != nil {
		log.Fatalf("Failed to load migrations: %v", err)
	}

	if direction == "up" {
		fmt.Println("Running migrations up...")
		if err := migrateUp(db, migrations); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("✅ All migrations completed successfully")
	} else {
		fmt.Println("Running migrations down...")
		if err := migrateDown(db, migrations); err != nil {
			log.Fatalf("Migration rollback failed: %v", err)
		}
		fmt.Println("✅ Migrations rolled back successfully")
	}
}

func loadMigrations() ([]Migration, error) {
	var migrations []Migration

	// Read all .sql files in migrations directory
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Group files by version and name
	migrationFiles := make(map[string]struct {
		up   string
		down string
	})

	for _, file := range files {
		filename := filepath.Base(file)

		// Parse filename format: {version}_{name}.{up|down}.sql
		parts := strings.Split(strings.TrimSuffix(filename, ".sql"), ".")
		if len(parts) != 2 {
			continue // Skip files that don't match pattern
		}

		baseName := parts[0]
		direction := parts[1]

		// Split base name to get version and migration name
		baseParts := strings.SplitN(baseName, "_", 2)
		if len(baseParts) != 2 {
			continue
		}

		version := baseParts[0]
		name := baseParts[1]

		key := version + "_" + name

		if _, exists := migrationFiles[key]; !exists {
			migrationFiles[key] = struct {
				up   string
				down string
			}{}
		}

		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		fileData := migrationFiles[key]
		if direction == "up" {
			fileData.up = string(content)
		} else if direction == "down" {
			fileData.down = string(content)
		}
		migrationFiles[key] = fileData
	}

	// Convert to Migration slices and sort by version
	for key, files := range migrationFiles {
		parts := strings.SplitN(key, "_", 2)
		migrations = append(migrations, Migration{
			Version: parts[0],
			Name:    parts[1],
			UpSQL:   files.up,
			DownSQL: files.down,
		})
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func migrateUp(db *sql.DB, migrations []Migration) error {
	// Get applied migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	appliedMap := make(map[string]bool)
	for _, m := range applied {
		appliedMap[m] = true
	}

	for _, migration := range migrations {
		if appliedMap[migration.Version] {
			fmt.Printf("✓ Migration %s (%s) already applied\n", migration.Version, migration.Name)
			continue
		}

		fmt.Printf("→ Applying migration %s (%s)...\n", migration.Version, migration.Name)

		// Start transaction
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Execute migration
		if _, err := tx.Exec(migration.UpSQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", migration.Version, err)
		}

		// Record migration
		if err := recordMigration(tx, migration.Version, migration.Name, "up"); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		fmt.Printf("✅ Applied migration %s (%s)\n", migration.Version, migration.Name)
	}

	return nil
}

func migrateDown(db *sql.DB, migrations []Migration) error {
	// Get applied migrations in reverse order
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Sort migrations in reverse order
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version > migrations[j].Version
	})

	for _, migration := range migrations {
		if !contains(applied, migration.Version) {
			continue
		}

		fmt.Printf("→ Rolling back migration %s (%s)...\n", migration.Version, migration.Name)

		// Start transaction
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Execute rollback
		if _, err := tx.Exec(migration.DownSQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to rollback migration %s: %w", migration.Version, err)
		}

		// Remove migration record
		if err := removeMigration(tx, migration.Version); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to remove migration record %s: %w", migration.Version, err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		fmt.Printf("✅ Rolled back migration %s (%s)\n", migration.Version, migration.Name)
	}

	return nil
}

func showMigrationStatus() {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewMySQLDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Ensure migration log table exists
	if err := createMigrationLogTable(db); err != nil {
		log.Fatalf("Failed to create migration log table: %v", err)
	}

	migrations, err := loadMigrations()
	if err != nil {
		log.Fatalf("Failed to load migrations: %v", err)
	}

	applied, err := getAppliedMigrations(db)
	if err != nil {
		log.Fatalf("Failed to get applied migrations: %v", err)
	}

	appliedMap := make(map[string]bool)
	for _, m := range applied {
		appliedMap[m] = true
	}

	fmt.Println("📊 Migration Status:")
	fmt.Println("====================")

	for _, migration := range migrations {
		status := "❌ Pending"
		if appliedMap[migration.Version] {
			status = "✅ Applied"
		}
		fmt.Printf("%s %s - %s\n", status, migration.Version, migration.Name)
	}

	fmt.Printf("\nTotal: %d migrations (%d applied, %d pending)\n",
		len(migrations), len(applied), len(migrations)-len(applied))
}

func createMigration(name string) {
	timestamp := time.Now().Format("20060102150405")

	upFile := filepath.Join(migrationsDir, fmt.Sprintf("%s_%s.up.sql", timestamp, name))
	downFile := filepath.Join(migrationsDir, fmt.Sprintf("%s_%s.down.sql", timestamp, name))

	// Create migrations directory if it doesn't exist
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		log.Fatalf("Failed to create migrations directory: %v", err)
	}

	// Create up migration file
	upContent := fmt.Sprintf("-- Migration: %s\n-- Version: %s\n-- Description: %s\n\n", name, timestamp, name)
	if err := os.WriteFile(upFile, []byte(upContent), 0644); err != nil {
		log.Fatalf("Failed to create up migration file: %v", err)
	}

	// Create down migration file
	downContent := fmt.Sprintf("-- Rollback: %s\n-- Version: %s\n\n", name, timestamp)
	if err := os.WriteFile(downFile, []byte(downContent), 0644); err != nil {
		log.Fatalf("Failed to create down migration file: %v", err)
	}

	fmt.Printf("✅ Created migration files:\n")
	fmt.Printf("   Up: %s\n", upFile)
	fmt.Printf("   Down: %s\n", downFile)
}

// Migration log table functions
func createMigrationLogTable(db *sql.DB) error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS migration_log (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			version VARCHAR(255) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			direction VARCHAR(10) NOT NULL,
			INDEX idx_version (version),
			INDEX idx_applied_at (applied_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`
	_, err := db.Exec(createTableSQL)
	return err
}

func recordMigration(tx *sql.Tx, version, name, direction string) error {
	query := `
		INSERT INTO migration_log (version, name, direction) 
		VALUES (?, ?, ?)
	`
	_, err := tx.Exec(query, version, name, direction)
	return err
}

func removeMigration(tx *sql.Tx, version string) error {
	query := `DELETE FROM migration_log WHERE version = ?`
	_, err := tx.Exec(query, version)
	return err
}

func getAppliedMigrations(db *sql.DB) ([]string, error) {
	var migrations []string

	query := `
		SELECT version FROM migration_log 
		WHERE direction = 'up' 
		ORDER BY applied_at ASC
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		migrations = append(migrations, version)
	}

	return migrations, nil
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
