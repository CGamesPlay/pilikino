package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/pelletier/go-toml"
)

// TomlFilename is the filename in the root directory where the configuration
// file lives.
const TomlFilename = ".pilikino.toml"

// Config holds all of the configuration for the current invocation of pilikino.
type Config struct {
	// FileNameTemplate is the fileew
	FileNameTemplate string `toml:"filename" comment:"defines the filename of new notes"`
	NewFileTemplate  string `toml:"template" multiline:"true" comment:"defines the content of new notes"`
}

// NewFileInfo holds the information necessary to create a new document.
type NewFileInfo struct {
	Title string
	Date  time.Time
	Tags  []string
}

// DefaultConfig is the default configuration object. It's prepopulated with
// default values, but is modified by the Load method.
var DefaultConfig = Config{
	FileNameTemplate: `{{.Date | date "20060102-1504"}}-{{.Title | kebabcase}}.md`,
	NewFileTemplate: `---
title: {{.Title}}
date: {{.Date | date "2006-01-02 15:04"}}
tags: {{.Tags | join " "}}
---

`,
}

// Load attempts to load the configuration file in the current directory.
func (config *Config) Load() error {
	bytes, err := ioutil.ReadFile(TomlFilename)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	return toml.Unmarshal(bytes, config)
}

// Save attempts to write out the current configuration to the configuration
// file.
func (config *Config) Save(overwrite bool) error {
	bytes, err := toml.Marshal(*config)
	if err != nil {
		return err
	}
	flags := os.O_WRONLY | os.O_CREATE
	if !overwrite {
		flags |= os.O_EXCL
	}
	f, err := os.OpenFile(TomlFilename, flags, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(bytes)
	return err
}

// GetFilename returns the correct name of the note file given the title and
// creation time.
func (config *Config) GetFilename(info *NewFileInfo) (string, error) {
	t, err := template.New("filename").
		Funcs(sprig.TxtFuncMap()).
		Parse(config.FileNameTemplate)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	err = t.Execute(buf, info)
	return buf.String(), err
}

// GetTemplate returns the configured default contents of a new file.
func (config *Config) GetTemplate(info *NewFileInfo) (string, error) {
	t, err := template.New("template").
		Funcs(sprig.TxtFuncMap()).
		Parse(config.NewFileTemplate)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	err = t.Execute(buf, info)
	return buf.String(), err
}
