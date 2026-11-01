package checkpoint

import (
	"os"
	"time"

	"github.com/Benjamin-van-Heerden/mem/internal/project"
	"github.com/Benjamin-van-Heerden/mem/internal/structure"
	"github.com/Benjamin-van-Heerden/mem/internal/work"
)

// folderBaseline is when work in a folder project last reached a checkpoint: the newest log, or the moment the project
// was set up.
func folderBaseline(p project.Project) (time.Time, error) {
	l, ok, err := work.LatestLog(p, "")
	if err != nil {
		return time.Time{}, err
	}
	if ok {
		if created, err := time.Parse(time.RFC3339, l.Meta.Created); err == nil {
			return created, nil
		}
	}
	info, err := os.Stat(project.ConfigPath(p.Root))
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

// FolderCover lists the files a work log written now covers in a folder project, most recently changed first.
func FolderCover(p project.Project) ([]structure.Change, error) {
	since, err := folderBaseline(p)
	if err != nil {
		return nil, err
	}
	return structure.ChangedSince(p, since)
}

// FolderSince measures the work in a folder project since the last log. With no commits or lines to go on, the changed
// files set the level on their own.
func FolderSince(p project.Project) (Work, error) {
	files, err := FolderCover(p)
	if err != nil {
		return Work{}, err
	}
	return Work{Files: len(files), Folder: true}, nil
}
