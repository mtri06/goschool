package integration

import (
	"encoding/json"
	"goschool/pkg/model"
	"io"
	"net/http"
	"testing"
)

func seedSubjects(t *testing.T, creCookies []*http.Cookie, subjectNames ...string) []int {
	t.Helper()
	var subjectIDs []int
	for _, name := range subjectNames {
		newSubject := model.NewSubject{Name: name}
		subjectResp := requestJSON(t, http.MethodPost, "/subjects", newSubject, withCookies(creCookies))
		defer subjectResp.Body.Close()

		if subjectResp.StatusCode != http.StatusCreated {
			respBody, _ := io.ReadAll(subjectResp.Body)
			t.Fatalf("seed subject: expected 201, got %d with response: %s", subjectResp.StatusCode, respBody)
		}

		var subject model.SubjectDetails
		if err := json.NewDecoder(subjectResp.Body).Decode(&subject); err != nil {
			t.Fatalf("decode subject response: %v", err)
		}
		subjectIDs = append(subjectIDs, subject.ID)
	}
	return subjectIDs
}
