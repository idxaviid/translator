package middlewares

import (
	"net/http"
	"strings"

	"github.com/translator/app/models"
)

// Middleware Middleware
type Middleware func(http.HandlerFunc) http.HandlerFunc

// ValidateContentType ValidateContentType
func ValidateContentType() Middleware {

	// Create a new Middleware
	return func(f http.HandlerFunc) http.HandlerFunc {

		// Define the http.HandlerFunc
		return func(w http.ResponseWriter, r *http.Request) {

			// Do middleware things
			contentType := r.Header.Get("content-type")
			if !strings.Contains(contentType, "application/json") {
				mR := models.MyResponse{}
				mR.Code = 1
				mR.Msg = "Invalid Content-Type"
				models.GenerateResponse(w, mR, http.StatusBadRequest)
				return
			}

			// Call the next middleware/handler in chain
			f(w, r)
		}
	}
}

// ValidateAuthorization ValidateAuthorization
func ValidateAuthorization() Middleware {

	// Create a new Middleware
	return func(f http.HandlerFunc) http.HandlerFunc {

		// Define the http.HandlerFunc
		return func(w http.ResponseWriter, r *http.Request) {

			// Do middleware things
			authorization := r.Header.Get("Authorization")
			if authorization == "" {
				mR := models.MyResponse{}
				mR.Code = 1
				mR.Msg = "Authorization is missing!"
				models.GenerateResponse(w, mR, http.StatusBadRequest)
				return
			}

			if authorization != "Basic bXl0cmFuc2xhdG9yOnRoaXNpc2FwYXNzd29yZA==" {
				mR := models.MyResponse{}
				mR.Code = 1
				mR.Msg = "Authorization is invalid!"
				models.GenerateResponse(w, mR, http.StatusBadRequest)
				return
			}

			// Call the next middleware/handler in chain
			f(w, r)
		}
	}
}

// Chain applies middlewares to a http.HandlerFunc
func Chain(f http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for _, m := range middlewares {
		f = m(f)
	}
	return f
}

func RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				mR := models.MyResponse{
					Code: 1,
					Msg:  "the server encountered a problem and could not process your request",
				}
				models.GenerateResponse(w, mR, http.StatusInternalServerError)
				return
			}
		}()
		next.ServeHTTP(w, r)
	})
}
