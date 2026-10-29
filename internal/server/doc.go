// Package server checks that a machine can run Astra, installs the daemon on it, and pairs a
// client with it (F1-02 to F1-06, ADR 0007). It works through a Runner, which runs commands on
// the machine: over SSH from a laptop, or locally when installing on the VPS itself, so both
// paths run exactly the same steps.
//
// Nothing here changes a machine before a person has seen the plan and agreed to it, and every
// privileged step is a short script fed to `sudo sh` on stdin, so no value is ever quoted into a
// command line.
package server
