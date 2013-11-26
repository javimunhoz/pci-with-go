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
	"log"
)

type Builder struct {
	name    string
	builds  []*Build
	running bool
}

func NewBuilder(name string) *Builder {
	return &Builder{name: name, running: false}
}

func (b *Builder) AddBuild(build *Build) {
	b.builds = append(b.builds, build)
}

func (b *Builder) PickBuildByPriority() (winner *Build) {
	var build *Build
	var i int
	winner = nil
	for i, build = range b.builds {
		if build.state == State_ready {
			winner = build
			i++
			break
		}
	}
	if winner == nil {
		return nil
	}
	for _, build = range b.builds[i:] {
		if build.state == State_ready {
			if build.priority < winner.priority {
				winner = build
				break
			}
		}
	}
	winner.state = State_building
	return winner
}

func (b *Builder) Schedule(global_state *GlobalState) (build *Build, stage *Stage) {
	if global_state.Current_build == nil {
		build = b.PickBuildByPriority()
	} else {
		build = global_state.Current_build
	}
	if build != nil {
		global_state.Current_build = build
		stage = build.PickStageByPriority()
		if stage != nil {
			stage.state = State_building
		} else {
			build.state = State_finished
			global_state.Current_build = nil
		}
	}
	return build, stage
}

func (b *Builder) BuildStep(build *Build, stage *Stage) {
	stage.Execute()
	stage.state = State_finished
	if !stage.status {
		for _, s := range build.stages {
			s.state = State_finished
		}
		build.state = State_finished
	}
}

func (b *Builder) RunStage() func() bool {
	global_state := NewGlobalState()
	return func() bool {
		build, stage := b.Schedule(global_state)
		if stage != nil {
			log.Printf("Executing %s %s", build.name, stage.name)
			b.BuildStep(build, stage)
		}
		if build == nil {
			builder.SetIdle(true)
			return false
		} else {
			builder.SetIdle(false)
			UpdateJSONFromBuilder(b, on_disk)
			return true
		}
	}
}

func (b *Builder) GetBuildsByState(state int) []*Build {
	log.Printf("--> ")
	for _, build := range b.builds {
		if build.state == state {
			log.Printf("   %s\n", build.name)
		}
	}
	return nil
}

var on_disk bool = false

func (b *Builder) UpdateOnDisk(update_json *bool) {
	// we want on_disk being global while related to Builder
	on_disk = *update_json
}

func (b *Builder) IsIdle() bool {
	return !b.running
}

func (b *Builder) SetIdle(v bool) {
	b.running = !v
}

func (b *Builder) Reload() func() bool {
	builder := NewBuilderFromCurrentJSON()
	runNextStage := builder.RunStage()
	UpdateJSONFromBuilder(builder, on_disk)
	return runNextStage
}
