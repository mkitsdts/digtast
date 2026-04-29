package api

import (
	"context"
	"digital-labor/internal/center"
	"digital-labor/pkg/model"
	pb "digital-labor/proto"
)

func (s *ContainerServer) StartVirtualDesktop(ctx context.Context, req *pb.StartVirtualDesktopRequest) (*pb.StartVirtualDesktopResponse, error) {
	if center.Vdisplay == nil {
		return &pb.StartVirtualDesktopResponse{
			Success: false,
			Message: "Virtual desktop not configured",
		}, nil
	}

	resp, err := center.Vdisplay.GetOrStartVisualDisplay(model.GetDesktopDisplayRequest{
		Key:  req.AgentName,
		Kind: "vnc",
	})
	if err != nil {
		return &pb.StartVirtualDesktopResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.StartVirtualDesktopResponse{
		Success: true,
		VncPort: int32(resp.Port),
	}, nil
}

func (s *ContainerServer) StopVirtualDesktop(ctx context.Context, req *pb.StopVirtualDesktopRequest) (*pb.StopVirtualDesktopResponse, error) {
	if center.Vdisplay == nil {
		return &pb.StopVirtualDesktopResponse{
			Success: false,
			Message: "Virtual desktop not configured",
		}, nil
	}

	_, err := center.Vdisplay.ShutdownVisualDisplay(model.ShutdownDesktopDisplayRequest{
		Key:  req.AgentName,
		Kind: "vnc",
	})
	if err != nil {
		return &pb.StopVirtualDesktopResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.StopVirtualDesktopResponse{
		Success: true,
	}, nil
}
