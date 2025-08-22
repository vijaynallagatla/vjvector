// Package main provides the command-line interface for the VJVector database.
// It includes commands for serving the database, generating embeddings, and querying data.
package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/vijaynallagatla/vjvector/internal/server"
	"github.com/vijaynallagatla/vjvector/pkg/config"
	"github.com/vijaynallagatla/vjvector/pkg/utils/logger"
)

func main() {
	app := &cli.App{
		Name:        "vjvector",
		Usage:       "AI-first vector database for RAG applications",
		Description: "A high-performance vector database built from scratch with native AI embedding support",
		Version:     "0.1.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Path to configuration file",
				Value:   "config.yaml",
			},
			&cli.StringFlag{
				Name:    "log-level",
				Aliases: []string{"l"},
				Usage:   "Log level (debug, info, warn, error)",
				Value:   "info",
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "serve",
				Usage:  "Start the vector database server",
				Action: startServer,
			},
			{
				Name:   "embed",
				Usage:  "Generate embeddings for text",
				Action: generateEmbeddings,
			},
			{
				Name:   "query",
				Usage:  "Query the vector database",
				Action: queryDatabase,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func startServer(c *cli.Context) error {
	configPath := c.String("config")
	logLevel := c.String("log-level")

	// Initialize logger
	logger.Init(logLevel)

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Start server
	srv := server.New(cfg)
	return srv.Start()
}

func generateEmbeddings(_ *cli.Context) error {
	// TODO: Implement embedding generation
	fmt.Println("Embedding generation not yet implemented")
	return nil
}

func queryDatabase(_ *cli.Context) error {
	// TODO: Implement database querying
	fmt.Println("Database querying not yet implemented")
	return nil
}
