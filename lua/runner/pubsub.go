package runner

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/nais/tester/lua/spec"
	lua "github.com/yuin/gopher-lua"
)

type PubSubTopic struct {
	Sent     []PubSubMessage
	Received []PubSubMessage
}

type PubSubMessage struct {
	Msg        map[string]any    `json:"msg"`
	Attributes map[string]string `json:"attributes"`
}

type PubSubHook func(topic string, msg PubSubMessage) error

type PubSub struct {
	lock      sync.Mutex
	topics    map[string]PubSubTopic
	doPublish PubSubHook
}

var _ spec.Runner = (*PubSub)(nil)

func NewPubSub(doPublish PubSubHook) *PubSub {
	return &PubSub{
		doPublish: doPublish,
	}
}

func (p *PubSub) Name() string {
	return "pubsub"
}

func (g *PubSub) Functions() []*spec.Function {
	return []*spec.Function{
		{
			Name: "check",
			Args: []spec.Argument{
				{
					Name: "topic",
					Type: []spec.ArgumentType{spec.ArgumentTypeString},
					Doc:  "The topic to check",
				},
				{
					Name: "resp",
					Type: []spec.ArgumentType{spec.ArgumentTypeTable},
					Doc:  "The message to check for. Must match both data and attributes",
				},
			},
			Doc:  "Check comment",
			Func: g.check,
		},
	}
}

func (g *PubSub) HelperFunctions() []*spec.Function {
	return []*spec.Function{
		{
			Name: "emptyPubSubTopic",
			Args: []spec.Argument{
				{
					Name: "topic",
					Type: []spec.ArgumentType{spec.ArgumentTypeString},
					Doc:  "The topic to check",
				},
			},
			Doc:  "Check comment",
			Func: g.emptyTopic,
		},
	}
}

func (r *PubSub) check(L *lua.LState) int {
	topic := L.CheckString(1)
	tbl := L.CheckTable(2)

	if !r.hasTopic(topic) {
		L.RaiseError("topic %q not registered, has: %v", topic, r.topicsNames())
	}

	msgs := r.messages(topic)
	if len(msgs) == 0 {
		L.RaiseError("no messages received on topic %q", topic)
	}

	var errs []string
	for _, msg := range msgs {
		target := map[string]any{}
		b := map[string]any{
			"data":       msg.Msg,
			"attributes": msg.Attributes,
		}
		bs, _ := json.Marshal(b)
		_ = json.Unmarshal(bs, &target)

		if err := StdCheckError(L.Context(), tbl, target); err != nil {
			errs = append(errs, err.Error())
		} else {
			return 0
		}
	}

	if len(errs) > 0 {
		L.RaiseError("%v", strings.Join(errs, "\n"))
	}

	L.RaiseError("no matching messages received on topic %q", topic)
	return 0
}

func (r *PubSub) emptyTopic(L *lua.LState) int {
	topic := L.CheckString(1)

	r.lock.Lock()
	defer r.lock.Unlock()

	r.topics[topic] = PubSubTopic{}
	return 0
}

func (p *PubSub) Send(topic string, msg PubSubMessage) {
	p.lock.Lock()
	defer p.lock.Unlock()

	t, ok := p.topics[topic]
	if !ok {
		t = PubSubTopic{}
	}

	t.Sent = append(t.Sent, msg)

	if p.topics == nil {
		p.topics = map[string]PubSubTopic{}
	}
	p.topics[topic] = t
}

func (p *PubSub) Receive(topic string, msg PubSubMessage) {
	p.lock.Lock()
	defer p.lock.Unlock()

	t, ok := p.topics[topic]
	if !ok {
		t = PubSubTopic{}
	}

	t.Received = append(t.Received, msg)
	if p.topics == nil {
		p.topics = map[string]PubSubTopic{}
	}
	p.topics[topic] = t
}

func (p *PubSub) hasTopic(name string) bool {
	p.lock.Lock()
	defer p.lock.Unlock()

	_, ok := p.topics[name]
	return ok
}

func (p *PubSub) topicsNames() []string {
	p.lock.Lock()
	defer p.lock.Unlock()

	names := []string{}
	for k := range p.topics {
		names = append(names, k)
	}

	return names
}

func (p *PubSub) messages(topic string) []PubSubMessage {
	p.lock.Lock()
	defer p.lock.Unlock()

	t, ok := p.topics[topic]
	if !ok {
		return nil
	}

	return t.Received
}
