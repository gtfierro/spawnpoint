package spawnctl

import (
	"errors"
	"time"

	"github.com/immesys/spawnpoint/objects"

	bw2 "gopkg.in/immesys/bw2bind.v1"
	yaml "gopkg.in/yaml.v2"
)

type SpawnClient struct {
	bwClient *bw2.BW2Client
	pac      string
}

func FromHeartbeat(msg *bw2.SimpleMessage) (*objects.SpawnPoint, error) {
	for _, po := range msg.POs {
		if po.IsTypeDF(bw2.PODFSpawnpointHeartbeat) {
			hb := objects.SpawnPointHb{}
			po.(bw2.YAMLPayloadObject).ValueInto(&hb)

			uri := msg.URI[:len(msg.URI)-len("info/spawn/!heartbeat")]
			seen, err := time.Parse(time.RFC3339, hb.Time)
			if err != nil {
				return nil, err
			}

			sp := objects.SpawnPoint{uri, seen, hb.Alias, hb.AvailableCpuShares, hb.AvailableMem}
			return &sp, nil
		}
	}

	err := errors.New("Heartbeat contained no valid payload object")
	return nil, err
}

func (client *SpawnClient) DeployService(spawnPoint *objects.SpawnPoint, config *objects.SvcConfig) error {
	rawYaml, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	configPo, err := bw2.LoadYAMLPayloadObject(bw2.PONumSpawnpointConfig, rawYaml)
	if err != nil {
		return err
	}

	// Send service config to spawnpoint
	uri := spawnPoint.URI + "ctl/cfg"
	err = client.bwClient.Publish(&bw2.PublishParams{
		URI:                uri,
		PayloadObjects:     []bw2.PayloadObject{configPo},
		PrimaryAccessChain: client.pac,
		ElaboratePAC:       bw2.ElaborateFull,
	})
	if err != nil {
		return err
	}

	// Instruct spawnpoint to launch service
	uri = spawnPoint.URI + "ctl/restart"
	namePo := bw2.CreateStringPayloadObject(config.ServiceName)

	return client.bwClient.Publish(&bw2.PublishParams{
		URI:                uri,
		PayloadObjects:     []bw2.PayloadObject{namePo},
		PrimaryAccessChain: client.pac,
		ElaboratePAC:       bw2.ElaborateFull,
	})
}

func (client *SpawnClient) RestartService(spawnPoint *objects.SpawnPoint, svcName string) error {
	po := bw2.CreateStringPayloadObject(svcName)
	uri := spawnPoint.URI + "ctl/restart"

	return client.bwClient.Publish(&bw2.PublishParams{
		URI:                uri,
		PayloadObjects:     []bw2.PayloadObject{po},
		PrimaryAccessChain: client.pac,
		ElaboratePAC:       bw2.ElaborateFull,
	})
}

func (client *SpawnClient) StopService(spawnPoint *objects.SpawnPoint, svcName string) error {
	po := bw2.CreateStringPayloadObject(svcName)
	uri := spawnPoint.URI + "ctl/stop"

	return client.bwClient.Publish(&bw2.PublishParams{
		URI:                uri,
		PayloadObjects:     []bw2.PayloadObject{po},
		PrimaryAccessChain: client.pac,
		ElaboratePAC:       bw2.ElaborateFull,
	})
}