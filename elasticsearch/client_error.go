package elasticsearch

import (
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func responseError(action string, resp *esapi.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("elastic: %s failed with status %d: read response body: %w", action, resp.StatusCode, err)
	}
	return fmt.Errorf("elastic: %s failed with status %d: %s", action, resp.StatusCode, string(body))
}
