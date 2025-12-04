package network

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// =============================================================================
// Constants and Configuration
// =============================================================================

const (
	// Download test configuration
	// Using Cloudflare's speed test files (reliable CDN)
	downloadURL          = "https://speed.cloudflare.com/__down?bytes=26214400" // 25MB chunk
	downloadConcurrency  = 4                                                    // Number of parallel download goroutines
	downloadTestDuration = 10 * time.Second                                     // Maximum test duration

	// Upload test configuration
	// Using httpbin's /post endpoint for uploads (echoes back data)
	uploadURL          = "https://httpbin.org/post"
	uploadConcurrency  = 4                // Number of parallel upload goroutines
	uploadChunkSize    = 1 * 1024 * 1024  // 1MB per upload chunk
	uploadTestDuration = 10 * time.Second // Maximum test duration

	// Ping test configuration
	pingURL   = "https://speed.cloudflare.com/__down?bytes=0"
	pingCount = 5 // Number of ping requests

	// HTTP client configuration
	httpTimeout = 30 * time.Second

	// Buffer size for reading responses (streaming, not loading into RAM)
	readBufferSize = 32 * 1024 // 32KB buffer
)

// =============================================================================
// Speed Test Messages and Types
// =============================================================================

// SpeedTestResult holds the complete results of a speed test
type SpeedTestResult struct {
	PingResult   *SpeedTestPingResult
	DownloadMbps float64
	UploadMbps   float64
	Error        string
	Server       string
	TestDuration time.Duration
}

// SpeedTestMsg is a tea.Msg that contains speed test results
type SpeedTestMsg SpeedTestResult

// SpeedTestProgressMsg is a tea.Msg that contains progress updates during the test
type SpeedTestProgressMsg struct {
	Stage       string  // "ping", "download", "upload"
	Progress    float64 // 0.0 to 1.0
	Message     string
	CurrentMbps float64 // Current speed for download/upload stages
}

// SpeedTestErrorMsg is a tea.Msg for speed test errors
type SpeedTestErrorMsg struct {
	Error string
}

// =============================================================================
// HTTP Client Setup
// =============================================================================

// createHTTPClient creates a configured HTTP client with timeouts.
func createHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   false,
		},
	}
}

// =============================================================================
// Ping Test
// =============================================================================

// SpeedTestPingResult holds the results of the ping test for speed tests.
type SpeedTestPingResult struct {
	Min     time.Duration
	Max     time.Duration
	Average time.Duration
	Samples int
}

// testPing performs latency measurements to the specified URL.
// It makes multiple requests and calculates min, max, and average RTT.
func testPing(client *http.Client, url string, count int) (*SpeedTestPingResult, error) {
	var samples []time.Duration

	for range count {
		start := time.Now()

		req, err := http.NewRequest(http.MethodHead, url, nil)
		if err != nil {
			// Skip failed request creation, try next
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			// Skip failed requests, try next
			continue
		}
		resp.Body.Close()

		rtt := time.Since(start)
		samples = append(samples, rtt)

		// Small delay between pings
		time.Sleep(100 * time.Millisecond)
	}

	if len(samples) == 0 {
		return nil, fmt.Errorf("all ping attempts failed")
	}

	// Calculate statistics
	result := &SpeedTestPingResult{
		Min:     samples[0],
		Max:     samples[0],
		Samples: len(samples),
	}

	var total time.Duration
	for _, s := range samples {
		total += s
		if s < result.Min {
			result.Min = s
		}
		if s > result.Max {
			result.Max = s
		}
	}
	result.Average = total / time.Duration(len(samples))

	return result, nil
}

// =============================================================================
// Download Test
// =============================================================================

// downloadWorker performs a single download stream and reports bytes downloaded.
func downloadWorker(ctx context.Context, client *http.Client, url string, bytesDownloaded *int64, wg *sync.WaitGroup, progressChan chan<- int64) {
	defer wg.Done()

	buffer := make([]byte, readBufferSize)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			// Check if context was cancelled
			if ctx.Err() != nil {
				return
			}
			// Retry on error
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Stream the response body to discard, counting bytes
		for {
			select {
			case <-ctx.Done():
				resp.Body.Close()
				return
			default:
			}

			n, err := resp.Body.Read(buffer)
			if n > 0 {
				atomic.AddInt64(bytesDownloaded, int64(n))
				if progressChan != nil {
					progressChan <- int64(n)
				}
			}
			if err != nil {
				break
			}
		}
		resp.Body.Close()
	}
}

// testDownload measures download speed using parallel connections.
// Returns the speed in Mbps (megabits per second).
func testDownload(client *http.Client, url string, concurrency int, duration time.Duration, progressChan chan<- int64) (float64, error) {
	var bytesDownloaded int64
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	// Start download workers
	for range concurrency {
		wg.Add(1)
		go downloadWorker(ctx, client, url, &bytesDownloaded, &wg, progressChan)
	}

	start := time.Now()

	// Wait for all workers to complete
	wg.Wait()

	elapsed := time.Since(start).Seconds()

	if bytesDownloaded == 0 {
		return 0, fmt.Errorf("no data downloaded")
	}

	// Calculate Mbps: (bytes * 8) / seconds / 1,000,000
	mbps := (float64(bytesDownloaded) * 8) / elapsed / 1_000_000

	return mbps, nil
}

// =============================================================================
// Upload Test
// =============================================================================

// randomDataReader provides a reader that generates random data on demand.
type randomDataReader struct {
	remaining int64
}

// Read implements io.Reader for randomDataReader.
func (r *randomDataReader) Read(p []byte) (n int, err error) {
	if r.remaining <= 0 {
		return 0, io.EOF
	}

	toRead := min(int64(len(p)), r.remaining)

	// Generate random data
	n, err = rand.Read(p[:toRead])
	r.remaining -= int64(n)

	return n, err
}

// uploadWorker performs upload operations and reports bytes uploaded.
func uploadWorker(ctx context.Context, client *http.Client, url string, chunkSize int, bytesUploaded *int64, wg *sync.WaitGroup, progressChan chan<- int64) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Create a reader with random data
		reader := &randomDataReader{remaining: int64(chunkSize)}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, reader)
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/octet-stream")
		req.ContentLength = int64(chunkSize)

		resp, err := client.Do(req)
		if err != nil {
			// Check if context was cancelled
			if ctx.Err() != nil {
				return
			}
			// Retry on error
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Discard response body
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		// Count the uploaded bytes
		atomic.AddInt64(bytesUploaded, int64(chunkSize))
		if progressChan != nil {
			progressChan <- int64(chunkSize)
		}
	}
}

// testUpload measures upload speed using parallel connections.
// Returns the speed in Mbps (megabits per second).
func testUpload(client *http.Client, url string, concurrency int, chunkSize int, duration time.Duration, progressChan chan<- int64) (float64, error) {
	var bytesUploaded int64
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	// Start upload workers
	for range concurrency {
		wg.Add(1)
		go uploadWorker(ctx, client, url, chunkSize, &bytesUploaded, &wg, progressChan)
	}

	start := time.Now()

	// Wait for all workers to complete
	wg.Wait()

	elapsed := time.Since(start).Seconds()

	if bytesUploaded == 0 {
		return 0, fmt.Errorf("no data uploaded")
	}

	// Calculate Mbps: (bytes * 8) / seconds / 1,000,000
	mbps := (float64(bytesUploaded) * 8) / elapsed / 1_000_000

	return mbps, nil
}

// =============================================================================
// Server Selection
// =============================================================================

// selectServer verifies server connectivity with retries.
func selectServer(client *http.Client, url string, maxRetries int) error {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Use GET instead of HEAD for better compatibility with CDN servers
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			continue
		}

		// Use a short timeout for server selection
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		req = req.WithContext(ctx)

		resp, err := client.Do(req)
		cancel()

		if err != nil {
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			return fmt.Errorf("unable to contact server after %d attempts: %v", maxRetries, err)
		}

		// Discard response body
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			return nil
		}

		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}

	return fmt.Errorf("server unavailable after %d attempts", maxRetries)
}

// =============================================================================
// Main Speed Test Function
// =============================================================================

// RunSpeedTest performs a complete network speed test and returns results as tea.Msg
func RunSpeedTest() tea.Cmd {
	return func() tea.Msg {
		startTime := time.Now()

		// Create HTTP client with timeout
		client := createHTTPClient(httpTimeout)

		// Server selection
		if err := selectServer(client, pingURL, 3); err != nil {
			return SpeedTestErrorMsg{Error: fmt.Sprintf("Server selection failed: %v", err)}
		}

		// Ping test
		pingResult, err := testPing(client, pingURL, pingCount)
		if err != nil {
			return SpeedTestErrorMsg{Error: fmt.Sprintf("Ping test failed: %v", err)}
		}

		// Download test
		downloadMbps, err := testDownload(client, downloadURL, downloadConcurrency, downloadTestDuration, nil)
		if err != nil {
			return SpeedTestErrorMsg{Error: fmt.Sprintf("Download test failed: %v", err)}
		}

		// Upload test
		uploadMbps, err := testUpload(client, uploadURL, uploadConcurrency, uploadChunkSize, uploadTestDuration, nil)
		if err != nil {
			return SpeedTestErrorMsg{Error: fmt.Sprintf("Upload test failed: %v", err)}
		}

		testDuration := time.Since(startTime)

		return SpeedTestMsg(SpeedTestResult{
			PingResult:   pingResult,
			DownloadMbps: downloadMbps,
			UploadMbps:   uploadMbps,
			Server:       "Cloudflare Speed Test",
			TestDuration: testDuration,
		})
	}
}

// RunSpeedTestWithProgress performs a speed test with progress updates
func RunSpeedTestWithProgress(progressChan chan<- SpeedTestProgressMsg) tea.Cmd {
	return func() tea.Msg {
		startTime := time.Now()

		// Create HTTP client with timeout
		client := createHTTPClient(httpTimeout)

		// Server selection
		if progressChan != nil {
			progressChan <- SpeedTestProgressMsg{
				Stage:    "server",
				Progress: 0.1,
				Message:  "Selecting server...",
			}
		}

		if err := selectServer(client, pingURL, 3); err != nil {
			return SpeedTestErrorMsg{Error: fmt.Sprintf("Server selection failed: %v", err)}
		}

		// Ping test
		if progressChan != nil {
			progressChan <- SpeedTestProgressMsg{
				Stage:    "ping",
				Progress: 0.2,
				Message:  "Testing latency...",
			}
		}

		pingResult, err := testPing(client, pingURL, pingCount)
		if err != nil {
			return SpeedTestErrorMsg{Error: fmt.Sprintf("Ping test failed: %v", err)}
		}

		// Download test with progress
		if progressChan != nil {
			progressChan <- SpeedTestProgressMsg{
				Stage:    "download",
				Progress: 0.4,
				Message:  "Testing download speed...",
			}
		}

		downloadMbps, err := testDownload(client, downloadURL, downloadConcurrency, downloadTestDuration, nil)
		if err != nil {
			return SpeedTestErrorMsg{Error: fmt.Sprintf("Download test failed: %v", err)}
		}

		// Upload test
		if progressChan != nil {
			progressChan <- SpeedTestProgressMsg{
				Stage:    "upload",
				Progress: 0.7,
				Message:  "Testing upload speed...",
			}
		}

		uploadMbps, err := testUpload(client, uploadURL, uploadConcurrency, uploadChunkSize, uploadTestDuration, nil)
		if err != nil {
			return SpeedTestErrorMsg{Error: fmt.Sprintf("Upload test failed: %v", err)}
		}

		testDuration := time.Since(startTime)

		if progressChan != nil {
			progressChan <- SpeedTestProgressMsg{
				Stage:    "complete",
				Progress: 1.0,
				Message:  "Speed test completed",
			}
		}

		return SpeedTestMsg(SpeedTestResult{
			PingResult:   pingResult,
			DownloadMbps: downloadMbps,
			UploadMbps:   uploadMbps,
			Server:       "Cloudflare Speed Test",
			TestDuration: testDuration,
		})
	}
}
