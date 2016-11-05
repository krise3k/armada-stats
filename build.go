// +build ignore

package main

import (
	"bytes"
	"crypto/md5"
	"flag"
	"fmt"
	"github.com/krise3k/armada-stats/utils"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var (
	versionRe             = regexp.MustCompile(`-[0-9]{1,3}-g[0-9a-f]{5,10}`)
	goarch                string
	goos                  string
	version               string = "v1"
	linuxPackageIteration string = ""
	race                  bool
	workingDir            string
	serverBinaryName      string = "armada-stats"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)

	ensureGoPath()
	workingDir, _ = os.Getwd()
	version = utils.ReadVersion()
	log.Printf("Version: %s, Package Iteration: %s\n", version, linuxPackageIteration)

	flag.StringVar(&goarch, "goarch", runtime.GOARCH, "GOARCH")
	flag.StringVar(&goos, "goos", runtime.GOOS, "GOOS")
	flag.BoolVar(&race, "race", race, "Use race detector")
	flag.Parse()

	if flag.NArg() == 0 {
		log.Println("Usage: go run build.go build")
		return
	}

	for _, cmd := range flag.Args() {
		switch cmd {

		case "build":
			pkg := "."
			clean()
			build(pkg, []string{})

		case "package":
			createLinuxPackages()

		case "pkg-deb":
			createDebPackages()

		case "pkg-rpm":
			createRpmPackages()

		case "clean":
			clean()

		default:
			log.Fatalf("Unknown command %q", cmd)
		}
	}
}

func createLinuxPackages() {
	createDebPackages()
	createRpmPackages()
}

func createDebPackages() {
	createPackage(linuxPackageOptions{
		packageType:    "deb",
		binPath:        "/usr/local/bin/armada-stats",
		configDir:      "/etc/armada-stats",
		configFilePath: "/etc/armada-stats/armada-stats.yml",

		etcDefaultPath:         "/etc/default",
		etcDefaultFilePath:     "/etc/default/armada-stats.yml",
		systemdServiceFilePath: "/usr/lib/systemd/system/armada-stats.service",

		systemdFileSrc: "packaging/systemd/armada-stats.service",
		postinstSrc:    "packaging/deb/control/postinst",
		defaultFileSrc: "conf/defaults.yml",
	})
}

func createRpmPackages() {
	createPackage(linuxPackageOptions{
		packageType:    "rpm",
		binPath:        "/usr/local/bin/armada-stats",
		configDir:      "/etc/armada-stats",
		configFilePath: "/etc/armada-stats/armada-stats.yml",

		etcDefaultPath:         "/etc/default",
		etcDefaultFilePath:     "/etc/default/armada-stats.yml",
		systemdServiceFilePath: "/usr/lib/systemd/system/armada-stats.service",

		systemdFileSrc: "packaging/systemd/armada-stats.service",
		postinstSrc:    "packaging/rpm/control/postinst",
		defaultFileSrc: "conf/defaults.yml",
	})
}
func createPackage(options linuxPackageOptions) {
	packageRoot, _ := ioutil.TempDir("", "armada-stats-linux-pack")

	// create directories
	runPrint("mkdir", "-p", filepath.Join(packageRoot, options.configDir))
	runPrint("mkdir", "-p", filepath.Join(packageRoot, options.etcDefaultPath))
	runPrint("mkdir", "-p", filepath.Join(packageRoot, "/usr/local/bin"))
	runPrint("mkdir", "-p", filepath.Join(packageRoot, "/usr/lib/systemd/system"))
	//create package directory
	runPrint("mkdir", "-p", filepath.Join(workingDir, "/dist"))

	// copy binary
	runPrint("cp", "-p", filepath.Join(workingDir, "tmp/bin/"+serverBinaryName), filepath.Join(packageRoot, options.binPath))
	// copy environment var file
	runPrint("cp", "-p", options.defaultFileSrc, filepath.Join(packageRoot, options.etcDefaultFilePath))
	// copy config file
	runPrint("cp", "conf/armada-stats.yml", filepath.Join(packageRoot, options.configFilePath))
	// copy systemd file
	runPrint("cp", "-p", options.systemdFileSrc, filepath.Join(packageRoot, options.systemdServiceFilePath))
	args := []string{
		"-s", "dir",
		"--description", "armada-stats",
		"-C", packageRoot,
		"--license", "\"Apache 2.0\"",
		"--maintainer", "krise3k@github.com",
		"--url", "https://github.com/krise3k/armada-stats",
		"--config-files", options.configFilePath,
		"--config-files", options.etcDefaultFilePath,
		"--config-files", options.systemdServiceFilePath,
		"--after-install", options.postinstSrc,
		"--name", "armada-stats",
		"--version", version,
		"-p", "./dist",
	}

	if linuxPackageIteration != "" {
		args = append(args, "--iteration", linuxPackageIteration)
	}

	args = append(args, ".")

	fmt.Println("Creating package: ", options.packageType)
	runPrint("fpm", append([]string{"-t", options.packageType}, args...)...)
}

type linuxPackageOptions struct {
	packageType            string
	binPath                string
	configDir              string
	configFilePath         string
	etcDefaultPath         string
	etcDefaultFilePath     string
	systemdServiceFilePath string

	systemdFileSrc string
	postinstSrc    string
	defaultFileSrc string
}

func runPrint(cmd string, args ...string) {
	log.Println(cmd, strings.Join(args, " "))
	ecmd := exec.Command(cmd, args...)
	ecmd.Stdout = os.Stdout
	ecmd.Stderr = os.Stderr
	err := ecmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func ensureGoPath() {
	if os.Getenv("GOPATH") == "" {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		gopath := filepath.Clean(filepath.Join(cwd, "../../../../"))
		log.Println("GOPATH is", gopath)
		os.Setenv("GOPATH", gopath)
	}
}

func build(pkg string, tags []string) {
	binary := "./tmp/bin/" + serverBinaryName

	rmr(binary, binary+".md5")
	args := []string{"build", "-ldflags", ldflags()}
	if len(tags) > 0 {
		args = append(args, "-tags", strings.Join(tags, ","))
	}
	if race {
		args = append(args, "-race")
	}

	args = append(args, "-o", binary)
	args = append(args, pkg)
	setBuildEnv()

	runPrint("go", "version")
	runPrint("go", args...)

	// Create an md5 checksum of the binary, to be included in the archive for
	// automatic upgrades.
	err := md5File(binary)
	if err != nil {
		log.Fatal(err)
	}
}

func setBuildEnv() {
	os.Setenv("GOOS", goos)
	if strings.HasPrefix(goarch, "armv") {
		os.Setenv("GOARCH", "arm")
		os.Setenv("GOARM", goarch[4:])
	} else {
		os.Setenv("GOARCH", goarch)
	}
	if goarch == "386" {
		os.Setenv("GO386", "387")
	}
	wd, err := os.Getwd()
	if err != nil {
		log.Println("Warning: can't determine current dir:", err)
		log.Println("Build might not work as expected")
	}
	os.Setenv("GOPATH", fmt.Sprintf("%s%c%s", filepath.Join(wd, "Godeps", "_workspace"), os.PathListSeparator, os.Getenv("GOPATH")))
	log.Println("GOPATH=" + os.Getenv("GOPATH"))
}

func md5File(file string) error {
	fd, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fd.Close()

	h := md5.New()
	_, err = io.Copy(h, fd)
	if err != nil {
		return err
	}

	out, err := os.Create(file + ".md5")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "%x\n", h.Sum(nil))
	if err != nil {
		return err
	}

	return out.Close()
}

func ldflags() string {
	var b bytes.Buffer
	b.WriteString("-w")
	b.WriteString(fmt.Sprintf(" -X main.version=%s", version))

	return b.String()
}

func rmr(paths ...string) {
	for _, path := range paths {
		log.Println("rm -r", path)
		os.RemoveAll(path)
	}
}

func clean() {
	rmr("bin", "Godeps/_workspace/pkg", "Godeps/_workspace/bin")
	rmr("dist")
	rmr("tmp")
	rmr(filepath.Join(os.Getenv("GOPATH"), fmt.Sprintf("pkg/%s_%s/github.com/krise3k", goos, goarch)))
}
