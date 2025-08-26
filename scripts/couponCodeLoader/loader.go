package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var couponFiles = []string{
	"couponbase1.gz",
	"couponbase2.gz",
}

func main() {
	LoadCouponCodes()
}

func dataFilePath(name string) string {
	_, thisFile, _, _ := runtime.Caller(0) // 0 = this function
	baseDir := filepath.Join(filepath.Dir(thisFile), "../../promocodes/")
	return filepath.Join(baseDir, name)
}

func LoadCouponCodes() {
	user := "test_user"
	pass := "test_password"
	host := "localhost"
	port := "5432"
	name := "test"

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, pass, host, port, name,
	)
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	files := couponFiles
	concurrency := 2 // number of concurrent files
	batchSize := 1000000

	filePaths := make(chan string, len(files))

	var wg sync.WaitGroup

	// Start worker pool
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range filePaths {
				fmt.Printf("Processing file: %s\n", path)
				fileId, err := createFileRecord(ctx, pool, path)
				if err != nil {
					log.Fatalf("Failed to create file record for %s: %v", path, err)
				}
				if err := copyFromGzipBatched(ctx, pool, fileId, path, batchSize); err != nil {
					log.Fatalf("Failed to copy file %s: %v", path, err)
				}
				fmt.Printf("Finished file: %s\n", path)
			}
		}()
	}

	// Send filePaths
	for _, path := range files {
		filePaths <- path
	}
	close(filePaths)

	wg.Wait()
	fmt.Println("All files processed successfully.")
}

func createFileRecord(ctx context.Context, pool *pgxpool.Pool, fileName string) (int, error) {
	_, err := pool.Exec(ctx, "INSERT INTO files (file_name) VALUES ($1)", fileName)
	if err != nil {
		return 0, fmt.Errorf("failed to insert file record for %s: %w", fileName, err)
	}

	var fileID int
	err = pool.QueryRow(ctx, "SELECT id FROM files WHERE file_name = $1", fileName).Scan(&fileID)
	if err != nil {
		return 0, fmt.Errorf("failed to get file ID for %s: %w", fileName, err)
	}

	log.Printf("Created file record for %s with ID %d", fileName, fileID)
	return fileID, nil
}

func copyFromGzipBatched(ctx context.Context, pool *pgxpool.Pool, fileID int, path string, batchSize int) error {
	path = dataFilePath(path)
	log.Printf("Copying code from file %s with file id %d", path, fileID)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	scanner := bufio.NewScanner(gzReader)
	batch := make([][]interface{}, 0, batchSize)

	for scanner.Scan() {
		batch = append(batch, []interface{}{fileID, scanner.Text()})

		if len(batch) >= batchSize {
			if err := copyBatch(ctx, pool, batch); err != nil {
				return err
			}
			batch = batch[:0] // reset batch
		}
	}

	// Insert remaining rows
	if len(batch) > 0 {
		if err := copyBatch(ctx, pool, batch); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func copyBatch(ctx context.Context, pool *pgxpool.Pool, batch [][]interface{}) error {
	log.Println("inserting a batch")
	rows := &batchCopySource{batch: batch, idx: -1}
	_, err := pool.CopyFrom(ctx, pgx.Identifier{"coupon_codes"}, []string{"file_id", "code"}, rows)
	return err
}

// Implements CopyFromSource for a batch
type batchCopySource struct {
	batch [][]interface{}
	idx   int
}

func (b *batchCopySource) Next() bool {
	b.idx++
	return b.idx < len(b.batch)
}

func (b *batchCopySource) Values() ([]interface{}, error) {
	return b.batch[b.idx], nil
}

func (b *batchCopySource) Err() error {
	return nil
}
