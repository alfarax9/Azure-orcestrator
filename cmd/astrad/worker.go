package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/alfarax9/astra/internal/domain"
)

// workerView is a worker with the status people see, which is derived from its facts.
type workerView struct {
	domain.Worker
	Status string `json:"status"`
}

// agentCommand runs the commands that act on a worker's agent. They are one-shot, except
// watch, which stays attached and keeps the worker's facts true until the agent exits; the
// daemon will do that itself once it has an API.
func agentCommand(ctx context.Context, a app, command string, args []string) (any, error) {
	if len(args) == 0 {
		return nil, errUsage
	}
	id := domain.WorkerID(args[0])
	switch command {
	case "worker start":
		return done(a.ag.Start(ctx, id), "started", id)
	case "worker stop":
		return done(a.ag.Stop(ctx, id), "stopped", id)
	case "worker interrupt":
		return done(a.ag.Interrupt(ctx, id), "interrupted", id)
	case "worker send":
		if len(args) != 2 {
			return nil, errUsage
		}
		return done(a.ag.Send(ctx, id, args[1]), "sent", id)
	case "worker press":
		if len(args) != 2 {
			return nil, errUsage
		}
		keys, err := strconv.Unquote(`"` + args[1] + `"`)
		if err != nil {
			return nil, fmt.Errorf("keys: %w", err)
		}
		return done(a.ag.Press(ctx, id, keys), "pressed", id)
	case "worker show":
		w, err := a.store.Worker(ctx, id)
		return workerView{Worker: w, Status: string(w.Status())}, err
	case "worker watch":
		code, err := a.ag.Watch(ctx, id)
		return map[string]any{"worker": id, "exited": err == nil && ctx.Err() == nil, "exitCode": code}, err
	}
	return nil, errUsage
}

func done(err error, what string, id domain.WorkerID) (any, error) {
	if err != nil {
		return nil, err
	}
	return map[string]any{what: id}, nil
}
