package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const projectConfigName = ".skills.yaml"
const globalConfigName = ".gym.yaml"

type GlobalConfig struct {
	SkillRepository string `yaml:"skillRepository"`
}

type ProjectConfig struct {
	Agents   []string                     `yaml:"agents"`
	SkillMap map[string]map[string]string `yaml:"skillMap"`
}

func loadGlobalConfig() (GlobalConfig, error) {
	path, err := globalConfigPath()
	if err != nil {
		return GlobalConfig{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return GlobalConfig{}, fmt.Errorf("read global config %s: %w", path, err)
	}
	var cfg GlobalConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return GlobalConfig{}, fmt.Errorf("parse global config %s: %w", path, err)
	}
	if cfg.SkillRepository == "" {
		return GlobalConfig{}, errors.New("global config skillRepository is empty")
	}
	return cfg, nil
}

func writeGlobalConfig(cfg GlobalConfig) error {
	path, err := globalConfigPath()
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal global config: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func loadProjectConfig(projectRoot string) (ProjectConfig, error) {
	path := filepath.Join(projectRoot, projectConfigName)
	data, err := os.ReadFile(path)
	if err != nil {
		return ProjectConfig{}, fmt.Errorf("read project config %s: %w", path, err)
	}
	var cfg ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return ProjectConfig{}, fmt.Errorf("parse project config %s: %w", path, err)
	}
	if len(cfg.Agents) == 0 {
		return ProjectConfig{}, errors.New("project config agents list is empty")
	}
	return cfg, nil
}

func writeProjectConfig(projectRoot string, cfg ProjectConfig) error {
	path := filepath.Join(projectRoot, projectConfigName)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal project config: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func projectConfigExists(projectRoot string) (bool, error) {
	path := filepath.Join(projectRoot, projectConfigName)
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func globalConfigExists() (bool, error) {
	path, err := globalConfigPath()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func globalConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, globalConfigName), nil
}
