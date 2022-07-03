package main

// import "path/filepath"

type Config struct {
	SrcDir string // defaults to the current directory
	OutFile string // defaults to "wasm-wrappers.go"
	ExportWrappers bool 
}

// func NewConfig() *Config {
// 	config := &Config{}

// }