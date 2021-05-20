package lib

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

// construct the body of the request
func ConstructBody(data string) (io.Reader, error) {

	var body io.Reader
	var err error

	if strings.HasPrefix(data, "@") {
		// Read data from file
		fPath := data[1:]
		body, err = os.Open(fPath)
		if err != nil {
			return body, err
		}
	} else {
		body = strings.NewReader(data)
	}
	return body, nil
}

func AddHeaders(request *http.Request, headers []string) error {
	for _, h := range headers {
		hParts := strings.Split(h, ":")
		if len(hParts) != 2 {
			return fmt.Errorf(`Error: Invalid header: %s. It should be in the format "Name: Value"`, h)
		}
		hKey := strings.TrimSpace(hParts[0])
		hVal := strings.TrimSpace(hParts[1])
		request.Header.Add(hKey, hVal)
	}
	return nil
}

func Sign(request *http.Request, cfg aws.Config, service string, region string) error {
	body := readAndReplaceBody(request)
	bodySHA256 := hashSHA256(body)
	signer := v4.NewSigner(cfg.Credentials)
	err := signer.SignHTTP(request.Context(), request, bodySHA256, service, region, time.Now())
	if err != nil {
		return err
	}
	return nil
}

func readAndReplaceBody(request *http.Request) []byte {
	if request.Body == nil {
		return []byte{}
	}
	payload, _ := ioutil.ReadAll(request.Body)
	request.Body = ioutil.NopCloser(bytes.NewReader(payload))
	return payload
}
