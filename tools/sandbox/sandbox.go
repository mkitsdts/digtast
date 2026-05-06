package sandbox

import (
	"digital-labor/pkg/ctxmanager"
	"log/slog"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino-ext/components/tool/commandline/sandbox"
)

func NewSandbox(name string) *commandline.StrReplaceEditor {
	ctx := ctxmanager.GetOrCreate(name)

	op, err := sandbox.NewDockerSandbox(ctx, &sandbox.Config{})
	if err != nil {
		slog.Error("failed to create sandbox", "error", err)
		return nil
	}
	// create docker
	if err = op.Create(ctx); err != nil {
		slog.Error("failed to create docker container", "error", err)
	}
	// TODO: 需要正确时机关闭 docker 的时机
	// defer op.Cleanup(ctx)

	sre, err := commandline.NewStrReplaceEditor(ctx, &commandline.EditorConfig{Operator: op})
	if err != nil {
		slog.Error("failed to create str replace editor", "error", err)
		return nil
	}

	return sre
}
