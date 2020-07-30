package clients

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/makkalot/eskit/generated/grpc/go/common"
	"github.com/makkalot/eskit/generated/grpc/go/crudstore"
	common3 "github.com/makkalot/eskit/services/lib/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"reflect"
)

func NewCrudStoreGRPCClient(ctx context.Context, storeEndpoint string) (crudstore.CrudStoreServiceClient, error) {
	conn, err := grpc.Dial(storeEndpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return crudstore.NewCrudStoreServiceClient(conn), nil

}

func NewCrudStoreGrpcClientWithWait(ctx context.Context, storeEndpoint string) (crudstore.CrudStoreServiceClient, error) {
	var conn *grpc.ClientConn
	var storeClient crudstore.CrudStoreServiceClient

	err := common3.RetryNormal(func() error {
		var err error
		conn, err = grpc.Dial(storeEndpoint, grpc.WithInsecure())
		if err != nil {
			return err
		}

		storeClient = crudstore.NewCrudStoreServiceClient(conn)
		_, err = storeClient.Healtz(ctx, &crudstore.HealthRequest{})
		if err != nil {
			return err
		}

		return nil
	})
	return storeClient, err
}

type CrudStoreClient struct {
	CrudGRPC crudstore.CrudStoreServiceClient
	ctx      context.Context
}

func NewCrudStoreClient(ctx context.Context, storeEndpoint string) (*CrudStoreClient, error) {

	client, err := NewCrudStoreGRPCClient(ctx, storeEndpoint)
	if err != nil {
		return nil, err
	}

	return NewCrudStoreWithActiveConn(ctx, client)
}

func NewCrudStoreWithActiveConn(ctx context.Context, client crudstore.CrudStoreServiceClient) (*CrudStoreClient, error) {
	return &CrudStoreClient{
		CrudGRPC: client,
		ctx:      ctx,
	}, nil
}

func (crud *CrudStoreClient) Create(msg proto.Message) (*common.Originator, error) {

	var originator *common.Originator

	o, ok := crud.extractOriginatorFromMsg(msg)
	if ok {
		originator = o
	}

	entityType := EntityTypeFromMsg(msg)

	marshaller := &jsonpb.Marshaler{}
	payloadJSON, err := marshaller.MarshalToString(msg)
	if err != nil {
		return nil, err
	}

	req := &crudstore.CreateRequest{
		EntityType: entityType,
		Originator: originator,
		Payload:    payloadJSON,
	}

	var createErr error
	var createResp *crudstore.CreateResponse

	if err := common3.RetryShort(func() error {
		createResp, err = crud.CrudGRPC.Create(crud.ctx, req)
		if err != nil {
			if status.Code(err) == codes.FailedPrecondition {
				return err
			}
			createErr = err
			return nil
		}
		return nil

	}); err != nil {
		return nil, err
	}

	if createErr != nil {
		return nil, createErr
	}

	return createResp.Originator, nil
}

func (crud *CrudStoreClient) Get(originator *common.Originator, msg proto.Message, deleted bool) error {
	if originator == nil {
		return fmt.Errorf("empty originator")
	}

	getResp, err := crud.CrudGRPC.Get(crud.ctx, &crudstore.GetRequest{
		EntityType: EntityTypeFromMsg(msg),
		Originator: originator,
		Deleted:    deleted,
	})

	if err != nil {
		return err
	}

	//log.Println("Unmarhal in crud.Get ", getResp.Payload)
	if err := jsonpb.UnmarshalString(getResp.Payload, msg); err != nil {
		return err
	}

	if err := crud.setOriginatorForMsg(msg, getResp.Originator); err != nil {
		return err
	}

	return nil
}

// TODO maybe can add some concurrency retry inside of it ?
func (crud *CrudStoreClient) Update(msg proto.Message) (*common.Originator, error) {
	var originator *common.Originator
	var ok bool

	originator, ok = crud.extractOriginatorFromMsg(msg)
	if !ok {
		return nil, fmt.Errorf("could not find the originator inside the message, can't continue")
	}

	entityType := EntityTypeFromMsg(msg)

	marshaller := &jsonpb.Marshaler{}
	payloadJSON, err := marshaller.MarshalToString(msg)
	if err != nil {
		return nil, err
	}

	updateResp, err := crud.CrudGRPC.Update(crud.ctx, &crudstore.UpdateRequest{
		EntityType: entityType,
		Originator: originator,
		Payload:    payloadJSON,
	})
	if err != nil {
		return nil, err
	}

	return updateResp.Originator, nil
}

func (crud *CrudStoreClient) Delete(originator *common.Originator, msg proto.Message) (*common.Originator, error) {
	if originator == nil {
		return nil, fmt.Errorf("empty originator")
	}

	deleteResp, err := crud.CrudGRPC.Delete(crud.ctx, &crudstore.DeleteRequest{
		EntityType: EntityTypeFromMsg(msg),
		Originator: originator,
	})

	if err != nil {
		return nil, err
	}

	return deleteResp.Originator, nil
}

// List is clever enough to fetch the results and populate the result interface with values
func (crud *CrudStoreClient) List(result interface{}) error {
	_, err := crud.ListWithPagination(result, "", 0)
	return err
}

func (crud *CrudStoreClient) ListWithPagination(result interface{}, fromPage string, size int) (string, error) {
	resultv := reflect.ValueOf(result)
	if resultv.Kind() != reflect.Ptr || resultv.Elem().Kind() != reflect.Slice {
		return "", fmt.Errorf("result argument must be a slice address")
	}

	slicev := resultv.Elem()
	slicev = slicev.Slice(0, slicev.Cap())
	elemType := slicev.Type()
	elemt := elemType.Elem()
	if elemt.Kind() != reflect.Ptr {
		return "", fmt.Errorf("the slice should contain addresses to objects ie. []*Object")
	}

	elemp := reflect.New(elemt.Elem())
	msgInterface := elemp.Interface()
	msg, ok := msgInterface.(proto.Message)
	if !ok {
		return "", fmt.Errorf("couldn't convert to proto message")
	}

	entityType := EntityTypeFromMsg(msg)
	resp, err := crud.CrudGRPC.List(crud.ctx, &crudstore.ListRequest{
		EntityType:   entityType,
		PaginationId: fromPage,
		Limit:        uint32(size),
	})
	if err != nil {
		return "", err
	}

	i := 0
	for _, res := range resp.Results {
		elemp := reflect.New(elemt.Elem())
		msgInterface := elemp.Interface()
		msg, ok := msgInterface.(proto.Message)
		if !ok {
			return "", fmt.Errorf("couldn't convert list result to proto message")
		}


		if err := jsonpb.UnmarshalString(res.Payload, msg); err != nil {
			log.Println("List: unmarshall : ", res.Payload, entityType)
			return "", err
		}

		if err := crud.setOriginatorForMsg(msg, res.Originator); err != nil {
			return "", err
		}

		msgValue := reflect.ValueOf(msg)

		if slicev.Len() == i {
			slicev = reflect.Append(slicev, msgValue)
			slicev = slicev.Slice(0, slicev.Cap())
		} else {
			slicev.Index(i).Set(msgValue)
		}

		i++
	}
	resultv.Elem().Set(slicev.Slice(0, i))
	return resp.NextPageId, nil
}

func (crud *CrudStoreClient) extractOriginatorFromMsg(msg proto.Message) (*common.Originator, bool) {
	s := reflect.ValueOf(msg).Elem()
	typeOfT := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)

		//log.Println("Checking the field : ", typeOfT.Field(i).Name)
		if typeOfT.Field(i).Name == "Originator" {
			i := f.Interface()
			originator, ok := i.(*common.Originator)
			//log.Println("Found the originator inside the message : ", originator)
			return originator, ok
		}
	}

	return nil, false
}

func (crud *CrudStoreClient) setOriginatorForMsg(msg proto.Message, originator *common.Originator) error {
	s := reflect.ValueOf(msg).Elem()
	typeOfT := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)

		if typeOfT.Field(i).Name == "Originator" {
			originatorValue := reflect.ValueOf(originator)
			f.Set(originatorValue)
			return nil
		}
	}

	return fmt.Errorf("originator field was not found in the message")
}

func EntityTypeFromMsg(msg proto.Message) string {
	return proto.MessageName(msg)
}

type crudStoreWithNoNetworking struct {
	server crudstore.CrudStoreServiceServer
}

func NewCrudStoreClientWithNoNetworking(server crudstore.CrudStoreServiceServer) crudstore.CrudStoreServiceClient {
	return &crudStoreWithNoNetworking{
		server: server,
	}
}

func (c *crudStoreWithNoNetworking) Healtz(ctx context.Context, in *crudstore.HealthRequest, opts ...grpc.CallOption) (*crudstore.HealthResponse, error) {
	return c.server.Healtz(ctx, in)
}

func (c *crudStoreWithNoNetworking) Create(ctx context.Context, in *crudstore.CreateRequest, opts ...grpc.CallOption) (*crudstore.CreateResponse, error) {
	return c.server.Create(ctx, in)
}

func (c *crudStoreWithNoNetworking) Update(ctx context.Context, in *crudstore.UpdateRequest, opts ...grpc.CallOption) (*crudstore.UpdateResponse, error) {
	return c.server.Update(ctx, in)
}

func (c *crudStoreWithNoNetworking) Delete(ctx context.Context, in *crudstore.DeleteRequest, opts ...grpc.CallOption) (*crudstore.DeleteResponse, error) {
	return c.server.Delete(ctx, in)
}

func (c *crudStoreWithNoNetworking) Get(ctx context.Context, in *crudstore.GetRequest, opts ...grpc.CallOption) (*crudstore.GetResponse, error) {
	return c.server.Get(ctx, in)
}

func (c *crudStoreWithNoNetworking) List(ctx context.Context, in *crudstore.ListRequest, opts ...grpc.CallOption) (*crudstore.ListResponse, error) {
	return c.server.List(ctx, in)
}

func (c *crudStoreWithNoNetworking) RegisterType(ctx context.Context, in *crudstore.RegisterTypeRequest, opts ...grpc.CallOption) (*crudstore.RegisterTypeResponse, error) {
	return c.server.RegisterType(ctx, in)
}

func (c *crudStoreWithNoNetworking) GetType(ctx context.Context, in *crudstore.GetTypeRequest, opts ...grpc.CallOption) (*crudstore.GetTypeResponse, error) {
	return c.server.GetType(ctx, in)
}

func (c *crudStoreWithNoNetworking) UpdateType(ctx context.Context, in *crudstore.UpdateTypeRequest, opts ...grpc.CallOption) (*crudstore.UpdateTypeResponse, error) {
	return c.server.UpdateType(ctx, in)
}

func (c *crudStoreWithNoNetworking) ListTypes(ctx context.Context, in *crudstore.ListTypesRequest, opts ...grpc.CallOption) (*crudstore.ListTypesResponse, error) {
	return c.server.ListTypes(ctx, in)
}
