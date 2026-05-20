package info

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockInfoService struct {
	mock.Mock
}

func (m *MockInfoService) GetCandidates(areaID int) ([]Candidate, error) {
	args := m.Called(areaID)
	return args.Get(0).([]Candidate), args.Error(1)
}

func (m *MockInfoService) GetParties() ([]Party, error) {
	args := m.Called()
	return args.Get(0).([]Party), args.Error(1)
}

func TestGetCandidatesHandler_Success(t *testing.T) {
	svc := new(MockInfoService)
	svc.On("GetCandidates", 1).Return([]Candidate{{CandidateID: 1, CandidateNo: 1}}, nil)

	h := NewInfoHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/candidates?area_id=1", nil)
	w := httptest.NewRecorder()
	h.GetCandidatesHandler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result []Candidate
	json.Unmarshal(w.Body.Bytes(), &result)
	assert.Len(t, result, 1)
}

func TestGetCandidatesHandler_MissingAreaID(t *testing.T) {
	h := NewInfoHandler(new(MockInfoService))
	req := httptest.NewRequest(http.MethodGet, "/candidates", nil)
	w := httptest.NewRecorder()
	h.GetCandidatesHandler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetCandidatesHandler_InvalidAreaID(t *testing.T) {
	h := NewInfoHandler(new(MockInfoService))
	req := httptest.NewRequest(http.MethodGet, "/candidates?area_id=abc", nil)
	w := httptest.NewRecorder()
	h.GetCandidatesHandler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetCandidatesHandler_ServiceError(t *testing.T) {
	svc := new(MockInfoService)
	svc.On("GetCandidates", 1).Return([]Candidate{}, errors.New("db error"))

	h := NewInfoHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/candidates?area_id=1", nil)
	w := httptest.NewRecorder()
	h.GetCandidatesHandler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetPartiesHandler_Success(t *testing.T) {
	svc := new(MockInfoService)
	svc.On("GetParties").Return([]Party{{PartyID: 1, PartyName: "Test Party"}}, nil)

	h := NewInfoHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/parties", nil)
	w := httptest.NewRecorder()
	h.GetPartiesHandler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetPartiesHandler_ServiceError(t *testing.T) {
	svc := new(MockInfoService)
	svc.On("GetParties").Return([]Party{}, errors.New("db error"))

	h := NewInfoHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/parties", nil)
	w := httptest.NewRecorder()
	h.GetPartiesHandler().ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
