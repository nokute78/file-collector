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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func validateCommandOutput(t *testing.T, buf *bytes.Buffer, expect string) error {
	t.Helper()

	ret, err := buf.ReadString('\n')
	if err != nil {
		newerr := fmt.Errorf("ReadString error:%s", err)
		t.Errorf("%s", newerr)
		return newerr
	}
	if ret != expect {
		err = fmt.Errorf("mismatch:\n given= %s\n expect=%s", ret, expect)
		t.Errorf("%s", err)
		return err
	}
	return nil
}

func TestExecCommand(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})

	src := &SrcFile{BeforeCmd: []string{"echo"}, AfterCmd: []string{"echo"}}

	err := src.ExecBeforeCmd(buf, buf)
	if err != nil {
		t.Errorf("ExecBeforeCmd:%s", err)
	}
	err = validateCommandOutput(t, buf, "\n")
	if err != nil {
		t.Errorf("validateCommand(ExecBeforeCmd):%s", err)
	}
	buf.Reset()

	err = src.ExecAfterCmd(buf, buf)
	if err != nil {
		t.Errorf("ExecAfterCmd:%s", err)
	}
	err = validateCommandOutput(t, buf, "\n")
	if err != nil {
		t.Errorf("validateCommand(ExecAfterCmd):%s", err)
	}
	buf.Reset()

	src = &SrcFile{Path: "input", DstPath: "output",
		BeforeCmd: []string{"echo", "${target}"}, AfterCmd: []string{"echo", "${target}"}}

	err = src.ExecBeforeCmd(buf, buf)
	if err != nil {
		t.Errorf("ExecBeforeCmd:%s", err)
	}
	err = validateCommandOutput(t, buf, "input\n")
	if err != nil {
		t.Errorf("validateCommand(ExecBeforeCmd):%s", err)
	}
	buf.Reset()

	err = src.ExecAfterCmd(buf, buf)
	if err != nil {
		t.Errorf("ExecAfterCmd:%s", err)
	}
	err = validateCommandOutput(t, buf, "output\n")
	if err != nil {
		t.Errorf("validateCommand(ExecAfterCmd):%s", err)
	}
	buf.Reset()

	src.BeforeCmd = []string{"echo", "file", "is", "${target}"}
	src.AfterCmd = []string{"echo", "file", "is", "${target}"}

	err = src.ExecBeforeCmd(buf, buf)
	if err != nil {
		t.Errorf("ExecBeforeCmd 2:%s", err)
	}
	err = validateCommandOutput(t, buf, "file is input\n")
	if err != nil {
		t.Errorf("validateCommand(ExecBeforeCmd 2):%s", err)
	}
	buf.Reset()

	err = src.ExecAfterCmd(buf, buf)
	if err != nil {
		t.Errorf("ExecAfterCmd 2:%s", err)
	}
	err = validateCommandOutput(t, buf, "file is output\n")
	if err != nil {
		t.Errorf("validateCommand(ExecAfterCmd):%s", err)
	}
	buf.Reset()
}

func TestChecksumStr(t *testing.T) {
	type testcase struct {
		sumType string
		expect  string
	}
	cases := []testcase{
		{"sha256", "7d1a54127b222502f5b79b5fb0803061152a44f92b37e23c6527baf665d4da9a"},
		{"sha1", "2fb5e13419fc89246865e7a324f476ec624e8740"},
		{"md5", "7ac66c0f148de9519b8bd264312c4d64"},
	}

	f, err := ioutil.TempFile("", "testchecksumstr")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	rowstr := "abcdefg"
	n, err := f.WriteString(rowstr)
	if err != nil {
		t.Fatal(err)
	} else if n < len(rowstr) {
		t.Fatalf("len(rowstr):%d", n)
	}

	src := &SrcFile{Path: f.Name()}
	_, err = src.ChecksumStr(f.Name())
	if err == nil {
		t.Error("No error src.ChecksumStr()")
	}

	for _, v := range cases {
		src.ChecksumType = v.sumType
		sumstr, err := src.ChecksumStr(f.Name())
		if err != nil {
			t.Fatal(err)
		}

		if sumstr != v.expect {
			t.Errorf("%s error:\n given =%s\n expect=%s", v.sumType, sumstr, v.expect)
		}
	}
}

func TestDecodeJSON(t *testing.T) {
	src := &SrcFile{}
	err := json.Unmarshal([]byte{}, &src)
	if err == nil {
		t.Error("It should be error")
	}

	input := `{"path": "Path", "dst_path": "Dst_path"}`
	err = json.Unmarshal([]byte(input), &src)
	if err != nil {
		t.Errorf("simple case: %s\ninput=%s", err, input)
	}
	if src.Path == "" {
		t.Errorf("SrcPath is blank")
	}
	if src.DstPath == "" {
		t.Errorf("DstPath is blank %s", src)
	}

	input = `{"path": "Path", "dst_path": "Dst_path", "before_cmd": ["echo"], "after_cmd": ["echo"], "checksum": "md5"}`
	err = json.Unmarshal([]byte(input), &src)
	if err != nil {
		t.Errorf("full case: %s\ninput=%s", err, input)
	}
	if src.Path == "" {
		t.Errorf("SrcPath is blank")
	}
	if src.DstPath == "" {
		t.Errorf("DstPath is blank %s", src)
	}
}

func TestIsSubDir(t *testing.T) {
	type testcase struct {
		name   string
		root   string
		path   string
		expect bool
	}

	cases := []testcase{
		{"normal", "hoge/", "hoge/a", true},
		{"relpath", "hoge/", "../hoge/", false},
		{"relpath2", "", "a", true},
	}

	for _, v := range cases {
		ret := IsSubDir(v.root, v.path)
		if ret != v.expect {
			t.Errorf("%s:given =%t expect=%t", v.name, ret, v.expect)
		}
	}
}

func TestSrcCheckConfiguration(t *testing.T) {
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

	in := &SrcFile{}
	in.Path = tmpdir

	err = in.CheckConfiguration(tmpdir)
	if err == nil {
		t.Errorf("i.Path should be a directory")
	}

	in.Path = tmpfile.Name()
	err = in.CheckConfiguration(tmpfile.Name())
	if err == nil {
		t.Errorf("outroot should be a directory")
	}

	in.Path = "/tmp/hoge"
	err = in.CheckConfiguration(tmpdir)
	if err == nil {
		t.Errorf("i.Path should not be a absolute path")
	}
}

func createTxtFile(t *testing.T, path string) error {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Create %s", err)
	}
	defer f.Close()
	_, err = f.Write([]byte("test"))
	if err != nil {
		return fmt.Errorf("write  %s", err)
	}
	return nil
}

func sameFile(t *testing.T, src string, dst string) (bool, error) {
	t.Helper()
	s, err := ioutil.ReadFile(src)
	if err != nil {
		return false, err
	}
	d, err := ioutil.ReadFile(dst)
	if err != nil {
		return false, err
	}

	ret := bytes.Compare(s, d)
	if ret != 0 {
		return false, nil
	}
	return true, nil
}

func TestCopyFile(t *testing.T) {
	srcdir, err := ioutil.TempDir("", "checkconfiguration")
	if err != nil {
		t.Fatalf("TempDir(src):%s", err)
	}
	defer os.RemoveAll(srcdir)

	dstdir, err := ioutil.TempDir("", "checkconfiguration")
	if err != nil {
		t.Fatalf("TempDir(dst):%s", err)
	}
	defer os.RemoveAll(dstdir)

	type testcase struct {
		name string
		src  string
		dst  string
	}

	cases := []testcase{
		{"normal", "a.txt", "a.txt"},
		{"rename", "a.txt", "b.txt"},
		{"dst sub dir", "a.txt", "src/a.txt"},
	}

	s := &SrcFile{}

	for _, v := range cases {
		s.Path = filepath.Join(srcdir, v.src)
		s.DstPath = filepath.Join(dstdir, v.dst)
		err := createTxtFile(t, s.Path)
		if err != nil {
			t.Errorf("%s: %s", v.name, err)
			continue
		}

		err = s.CopyFile()
		if err != nil {
			t.Errorf("%s: error %s", v.name, err)
		}

		ok, err := sameFile(t, s.Path, s.DstPath)
		if err != nil {
			t.Errorf("%s: error %s", v.name, err)
		} else if !ok {
			t.Errorf("%s: file is not same", v.name)
		}
	}
}
