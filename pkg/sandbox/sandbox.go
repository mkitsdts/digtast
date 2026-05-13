package sandbox

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/cloudwego/eino/adk/filesystem"
	"github.com/cloudwego/eino/schema"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Config struct {
	Image        string // Docker image, default "debian:latest"
	HostWorkPath string // Host workspace path to bind-mount
	ContWorkPath string // Container workspace path, default "/workspace"
}

type Sandbox struct {
	cli          *client.Client
	containerID  string
	contWorkPath string
	mu           sync.Mutex
}

func New(ctx context.Context, cfg *Config) (*Sandbox, error) {
	if cfg.HostWorkPath == "" {
		return nil, fmt.Errorf("HostWorkPath is required")
	}

	img := cfg.Image
	if img == "" {
		img = "debian:latest"
	}
	contWorkPath := cfg.ContWorkPath
	if contWorkPath == "" {
		contWorkPath = "/workspace"
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	// Pull image if not present
	_, err = cli.ImageInspect(ctx, img)
	if err != nil {
		slog.Info("pulling sandbox image", "image", img)
		rc, err := cli.ImagePull(ctx, img, image.PullOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to pull image %s: %w", img, err)
		}
		io.Copy(io.Discard, rc)
		rc.Close()
	}

	// Create container with bind mount
	containerCfg := &container.Config{
		Image:      img,
		Cmd:        []string{"sleep", "infinity"},
		Tty:        false,
		WorkingDir: contWorkPath,
	}

	hostCfg := &container.HostConfig{
		Binds:        []string{fmt.Sprintf("%s:%s", cfg.HostWorkPath, contWorkPath)},
		PortBindings: nat.PortMap{},
		// Security: read-only root filesystem could be added later
	}

	resp, err := cli.ContainerCreate(ctx, containerCfg, hostCfg, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	slog.Info("sandbox container started", "id", resp.ID[:12], "image", img, "workPath", contWorkPath)

	sb := &Sandbox{
		cli:          cli,
		containerID:  resp.ID,
		contWorkPath: contWorkPath,
	}

	// Install ripgrep for GrepRaw support
	if err := sb.installRipgrep(ctx); err != nil {
		slog.Warn("failed to install ripgrep in sandbox", "error", err)
	}

	return sb, nil
}

func (s *Sandbox) installRipgrep(ctx context.Context) error {
	_, err := s.execSync(ctx, "apt-get update && apt-get install -y ripgrep > /dev/null 2>&1")
	return err
}

func (s *Sandbox) Exec(ctx context.Context, cmd string) (*schema.StreamReader[*filesystem.ExecuteResponse], error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	execCfg := container.ExecOptions{
		Cmd:          []string{"/bin/sh", "-c", cmd},
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   s.contWorkPath,
	}

	execIDResp, err := s.cli.ContainerExecCreate(ctx, s.containerID, execCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec: %w", err)
	}

	attachResp, err := s.cli.ContainerExecAttach(ctx, execIDResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to attach exec: %w", err)
	}

	sr, w := schema.Pipe[*filesystem.ExecuteResponse](100)

	go func() {
		defer func() {
			attachResp.Close()
			w.Close()
		}()

		buf := make([]byte, 4096)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			n, err := attachResp.Reader.Read(buf)
			if n > 0 {
				w.Send(&filesystem.ExecuteResponse{
					Output: string(buf[:n]),
				}, nil)
			}
			if err != nil {
				if err == io.EOF {
					// Get exit code
					inspectResp, inspectErr := s.cli.ContainerExecInspect(ctx, execIDResp.ID)
					if inspectErr == nil {
						exitCode := inspectResp.ExitCode
						w.Send(&filesystem.ExecuteResponse{
							ExitCode: &exitCode,
						}, nil)
					}
				}
				return
			}
		}
	}()

	return sr, nil
}

// execSync runs a command and waits for completion (used for setup commands).
func (s *Sandbox) execSync(ctx context.Context, cmd string) (string, error) {
	return s.execOutput(ctx, cmd)
}

// execOutput runs a command in the container and returns stdout.
func (s *Sandbox) execOutput(ctx context.Context, cmd string) (string, error) {
	execCfg := container.ExecOptions{
		Cmd:          []string{"/bin/sh", "-c", cmd},
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   s.contWorkPath,
	}

	execIDResp, err := s.cli.ContainerExecCreate(ctx, s.containerID, execCfg)
	if err != nil {
		return "", fmt.Errorf("failed to create exec: %w", err)
	}

	attachResp, err := s.cli.ContainerExecAttach(ctx, execIDResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to attach exec: %w", err)
	}
	defer attachResp.Close()

	output, _ := io.ReadAll(attachResp.Reader)

	inspectResp, err := s.cli.ContainerExecInspect(ctx, execIDResp.ID)
	if err != nil {
		return string(output), nil
	}
	if inspectResp.ExitCode != 0 {
		return string(output), fmt.Errorf("command exited with code %d", inspectResp.ExitCode)
	}
	return string(output), nil
}

// execWithStdin runs a command in the container with stdin piped.
func (s *Sandbox) execWithStdin(ctx context.Context, cmd string, stdinData string) error {
	execCfg := container.ExecOptions{
		Cmd:          []string{"/bin/sh", "-c", cmd},
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   s.contWorkPath,
	}

	execIDResp, err := s.cli.ContainerExecCreate(ctx, s.containerID, execCfg)
	if err != nil {
		return fmt.Errorf("failed to create exec: %w", err)
	}

	attachResp, err := s.cli.ContainerExecAttach(ctx, execIDResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return fmt.Errorf("failed to attach exec: %w", err)
	}
	defer attachResp.Close()

	// Write stdin and close
	attachResp.Conn.Write([]byte(stdinData))
	attachResp.CloseWrite()

	// Drain output
	io.Copy(io.Discard, attachResp.Reader)

	inspectResp, err := s.cli.ContainerExecInspect(ctx, execIDResp.ID)
	if err != nil {
		return nil
	}
	if inspectResp.ExitCode != 0 {
		return fmt.Errorf("command exited with code %d", inspectResp.ExitCode)
	}
	return nil
}

func (s *Sandbox) Close() error {
	if s.cli == nil {
		return nil
	}
	ctx := context.Background()
	timeout := 5
	err := s.cli.ContainerStop(ctx, s.containerID, container.StopOptions{Timeout: &timeout})
	if err != nil {
		slog.Error("failed to stop sandbox container", "error", err)
	}
	err = s.cli.ContainerRemove(ctx, s.containerID, container.RemoveOptions{Force: true})
	if err != nil {
		slog.Error("failed to remove sandbox container", "error", err)
	}
	slog.Info("sandbox container removed", "id", s.containerID[:12])
	return s.cli.Close()
}
