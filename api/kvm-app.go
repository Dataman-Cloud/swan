package api

import (
	"net/http"
)

// Kvm App
//
//

func (r *Server) listKvmApps(w http.ResponseWriter, req *http.Request) {
}

func (r *Server) createKvmApp(w http.ResponseWriter, req *http.Request) {
}

func (r *Server) getKvmApp(w http.ResponseWriter, req *http.Request) {
}

func (r *Server) deleteKvmApp(w http.ResponseWriter, req *http.Request) {
}

func (r *Server) startKvmApp(w http.ResponseWriter, req *http.Request) {
}

func (r *Server) stopKvmApp(w http.ResponseWriter, req *http.Request) {
}

func (r *Server) suspendKvmApp(w http.ResponseWriter, req *http.Request) {
}

func (r *Server) resumeKvmApp(w http.ResponseWriter, req *http.Request) {
}

// Kvm Task
//
//
func (r *Server) getKvmTasks(w http.ResponseWriter, req *http.Request) {
}

func (r *Server) getKvmTask(w http.ResponseWriter, req *http.Request) {
}

func (r *Server) deleteKvmTask(w http.ResponseWriter, req *http.Request) {
}

func (r *Server) startKvmTask(w http.ResponseWriter, req *http.Request) {
}

func (r *Server) stopKvmTask(w http.ResponseWriter, req *http.Request) {
}

func (r *Server) suspendKvmTask(w http.ResponseWriter, req *http.Request) {
}

func (r *Server) resumeKvmTask(w http.ResponseWriter, req *http.Request) {
}
