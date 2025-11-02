package users_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/makkalot/eskit/lib/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type CreateUserRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type UserResponse struct {
	Originator *types.Originator `json:"originator"`
	Email      string            `json:"email"`
	FirstName  string            `json:"firstName"`
	LastName   string            `json:"lastName"`
	Active     bool              `json:"active"`
	Workspaces []string          `json:"workspaces"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

var _ = Describe("Users", func() {

	var httpClient *http.Client
	var baseURL string

	BeforeEach(func() {
		httpClient = &http.Client{}
		baseURL = fmt.Sprintf("http://%s/v1/users", userEndpoint)
	})

	Context("When User is Created", func() {
		It("Should be returned via its ID", func() {

			req := &CreateUserRequest{
				Email:     "ahmet@ahmet.com",
				FirstName: "SameName",
				LastName:  "Eskit",
			}

			// Create user
			jsonData, err := json.Marshal(req)
			Expect(err).To(BeNil())

			resp, err := httpClient.Post(baseURL, "application/json", bytes.NewBuffer(jsonData))
			Expect(err).To(BeNil())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			var createResp UserResponse
			err = json.NewDecoder(resp.Body).Decode(&createResp)
			Expect(err).To(BeNil())
			Expect(createResp.Originator).NotTo(BeNil())
			Expect(createResp.Originator.ID).NotTo(BeEmpty())

			// Get user by ID
			getURL := fmt.Sprintf("%s?id=%s", baseURL, createResp.Originator.ID)
			getResp, err := httpClient.Get(getURL)
			Expect(err).To(BeNil())
			defer getResp.Body.Close()
			Expect(getResp.StatusCode).To(Equal(http.StatusOK))

			var getUser UserResponse
			err = json.NewDecoder(getResp.Body).Decode(&getUser)
			Expect(err).To(BeNil())

			Expect(getUser.Email).To(Equal(req.Email))
			Expect(getUser.FirstName).To(Equal(req.FirstName))
			Expect(getUser.LastName).To(Equal(req.LastName))

			// Create second user
			req2 := &CreateUserRequest{
				Email:     "osman@osman.com",
				FirstName: "SameName",
				LastName:  "Eskit",
			}

			jsonData2, err := json.Marshal(req2)
			Expect(err).To(BeNil())

			resp2, err := httpClient.Post(baseURL, "application/json", bytes.NewBuffer(jsonData2))
			Expect(err).To(BeNil())
			defer resp2.Body.Close()
			Expect(resp2.StatusCode).To(Equal(http.StatusCreated))

			var createResp2 UserResponse
			err = json.NewDecoder(resp2.Body).Decode(&createResp2)
			Expect(err).To(BeNil())
			Expect(createResp2.Originator).NotTo(BeNil())

			// Get second user by ID
			getURL2 := fmt.Sprintf("%s?id=%s", baseURL, createResp2.Originator.ID)
			getResp2, err := httpClient.Get(getURL2)
			Expect(err).To(BeNil())
			defer getResp2.Body.Close()
			Expect(getResp2.StatusCode).To(Equal(http.StatusOK))

			var getUser2 UserResponse
			err = json.NewDecoder(getResp2.Body).Decode(&getUser2)
			Expect(err).To(BeNil())

			Expect(getUser2.Email).To(Equal(req2.Email))
			Expect(getUser2.FirstName).To(Equal(req2.FirstName))
			Expect(getUser2.LastName).To(Equal(req2.LastName))

		})
	})

	Context("When User is Deleted", func() {
		var deletedUserID string
		var runOnce bool

		BeforeEach(func() {
			if runOnce {
				return
			}

			req := &CreateUserRequest{
				Email:     "eskit@gmail.com",
				FirstName: "Ahmet",
				LastName:  "Abi",
			}

			jsonData, err := json.Marshal(req)
			Expect(err).To(BeNil())

			resp, err := httpClient.Post(baseURL, "application/json", bytes.NewBuffer(jsonData))
			Expect(err).To(BeNil())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			var createResp UserResponse
			err = json.NewDecoder(resp.Body).Decode(&createResp)
			Expect(err).To(BeNil())
			deletedUserID = createResp.Originator.ID

			// Delete user
			deleteURL := fmt.Sprintf("%s?id=%s", baseURL, deletedUserID)
			deleteReq, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, deleteURL, nil)
			Expect(err).To(BeNil())

			deleteResp, err := httpClient.Do(deleteReq)
			Expect(err).To(BeNil())
			defer deleteResp.Body.Close()

			runOnce = true
		})

		It("Should not be returned via its ID", func() {
			getURL := fmt.Sprintf("%s?id=%s", baseURL, deletedUserID)
			getResp, err := httpClient.Get(getURL)
			Expect(err).To(BeNil())
			defer getResp.Body.Close()

			Expect(getResp.StatusCode).To(Equal(http.StatusNotFound))

			body, err := io.ReadAll(getResp.Body)
			Expect(err).To(BeNil())

			var errResp ErrorResponse
			err = json.Unmarshal(body, &errResp)
			Expect(err).To(BeNil())
			Expect(errResp.Error).To(Equal("not_found"))
		})
	})
})
