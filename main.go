package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// Voucher represents a voucher with its conditions
type Voucher struct {
	Id                 int64
	Code               string
	MinOrderAmount     float64
	DiscountAmount     sql.NullFloat64
	DiscountPercentage sql.NullInt64
	MaxDiscountAmount  sql.NullFloat64
}

// Database connection details
const (
	dbHost     = "localhost"
	dbPort     = 5432
	dbUser     = "postgres"
	dbPassword = "admin"
	dbName     = "learning_db"
)

var (
	logger *zap.Logger
)

func init() {
	// Initialize logger
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
}

// fetchVouchers fetches all vouchers from the database
func fetchVouchers(ctx context.Context, db *sql.DB) ([]Voucher, error) {
	query := `
		SELECT id, code, min_order_amount, discount_amount, discount_percentage, max_discount_amount
		FROM vouchers
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch vouchers: %w", err)
	}
	defer rows.Close()

	var vouchers []Voucher
	for rows.Next() {
		var v Voucher
		if err := rows.Scan(&v.Id, &v.Code, &v.MinOrderAmount, &v.DiscountAmount, &v.DiscountPercentage, &v.MaxDiscountAmount); err != nil {
			return nil, fmt.Errorf("failed to scan voucher: %w", err)
		}
		vouchers = append(vouchers, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return vouchers, nil
}

// calculateDiscount calculates the discount for a voucher
func calculateDiscount(v Voucher, orderAmount float64) (float64, error) {
	if orderAmount < v.MinOrderAmount {
		return 0, errors.New("order amount does not meet minimum requirement")
	}

	var discount float64
	if v.DiscountAmount.Valid {
		// Flat discount
		discount = v.DiscountAmount.Float64
	} else if v.DiscountPercentage.Valid {
		// Percentage discount
		discount = orderAmount * float64(v.DiscountPercentage.Int64) / 100
		if v.MaxDiscountAmount.Valid && discount > v.MaxDiscountAmount.Float64 {
			discount = v.MaxDiscountAmount.Float64
		}
	} else {
		return 0, errors.New("invalid voucher discount configuration")
	}

	// Floor the discount to the nearest integer
	discount = math.Floor(discount)
	return discount, nil
}

// findBestVoucher finds the best voucher in parallel
func findBestVoucher(ctx context.Context, vouchers []Voucher, orderAmount float64) (Voucher, float64, error) {
	var (
		wg          sync.WaitGroup
		mu          sync.Mutex
		bestVoucher Voucher
		maxDiscount float64
	)

	// Create a worker pool with a limit of 10 workers
	workerPool := make(chan struct{}, 10)
	defer close(workerPool)

	resultChan := make(chan struct {
		Voucher  Voucher
		Discount float64
	}, len(vouchers))

	for _, v := range vouchers {
		wg.Add(1)
		workerPool <- struct{}{} // Acquire a worker slot

		go func(v Voucher) {
			defer wg.Done()
			defer func() { <-workerPool }() // Release the worker slot

			discount, err := calculateDiscount(v, orderAmount)
			if err != nil {
				logger.Warn("Voucher condition not met", zap.String("code", v.Code), zap.Error(err))
				return
			}

			resultChan <- struct {
				Voucher  Voucher
				Discount float64
			}{Voucher: v, Discount: discount}
		}(v)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		mu.Lock()
		if result.Discount > maxDiscount {
			maxDiscount = result.Discount
			bestVoucher = result.Voucher
		}
		mu.Unlock()
	}

	if maxDiscount == 0 {
		return Voucher{}, 0, errors.New("no applicable voucher found")
	}

	return bestVoucher, maxDiscount, nil
}

func main() {
	// Set up database connection
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	db.SetMaxOpenConns(50)                 // Limit the number of open connections
	db.SetMaxIdleConns(10)                 // Limit idle connections
	db.SetConnMaxLifetime(time.Minute * 5) // Close connections after 5 minutes
	defer db.Close()

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Fetch vouchers from the database
	vouchers, err := fetchVouchers(ctx, db)
	if err != nil {
		logger.Fatal("Failed to fetch vouchers", zap.Error(err))
	}

	// Define order amount
	orderAmount := 500.0

	// Find the best voucher
	bestVoucher, discount, err := findBestVoucher(ctx, vouchers, orderAmount)
	if err != nil {
		logger.Error("Failed to find best voucher", zap.Error(err))
		return
	}

	logger.Info("Best voucher found",
		zap.Int64("id", bestVoucher.Id),
		zap.String("code", bestVoucher.Code),
		zap.Float64("discount", discount),
	)
}
