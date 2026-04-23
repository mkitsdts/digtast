package model

type GetDesktopDisplayRequest struct {
	Key  string
	Kind string
}

type GetDesktopDisplayResponse struct {
	Port int
}

type ShutdownDesktopDisplayRequest struct {
	Kind string
}

type ShutdownDesktopDisplayResponse struct {
}
