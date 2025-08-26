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

	"github.com/malakagl/kart-challenge/pkg/cache"
	"github.com/malakagl/kart-challenge/pkg/errors"
	"github.com/malakagl/kart-challenge/pkg/log"
	"github.com/malakagl/kart-challenge/pkg/otel"
)

var (
	couponCodeFiles []string
	rwMutex         sync.RWMutex
	couponCodeCache *cache.LRUCache
)

func InitCache(maxSize int) {
	couponCodeCache = cache.NewLRUCache(maxSize)
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

	f, err := os.Open(path)
	if err != nil {
		errCh <- fmt.Errorf("couponcode: couponcode file open error: %w", err)
		return
	}
	defer func() { _ = f.Close() }()

	var reader io.Reader = f
	if strings.HasSuffix(strings.ToLower(filepath.Ext(path)), ".gz") {
		gz, err := gzip.NewReader(f)
		if err != nil {
			errCh <- fmt.Errorf("error creating gzip reader: %v", err)
			return
		}
		defer func() { _ = gz.Close() }()
		reader = gz
	}

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
		couponCodeCache.Set(code, true)
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

	couponCodeCache.Set(code, false)
	return false, nil
}
