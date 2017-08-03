package api

import (
	"reflect"
	"testing"
)

func TestRoute_Methods(t *testing.T) {
	type fields struct {
		method  string
		path    string
		handler HandlerFunc
		prefix  bool
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "test1",
			fields: fields{
				method: "GET",
				path:   "/test/abcd",
				prefix: false,
			},
			want: []string{"GET"},
		},
		{
			name: "test1",
			fields: fields{
				method: "PATCH",
				path:   "/test/abcd",
				prefix: false,
			},
			want: []string{"PATCH"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Route{
				method:  tt.fields.method,
				path:    tt.fields.path,
				handler: tt.fields.handler,
				prefix:  tt.fields.prefix,
			}
			if got := r.Methods(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Route.Methods() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoute_Path(t *testing.T) {
	type fields struct {
		method  string
		path    string
		handler HandlerFunc
		prefix  bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test1",
			fields: fields{
				method: "GET",
				path:   "/test/abcd",
				prefix: false,
			},
			want: "/test/abcd",
		},
		{
			name: "test1",
			fields: fields{
				method: "GET",
				path:   "/test/abd",
				prefix: false,
			},
			want: "/test/abd",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Route{
				method:  tt.fields.method,
				path:    tt.fields.path,
				handler: tt.fields.handler,
				prefix:  tt.fields.prefix,
			}
			if got := r.Path(); got != tt.want {
				t.Errorf("Route.Path() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoute_Handler(t *testing.T) {
	type fields struct {
		method  string
		path    string
		handler HandlerFunc
		prefix  bool
	}
	tests := []struct {
		name   string
		fields fields
		want   HandlerFunc
	}{
		{
			name: "test1",
			fields: fields{
				method: "GET",
				path:   "/test/abd",
				prefix: false,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Route{
				method:  tt.fields.method,
				path:    tt.fields.path,
				handler: tt.fields.handler,
				prefix:  tt.fields.prefix,
			}
			if got := r.Handler(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Route.Handler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewRoute(t *testing.T) {
	type args struct {
		method  string
		path    string
		handler HandlerFunc
	}
	tests := []struct {
		name string
		args args
		want *Route
	}{
		{
			name: "test",
			args: args{
				method: "GET",
				path:   "/new/abcd",
			},
			want: &Route{
				method: "GET",
				path:   "/new/abcd",
				prefix: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRoute(tt.args.method, tt.args.path, tt.args.handler); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRoute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewPrefixRoute(t *testing.T) {
	type args struct {
		method  string
		path    string
		handler HandlerFunc
	}
	tests := []struct {
		name string
		args args
		want *Route
	}{
		{
			name: "test",
			args: args{
				method: "GET",
				path:   "/new/abcd",
			},
			want: &Route{
				method: "GET",
				path:   "/new/abcd",
				prefix: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPrefixRoute(tt.args.method, tt.args.path, tt.args.handler); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPrefixRoute() = %v, want %v", got, tt.want)
			}
		})
	}
}
