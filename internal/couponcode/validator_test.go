package couponcode_test

import (
	"compress/gzip"
	"context"
	"os"
	"testing"

	"github.com/malakagl/kart-challenge/internal/couponcode"
	"github.com/malakagl/kart-challenge/pkg/util"
)

// helper: create a plain text file with given lines
func createTempFile(t *testing.T, lines []string) string {
	t.Helper()
	file, err := os.CreateTemp("", "coupon_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	for _, line := range lines {
		_, err := file.WriteString(line + "\n")
		if err != nil {
			t.Fatal(err)
		}
	}
	return file.Name()
}

// helper: create a gzipped file with given lines
func createTempGzipFile(t *testing.T, lines []string) string {
	t.Helper()
	file, err := os.CreateTemp("", "coupon_*.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	for _, line := range lines {
		_, err := gw.Write([]byte(line + "\n"))
		if err != nil {
			t.Fatal(err)
		}
	}
	return file.Name()
}

func TestValidateCouponCode_PlainText(t *testing.T) {
	// create files
	file1 := createTempFile(t, []string{"ABC12345", "XYZ98765"})
	defer os.Remove(file1)

	file2 := createTempFile(t, []string{"ABC12345", "LMN11111"})
	defer os.Remove(file2)

	file3 := createTempFile(t, []string{"QWE22222"})
	defer os.Remove(file3)

	couponcode.SetCouponCodeFiles([]string{file1, file2, file3})
	couponcode.InitCache(10)
	// code present in 2 files → should return true
	if isValid, err := couponcode.ValidateCouponCode(t.Context(), "ABC12345"); !isValid {
		t.Error("Expected true and no error, got ", isValid, err)
	}

	// code present in 1 file → should return false
	if isValid, err := couponcode.ValidateCouponCode(t.Context(), "QWE22222"); isValid {
		t.Error("Expected false, got ", isValid, err)
	}

	// code not present → should return false
	if isValid, err := couponcode.ValidateCouponCode(t.Context(), "NOTFOUND"); isValid {
		t.Error("Expected false, got ", isValid, err)
	}

	// code too short → should return false
	if isValid, err := couponcode.ValidateCouponCode(t.Context(), "SHORT"); isValid {
		t.Error("Expected false for short code, got ", isValid, err)
	}
}

func TestValidateCouponCode_GzipFiles(t *testing.T) {
	file1 := createTempGzipFile(t, []string{"ABC12345", "XYZ98765"})
	defer os.Remove(file1)

	file2 := createTempGzipFile(t, []string{"ABC12345", "LMN11111"})
	defer os.Remove(file2)

	couponcode.SetCouponCodeFiles([]string{file1, file2})

	// code present in 2 files → should return true
	if isValid, err := couponcode.ValidateCouponCode(t.Context(), "ABC12345"); !isValid {
		t.Error("Expected true and no error, got ", isValid, err)
	}

	// code present in 1 file → should return false
	if isValid, err := couponcode.ValidateCouponCode(t.Context(), "LMN11111"); isValid {
		t.Error("Expected false, got ", isValid, err)
	}
}

func TestValidateCouponCode_MixedFiles(t *testing.T) {
	file1 := createTempFile(t, []string{"ABC12345", "XYZ98765"})
	defer os.Remove(file1)

	file2 := createTempGzipFile(t, []string{"ABC12345", "LMN11111"})
	defer os.Remove(file2)

	couponcode.SetCouponCodeFiles([]string{file1, file2})

	// code present in both → should return true
	if isValid, err := couponcode.ValidateCouponCode(t.Context(), "ABC12345"); !isValid {
		t.Error("Expected true, got ", isValid, err)
	}

	// code present in only one → should return false
	if isValid, err := couponcode.ValidateCouponCode(t.Context(), "LMN11111"); isValid {
		t.Error("Expected false, got ", isValid, err)
	}
}

// Benchmark tests
func BenchmarkValidateCouponCode_ValidCodeWithDecompressedFiles(b *testing.B) {
	validCode := "FIFTYOFF"
	files := []string{
		util.AbsoluteFilePath("couponbase1.txt", "../.."),
		util.AbsoluteFilePath("couponbase2.txt", "../.."),
		util.AbsoluteFilePath("couponbase3.txt", "../.."),
	}
	couponcode.SetCouponCodeFiles(files)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = couponcode.ValidateCouponCode(ctx, validCode)
	}
}

func BenchmarkValidateCouponCode_InvalidCodeUsingDecompressedFiles(b *testing.B) {
	files := []string{
		util.AbsoluteFilePath("couponbase1.txt", "../.."),
		util.AbsoluteFilePath("couponbase2.txt", "../.."),
		util.AbsoluteFilePath("couponbase3.txt", "../.."),
	}
	couponcode.SetCouponCodeFiles(files)
	invalidCode := "NOTFOUND"

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = couponcode.ValidateCouponCode(ctx, invalidCode)
	}
}

func BenchmarkValidateCouponCode_ValidCodeWithCompressedFiles(b *testing.B) {
	validCode := "FIFTYOFF"
	files := []string{
		util.AbsoluteFilePath("couponbase1.gz", "../.."),
		util.AbsoluteFilePath("couponbase2.gz", "../.."),
		util.AbsoluteFilePath("couponbase3.gz", "../.."),
	}
	couponcode.SetCouponCodeFiles(files)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = couponcode.ValidateCouponCode(ctx, validCode)
	}
}

func BenchmarkValidateCouponCode_InvalidCodeUsingCompressedFiles(b *testing.B) {
	files := []string{
		util.AbsoluteFilePath("couponbase1.gz", "../.."),
		util.AbsoluteFilePath("couponbase2.gz", "../.."),
		util.AbsoluteFilePath("couponbase3.gz", "../.."),
	}
	couponcode.SetCouponCodeFiles(files)
	invalidCode := "NOTFOUND"

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = couponcode.ValidateCouponCode(ctx, invalidCode)
	}
}
