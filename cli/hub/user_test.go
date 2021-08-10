package hub

import (
	"testing"

	"github.com/docker/cli/internal/test/environment"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/skip"
)

func TestGetUserInfo(t *testing.T) {
	environment.SkipIfDaemonNotLinux(t)

	c := NewClient(testUserAgent())
	err := c.Login()
	skip.If(t, err != nil)

	ui, err := c.GetUserInfo()
	assert.NilError(t, err)
	assert.Assert(t, len(ui.Name) > 0)

	//b, _ := json.MarshalIndent(ui, "", "  ")
	//fmt.Println(string(b))
}
