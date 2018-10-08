package middleware

import (
	"bitbucket.org/cfchou/icecream/cmd/apiserver/mocks"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIKeyMiddleWare_Handle(t *testing.T) {
	mB := &mocks.APIKeyBackend{}

	apiKey := "123"
	expected := []byte("valid apikey")
	am := CreateAPIKeyMiddleWare(mB)

	mB.On("Authenticate", apiKey).Return(nil)

	f := func(w http.ResponseWriter, r *http.Request) {
		w.Write(expected)
	}

	mux := http.NewServeMux()
	//mux.Handle("/", am.Handle(mH))
	mux.Handle("/", am.Handle(http.HandlerFunc(f)))

	writer := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/", nil)
	request.Header.Set("Authorization", apiKey)
	mux.ServeHTTP(writer, request)
	assert.Equal(t, 200, writer.Code)

	// f is not called
	bs, _ := ioutil.ReadAll(writer.Body)
	assert.Equal(t, expected, bs)
}

func TestAPIKeyMiddleWare_Handle_401NoAuthorization(t *testing.T) {
	mB := &mocks.APIKeyBackend{}

	expected := []byte("valid apikey")
	am := CreateAPIKeyMiddleWare(mB)

	f := func(w http.ResponseWriter, r *http.Request) {
		w.Write(expected)
	}

	mux := http.NewServeMux()
	mux.Handle("/", am.Handle(http.HandlerFunc(f)))

	writer := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/", nil)
	mux.ServeHTTP(writer, request)
	assert.Equal(t, 401, writer.Code)

	// f is not called
	bs, _ := ioutil.ReadAll(writer.Body)
	assert.NotEqual(t, expected, bs)

	// mB is not called
	mB.AssertNotCalled(t, "Authenticate", mock.Anything)
}

func TestAPIKeyMiddleWare_Handle_401InvalidAPIKey(t *testing.T) {
	mB := &mocks.APIKeyBackend{}

	apiKey := "123"
	expected := []byte("valid apikey")

	am := CreateAPIKeyMiddleWare(mB)

	mB.On("Authenticate", mock.Anything).
		Return(fmt.Errorf("any error"))

	f := func(w http.ResponseWriter, r *http.Request) {
		w.Write(expected)
	}

	mux := http.NewServeMux()
	mux.Handle("/", am.Handle(http.HandlerFunc(f)))

	writer := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/", nil)
	request.Header.Set("Authorization", apiKey)
	mux.ServeHTTP(writer, request)
	assert.Equal(t, 401, writer.Code)

	// f is not called
	bs, _ := ioutil.ReadAll(writer.Body)
	assert.NotEqual(t, expected, bs)

	// mB is called
	mB.AssertCalled(t, "Authenticate", mock.Anything)
}
