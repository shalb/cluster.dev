package cdev

import (
	"runtime"

	"github.com/shalb/cluster.dev/internal/config"
	"github.com/shalb/cluster.dev/internal/project"
)

type CmdErrExtended struct {
	CdevUsage  *config.CdevUsage
	Command    string
	Err        error
	ProjectPtr *project.Project
}

func (e *CmdErrExtended) Error() string {
	if e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func NewCmdErr(p *project.Project, cmdName string, err error) *CmdErrExtended {
	return &CmdErrExtended{
		Err:        err,
		Command:    cmdName,
		ProjectPtr: p,
		CdevUsage: &config.CdevUsage{
			CdevVersion: config.Version,
			OsVersion:   runtime.GOOS,
		},
	}
}
