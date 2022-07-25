package main

import (
	"os"
	"testing"

	"go.strv.io/tea/util"

	"github.com/stretchr/testify/assert"
)

func Test_runGenerateIDs(t *testing.T) {
	cleanupAfterTest := util.CleanupAfterTest(t)
	input := "../../tests/gen/id/id.go"
	output := "../../tests/gen/id/id_gen.go"

	type args struct {
		opts *GenIDOptions
	}
	tests := []struct {
		name    string
		args    args
		cond    func()
		wantErr bool
	}{
		{
			name: "success:generate-ids",
			args: args{
				opts: &GenIDOptions{
					SourceFilePath: input,
					OutputFilePath: output,
				},
			},
			cond: func() {
				_, err := os.Stat(output)
				assert.NoError(t, err)
				if cleanupAfterTest {
					assert.NoError(t, os.Remove(output))
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := runGenerateIDs(tt.args.opts.SourceFilePath, tt.args.opts.OutputFilePath); (err != nil) != tt.wantErr {
				t.Errorf("runGenerateIDs() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.cond()
		})
	}
}
