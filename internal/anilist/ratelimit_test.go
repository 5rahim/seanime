package anilist

import (
	"github.com/seanime-app/seanime/internal/limiter"
	"sync"
	"testing"
	"time"
)

func performAPICall(rateLimiter *limiter.Limiter, i int, wg *sync.WaitGroup) int {
	defer wg.Done()

	rateLimiter.Wait()

	responseTime := 20 * time.Millisecond
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

	rateLimit := limiter.NewAnilistLimiter()

	go func() {
		for {
			<-time.Tick(time.Minute)
			println("1 minute")
		}
	}()

	for i := 1; i <= 100; i++ {
		wg.Add(1)
		go performAPICall(rateLimit, i, &wg)
	}

	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go performAPICall(rateLimit, i, &wg)
	}

	wg.Wait()
	t.Log("API calls completed without rate limiting issues.")
}
