package couponcode

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/malakagl/go-template/pkg/cache"
	"github.com/malakagl/go-template/pkg/errors"
	"github.com/malakagl/go-template/pkg/log"
	"github.com/malakagl/go-template/pkg/otel"
)

var (
	couponCodeFiles []string
	rwMutex         sync.RWMutex
	couponCodeCache *cache.LRUCache[bool]
)

func InitCache(maxSize int) {
	couponCodeCache = cache.NewLRUCache[bool](maxSize, time.Hour)
}

func SetCouponCodeFiles(f []string) {
	rwMutex.Lock()
	defer rwMutex.Unlock()
	couponCodeFiles = f
}

func worker(ctx context.Context, path, code string, count *atomic.Int32, wg *sync.WaitGroup, cancel context.CancelFunc, errCh chan error) {
	defer wg.Done()
	ctx, span := otel.Tracer(ctx, "worker:"+path+":"+code)
	defer span.End()

	if strings.HasSuffix(strings.ToLower(filepath.Ext(path)), ".gz") {
		f, err := os.Open(path)
		if err != nil {
			errCh <- fmt.Errorf("couponcode: couponcode file open error: %w", err)
			return
		}
		defer func() { _ = f.Close() }()
		reader, err := gzip.NewReader(f)
		if err != nil {
			errCh <- fmt.Errorf("error creating gzip reader: %v", err)
			return
		}
		defer func() { _ = reader.Close() }()

		scanner := bufio.NewScanner(reader)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				log.WithCtx(ctx).Debug().Msgf("Context done: %v", path)
				return
			default:
				if strings.TrimSpace(scanner.Text()) == code {
					if count.Add(1) >= 2 { // found in 2 files
						cancel() // stop all other workers
						return
					}

					return
				}
			}
		}
	} else if strings.HasSuffix(strings.ToLower(filepath.Ext(path)), ".txt") {
		f, err := os.Open(path)
		if err != nil {
			errCh <- fmt.Errorf("couponcode: couponcode file open error: %w", err)
			return
		}
		defer func() { _ = f.Close() }()
		findCodeInTextFile(ctx, code, count, f, errCh, cancel)
	}
}

type fileChunk struct {
	start int64
	end   int64
}

func findCodeInTextFile(ctx context.Context, code string, count *atomic.Int32, f *os.File, errCh chan error, cancel context.CancelFunc) {
	stat, err := f.Stat()
	if err != nil {
		errCh <- fmt.Errorf("could not stat file %s: %w", f.Name(), err)
		return
	}

	fileSize := stat.Size()
	const chunkSize = int64(1024 * 1024 * 100) // 100MB
	const overlap = 10                         // to avoid cutting off lines. coupon code is less than or equal 10 bytes
	numChunks := int((fileSize / chunkSize) + 1)

	// prepare channel of chunks
	chunkCh := make(chan fileChunk, numChunks)
	for i := 0; i < numChunks; i++ {
		start := int64(i) * chunkSize
		end := start + chunkSize
		if i == numChunks-1 {
			end = fileSize
		} else {
			end += overlap
		}
		chunkCh <- fileChunk{start, end}
	}
	close(chunkCh)

	// limit workers
	const numWorkers = 100
	var wg sync.WaitGroup
	var foundInThisFile atomic.Bool
	workerFn := func() {
		defer wg.Done()
		for chunk := range chunkCh {
			log.WithCtx(ctx).Debug().Msgf("starting worker for file %s for chunk %d-%d", f.Name(), chunk.start, chunk.end)
			// if already found in this file, stop early
			if foundInThisFile.Load() {
				return
			}

			section := io.NewSectionReader(f, chunk.start, chunk.end-chunk.start)
			scanner := bufio.NewScanner(section)
			buf := make([]byte, 0, 64*1024)
			scanner.Buffer(buf, 1024*1024)

			for scanner.Scan() {
				select {
				case <-ctx.Done():
					return
				default:
					if strings.TrimSpace(scanner.Text()) == code {
						if !foundInThisFile.Swap(true) { // only once per file
							if count.Add(1) >= 2 { // found in 2+ files
								cancel()
							}
						}
						return
					}
				}
			}
			if err := scanner.Err(); err != nil {
				errCh <- fmt.Errorf("scanner error in %s [%d-%d]: %w", f.Name(), chunk.start, chunk.end, err)
			}
		}
	}

	// spawn fixed number of workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go workerFn()
	}

	wg.Wait()
}

func ValidateCouponCode(ctx context.Context, code string) (bool, error) {
	ctx, span := otel.Tracer(ctx, "validateCouponCode")
	defer span.End()
	log.WithCtx(ctx).Debug().Msgf("validating coupon code %s", code)
	isValid := false
	defer func(start time.Time) {
		log.WithCtx(ctx).Debug().Msgf("validated coupon code in %s: %v", time.Since(start).String(), isValid)
	}(time.Now())

	if len(code) < 8 || len(code) > 10 {
		log.WithCtx(ctx).Warn().Msgf("invalid coupon code length. code: %s", code)
		return false, nil
	}

	if value, found := couponCodeCache.Get(code); found {
		log.WithCtx(ctx).Debug().Msgf("found coupon code in cache %s", code)
		isValid = value
		return value, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	var count atomic.Int32
	errChan := make(chan error, len(couponCodeFiles))
	for _, f := range couponCodeFiles {
		log.WithCtx(ctx).Debug().Msgf("checking file %s", f)
		wg.Add(1)
		go worker(ctx, f, code, &count, &wg, cancel, errChan)
	}

	wg.Wait()
	close(errChan)
	if count.Load() >= 2 {
		couponCodeCache.Put(code, true)
		isValid = true
		return isValid, nil
	}

	var err error
	for e := range errChan {
		if e != nil {
			err = errors.Join(err, e)
		}
	}
	if err != nil {
		return false, err
	}

	couponCodeCache.Put(code, false)
	return false, nil
}
