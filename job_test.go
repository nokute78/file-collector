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
	"io/ioutil"
	"os"
	"testing"
)

func TestDecodeJson(t *testing.T) {
	job := &Job{}
	err := json.Unmarshal([]byte{}, &job)
	if err == nil {
		t.Error("It should be error")
	}

	tmpdir, err := ioutil.TempDir("", "checkconfiguration")
	if err != nil {
		t.Fatalf("TempDir:%s", err)
	}
	defer os.RemoveAll(tmpdir)

	tmpfile, err := ioutil.TempFile("", "checkconfiguration")
	if err != nil {
		t.Fatalf("TempFile:%s", err)
	}
	defer os.Remove(tmpfile.Name())

	input := `{"srcs":[{"path":"` + tmpfile.Name() + `", "dst_path": "hoge"}],"dst":"dst"}`
	err = json.Unmarshal([]byte(input), &job)
	if err != nil {
		t.Errorf("normal input:%s", err)
	}
	if len(job.Srcs) == 0 {
		t.Errorf("Srcs is missing")
	}
	if job.DstDir == "" {
		t.Errorf("Dst is missing")
	}
}

func TestJobCheckConfiguration(t *testing.T) {
	j := &Job{}

	err := j.CheckConfiguration()
	if err == nil {
		t.Errorf("j is a blank. It should be error.")
	}

	tmpdir, err := ioutil.TempDir("", "dstdir")
	if err != nil {
		t.Fatalf("tempdir:%s", err)
	}
	defer os.RemoveAll(tmpdir)

	j.DstDir = tmpdir
	err = j.CheckConfiguration()
	if err == nil {
		t.Errorf("Srcs are blank. It should be error")
	}

	j.Srcs = append(j.Srcs, &SrcFile{})
	err = j.CheckConfiguration()
	if err != nil {
		t.Errorf("CheckConfiguration:%s", err)
	}
}
