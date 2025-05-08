package handlers

import (
	"testing"
)

func TestIsUrl(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// Gyldige URL-er
		{"Valid URL with domain", args{"http://google.com"}, true},
		{"Valid URL with subdomain", args{"http://sub.google.com"}, true},
		{"Valid URL with IP", args{"http://192.168.0.1"}, true},
		{"Valid URL with IP and port", args{"http://192.168.0.1:8080"}, true},
		{"Valid URL with path", args{"http://google.com/search"}, true},
		{"Invalid IP URL", args{"http://192.168.1.300"}, true},

		// Ugyldige URL-er
		{"Invalid URL without schema", args{"google.com"}, false},
		{"Invalid URL with missing TLD", args{"http://google"}, false},
		{"Invalid URL with special characters", args{"http://!@#$%^&*()"}, false},

		{"Empty string", args{""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsURL(tt.args.str); got != tt.want {
				t.Errorf("IsUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestRedirect(t *testing.T) {
// 	type args struct {
// 		rdb *redis.Client
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want http.HandlerFunc
// 	}{
// 		{},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := Redirect(tt.args.rdb); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("Redirect() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestDeleteRedirect(t *testing.T) {
// 	type args struct {
// 		rdb *redis.Client
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want http.HandlerFunc
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := DeleteRedirect(tt.args.rdb); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("DeleteRedirect() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
