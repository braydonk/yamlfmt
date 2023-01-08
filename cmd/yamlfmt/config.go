package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/braydonk/yaml"
	"github.com/google/yamlfmt"
	"github.com/google/yamlfmt/command"
	"github.com/mitchellh/mapstructure"
)

const (
	configFileName string = ".yamlfmt"
	configHomeDir  string = "yamlfmt"
)

var (
	errNoConfFlag       = errors.New("config path not specified in --conf")
	errConfPathInvalid  = errors.New("config path specified in --conf was invalid")
	errConfPathNotExist = errors.New("config path does not exist")
	errConfPathIsDir    = errors.New("config path is dir")
	errNoConfigHome     = errors.New("missing required env var for config home")
)

type configPathError struct {
	path string
	err  error
}

func (e *configPathError) Error() string {
	if errors.Is(e.err, errConfPathInvalid) {
		return fmt.Sprintf("Config path %s was invalid", e.path)
	}
	if errors.Is(e.err, errConfPathNotExist) {
		return fmt.Sprintf("Config path %s does not exist", e.path)
	}
	if errors.Is(e.err, errConfPathIsDir) {
		return fmt.Sprintf("Config path %s is a directory", e.path)
	}
	return e.err.Error()
}

func (e *configPathError) Unwrap() error {
	return e.err
}

func readConfig(path string) (map[string]any, error) {
	yamlBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var configData map[string]interface{}
	err = yaml.Unmarshal(yamlBytes, &configData)
	if err != nil {
		return nil, err
	}
	return configData, nil
}

func getConfigPath() (string, error) {
	// First priority: specified in cli flag
	configPath, err := getConfigPathFromFlag()
	if err != nil {
		// If they don't provide a conf flag, we continue. If
		// a conf flag is provided and it's wrong, we consider
		// that a failure state.
		if !errors.Is(err, errNoConfFlag) {
			return "", err
		}
	} else {
		return configPath, nil
	}

	// Second priority: in the working directory
	configPath, err = getConfigPathFromWd()
	// In this scenario, no error constitutes a failure state,
	// so we continue to the next fallback.
	if err == nil {
		return configPath, nil
	}

	// Third priority: in home config directory
	configPath, err = getConfigPathFromConfigHome()
	// In this scenario, no error constitutes a failure state,
	// so we continue to the next fallback.
	if err == nil {
		return configPath, nil
	}

	// All else fails, no path and no error (signals to
	// use default config).
	return "", nil
}

func getConfigPathFromFlag() (string, error) {
	configPath := *flagConf
	if configPath == "" {
		return configPath, errNoConfFlag
	}
	return configPath, validatePath(configPath)
}

func getConfigPathFromWd() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	configPath := filepath.Join(wd, configFileName)
	return configPath, validatePath(configPath)
}

func getConfigPathFromConfigHome() (string, error) {
	// Build tags are a veritable pain in the behind,
	// I'm putting both config home functions in this
	// file. You can't stop me.
	if runtime.GOOS == "windows" {
		return getConfigPathFromAppDataLocal()
	}
	return getConfigPathFromXdgConfigHome()
}

func getConfigPathFromXdgConfigHome() (string, error) {
	configHome, configHomePresent := os.LookupEnv("XDG_CONFIG_HOME")
	if !configHomePresent {
		home, homePresent := os.LookupEnv("HOME")
		if !homePresent {
			// I fear whom's'tever does not have a $HOME set
			return "", errNoConfigHome
		}
		configHome = filepath.Join(home, ".config")
	}
	homeConfigPath := filepath.Join(configHome, configHomeDir, configFileName)
	return homeConfigPath, validatePath(homeConfigPath)
}

func getConfigPathFromAppDataLocal() (string, error) {
	configHome, configHomePresent := os.LookupEnv("LOCALAPPDATA")
	if !configHomePresent {
		// I think you'd have to go out of your way to unset this,
		// so this should only happen to sickos with broken setups.
		return "", errNoConfigHome
	}
	homeConfigPath := filepath.Join(configHome, configHomeDir, configFileName)
	return homeConfigPath, validatePath(homeConfigPath)
}

func validatePath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &configPathError{
				path: path,
				err:  errConfPathNotExist,
			}
		}
		if info.IsDir() {
			return &configPathError{
				path: path,
				err:  errConfPathIsDir,
			}
		}
		return &configPathError{
			path: path,
			err:  err,
		}
	}
	return nil
}

func makeCommandConfigFromData(configData map[string]any) (*command.Config, error) {
	var config *command.Config
	err := mapstructure.Decode(configData, &config)
	if err != nil {
		return nil, err
	}

	// Default to OS line endings
	if config.LineEnding == "" {
		config.LineEnding = yamlfmt.LineBreakStyleLF
		if runtime.GOOS == "windows" {
			config.LineEnding = yamlfmt.LineBreakStyleCRLF
		}
	}

	// Default to yaml and yml extensions
	if len(config.Extensions) == 0 {
		config.Extensions = []string{"yaml", "yml"}
	}

	// Default to flag if not set in config
	if !config.Doublestar {
		config.Doublestar = useDoublestar()
	}

	// Overwrite config if includes are provided through args
	if len(flag.Args()) > 0 {
		config.Include = flag.Args()
	}

	return config, nil
}
