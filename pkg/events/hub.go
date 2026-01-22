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

type FramePromptTaskHub struct {
	mu   sync.RWMutex
	subs map[string]*framePromptSubscriber
}

type framePromptSubscriber struct {
	ch     chan *models.FramePromptTask
	filter func(*models.FramePromptTask) bool
}

func NewFramePromptTaskHub() *FramePromptTaskHub {
	return &FramePromptTaskHub{
		subs: make(map[string]*framePromptSubscriber),
	}
}

func (h *FramePromptTaskHub) Subscribe(filter func(*models.FramePromptTask) bool) (string, <-chan *models.FramePromptTask, func()) {
	id := uuid.New().String()
	ch := make(chan *models.FramePromptTask, 20)

	h.mu.Lock()
	h.subs[id] = &framePromptSubscriber{ch: ch, filter: filter}
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

func (h *FramePromptTaskHub) Publish(task *models.FramePromptTask) {
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

type VideoMergeHub struct {
	mu   sync.RWMutex
	subs map[string]*videoMergeSubscriber
}

type videoMergeSubscriber struct {
	ch     chan *models.VideoMerge
	filter func(*models.VideoMerge) bool
}

func NewVideoMergeHub() *VideoMergeHub {
	return &VideoMergeHub{
		subs: make(map[string]*videoMergeSubscriber),
	}
}

func (h *VideoMergeHub) Subscribe(filter func(*models.VideoMerge) bool) (string, <-chan *models.VideoMerge, func()) {
	id := uuid.New().String()
	ch := make(chan *models.VideoMerge, 20)

	h.mu.Lock()
	h.subs[id] = &videoMergeSubscriber{ch: ch, filter: filter}
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

func (h *VideoMergeHub) Publish(merge *models.VideoMerge) {
	if merge == nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, sub := range h.subs {
		if sub.filter != nil && !sub.filter(merge) {
			continue
		}
		mergeCopy := *merge
		select {
		case sub.ch <- &mergeCopy:
		default:
		}
	}
}

type EventHub struct {
	Tasks             *TaskHub
	ImageGenerations  *ImageGenerationHub
	VideoGenerations  *VideoGenerationHub
	FramePromptTasks  *FramePromptTaskHub
	VideoMerges       *VideoMergeHub
}

func NewEventHub() *EventHub {
	return &EventHub{
		Tasks:            NewTaskHub(),
		ImageGenerations: NewImageGenerationHub(),
		VideoGenerations: NewVideoGenerationHub(),
		FramePromptTasks: NewFramePromptTaskHub(),
		VideoMerges:      NewVideoMergeHub(),
	}
}
