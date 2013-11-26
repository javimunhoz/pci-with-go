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
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
)

type shellCommand struct {
	name    string
	command string
	params  string
	dir     string
	stdio   string
	status  bool
}

func NewShellCommand(
	name string,
	command string,
	params string,
	dir string,
	stdio string) *shellCommand {
	return &shellCommand{
		name:    name,
		command: command,
		params:  params,
		dir:     dir,
		stdio:   stdio,
		status:  false}
}

func (c *shellCommand) Execute() {
	var out []byte
	var ok bool
	cmd := fmt.Sprintf("$ %s %s\n", c.command, c.params)
	c.writeOutputToFile([]byte(cmd))
	if ok, out = c.runCommand(); ok {
		c.status = true
	} else {
		c.status = false
	}
	c.writeOutputToFile(out)
}

func (c *shellCommand) runCommand() (bool, []byte) {
	ok := true
	cmd := exec.Command(c.command, c.params)
	cmd.Dir = c.stdio
	out, err := cmd.CombinedOutput()
	if err != nil {
		ok = false
	}
	return ok, out
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (c *shellCommand) writeOutputToFile(out []byte) {
	var buffer bytes.Buffer
	if ok, _ := exists(c.stdio); !ok {
		log.Printf("error: %s doesn't exist", c.stdio)
		os.Exit(-1)
	}
	buffer.WriteString(c.stdio)
	buffer.WriteString("/stdio.txt")
	fo, err := os.OpenFile(
		buffer.String(),
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0660)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()
	if _, err := fo.Write(out); err != nil {
		panic(err)
	}
}

type shellCommands []*shellCommand

func NewShellCommands() shellCommands {
	var commands shellCommands
	return commands
}

func (commands *shellCommands) Add(sc *shellCommand) {
	*commands = append(*commands, sc)
}

func (commands *shellCommands) Execute() {
	for _, c := range *commands {
		c.Execute()
	}
}

func (commands *shellCommands) GetCommands() *commandIterator {
	return NewCommandIterator(commands)
}
