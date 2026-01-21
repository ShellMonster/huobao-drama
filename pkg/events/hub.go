package events

import (
	"sync"

	"github.com/drama-generator/backend/domain/models"
	"github.com/google/uuid"
)

type TaskHub struct {
	mu   sync.RWMutex
	subs map[string]*taskSubscriber
}

type taskSubscriber struct {
	ch     chan *models.AsyncTask
	filter func(*models.AsyncTask) bool
}

func NewTaskHub() *TaskHub {
	return &TaskHub{
		subs: make(map[string]*taskSubscriber),
	}
}

func (h *TaskHub) Subscribe(filter func(*models.AsyncTask) bool) (string, <-chan *models.AsyncTask, func()) {
	id := uuid.New().String()
	ch := make(chan *models.AsyncTask, 20)

	h.mu.Lock()
	h.subs[id] = &taskSubscriber{ch: ch, filter: filter}
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		if sub, ok := h.subs[id]; ok {
			delete(h.subs, id)
			close(sub.ch)
		}
		h.mu.Unlock()
	}

	return id, ch, unsubscribe
}

func (h *TaskHub) Publish(task *models.AsyncTask) {
	if task == nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, sub := range h.subs {
		if sub.filter != nil && !sub.filter(task) {
			continue
		}
		taskCopy := *task
		select {
		case sub.ch <- &taskCopy:
		default:
		}
	}
}

type ImageGenerationHub struct {
	mu   sync.RWMutex
	subs map[string]*imageSubscriber
}

type imageSubscriber struct {
	ch     chan *models.ImageGeneration
	filter func(*models.ImageGeneration) bool
}

func NewImageGenerationHub() *ImageGenerationHub {
	return &ImageGenerationHub{
		subs: make(map[string]*imageSubscriber),
	}
}

func (h *ImageGenerationHub) Subscribe(filter func(*models.ImageGeneration) bool) (string, <-chan *models.ImageGeneration, func()) {
	id := uuid.New().String()
	ch := make(chan *models.ImageGeneration, 20)

	h.mu.Lock()
	h.subs[id] = &imageSubscriber{ch: ch, filter: filter}
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		if sub, ok := h.subs[id]; ok {
			delete(h.subs, id)
			close(sub.ch)
		}
		h.mu.Unlock()
	}

	return id, ch, unsubscribe
}

func (h *ImageGenerationHub) Publish(imageGen *models.ImageGeneration) {
	if imageGen == nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, sub := range h.subs {
		if sub.filter != nil && !sub.filter(imageGen) {
			continue
		}
		imageCopy := *imageGen
		select {
		case sub.ch <- &imageCopy:
		default:
		}
	}
}

type VideoGenerationHub struct {
	mu   sync.RWMutex
	subs map[string]*videoSubscriber
}

type videoSubscriber struct {
	ch     chan *models.VideoGeneration
	filter func(*models.VideoGeneration) bool
}

func NewVideoGenerationHub() *VideoGenerationHub {
	return &VideoGenerationHub{
		subs: make(map[string]*videoSubscriber),
	}
}

func (h *VideoGenerationHub) Subscribe(filter func(*models.VideoGeneration) bool) (string, <-chan *models.VideoGeneration, func()) {
	id := uuid.New().String()
	ch := make(chan *models.VideoGeneration, 20)

	h.mu.Lock()
	h.subs[id] = &videoSubscriber{ch: ch, filter: filter}
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		if sub, ok := h.subs[id]; ok {
			delete(h.subs, id)
			close(sub.ch)
		}
		h.mu.Unlock()
	}

	return id, ch, unsubscribe
}

func (h *VideoGenerationHub) Publish(videoGen *models.VideoGeneration) {
	if videoGen == nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, sub := range h.subs {
		if sub.filter != nil && !sub.filter(videoGen) {
			continue
		}
		videoCopy := *videoGen
		select {
		case sub.ch <- &videoCopy:
		default:
		}
	}
}

type EventHub struct {
	Tasks            *TaskHub
	ImageGenerations *ImageGenerationHub
	VideoGenerations *VideoGenerationHub
}

func NewEventHub() *EventHub {
	return &EventHub{
		Tasks:            NewTaskHub(),
		ImageGenerations: NewImageGenerationHub(),
		VideoGenerations: NewVideoGenerationHub(),
	}
}
