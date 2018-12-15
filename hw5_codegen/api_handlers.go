
package main
import (
	"net/http"
	"strconv"
	"encoding/json"
	"strings"
	"fmt"
)
	
func (h *MyApi ) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/profile":
		h.ProfileHandler(w, r)
	case "/user/create":
		h.CreateHandler(w, r)
	default:
		// 404
		w.WriteHeader(http.StatusNotFound)
		mk := make(map[string]interface{})
		mk["error"] = "unknown method"
		resp, _ := json.Marshal(mk)
		w.Write(resp)
	}
}
	
func (h *OtherApi ) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/create":
		h.CreateHandler(w, r)
	default:
		// 404
		w.WriteHeader(http.StatusNotFound)
		mk := make(map[string]interface{})
		mk["error"] = "unknown method"
		resp, _ := json.Marshal(mk)
		w.Write(resp)
	}
}
	
// URL: /user/profile
func (h *MyApi) ProfileHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	tmp := ""
	fmt.Printf("tmp = %+v\n", tmp)
	params:=new(ProfileParams)

	if r.Method == "POST" {
		params.Login = r.FormValue(strings.ToLower("Login"))
	} else {
		params.Login = r.URL.Query().Get(strings.ToLower("Login"))
	}
	
	if params.Login == "" {
		w.WriteHeader(http.StatusBadRequest)
		mk := make(map[string]interface{})
		mk["error"] = strings.ToLower("Login") + " must me not empty"
		resp, _ := json.Marshal(mk)
		w.Write(resp)
		return
	}
	
	ctx := r.Context()

	res, err := h.Profile(ctx, *params)
	if err != nil {
		e, ok := err.(ApiError)
		if ok {
			w.WriteHeader(e.HTTPStatus)
			mk := make(map[string]interface{})
			mk["error"] = err.Error()
			resp, _ := json.Marshal(mk)
			w.Write(resp)
			return
		} else {
			if err != nil && err.Error() == "bad user" {
				w.WriteHeader(http.StatusInternalServerError)
				mk := make(map[string]interface{})
				mk["error"] = err.Error()
				resp, _ := json.Marshal(mk)
				w.Write(resp)
				return
			}
		}
	}
	w.WriteHeader(http.StatusOK)
	mk := make(map[string]interface{})
	mk["error"] =	 ""
	mk["response"] = res
	resp, _ := json.Marshal(mk)
	w.Write(resp)
	
}

// URL: /user/create
func (h *MyApi) CreateHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	tmp := ""
	fmt.Printf("tmp = %+v\n", tmp)
	params:=new(CreateParams)

	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotAcceptable)
		mk := make(map[string]interface{})
		mk["error"] = "bad method"
		resp, _ := json.Marshal(mk)
		w.Write(resp)
		return
	}
	
	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		mp := make(map[string]interface{})
		mp["error"] = "unauthorized"
		js, _ := json.Marshal(mp)
		w.Write(js)
		return
	}
	
	if r.Method == "POST" {
		params.Login = r.FormValue(strings.ToLower("Login"))
	} else {
		params.Login = r.URL.Query().Get(strings.ToLower("Login"))
	}
	
	if r.Method == "POST" {
		params.Name = r.FormValue(strings.ToLower("full_name"))
	} else {
		params.Name = r.URL.Query().Get(strings.ToLower("full_name"))
	}
	
	if r.Method == "POST" {
		params.Status = r.FormValue(strings.ToLower("Status"))
	} else {
		params.Status = r.URL.Query().Get(strings.ToLower("Status"))
	}
	
	if r.Method == "POST" {
		tmp = r.FormValue(strings.ToLower("Age"))
	} else {
		tmp = r.URL.Query().Get(strings.ToLower("Age"))
	}
	params.Age, err = strconv.Atoi(tmp)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		mk := make(map[string]interface{})
		mk["error"] = strings.ToLower("Age") + " must be int"
		resp, _ := json.Marshal(mk)
		w.Write(resp)
		return
	}
	
	if params.Login == "" {
		w.WriteHeader(http.StatusBadRequest)
		mk := make(map[string]interface{})
		mk["error"] = strings.ToLower("Login") + " must me not empty"
		resp, _ := json.Marshal(mk)
		w.Write(resp)
		return
	}
	
	if !(len(params.Login) >= 10) {
		w.WriteHeader(http.StatusBadRequest)
		mk := make(map[string]interface{})
		mk["error"] = strings.ToLower("Login") + " len must be >= 10"
		resp, _ := json.Marshal(mk)
		w.Write(resp)
		return
	}
	
	
	switch params.Status {
		case "user":
		case "moderator":
		case "admin":
		// params.Status = Status
	default:
		if params.Status != "" {
			w.WriteHeader(http.StatusBadRequest)
			mk := make(map[string]interface{})
			mk["error"] = strings.ToLower("Status") + " must be one of [user, moderator, admin]"
			resp, _ := json.Marshal(mk)
			w.Write(resp)
			return
		}
		params.Status = "user"
	}
	if !(params.Age >= 0) {
		w.WriteHeader(http.StatusBadRequest)
		mk := make(map[string]interface{})
		mk["error"] = strings.ToLower("Age") + " must be >= 0"
		resp, _ := json.Marshal(mk)
		w.Write(resp)
		return
	}
	
	if !(params.Age < 128) {
		w.WriteHeader(http.StatusBadRequest)
		mk := make(map[string]interface{})
		mk["error"] = strings.ToLower("Age") + " must be <= 128"
		resp, _ := json.Marshal(mk)
		w.Write(resp)
		return
	}
	
	ctx := r.Context()

	res, err := h.Create(ctx, *params)
	if err != nil {
		e, ok := err.(ApiError)
		if ok {
			w.WriteHeader(e.HTTPStatus)
			mk := make(map[string]interface{})
			mk["error"] = err.Error()
			resp, _ := json.Marshal(mk)
			w.Write(resp)
			return
		} else {
			if err != nil && err.Error() == "bad user" {
				w.WriteHeader(http.StatusInternalServerError)
				mk := make(map[string]interface{})
				mk["error"] = err.Error()
				resp, _ := json.Marshal(mk)
				w.Write(resp)
				return
			}
		}
	}
	w.WriteHeader(http.StatusOK)
	mk := make(map[string]interface{})
	mk["error"] =	 ""
	mk["response"] = res
	resp, _ := json.Marshal(mk)
	w.Write(resp)
	
}

// URL: /user/create
func (h *OtherApi) CreateHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	tmp := ""
	fmt.Printf("tmp = %+v\n", tmp)
	params:=new(OtherCreateParams)

	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotAcceptable)
		mk := make(map[string]interface{})
		mk["error"] = "bad method"
		resp, _ := json.Marshal(mk)
		w.Write(resp)
		return
	}
	
	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		mp := make(map[string]interface{})
		mp["error"] = "unauthorized"
		js, _ := json.Marshal(mp)
		w.Write(js)
		return
	}
	
	if r.Method == "POST" {
		params.Username = r.FormValue(strings.ToLower("Username"))
	} else {
		params.Username = r.URL.Query().Get(strings.ToLower("Username"))
	}
	
	if r.Method == "POST" {
		params.Name = r.FormValue(strings.ToLower("account_name"))
	} else {
		params.Name = r.URL.Query().Get(strings.ToLower("account_name"))
	}
	
	if r.Method == "POST" {
		params.Class = r.FormValue(strings.ToLower("Class"))
	} else {
		params.Class = r.URL.Query().Get(strings.ToLower("Class"))
	}
	
	if r.Method == "POST" {
		tmp = r.FormValue(strings.ToLower("Level"))
	} else {
		tmp = r.URL.Query().Get(strings.ToLower("Level"))
	}
	params.Level, err = strconv.Atoi(tmp)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		mk := make(map[string]interface{})
		mk["error"] = strings.ToLower("Level") + " must be int"
		resp, _ := json.Marshal(mk)
		w.Write(resp)
		return
	}
	
	if params.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		mk := make(map[string]interface{})
		mk["error"] = strings.ToLower("Username") + " must me not empty"
		resp, _ := json.Marshal(mk)
		w.Write(resp)
		return
	}
	
	if !(len(params.Username) >= 3) {
		w.WriteHeader(http.StatusBadRequest)
		mk := make(map[string]interface{})
		mk["error"] = strings.ToLower("Username") + " len must be >= 3"
		resp, _ := json.Marshal(mk)
		w.Write(resp)
		return
	}
	
	
	switch params.Class {
		case "warrior":
		case "sorcerer":
		case "rouge":
		// params.Class = Class
	default:
		if params.Class != "" {
			w.WriteHeader(http.StatusBadRequest)
			mk := make(map[string]interface{})
			mk["error"] = strings.ToLower("Class") + " must be one of [warrior, sorcerer, rouge]"
			resp, _ := json.Marshal(mk)
			w.Write(resp)
			return
		}
		params.Class = "warrior"
	}
	if !(params.Level >= 1) {
		w.WriteHeader(http.StatusBadRequest)
		mk := make(map[string]interface{})
		mk["error"] = strings.ToLower("Level") + " must be >= 1"
		resp, _ := json.Marshal(mk)
		w.Write(resp)
		return
	}
	
	if !(params.Level < 50) {
		w.WriteHeader(http.StatusBadRequest)
		mk := make(map[string]interface{})
		mk["error"] = strings.ToLower("Level") + " must be <= 50"
		resp, _ := json.Marshal(mk)
		w.Write(resp)
		return
	}
	
	ctx := r.Context()

	res, err := h.Create(ctx, *params)
	if err != nil {
		e, ok := err.(ApiError)
		if ok {
			w.WriteHeader(e.HTTPStatus)
			mk := make(map[string]interface{})
			mk["error"] = err.Error()
			resp, _ := json.Marshal(mk)
			w.Write(resp)
			return
		} else {
			if err != nil && err.Error() == "bad user" {
				w.WriteHeader(http.StatusInternalServerError)
				mk := make(map[string]interface{})
				mk["error"] = err.Error()
				resp, _ := json.Marshal(mk)
				w.Write(resp)
				return
			}
		}
	}
	w.WriteHeader(http.StatusOK)
	mk := make(map[string]interface{})
	mk["error"] =	 ""
	mk["response"] = res
	resp, _ := json.Marshal(mk)
	w.Write(resp)
	
}
