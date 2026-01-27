package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a project for skill management",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve project root: %w", err)
			}
			exists, err := projectConfigExists(projectRoot)
			if err != nil {
				return err
			}
			if exists {
				return errors.New(".skills.yaml already exists")
			}

			agents, err := promptAgents(os.Stdin, os.Stdout)
			if err != nil {
				return err
			}
			if err := ensureSupportedAgents(agents); err != nil {
				return err
			}

			cfg := ProjectConfig{
				Agents:   agents,
				SkillMap: map[string]map[string]string{},
			}
			if err := writeProjectConfig(projectRoot, cfg); err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, "Created .skills.yaml")
			return nil
		},
	}
}

func addCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <skill-name>",
		Short: "Add a skill from the central repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillName := args[0]
			projectRoot, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve project root: %w", err)
			}
			globalCfg, err := loadGlobalConfig()
			if err != nil {
				return err
			}
			projectCfg, err := loadProjectConfig(projectRoot)
			if err != nil {
				return err
			}
			if err := ensureSupportedAgents(projectCfg.Agents); err != nil {
				return err
			}

			skillSrc := filepath.Join(globalCfg.SkillRepository, skillName)
			if _, err := os.Stat(skillSrc); err != nil {
				return fmt.Errorf("skill %q not found in repository: %w", skillName, err)
			}

			if projectCfg.SkillMap == nil {
				projectCfg.SkillMap = map[string]map[string]string{}
			}
			if _, ok := projectCfg.SkillMap[skillName]; !ok {
				projectCfg.SkillMap[skillName] = map[string]string{}
			}

			for _, agent := range projectCfg.Agents {
				target, err := resolveSkillTarget(projectRoot, skillName, agent, projectCfg.SkillMap[skillName])
				if err != nil {
					return err
				}
				if err := copySkillDir(skillSrc, target); err != nil {
					return fmt.Errorf("copy skill to %s: %w", target, err)
				}
				fmt.Fprintf(os.Stdout, "Synced %s for %s -> %s\n", skillName, agent, target)
			}

			if err := writeProjectConfig(projectRoot, projectCfg); err != nil {
				return err
			}
			return nil
		},
	}
}

func syncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Synchronize all registered skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve project root: %w", err)
			}
			globalCfg, err := loadGlobalConfig()
			if err != nil {
				return err
			}
			projectCfg, err := loadProjectConfig(projectRoot)
			if err != nil {
				return err
			}
			if err := ensureSupportedAgents(projectCfg.Agents); err != nil {
				return err
			}
			if len(projectCfg.SkillMap) == 0 {
				fmt.Fprintln(os.Stdout, "No skills registered in .skills.yaml")
				return nil
			}

			for skillName, overrides := range projectCfg.SkillMap {
				skillSrc := filepath.Join(globalCfg.SkillRepository, skillName)
				if _, err := os.Stat(skillSrc); err != nil {
					return fmt.Errorf("skill %q not found in repository: %w", skillName, err)
				}
				for _, agent := range projectCfg.Agents {
					target, err := resolveSkillTarget(projectRoot, skillName, agent, overrides)
					if err != nil {
						return err
					}
					if err := copySkillDir(skillSrc, target); err != nil {
						return fmt.Errorf("copy skill to %s: %w", target, err)
					}
					fmt.Fprintf(os.Stdout, "Synced %s for %s -> %s\n", skillName, agent, target)
				}
			}
			return nil
		},
	}
}
