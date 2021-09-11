package main

import "testing"

func Test_b32encode(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{""}, "DNG3P22AK5O57EPAG3PMOIDR6U"},
		{"google", args{"https://www.google.com/search?q=zat.is"}, "HZHN3BHZTZM65P7WPH3PSML5VA"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := b32encode(tt.args.path); got != tt.want {
				t.Errorf("b32encode() = %v, want %v", got, tt.want)
			}
		})
	}
}
