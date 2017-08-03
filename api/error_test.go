package api

import "testing"

func Test_httpError_Error(t *testing.T) {
	type fields struct {
		errmsg     string
		statuscode int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "test1",
			fields: fields{"err1", 1},
			want:   "err1",
		},
		{
			name:   "test2",
			fields: fields{"err2", 1},
			want:   "err2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := httpError{
				errmsg:     tt.fields.errmsg,
				statuscode: tt.fields.statuscode,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("httpError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_httpError_StatusCode(t *testing.T) {
	type fields struct {
		errmsg     string
		statuscode int
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name:   "test1",
			fields: fields{"err1", 1},
			want:   1,
		},
		{
			name:   "test2",
			fields: fields{"err2", 1},
			want:   1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := httpError{
				errmsg:     tt.fields.errmsg,
				statuscode: tt.fields.statuscode,
			}
			if got := e.StatusCode(); got != tt.want {
				t.Errorf("httpError.StatusCode() = %v, want %v", got, tt.want)
			}
		})
	}
}
