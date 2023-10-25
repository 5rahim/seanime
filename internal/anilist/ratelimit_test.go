package anilist

import (
	"github.com/seanime-app/seanime-server/internal/limiter"
	"sync"
	"testing"
	"time"
)

func performAPICall(rateLimiter *limiter.Limiter, i int, wg *sync.WaitGroup) int {
	wg.Add(1)
	defer wg.Done()

	rateLimiter.Wait()

	responseTime := 200 * time.Millisecond
	time.Sleep(responseTime)
	println("performed api call", i)
	return i
}

func TestRateLimitedAPICalls(t *testing.T) {

	var wg = sync.WaitGroup{}

	// Create a rate limiter to limit to 90 requests per 5 seconds
	rateLimit := limiter.NewLimiter(time.Second*5, 5)

	// Perform 10 API calls with rate limiting
	for i := 0; i < 10; i++ {
		go performAPICall(rateLimit, i, &wg)
	}

	wg.Wait()
	t.Log("API calls completed without rate limiting issues.")
}

func TestRateLimitedAPICalls2(t *testing.T) {

	var wg = sync.WaitGroup{}

	// Create a rate limiter to limit to 90 requests per 5 seconds
	rateLimit := limiter.NewLimiter(time.Minute, 90)

	// Perform 10 API calls with rate limiting
	for i := 1; i <= 120; i++ {
		go performAPICall(rateLimit, i, &wg)
	}

	wg.Wait()
	t.Log("API calls completed without rate limiting issues.")
}
