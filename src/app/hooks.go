package main

import (
	"bitbucket"
	"builder"
	"github"
	"google"
	"io"
	"log"
	"net/http"
	"strings"
	"worker"
)

type work_loader interface {
	builder.Work
	Load(io.Reader) error
}

func perform_hook(w http.ResponseWriter, req *http.Request, ctx *Context, l work_loader) {
	body := strings.NewReader(req.FormValue("payload"))
	if err := l.Load(body); err != nil {
		log.Println("error loading hook message:", err)
		perform_status(w, ctx, http.StatusInternalServerError)
		return
	}
	worker.Schedule(l)
}

func perform_google_hook(w http.ResponseWriter, req *http.Request, ctx *Context, m *google.HookMessage) {
	if err := m.Load(req.Body); err != nil {
		log.Println("error loading hook message:", err)
		perform_status(w, ctx, http.StatusInternalServerError)
		return
	}
	switch req.FormValue(":vcs") {
	case "git":
		m.Vcs = builder.Git
	case "hg":
		m.Vcs = builder.HG
	default:
		perform_status(w, ctx, http.StatusOK)
		return
	}
	worker.Schedule(m)
}

func handle_github_hook_package(w http.ResponseWriter, req *http.Request, ctx *Context) {
	m := &github.HookMessage{Workspace: false}
	perform_hook(w, req, ctx, m)
}

func handle_github_hook_workspace(w http.ResponseWriter, req *http.Request, ctx *Context) {
	m := &github.HookMessage{Workspace: true}
	perform_hook(w, req, ctx, m)
}

func handle_bitbucket_hook_package(w http.ResponseWriter, req *http.Request, ctx *Context) {
	m := &bitbucket.HookMessage{Workspace: false}
	perform_hook(w, req, ctx, m)
}

func handle_bitbucket_hook_workspace(w http.ResponseWriter, req *http.Request, ctx *Context) {
	m := &bitbucket.HookMessage{Workspace: true}
	perform_hook(w, req, ctx, m)
}

func handle_google_hook_package(w http.ResponseWriter, req *http.Request, ctx *Context) {
	m := &google.HookMessage{Workspace: false}
	perform_google_hook(w, req, ctx, m)
}

func handle_google_hook_workspace(w http.ResponseWriter, req *http.Request, ctx *Context) {
	m := &google.HookMessage{Workspace: true}
	perform_google_hook(w, req, ctx, m)
}
