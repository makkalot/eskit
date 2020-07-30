package provider

import (
	"context"
	"fmt"
	"github.com/makkalot/eskit/generated/grpc/go/common"
	"github.com/makkalot/eskit/generated/grpc/go/crudstore"
	eskitcommon "github.com/makkalot/eskit/services/lib/common"
	crud "github.com/makkalot/eskit/services/lib/crudstore"
	"github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/protobuf/jsonpb"
	"github.com/xeipuuv/gojsonschema"
	"log"
)

type CrudStoreSvcProvider struct {
	storage crud.CrudStore
}

func NewCrudStoreApiProvider(storage crud.CrudStore) (crudstore.CrudStoreServiceServer, error) {
	return &CrudStoreSvcProvider{
		storage: storage,
	}, nil
}

func (svc *CrudStoreSvcProvider) Healtz(ctx context.Context, req *crudstore.HealthRequest) (*crudstore.HealthResponse, error) {
	return &crudstore.HealthResponse{}, nil
}

func (svc *CrudStoreSvcProvider) validateOriginator(originator *common.Originator) error {
	if originator == nil {
		return status.Error(codes.InvalidArgument, "missing originator")
	}

	if originator.Id == "" {
		return status.Error(codes.InvalidArgument, "missing originator id")
	}

	if originator.Version == "" {
		return status.Error(codes.InvalidArgument, "missing originator version")
	}

	return nil
}

func (svc *CrudStoreSvcProvider) Create(ctx context.Context, req *crudstore.CreateRequest) (*crudstore.CreateResponse, error) {
	if req.EntityType == "" {
		return nil, status.Error(codes.InvalidArgument, "missing entity type")
	}

	if req.Payload == "" {
		return nil, status.Error(codes.InvalidArgument, "missing payload")
	}

	originator := req.Originator
	if originator == nil {
		originator = &common.Originator{
			Id:      uuid.Must(uuid.NewV4()).String(),
			Version: "1",
		}
	} else if originator.Version == "" {
		originator.Version = "1"
	}

	if err := svc.validateOriginator(originator); err != nil {
		return nil, err
	}

	if err := svc.validateDoc(req.EntityType, req.Payload); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation : %s", err)
	}

	err := svc.storage.Create(req.EntityType, originator, req.Payload)
	if err != nil {
		if crud.IsDuplicate(err) {
			return nil, status.Errorf(codes.AlreadyExists, "duplicate")
		}
		return nil, status.Errorf(codes.Internal, "creation failed : %v", err)
	}

	return &crudstore.CreateResponse{Originator: originator}, nil
}

func (svc *CrudStoreSvcProvider) Update(ctx context.Context, req *crudstore.UpdateRequest) (*crudstore.UpdateResponse, error) {
	if req.EntityType == "" {
		return nil, status.Error(codes.InvalidArgument, "missing entity type")
	}

	if req.Payload == "" {
		return nil, status.Error(codes.InvalidArgument, "missing payload")
	}

	originator := req.Originator
	if err := svc.validateOriginator(originator); err != nil {
		return nil, err
	}

	if err := svc.validateDoc(req.EntityType, req.Payload); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation : %s", err)
	}

	updatedOriginator, err := svc.storage.Update(req.EntityType, originator, req.Payload)
	if err != nil {
		if crud.IsDuplicate(err) {
			return nil, status.Errorf(codes.AlreadyExists, "update failed old version : %v", err)
		}
		return nil, status.Errorf(codes.Internal, "update failed : %v", err)
	}

	return &crudstore.UpdateResponse{
		Originator: updatedOriginator,
	}, nil
}

// validates the document against a schema if registered at all
func (svc *CrudStoreSvcProvider) validateDoc(entityType, payload string) error {
	spec, _, err := svc.getSpecForEntity(&common.Originator{Id: entityType})
	if err != nil {
		return err
	}

	if spec == nil {
		return nil
	}

	if spec.SchemaSpec == nil || spec.SchemaSpec.JsonSchema == "" {
		return nil
	}

	schemaLoader := gojsonschema.NewStringLoader(spec.SchemaSpec.JsonSchema)
	documentLoader := gojsonschema.NewStringLoader(payload)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if result.Valid() {
		return nil
	}

	var errStr string
	for _, err := range result.Errors() {
		errStr += fmt.Sprintf("- %s\n", err)
	}

	return fmt.Errorf("schema constraint failed : %s", errStr)
}

func (svc *CrudStoreSvcProvider) Get(ctx context.Context, req *crudstore.GetRequest) (*crudstore.GetResponse, error) {
	if req.Originator == nil || req.Originator.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "missing originator id")
	}

	if req.EntityType == "" {
		return nil, status.Error(codes.InvalidArgument, "missing entity type")
	}

	payload, originator, err := svc.storage.Get(req.Originator, req.Deleted)
	if err != nil {
		if crud.IsErrDeleted(err) || crud.IsErrNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "getting entity failed : %v", err)
	}

	return &crudstore.GetResponse{Originator: originator, Payload: payload}, nil
}

func (svc *CrudStoreSvcProvider) Delete(ctx context.Context, req *crudstore.DeleteRequest) (*crudstore.DeleteResponse, error) {
	if req.Originator == nil || req.Originator.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "missing originator id")
	}

	if req.EntityType == "" {
		return nil, status.Error(codes.InvalidArgument, "missing entity type")
	}

	deletedOriginator, err := svc.storage.Delete(req.EntityType, req.Originator)
	if err != nil {
		return nil, err
	}

	return &crudstore.DeleteResponse{
		Originator: deletedOriginator,
	}, nil
}

func (svc *CrudStoreSvcProvider) List(ctx context.Context, req *crudstore.ListRequest) (*crudstore.ListResponse, error) {
	if req.EntityType == "" {
		return nil, status.Error(codes.InvalidArgument, "missing entity type")
	}

	originators, nextPage, err := svc.storage.List(req.EntityType, req.PaginationId, int(req.Limit))
	if err != nil {
		return nil, err
	}

	var results []*crudstore.ListResponseItem
	skipPayload := req.SkipPayload
	var latestOriginator *common.Originator
	for _, o := range originators {
		var payload string


		if !skipPayload {
			p, originator, err := svc.storage.Get(o, false)
			if err != nil {
				log.Printf("Skipping originator : %+v because of : %v \n", o, err)
				continue
			}
			payload = p
			latestOriginator = originator
		}

		results = append(results, &crudstore.ListResponseItem{
			EntityType: req.EntityType,
			Originator: latestOriginator,
			Payload:    payload,
		})
	}

	if results == nil {
		results = []*crudstore.ListResponseItem{}
	}

	return &crudstore.ListResponse{
		Results:    results,
		NextPageId: nextPage,
	}, nil
}

func (svc *CrudStoreSvcProvider) RegisterType(ctx context.Context, req *crudstore.RegisterTypeRequest) (*crudstore.RegisterTypeResponse, error) {
	spec := req.Spec
	if spec == nil {
		return nil, status.Error(codes.InvalidArgument, "missing spec")
	}

	if spec.EntityType == "" {
		return nil, status.Error(codes.InvalidArgument, "missing entity type")
	}

	existingSpec, _, err := svc.getSpecForEntity(&common.Originator{
		Id: spec.EntityType,
	})
	if err != nil && !crud.IsErrNotFound(err) {
		return nil, status.Errorf(codes.Internal, "fetching spec failed : %v", err)
	}

	if existingSpec != nil {
		if !req.SkipDuplicate {
			return nil, status.Error(codes.AlreadyExists, "already registered")
		}
		return &crudstore.RegisterTypeResponse{}, nil
	}

	if spec.GetSchemaSpec() != nil && spec.SchemaSpec.JsonSchema != "" {
		if _, err := svc.validateJSONSchema(spec.SchemaSpec); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "schema : %v", err)
		}
	}

	marshaller := &jsonpb.Marshaler{}
	specJSON, err := marshaller.MarshalToString(spec)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "json marshall")
	}

	originator := &common.Originator{
		Id:      spec.EntityType,
		Version: "1",
	}
	log.Println("Registering type with originator ", originator)

	err = svc.storage.Create(eskitcommon.RegisterTypeEntity, originator, specJSON)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "saving failed : %v", err)
	}

	return &crudstore.RegisterTypeResponse{}, nil
}

// checks if the specified schema is valid JSON Schema
func (svc *CrudStoreSvcProvider) validateJSONSchema(spec *crudstore.SchemaSpec) (gojsonschema.JSONLoader, error) {
	if spec.SchemaVersion == 0 {
		return nil, fmt.Errorf("version is required")
	}

	loader := gojsonschema.NewStringLoader(spec.JsonSchema)
	_, err := loader.LoadJSON()
	if err != nil {
		return nil, fmt.Errorf("invalid spec : %v", err)
	}

	sl := gojsonschema.NewSchemaLoader()
	_, err = sl.Compile(loader)
	if err != nil {
		return nil, fmt.Errorf("compile schema : %v", err)
	}

	return loader, nil
}

func (svc *CrudStoreSvcProvider) getSpecForEntity(originator *common.Originator) (*crudstore.CrudEntitySpec, *common.Originator, error) {

	payload, originator, err := svc.storage.Get(originator, false)
	if err != nil {
		if crud.IsErrNotFound(err) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	crudSpec := &crudstore.CrudEntitySpec{}
	if err := jsonpb.UnmarshalString(payload, crudSpec); err != nil {
		return nil, nil, err
	}

	return crudSpec, originator, nil
}

func (svc *CrudStoreSvcProvider) GetType(ctx context.Context, req *crudstore.GetTypeRequest) (*crudstore.GetTypeResponse, error) {
	if req.EntityType == "" {
		return nil, status.Error(codes.InvalidArgument, "missing entity type")
	}

	existingSpec, _, err := svc.getSpecForEntity(&common.Originator{
		Id: req.EntityType,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "missing entity type")
	}

	if existingSpec == nil {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &crudstore.GetTypeResponse{Spec: existingSpec}, nil
}

func (svc *CrudStoreSvcProvider) UpdateType(ctx context.Context, req *crudstore.UpdateTypeRequest) (*crudstore.UpdateTypeResponse, error) {
	spec := req.Spec
	if spec == nil {
		return nil, status.Error(codes.InvalidArgument, "missing spec")
	}

	if spec.EntityType == "" {
		return nil, status.Error(codes.InvalidArgument, "missing entity type")
	}

	oldSpec, originator, err := svc.getSpecForEntity(&common.Originator{
		Id: spec.EntityType,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "missing entity type")
	}

	if oldSpec == nil {
		return nil, status.Errorf(codes.NotFound, "schema not found: %v", err)
	}

	if spec.GetSchemaSpec() != nil && spec.SchemaSpec.JsonSchema != "" {
		if _, err := svc.validateJSONSchema(spec.SchemaSpec); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "schema : %v", err)
		}

		if oldSpec.SchemaSpec != nil && oldSpec.SchemaSpec.SchemaVersion == spec.SchemaSpec.SchemaVersion {
			return nil, status.Error(codes.InvalidArgument, "schema version not changed")
		}
	}

	m := &jsonpb.Marshaler{}
	specJSON, err := m.MarshalToString(spec)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "marshal : %v", err)
	}

	_, err = svc.storage.Update(eskitcommon.RegisterTypeEntity, originator, specJSON)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "saving type : %v", err)
	}

	return &crudstore.UpdateTypeResponse{}, nil
}

func (svc *CrudStoreSvcProvider) ListTypes(ctx context.Context, req *crudstore.ListTypesRequest) (*crudstore.ListTypesResponse, error) {

	size := 20
	if req.Limit != 0 {
		size = int(req.Limit)
	}
	originators, _, err := svc.storage.List(eskitcommon.RegisterTypeEntity, "0", size)
	if err != nil {
		return nil, err
	}

	var results []*crudstore.CrudEntitySpec
	for _, o := range originators {
		var payload string

		p, _, err := svc.storage.Get(o, false)
		if err != nil {
			log.Printf("Skipping originator : %+v because of : %v \n", o, err)
			continue
		}
		payload = p

		crudSpec := &crudstore.CrudEntitySpec{}
		if err := jsonpb.UnmarshalString(payload, crudSpec); err != nil {
			return nil, err
		}

		results = append(results, crudSpec)
	}

	return &crudstore.ListTypesResponse{
		Results: results,
	}, nil
}
