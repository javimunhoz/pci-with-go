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

type Build struct {
	name      string
	directory string
	priority  int
	state     int
	stages    []*Stage
	status    bool
}

func NewBuild(name string,
	directory string,
	priority int,
	state int) *Build {
	return &Build{name: name,
		directory: directory,
		priority:  priority,
		state:     state,
		status:    true}
}

func (b *Build) AddStage(s *Stage) {
	b.stages = append(b.stages, s)
}

func (b *Build) GetStages() *stageIterator {
	return NewStageIterator(b.stages)
}

func (b *Build) PickStageByPriority() (winner *Stage) {
	var stage *Stage
	var i int
	winner = nil
	for i, stage = range b.stages {
		if stage.state == State_ready {
			winner = stage
			i++
			break
		}
	}
	if winner == nil {
		return nil
	}
	for _, stage = range b.stages[i:] {
		if stage.state == State_ready {
			if stage.priority < winner.priority {
				winner = stage
				break
			}
		}
	}
	winner.state = State_building
	return winner
}
