package api

import (
	"context"
	pb "proto/digital_labor"
)

// CreateSkill 在容器里创建 Skill
func (s *ContainerServer) CreateSkill(ctx context.Context, req *pb.CreateSkillRequest) (*pb.CreateSkillResponse, error) {
	// TODO: 实现创建 Skill 逻辑，可以使用 req.Metadata
	return &pb.CreateSkillResponse{
		SkillId: "new_skill_id",
	}, nil
}

// DisableSkill 在容器里禁用 Skill
func (s *ContainerServer) DisableSkill(ctx context.Context, req *pb.DisableSkillRequest) (*pb.DisableSkillResponse, error) {
	// TODO: 实现禁用 Skill 逻辑
	return &pb.DisableSkillResponse{
		Success: true,
	}, nil
}

// RemoveSkill 在容器里移除 Skill
func (s *ContainerServer) RemoveSkill(ctx context.Context, req *pb.RemoveSkillRequest) (*pb.RemoveSkillResponse, error) {
	// TODO: 实现移除 Skill 逻辑
	return &pb.RemoveSkillResponse{
		Success: true,
	}, nil
}
