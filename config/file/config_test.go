// inspired by / taken from github.com/spf13/viper

package file

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.linka.cloud/grpc/config"
)

func newConfigFile(t *testing.T) (config.Config, string, func()){
	path := filepath.Join(os.TempDir(), "config.yaml")
	if err := ioutil.WriteFile(path, []byte("ok"), os.ModePerm); err != nil {
		t.Fatal(err)
	}
	cleanUp := func() {
		if err := os.Remove(path); err != nil {
			t.Error(err)
		}
	}
	return &file{path: path}, path, cleanUp
}

func newSymlinkedConfigFile(t *testing.T) (config.Config, string, string, func()) {
	watchDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)
	dataDir1 := path.Join(watchDir, "data1")
	err = os.Mkdir(dataDir1, 0o777)
	require.Nil(t, err)
	realConfigFile := path.Join(dataDir1, "config.yaml")
	t.Logf("Real config file location: %s\n", realConfigFile)
	err = ioutil.WriteFile(realConfigFile, []byte("foo: bar\n"), 0o640)
	require.Nil(t, err)
	cleanup := func() {
		os.RemoveAll(watchDir)
	}
	// now, symlink the tm `data1` dir to `data` in the baseDir
	os.Symlink(dataDir1, path.Join(watchDir, "data"))
	// and link the `<watchdir>/datadir1/config.yaml` to `<watchdir>/config.yaml`
	configFile := path.Join(watchDir, "config.yaml")
	os.Symlink(path.Join(watchDir, "data", "config.yaml"), configFile)
	path := path.Join(watchDir, "config.yaml")
	t.Logf("Config file location: %s\n", path)
	return &file{path: path}, watchDir, configFile, cleanup
}

func TestWatch(t *testing.T) {
	t.Run("file content changed", func(t *testing.T) {
		// given a `config.yaml` file being watched
		v, cpath, cleanup := newConfigFile(t)
		defer cleanup()
		updates := make(chan []byte, 1)
		if err := v.Watch(context.Background(), updates); err != nil {
			t.Fatal(err)
		}
		// when overwriting the file and waiting for the custom change notification handler to be triggered
		err := ioutil.WriteFile(cpath, []byte("foo: baz\n"), 0o640)
		b := <- updates
		// then the config value should have changed
		require.Nil(t, err)
		assert.Equal(t, []byte("foo: baz\n"), b)
	})

	t.Run("link to real file changed (à la Kubernetes)", func(t *testing.T) {
		// skip if not executed on Linux
		if runtime.GOOS != "linux" {
			t.Skipf("Skipping test as symlink replacements don't work on non-linux environment...")
		}
		v, watchDir, _, cleanup := newSymlinkedConfigFile(t)
		defer cleanup()
		updates := make(chan []byte, 1)
		if err := v.Watch(context.Background(), updates); err != nil {
			t.Fatal(err)
		}
		// when link to another `config.yaml` file
		dataDir2 := path.Join(watchDir, "data2")
		err := os.MkdirAll(dataDir2, 0o777)
		require.NoError(t, err)
		configFile2 := path.Join(dataDir2, "config.yaml")
		err = ioutil.WriteFile(configFile2, []byte("foo: baz\n"), 0o640)
		require.NoError(t, err)
		// change the symlink using the `ln -sfn` command
		err = exec.Command("ln", "-sfn", dataDir2, path.Join(watchDir, "data")).Run()
		require.NoError(t, err)
		b := <-updates
		assert.Equal(t, []byte("foo: baz\n"), b)
	})
}
