package vnc

import (
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
)

func startVNC(key string) error {
	//TODO:还需要优化逻辑，增强适配性
	port, err := s.AddVNCService(key)
	if err != nil {
		return errors.New("no available id")
	}

	arg := fmt.Sprintf(":%d", port)
	cmd := exec.Command("vncserver", arg, "-geometry", "1920x1080")
	if err := cmd.Run(); err != nil {
		slog.Error("启动失败", "error", err)
		return err
	}
	go listenVNC(key, port, cmd)
	return nil
}

func stopVNC(port int) error {
	//TODO:还需要优化逻辑，当前启动逻辑过于死板，适配性有限
	if err := exec.Command("vncserver", "-kill", fmt.Sprintf(":%d", port)).Run(); err != nil {
		slog.Error("停止失败", "error", err)
		return err
	}
	return nil
}

func listenVNC(key string, port int, cmd *exec.Cmd) {
	// TODO：
	err := cmd.Wait()
	if err == nil {
		// normal exit
		slog.Info("VNC service exit normally", "port", port)
		return
	}

	// Error exit
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode := exitErr.ExitCode()
		//
		slog.Error("VNC service exited with error",
			"port", port,
			"exit_code", exitCode,
			"stderr", string(exitErr.Stderr),
		)

		// 可以在这里触发重启逻辑
		go startVNC(key)
	} else {
		// system error: maybe the program file does not exist, or other system issues
		slog.Error("VNC environment error", "port", port, "error", err)
	}
}
