package healthchecker

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/speijnik/go-errortree"
)

// TCPDialCheck returns a Check that checks TCP connectivity to the provided
// endpoint.
func TCPDialCheck(addr string, timeout time.Duration) Check {

	return func(ctx context.Context) error {
		var rcerror error

		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err != nil {
			return errortree.Add(rcerror, "TCPDialCheck", err)
		}

		return conn.Close()
	}
}

// HTTPGetCheck returns a Check that performs an HTTP GET request against the
// specified URL. The check fails if the response times out or returns a non-200
// status code.
func HTTPGetCheck(url string, timeout time.Duration) Check {
	client := http.Client{
		Timeout: timeout,
		// never follow redirects
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return func(ctx context.Context) error {
		var rcerror error

		resp, err := client.Get(url)
		if err != nil {
			return errortree.Add(rcerror, "HTTPGetCheck", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return errortree.Add(rcerror, "HTTPGetCheck", fmt.Errorf("returned status %d", resp.StatusCode))
		}

		return nil
	}
}
