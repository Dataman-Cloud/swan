package api

/*
import (
	"net"
	"net/http"
	"reflect"
	"sync"
	"testing"

	"github.com/Dataman-Cloud/swan/store"
	"github.com/gorilla/mux"
)

func TestNewServer(t *testing.T) {
	fakeCfg := &Config{
		Advertise: "hello",
		LogLevel:  "debug",
	}
	type args struct {
		cfg    *Config
		l      net.Listener
		leader string
		driver Driver
		db     store.Store
	}
	tests := []struct {
		name string
		args args
		want *Server
	}{
		{
			name: "test",
			args: args{
				cfg: fakeCfg,
			},
			want: &Server{
				cfg: fakeCfg,
				server: &http.Server{
					Handler: nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewServer(tt.args.cfg, tt.args.l, tt.args.driver, tt.args.db); !reflect.DeepEqual(got.cfg, tt.want.cfg) {
				t.Errorf("NewServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_Run(t *testing.T) {
	type fields struct {
		cfg      *Config
		listener net.Listener
		leader   string
		server   *http.Server
		driver   Driver
		db       store.Store
		Mutex    sync.Mutex
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
	//TODO: Do TestServer_Run
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				cfg:      tt.fields.cfg,
				listener: tt.fields.listener,
				leader:   tt.fields.leader,
				server:   tt.fields.server,
				driver:   tt.fields.driver,
				db:       tt.fields.db,
				Mutex:    tt.fields.Mutex,
			}
			if err := s.Run(); (err != nil) != tt.wantErr {
				t.Errorf("Server.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServer_Shutdown(t *testing.T) {
	type fields struct {
		cfg      *Config
		listener net.Listener
		leader   string
		server   *http.Server
		driver   Driver
		db       store.Store
		Mutex    sync.Mutex
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				cfg:      tt.fields.cfg,
				listener: tt.fields.listener,
				leader:   tt.fields.leader,
				server:   tt.fields.server,
				driver:   tt.fields.driver,
				db:       tt.fields.db,
				Mutex:    tt.fields.Mutex,
			}
			if err := s.Shutdown(); (err != nil) != tt.wantErr {
				t.Errorf("Server.Shutdown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServer_Stop(t *testing.T) {
	type fields struct {
		cfg      *Config
		listener net.Listener
		leader   string
		server   *http.Server
		driver   Driver
		db       store.Store
		Mutex    sync.Mutex
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				cfg:      tt.fields.cfg,
				listener: tt.fields.listener,
				leader:   tt.fields.leader,
				server:   tt.fields.server,
				driver:   tt.fields.driver,
				db:       tt.fields.db,
				Mutex:    tt.fields.Mutex,
			}
			if err := s.Stop(); (err != nil) != tt.wantErr {
				t.Errorf("Server.Stop() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServer_Reload(t *testing.T) {
	type fields struct {
		cfg      *Config
		listener net.Listener
		leader   string
		server   *http.Server
		driver   Driver
		db       store.Store
		Mutex    sync.Mutex
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				cfg:      tt.fields.cfg,
				listener: tt.fields.listener,
				leader:   tt.fields.leader,
				server:   tt.fields.server,
				driver:   tt.fields.driver,
				db:       tt.fields.db,
				Mutex:    tt.fields.Mutex,
			}
			if err := s.Reload(); (err != nil) != tt.wantErr {
				t.Errorf("Server.Reload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServer_UpdateLeader(t *testing.T) {
	type fields struct {
		cfg      *Config
		listener net.Listener
		leader   string
		server   *http.Server
		driver   Driver
		db       store.Store
		Mutex    sync.Mutex
	}
	type args struct {
		leader string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				cfg:      tt.fields.cfg,
				listener: tt.fields.listener,
				leader:   tt.fields.leader,
				server:   tt.fields.server,
				driver:   tt.fields.driver,
				db:       tt.fields.db,
				Mutex:    tt.fields.Mutex,
			}
			s.UpdateLeader(tt.args.leader)
		})
	}
}

func TestServer_GetLeader(t *testing.T) {
	type fields struct {
		cfg      *Config
		listener net.Listener
		leader   string
		server   *http.Server
		driver   Driver
		db       store.Store
		Mutex    sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				cfg:      tt.fields.cfg,
				listener: tt.fields.listener,
				leader:   tt.fields.leader,
				server:   tt.fields.server,
				driver:   tt.fields.driver,
				db:       tt.fields.db,
				Mutex:    tt.fields.Mutex,
			}
			if got := s.GetLeader(); got != tt.want {
				t.Errorf("Server.GetLeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_createMux(t *testing.T) {
	type fields struct {
		cfg      *Config
		listener net.Listener
		leader   string
		server   *http.Server
		driver   Driver
		db       store.Store
		Mutex    sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
		want   *mux.Router
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				cfg:      tt.fields.cfg,
				listener: tt.fields.listener,
				leader:   tt.fields.leader,
				server:   tt.fields.server,
				driver:   tt.fields.driver,
				db:       tt.fields.db,
				Mutex:    tt.fields.Mutex,
			}
			if got := s.createMux(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server.createMux() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_enableCORS(t *testing.T) {
	type fields struct {
		cfg      *Config
		listener net.Listener
		leader   string
		server   *http.Server
		driver   Driver
		db       store.Store
		Mutex    sync.Mutex
	}
	type args struct {
		w http.ResponseWriter
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				cfg:      tt.fields.cfg,
				listener: tt.fields.listener,
				leader:   tt.fields.leader,
				server:   tt.fields.server,
				driver:   tt.fields.driver,
				db:       tt.fields.db,
				Mutex:    tt.fields.Mutex,
			}
			s.enableCORS(tt.args.w)
		})
	}
}

func Test_profilerSetup(t *testing.T) {
	type args struct {
		r    *mux.Router
		path string
	}
	tests := []struct {
		name string
		args args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profilerSetup(tt.args.r, tt.args.path)
		})
	}
}

func TestServer_makeHTTPHandler(t *testing.T) {
	type fields struct {
		cfg      *Config
		listener net.Listener
		leader   string
		server   *http.Server
		driver   Driver
		db       store.Store
		Mutex    sync.Mutex
	}
	type args struct {
		handler HandlerFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   http.HandlerFunc
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				cfg:      tt.fields.cfg,
				listener: tt.fields.listener,
				leader:   tt.fields.leader,
				server:   tt.fields.server,
				driver:   tt.fields.driver,
				db:       tt.fields.db,
				Mutex:    tt.fields.Mutex,
			}
			if got := s.makeHTTPHandler(tt.args.handler); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server.makeHTTPHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_forwardRequest(t *testing.T) {
	type fields struct {
		cfg      *Config
		listener net.Listener
		leader   string
		server   *http.Server
		driver   Driver
		db       store.Store
		Mutex    sync.Mutex
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				cfg:      tt.fields.cfg,
				listener: tt.fields.listener,
				leader:   tt.fields.leader,
				server:   tt.fields.server,
				driver:   tt.fields.driver,
				db:       tt.fields.db,
				Mutex:    tt.fields.Mutex,
			}
			s.forwardRequest(tt.args.w, tt.args.r)
		})
	}
}
*/
