// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/blang/semver/v4"
)

const defaultCGOEnabled = "1"

var defaultPackageCachePath = filepath.Join(os.TempDir(), "go-pkg")

// OsArch creates a easy map of OS and Arch combinations.
type OsArch struct {
	Os   string
	Arch string
}

var defaultOsArches = []OsArch{
	{"darwin", "amd64"},
	{"darwin", "arm64"},
	{"linux", "amd64"},
	{"linux", "arm"},
	{"linux", "arm64"},
	{"windows", "amd64"},
	{"freebsd", "amd64"},
}

var (
	components   string
	buildName    string
	osArches     string
	skipOsArches string
	buildTags    string
)

// Env contains all necessary environment settings for the build.
type Env struct {
	Log        *log.Logger
	BranchName string
	ReleaseDir string
	WorkDir    string

	Commit struct {
		Version   semver.Version
		Timestamp time.Time
		Hash      string
		Release   bool
	}

	GoVersion        string
	PackageCachePath string
	CGOENABLED       string
	LDFLAGS          string
}

func main() {
	env := Env{}
	env.Log = log.New(os.Stderr, "", log.Lshortfile)
	flag.StringVar(&env.ReleaseDir, "release-dir", "release", "release directory")
	flag.StringVar(&env.BranchName, "branch", "", "branch name to use for tagging")
	flag.StringVar(&components, "components", "", "comma separated list of components to build within the repo")
	flag.StringVar(&buildName, "build-name", "", "build name if building root of repo instead of providing a component list")
	flag.StringVar(&env.GoVersion, "go-version", "", "go version to use for building the image")
	flag.StringVar(&buildTags, "build-tags", "", "build tags")

	flag.StringVar(&osArches, "osarches", "", "comma-separated list of os/arch to build for")
	flag.StringVar(&skipOsArches, "skip-osarches", "", "comma-separated list of os/arch to skip build for")

	flag.Parse()

	var err error

	env.WorkDir, err = os.Getwd()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	if env.GoVersion == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Please specify go version with -go-version (tag of storjlabs/golang Docker image)")
		os.Exit(1)
	}

	env.LDFLAGS = strings.TrimSpace(os.Getenv("LDFLAGS"))

	gopath, err := getGoEnv("GOPATH")
	if err != nil {
		fmt.Println("could not find GOPATH, defaulting to", defaultPackageCachePath, "for package cache")
		env.PackageCachePath = defaultPackageCachePath
	} else {
		env.PackageCachePath = filepath.Join(gopath, "pkg")
	}

	cgoEnabled, err := getGoEnv("CGO_ENABLED")
	if err != nil {
		fmt.Println("could not find CGO_ENABLED, defaulting to", defaultCGOEnabled)
		env.CGOENABLED = defaultCGOEnabled
	} else {
		env.CGOENABLED = cgoEnabled
	}

	err = env.FetchVersionInfo()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to get version: %v\n", err)
		os.Exit(1)
	}

	componentList := strings.Split(components, ",")

	err = env.BuildComponents(componentList)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to build: %v\n", err)
		os.Exit(1)
	}
}

// getGoEnv attempts to get a particular Go environment variable. First it checks
// if the environment variable exists in the OS, and returns that if it exists.
// failing that, it attempts to find it using the `go env` command.
func getGoEnv(name string) (string, error) {
	env := strings.TrimSpace(os.Getenv(name))
	if env != "" {
		return env, nil
	}

	value, err := exec.Command("go", "env", name).CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(value)), nil
}

// FetchVersionInfo gets the version information from the git tag.
func (env *Env) FetchVersionInfo() error {
	out, err := exec.Command("git", "describe", "--tags", "--exact-match", "--match", `v[0-9]*.[0-9]*.[0-9]*`).CombinedOutput()
	env.Log.Printf("git describe tags output: %v", string(out))
	if err != nil {
		env.Log.Printf("failed to get git tag for last commit: %v", string(out))
		out = []byte("v0.0.0")
	}

	version, err := semver.ParseTolerant(string(out))
	if err != nil {
		return fmt.Errorf("failed to parse %q: %w", string(out), err)
	}

	out, err = exec.Command("git", "rev-parse", "--short", "HEAD").CombinedOutput()
	env.Log.Printf("git reverse parse output: %v", string(out))
	if err != nil {
		return fmt.Errorf("failed to get git commit hash: %w", err)
	}

	env.Commit.Version = version
	env.Commit.Timestamp = time.Now().UTC()
	env.Commit.Hash = strings.TrimSuffix(string(out), "\n")
	env.Commit.Release = true // ToDo: flip to false if not a tag or modified code exists

	return nil
}

// ConstructFolderName creates the folder name for storing the release assets.
func (env *Env) ConstructFolderName() string {
	// custom branch
	if env.BranchName != "" {
		return fmt.Sprintf("%s-%s-go%s", env.Commit.Hash, env.BranchName, env.GoVersion)
	}
	// main branch
	if env.Commit.Version.String() == "0.0.0" {
		return fmt.Sprintf("%s-go%s", env.Commit.Hash, env.GoVersion)
	}
	// release tag
	return fmt.Sprintf("%s-v%s-go%s", env.Commit.Hash, env.Commit.Version.String(), env.GoVersion)
}

// BuildComponents builds binaries for the passed in component list.
func (env *Env) BuildComponents(components []string) error {
	tagDir := filepath.Join(env.ReleaseDir, env.ConstructFolderName())
	if err := os.MkdirAll(tagDir, 0755); err != nil {
		return fmt.Errorf("failed to create tagdir %q: %w", tagDir, err)
	}

	skippedOsArches := make(map[string]struct{})

	for _, s := range strings.Split(skipOsArches, ",") {
		skippedOsArches[s] = struct{}{}
	}

	for _, component := range components {
		err := env.BuildComponent(tagDir, component, skippedOsArches)
		if err != nil {
			return fmt.Errorf("failed component %q: %w", component, err)
		}
	}
	return nil
}

// BuildComponent builds the binaries for the different OsArch's for the specified component.
func (env *Env) BuildComponent(tagdir, component string, skippedOsArches map[string]struct{}) error {
	var osarches []OsArch
	if osArches != "" {
		for _, v := range strings.Split(osArches, ",") {
			split := strings.Split(v, "/")
			if len(split) != 2 {
				return fmt.Errorf("wrong format for os/arch: %s", v)
			}
			osarches = append(osarches, OsArch{split[0], split[1]})
		}
	} else {
		osarches = defaultOsArches
	}

	for _, osarch := range osarches {
		if _, ok := skippedOsArches[fmt.Sprintf("%s/%s", osarch.Os, osarch.Arch)]; ok {
			env.Log.Printf("Skipped build for %v", osarch)
			continue
		}
		if err := env.BuildComponentBinary(tagdir, component, osarch); err != nil {
			return fmt.Errorf("failed osarch %q: %w", osarch, err)
		}
	}
	return nil
}

// BuildComponentBinary builds the actual binary for specified component on OsArch.
func (env *Env) BuildComponentBinary(tagdir, component string, osarch OsArch) error {
	name := filepath.Base(component)
	if component == "" {
		// if no component, we're building the root of the repo.
		name = buildName
	}

	binaryName := name + "_" + osarch.Os + "_" + osarch.Arch
	if osarch.Os == "windows" {
		binaryName += ".exe"
	}

	binaryPath := filepath.Join(tagdir, binaryName)

	if osarch.Os == "windows" {
		// build version info for Windows.
		versionInfoTemplate := filepath.Join(component, "versioninfo.json")
		defer func() { _ = os.Remove(versionInfoTemplate) }()

		if err := os.WriteFile(versionInfoTemplate, []byte(versionInfo), 0644); err != nil {
			return fmt.Errorf("failed to write versioninfo.json: %w", err)
		}

		iconfile := filepath.Join(component, "storj.ico")
		defer func() { _ = os.Remove(iconfile) }()

		if err := os.WriteFile(iconfile, icondata, 0644); err != nil {
			return fmt.Errorf("failed to write storj.ico: %w", err)
		}

		resourcesyso := filepath.Join(component, "resource.syso")
		defer func() { _ = os.Remove(resourcesyso) }()

		var args []string
		if osarch.Arch == "amd64" {
			args = append(args, "-64")
		}

		var version string
		if env.Commit.Release {
			version = "release"
		} else {
			version = "dev"
		}

		args = append(args,
			"-o", resourcesyso,
			"-original-name", binaryName,
			"-description", name+" program for Storj",
			"-product-ver-major", fmt.Sprintf("%d", env.Commit.Version.Major),
			"-ver-major", fmt.Sprintf("%d", env.Commit.Version.Major),
			"-product-ver-minor", fmt.Sprintf("%d", env.Commit.Version.Minor),
			"-ver-minor", fmt.Sprintf("%d", env.Commit.Version.Minor),
			"-product-ver-patch", fmt.Sprintf("%d", env.Commit.Version.Patch),
			"-ver-patch", fmt.Sprintf("%d", env.Commit.Version.Patch),
			"-product-version", version,
			"-special-build", version,
			"-icon", iconfile,
			versionInfoTemplate,
		)

		out, err := exec.Command("goversioninfo", args...).CombinedOutput()
		env.Log.Println("goversioninfo: ", string(out))
		if err != nil {
			return fmt.Errorf("failed to run goversioninfo: %w", err)
		}
	}

	user, err := user.Current()
	if err != nil {
		return fmt.Errorf("could not get current user: %w", err)
	}

	ldFlags := fmt.Sprintf("-X storj.io/private/version.buildTimestamp=%d ", env.Commit.Timestamp.Unix()) +
		fmt.Sprintf("-X storj.io/private/version.buildCommitHash=%s ", env.Commit.Hash) +
		fmt.Sprintf("-X storj.io/private/version.buildVersion=%s ", env.Commit.Version.String()) +
		fmt.Sprintf("-X storj.io/private/version.buildRelease=%t ", env.Commit.Release)
	if env.LDFLAGS != "" {
		ldFlags = ldFlags + " " + env.LDFLAGS
	}

	runArgs := []string{
		"run", "--rm",
		// don't build as root
		"-u", user.Uid + ":" + user.Gid,
		// setup build folder
		"-v", env.WorkDir + ":/go/build",
		"-w", "/go/build",
		// use a shared package cache to avoid downloading
		"-v", env.PackageCachePath + ":/go/pkg",
		// setup correct os/arch
		"-e", "GOOS=" + osarch.Os, "-e", "GOARCH=" + osarch.Arch,
		"-e", "GOARM=6", "-e", "CGO_ENABLED=" + env.CGOENABLED,
		// use goproxy
		"-e", "GOPROXY",
		// use our golang image
		"storjlabs/golang:" + env.GoVersion,
	}

	buildArgs := []string{
		"go", "build", "-ldflags", ldFlags, "-o", filepath.ToSlash(binaryPath),
	}
	if buildTags != "" {
		buildArgs = append(buildArgs, "-tags", buildTags)
	}

	buildArgs = append(buildArgs, "./"+component)

	runArgs = append(runArgs, buildArgs...)

	cmd := exec.Command("docker", runArgs...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	/*if err := os.Chmod(binaryPath, 0755); err != nil {
		return fmt.Errorf("failed to make executable %q: %w", binaryPath, err)
	}*/

	if osarch.Arch == "windows" {
		signer, err := exec.LookPath("storj-sign")
		if err != nil {
			env.Log.Printf("skipping signing because storj-sign is missing: %v", err)
		} else {
			out, err := exec.Command(signer, binaryPath).CombinedOutput()
			env.Log.Printf("storj-sign: %v", string(out))
			if err != nil {
				return fmt.Errorf("failed to sign %q: %w", binaryPath, err)
			}
		}
	}

	return nil
}

var versionInfo = `{
    "FixedFileInfo": {
        "FileVersion": {
            "Major": 0,
            "Minor": 0,
            "Patch": 0,
            "Build": 0
        },
        "ProductVersion": {
            "Major": 0,
            "Minor": 0,
            "Patch": 0,
            "Build": 0
        },
        "FileFlagsMask": "3f",
        "FileFlags ": "00",
        "FileOS": "040004",
        "FileType": "01",
        "FileSubType": "00"
    },
    "StringFileInfo": {
        "Comments": "",
        "CompanyName": "Storj Labs, Inc",
        "FileDescription": "OVERWRITE",
        "FileVersion": "",
        "InternalName": "",
        "LegalCopyright": "© Storj Labs, Inc",
        "LegalTrademarks": "Storj is a trademark of Storj Labs Inc.\nTardigrade is a trademark of Storj Labs Inc.",
        "OriginalFilename": "OVERWRITE",
        "PrivateBuild": "OVERWRITE",
        "ProductName": "Storj",
        "ProductVersion": "OVERWRITE",
        "SpecialBuild": ""
    },
    "VarFileInfo": {
        "Translation": {
            "LangID": "0409",
            "CharsetID": "04B0"
        }
    },
    "IconPath": "OVERWRITE",
    "ManifestPath": ""
}`
