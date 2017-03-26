package reusable

import (
	docker "github.com/fsouza/go-dockerclient"
)

// GetImageID - gets image ID from name
func GetImageID(image string, client *docker.Client) (id string, err error) {
	img, err := client.InspectImage(image)
	if err != nil {
		return
	}
	id = img.ID
	return
}
