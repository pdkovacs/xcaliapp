package local

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/rs/zerolog"
)

type CmdOpts struct {
	Cwd string
}

func (o CmdOpts) String() string {
	return fmt.Sprintf("{Cwd: %v}", o.Cwd)
}

type ExecCmdParams struct {
	Name string
	Args []string
	Opts *CmdOpts
}

func (e ExecCmdParams) String() string {
	option_string := "No options given"
	if e.Opts != nil {
		option_string = fmt.Sprintf("%v", *e.Opts)
	}
	return fmt.Sprintf("%v, %v, %v", e.Name, e.Args, option_string)
}

func ExecuteCommand(params ExecCmdParams, logger *zerolog.Logger) (string, error) {
	execCmdLogger := logger.With().Str("function", "ExecuteCommand").Logger()
	execCmdLogger.Info().Interface("params", params).Msg("Starting execution...")

	cmd := exec.Command(params.Name, params.Args...)
	if params.Opts != nil {
		cmd.Dir = params.Opts.Cwd
	}
	stderr, errStderr := cmd.StderrPipe()
	if errStderr != nil {
		return "", errStderr
	}
	stdout, errStdout := cmd.StdoutPipe()
	if errStdout != nil {
		return "", errStdout
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	slurpErr, _ := io.ReadAll(stderr)
	slurpOut, _ := io.ReadAll(stdout)

	err := cmd.Wait()
	if err != nil {
		errMsg := slurpErr
		if len(errMsg) == 0 {
			errMsg = slurpOut
		}
		return string(errMsg), err
	}
	return string(slurpOut), nil
}
