package update

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"fmt"
	"io"
	"path"
	"strings"
)

const maxBinarySize = 256 << 20

// ParseChecksums parses a sha256sum-style file into filename -> lowercase hash.
func ParseChecksums(data []byte) (map[string]string, error) {
	sums := make(map[string]string)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, fmt.Errorf("invalid checksum line %q", line)
		}
		hash := strings.ToLower(fields[0])
		if len(hash) != 64 {
			return nil, fmt.Errorf("invalid checksum line %q: hash is not 64 hex characters", line)
		}
		if _, err := hex.DecodeString(hash); err != nil {
			return nil, fmt.Errorf("invalid checksum line %q: %w", line, err)
		}
		name := strings.TrimPrefix(fields[1], "*")
		name = strings.TrimPrefix(name, "./")
		if name == "" {
			return nil, fmt.Errorf("invalid checksum line %q: empty filename", line)
		}
		sums[name] = hash
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read checksums: %w", err)
	}
	if len(sums) == 0 {
		return nil, fmt.Errorf("checksum file is empty")
	}
	return sums, nil
}

// extractBinary returns the named executable from a gzipped tarball.
func extractBinary(archive []byte, binary string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, fmt.Errorf("open release archive: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	wantExe := binary + ".exe"
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read release archive: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		clean := path.Clean(hdr.Name)
		if clean == ".." || strings.HasPrefix(clean, "../") || path.IsAbs(clean) {
			base := path.Base(clean)
			if base == binary || base == wantExe {
				return nil, fmt.Errorf("release archive contains unsafe path %q", hdr.Name)
			}
			continue
		}
		base := path.Base(clean)
		if base != binary && base != wantExe {
			continue
		}
		if hdr.Size < 0 || hdr.Size > maxBinarySize {
			return nil, fmt.Errorf("release binary is too large (%d bytes)", hdr.Size)
		}
		data, err := io.ReadAll(io.LimitReader(tr, maxBinarySize+1))
		if err != nil {
			return nil, fmt.Errorf("read release binary: %w", err)
		}
		if len(data) == 0 {
			return nil, fmt.Errorf("release binary is empty")
		}
		if int64(len(data)) > maxBinarySize {
			return nil, fmt.Errorf("release binary is too large")
		}
		return data, nil
	}
	return nil, fmt.Errorf("release archive does not contain %s", binary)
}
