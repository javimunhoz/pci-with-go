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

package builder

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var httpd_c chan int
var builder *Builder

const (
	Httpd_no_action = iota
	Httpd_run_build
)

var regexps = map[string]*regexp.Regexp{
	"builders_re":    regexp.MustCompile("^/builders$"),
	"builder_re":     regexp.MustCompile("^/builders/[a-zA-Z0-9-_]+$"),
	"builder_run_re": regexp.MustCompile("^/builders/[a-zA-Z0-9-_]+/run$"),
	"builds_re":      regexp.MustCompile("^/builders/[a-zA-Z0-9-_]+/builds$"),
	"build_re":       regexp.MustCompile("^/builders/[a-zA-Z0-9-_]+/builds/[a-zA-Z0-9-_]+$"),
	"stages_re":      regexp.MustCompile("^/builders/[a-zA-Z0-9-_]+/builds/[a-zA-Z0-9-_]+/stages$"),
	"stage_re":       regexp.MustCompile("^/builders/[a-zA-Z0-9-_]+/builds/[a-zA-Z0-9-_]+/stages/[a-zA-Z0-9-_]+$"),
	"commands_re":    regexp.MustCompile("^/builders/[a-zA-Z0-9-_]+/builds/[a-zA-Z0-9-_]+/stages/[a-zA-Z0-9-_]+/commands$"),
}

func showHttpErrorMessage(w http.ResponseWriter, m string) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, "{ \"error\": \"%s\" }", m)
}

func showHttpBuilderErrorMessage(w http.ResponseWriter) {
	showHttpErrorMessage(w, "builder name doesn't match")
}

func showHttpBuildErrorMessage(w http.ResponseWriter) {
	showHttpErrorMessage(w, "build name doesn't match")
}

func existBuilder(r *http.Request) bool {
	builder_name := strings.Split(r.URL.Path, "/")[2]
	return builder_name == builder.name
}

func getBuild(r *http.Request) (build *Build) {
	build_name := strings.Split(r.URL.Path, "/")[4]
	for _, v := range builder.builds {
		if v.name == build_name {
			return v
		}
	}
	return nil
}

func existBuild(r *http.Request) bool {
	return getBuild(r) != nil
}

func getStage(r *http.Request) (stage *Stage) {
	build := getBuild(r)
	if build == nil {
		return nil
	}
	stage_name := strings.Split(r.URL.Path, "/")[6]
	for _, v := range build.stages {
		if v.name == stage_name {
			return v
		}
	}
	return nil
}

func existStage(r *http.Request) bool {
	return getStage(r) != nil
}

func showBuilders(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "{ \"builders\": [ \"%s\" ] }", builder.name)
}

func showBuilder(w http.ResponseWriter, r *http.Request) {
	if !existBuilder(r) {
		showHttpBuilderErrorMessage(w)
		return
	}
	fmt.Fprintf(w, "{ \"builder\": { \"name\": \"%s\" } }", builder.name)
}

func showBuilds(w http.ResponseWriter, r *http.Request) {
	if !existBuilder(r) {
		showHttpBuilderErrorMessage(w)
		return
	}
	fmt.Fprintf(w, "{ \"builds\": [ ")
	end := len(builder.builds) - 1
	for i, v := range builder.builds {
		fmt.Fprintf(w, "\"%s\"", v.name)
		if i != end {
			fmt.Fprintf(w, ", ")
		}
	}
	fmt.Fprintf(w, " ] }")
}

func showRawBuild(w http.ResponseWriter, r *http.Request, build *Build) {
	fmt.Fprintf(w,
		"{ \"build\": { { \"name\": \"%s\" }, { \"directory\": \"%s\" }, { \"priority\" : %d }, { \"state\" : %d }, { \"status\" : \"%s\" } } }",
		build.name,
		build.directory,
		build.priority,
		build.state,
		strconv.FormatBool(build.status))
}

func showBuild(w http.ResponseWriter, r *http.Request) {
	if !existBuilder(r) {
		showHttpBuilderErrorMessage(w)
		return
	}
	if !existBuild(r) {
		showHttpBuildErrorMessage(w)
		return
	}
	build := getBuild(r)
	fmt.Fprintf(w,
		"{ \"build\": { { \"name\": \"%s\" }, { \"directory\": \"%s\" }, { \"priority\" : %d }, { \"state\" : %d }, { \"status\" : \"%s\" } } }",
		build.name,
		build.directory,
		build.priority,
		build.state,
		strconv.FormatBool(build.status))
}

func showStages(w http.ResponseWriter, r *http.Request) {
	if !existBuilder(r) {
		showHttpBuilderErrorMessage(w)
		return
	}
	if !existBuild(r) {
		showHttpBuildErrorMessage(w)
		return
	}
	build := getBuild(r)
	fmt.Fprintf(w, "{ \"stages\": [ ")
	end := len(build.stages) - 1
	for i, v := range build.stages {
		fmt.Fprintf(w, "\"%s\"", v.name)
		if i != end {
			fmt.Fprintf(w, ", ")
		}
	}
	fmt.Fprintf(w, " ] }")
}

func showStage(w http.ResponseWriter, r *http.Request) {
	if !existStage(r) {
		m := "builder/build/stage don't match"
		showHttpErrorMessage(w, m)
		return
	}
	stage := getStage(r)
	fmt.Fprintf(w,
		"{ \"stage\": { { \"name\": \"%s\" }, { \"priority\" : %d }, { \"state\" : %d }, { \"status\" : \"%v\" } } }",
		stage.name,
		stage.priority,
		stage.state,
		stage.status)
}

func showCommands(w http.ResponseWriter, r *http.Request) {
	if !existStage(r) {
		m := "builder/build/stage don't match"
		showHttpErrorMessage(w, m)
		return
	}
	commands := getStage(r).commands
	end := len(commands) - 1
	fmt.Fprintf(w, "{ \"commands\" : [ ")
	for i, c := range commands {
		fmt.Fprintf(w, "{ \"name\": \"%s\", \"command\": \"%s\", \"params\": \"%s\", \"dir\": \"%s\", \"stdio\": \"%s\", \"status\": \"%s\" }", c.name, c.command, c.params, c.dir, c.stdio, strconv.FormatBool(c.status))
		if i != end {
			fmt.Fprintf(w, ", ")
		}
	}
	fmt.Fprintf(w, " ] }")
}

func handleGetMethod(w http.ResponseWriter, r *http.Request) {
	switch {
	case regexps["builders_re"].MatchString(r.URL.Path):
		showBuilders(w, r)
	case regexps["builder_re"].MatchString(r.URL.Path):
		showBuilder(w, r)
	case regexps["builds_re"].MatchString(r.URL.Path):
		showBuilds(w, r)
	case regexps["build_re"].MatchString(r.URL.Path):
		showBuild(w, r)
	case regexps["stages_re"].MatchString(r.URL.Path):
		showStages(w, r)
	case regexps["stage_re"].MatchString(r.URL.Path):
		showStage(w, r)
	case regexps["commands_re"].MatchString(r.URL.Path):
		showCommands(w, r)
	default:
		m := fmt.Sprintf("resource doesn't exist (%s)", r.URL.Path)
		log.Printf("error: %s\n", m)
		showHttpErrorMessage(w, m)
	}
	httpd_c <- Httpd_no_action
}

func updateBuildName(w http.ResponseWriter, r *http.Request) {
	if !existBuilder(r) {
		showHttpBuilderErrorMessage(w)
		return
	}
	if !existBuild(r) {
		showHttpBuildErrorMessage(w)
		return
	}
	build := getBuild(r)
	new_name := r.PostFormValue("name")
	if new_name != "" {
		build.name = new_name
		UpdateJSONFromBuilder(builder, on_disk)
	} else {
		m := "build name is not valid"
		showHttpErrorMessage(w, m)
		return
	}
	showRawBuild(w, r, build)
}

func updateBuildPriority(w http.ResponseWriter, r *http.Request) {
	if !existBuilder(r) {
		showHttpBuilderErrorMessage(w)
		return
	}
	if !existBuild(r) {
		showHttpBuildErrorMessage(w)
		return
	}
	build := getBuild(r)
	new_priority_str := r.PostFormValue("priority")
	if new_priority, err := strconv.Atoi(new_priority_str); err == nil {
		build.priority = new_priority
		UpdateJSONFromBuilder(builder, on_disk)
	} else {
		m := "build priority is not valid"
		showHttpErrorMessage(w, m)
		return
	}
	showRawBuild(w, r, build)
}

func updateBuildState(w http.ResponseWriter, r *http.Request) {
	if !existBuilder(r) {
		showHttpBuilderErrorMessage(w)
		return
	}
	if !existBuild(r) {
		showHttpBuildErrorMessage(w)
		return
	}
	build := getBuild(r)
	new_state_str := r.PostFormValue("state")
	new_state := str2state(new_state_str)
	if new_state != State_undefined {
		build.state = new_state
		UpdateJSONFromBuilder(builder, on_disk)
	} else {
		m := "build state is not valid"
		showHttpErrorMessage(w, m)
		return
	}
	showRawBuild(w, r, build)
}

func handleBuilderRun(w http.ResponseWriter, r *http.Request) bool {
	if !existBuilder(r) {
		showHttpBuilderErrorMessage(w)
		return false
	}
	showBuilder(w, r)
	return true
}

func handlePostMethod(w http.ResponseWriter, r *http.Request) {
	switch {
	case regexps["build_re"].MatchString(r.URL.Path):
		switch {
		case r.PostFormValue("name") != "":
			updateBuildName(w, r)
			httpd_c <- Httpd_run_build
			return
		case r.PostFormValue("priority") != "":
			updateBuildPriority(w, r)
			httpd_c <- Httpd_run_build
			return
		case r.PostFormValue("state") != "":
			updateBuildState(w, r)
			httpd_c <- Httpd_run_build
			return
		}
	case regexps["builder_run_re"].MatchString(r.URL.Path):
		if handleBuilderRun(w, r) {
			httpd_c <- Httpd_run_build
		} else {
			httpd_c <- Httpd_no_action
		}
		return
	}
	m := fmt.Sprintf("resource doesn't exist (%s)", r.URL.Path)
	log.Printf("error: %s\n", m)
	showHttpErrorMessage(w, m)
	httpd_c <- Httpd_no_action
}

func dispatcher(w http.ResponseWriter, r *http.Request) {
	switch strings.ToUpper(r.Method) {
	case "GET":
		handleGetMethod(w, r)
	case "POST":
		handlePostMethod(w, r)
	default:
		fmt.Fprintf(w, "http method not supported\n")
	}
}

func HttpServer(b *Builder) chan int {
	builder = b
	httpd_c = make(chan int)
	go func() {
		http.HandleFunc("/", dispatcher)
		ln, err := net.Listen("tcp", ":8080")
		if err != nil {
			log.Printf("error Listen\n")
		}
		for {
			err = http.Serve(ln, nil)
			if err != nil {
				continue
			}
		}
	}()
	return httpd_c
}
