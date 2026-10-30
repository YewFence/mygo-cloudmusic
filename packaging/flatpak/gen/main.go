// Command gen prepares packaging/flatpak/_build for a flatpak-builder
// run: it fills build.sh's @NAME@/@VERSION@ placeholders from mygo.json
// and gives the metainfo its <releases> block, from CHANGELOG.md with
// each version dated by its tag. Nothing it writes is committed; the
// changelog a release already writes is all the input there is.
//
//	go run ./packaging/flatpak/gen packaging/flatpak
package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// out is the prepared directory the manifest builds from.
var out string

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: gen DIR")
		os.Exit(1)
	}
	dir := os.Args[1]
	out = filepath.Join(dir, "_build")
	if err := os.RemoveAll(out); err != nil {
		fail(err)
	}

	name, version := mygoJSON()
	write(filepath.Join(dir, "build.sh"), "build.sh", map[string]string{
		"@NAME@":    name,
		"@VERSION@": version,
	})
	sections := changelogReleases()
	write(filepath.Join(dir, "dev.egoist.cloudmusic.desktop"),
		"dev.egoist.cloudmusic.desktop", nil)
	write(filepath.Join(dir, "dev.egoist.cloudmusic.metainfo.xml"),
		"dev.egoist.cloudmusic.metainfo.xml", nil)
	addReleases(filepath.Join(out, "dev.egoist.cloudmusic.metainfo.xml"), sections)
}

// mygoJSON returns the app's name and version, which the flatpak's
// binary reports as its own.
func mygoJSON() (name, version string) {
	text, err := os.ReadFile("mygo.json")
	if err != nil {
		fail(err)
	}
	var config struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(text, &config); err != nil {
		fail(err)
	}
	return config.Name, config.Version
}

// A section of the changelog: the version of its heading, and the lines
// under it.
type section struct {
	version string
	date    string
	lines   []string
}

// changelogReleases takes the "## x.y.z" sections of the changelog, in
// order, dating each by its tag: an undated version would end up in
// appstream data with a made-up date.
func changelogReleases() []section {
	text, err := os.ReadFile("CHANGELOG.md")
	if err != nil {
		fail(err)
	}
	var sections []section
	var cur *section
	for _, line := range strings.Split(string(text), "\n") {
		if v, ok := strings.CutPrefix(line, "## "); ok {
			date, err := exec.Command("git", "log", "-1", "--format=%as", "v"+v).Output()
			if err != nil {
				fail(fmt.Errorf("tag v%s: %w", v, err))
			}
			sections = append(sections, section{version: v, date: strings.TrimSpace(string(date))})
			cur = &sections[len(sections)-1]
			continue
		}
		if cur != nil && line != "" {
			cur.lines = append(cur.lines, strings.TrimPrefix(line, "- "))
		}
	}
	if len(sections) == 0 {
		fail(fmt.Errorf("no \"## version\" sections in CHANGELOG.md"))
	}
	return sections
}

// write copies src into _build as name, @NAME@/@VERSION@ filled in when
// any: a placeholder left unfilled would end up in the package.
func write(src, name string, fill map[string]string) {
	text, err := os.ReadFile(src)
	if err != nil {
		fail(err)
	}
	s := string(text)
	for ph, v := range fill {
		if !strings.Contains(s, ph) {
			fail(fmt.Errorf("%s has no %s placeholder", src, ph))
		}
		s = strings.ReplaceAll(s, ph, v)
	}
	dst := filepath.Join(out, name)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		fail(err)
	}
	mode := os.FileMode(0o644)
	if strings.HasSuffix(name, ".sh") {
		mode = 0o755
	}
	if err := os.WriteFile(dst, []byte(s), mode); err != nil {
		fail(err)
	}
}

// metainfoRelease is one <release> of the metainfo's <releases> block,
// marshalled with encoding/xml: the changelog lines are character data
// of <li>, which the encoder escapes.
type metainfoRelease struct {
	XMLName xml.Name `xml:"release"`
	Version string   `xml:"version,attr"`
	Date    string   `xml:"date,attr"`
	Items   []string `xml:"description>ul>li"`
}

// addReleases inserts the changelog's sections as the <releases> block
// of the prepared metainfo, their lines kept verbatim: the changelog is
// written for the app's users, and so is this. Software centers list a
// handful of releases; the rest is noise.
func addReleases(path string, sections []section) {
	releases := make([]metainfoRelease, 0, len(sections))
	for i, s := range sections {
		if i == 8 {
			break
		}
		releases = append(releases, metainfoRelease{Version: s.version, Date: s.date, Items: s.lines})
	}
	// The block sits inside the metainfo document, so the lines it is
	// indented by are written here; the encoder indents what's between.
	var block bytes.Buffer
	block.WriteString("  <releases>\n")
	enc := xml.NewEncoder(&block)
	enc.Indent("    ", "  ")
	if err := enc.Encode(releases); err != nil {
		fail(err)
	}
	if err := enc.Close(); err != nil {
		fail(err)
	}
	block.WriteString("\n  </releases>")

	text, err := os.ReadFile(path)
	if err != nil {
		fail(err)
	}
	const open, close = "  <releases>", "  </releases>"
	i, j := strings.Index(string(text), open), strings.Index(string(text), close)
	if i < 0 || j < i {
		fail(fmt.Errorf("%s has no <releases> block", path))
	}
	out := string(text)[:i] + block.String() + string(text)[j+len(close):]
	if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "gen:", err)
	os.Exit(1)
}
