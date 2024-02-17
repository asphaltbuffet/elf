// Package fileyourself implements utility routines for manipulating filename paths in a way
// compatible with the target operating system-defined file paths.
//
// It can generally be used as a drop-in replacement for path/filepath with a notable
// exception for the return values from filepath.Base() and filepath.Ext().
package fileyourself
