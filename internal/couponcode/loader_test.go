package couponcode

import (
	"compress/gzip"
	"os"
	"testing"
)

func TestUnZipGzipFile(t *testing.T) {
	f, err := os.CreateTemp("", "testfile-*.gz")
	if err != nil {
		t.Fatalf("failed to create test gzip file: %v", err)
	}
	defer f.Close()
	gw := gzip.NewWriter(f)
	_, err = gw.Write([]byte("TESTCODE\nHAPPYHRS\n"))
	if err != nil {
		t.Fatalf("failed to write gzip content: %v", err)
	}
	gw.Close()

	err = UnZipGzipFile(t.Context(), f.Name())
	if err != nil {
		t.Errorf("UnZipGzipFile(%q) failed: %v", f.Name(), err)
	}
	outputFile := f.Name()[:len(f.Name())-3] + ".txt" // remove .gz extension
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("expected output file %q to exist, but it does not", outputFile)
	} else {
		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Errorf("failed to read output file %q: %v", outputFile, err)
		} else if string(content) != "TESTCODE\nHAPPYHRS\n" {
			t.Errorf("expected content of %q to be 'TESTCODE\\nHAPPYHRS\\n', got %q", outputFile, content)
		}
		os.Remove(outputFile) // clean up
	}
}
