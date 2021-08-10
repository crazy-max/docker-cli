package hub

import (
	"testing"

	"github.com/docker/cli/internal/test/environment"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/skip"
)

func TestGetRepositories(t *testing.T) {
	environment.SkipIfDaemonNotLinux(t)

	c, _ := NewClient()
	err := c.Login()
	skip.If(t, err != nil)

	repos, total, err := c.GetRepositories("")
	assert.NilError(t, err)
	assert.Assert(t, total > 0)
	assert.Assert(t, len(repos) > 0)

	//b, _ := json.MarshalIndent(repos, "", "  ")
	//fmt.Println(string(b))
}
