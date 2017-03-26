package types

import (
	"bytes"
	"encoding/gob"
)

// defaults
const (
	ImageDeploymentTopicName             = "image_deployment_topic"
	ImageDeploymentInfoRequestsTopicName = "image_deployment_info_req_topic"
	DefaultDistributionDHTPort           = 57000
)

// ImageToLeech is used
type ImageToLeech struct {
	ID      string // Image ID, provided by Docker
	Name    string // Name string
	Labels  map[string]string
	Torrent []byte // torrent file
}

// Encode - encodes request
func (r *ImageToLeech) Encode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(r)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecodeImageToLeech - decodes Gob bytes into function struct
func DecodeImageToLeech(data []byte) (*ImageToLeech, error) {
	var r *ImageToLeech
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// ImagesToLeechBatch - images to leech batch
type ImagesToLeechBatch struct {
	Images []*ImageToLeech
}

// Encode - encodes request
func (r *ImagesToLeechBatch) Encode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(r)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecodeImagesToLeechBatch - decodes Gob bytes into batch struct
func DecodeImagesToLeechBatch(data []byte) (*ImagesToLeechBatch, error) {
	var r *ImagesToLeechBatch
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&r)
	if err != nil {
		return nil, err
	}
	return r, nil
}
