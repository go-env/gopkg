// Command `get` for download and install packages
// and their dependencies
package main

import "github.com/go-env/pkglib"

var cmdGet = &pkglib.Command{
	UsageLine: "get [-r <repo>] [-e <env>] [-t <tag>] [-d] <packages...>",
	Short:     "download and install packages with their dependencies",
	Long: `
Get downloads packages named by the import paths and installs them with
the dependencies to registered environments. Dependencies parsed from
packages source code.

Packages switched to tags as pointed in args. If any of packages have
GOENV file, those they are used for resolving dependencies and their
versions. Command line args have higher priority.

Run 'go help packages' — more info about specifying packages
`,
	CustomFlags: true,
}

var getE = cmdGet.Flag.String("e", "default", "")
var getD = cmdGet.Flag.Bool("d", false, "")
var getR = cmdGet.Flag.String("r", "", "")
var getT = cmdGet.Flag.String("t", "", "")

func init() {
	//addBuildFlags(cmdGet)
	cmdGet.Run = runGet // break init loop
}

//
func runGet(cmd *pkglib.Command, args []string) {
	// just experiment
	// git := pkglib.VcsByCmd("git")
	// git.Create("XXX", "https://github.com/grafov/m3u8.git")
	// git.TagSync("XXX", "draft")

	// Phase 1.  Download/update.
	var stk pkglib.ImportStack
	for _, arg := range pkglib.DownloadPaths(args) {
		pkglib.Download(arg, &stk, false)
	}
	exitIfErrors()

	// Phase 2. Rescan packages and re-evaluate args list.

	// Code we downloaded and all code that depends on it
	// needs to be evicted from the package cache so that
	// the information will be recomputed.  Instead of keeping
	// track of the reverse dependency information, evict
	// everything.
	pkglib.CleanPackageCache()

	args = pkglib.ImportPaths(args)

	// Phase 3.  Install.
	if *getD {
		// Download only.
		// Check delayed until now so that importPaths
		// has a chance to print errors.
		return
	}

	pkglib.RunInstall(cmd, args)
}
