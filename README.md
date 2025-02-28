# Go-Routine, Channel & sync.Mutex Real Life Example

This is a Go-based application that automatically applies the best voucher code based on predefined conditions. It uses goroutines, channels, and a PostgreSQL database to efficiently handle concurrent requests and determine the best voucher for a given order amount.

## Features

- **Parallel Processing**: Uses goroutines to check voucher conditions concurrently.
- **Database Integration**: Fetches voucher data from a PostgreSQL database.
- **Concurrency Safety**: Uses `sync.Mutex` to safely handle shared resources.
- **Scalable**: Designed to handle high concurrent requests.
- **Production-Ready**: Includes error handling, logging, and context-based timeouts.

## Installation

1. Clone the repository:
   ```sh
   git clone https://github.com/faisal-porag/goroutine_channel_example.git
   cd voucher-discount
   ```

2. Install dependencies:
   ```sh
   go mod tidy
   ```

3. Set up your PostgreSQL database and update the connection string in the code:
   ```sh
   export DATABASE_URL="postgres://user:password@localhost:5432/dbname?sslmode=disable"
   ```

## Usage

Run the application with:
   ```sh
   go run main.go
   ```

## How It Works

1. **Fetches vouchers** from the PostgreSQL database.
2. **Processes each voucher in parallel** using goroutines.
3. **Calculates applicable discounts** based on the order amount.
4. **Uses a mutex to track the best voucher** safely.
5. **Returns the best voucher** with the highest discount.

## Example Output
```
Best Voucher: voucher1 | Discount: 85.00
```

## Performance Considerations

- **Parallel execution** improves speed.
- **Buffered channels** prevent blocking.
- **Mutex ensures thread safety**.
- **Optimized database queries** improve efficiency.

## License

This project is open-source under the [MIT License](LICENSE).

