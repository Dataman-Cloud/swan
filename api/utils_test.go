package api

/*
import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_writeJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	type args struct {
		w    http.ResponseWriter
		code int
		v    interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "testJSON",
			args: args{
				w:    rr,
				code: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeJSON(tt.args.w, tt.args.code, tt.args.v)
		})
	}
}

func Test_checkForJSON(t *testing.T) {
	req, _ := http.NewRequest("GET", "/health-check", nil)
	type args struct {
		req *http.Request
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "checkForJSON",
			args: args{
				req: req,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkForJSON(tt.args.req); (err != nil) != tt.wantErr {
				t.Errorf("checkForJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_decode(t *testing.T) {
	type args struct {
		b io.ReadCloser
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO add testing code
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := decode(tt.args.b, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("decode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
*/
