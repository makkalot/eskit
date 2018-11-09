package benchmark

import (
	"testing"
	"fmt"
	"os"
	"github.com/makkalot/eskit/services/clients"
	"context"
	"github.com/makkalot/eskit/generated/grpc/go/common"
	"github.com/makkalot/eskit/generated/grpc/go/users"
	"github.com/satori/go.uuid"
	"log"
	"github.com/davecgh/go-spew/spew"
)

func BenchmarkCrudCreate(b *testing.B) {
	crudStoreEndpoint := os.Getenv("CRUDSTORE_ENDPOINT")
	if crudStoreEndpoint == "" {
		b.Fatalf("CRUDSTORE_ENDPOINT is required")
	}

	ctx, cancel := context.WithCancel(context.Background())

	crudStoreClient, err := clients.NewCrudStoreGrpcClientWithWait(ctx, crudStoreEndpoint)
	if err != nil {
		b.Fatalf("crud store client initialization failed : %v", err)
	}

	crud, err := clients.NewCrudStoreWithActiveConn(ctx, crudStoreClient)
	if err != nil {
		b.Fatalf("crud store client initialization failed : %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		originator := &common.Originator{
			Id:      uuid.Must(uuid.NewV4()).String(),
			Version: "1",
		}

		user := &users.User{
			Originator: originator,
			Email:      fmt.Sprintf("testeskit_%d@gmail.com", i),
			FirstName:  "test",
			LastName:   "eskit",
		}

		_, err := crud.Create(user)
		if err != nil {
			log.Println("creation failed : ", err)
		}
	}

	cancel()
}

func BenchmarkGetNoUpdate(b *testing.B) {
	crudStoreEndpoint := os.Getenv("CRUDSTORE_ENDPOINT")
	if crudStoreEndpoint == "" {
		b.Fatalf("CRUDSTORE_ENDPOINT is required")
	}

	ctx, cancel := context.WithCancel(context.Background())

	crudStoreClient, err := clients.NewCrudStoreGrpcClientWithWait(ctx, crudStoreEndpoint)
	if err != nil {
		b.Fatalf("crud store client initialization failed : %v", err)
	}

	crud, err := clients.NewCrudStoreWithActiveConn(ctx, crudStoreClient)
	if err != nil {
		b.Fatalf("crud store client initialization failed : %v", err)
	}

	var originators []*common.Originator
	for i := 0; i < b.N; i++ {
		originator := &common.Originator{
			Id:      uuid.Must(uuid.NewV4()).String(),
			Version: "1",
		}

		user := &users.User{
			Originator: originator,
			Email:      fmt.Sprintf("testeskitget_%d@gmail.com", i),
			FirstName:  "test",
			LastName:   "eskit",
		}

		createOriginator, err := crud.Create(user)
		if err != nil {
			log.Println("creation failed : ", err)
			continue
		}

		originators = append(originators, createOriginator)
	}

	b.ResetTimer()

	for _, o := range originators {
		u := &users.User{}
		if err := crud.Get(o, u, false); err != nil {
			log.Println("fetching failed : ", err)
		}
	}

	cancel()
}

func BenchmarkGetWithUpdate(b *testing.B) {
	crudStoreEndpoint := os.Getenv("CRUDSTORE_ENDPOINT")
	if crudStoreEndpoint == "" {
		b.Fatalf("CRUDSTORE_ENDPOINT is required")
	}

	ctx, cancel := context.WithCancel(context.Background())

	crudStoreClient, err := clients.NewCrudStoreGrpcClientWithWait(ctx, crudStoreEndpoint)
	if err != nil {
		b.Fatalf("crud store client initialization failed : %v", err)
	}

	crud, err := clients.NewCrudStoreWithActiveConn(ctx, crudStoreClient)
	if err != nil {
		b.Fatalf("crud store client initialization failed : %v", err)
	}

	var originators []*common.Originator
	for i := 0; i < b.N; i++ {
		originator := &common.Originator{
			Id:      uuid.Must(uuid.NewV4()).String(),
			Version: "1",
		}

		user := &users.User{
			Originator: originator,
			Email:      fmt.Sprintf("testeskitget_%d@gmail.com", i),
			FirstName:  "test",
			LastName:   "eskit",
		}

		createOriginator, err := crud.Create(user)
		if err != nil {
			log.Println("creation failed : ", err)
			continue
		}

		originators = append(originators, createOriginator)
	}

	for _, o := range originators {
		for i := 0; i < 50; i++ {
			u := &users.User{}
			fetchOriginator := &common.Originator{Id: o.Id}
			if err := crud.Get(fetchOriginator, u, false); err != nil {
				log.Println("fetching failed : ", err)
			}
			log.Println("Fetched user : ", spew.Sdump(u))
			u.LastName = fmt.Sprintf("eskit_%d", i)
			log.Println("Saving : ", spew.Sdump(u))
			if _, err := crud.Update(u); err != nil {
				log.Println("update failed : ", err)
				continue
			}
		}
	}

	b.ResetTimer()
	for _, o := range originators {
		u := &users.User{}
		if err := crud.Get(o, u, false); err != nil {
			log.Println("updated fetching failed : ", err)
		}
	}

	cancel()
}
