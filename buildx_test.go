package dockerbuildkit

import (
	"os/exec"
	"reflect"
	"testing"
)

func TestCommandBake(t *testing.T) {
	tcs := []struct {
		name string
		bake Bake
		want *exec.Cmd
	}{
		{
			name: "first",
			bake: Bake{
				Files: []string{"docker-bake.hcl"},
			},
			want: exec.Command(
				dockerExe,
				"buildx",
				"bake",
				"--print",
				"--file",
				"docker-bake.hcl",
			),
		},
		{
			name: "tow",
			bake: Bake{
				Files:      []string{"docker-bake.hcl"},
				Provenance: "false",
				Variables: []string{
					"TAGS=1,1.4,1.4.31",
					"LABEL_CREATE_AT=1231237788",
				},
			},
			want: exec.Command(
				"TAGS=1,1.4,1.4.31",
				"LABEL_CREATE_AT=1231237788",
				dockerExe,
				"buildx",
				"bake",
				"--print",
				"--file",
				"docker-bake.hcl",
				"--provenance",
				"false",
			),
		},
	}

	for _, tc := range tcs {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {

			cmd := tc.bake.commandBakePrint(tc.bake.Variables)

			if !reflect.DeepEqual(cmd.String(), tc.want.String()) {
				t.Errorf("\nGot cmd %v, \nwant    %v", cmd, tc.want)
			}
		})
	}
}
