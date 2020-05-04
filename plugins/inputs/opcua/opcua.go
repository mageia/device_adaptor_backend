package opcua

import (
	"context"
	"device_adaptor"
	"device_adaptor/internal"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/inputs"
	"device_adaptor/utils"
	"fmt"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/rs/zerolog/log"
	"strings"
	"time"
)

type OpcUA struct {
	Endpoint     string            `json:"endpoint"`
	ObjectId     string            `json:"object_id"`
	Interval     internal.Duration `json:"interval"`
	Timeout      internal.Duration `json:"timeout"`
	FieldPrefix  string            `json:"field_prefix"`
	FieldSuffix  string            `json:"field_suffix"`
	NameOverride string            `json:"name_override"`

	client             *opcua.Client
	connected          bool
	originName         string
	quality            device_adaptor.Quality
	itemIdList         []string
	pointMap           map[string]points.PointDefine
	_pointAddressToKey map[string]string
	listening          bool
}

func (o *OpcUA) OriginName() string {
	return o.originName
}

func (o *OpcUA) SetValue(map[string]interface{}) error {
	return nil
}

func (o *OpcUA) UpdatePointMap(map[string]interface{}) error {
	return nil
}

func (o *OpcUA) RetrievePointMap([]string) map[string]points.PointDefine {
	return o.pointMap
}

func (o *OpcUA) Name() string {
	if o.NameOverride != "" {
		return o.NameOverride
	}
	return o.originName
}

func (o *OpcUA) ReadItems(nodeIds ...string) (map[string]interface{}, error) {
	defer func() {
		if e := recover(); e != nil {
			log.Error().Err(fmt.Errorf("%v", e)).Msg("ReadItems")
		}
	}()

	result := make(map[string]interface{})
	if len(nodeIds) == 0 {
		return result, nil
	}

	baseOffset := 50
	sliceLen := len(nodeIds)
	for baseAddr := 0; baseAddr < sliceLen/baseOffset+1; baseAddr++ {
		nodeToRead := make([]*ua.ReadValueID, 0)
		for _, v := range nodeIds[baseAddr*baseOffset : baseAddr*baseOffset+
			utils.MinInt(baseOffset, sliceLen-baseAddr*baseOffset)] {
			id, e := ua.ParseNodeID(v)
			if e != nil {
				log.Error().Err(e).Msg("ParseNodeID")
				continue
			}
			nodeToRead = append(nodeToRead, &ua.ReadValueID{NodeID: id})
			req := &ua.ReadRequest{
				MaxAge:             0,
				NodesToRead:        nodeToRead,
				TimestampsToReturn: ua.TimestampsToReturnServer,
			}
			resp, err := o.client.Read(req)
			if err != nil {
				//log.Error().Err(err).Int("len(nodeToRead)", len(nodeToRead)).Str("objectId", nodeIds[0]).Msg("Read")
				return nil, err
			}

			for i, v := range resp.Results {
				if v.Value != nil {
					result[nodeIds[baseOffset*baseAddr+i]] = v.Value.Value()
				}
			}
		}
	}

	return result, nil
}

func (o *OpcUA) CheckGather(acc device_adaptor.Accumulator) error {
	if !o.connected {
		o.Start()
	}

	fields := make(map[string]interface{})
	tags := make(map[string]string)

	defer func(ua *OpcUA) {
		if e := recover(); e != nil {
			acc.AddError(fmt.Errorf("%v", e))
		}
		acc.AddFields(ua.Name(), fields, tags, ua.SelfCheck())
	}(o)

	rtDataMap, e := o.ReadItems(o.itemIdList...)
	if e != nil {
		//log.Error().Str("objectId", o.ObjectId).Err(e).Msg("Polling Error")
		o.Stop()
		return e
	}
	fields = rtDataMap
	return nil
}

func (o *OpcUA) StartListen(ctx context.Context, acc device_adaptor.Accumulator) (bool, error) {
	if o.listening {
		return true, nil
	}

	notifyCh := make(chan *opcua.PublishNotificationData)
	sub, err := o.client.Subscribe(nil, notifyCh)
	if err != nil {
		//log.Error().Err(err).Msg("Init Subscribe instance failed")
		return false, err
	}

	baseOffset := 50
	sliceLen := len(o.itemIdList)
	for baseAddr := 0; baseAddr < sliceLen/baseOffset+1; baseAddr++ {
		nodeSubs := make([]*ua.MonitoredItemCreateRequest, 0)
		for i, v := range o.itemIdList[baseAddr*baseOffset : baseAddr*baseOffset+
			utils.MinInt(baseOffset, sliceLen-baseAddr*baseOffset)] {
			id, e := ua.ParseNodeID(v)
			if e != nil {
				log.Error().Str("nodeId", v).Err(e).Msg("ParseNodeID")
				o.Stop()
				return false, e
			}
			nodeSubs = append(nodeSubs, opcua.NewMonitoredItemCreateRequestWithDefaults(id, ua.AttributeIDValue, uint32(i)))
		}
		res, err := sub.Monitor(ua.TimestampsToReturnBoth, nodeSubs...)
		if err != nil || res.Results[0].StatusCode != ua.StatusOK {
			log.Error().Err(err).Interface("result", res).Msg("Monitor")
			o.Stop()
			return false, err
		}
	}

	go sub.Run(context.Background())

	o.listening = true
	log.Info().Str("plugin", o.Name()).Msg("StartListen")

	defer func() {
		if err := recover(); err != nil {
			log.Error().Err(err.(error)).Str("objectId", o.ObjectId).Msg("PANIC")
		}
	}()

	for {
		select {
		case res := <-notifyCh:
			if res.Error != nil {
				log.Error().Str("addr", o.Endpoint).Str("objectId", o.ObjectId).Err(res.Error).Msg("Sub Error")
				sub.Cancel()
				o.listening = false
				o.Stop()
				return false, res.Error
			}
			switch x := res.Value.(type) {
			case *ua.DataChangeNotification:
				miDataMap := make(map[string]interface{})

				for _, item := range x.MonitoredItems {
					if item.Value.Status == ua.StatusOK {
						//TODO: index out of range
						miDataMap[o.itemIdList[item.ClientHandle]] = item.Value.Value.Value()
					}
				}
				acc.AddFields(o.Name(), miDataMap, map[string]string{}, o.SelfCheck())
			default:
				log.Error().Msg(fmt.Sprintf("what's this publish result? %T", res.Value))
			}
		}
	}
}
func (o *OpcUA) GetListening() bool {
	return o.listening
}
func (o *OpcUA) SelfCheck() device_adaptor.Quality {
	return o.quality
}

func (o *OpcUA) SetPointMap(pointMap map[string]points.PointDefine) {
	o.pointMap = pointMap
}
func (o *OpcUA) RecursiveForItems(objectId string) []string {
	itemList := make([]string, 0)
	if objectId == "" {
		return itemList
	}

	id, err := ua.ParseNodeID(objectId)
	if err != nil {
		log.Error().Err(err).Str("objectId", objectId).Msg("ua.ParseNodeID")
		return itemList
	}
	desc := &ua.BrowseDescription{
		NodeID:          id,
		ReferenceTypeID: ua.NewNumericNodeID(0, 0),
		BrowseDirection: ua.BrowseDirectionForward,
		IncludeSubtypes: true,
		NodeClassMask:   uint32(ua.NodeClassAll),
		ResultMask:      uint32(ua.BrowseResultMaskAll),
	}
	repBrowse := &ua.BrowseRequest{
		View: &ua.ViewDescription{
			ViewID:    ua.NewTwoByteNodeID(0),
			Timestamp: time.Now(),
		},
		RequestedMaxReferencesPerNode: 0,
		NodesToBrowse:                 []*ua.BrowseDescription{desc},
	}
	resp, err := o.client.Browse(repBrowse)
	if err != nil {
		log.Error().Err(err).Msg("browse request")
		return itemList
	}

	for _, v := range resp.Results[0].References {
		if v.BrowseName.Name != "FolderType" && !strings.HasPrefix(v.BrowseName.Name, "_") {
			switch v.NodeClass {
			case ua.NodeClassObject:
				itemList = append(itemList, o.RecursiveForItems(v.NodeID.NodeID.String())...)
			case ua.NodeClassVariable:
				itemList = append(itemList, v.NodeID.NodeID.String())
			default:
				log.Debug().Interface("NodeClass", v.NodeClass).Msg("RecursiveForItems")
			}
		}
	}
	return itemList
}
func (o *OpcUA) ProbePointMap() map[string]points.PointDefine {
	pointMap := make(map[string]points.PointDefine)
	o.itemIdList = o.RecursiveForItems(o.ObjectId)
	for _, k := range o.itemIdList {
		pointMap[k] = points.PointDefine{Address: k}
	}
	o.pointMap = pointMap
	return pointMap
}

func (o *OpcUA) Start() error {
	if o.connected {
		return nil
	}

	c := opcua.NewClient(o.Endpoint)
	if e := c.Connect(nil); e != nil {
		log.Error().Err(e).Str("endpoint", o.Endpoint).Msg("connect failed")
		return e
	}
	o.client = c
	o.connected = true

	o.pointMap = o.ProbePointMap()
	//log.Debug().Interface("pointMap", o.pointMap).Msg("PointMap")
	return nil
}

func (o *OpcUA) Stop() {
	if o.connected {
		o.client.Close()
		o.connected = false
	}
}

func init() {
	inputs.Add("opc_ua", func() device_adaptor.Input {
		return &OpcUA{
			originName: "opc_ua",
			quality:    device_adaptor.QualityGood,
			Timeout:    internal.Duration{Duration: time.Second * 5},
		}
	})
}
