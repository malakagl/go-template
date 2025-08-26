package couponcode

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/malakagl/kart-challenge/pkg/log"
	"github.com/malakagl/kart-challenge/pkg/otel"
)

// SetupCouponCodeFiles processes a list of gzip files, unzipping each one into a corresponding text file.
// experimental
func SetupCouponCodeFiles(ctx context.Context, filePaths []string) error {
	defer func(start time.Time) {
		log.Info().Msgf("Coupon code files setup completed in %s", time.Since(start).String())
	}(time.Now())

	var wg sync.WaitGroup
	var errCh = make(chan error, len(filePaths))
	for _, filePath := range filePaths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			spanCtx, span := otel.Tracer(ctx, filepath.Base(path))
			defer span.End()
			if err := UnZipGzipFile(spanCtx, path); err != nil {
				log.Error().Msgf("error unzipping file %s: %v", path, err)
				span.RecordError(err)
				errCh <- err
			}
		}(filePath)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			log.Error().Msgf("error processing file: %v", err)
			return err
		}
	}

	return nil
}

func UnZipGzipFile(ctx context.Context, input string) error {
	if !strings.HasSuffix(input, ".gz") {
		return fmt.Errorf("input file must end with .gz")
	}

	output := input[:len(input)-3] + ".txt" // remove .gz extension
	// Open gzip file
	f, err := os.Open(input)
	if err != nil {
		log.Error().Msgf("failed to open gzip file %s: %v", input, err)
		return err
	}
	defer f.Close()

	// Create gzip reader
	gr, err := gzip.NewReader(f)
	if err != nil {
		log.Error().Msgf("failed to create gzip reader for %s: %v", input, err)
		return err
	}
	defer gr.Close()

	// Create output file
	out, err := os.Create(output)
	if err != nil {
		log.Error().Msgf("failed to create output file %s: %v", output, err)
		return err
	}
	defer out.Close()

	// Copy decompressed data
	_, err = io.Copy(out, gr)
	if err != nil {
		log.Error().Msgf("failed to copy data from %s to %s: %v", input, output, err)
		return err
	}

	log.Info().Msgf("unzipped file %s to %s", input, output)
	rwMutex.Lock()
	for i, file := range couponCodeFiles {
		if input == file {
			couponCodeFiles[i] = output
			break
		}
	}
	rwMutex.Unlock()

	return nil
}
