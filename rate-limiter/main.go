// https://codingchallenges.fyi/challenges/challenge-rate-limiter

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

// Token bucket

const BUCKET_CAPACITY = 10
const BUCKET_REFIL_INTERVAL = 5000
const BUCKET_REFIL_RATE = 1


type Bucket struct {
	tokens uint16
	capacity uint16
}

func (bucket *Bucket) refil(refilRate int16) {
	for {
		bucket.tokens = min(bucket.tokens+BUCKET_REFIL_RATE, bucket.capacity)
		time.Sleep(BUCKET_REFIL_INTERVAL*time.Millisecond)
	}
}

func createBucket(tokensCapacity uint16) *Bucket {
	bucket := &Bucket{tokensCapacity, tokensCapacity}
	go bucket.refil(BUCKET_REFIL_RATE)
	return bucket
}


var bucketMap map[string]*Bucket

type RateLimitExceeded struct {
	clientIP string
}
func (err *RateLimitExceeded) Error() string {
	return fmt.Sprintf("%s: exceeded rate limit.", err.clientIP)
}

func rateLimitTokenBucket(remoteAddr string) error {
	ip, _, _ := net.SplitHostPort(remoteAddr)
	bucket, exists := bucketMap[ip]
	if !exists {
		new_bucket := createBucket(BUCKET_CAPACITY)
		new_bucket.tokens--
		log.Printf("%s: remaining tokens - %d", ip, new_bucket.tokens)
		bucketMap[ip] = new_bucket
		return nil
	}

	if bucket.tokens > 0 {
		bucket.tokens--
		log.Printf("%s: remaining tokens - %d", ip, bucket.tokens)
		return nil
	}

	return &RateLimitExceeded{ip}
}

//--------------------------------------------------------

const RequestRate = 10	// 10 request per minute
const WindowLength = time.Duration(60)*time.Second

type FixedWindow struct {
	Counter map[string]int
}


func initFixedWindow() *FixedWindow {
	counter := make(map[string]int)
	return &FixedWindow{counter}
}


var fixedWindow *FixedWindow


func rateLimitFixedWindow(remoteAddr string) error {
	ip, _, _ := net.SplitHostPort(remoteAddr)

	count, exists := fixedWindow.Counter[ip]
	if !exists {
		fixedWindow.Counter[ip] =  1
		log.Printf("%s: request count - %d", ip, 1)
		return nil
	}

	if count + 1 < RequestRate {
		fixedWindow.Counter[ip] = count + 1
		log.Printf("%s: request count - %d", ip, count+1)
		return nil
	}

	return &RateLimitExceeded{}
}

func renewFixedWindow() {
	for {
		fixedWindow = initFixedWindow()
		time.Sleep(WindowLength)
	}
}


//-------------------------


const SlidingWindowLenght = time.Duration(60)*time.Second
const SlidingWindowRequestRate = 10
const SecToPercent = 0.016666666


type SlidingWindow struct {
	Counter map[string]int
	WindowTimestamp time.Time
}

func initSlidingWindow(windowTimestamp time.Time) *SlidingWindow {
	counter := make(map[string]int)
	return &SlidingWindow{counter, windowTimestamp}
}

var currSlidingWindow *SlidingWindow
var prevSlidingWindow *SlidingWindow

func moveSlidingWindow() {
	for {
		prevSlidingWindow = currSlidingWindow
		currTimestamp := time.Now()
		windowTimestamp := currTimestamp.Truncate(time.Minute)
		currSlidingWindow = initSlidingWindow(windowTimestamp)
		time.Sleep(time.Minute - time.Second*time.Duration(currTimestamp.Second()))
	}
}

func rateLimitSlidingWindow(remoteAddr string) error {
	ip, _, _ := net.SplitHostPort(remoteAddr)
	currWindowCount, exists := currSlidingWindow.Counter[ip]
	if !exists {
		currWindowCount = 0
	}
	currWindowCount++

	prevWindowCount := 0
	if prevSlidingWindow != nil {
		count, exists := prevSlidingWindow.Counter[ip]
		if exists {
			prevWindowCount = count
		}
	}

	currTimestamp := time.Now()
	prevWindowInfluence := 1-float32(currTimestamp.Second())/float32(60)

	requestCount := currWindowCount + int(prevWindowInfluence*float32(prevWindowCount))
	if requestCount > SlidingWindowRequestRate {
		log.Printf("%s: window - %s, rate limit exceeded\n", ip, currSlidingWindow.WindowTimestamp.Format("15:04:05"))
		return &RateLimitExceeded{ip}
	}

	currSlidingWindow.Counter[ip] = currWindowCount
	log.Printf("%s: window - %s, request count - %d\n", ip, currSlidingWindow.WindowTimestamp.Format("15:04:05"), requestCount)
	return nil
}

func limited(w http.ResponseWriter, req *http.Request) {
	err := rateLimitSlidingWindow(req.RemoteAddr)
	if err != nil {
		http.Error(w, "Too many requests!", http.StatusTooManyRequests)
		return
	}

	fmt.Fprint(w, "Limited\n")
}

func unlimited(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, "Unlimited\n")
}


func main() {

	go moveSlidingWindow()

	http.HandleFunc("/limited", limited)
	http.HandleFunc("/unlimited", unlimited)

	log.Fatal(http.ListenAndServe(":8080", nil))
}