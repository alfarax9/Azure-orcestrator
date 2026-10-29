package server

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Source says where the binaries come from: files on this machine, or a release the server
// downloads itself. Exactly one is set.
type Source struct {
	// BinaryDir holds astrad-linux-<arch>, astra-priv-linux-<arch>, optionally SHA256SUMS and a
	// web/ folder (the built web UI).
	BinaryDir string
	// ReleaseURL is a folder holding SHA256SUMS and astra-linux-<arch>.tar.gz.
	ReleaseURL string
	// AllowHTTP lets ReleaseURL be plain http, for a release mirror on a private network.
	AllowHTTP bool
}

// urlPattern is the only shape of release URL accepted: it goes into a script, so it has no
// quotes, spaces, or shell characters.
var urlPattern = regexp.MustCompile(`^(https?)://[A-Za-z0-9.-]+(:[0-9]{1,5})?(/[A-Za-z0-9._~/-]*)?$`)

func (s Source) validate() error {
	switch {
	case (s.BinaryDir == "") == (s.ReleaseURL == ""):
		return errors.New("give exactly one of a folder with the binaries and a release URL")
	case s.ReleaseURL == "":
		return nil
	case !urlPattern.MatchString(s.ReleaseURL):
		return fmt.Errorf("release URL %q: use plain letters, digits, dots and slashes (no query, no spaces)", s.ReleaseURL)
	case strings.HasPrefix(s.ReleaseURL, "http://") && !s.AllowHTTP:
		return errors.New("a release URL must be https; an http mirror needs an explicit opt-in")
	}
	return nil
}

// localFiles are the binaries of a BinaryDir for one architecture.
type localFiles struct {
	astrad, priv string
	web          string // empty when there is no web UI
	sums         string // SHA256SUMS content for the files as uploaded, named astrad and astra-priv
	provided     bool   // the sums came from a SHA256SUMS file, not from hashing here
}

// inspect finds the binaries for arch, and checks them against a SHA256SUMS file if there is one.
func inspect(dir, arch string) (localFiles, error) {
	f := localFiles{astrad: filepath.Join(dir, "astrad-linux-"+arch), priv: filepath.Join(dir, "astra-priv-linux-"+arch)}
	hashes := map[string]string{}
	for _, p := range []string{f.astrad, f.priv} {
		sum, err := fileSum(p)
		if err != nil {
			return f, fmt.Errorf("%w (build it with GOOS=linux GOARCH=%s)", err, arch)
		}
		hashes[filepath.Base(p)] = sum
	}
	if listed, err := readSums(filepath.Join(dir, "SHA256SUMS")); err == nil {
		for name, sum := range hashes {
			if want, ok := listed[name]; !ok || want != sum {
				return f, fmt.Errorf("%s does not match SHA256SUMS in %s: the file or the list was changed", name, dir)
			}
		}
		f.provided = true
	} else if !errors.Is(err, fs.ErrNotExist) {
		return f, err
	}
	f.sums = fmt.Sprintf("%s  astrad\n%s  astra-priv\n", hashes[filepath.Base(f.astrad)], hashes[filepath.Base(f.priv)])
	if info, err := os.Stat(filepath.Join(dir, "web", "index.html")); err == nil && !info.IsDir() {
		f.web = filepath.Join(dir, "web")
	}
	return f, nil
}

func fileSum(path string) (string, error) {
	fh, err := os.Open(path) //nolint:gosec // the binary the person pointed at
	if err != nil {
		return "", fmt.Errorf("find %s: %w", path, err)
	}
	defer func() { _ = fh.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, fh); err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// readSums reads a SHA256SUMS file: "<hash>  <name>" per line.
func readSums(path string) (map[string]string, error) {
	fh, err := os.Open(path) //nolint:gosec // the list next to the binaries
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	defer func() { _ = fh.Close() }()
	out := map[string]string{}
	sc := bufio.NewScanner(fh)
	for sc.Scan() {
		if sum, name, ok := strings.Cut(sc.Text(), "  "); ok {
			out[strings.TrimPrefix(name, "*")] = strings.ToLower(sum)
		}
	}
	return out, sc.Err()
}

// tarWeb packs the built web UI as a gzipped tar, for sending over SSH.
func tarWeb(dir string) (io.Reader, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || (!d.IsDir() && !d.Type().IsRegular()) { // links and devices are not part of a web build
			return err
		}
		return addEntry(tw, dir, path, d)
	})
	if err == nil {
		err = errors.Join(tw.Close(), gz.Close())
	}
	if err != nil {
		return nil, fmt.Errorf("pack the web UI: %w", err)
	}
	return &buf, nil
}

// addEntry writes one file or folder of the web build, owned by nobody in particular.
func addEntry(tw *tar.Writer, root, path string, d fs.DirEntry) error {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." {
		return err
	}
	info, err := d.Info()
	if err != nil {
		return err
	}
	hdr, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	hdr.Name = filepath.ToSlash(rel)
	hdr.Uid, hdr.Gid, hdr.Uname, hdr.Gname = 0, 0, "", ""
	if err := tw.WriteHeader(hdr); err != nil || d.IsDir() {
		return err
	}
	fh, err := os.Open(path) //nolint:gosec // a file inside the web build folder
	if err != nil {
		return err
	}
	defer func() { _ = fh.Close() }()
	_, err = io.Copy(tw, fh)
	return err
}
