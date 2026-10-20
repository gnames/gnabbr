/*
Copyright © 2026 Dmitry Mozzherin <dmozzherin@gmail.com>
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/gnames/gn"
	"github.com/gnames/gnabbr/internal/iofs"
	"github.com/gnames/gnabbr/internal/iologger"
	"github.com/gnames/gnabbr/pkg/config"
	"github.com/gnames/gnabbr/pkg/gnabbr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	homeDir string
	opts    []config.Option
	cfg     *config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Version: fmt.Sprintf("version: %s\nbuild:   %s", gnabbr.Version, gnabbr.Build),
	Use:     "gnabbr",
	Short:   "Resolves abbreviated genera of scientific names",
	Long: `gnabbr resolves abbreviated genera of scientific names found in
documents (C. elegans -> Caenorhabditis elegans), using the names around
them and gnames abbr_names.`,
	PersistentPreRunE: bootstrap,
	RunE:              runRoot,
	SilenceErrors:     true,
	SilenceUsage:      true,
}

func bootstrap(cmd *cobra.Command, args []string) error {
	var err error

	homeDir, err = os.UserHomeDir()
	if err != nil {
		gn.PrintErrorMessage(err)
		return err
	}

	if err = iofs.EnsureDirs(homeDir); err != nil {
		gn.PrintErrorMessage(err)
		return err
	}

	// Initialize logging with hardcoded defaults ASAP so all subsequent logs are captured
	// Will be reconfigured later with user's config settings
	defaultLog := config.LogConfig{
		Format:      "json",
		Level:       "info",
		Destination: "file",
	}

	if err = iologger.Init(config.LogDir(homeDir), defaultLog, false); err != nil {
		gn.PrintErrorMessage(err)
		return err
	}

	// Now that logging is initialized, all subsequent logs will be captured
	slog.Info("Bootstrap process started")
	slog.Info("User home directory resolved", "home_dir", homeDir)
	slog.Info("Required directories ensured",
		"config_dir", config.ConfigDir(homeDir),
		"log_dir", config.LogDir(homeDir),
		"cache_dir", config.CacheDir(homeDir))
	slog.Info("Logger initialized with default configuration",
		"format", defaultLog.Format,
		"level", defaultLog.Level,
		"destination", defaultLog.Destination)

	if err = iofs.EnsureConfigFile(homeDir); err != nil {
		slog.Error("Failed to ensure config file", "error", err)
		gn.PrintErrorMessage(err)
		return err
	}
	slog.Info("Config file ensured", "config_file", config.ConfigFilePath(homeDir))

	var cfgViper *config.Config
	if cfgViper, err = initConfig(homeDir); err != nil {
		slog.Error("Failed to initialize configuration", "error", err)
		gn.PrintErrorMessage(err)
		return err
	}
	slog.Info("Configuration initialized from config file")

	cfg = config.New()
	slog.Info("Default configuration created")

	opts = cfgViper.ToOptions()
	slog.Info("Configuration options extracted from config file", "options_count", len(opts))

	cfg.Update(opts)
	slog.Info("Configuration updated with config file options")

	// Set HomeDir after config is loaded
	cfg.Update([]config.Option{config.OptHomeDir(homeDir)})
	slog.Info("Home directory set in configuration", "home_dir", homeDir)

	// Reconfigure logging with user's settings and proper log file location
	if err = reconfigureLogging(cfg); err != nil {
		slog.Error("Failed to reconfigure logging", "error", err)
		gn.PrintErrorMessage(err)
		return err
	}

	slog.Info("Configuration loaded successfully",
		"config_file", config.ConfigFilePath(homeDir),
		"log_format", cfg.Log.Format,
		"log_level", cfg.Log.Level,
		"log_destination", cfg.Log.Destination,
		"jobs_number", cfg.JobsNumber)

	return nil
}

// reconfigureLogging reinitializes the logger with the loaded configuration.
// Creates log file in the proper location now that we know HomeDir.
// Appends to existing log file to preserve bootstrap logs.
func reconfigureLogging(cfg *config.Config) error {
	logDir := config.LogDir(cfg.HomeDir)
	slog.Info("Reconfiguring logger with user settings",
		"log_dir", logDir,
		"format", cfg.Log.Format,
		"level", cfg.Log.Level,
		"destination", cfg.Log.Destination)

	err := iologger.Init(logDir, cfg.Log, true)
	if err != nil {
		slog.Error("Failed to reconfigure logger", "error", err, "log_dir", logDir)
		return err
	}

	slog.Info("Logger reconfigured successfully", "log_file", logDir+"/gnabbr.log")
	return nil
}

func runRoot(cmd *cobra.Command, args []string) error {
	// When root is called with no subcommands, show help
	return cmd.Help()
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Remove the automatic "gnabbr version" prefix
	rootCmd.SetVersionTemplate("{{.Version}}\n")

	// Override version flag to use -V (consistent with gnames projects)
	rootCmd.Flags().BoolP("version", "V", false, "version for gnabbr")

	// Add subcommands
	rootCmd.AddCommand(getResolveCmd())
}

func initConfig(home string) (*config.Config, error) {
	var err error
	cfgPath := config.ConfigFilePath(home)
	slog.Info("Initializing configuration from file", "config_path", cfgPath)

	v := viper.New()
	v.SetConfigFile(cfgPath)
	slog.Info("Viper initialized with config file path")

	initEnvVars(v)

	if err = v.ReadInConfig(); err != nil {
		slog.Error("Failed to read config file", "error", err, "config_path", cfgPath)
		return nil, iofs.ReadFileError(cfgPath, err)
	}
	slog.Info("Config file read successfully", "config_path", cfgPath)

	var res config.Config
	if err = v.Unmarshal(&res); err != nil {
		slog.Error("Failed to unmarshal config", "error", err, "config_path", cfgPath)
		return nil, iofs.ReadFileError(cfgPath, err)
	}
	slog.Info("Configuration unmarshaled successfully",
		"log_level", res.Log.Level,
		"log_format", res.Log.Format,
		"jobs_number", res.JobsNumber)

	return &res, nil
}

func initEnvVars(v *viper.Viper) {
	// Set environment variables we want.
	// We set them manually so we can see clearly which env variables are allowed.
	// These match the fields included in config.ToOptions() - i.e., persistent
	// configuration that can be stored in config.yaml.
	slog.Info("Binding environment variables to configuration")

	v.SetEnvPrefix("GNABBR")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	slog.Info("Environment variable prefix and key replacer set", "prefix", "GNABBR")

	// Log configuration
	v.BindEnv("log.level")
	v.BindEnv("log.format")
	v.BindEnv("log.destination")
	slog.Info("Log environment variables bound")

	// General configuration
	v.BindEnv("jobs_number")
	slog.Info("General environment variables bound")

	// Resolver configuration
	v.BindEnv("gnames_url")
	v.BindEnv("gap_chars")
	v.BindEnv("span_chars")
	v.BindEnv("hint_window")
	v.BindEnv("hints_limit")
	v.BindEnv("genera_limit")
	v.BindEnv("threshold")
	slog.Info("Resolver environment variables bound")

	v.AutomaticEnv()
	slog.Info("Environment variable binding complete, automatic env lookup enabled")
}
