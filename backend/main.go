package backend

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/favclip/ucon"
	"github.com/favclip/ucon/swagger"
)

func Setup() {
	ucon.Orthodox()
	ucon.Middleware(swagger.RequestValidator())

	swPlugin := swagger.NewPlugin(&swagger.Options{
		Object: &swagger.Object{
			Info: &swagger.Info{
				Title:   "GCPUG",
				Version: "v1",
			},
			Schemes: []string{"http", "https"},
		},
		DefinitionNameModifier: func(refT reflect.Type, defName string) string {
			if strings.HasSuffix(defName, "JSON") {
				return defName[:len(defName)-4]
			}
			return defName
		},
	})
	ucon.Plugin(swPlugin)

	setupSpannerAccountAPI(swPlugin)
	http.Handle("/api/", ucon.DefaultMux)
	ucon.DefaultMux.Prepare()
}

// HTTPError is API Resposeとして返すError
type HTTPError struct {
	Code    int         `json:"code"`
	Message interface{} `json:"message"`
}

// StatusCode is Http Response Status Codeを返す
func (he *HTTPError) StatusCode() int {
	return he.Code
}

// ErrorMessage is Clientに返すErrorMessageを返す
func (he *HTTPError) ErrorMessage() interface{} {
	return he
}

// Error is error interfaceを実装
func (he *HTTPError) Error() string {
	return fmt.Sprintf("status code %d: %s", he.StatusCode(), he.ErrorMessage())
}
