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
	"os/exec"
	"strings"
)

func execCommand(f map[string]string, args []string, outio io.Writer, errio io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("command not found")
	}

	// replace placeholder
	for k, v := range f {
		for i, arg := range args {
			if strings.Compare(k, arg) == 0 {
				args[i] = v
			}
		}
	}

	var cmd *exec.Cmd

	cmd = exec.Command(args[0], args[1:]...)
	//	fmt.Printf("cmd:%s %s\n", cmdargs[0], cmdargs[1:])
	if outio != nil {
		cmd.Stdout = outio
	}
	if errio != nil {
		cmd.Stderr = errio
	}

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%s:%s\n", args, err)
	}
	return nil
}
