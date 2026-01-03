package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/aquamarinepk/aqm/config"
	"github.com/aquamarinepk/aqm/log"
	"github.com/aquamarinepk/aqm/migrate"
)

type Database struct {
	DB            *sql.DB
	assetsFS      embed.FS
	engine        string
	migrationPath string
	cfg           *config.Config
	log           logger.Logger
}

func New(assetsFS embed.FS, engine string, cfg *config.Config, log logger.Logger) *Database {
	return &Database{
		assetsFS: assetsFS,
		engine:   engine,
		cfg:      cfg,
		log:      log,
	}
}

func (d *Database) SetMigrationPath(path string) {
	d.migrationPath = path
}

func (d *Database) Start(ctx context.Context) error {
	db, err := sql.Open("pgx", d.cfg.Database.ConnectionString())
	if err != nil {
		return fmt.Errorf("cannot open database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("cannot ping database: %w", err)
	}

	d.DB = db
	d.log.Info("Database connection established")

	if err := d.ensureSchema(ctx); err != nil {
		return fmt.Errorf("cannot ensure schema: %w", err)
	}

	migrator := migrate.New(d.assetsFS, d.engine, d.log)
	migrator.SetDB(d.DB)
	if d.migrationPath != "" {
		migrator.SetPath(d.migrationPath)
	}
	if err := migrator.Run(ctx); err != nil {
		return fmt.Errorf("cannot run migrations: %w", err)
	}

	return nil
}

func (d *Database) Stop(ctx context.Context) error {
	if d.DB != nil {
		d.log.Info("Closing database connection")
		return d.DB.Close()
	}
	return nil
}

func (d *Database) GetDB() *sql.DB {
	return d.DB
}

func (d *Database) ensureSchema(ctx context.Context) error {
	if d.cfg.Database.Schema == "" {
		return nil
	}

	query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", d.cfg.Database.Schema)
	if _, err := d.DB.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("cannot create schema %s: %w", d.cfg.Database.Schema, err)
	}

	d.log.Infof("Schema %s ensured", d.cfg.Database.Schema)
	return nil
}
