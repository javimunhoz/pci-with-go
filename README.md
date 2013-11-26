Personal Continuous Integration (PCI) with Go
=============================================

This project explores concurrency and light architectures in Personal
Continuous Integration software.

Source code is implemented in Go. REST API over HTTP is available.

More information about this project:

[http://javiermunhoz.com/blog/2013/11/26/personal-continuous-integration-with-go.html](http://javiermunhoz.com/blog/2013/11/26/personal-continuous-integration-with-go.html)

Licensing
=========

PCI is freely redistributable under the two-clause BSD License. Use of this
source code is governed by a BSD-style license that can be found in the
`LICENSE` file.

Dependencies
============

1. This code was developed and tested in a GNU/Linux system ([Debian GNU/Linux](http://www.debian.org))
2. It requires Go installed. Tested with 'go version go1.1.2 linux/amd64'

Compiling and running
=====================

1. Grab the code with Git. Use the following command:

   ~$ git clone https://github.com/javimunhoz/pci-with-go

2. Set up your environment

   ~$ cd pci-with-go

   ~$ . scripts/environ.sh

3. Compile the sources

   ~$ make -C src/

4. Run it

   ~$ pci -conf-json conf/example-conf.json

Notes:

   - you can use example-conf.json to craft your new configuration file
   - use '-update-json=true' to save the updated runtime configuration to disk
