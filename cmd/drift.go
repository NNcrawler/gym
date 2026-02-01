package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"
)

var errDirMismatch = errors.New("directory mismatch")

func driftCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "drift",
		Short: "List drifting skills for the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve project root: %w", err)
			}
			globalCfg, err := loadGlobalConfig()
			if err != nil {
				return err
			}
			drifted, err := projectDriftSkills(projectRoot, globalCfg.SkillRepository)
			if err != nil {
				return fmt.Errorf("check drift for %s: %w", projectRoot, err)
			}
			if len(drifted) == 0 {
				fmt.Fprintln(os.Stdout, "No drifting skills found")
				return nil
			}
			sort.Slice(drifted, func(i, j int) bool {
				return drifted[i].Skill < drifted[j].Skill
			})
			for _, item := range drifted {
				fmt.Fprintf(
					os.Stdout,
					"%s: repo=%s project=%s status=%s\n",
					item.Skill,
					formatModTime(item.RepoTime),
					formatModTime(item.ProjectTime),
					item.Status,
				)
			}
			return nil
		},
	}
}

type driftInfo struct {
	Skill       string
	RepoTime    time.Time
	ProjectTime time.Time
	Status      string
}

func projectDriftSkills(projectRoot, skillRepo string) ([]driftInfo, error) {
	projectCfg, err := loadProjectConfig(projectRoot)
	if err != nil {
		return nil, err
	}
	if err := ensureSupportedAgents(projectCfg.Agents); err != nil {
		return nil, err
	}
	if len(projectCfg.SkillMap) == 0 {
		return nil, nil
	}
	drifted := make([]driftInfo, 0)
	for skillName, overrides := range projectCfg.SkillMap {
		skillSrc := filepath.Join(skillRepo, skillName)
		if _, err := os.Stat(skillSrc); err != nil {
			return nil, fmt.Errorf("skill %q not found in repository: %w", skillName, err)
		}
		repoTime, err := latestModTime(skillSrc)
		if err != nil {
			return nil, fmt.Errorf("read repository mtime for %s: %w", skillSrc, err)
		}
		projectTime := time.Time{}
		skillHasDrift := false
		for _, agent := range projectCfg.Agents {
			target, err := resolveSkillTarget(projectRoot, skillName, agent, overrides)
			if err != nil {
				return nil, err
			}
			targetTime, err := latestModTime(target)
			if err != nil {
				return nil, fmt.Errorf("read project mtime for %s: %w", target, err)
			}
			if targetTime.After(projectTime) {
				projectTime = targetTime
			}
			match, err := dirsEqual(skillSrc, target)
			if err != nil {
				return nil, err
			}
			if !match {
				skillHasDrift = true
			}
		}
		if skillHasDrift {
			status := driftStatus(repoTime, projectTime)
			drifted = append(drifted, driftInfo{
				Skill:       skillName,
				RepoTime:    repoTime,
				ProjectTime: projectTime,
				Status:      status,
			})
		}
	}
	return drifted, nil
}

func dirsEqual(src, dst string) (bool, error) {
	info, err := os.Stat(src)
	if err != nil {
		return false, err
	}
	if !info.IsDir() {
		return false, fmt.Errorf("source %s is not a directory", src)
	}
	dstInfo, err := os.Stat(dst)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if !dstInfo.IsDir() {
		return false, nil
	}

	if err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == src {
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		targetInfo, err := os.Lstat(target)
		if err != nil {
			if os.IsNotExist(err) {
				return errDirMismatch
			}
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		mode := info.Mode()
		if mode&os.ModeSymlink != 0 {
			if targetInfo.Mode()&os.ModeSymlink == 0 {
				return errDirMismatch
			}
			srcLink, err := os.Readlink(path)
			if err != nil {
				return err
			}
			dstLink, err := os.Readlink(target)
			if err != nil {
				return err
			}
			if srcLink != dstLink {
				return errDirMismatch
			}
			return nil
		}
		if d.IsDir() {
			if !targetInfo.IsDir() {
				return errDirMismatch
			}
			return nil
		}
		if !mode.IsRegular() || !targetInfo.Mode().IsRegular() {
			return errDirMismatch
		}
		if mode.Perm() != targetInfo.Mode().Perm() {
			return errDirMismatch
		}
		equal, err := filesEqual(path, target)
		if err != nil {
			return err
		}
		if !equal {
			return errDirMismatch
		}
		return nil
	}); err != nil {
		if errors.Is(err, errDirMismatch) {
			return false, nil
		}
		return false, err
	}

	if err := filepath.WalkDir(dst, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == dst {
			return nil
		}
		rel, err := filepath.Rel(dst, path)
		if err != nil {
			return err
		}
		source := filepath.Join(src, rel)
		if _, err := os.Lstat(source); err != nil {
			if os.IsNotExist(err) {
				return errDirMismatch
			}
			return err
		}
		return nil
	}); err != nil {
		if errors.Is(err, errDirMismatch) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func filesEqual(src, dst string) (bool, error) {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return false, err
	}
	dstInfo, err := os.Stat(dst)
	if err != nil {
		return false, err
	}
	if srcInfo.Size() != dstInfo.Size() {
		return false, nil
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return false, err
	}
	defer srcFile.Close()

	dstFile, err := os.Open(dst)
	if err != nil {
		return false, err
	}
	defer dstFile.Close()

	bufA := make([]byte, 32*1024)
	bufB := make([]byte, 32*1024)
	for {
		readA, errA := srcFile.Read(bufA)
		readB, errB := dstFile.Read(bufB)
		if readA != readB {
			return false, nil
		}
		if readA > 0 && !equalBytes(bufA[:readA], bufB[:readB]) {
			return false, nil
		}
		if errA != nil || errB != nil {
			if errA == io.EOF && errB == io.EOF {
				return true, nil
			}
			if errA == io.EOF || errB == io.EOF {
				return false, nil
			}
			if errA != nil {
				return false, errA
			}
			return false, errB
		}
	}
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func latestModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	if !info.IsDir() {
		return info.ModTime(), nil
	}
	latest := info.ModTime()
	if err := filepath.WalkDir(path, func(walkPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.ModTime().After(latest) {
			latest = info.ModTime()
		}
		return nil
	}); err != nil {
		return time.Time{}, err
	}
	return latest, nil
}

func formatModTime(value time.Time) string {
	if value.IsZero() {
		return "missing"
	}
	return value.UTC().Format(time.RFC3339)
}

func driftStatus(repoTime, projectTime time.Time) string {
	if projectTime.IsZero() && !repoTime.IsZero() {
		return "project missing"
	}
	if repoTime.After(projectTime) {
		return "repo newer"
	}
	if projectTime.After(repoTime) {
		return "project newer"
	}
	return "in sync"
}
