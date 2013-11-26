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
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type jsonobject struct {
	Builder BuilderBody
}

type BuilderBody struct {
	Name   string
	Builds []BuildBody
}

type BuildBody struct {
	Name      string
	Directory string
	Priority  int
	State     string
	Stages    []StageBody
}

type StageBody struct {
	Name     string
	Priority int
	State    string
	Commands []CommandBody
}

type CommandBody struct {
	Name      string
	Command   string
	Args      string
	Directory string
}

var file_json string

func loadJSON(file string) jsonobject {
	log.Printf("Loading configuration '%s'", file)
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	var object jsonobject
	json.Unmarshal(content, &object)
	return object
}

func saveJSON(object jsonobject, file string) {
	log.Printf("Saving configuration '%s'", file)
	content, err := json.MarshalIndent(object, "", "   ")
	if err != nil {
		log.Println("error:", err)
		return
	}
	err = ioutil.WriteFile(file, content, 0666)
	if err != nil {
		log.Println("error:", err)
		return
	}
}

func NewBuilderFromCurrentJSON() *Builder {
	return NewBuilderFromJSON(file_json)
}

func NewBuilderFromJSON(file string) *Builder {
	file_json = file
	object := loadJSON(file_json)
	// TODO: check jsonobject is right!
	builder := NewBuilder(object.Builder.Name)
	for build_i, build_v := range object.Builder.Builds {
		build := NewBuild(build_v.Name, build_v.Directory, build_v.Priority, str2state(build_v.State))
		for stage_i, stage_v := range object.Builder.Builds[build_i].Stages {
			commands := NewShellCommands()
			for _, command_v := range object.Builder.Builds[build_i].Stages[stage_i].Commands {
				command := NewShellCommand(command_v.Name,
					command_v.Command,
					command_v.Args,
					command_v.Directory,
					build_v.Directory)
				commands.Add(command)
			}
			stage := NewStage(stage_v.Name,
				stage_v.Priority,
				str2state(stage_v.State))
			stage.AddCommands(commands)
			build.AddStage(stage)
		}
		builder.AddBuild(build)
	}
	builder.SetIdle(false)
	return builder
}

func UpdateJSONFromBuilder(builder *Builder, flag bool) {
	// save to disk?
	if !flag {
		return
	}
	// build jsonobject from builder
	var object jsonobject
	object.Builder.Name = builder.name
	for _, build_v := range builder.builds {
		var build_body BuildBody
		build_body.Name = build_v.name
		build_body.Directory = build_v.directory
		build_body.Priority = build_v.priority
		build_body.State = state2str(build_v.state)
		for _, stage_v := range build_v.stages {
			var stage_body StageBody
			stage_body.Name = stage_v.name
			stage_body.Priority = stage_v.priority
			stage_body.State = state2str(stage_v.state)
			for _, command_v := range stage_v.commands {
				var command_body CommandBody
				command_body.Name = command_v.name
				command_body.Command = command_v.command
				command_body.Args = command_v.params
				command_body.Directory = command_v.dir
				stage_body.Commands = append(stage_body.Commands, command_body)
			}
			build_body.Stages = append(build_body.Stages, stage_body)
		}
		object.Builder.Builds = append(object.Builder.Builds, build_body)
	}
	// save to disk
	saveJSON(object, file_json+".new")
}
