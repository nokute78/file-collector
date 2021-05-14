/*
   Copyright 2020 Takahiro Yamashita

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

const version string = "0.0.3"

// Exit status
const (
	ExitOK int = iota
	ExitArgError
	ExitCmdError
)

// CLI has In/Out/Err streams.
type CLI struct {
	OutStream io.Writer
	InStream  io.Reader
	ErrStream io.Writer
	quiet     bool // for testing to suppress output
}

// Run executes real main function.
func (cli *CLI) Run(args []string) (ret int) {
	cnf, err := Configure(args[1:], cli.quiet)
	if err != nil {
		if err == flag.ErrHelp {
			return ExitOK
		}
		fmt.Fprintf(cli.ErrStream, "%s\n", err)
		return ExitArgError
	}

	if cnf.showVersion {
		fmt.Fprintf(cli.OutStream, "Ver: %s\n", version)
		return ExitOK
	}
	if cnf.ConfigFilePath == "" {
		fmt.Fprintf(cli.ErrStream, "config file is missing")
		return ExitArgError
	}

	b, err := ioutil.ReadFile(cnf.ConfigFilePath)
	if err != nil {
		fmt.Fprintf(cli.ErrStream, "ReadFile:%s", err)
		return ExitCmdError
	}
	job := &Job{}
	err = json.Unmarshal(b, &job)
	if err != nil {
		fmt.Fprintf(cli.ErrStream, "Unmarshal:%s\n", err)
		synerr, ok := err.(*json.SyntaxError)
		if ok {
			fmt.Fprintf(cli.ErrStream, "  %s\n", string(b[synerr.Offset:]))
		}
		return ExitCmdError
	}

	err = job.CopyAndExec(cli.OutStream, cli.ErrStream)
	if err != nil {
		fmt.Fprintf(cli.ErrStream, "%s\n", err)
		return ExitCmdError
	}

	return ExitOK
}

func main() {
	cli := &CLI{OutStream: os.Stdout, InStream: os.Stdin, ErrStream: os.Stderr}

	os.Exit(cli.Run(os.Args))
}
