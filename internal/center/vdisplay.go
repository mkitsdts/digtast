package center

import "digital-labor/pkg/model"

type VDisplayParams struct {
	Key  string
	Kind string
}

func (c *Center) StartVDisplay(params *VDisplayParams) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	resp, err := c.visualDisplay.GetOrStartVisualDisplay(model.GetDesktopDisplayRequest{
		Key:  params.Key,
		Kind: params.Kind,
	})
	if err != nil {
		return -1, err
	}
	return resp.Port, nil
}

func (c *Center) StopVDisplay(params *VDisplayParams) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.visualDisplay.ShutdownVisualDisplay(model.ShutdownDesktopDisplayRequest{
		Key:  params.Key,
		Kind: params.Kind,
	})
	if err != nil {
		return err
	}
	return nil
}
