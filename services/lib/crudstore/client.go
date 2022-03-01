package crudstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/makkalot/eskit/generated/grpc/go/common"
	uuid "github.com/satori/go.uuid"
	"log"
	"reflect"
)

var (
	InvalidArgumentError = errors.New("invalid argument")
)

type StructCrudStoreClient struct {
	crudStore CrudStore
}

func NewStructCrudStoreClient(crudStore CrudStore) *StructCrudStoreClient {
	return &StructCrudStoreClient{crudStore: crudStore}
}

func (client *StructCrudStoreClient) checkIfPtr(msg interface{}) error {
	t := reflect.TypeOf(msg)
	if t.Kind() == reflect.Ptr {
		return nil
	}
	return fmt.Errorf("non pointer : %w", InvalidArgumentError)
}

// Create creates a new entry into crudstore for the given struct, it uses its structname for
// entity type for now
func (client *StructCrudStoreClient) Create(msg interface{}) (*common.Originator, error) {
	var originator *common.Originator

	if err := client.checkIfPtr(msg); err != nil {
		return nil, err
	}

	o, ok := client.extractOriginatorFromMsg(msg)
	if ok {
		originator = o
	}

	if originator == nil {
		originator = &common.Originator{
			Id:      uuid.Must(uuid.NewV4()).String(),
			Version: "1",
		}
	}

	entityType := EntityTypeFromStruct(msg)
	payloadJSON, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	err = client.crudStore.Create(entityType, originator, string(payloadJSON))
	if err != nil {
		return nil, err
	}

	if err := client.setOriginatorForMsg(msg, originator); err != nil {
		return nil, err
	}

	return originator, nil
}

func (client *StructCrudStoreClient) Get(originator *common.Originator, msg interface{}, deleted bool) error {
	if originator == nil {
		return fmt.Errorf("empty originator : %w", InvalidArgumentError)
	}

	if err := client.checkIfPtr(msg); err != nil {
		return err
	}

	payload, originator, err := client.crudStore.Get(
		originator,
		deleted)

	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(payload), msg); err != nil {
		return fmt.Errorf("restoring the payload : %w", err)
	}

	if err := client.setOriginatorForMsg(msg, originator); err != nil {
		return err
	}

	return nil
}

// Update updates the object, it should have the originator set
func (client *StructCrudStoreClient) Update(msg interface{}) (*common.Originator, error) {
	var originator *common.Originator
	var ok bool

	originator, ok = client.extractOriginatorFromMsg(msg)
	if !ok {
		return nil, fmt.Errorf("could not find the originator inside the message, can't continue")
	}

	if originator == nil {
		return nil, fmt.Errorf("empty originator found inside the message")
	}

	entityType := EntityTypeFromStruct(msg)

	payloadJSON, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	updatedOriginator, err := client.crudStore.Update(
		entityType,
		originator,
		string(payloadJSON),
	)
	if err != nil {
		return nil, err
	}

	if err := client.setOriginatorForMsg(msg, updatedOriginator); err != nil {
		return nil, err
	}

	return updatedOriginator, nil
}

func (client *StructCrudStoreClient) Delete(originator *common.Originator, msg interface{}) (*common.Originator, error) {
	if originator == nil {
		return nil, fmt.Errorf("empty originator")
	}

	deletedOriginator, err := client.crudStore.Delete(
		EntityTypeFromStruct(msg),
		originator,
	)

	if err != nil {
		return nil, err
	}

	return deletedOriginator, nil
}

func (client *StructCrudStoreClient) ListWithPagination(result interface{}, fromID string, size int) (string, error) {
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
	msg := elemp.Interface()

	entityType := EntityTypeFromStruct(msg)
	results, lastID, err := client.crudStore.List(
		entityType,
		fromID,
		size,
	)

	if err != nil {
		return "", err
	}

	i := 0
	var latestOriginator *common.Originator
	for _, resOriginator := range results {
		elemp := reflect.New(elemt.Elem())
		msg := elemp.Interface()

		p, originator, err := client.crudStore.Get(resOriginator, false)
		if err != nil {
			log.Printf("Skipping originator : %+v because of : %v \n", originator, err)
			continue
		}
		latestOriginator = originator

		if err := json.Unmarshal([]byte(p), msg); err != nil {
			return "", fmt.Errorf("list : payload : %s, entityType : %s: %w", p, entityType, err)
		}

		if err := client.setOriginatorForMsg(msg, originator); err != nil {
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
	if latestOriginator == nil {
		return "", nil
	}
	return lastID, nil
}

func (client *StructCrudStoreClient) setOriginatorForMsg(msg interface{}, originator *common.Originator) error {
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

func (client *StructCrudStoreClient) extractOriginatorFromMsg(msg interface{}) (*common.Originator, bool) {
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

func EntityTypeFromStruct(msg interface{}) string {
	t := reflect.TypeOf(msg)
	return t.Elem().Name()
}
