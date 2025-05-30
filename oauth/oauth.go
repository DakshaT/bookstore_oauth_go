package oauth

import (
	//"errors"
	//"github.com/DakshaT/bookstoreapp/bookstore_utils-go/rest_errors"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DakshaT/bookstore_utils-go/rest_errors"

	"github.com/mercadolibre/golang-restclient/rest"
)

var (
	oauthResClient = rest.RequestBuilder{
		BaseURL: "http://localhost:8083",
		Timeout: 200 * time.Millisecond,
	}
)

const (
	headerXPublic        = "X-Public"
	headerXClientId      = "X-Client-Id"
	headerXCallerId      = "X-Caller-Id"
	parameterAccessToken = "access_token"
)

type accessToken struct {
	Id       string `json:"Id"`
	UserId   int64  `json:"user_id"`
	ClientId int64  `json:"client_id"`
}

func IsPublic(request *http.Request) bool {
	if request == nil {
		return true
	}
	return request.Header.Get(headerXPublic) == "true"

}

func AuthenticateRequest(request *http.Request) *rest_errors.RestErr {
	if request == nil {
		return nil
	}
	cleanRequest(request)
	accessTokenId := strings.TrimSpace(request.URL.Query().Get(parameterAccessToken))
	if accessTokenId == "" {
		return nil
	}
	at, err := getAccessToken(accessTokenId)
	if err != nil {
		if err.Code == http.StatusNotFound {
			return nil
		}
		return err
	}
	request.Header.Add(headerXCallerId, fmt.Sprintf("%v", at.UserId))
	request.Header.Add(headerXClientId, fmt.Sprintf("%v", at.ClientId))
	return nil
}

func cleanRequest(request *http.Request) {
	if request == nil {
		return
	}
	request.Header.Del(headerXClientId)
	request.Header.Del(headerXCallerId)
}

func getCallerId(request *http.Request) int64 {
	if request == nil {
		return 0
	}
	callerId, err := strconv.ParseInt(request.Header.Get(headerXCallerId), 10, 64)
	if err != nil {
		return 0
	}
	return callerId

}

func getClientId(request *http.Request) int64 {
	if request == nil {
		return 0
	}
	clientId, err := strconv.ParseInt(request.Header.Get(headerXClientId), 10, 64)
	if err != nil {
		return 0
	}
	return clientId
}

func getAccessToken(accessTokenId string) (*accessToken, *rest_errors.RestErr) {
	response := oauthResClient.Get(fmt.Sprintf("/oauth/access_token/%s", accessTokenId))
	if response == nil || response.Response == nil {
		return nil, rest_errors.NewInternalServerError("invalid restclient response when trying to get access token.")
	}
	if response.StatusCode > 299 {
		var restErr rest_errors.RestErr
		err := json.Unmarshal(response.Bytes(), &restErr)
		if err != nil {
			return nil, rest_errors.NewInternalServerError("invalid error interface while trying to get access token.")
		}
		return nil, &restErr
	}

	var at accessToken
	err := json.Unmarshal(response.Bytes(), &at)
	if err != nil {
		return nil, rest_errors.NewInternalServerError("Error while trying to unmarshal access token response")
	}
	return &at, nil

}
