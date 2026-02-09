package cli

import (
	"testing"

	"github.com/johnqtcg/issue2md/internal/config"
)

func TestValidateArgs(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name    string
		cfg     config.Config
		want    Args
		wantErr bool
	}{
		{
			name: "single mode valid",
			cfg: config.Config{
				Positional: []string{"https://github.com/octo/repo/issues/1"},
			},
			want: Args{
				Mode: ModeSingle,
				URL:  "https://github.com/octo/repo/issues/1",
			},
		},
		{
			name:    "single mode missing url",
			cfg:     config.Config{},
			wantErr: true,
		},
		{
			name: "single mode too many urls",
			cfg: config.Config{
				Positional: []string{"u1", "u2"},
			},
			wantErr: true,
		},
		{
			name: "batch mode valid",
			cfg: config.Config{
				InputFile:  "urls.txt",
				OutputPath: "out",
			},
			want: Args{
				Mode: ModeBatch,
			},
		},
		{
			name: "batch mode requires output path",
			cfg: config.Config{
				InputFile: "urls.txt",
			},
			wantErr: true,
		},
		{
			name: "batch mode cannot accept positional url",
			cfg: config.Config{
				InputFile:  "urls.txt",
				OutputPath: "out",
				Positional: []string{"https://github.com/octo/repo/issues/1"},
			},
			wantErr: true,
		},
		{
			name: "batch mode cannot use stdout",
			cfg: config.Config{
				InputFile:  "urls.txt",
				OutputPath: "out",
				Stdout:     true,
			},
			wantErr: true,
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ValidateArgs(tc.cfg)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ValidateArgs error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateArgs error = %v, want nil", err)
			}
			if got != tc.want {
				t.Fatalf("ValidateArgs = %#v, want %#v", got, tc.want)
			}
		})
	}
}
