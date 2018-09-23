// +build mage

package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/mholt/archiver"
)

var Default = Protoc
var basedir string
var protocPath string

func init() {
	var err error
	basedir, err = os.Getwd()
	if err != nil {
		panic("unable to determine local directory")
	}
	for {
		if _, err := os.Stat(filepath.Join(basedir, "go.mod")); err == nil {
			break
		}
		if basedir == filepath.Dir(basedir) {
			panic("unable to find project root. no go.mod file in path")
		}
		basedir = filepath.Dir(basedir)
	}
	protocPath = filepath.Join(mg.CacheDir(), "build-cache", "protoc")
}

func DownloadProtoc() error {
	var url string
	switch runtime.GOOS {
	case "linux":
		url = "https://github.com/protocolbuffers/protobuf/releases/download/v3.6.1/protoc-3.6.1-linux-x86_64.zip"
	case "windows":
		url = "https://github.com/protocolbuffers/protobuf/releases/download/v3.6.1/protoc-3.6.1-win32.zip"
	default:
		url = "https://github.com/protocolbuffers/protobuf/releases/download/v3.6.1/protoc-3.6.1-osx-x86_64.zip"
	}
	if _, err := os.Stat(protocPath); err != nil {
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("unable to download protoc from %s: %v", url, err)
		}
		defer resp.Body.Close()
		if err := archiver.Zip.Read(resp.Body, filepath.Join(protocPath)); err != nil {
			return fmt.Errorf("unable to unzip protoc: %v", err)
		}
	}
	return nil
}

func getModPath(modName string) string {
	c := exec.Command("go", "list", "-m", "-f={{.Dir}}", modName)
	out, err := c.Output()
	if err != nil {
		panic(fmt.Errorf("unable to locate %s module: %v", modName, err))
	}
	return strings.TrimSpace(string(out))
}

func BuildGogoFast() error {
	genPath := filepath.Join(protocPath, "bin", "protoc-gen-gogofast")
	if _, err := os.Stat(genPath); err != nil {
		c := exec.Command("go", "build", "-o", genPath, "github.com/gogo/protobuf/protoc-gen-gogofast")
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}
	return nil
}

func Protoc() error {
	mg.Deps(DownloadProtoc, BuildGogoFast)
	winExt := ""
	if runtime.GOOS == "windows" {
		winExt = ".exe"
	}

	gogoProtoMapping := "Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types," +
		"Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types," +
		"Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types," +
		"Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types," +
		"Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types," +
		"Mgoogle/protobuf/field_mask.proto=github.com/gogo/protobuf/types"

	return filepath.Walk(basedir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".proto" {
			c := exec.Command(
				filepath.Join(protocPath, "bin", "protoc"+winExt),
				"-I="+basedir,
				"-I="+filepath.Join(getModPath("github.com/gogo/protobuf"), "gogoproto"),
				"--plugin=protoc-gen-gogofast="+filepath.Join(protocPath, "bin", "protoc-gen-gogofast"),
				"--gogofast_out=plugins=grpc,"+gogoProtoMapping+",paths=source_relative:.",
				path,
			)
			c.Stderr = os.Stderr
			c.Stdout = os.Stdout

			return c.Run()
		}
		return nil
	})
}

func Clean() {
	filepath.Walk(basedir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".pb.go") {
			return os.Remove(path)
		}
		return nil
	})
	os.RemoveAll("build")
}
