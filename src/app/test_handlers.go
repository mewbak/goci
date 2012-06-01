package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func build_url_pair(host, id string) (req, res, err string) {
	req = fmt.Sprintf("http://%s%s", host, reverse("test_request", "id", id))
	res = fmt.Sprintf("http://%s%s", host, reverse("test_response", "id", id))
	err = fmt.Sprintf("http://%s%s", host, reverse("test_error", "id", id))
	return
}

func handle_test_request(w http.ResponseWriter, req *http.Request, ctx *Context) {
	active_tests_lock.RLock()
	defer active_tests_lock.RUnlock()

	id := req.FormValue(":id")
	test, ex := active_tests[id]
	if !ex {
		log.Printf("test id not found: %q", id)
		perform_status(w, ctx, http.StatusNotFound)
		return
	}
	f, err := os.Open(test.Path)
	if err != nil {
		log.Printf("%s: couldn't open path: %s", test.WholeID(), err)
		perform_status(w, ctx, http.StatusNotFound)
		return
	}
	defer f.Close()
	_, err = io.Copy(w, f)
	if err != nil {
		log.Printf("%s: error copying binary: %s", test.WholeID(), err)
		return
	}

	test.Start()
}

func handle_test_response(w http.ResponseWriter, req *http.Request, ctx *Context) {
	active_tests_lock.RLock()
	defer active_tests_lock.RUnlock()

	id := req.URL.Query().Get(":id")
	test, ex := active_tests[id]
	if !ex {
		log.Printf("test id not found: %q", id)
		perform_status(w, ctx, http.StatusNotFound)
		return
	}
	by, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("error reading response data: %v", err)
		test.Error = err.Error()
		perform_status(w, ctx, http.StatusInternalServerError)
		return
	}
	s := string(by)
	test.Output = s
	test.Passed = strings.HasSuffix(s, "\nPASS\n")
	test.Finish()

	test_complete <- id
}

func handle_test_error(w http.ResponseWriter, req *http.Request, ctx *Context) {
	active_tests_lock.RLock()
	defer active_tests_lock.RUnlock()

	id := req.URL.Query().Get(":id")
	test, ex := active_tests[id]
	if !ex {
		log.Printf("test id not found: %q", id)
		perform_status(w, ctx, http.StatusNotFound)
		return
	}
	by, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("error reading response data: %v", err)
		test.Error = err.Error()
		perform_status(w, ctx, http.StatusInternalServerError)
		return
	}
	test.Error = string(by)
	test.Finish()

	test_complete <- id
}
