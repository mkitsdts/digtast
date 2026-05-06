package grpc

import (
	"testing"

	pb "digital-labor/proto"
)

func TestSkillMethods(t *testing.T) {
	ctx := getContext()
	skillName := "test-skill"

	t.Run("AddSkill", func(t *testing.T) {
		resp, err := client.AddSkill(ctx, &pb.AddSkillRequest{
			Name:        skillName,
			Description: "test skill description",
		})
		if err != nil {
			t.Fatalf("AddSkill failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success, got false: %s", resp.Message)
		}
	})

	t.Run("GetAllSkills", func(t *testing.T) {
		resp, err := client.GetAllSkills(ctx, &pb.GetAllSkillsRequest{})
		if err != nil {
			t.Fatalf("GetAllSkills failed: %v", err)
		}
		found := false
		for _, s := range resp.Skills {
			if s.Name == skillName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("skill %s not found in list", skillName)
		}
	})

	t.Run("DisableSkill", func(t *testing.T) {
		resp, err := client.DisableSkill(ctx, &pb.DisableSkillRequest{
			SkillName: skillName,
		})
		if err != nil {
			t.Fatalf("DisableSkill failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success, got false: %s", resp.Message)
		}
	})

	t.Run("EnableSkill", func(t *testing.T) {
		resp, err := client.EnableSkill(ctx, &pb.EnableSkillRequest{
			SkillName: skillName,
		})
		if err != nil {
			t.Fatalf("EnableSkill failed: %v", err)
		}
		if !resp.Success {
			t.Errorf("expected success, got false: %s", resp.Message)
		}
	})
}
