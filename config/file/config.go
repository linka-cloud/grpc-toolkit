// inspired by / taken from github.com/spf13/viper

package file

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"

	"go.linka.cloud/grpc-toolkit/config"
	"go.linka.cloud/grpc-toolkit/logger"
)

func NewConfig(path string) (config.Config, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	return &file{path: path}, nil
}

type file struct {
	path string
}

func (c *file) Read() ([]byte, error) {
	return os.ReadFile(c.path)
}

// Watch listen for config changes and send updated content to the updates channel
func (c *file) Watch(ctx context.Context, updates chan<- []byte) error {
	log := logger.From(ctx)
	errs := make(chan error, 1)
	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			errs <- err
			return
		}
		defer watcher.Close()
		// we have to watch the entire directory to pick up renames/atomic saves in a cross-platform way
		configFile := filepath.Clean(c.path)
		configDir, _ := filepath.Split(configFile)
		realConfigFile, _ := filepath.EvalSymlinks(c.path)

		eventsWG := sync.WaitGroup{}
		eventsWG.Add(1)
		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok { // 'Events' channel is closed
						eventsWG.Done()
						return
					}
					currentConfigFile, _ := filepath.EvalSymlinks(c.path)
					// we only care about the config file with the following cases:
					// 1 - if the config file was modified or created
					// 2 - if the real path to the config file changed (eg: k8s ConfigMap replacement)
					const writeOrCreateMask = fsnotify.Write | fsnotify.Create
					if (filepath.Clean(event.Name) == configFile &&
						event.Op&writeOrCreateMask != 0) ||
						(currentConfigFile != "" && currentConfigFile != realConfigFile) {
						realConfigFile = currentConfigFile
						b, err := c.Read()
						if err != nil {
							log.WithError(err).Error("failed to read config")
							break
						}
						out := make([]byte, len(b))
						copy(out, b)
						updates <- out
					} else if filepath.Clean(event.Name) == configFile &&
						event.Op&fsnotify.Remove&fsnotify.Remove != 0 {
						eventsWG.Done()
						return
					}

				case err, ok := <-watcher.Errors:
					if ok { // 'Errors' channel is not closed
						log.WithError(err).Error("watcher failed")
					}
					eventsWG.Done()
					return
				case <-ctx.Done():
					return
				}
			}
		}()

		errs <- watcher.Add(configDir) // done initializing the watch in this go routine, so the parent routine can move on...
		eventsWG.Wait()                // now, wait for event loop to end in this go-routine...
	}()
	// initWG.Wait() // make sure that the go routine above fully ended before returning
	return <-errs
}
