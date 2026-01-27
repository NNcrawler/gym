package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var supportedAgents = map[string]string{
	"kilo-code": ".kilocode/skills",
	"codex":     ".codex/skills",
}

func listSupportedAgents() []string {
	agents := make([]string, 0, len(supportedAgents))
	for agent := range supportedAgents {
		agents = append(agents, agent)
	}
	sort.Strings(agents)
	return agents
}

func promptAgents(r io.Reader, w io.Writer) ([]string, error) {
	agents := listSupportedAgents()
	fmt.Fprintln(w, "Select agents used in this project:")
	for i, agent := range agents {
		fmt.Fprintf(w, "  %d) %s\n", i+1, agent)
	}
	fmt.Fprint(w, "Enter comma-separated numbers: ")

	reader := bufio.NewReader(r)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("read selection: %w", err)
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, errors.New("no agents selected")
	}

	parts := strings.Split(line, ",")
	selected := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		idx, err := parsePositiveInt(part)
		if err != nil {
			return nil, fmt.Errorf("invalid selection %q: %w", part, err)
		}
		if idx < 1 || idx > len(agents) {
			return nil, fmt.Errorf("selection %d out of range", idx)
		}
		agent := agents[idx-1]
		if !seen[agent] {
			selected = append(selected, agent)
			seen[agent] = true
		}
	}
	if len(selected) == 0 {
		return nil, errors.New("no agents selected")
	}
	return selected, nil
}

func promptSkillRepository(r io.Reader, w io.Writer) (string, error) {
	fmt.Fprint(w, "Enter the central skill repository path: ")
	reader := bufio.NewReader(r)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read repository path: %w", err)
	}
	repo := strings.TrimSpace(line)
	if repo == "" {
		return "", errors.New("skill repository path is empty")
	}
	info, err := os.Stat(repo)
	if err != nil {
		return "", fmt.Errorf("stat skill repository: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("skill repository %s is not a directory", repo)
	}
	return repo, nil
}

func parsePositiveInt(value string) (int, error) {
	var n int
	_, err := fmt.Sscanf(value, "%d", &n)
	if err != nil {
		return 0, err
	}
	if n <= 0 {
		return 0, errors.New("must be positive")
	}
	return n, nil
}

func ensureSupportedAgents(agents []string) error {
	for _, agent := range agents {
		if _, ok := supportedAgents[agent]; !ok {
			return fmt.Errorf("unsupported agent %q", agent)
		}
	}
	return nil
}

func defaultSkillDir(agent string) (string, error) {
	dir, ok := supportedAgents[agent]
	if !ok {
		return "", fmt.Errorf("unsupported agent %q", agent)
	}
	return dir, nil
}

func copySkillDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat source %s: %w", src, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("source %s is not a directory", src)
	}

	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("remove existing %s: %w", dst, err)
	}
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return fmt.Errorf("create destination %s: %w", dst, err)
	}

	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
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
		info, err := d.Info()
		if err != nil {
			return err
		}

		if d.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, target)
		}

		if err := copyFile(path, target, info.Mode()); err != nil {
			return err
		}
		return nil
	})
}

func copyFile(src, dst string, mode fs.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return nil
}

func resolveSkillTarget(projectRoot, skillName, agent string, overrides map[string]string) (string, error) {
	if overrides != nil {
		if override, ok := overrides[agent]; ok && override != "" {
			return filepath.Join(projectRoot, override), nil
		}
	}
	baseDir, err := defaultSkillDir(agent)
	if err != nil {
		return "", err
	}
	return filepath.Join(projectRoot, baseDir, skillName), nil
}
