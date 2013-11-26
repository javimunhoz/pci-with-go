/*-
 * Copyright (c) 2013 Javier M. Mellid
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 * 1. Redistributions of source code must retain the above copyright
 *    notice, this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright
 *    notice, this list of conditions and the following disclaimer in the
 *    documentation and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE NETBSD FOUNDATION, INC. AND CONTRIBUTORS
 * ``AS IS'' AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED
 * TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR
 * PURPOSE ARE DISCLAIMED.  IN NO EVENT SHALL THE FOUNDATION OR CONTRIBUTORS
 * BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */

package main

import (
	"flag"
	_b "pci/builder"
)

func main() {
	update_json := flag.Bool("update-json", false, "update json conf file")
	conf_json := flag.String("conf-json", "", "json conf file")
	flag.Parse()
	builder := _b.NewBuilderFromJSON(*conf_json)
	builder.UpdateOnDisk(update_json)
	runNextStage := builder.RunStage()
	httpd_c := _b.HttpServer(builder)
	build_c := make(chan bool)
	go func() { build_c <- true }()
	for {
		select {
		case next_one := <-build_c:
			go func() {
				if next_one {
					if builder.IsIdle() {
						builder.SetIdle(false)
						runNextStage = builder.Reload()
					}
					build_c <- runNextStage()
				}
			}()
		case httpd_action := <-httpd_c:
			go func() {
				if httpd_action == _b.Httpd_run_build {
					if builder.IsIdle() {
						build_c <- true
					}
				}
			}()
		}
	}
}
