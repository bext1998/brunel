package tools

import (
	"io/fs"
	"path/filepath"
	"sort"

	"github.com/bext1998/brunel/internal/filetools"
)

func listFiles(r filetools.Resolver, path string, glob *string, maxDepth *int) ([]FileEntry, error) {
	root, err := r.Resolve(path)
	if err != nil {
		return nil, err
	}
	base := filepath.Clean(path)
	entries := make([]FileEntry, 0)
	err = filepath.WalkDir(root, func(current string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		rel, err := filepath.Rel(root, current)
		if err != nil {
			return err
		}
		depth := 0
		if rel != "." {
			depth = len(splitPath(rel))
		}
		if d.IsDir() {
			if maxDepth != nil && *maxDepth > 0 && depth >= *maxDepth {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		if maxDepth != nil && *maxDepth > 0 && depth > *maxDepth {
			return nil
		}
		if glob != nil {
			matched, err := filepath.Match(*glob, d.Name())
			if err != nil {
				return err
			}
			if !matched {
				return nil
			}
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		displayPath := base
		if rel != "." {
			displayPath = filepath.Join(base, rel)
		}
		entries = append(entries, FileEntry{Path: filepath.ToSlash(displayPath), Size: info.Size()})
		return nil
	})
	if err != nil {
		return nil, codeError(ErrToolIO.Code, "cannot list workspace files", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return entries, nil
}

func splitPath(path string) []string {
	if path == "." || path == "" {
		return nil
	}
	return splitPathComponents(filepath.Clean(path))
}

func splitPathComponents(path string) []string {
	var parts []string
	for {
		dir, file := filepath.Dir(path), filepath.Base(path)
		parts = append(parts, file)
		if dir == "." || dir == path {
			break
		}
		path = dir
	}
	return parts
}
