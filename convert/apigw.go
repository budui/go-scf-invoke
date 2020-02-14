package convert

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/tencentyun/scf-go-lib/events"
)

// headertoSingleMap is workaroud of https://github.com/tencentyun/scf-go-lib/issues/15
func headertoSingleMap(header http.Header) map[string]string {
	newReqHeader := map[string]string{}
	for k, v := range header {
		newReqHeader[k] = v[0]
	}
	return newReqHeader
}

// NewAPIGatewayRequestFromRequest convert *http.Request to evnets.APIGatewayRequest
func NewAPIGatewayRequestFromRequest(req *http.Request) events.APIGatewayRequest {
	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return events.APIGatewayRequest{}
	}

	agwpq := events.APIGatewayRequest{
		Headers:     headertoSingleMap(req.Header),
		Method:      req.Method,
		Path:        req.URL.Path,
		QueryString: events.APIGatewayQueryString(req.URL.Query()),
		Body:        string(bodyBytes),
		Context: events.APIGatewayRequestContext{
			ServiceID: "service-invalid",
			RequestID: "invalid-request-id",
			Method:    "ANY",
			Path:      req.URL.Path,
			Stage:     "dev",
			SourceIP:  strings.Split(req.RemoteAddr, ":")[0],
		},
	}
	return agwpq
}
