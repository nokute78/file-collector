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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Job struct {
	Srcs     []*SrcFile `json:"srcs"`
	DstDir   string     `json:"dst"`
	AfterCmd []string   `json:"after_cmd",omitempty`
}

func (j Job) CheckConfiguration() error {
	if len(j.Srcs) == 0 {
		return fmt.Errorf("Srcs missing")
	}
	/*
		outrootinfo, err := os.Stat(j.DstDir)
		if err != nil {
			return err
		}
		if !outrootinfo.IsDir() {
			return fmt.Errorf("dstRoot is a file")
		}
	*/
	return nil
}

func (j Job) CopyAndExec(cmdout io.Writer, cmderr io.Writer) error {
	err := j.CheckConfiguration()
	if err != nil {
		return err
	}

	tmpdir, err := ioutil.TempDir("", "job")
	if err != nil {
		return fmt.Errorf("Job.CopyAndExec Tempdir:%w", err)
	}
	defer os.RemoveAll(tmpdir)
	tmproot := filepath.Join(tmpdir, "root")
	err = os.Mkdir(tmproot, 0744)
	if err != nil {
		return fmt.Errorf("Job.CopyAndExec Mkdir:%w", err)
	}

	for _, v := range j.Srcs {
		err = v.CopyAndExec(tmproot)
		if err != nil {
			return fmt.Errorf("%s error:%s", v.Path, err)
		}
	}

	if len(j.AfterCmd) > 1 {
		mp := make(map[string]string)
		err = execCommand(mp, j.AfterCmd, cmdout, cmderr)
		if err != nil {
			return err
		}
	}

	err = os.Rename(tmproot, j.DstDir)
	if err != nil {
		return fmt.Errorf("Rename:%w", err)
	}

	return nil
}
