package integration

import (
	"encoding/json"
	"goschool/pkg/model"
	"io"
	"math/rand"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func toPtr[T any](v T) *T {
	return &v
}

func pickRandom[T any](arr ...T) T {
	return arr[rand.Intn(len(arr))]
}

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

func seedTeachers(t *testing.T, creCookies []*http.Cookie, newTeachers ...model.NewTeacher) (ids []int) {
	t.Helper()
	for i, teacher := range newTeachers {
		resp := requestJSON(t, http.MethodPost, "/teachers", teacher, withCookies(creCookies))
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode, "failed to create teacher %d: expected 201, got %d", i, resp.StatusCode)

		var teacherDetails model.TeacherDetails
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&teacherDetails), "decode teacher response")

		ids = append(ids, teacherDetails.ID)
	}
	return ids
}

func logAsJSON(t *testing.T, v any) {
	t.Helper()
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal to JSON: %v", err)
	}
	t.Log(string(data))
}
