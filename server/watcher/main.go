package watcher

import (
	"log"
	"sync"

	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/message"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/pubsub"
	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watcher   *fsnotify.Watcher
	publisher pubsub.Publisher
}

type Manager interface {
	Add(name string) error
	Event() chan fsnotify.Event
	Error() chan error
	Close() error
	StartDirectoryMonitor(wg *sync.WaitGroup)
}

func NewWatcher(
	watcher *fsnotify.Watcher,
	publisher pubsub.Publisher,
) *Watcher {
	return &Watcher{
		watcher:   watcher,
		publisher: publisher,
	}
}

func (w *Watcher) Add(name string) error { return w.watcher.Add(name) }

func (w *Watcher) Event() chan fsnotify.Event { return w.watcher.Events }

func (w *Watcher) Error() chan error { return w.watcher.Errors }

func (w *Watcher) Close() error { return w.watcher.Close() }

func (w *Watcher) StartDirectoryMonitor(wg *sync.WaitGroup) {
	defer wg.Done()

	if err := w.watcher.Add("subsetdata"); err != nil {
		log.Println("error adding watcher", err.Error())
	}

	for {
		select {
		case event, ok := <-w.watcher.Events:
			switch {
			case !ok:
				return
			case event.Has(fsnotify.Create):
				if err := w.publisher.Publish(pubsub.TopicFileUpdated, message.FileUpdatedSchemaMessage{
					FileName: event.Name,
					IsCreate: true,
				}); err != nil {
					log.Println("error publishing message", err.Error())
				}
			case event.Has(fsnotify.Write):
				if err := w.publisher.Publish(pubsub.TopicFileUpdated, message.FileUpdatedSchemaMessage{
					FileName: event.Name,
					IsCreate: false,
				}); err != nil {
					log.Println("error publishing message", err.Error())
				}
			}
		case errWatcher, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Println("error watching directory", errWatcher.Error())
		}
	}
}
