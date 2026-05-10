package api

import (
	"context"
	"digital-labor/pkg/skill"
	pb "digital-labor/proto"
)

func (s *ContainerServer) AddSkill(ctx context.Context, req *pb.AddSkillRequest) (*pb.AddSkillResponse, error) {
	mag := skill.NewManagerForAgent(req.AgentId)
	err := mag.AddSkill(req.Name, req.Description)
	if err != nil {
		return &pb.AddSkillResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &pb.AddSkillResponse{
		Success: true,
	}, nil
}

func (s *ContainerServer) GetAllSkills(ctx context.Context, req *pb.GetAllSkillsRequest) (*pb.GetAllSkillsResponse, error) {
	mag := skill.NewManagerForAgent(req.AgentId)
	skills, err := mag.GetAllSkills()
	if err != nil {
		return nil, err
	}

	res := make([]*pb.SkillInfo, 0, len(skills))
	for _, sk := range skills {
		res = append(res, &pb.SkillInfo{
			Name:        sk.Name,
			Description: sk.Description,
			Enabled:     sk.Enabled,
		})
	}

	return &pb.GetAllSkillsResponse{
		Skills: res,
	}, nil
}

func (s *ContainerServer) DisableSkill(ctx context.Context, req *pb.DisableSkillRequest) (*pb.DisableSkillResponse, error) {
	mag := skill.NewManagerForAgent(req.AgentId)
	err := mag.DisableSkill(req.SkillName)
	if err != nil {
		return &pb.DisableSkillResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &pb.DisableSkillResponse{
		Success: true,
	}, nil
}

func (s *ContainerServer) EnableSkill(ctx context.Context, req *pb.EnableSkillRequest) (*pb.EnableSkillResponse, error) {
	mag := skill.NewManagerForAgent(req.AgentId)
	err := mag.EnableSkill(req.SkillName)
	if err != nil {
		return &pb.EnableSkillResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &pb.EnableSkillResponse{
		Success: true,
	}, nil
}
