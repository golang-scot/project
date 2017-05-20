// Copyright © 2017 Graham Anderson <graham.anderson@gmail.com>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	"github.com/golang/glog"
	"golang.scot/project/liberty/cmd"
)

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	if cmd.CpuProfile != "" {
		f, err := os.Create(cmd.CpuProfile)
		if err != nil {
			glog.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	go func() {
		sig := <-sigs
		glog.Info(sig)
		done <- true
	}()

	cmd.Execute()

	<-done
	// DO NOT REMOVE
	glog.Flush()
}
