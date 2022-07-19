package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var cleanupAfterTest = false // os.Getenv("GO_TEST_CLEANUP") != "false"

func Test_runOAPICompose(t *testing.T) {
	type args struct {
		opts *OAPIComposeOptions
	}
	type test struct {
		name    string
		args    args
		cond    func(*test) error
		wantErr bool
	}
	tests := []test{
		{
			/*
			   @given valid config and options
			   @then OpenAPI compose file is composeed and stored into single openapi.yaml
			*/
			name: "success:compose-openapi-compose",
			args: args{
				opts: &OAPIComposeOptions{
					SourceFilePath: "../../tests/oapi/compose/v1/openapi_compose.yaml",
					OutputFilePath: "../../tests/oapi/compose/v1/openapi.yaml",
				},
			},
			cond: func(tt *test) error {
				_, err := os.Stat(tt.args.opts.OutputFilePath)
				require.NoError(t, err)

				defer func() {
					if cleanupAfterTest {
						require.NoError(t, os.Remove(tt.args.opts.OutputFilePath))
					}
				}()
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := runOAPICompose(tt.args.opts); (err != nil) != tt.wantErr {
				t.Errorf("runOAPICompose() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NoError(t, tt.cond(&tt))
		})
	}
}
