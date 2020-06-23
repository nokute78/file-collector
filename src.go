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
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var sumList sync.Map

func init() {
	sumList.Store("sha256", sha256.New())
	sumList.Store("sha1", sha1.New())
	sumList.Store("md5", md5.New())
}

type SrcFile struct {
	Path         string   `json:"path"`
	DstPath      string   `json:"dst_path"` // relative file path
	ChecksumType string   `json:"checksum,omitempty"`
	BeforeCmd    []string `json:"before_cmd",omitempty`
	AfterCmd     []string `json:"after_cmd",omitempty`
}

func (i SrcFile) String() string {
	return fmt.Sprintf("Path:%s, DstPath: %s, CheckSumType: %s", i.Path, i.DstPath, i.ChecksumType)
}

// IsSubDir checks if path is a sub directory of root.
func IsSubDir(root string, path string) bool {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	//	fmt.Printf("absPath:%s\nabsRoot:%s\n", absPath, absRoot)
	return strings.HasPrefix(absPath, absRoot)
}

func copyFile(dstPath string, srcPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("src open:%w", err)
	}
	defer src.Close()
	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("dst create:%w", err)
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}

func (i SrcFile) ExecBeforeCmd(out io.Writer, err io.Writer) error {
	return execCommand(i.Path, i.BeforeCmd, out, err)
}

func (i SrcFile) ExecAfterCmd(out io.Writer, err io.Writer) error {
	return execCommand(i.DstPath, i.AfterCmd, out, err)
}

func (i SrcFile) Checksum(path string) ([]byte, error) {
	l, ok := sumList.Load(i.ChecksumType)
	if !ok {
		return nil, fmt.Errorf("Unknown checksum :%s", i.ChecksumType)
	}
	h, ok := l.(hash.Hash)
	if !ok {
		return nil, fmt.Errorf("Not hash.Hash. %v", h)
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	n, err := h.Write(b)
	if err != nil {
		return nil, err
	} else if n < len(b) {
		return nil, fmt.Errorf("calc sum size error n=%d", n)
	}

	return h.Sum(nil), nil
}

func (i SrcFile) ChecksumStr(path string) (string, error) {
	b, err := i.Checksum(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}

// CheckConditions checks configuration
//   src should be a file.
//   dst root should be a directory.
func (i *SrcFile) CheckConfiguration(outRoot string) error {
	srcinfo, err := os.Stat(i.Path)
	if err != nil {
		return err
	}
	if srcinfo.IsDir() {
		return fmt.Errorf("SrcPath is a directory")
	}

	outrootinfo, err := os.Stat(outRoot)
	if err != nil {
		return fmt.Errorf("stat(outroot):%w", err)
	}
	if !outrootinfo.IsDir() {
		return fmt.Errorf("dstRoot is a file")
	}

	if !IsSubDir(outRoot, i.DstPath) {
		return fmt.Errorf("DstPath:%s is outside of root %s", i.DstPath, outRoot)
	}
	return nil
}

func (i *SrcFile) Normalize(outRoot string) error {
	if len(i.DstPath) > 1 && i.DstPath[0] == '/' {
		return fmt.Errorf("DstPath:%s should not be absolute path", i.DstPath)
	}

	outputPath := filepath.Join(outRoot, i.DstPath)
	if len(i.DstPath) == 0 {
		outputPath = filepath.Join(outputPath, filepath.Base(i.Path))
	}
	i.DstPath = outputPath
	return nil
}

func (i *SrcFile) CopyAndExec(outRoot string) error {
	err := i.Normalize(outRoot)
	if err != nil {
		return err
	}

	err = i.CheckConfiguration(outRoot)
	if err != nil {
		return err
	}

	if len(i.BeforeCmd) > 1 {
		err = i.ExecBeforeCmd(nil, nil)
		if err != nil {
			return err
		}
	}

	// filecopy
	err = copyFile(i.DstPath, i.Path)
	if err != nil {
		return fmt.Errorf("copyFile:%w", err)
	}

	if len(i.AfterCmd) > 1 {
		err = i.ExecAfterCmd(nil, nil)
		if err != nil {
			return err
		}
	}

	if i.ChecksumType != "" {
		sum, err := i.ChecksumStr(i.DstPath)
		if err != nil {
			return fmt.Errorf("CheckSumStr:%w", err)
		}
		sumPath := i.DstPath + "." + i.ChecksumType
		err = ioutil.WriteFile(sumPath, []byte(sum), 0644)
		if err != nil {
			return fmt.Errorf("ioutil.WriteFile:%w", err)
		}
	}

	return nil
}
