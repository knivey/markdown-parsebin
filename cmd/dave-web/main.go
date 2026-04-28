package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/knivey/dave-web/internal/db"
	mcpserver "github.com/knivey/dave-web/internal/mcp"
	"github.com/knivey/dave-web/internal/ttl"
	"github.com/knivey/dave-web/internal/web"
)

//go:embed all:migrations
var migrationsFS embed.FS

//go:embed all:templates
var templatesFS embed.FS

//go:embed all:static
var staticFS embed.FS

var (
	webAddr   string
	mcpAddr   string
	dbPath    string
	baseURL   string
	keyDesc   string
	keyRevoke string
)

func main() {
	rootCmd := buildRootCmd()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func buildRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "dave-web",
		Short: "Markdown pastebin with MCP server",
	}

	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "dave-web.db", "Path to SQLite database")
	rootCmd.PersistentFlags().StringVar(&baseURL, "base-url", "http://localhost:8080", "Base URL for paste links")

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the web and MCP servers",
		RunE:  runServe,
	}
	serveCmd.Flags().StringVar(&webAddr, "addr", ":8080", "Web server listen address")
	serveCmd.Flags().StringVar(&mcpAddr, "mcp-addr", ":8081", "MCP server listen address")

	rootCmd.AddCommand(serveCmd, keysCmd())

	return rootCmd
}

func openDB() (*db.DB, error) {
	return db.New(dbPath, migrationsFS)
}

func getTemplateFS() fs.FS {
	sub, err := fs.Sub(templatesFS, "templates")
	if err != nil {
		log.Fatal(err)
	}
	return sub
}

func getStaticFS() fs.FS {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatal(err)
	}
	return sub
}

func runServe(cmd *cobra.Command, args []string) error {
	database, err := openDB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer database.Close()

	ttl.StartCleaner(database, 5*time.Minute)

	mcpSrv := mcpserver.NewMCPServer(database, baseURL)
	go func() {
		if err := mcpSrv.Run(mcpAddr); err != nil {
			log.Fatalf("MCP server error: %v", err)
		}
	}()

	srv := web.NewServer(database, getTemplateFS(), getStaticFS(), baseURL)

	log.Printf("Web server starting on %s", webAddr)
	return srv.Run(webAddr)
}

func keysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Manage API keys",
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := openDB()
			if err != nil {
				return err
			}
			defer database.Close()

			ak, err := database.CreateAPIKey(keyDesc)
			if err != nil {
				return err
			}

			fmt.Println(ak.Key)
			return nil
		},
	}
	createCmd.Flags().StringVar(&keyDesc, "description", "", "Description for the key")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all API keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := openDB()
			if err != nil {
				return err
			}
			defer database.Close()

			keys, err := database.ListAPIKeys()
			if err != nil {
				return err
			}

			if len(keys) == 0 {
				fmt.Println("No API keys found.")
				return nil
			}

			fmt.Printf("%-24s  %-30s  %s\n", "KEY", "DESCRIPTION", "CREATED")
			for _, ak := range keys {
				short := ak.Key[:12] + "..."
				fmt.Printf("%-24s  %-30s  %s\n", short, ak.Description, ak.CreatedAt.Format("2006-01-02 15:04"))
			}
			return nil
		},
	}

	revokeCmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke an API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := openDB()
			if err != nil {
				return err
			}
			defer database.Close()

			if len(keyRevoke) < 12 {
			return fmt.Errorf("invalid key: too short")
		}
		if err := database.DeleteAPIKey(keyRevoke); err != nil {
				return err
			}

			fmt.Printf("Key %s... revoked\n", keyRevoke[:12])
			return nil
		},
	}
	revokeCmd.Flags().StringVar(&keyRevoke, "key", "", "Full API key to revoke")
	_ = revokeCmd.MarkFlagRequired("key")

	cmd.AddCommand(createCmd, listCmd, revokeCmd)
	return cmd
}
