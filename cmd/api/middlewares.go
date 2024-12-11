package main

import (
	"fmt"
	"golang.org/x/time/rate"
	"net"
	"net/http"
	"sync"
	"time"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function which always run in the event
		// of panic as Go unwinds the stack.
		defer func() {
			// Use the built-in recover function to check if
			// there has been a panic or not.
			if err := recover(); err != nil {
				// If there was a panic, set "Connection: close" header
				// on the response. This acts as a trigger to make
				// Go's HTTP server automatically close the current connection
				// after a response has been sent
				w.Header().Set("Connection", "close")

				// The value returned by recover is a type of any, so we use
				// fmt.Errorf to normalize it into an error and call
				// app.serverErrorResponse method.
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	// client is a struct for holding rate limiter and last seen for each client.
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Launch a background goroutine which removes old entries from the clients map
	// once every minute.
	go func() {
		for {
			time.Sleep(time.Minute)

			// Lock the mutex to prevent any rate limiter checks from happening
			// while the cleanup is taking place.
			mu.Lock()

			// Loop through all the clients. If they haven't been seen within
			// the last 3 minutes, delete the corresponding entry from the map.
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			// Unlock the mutex when the cleanup is complete.
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enabled {
			// Extract the client ip address from the request.
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			mu.Lock()

			// Check to see if the IP address already exists in the map. If it doesn't, then
			// initialize a new client and add the IP address and client to the map
			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
				}
			}

			// Update the last seen for the client.
			clients[ip].lastSeen = time.Now()

			// Call Allow() method on the rate limiter for the current IP address. If the
			// request isn't allowed, unlock the mutex and send a 429 Too Many Requests
			// response.
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			// Unlock the mutex before calling next handler
			mu.Unlock()

		}
		next.ServeHTTP(w, r)
	})
}
