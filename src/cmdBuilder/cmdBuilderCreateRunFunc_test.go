package cmdBuilder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertArgs(t *testing.T) {
	type args struct {
		cmd  *Cmd
		args []string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string][]string
		wantErr string
	}{
		{
			name: "no args",
			args: args{
				cmd:  NewCmd(),
				args: []string{},
			},
			want:    map[string][]string{},
			wantErr: "",
		},
		{
			name: "one required arg",
			args: args{
				cmd:  (NewCmd()).Arg("arg1"),
				args: []string{"value1"},
			},
			want:    map[string][]string{"arg1": {"value1"}},
			wantErr: "",
		},
		{
			name: "one required arg, one optional arg, one value",
			args: args{
				cmd:  (NewCmd()).Arg("arg1").Arg("arg2", OptionalArg()),
				args: []string{"value1"},
			},
			want:    map[string][]string{"arg1": {"value1"}},
			wantErr: "",
		},
		{
			name: "one required arg, one optional arg, two values",
			args: args{
				cmd:  (NewCmd()).Arg("arg1").Arg("arg2", OptionalArg()),
				args: []string{"value1", "value2"},
			},
			want:    map[string][]string{"arg1": {"value1"}, "arg2": {"value2"}},
			wantErr: "",
		},
		{
			name: "one required arg, one optional array arg, one value",
			args: args{
				cmd:  (NewCmd()).Arg("arg1").Arg("arg2", ArrayArg(), OptionalArg()),
				args: []string{"value1"},
			},
			want:    map[string][]string{"arg1": {"value1"}},
			wantErr: "",
		},
		{
			name: "one required arg, one optional array arg, two values",
			args: args{
				cmd:  (NewCmd()).Arg("arg1").Arg("arg2", ArrayArg(), OptionalArg()),
				args: []string{"value1", "value2"},
			},
			want:    map[string][]string{"arg1": {"value1"}, "arg2": {"value2"}},
			wantErr: "",
		},
		{
			name: "one required arg, one array arg, two values",
			args: args{
				cmd:  (NewCmd()).Arg("arg1").Arg("arg2", ArrayArg()),
				args: []string{"value1", "value2"},
			},
			want:    map[string][]string{"arg1": {"value1"}, "arg2": {"value2"}},
			wantErr: "",
		},
		{
			name: "one required arg, one array arg, three values",
			args: args{
				cmd:  (NewCmd()).Arg("arg1").Arg("arg2", ArrayArg()),
				args: []string{"value1", "value2", "value3"},
			},
			want:    map[string][]string{"arg1": {"value1"}, "arg2": {"value2", "value3"}},
			wantErr: "",
		},
		// errors
		{
			name: "no args, one value",
			args: args{
				cmd:  NewCmd(),
				args: []string{"value1"},
			},
			want:    nil,
			wantErr: "expected no more than 0 arg(s), got 1",
		},
		{
			name: "one required arg, no value",
			args: args{
				cmd:  (NewCmd()).Arg("arg1"),
				args: []string{},
			},
			want:    nil,
			wantErr: "expected at least 1 arg(s), got 0",
		},
		{
			name: "one required arg, one optional arg, no values",
			args: args{
				cmd:  (NewCmd()).Arg("arg1").Arg("arg2", OptionalArg()),
				args: []string{},
			},
			want:    nil,
			wantErr: "expected at least 1 arg(s), got 0",
		},
		{
			name: "one required arg, one optional arg, three values",
			args: args{
				cmd:  (NewCmd()).Arg("arg1").Arg("arg2", OptionalArg()),
				args: []string{"value1", "value2", "value3"},
			},
			want:    nil,
			wantErr: "expected no more than 2 arg(s), got 3",
		},
		{
			name: "one required arg, one optional array arg, no values",
			args: args{
				cmd:  (NewCmd()).Arg("arg1").Arg("arg2", ArrayArg(), OptionalArg()),
				args: []string{},
			},
			want:    nil,
			wantErr: "expected at least 1 arg(s), got 0",
		},
		{
			name: "one required arg, one array arg, one value",
			args: args{
				cmd:  (NewCmd()).Arg("arg1").Arg("arg2", ArrayArg()),
				args: []string{"value1"},
			},
			want:    nil,
			wantErr: "expected at least 2 arg(s), got 1",
		},
		// settings errors
		{
			name: "optional arg must be the last one",
			args: args{
				cmd:  (NewCmd()).Arg("arg1", OptionalArg()).Arg("arg2"),
				args: nil,
			},
			want:    nil,
			wantErr: "optional arg arg1 can be only the last on",
		},
		{
			name: "array arg must be the last one",
			args: args{
				cmd:  (NewCmd()).Arg("arg1", ArrayArg()).Arg("arg2"),
				args: nil,
			},
			want:    nil,
			wantErr: "array arg arg1 can be only the last on",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertArgs(tt.args.cmd, tt.args.args)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}
