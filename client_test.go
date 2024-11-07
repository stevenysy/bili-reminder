package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestCookieValid tests if the SESSDATA cookie can be used to access the watch later list
func TestCookieValid(t *testing.T) {
	sessdataCookie := os.Getenv("SESSDATA")
	require.NotEmpty(t, sessdataCookie)

	req, err := http.NewRequest("GET", WATCH_LATER_URL, nil)
	require.NoError(t, err)

	req.Header.Set("Cookie", fmt.Sprintf("SESSDATA=%s", sessdataCookie))

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	responseBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NotEmpty(t, responseBody)

	var watchLater watchLaterResponse
	err = json.Unmarshal(responseBody, &watchLater)
	require.NoError(t, err)
	require.Equal(t, 0, watchLater.Code) // 0 means success
}
