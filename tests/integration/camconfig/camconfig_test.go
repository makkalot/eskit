package camconfig_test

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

type CreateCamConfigRequest struct {
	CameraID   string  `json:"cameraId"`
	Name       string  `json:"name"`
	Gamma      float64 `json:"gamma"`
	Exposure   int     `json:"exposure"`
	Saturation int     `json:"saturation"`
	Sharpness  int     `json:"sharpness"`
	Gain       int     `json:"gain"`
}

type UpdateCamConfigRequest struct {
	CameraID   string  `json:"cameraId,omitempty"`
	Name       string  `json:"name,omitempty"`
	Gamma      float64 `json:"gamma,omitempty"`
	Exposure   int     `json:"exposure,omitempty"`
	Saturation int     `json:"saturation,omitempty"`
	Sharpness  int     `json:"sharpness,omitempty"`
	Gain       int     `json:"gain,omitempty"`
}

type CamConfigResponse struct {
	Originator *types.Originator `json:"originator"`
	CameraID   string            `json:"cameraId"`
	Name       string            `json:"name"`
	Gamma      float64           `json:"gamma"`
	Exposure   int               `json:"exposure"`
	Saturation int               `json:"saturation"`
	Sharpness  int               `json:"sharpness"`
	Gain       int               `json:"gain"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

var _ = Describe("CamConfig", func() {

	var httpClient *http.Client
	var baseURL string

	BeforeEach(func() {
		httpClient = &http.Client{}
		baseURL = fmt.Sprintf("http://%s/v1/camconfigs", camConfigEndpoint)
	})

	Context("Health Check", func() {
		It("Should return healthy status", func() {
			healthURL := fmt.Sprintf("http://%s/v1/health", camConfigEndpoint)
			resp, err := httpClient.Get(healthURL)
			Expect(err).To(BeNil())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			var health map[string]string
			err = json.NewDecoder(resp.Body).Decode(&health)
			Expect(err).To(BeNil())
			Expect(health["status"]).To(Equal("ok"))
		})
	})

	Context("When CamConfig is Created", func() {
		It("Should be returned via its ID", func() {
			req := &CreateCamConfigRequest{
				CameraID:   "CAM-TEST-001",
				Name:       "Test Camera 1",
				Gamma:      1.2,
				Exposure:   1500,
				Saturation: 60,
				Sharpness:  55,
				Gain:       30,
			}

			// Create config
			jsonData, err := json.Marshal(req)
			Expect(err).To(BeNil())

			resp, err := httpClient.Post(baseURL, "application/json", bytes.NewBuffer(jsonData))
			Expect(err).To(BeNil())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			var createResp CamConfigResponse
			err = json.NewDecoder(resp.Body).Decode(&createResp)
			Expect(err).To(BeNil())
			Expect(createResp.Originator).NotTo(BeNil())
			Expect(createResp.Originator.ID).NotTo(BeEmpty())
			Expect(createResp.Originator.Version).To(Equal(uint64(1)))

			// Verify created values
			Expect(createResp.CameraID).To(Equal(req.CameraID))
			Expect(createResp.Name).To(Equal(req.Name))
			Expect(createResp.Gamma).To(Equal(req.Gamma))
			Expect(createResp.Exposure).To(Equal(req.Exposure))
			Expect(createResp.Saturation).To(Equal(req.Saturation))
			Expect(createResp.Sharpness).To(Equal(req.Sharpness))
			Expect(createResp.Gain).To(Equal(req.Gain))

			// Get config by ID
			getURL := fmt.Sprintf("%s?id=%s", baseURL, createResp.Originator.ID)
			getResp, err := httpClient.Get(getURL)
			Expect(err).To(BeNil())
			defer getResp.Body.Close()
			Expect(getResp.StatusCode).To(Equal(http.StatusOK))

			var getConfig CamConfigResponse
			err = json.NewDecoder(getResp.Body).Decode(&getConfig)
			Expect(err).To(BeNil())

			Expect(getConfig.CameraID).To(Equal(req.CameraID))
			Expect(getConfig.Name).To(Equal(req.Name))
			Expect(getConfig.Gamma).To(Equal(req.Gamma))
			Expect(getConfig.Exposure).To(Equal(req.Exposure))
			Expect(getConfig.Saturation).To(Equal(req.Saturation))
			Expect(getConfig.Sharpness).To(Equal(req.Sharpness))
			Expect(getConfig.Gain).To(Equal(req.Gain))
		})

		It("Should create multiple configs independently", func() {
			// Create first config
			req1 := &CreateCamConfigRequest{
				CameraID:   "CAM-TEST-002",
				Name:       "Test Camera 2",
				Gamma:      1.0,
				Exposure:   1000,
				Saturation: 50,
				Sharpness:  50,
				Gain:       25,
			}

			jsonData1, err := json.Marshal(req1)
			Expect(err).To(BeNil())

			resp1, err := httpClient.Post(baseURL, "application/json", bytes.NewBuffer(jsonData1))
			Expect(err).To(BeNil())
			defer resp1.Body.Close()
			Expect(resp1.StatusCode).To(Equal(http.StatusCreated))

			var createResp1 CamConfigResponse
			err = json.NewDecoder(resp1.Body).Decode(&createResp1)
			Expect(err).To(BeNil())

			// Create second config
			req2 := &CreateCamConfigRequest{
				CameraID:   "CAM-TEST-003",
				Name:       "Test Camera 3",
				Gamma:      1.5,
				Exposure:   2000,
				Saturation: 70,
				Sharpness:  60,
				Gain:       35,
			}

			jsonData2, err := json.Marshal(req2)
			Expect(err).To(BeNil())

			resp2, err := httpClient.Post(baseURL, "application/json", bytes.NewBuffer(jsonData2))
			Expect(err).To(BeNil())
			defer resp2.Body.Close()
			Expect(resp2.StatusCode).To(Equal(http.StatusCreated))

			var createResp2 CamConfigResponse
			err = json.NewDecoder(resp2.Body).Decode(&createResp2)
			Expect(err).To(BeNil())

			// Verify IDs are different
			Expect(createResp1.Originator.ID).NotTo(Equal(createResp2.Originator.ID))

			// Get first config
			getURL1 := fmt.Sprintf("%s?id=%s", baseURL, createResp1.Originator.ID)
			getResp1, err := httpClient.Get(getURL1)
			Expect(err).To(BeNil())
			defer getResp1.Body.Close()
			Expect(getResp1.StatusCode).To(Equal(http.StatusOK))

			var getConfig1 CamConfigResponse
			err = json.NewDecoder(getResp1.Body).Decode(&getConfig1)
			Expect(err).To(BeNil())
			Expect(getConfig1.CameraID).To(Equal(req1.CameraID))

			// Get second config
			getURL2 := fmt.Sprintf("%s?id=%s", baseURL, createResp2.Originator.ID)
			getResp2, err := httpClient.Get(getURL2)
			Expect(err).To(BeNil())
			defer getResp2.Body.Close()
			Expect(getResp2.StatusCode).To(Equal(http.StatusOK))

			var getConfig2 CamConfigResponse
			err = json.NewDecoder(getResp2.Body).Decode(&getConfig2)
			Expect(err).To(BeNil())
			Expect(getConfig2.CameraID).To(Equal(req2.CameraID))
		})
	})

	Context("When CamConfig is Updated", func() {
		var configID string
		var currentVersion uint64

		BeforeEach(func() {
			// Create a config to update
			req := &CreateCamConfigRequest{
				CameraID:   "CAM-UPDATE-TEST",
				Name:       "Update Test Camera",
				Gamma:      1.0,
				Exposure:   1000,
				Saturation: 50,
				Sharpness:  50,
				Gain:       25,
			}

			jsonData, err := json.Marshal(req)
			Expect(err).To(BeNil())

			resp, err := httpClient.Post(baseURL, "application/json", bytes.NewBuffer(jsonData))
			Expect(err).To(BeNil())
			defer resp.Body.Close()

			var createResp CamConfigResponse
			err = json.NewDecoder(resp.Body).Decode(&createResp)
			Expect(err).To(BeNil())

			configID = createResp.Originator.ID
			currentVersion = createResp.Originator.Version
		})

		It("Should increment version number", func() {
			updateReq := &UpdateCamConfigRequest{
				Gamma:    1.5,
				Exposure: 2000,
			}

			jsonData, err := json.Marshal(updateReq)
			Expect(err).To(BeNil())

			updateURL := fmt.Sprintf("%s?id=%s&version=%d", baseURL, configID, currentVersion)
			req, err := http.NewRequest(http.MethodPut, updateURL, bytes.NewBuffer(jsonData))
			Expect(err).To(BeNil())
			req.Header.Set("Content-Type", "application/json")

			resp, err := httpClient.Do(req)
			Expect(err).To(BeNil())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			var updateResp CamConfigResponse
			err = json.NewDecoder(resp.Body).Decode(&updateResp)
			Expect(err).To(BeNil())

			// Version should be incremented
			Expect(updateResp.Originator.Version).To(Equal(currentVersion + 1))
			Expect(updateResp.Gamma).To(Equal(1.5))
			Expect(updateResp.Exposure).To(Equal(2000))
		})

		It("Should preserve unchanged fields", func() {
			updateReq := &UpdateCamConfigRequest{
				Gamma: 1.8,
			}

			jsonData, err := json.Marshal(updateReq)
			Expect(err).To(BeNil())

			updateURL := fmt.Sprintf("%s?id=%s&version=%d", baseURL, configID, currentVersion)
			req, err := http.NewRequest(http.MethodPut, updateURL, bytes.NewBuffer(jsonData))
			Expect(err).To(BeNil())
			req.Header.Set("Content-Type", "application/json")

			resp, err := httpClient.Do(req)
			Expect(err).To(BeNil())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			var updateResp CamConfigResponse
			err = json.NewDecoder(resp.Body).Decode(&updateResp)
			Expect(err).To(BeNil())

			// Changed field
			Expect(updateResp.Gamma).To(Equal(1.8))
			// Unchanged fields
			Expect(updateResp.CameraID).To(Equal("CAM-UPDATE-TEST"))
			Expect(updateResp.Name).To(Equal("Update Test Camera"))
			Expect(updateResp.Exposure).To(Equal(1000))
			Expect(updateResp.Saturation).To(Equal(50))
		})

		It("Should support multiple updates", func() {
			// First update
			updateReq1 := &UpdateCamConfigRequest{
				Gamma: 1.2,
			}

			jsonData1, err := json.Marshal(updateReq1)
			Expect(err).To(BeNil())

			updateURL1 := fmt.Sprintf("%s?id=%s&version=%d", baseURL, configID, currentVersion)
			req1, err := http.NewRequest(http.MethodPut, updateURL1, bytes.NewBuffer(jsonData1))
			Expect(err).To(BeNil())
			req1.Header.Set("Content-Type", "application/json")

			resp1, err := httpClient.Do(req1)
			Expect(err).To(BeNil())
			defer resp1.Body.Close()
			Expect(resp1.StatusCode).To(Equal(http.StatusOK))

			var updateResp1 CamConfigResponse
			err = json.NewDecoder(resp1.Body).Decode(&updateResp1)
			Expect(err).To(BeNil())
			Expect(updateResp1.Originator.Version).To(Equal(uint64(2)))

			// Second update
			updateReq2 := &UpdateCamConfigRequest{
				Exposure: 1500,
			}

			jsonData2, err := json.Marshal(updateReq2)
			Expect(err).To(BeNil())

			updateURL2 := fmt.Sprintf("%s?id=%s&version=%d", baseURL, configID, 2)
			req2, err := http.NewRequest(http.MethodPut, updateURL2, bytes.NewBuffer(jsonData2))
			Expect(err).To(BeNil())
			req2.Header.Set("Content-Type", "application/json")

			resp2, err := httpClient.Do(req2)
			Expect(err).To(BeNil())
			defer resp2.Body.Close()
			Expect(resp2.StatusCode).To(Equal(http.StatusOK))

			var updateResp2 CamConfigResponse
			err = json.NewDecoder(resp2.Body).Decode(&updateResp2)
			Expect(err).To(BeNil())
			Expect(updateResp2.Originator.Version).To(Equal(uint64(3)))
			Expect(updateResp2.Gamma).To(Equal(1.2))
			Expect(updateResp2.Exposure).To(Equal(1500))
		})
	})

	Context("When CamConfig is Deleted", func() {
		var deletedConfigID string

		BeforeEach(func() {
			req := &CreateCamConfigRequest{
				CameraID:   "CAM-DELETE-TEST",
				Name:       "Delete Test Camera",
				Gamma:      1.0,
				Exposure:   1000,
				Saturation: 50,
				Sharpness:  50,
				Gain:       25,
			}

			jsonData, err := json.Marshal(req)
			Expect(err).To(BeNil())

			resp, err := httpClient.Post(baseURL, "application/json", bytes.NewBuffer(jsonData))
			Expect(err).To(BeNil())
			defer resp.Body.Close()

			var createResp CamConfigResponse
			err = json.NewDecoder(resp.Body).Decode(&createResp)
			Expect(err).To(BeNil())
			deletedConfigID = createResp.Originator.ID

			// Delete config
			deleteURL := fmt.Sprintf("%s?id=%s", baseURL, deletedConfigID)
			deleteReq, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, deleteURL, nil)
			Expect(err).To(BeNil())

			deleteResp, err := httpClient.Do(deleteReq)
			Expect(err).To(BeNil())
			defer deleteResp.Body.Close()
			Expect(deleteResp.StatusCode).To(Equal(http.StatusOK))
		})

		It("Should not be returned via its ID", func() {
			getURL := fmt.Sprintf("%s?id=%s", baseURL, deletedConfigID)
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

	Context("Error Handling", func() {
		It("Should return 400 for invalid JSON", func() {
			invalidJSON := []byte(`{"cameraId": "invalid json}`)

			resp, err := httpClient.Post(baseURL, "application/json", bytes.NewBuffer(invalidJSON))
			Expect(err).To(BeNil())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

			var errResp ErrorResponse
			err = json.NewDecoder(resp.Body).Decode(&errResp)
			Expect(err).To(BeNil())
			Expect(errResp.Error).To(Equal("invalid_request"))
		})

		It("Should return 404 for non-existent config", func() {
			getURL := fmt.Sprintf("%s?id=non-existent-id", baseURL)
			getResp, err := httpClient.Get(getURL)
			Expect(err).To(BeNil())
			defer getResp.Body.Close()

			Expect(getResp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("Should return 400 when id is missing", func() {
			getURL := fmt.Sprintf("%s", baseURL)
			getResp, err := httpClient.Get(getURL)
			Expect(err).To(BeNil())
			defer getResp.Body.Close()

			Expect(getResp.StatusCode).To(Equal(http.StatusBadRequest))

			var errResp ErrorResponse
			err = json.NewDecoder(getResp.Body).Decode(&errResp)
			Expect(err).To(BeNil())
			Expect(errResp.Error).To(Equal("invalid_request"))
		})
	})
})
