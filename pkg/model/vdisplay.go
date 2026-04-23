package model

type GetDesktopDisplayRequest struct {
	Key string
}

type GetDesktopDisplayResponse struct {
	Port int
}

type ShutdownDesktopDisplayRequest struct {
}

type ShutdownDesktopDisplayResponse struct {
}
