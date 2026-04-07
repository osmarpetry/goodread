package main

import (
	"fmt"
	"os"
	"path/filepath"

	"media2goodreads/internal/api"
	"media2goodreads/internal/export"
	"media2goodreads/internal/ingest"
	"media2goodreads/internal/model"
	"media2goodreads/internal/tui"
	"media2goodreads/internal/util"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
	fs      = afero.NewOsFs()
	logger  *util.Logger
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "media2goodreads",
	Short: "Import media libraries and export to Goodreads CSV",
	Long: `media2goodreads consolidates your reading history from Audible, Kindle, 
and Storytel into a unified JSON library, then exports it to a Goodreads-compatible CSV.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
		if verbose {
			logger = util.NewLogger(util.LogLevelVerbose)
		} else {
			logger = util.NewLogger(util.LogLevelNormal)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .env)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Add commands
	rootCmd.AddCommand(audibleCmd)
	rootCmd.AddCommand(kindleCmd)
	rootCmd.AddCommand(storytelCmd)
	rootCmd.AddCommand(goodreadsCSVCmd)
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(webCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".env")
		viper.SetConfigType("env")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME")
	}

	viper.SetEnvPrefix("MEDIA2GR")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("OUT_DIR", "./out")
	viper.SetDefault("TZ", "America/Sao_Paulo")
	viper.SetDefault("DEFAULT_SHELF", "read")
	viper.SetDefault("AUDIBLE_FORMAT", "json")
	viper.SetDefault("AUDIBLE_REGION", "us")
	viper.SetDefault("STORYTEL_FORMAT", "csv")

	// Ignore error if config file doesn't exist
	_ = viper.ReadInConfig()
}

// Audible import command
var audibleCmd = &cobra.Command{
	Use:   "audible",
	Short: "Audible import commands",
}

var audibleImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import Audible library from OpenAudible export",
	RunE: func(cmd *cobra.Command, args []string) error {
		from, _ := cmd.Flags().GetString("from")
		format, _ := cmd.Flags().GetString("format")
		region, _ := cmd.Flags().GetString("region")
		
		if from == "" {
			from = viper.GetString("AUDIBLE_EXPORT_PATH")
		}
		if from == "" {
			return fmt.Errorf("--from flag or MEDIA2GR_AUDIBLE_EXPORT_PATH required")
		}

		timezone := viper.GetString("TZ")
		outDir := viper.GetString("OUT_DIR")
		libraryPath := filepath.Join(outDir, "library.json")

		logger.Info("Importing Audible library from %s (format: %s, region: %s)", from, format, region)

		var items []model.BookItem
		var err error

		if format == "csv" {
			items, err = ingest.ParseAudibleCSV(fs, from, timezone)
		} else {
			items, err = ingest.ParseAudibleJSON(fs, from, timezone)
		}

		if err != nil {
			return fmt.Errorf("failed to parse Audible export: %w", err)
		}

		logger.Info("Parsed %d items", len(items))

		// Load existing library
		existingLib, _ := util.LoadLibrary(fs, libraryPath)

		// Merge and save
		merged, err := ingest.MergeIntoLibrary(fs, libraryPath, items)
		if err != nil {
			return err
		}

		stats := ingest.CalculateImportStats(existingLib, merged, len(items))
		logger.Success("Imported %d items (%d new, %d merged)", stats.Total, stats.Added, stats.Merged)
		logger.Info("Library saved to %s", libraryPath)

		return nil
	},
}

func init() {
	audibleImportCmd.Flags().String("from", "", "Path to OpenAudible export file")
	audibleImportCmd.Flags().String("format", "json", "Export format (json|csv)")
	audibleImportCmd.Flags().String("region", "us", "Audible region (us|uk|de|br)")
	audibleCmd.AddCommand(audibleImportCmd)
}

// Kindle import command
var kindleCmd = &cobra.Command{
	Use:   "kindle",
	Short: "Kindle import commands",
}

var kindleImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import Kindle library from My Clippings.txt",
	RunE: func(cmd *cobra.Command, args []string) error {
		clippings, _ := cmd.Flags().GetString("clippings")
		notebookDir, _ := cmd.Flags().GetString("notebook-html")

		if clippings == "" {
			clippings = viper.GetString("KINDLE_CLIPPINGS_PATH")
		}
		if notebookDir == "" {
			notebookDir = viper.GetString("KINDLE_NOTEBOOK_HTML_DIR")
		}

		timezone := viper.GetString("TZ")
		outDir := viper.GetString("OUT_DIR")
		libraryPath := filepath.Join(outDir, "library.json")

		var allItems []model.BookItem

		// Parse clippings if provided
		if clippings != "" {
			logger.Info("Importing Kindle clippings from %s", clippings)
			items, err := ingest.ParseKindleClippings(fs, clippings, timezone)
			if err != nil {
				return fmt.Errorf("failed to parse Kindle clippings: %w", err)
			}
			logger.Info("Parsed %d items from clippings", len(items))
			allItems = append(allItems, items...)
		}

		// Parse notebook HTML if provided
		if notebookDir != "" {
			logger.Info("Importing Kindle notebooks from %s", notebookDir)
			items, err := ingest.ParseKindleNotebookHTML(fs, notebookDir, timezone)
			if err != nil {
				logger.Warn("Failed to parse Kindle notebooks: %v", err)
			} else {
				logger.Info("Parsed %d items from notebooks", len(items))
				allItems = append(allItems, items...)
			}
		}

		if len(allItems) == 0 {
			return fmt.Errorf("no items to import; provide --clippings or --notebook-html")
		}

		// Load existing library
		existingLib, _ := util.LoadLibrary(fs, libraryPath)

		// Merge and save
		merged, err := ingest.MergeIntoLibrary(fs, libraryPath, allItems)
		if err != nil {
			return err
		}

		stats := ingest.CalculateImportStats(existingLib, merged, len(allItems))
		logger.Success("Imported %d items (%d new, %d merged)", stats.Total, stats.Added, stats.Merged)
		logger.Info("Library saved to %s", libraryPath)

		return nil
	},
}

func init() {
	kindleImportCmd.Flags().String("clippings", "", "Path to My Clippings.txt")
	kindleImportCmd.Flags().String("notebook-html", "", "Directory containing Kindle Notebook HTML files")
	kindleCmd.AddCommand(kindleImportCmd)
}

// Storytel import command
var storytelCmd = &cobra.Command{
	Use:   "storytel",
	Short: "Storytel import commands",
}

var storytelImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import Storytel library from export file",
	RunE: func(cmd *cobra.Command, args []string) error {
		from, _ := cmd.Flags().GetString("from")
		format, _ := cmd.Flags().GetString("format")

		if from == "" {
			from = viper.GetString("STORYTEL_EXPORT_PATH")
		}
		if from == "" {
			return fmt.Errorf("--from flag or MEDIA2GR_STORYTEL_EXPORT_PATH required")
		}

		timezone := viper.GetString("TZ")
		outDir := viper.GetString("OUT_DIR")
		libraryPath := filepath.Join(outDir, "library.json")

		logger.Info("Importing Storytel library from %s (format: %s)", from, format)

		var items []model.BookItem
		var err error

		if format == "csv" {
			items, err = ingest.ParseStorytelCSV(fs, from, timezone)
		} else {
			items, err = ingest.ParseStorytelJSON(fs, from, timezone)
		}

		if err != nil {
			return fmt.Errorf("failed to parse Storytel export: %w", err)
		}

		logger.Info("Parsed %d items", len(items))

		// Load existing library
		existingLib, _ := util.LoadLibrary(fs, libraryPath)

		// Merge and save
		merged, err := ingest.MergeIntoLibrary(fs, libraryPath, items)
		if err != nil {
			return err
		}

		stats := ingest.CalculateImportStats(existingLib, merged, len(items))
		logger.Success("Imported %d items (%d new, %d merged)", stats.Total, stats.Added, stats.Merged)
		logger.Info("Library saved to %s", libraryPath)

		return nil
	},
}

func init() {
	storytelImportCmd.Flags().String("from", "", "Path to Storytel export file")
	storytelImportCmd.Flags().String("format", "csv", "Export format (json|csv)")
	storytelCmd.AddCommand(storytelImportCmd)
}

// Goodreads CSV export command
var goodreadsCSVCmd = &cobra.Command{
	Use:   "goodreads-csv",
	Short: "Export library to Goodreads-compatible CSV",
	RunE: func(cmd *cobra.Command, args []string) error {
		inPath, _ := cmd.Flags().GetString("in")
		outPath, _ := cmd.Flags().GetString("out")
		shelf, _ := cmd.Flags().GetString("shelf")
		added, _ := cmd.Flags().GetString("added")

		outDir := viper.GetString("OUT_DIR")
		if inPath == "" {
			inPath = filepath.Join(outDir, "library.json")
		}
		if outPath == "" {
			outPath = filepath.Join(outDir, "goodreads_import.csv")
		}
		if shelf == "" {
			shelf = viper.GetString("DEFAULT_SHELF")
		}
		if added == "" {
			added = viper.GetString("DATE_ADDED")
		}

		timezone := viper.GetString("TZ")

		logger.Info("Exporting library to Goodreads CSV")
		logger.Info("Input: %s", inPath)
		logger.Info("Output: %s", outPath)

		// Load library
		library, err := util.LoadLibrary(fs, inPath)
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		logger.Info("Loaded %d items", len(library.Items))

		// Export to CSV
		opts := export.ExportOptions{
			Shelf:     shelf,
			DateAdded: added,
			Timezone:  timezone,
		}

		if err := export.ExportGoodreadsCSV(fs, library, outPath, opts); err != nil {
			return fmt.Errorf("failed to export CSV: %w", err)
		}

		logger.Success("Exported %d items to %s", len(library.Items), outPath)
		logger.Info("You can now import this CSV at https://www.goodreads.com/review/import")

		return nil
	},
}

func init() {
	goodreadsCSVCmd.Flags().String("in", "", "Input library JSON path")
	goodreadsCSVCmd.Flags().String("out", "", "Output CSV path")
	goodreadsCSVCmd.Flags().String("shelf", "", "Exclusive shelf (read|to-read|currently-reading)")
	goodreadsCSVCmd.Flags().String("added", "", "Date added (YYYY-MM-DD)")
}

// TUI command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive text-based UI",
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.Run()
	},
}

// Web command
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Launch web UI server",
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, _ := cmd.Flags().GetString("addr")
		staticDir, _ := cmd.Flags().GetString("static")
		outDir := viper.GetString("OUT_DIR")
		timezone := viper.GetString("TZ")
		logger.Info("Starting web server on http://%s", addr)
		return api.Run(addr, outDir, timezone, staticDir)
	},
}

func init() {
	webCmd.Flags().String("addr", "localhost:8080", "Address to listen on")
	webCmd.Flags().String("static", "./web/dist", "Path to static frontend files")
}
