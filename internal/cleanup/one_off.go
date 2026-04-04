package cleanup

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/pashkov256/deletor/internal/filemanager"
	"github.com/pashkov256/deletor/internal/rules"
	"github.com/pashkov256/deletor/internal/utils"
)

// OneOffCleanSpec is a snapshot of the saved cleanup rules used for a
// scheduled one-off clean run.
type OneOffCleanSpec struct {
	Path                  string
	Extensions            []string
	Exclude               []string
	MinSize               int64
	MaxSize               int64
	OlderThan             time.Time
	NewerThan             time.Time
	IncludeSubfolders     bool
	DeleteEmptySubfolders bool
	SendFilesToTrash      bool
	LogToFile             bool
}

// OneOffCleanResult captures the outcome of a scheduled clean execution.
type OneOffCleanResult struct {
	Path             string
	FilesCleaned     int
	BytesCleared     int64
	EmptyDirsDeleted int
	UsedTrash        bool
	CompletedAt      time.Time
}

// LoadOneOffCleanSpec snapshots the currently saved rules for a scheduled
// cleanup run.
func LoadOneOffCleanSpec(ruleManager rules.Rules) (*OneOffCleanSpec, error) {
	if ruleManager == nil {
		return nil, errors.New("rules manager is required")
	}

	savedRules, err := ruleManager.GetRules()
	if err != nil {
		return nil, err
	}

	targetPath := utils.ExpandTilde(savedRules.Path)
	if targetPath == "" {
		return nil, errors.New("save a target path in Manage Rules before scheduling a clean")
	}

	if _, err := os.Stat(targetPath); err != nil {
		return nil, fmt.Errorf("saved rules path is invalid: %w", err)
	}

	var minSize, maxSize int64
	var olderThan, newerThan time.Time

	if savedRules.MinSize != "" {
		minSize, err = utils.ToBytes(savedRules.MinSize)
		if err != nil {
			return nil, fmt.Errorf("invalid saved minimum size: %w", err)
		}
	}

	if savedRules.MaxSize != "" {
		maxSize, err = utils.ToBytes(savedRules.MaxSize)
		if err != nil {
			return nil, fmt.Errorf("invalid saved maximum size: %w", err)
		}
	}

	if savedRules.OlderThan != "" {
		olderThan, err = utils.ParseTimeDuration(savedRules.OlderThan)
		if err != nil {
			return nil, fmt.Errorf("invalid saved older-than value: %w", err)
		}
	}

	if savedRules.NewerThan != "" {
		newerThan, err = utils.ParseTimeDuration(savedRules.NewerThan)
		if err != nil {
			return nil, fmt.Errorf("invalid saved newer-than value: %w", err)
		}
	}

	return &OneOffCleanSpec{
		Path:                  targetPath,
		Extensions:            append([]string(nil), savedRules.Extensions...),
		Exclude:               append([]string(nil), savedRules.Exclude...),
		MinSize:               minSize,
		MaxSize:               maxSize,
		OlderThan:             olderThan,
		NewerThan:             newerThan,
		IncludeSubfolders:     savedRules.IncludeSubfolders,
		DeleteEmptySubfolders: savedRules.DeleteEmptySubfolders,
		SendFilesToTrash:      savedRules.SendFilesToTrash,
		LogToFile:             savedRules.LogToFile,
	}, nil
}

// RunOneOffClean executes a one-off cleanup run using a previously loaded
// cleanup spec.
func RunOneOffClean(fm filemanager.FileManager, spec *OneOffCleanSpec) (*OneOffCleanResult, error) {
	if fm == nil {
		return nil, errors.New("file manager is required")
	}
	if spec == nil {
		return nil, errors.New("cleanup spec is required")
	}

	filter := fm.NewFileFilter(
		spec.MinSize,
		spec.MaxSize,
		utils.ParseExtToMap(spec.Extensions),
		spec.Exclude,
		spec.OlderThan,
		spec.NewerThan,
	)

	scanner := filemanager.NewFileScanner(fm, filter, false)

	var toClean map[string]string
	var totalBytes int64

	if spec.IncludeSubfolders {
		toClean, totalBytes = scanner.ScanFilesRecursively(spec.Path)
	} else {
		toClean, totalBytes = scanner.ScanFilesCurrentLevel(spec.Path)
	}

	for filePath := range toClean {
		if spec.SendFilesToTrash {
			fm.MoveFileToTrash(filePath)
		} else {
			fm.DeleteFile(filePath)
		}
	}

	emptyDirsDeleted := 0
	if spec.DeleteEmptySubfolders {
		emptyDirs := scanner.ScanEmptySubFolders(spec.Path)
		emptyDirsDeleted = len(emptyDirs)
		if emptyDirsDeleted > 0 {
			fm.DeleteEmptySubfolders(spec.Path)
		}
	}

	if spec.LogToFile && len(toClean) > 0 {
		utils.LogDeletionToFile(toClean)
	}

	return &OneOffCleanResult{
		Path:             spec.Path,
		FilesCleaned:     len(toClean),
		BytesCleared:     totalBytes,
		EmptyDirsDeleted: emptyDirsDeleted,
		UsedTrash:        spec.SendFilesToTrash,
		CompletedAt:      time.Now(),
	}, nil
}
