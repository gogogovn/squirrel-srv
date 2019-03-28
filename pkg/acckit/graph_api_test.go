package acckit

import (
	"reflect"
	"testing"
)

func Test_getMe(t *testing.T) {
	type args struct {
		accessToken string
		secretKey   string
	}
	tests := []struct {
		name    string
		args    args
		want    *GraphAPIResponse
		wantErr bool
	}{
		{
			"Empty data should be error",
			args{
				"",
				"",
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getMe(tt.args.accessToken, tt.args.secretKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("getMe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getMe() = %v, want %v", got, tt.want)
			}
		})
	}
}
