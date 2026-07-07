package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/tianyi-zhang-02/coact/internal/versionmgr"
)

func cmdVersions(args []string) int {
	fs := flag.NewFlagSet("versions", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	home := fs.String("home", versionmgr.DefaultHome(), "managed coact home")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "coact versions: unexpected argument %q\n", fs.Arg(0))
		return 2
	}

	versions := versionmgr.LocalVersions(*home)
	if len(versions) == 0 {
		fmt.Printf("no managed coact versions in %s\n", *home)
		return 0
	}
	fmt.Printf("managed coact versions in %s:\n", *home)
	for _, v := range versions {
		marker := " "
		if v.Active {
			marker = "*"
		}
		fmt.Printf("  %s %-16s %s\n", marker, v.Version, v.Path)
	}
	return 0
}

func cmdUpdate(args []string) int {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	channel := fs.String("channel", "stable", "release channel: stable, beta, or alpha")
	repo := fs.String("repo", "tianyi-zhang-02/coact", "GitHub owner/repo")
	home := fs.String("home", versionmgr.DefaultHome(), "managed coact home")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "coact update: unexpected argument %q\n", fs.Arg(0))
		return 2
	}

	// Experimental: releases are verified by SHA-256 checksum over HTTPS, but not
	// yet by a cryptographic signature. Until signing lands, treat self-update as
	// a convenience, not an authenticity guarantee.
	fmt.Fprintln(os.Stderr, "coact update is experimental (checksum-verified over HTTPS, not yet signed).")

	manifest, installed, err := versionmgr.Update(versionmgr.UpdateOptions{
		Repo:    *repo,
		Channel: *channel,
		Home:    *home,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact update: %v\n", err)
		return 1
	}

	fmt.Printf("downloaded coact %s to %s\n", manifest.Version, installed)
	if managed, exe := versionmgr.CurrentBinaryManaged(*home); managed {
		if err := versionmgr.Switch(*home, manifest.Version); err != nil {
			fmt.Fprintf(os.Stderr, "coact update: installed but could not switch: %v\n", err)
			return 1
		}
		fmt.Printf("switched %s to %s\n", versionmgr.LinkPath(*home), manifest.Version)
	} else {
		fmt.Printf("current binary is not managed by coact: %s\n", exe)
		fmt.Printf("to use the managed shim, add %s to PATH and run: coact switch %s\n", *home, manifest.Version)
	}
	return 0
}

func cmdSwitch(args []string) int {
	fs := flag.NewFlagSet("switch", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	home := fs.String("home", versionmgr.DefaultHome(), "managed coact home")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: coact switch [--home <dir>] <version>")
		return 2
	}
	version := fs.Arg(0)
	if err := versionmgr.Switch(*home, version); err != nil {
		fmt.Fprintf(os.Stderr, "coact switch: %v\n", err)
		return 1
	}
	fmt.Printf("switched %s to %s\n", versionmgr.LinkPath(*home), version)
	return 0
}
