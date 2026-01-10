package web

import (
	"net/http"
	"strconv"

	"github.com/delaneyj/toolbelt/id"
	"github.com/go-chi/chi/v5"
)

func ChiParamInt64(r *http.Request, name string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, name), 10, 64)
}

func ChiParamEncodedID(r *http.Request, name string) (int64, error) {
	return id.EncodedIDToInt64(chi.URLParam(r, name))
}
