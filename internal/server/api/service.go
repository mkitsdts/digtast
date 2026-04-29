package api

import (
	"archive/zip"
	"digital-labor/pkg/workspace"
	pb "digital-labor/proto"
	"io"
	"os"
	"path/filepath"
)

// BackupService 备份服务
func (s *ContainerServer) BackupService(req *pb.BackupServiceRequest, stream pb.ContainerService_BackupServiceServer) error {
	wp := workspace.GetWorkspacePath()
	
	// Create a temporary zip file
	tmpFile, err := os.CreateTemp("", "backup-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	zw := zip.NewWriter(tmpFile)
	err = filepath.Walk(wp, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(wp, path)
		if err != nil {
			return err
		}
		w, err := zw.Create(relPath)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		return err
	})
	if err != nil {
		return err
	}
	zw.Close()

	// Stream the zip file back
	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return err
	}

	buf := make([]byte, 1024*64) // 64KB chunks
	for {
		n, err := tmpFile.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := stream.Send(&pb.BackupServiceResponse{
			Data: buf[:n],
		}); err != nil {
			return err
		}
	}

	return nil
}
