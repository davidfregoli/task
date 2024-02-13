package taskfile

import (
	"context"
	"os"
	"strings"

	"github.com/go-task/task/v3/errors"
	"github.com/go-task/task/v3/internal/experiments"
	"github.com/go-task/task/v3/internal/logger"
	"github.com/go-task/task/v3/taskfile/ast"
)

type Node interface {
	Read(ctx context.Context) ([]byte, error)
	Parent() Node
	Location() string
	Dir() string
	Optional() bool
	Remote() bool
	ResolveIncludeEntrypoint(include ast.Include) (string, error)
	ResolveIncludeDir(include ast.Include) (string, error)
}

func NewRootNode(
	l *logger.Logger,
	entrypoint string,
	dir string,
	insecure bool,
) (Node, error) {
	// Check if there is something to read on STDIN
	stat, _ := os.Stdin.Stat()
	if (stat.Mode()&os.ModeCharDevice) == 0 && stat.Size() > 0 {
		return NewStdinNode(dir)
	}
	return NewNode(l, entrypoint, dir, insecure)
}

func NewNode(
	l *logger.Logger,
	entrypoint string,
	dir string,
	insecure bool,
	opts ...NodeOption,
) (Node, error) {
	var node Node
	var err error
	switch getScheme(entrypoint) {
	case "http", "https":
		node, err = NewHTTPNode(l, entrypoint, dir, insecure, opts...)
	default:
		// If no other scheme matches, we assume it's a file
		node, err = NewFileNode(l, entrypoint, dir, opts...)
	}
	if node.Remote() && !experiments.RemoteTaskfiles.Enabled {
		return nil, errors.New("task: Remote taskfiles are not enabled. You can read more about this experiment and how to enable it at https://taskfile.dev/experiments/remote-taskfiles")
	}
	return node, err
}

func getScheme(uri string) string {
	if i := strings.Index(uri, "://"); i != -1 {
		return uri[:i]
	}
	return ""
}
