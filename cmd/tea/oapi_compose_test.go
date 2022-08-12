package main

import (
	"os"
	"testing"

	"go.strv.io/tea/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_runOAPICompose(t *testing.T) {
	cleanupAfterTest := util.CleanupAfterTest(t)
	sourceFilePath := "../../tests/oapi/compose/v1/openapi_compose.yaml"
	outputFilePath := "../../tests/oapi/compose/v1/openapi.yaml"

	type args struct {
		opts *OAPIComposeOptions
	}
	type test struct {
		name    string
		args    args
		cond    func(t *testing.T) error
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
					SourceFilePath: sourceFilePath,
					OutputFilePath: outputFilePath,
				},
			},
			cond: func(t *testing.T) error {
				t.Helper()
				_, err := os.Stat(outputFilePath)
				require.NoError(t, err)
				defer func() {
					if cleanupAfterTest {
						require.NoError(t, os.Remove(outputFilePath))
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
			assert.NoError(t, tt.cond(t))
		})
	}
}
