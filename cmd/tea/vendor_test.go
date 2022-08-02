package main

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GetPackages(t *testing.T) {
	type versionedPackage struct {
		packageName string
		ver         string
	}

	type test struct {
		name   string
		filter string
		line   string
		want   versionedPackage
		err    error
	}
	tests := []test{
		{
			name:   "success:process-line:regex-match",
			filter: "go.strv.io/*",
			line:   "    go.strv.io/net v1.0.0",
			want: versionedPackage{
				packName: "go.strv.io/net",
				ver:      "v1.0.0",
				found:    true,
			},
		},
		{
			name:   "success:process-line:regex-match-with-require",
			filter: "go.strv.io/*",
			line:   "require go.strv.io/net v1.0.0",
			want: versionedPackage{
				packName: "go.strv.io/net",
				ver:      "v1.0.0",
				found:    true,
			},
		},
		{
			name:   "success:process-line:regex-match-without-asterisk",
			filter: "go.strv.io",
			line:   "    go.strv.io/net v1.0.0",
			want: versionedPackage{
				packName: "go.strv.io/net",
				ver:      "v1.0.0",
				found:    true,
			},
		},
		{
			name:   "fail:process-line:regex-match",
			filter: "go.strv.io/*",
			line:   "    go.strv.fio/net v1.0.0",
			want: versionedPackage{
				packName: "",
				ver:      "",
				found:    false,
			},
		},
		{
			name:   "success:process-line:regex-match-with-exclude",
			filter: "go.strv.io/*",
			line:   "exclude go.strv.io/net v1.0.0",
			want: versionedPackage{
				packName: "go.strv.io/net",
				ver:      "v1.0.0",
				found:    true,
			},
		},
		{
			name:   "fail:process-line:not-regex-match",
			filter: "go.strv.io/*",
			line:   "    jackc/pgx v1.0.0",
			want: versionedPackage{
				packName: "",
				ver:      "",
				found:    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := regexp.Compile(tt.filter)
			if err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}

			getPackageList(re)

			gotPackage, gotVersion, found := processLine(tt.line, re)
			require.Equal(t, tt.want.found, found)
			require.Equal(t, tt.want.packName, gotPackage)
			require.Equal(t, tt.want.ver, gotVersion)
		})
	}
}
