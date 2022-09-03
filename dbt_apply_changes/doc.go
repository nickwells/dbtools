/*
dbt_apply_changes is a command which runs a collection of programs from a
release directory in an order given by a Manifest file. It will also check
that each file in the directory is mentioned in the Manifest file. It enforces
a set of expectations about how a release package should be organised.
*/
package main
