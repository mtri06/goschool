package integration

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"goschool/pkg/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validNewTeacher() *model.NewTeacher {
	return &model.NewTeacher{
		Username:      "fake_teacher_" + strconv.Itoa(rand.Intn(1000)),
		Password:      "Password1!",
		Name:          "Fake Teacher " + strconv.Itoa(rand.Intn(1000)),
		DateOfBirth:   time.Date(1990, 6, 15, 0, 0, 0, 0, time.UTC),
		Gender:        pickRandom("male", "female"),
		HireDate:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		WorkingStatus: "active",
	}
}

func validUpdateTeacher() *model.UpdateTeacher {
	return &model.UpdateTeacher{
		Name:          "Updated Teacher",
		DateOfBirth:   time.Date(1990, 6, 15, 0, 0, 0, 0, time.UTC),
		Gender:        "male",
		SubjectID:     toPtr(2),
		HireDate:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		WorkingStatus: "on_leave",
	}
}

func TestTeacher_CRUD(t *testing.T) {
	t.Cleanup(func() { clearDB(t) })

	cookies := loginAsAdmin(t)

	// 1. Seed subjects (required for teacher's subject_id foreign key)
	subjectIDs := seedSubjects(t, cookies, "Math", "Science")

	// ── 1. Create a teacher ──────────────────────────────────────────────────
	// Seed a subject first (required foreign key). Call the DB directly via
	// the app's own endpoint if available, or skip if not yet exposed.
	// For now we rely on the migration having seeded at least subject id=1.
	// If your migrations don't seed a subject, run an INSERT here via sqlx.

	newTeacher := validNewTeacher()
	newTeacher.SubjectID = &subjectIDs[0] // Assign to first seeded subject

	createResp := requestJSON(t, http.MethodPost, "/teachers", newTeacher, withCookies(cookies))
	defer createResp.Body.Close()

	respBody, err := io.ReadAll(createResp.Body)
	require.NoError(t, err)
	require.Equalf(t, http.StatusCreated, createResp.StatusCode, "expected status code 201 with response: %s", respBody)

	var created model.TeacherDetails
	require.NoError(t, json.Unmarshal(respBody, &created), "decode create response")
	require.NotZero(t, created.ID, "expected created teacher to have non-zero ID")

	// ── 2. List teachers — should include the new one ─────────────────────────
	listResp := requestJSON(t, http.MethodGet, "/teachers", nil, withCookies(cookies))
	var listBody struct {
		Teachers []model.TeacherDetails `json:"teachers"`
		Total    int                    `json:"total"`
	}
	decodeJSON(t, listResp, &listBody)

	require.Equal(t, http.StatusOK, listResp.StatusCode, "expected status code 200")
	require.Equal(t, 1, listBody.Total, "expected total 1 teacher")
	require.Len(t, listBody.Teachers, 1, "expected 1 teacher in list")
	require.Equal(t, created.ID, listBody.Teachers[0].ID, "expected teacher ID to match created teacher")

	// ── 3. Get by ID ──────────────────────────────────────────────────────────
	getResp := requestJSON(t, http.MethodGet, "/teachers/"+itoa(created.ID), nil, withCookies(cookies))
	var got model.TeacherDetails
	decodeJSON(t, getResp, &got)

	require.Equal(t, http.StatusOK, getResp.StatusCode, "expected status code 200")
	require.Equal(t, created.ID, got.ID, "expected teacher ID to match")
	require.Equal(t, created.Username, got.Username, "expected username to match")

	// ── 4. Update the teacher ─────────────────────────────────────────────────
	update := validUpdateTeacher()
	update.SubjectID = &subjectIDs[1]
	update.Name = "Updated Teacher"
	update.WorkingStatus = "on_leave"

	updateResp := requestJSON(t, http.MethodPut, "/teachers/"+itoa(created.ID), update, withCookies(cookies))
	defer updateResp.Body.Close()

	require.Equal(t, http.StatusNoContent, updateResp.StatusCode, "expected status code 204")

	// ── 5. Verify update ──────────────────────────────────────────────────────
	getAfterUpdate := requestJSON(t, http.MethodGet, "/teachers/"+itoa(created.ID), nil, withCookies(cookies))
	var updated model.TeacherDetails
	decodeJSON(t, getAfterUpdate, &updated)

	require.Equal(t, "Updated Teacher", updated.Name, "expected updated name 'Updated Teacher'")
	require.Equal(t, "on_leave", updated.WorkingStatus, "expected working_status 'on_leave'")

	// ── 6. Delete the teacher ─────────────────────────────────────────────────
	deleteResp := requestJSON(t, http.MethodDelete, "/teachers/"+itoa(created.ID), nil, withCookies(cookies))
	defer deleteResp.Body.Close()

	require.Equal(t, http.StatusNoContent, deleteResp.StatusCode, "expected status code 204")

	// ── 7. Get after delete — should be 404 ───────────────────────────────────
	getAfterDelete := requestJSON(t, http.MethodGet, "/teachers/"+itoa(created.ID), nil, withCookies(cookies))
	defer getAfterDelete.Body.Close()

	require.Equal(t, http.StatusNotFound, getAfterDelete.StatusCode, "expected status code 404")
}

func TestTeacher_GetList_Unauthorized(t *testing.T) {
	t.Cleanup(func() { clearDB(t) })

	resp := requestJSON(t, http.MethodGet, "/teachers", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d", resp.StatusCode)
	}
}

func TestTeacher_GetByID_NotFound(t *testing.T) {
	t.Cleanup(func() { clearDB(t) })

	cookies := loginAsAdmin(t)

	resp := requestJSON(t, http.MethodGet, "/teachers/99999", nil, withCookies(cookies))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 for non-existent teacher, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// GetList Teachers Tests
// ---------------------------------------------------------------------------

func TestTeacher_GetList_Success_BareURL(t *testing.T) {
	t.Cleanup(func() { clearDB(t) })

	cookies := loginAsAdmin(t)

	// Create 3 teachers with predictable names
	var newTeachers []model.NewTeacher
	for i := 1; i <= 3; i++ {
		teacher := validNewTeacher()
		teacher.Username += itoa(i)
		newTeachers = append(newTeachers, *teacher)
	}
	seedTeachers(t, cookies, newTeachers...)

	// Call bare URL without any query parameters
	listResp := requestJSON(t, http.MethodGet, "/teachers", nil, withCookies(cookies))
	var listBody struct {
		Teachers []model.TeacherDetails `json:"teachers"`
		Total    int                    `json:"total"`
	}
	decodeJSON(t, listResp, &listBody)

	require.Equal(t, listResp.StatusCode, http.StatusOK, "expect status code 200")
	require.NotNil(t, listBody.Teachers, "teachers should not be nil")
	assert.Equal(t, 3, listBody.Total, "expected total 3 teachers")
	assert.Equal(t, 3, len(listBody.Teachers), "expected 3 teachers in response")

	// Verify returned objects have expected names
	for i, teacher := range listBody.Teachers {
		if i >= len(newTeachers) {
			break
		}
		assert.Equal(t, newTeachers[i].Name, teacher.Name, "expected teacher name at index %d", i)
	}
}

func TestTeacher_GetList_Paging_Success(t *testing.T) {
	t.Cleanup(func() { clearDB(t) })

	cookies := loginAsAdmin(t)
	subjectIDs := seedSubjects(t, cookies, "Math", "Physics")

	// Create 25 teachers (more than default page size of 20)
	// Store the created teachers for later verification
	newTeachers := make([]model.NewTeacher, 0, 25)
	for i := 1; i <= 25; i++ {
		teacher := validNewTeacher()
		teacher.Username += itoa(i)
		teacher.SubjectID = &subjectIDs[0]
		newTeachers = append(newTeachers, *teacher)
	}
	seedTeachers(t, cookies, newTeachers...)

	expectTotal := 25
	tests := []struct {
		name            string
		page            *int
		pageSize        *int
		shouldHaveError bool
		expectTeachers  []model.NewTeacher
	}{
		{
			name:           "First page with default size",
			page:           nil,
			pageSize:       nil,
			expectTeachers: newTeachers[:20],
		},
		{
			name:           "Second page with default size",
			page:           toPtr(2),
			pageSize:       nil,
			expectTeachers: newTeachers[20:25],
		},
		{
			name:           "Custom page size 10, page 1",
			page:           toPtr(1),
			pageSize:       toPtr(10),
			expectTeachers: newTeachers[:10],
		},
		{
			name:           "Custom page size 12, page 2",
			page:           toPtr(2),
			pageSize:       toPtr(12),
			expectTeachers: newTeachers[12:24],
		},
		{
			name:           "Custom page size 10, page 3",
			page:           toPtr(3),
			pageSize:       toPtr(10),
			expectTeachers: newTeachers[20:25],
		},
		{
			name:            "Invalid page (0) should use default page 1",
			page:            toPtr(0),
			pageSize:        nil,
			shouldHaveError: false,
			expectTeachers:  newTeachers[:20],
		},
		{
			name:            "Negative page should use default page 1",
			page:            toPtr(-5),
			pageSize:        toPtr(15),
			shouldHaveError: false,
			expectTeachers:  newTeachers[:15],
		},
		{
			name:            "Invalid pageSize (0) should use default 20",
			page:            toPtr(1),
			pageSize:        toPtr(0),
			shouldHaveError: false,
			expectTeachers:  newTeachers[:20],
		},
		{
			name:            "Negative pageSize should use default 20",
			page:            toPtr(1),
			pageSize:        toPtr(-10),
			shouldHaveError: false,
			expectTeachers:  newTeachers[:20],
		},
		{
			name:            "PageSize > 100 should use 100",
			page:            toPtr(1),
			pageSize:        toPtr(150),
			shouldHaveError: false,
			expectTeachers:  newTeachers[:25],
		},
		{
			name:           "Max valid pageSize (100)",
			page:           toPtr(1),
			pageSize:       toPtr(100),
			expectTeachers: newTeachers[:25],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/teachers"
			queryParts := []string{}
			if tt.page != nil {
				queryParts = append(queryParts, "page="+itoa(*tt.page))
			}
			if tt.pageSize != nil {
				queryParts = append(queryParts, "pageSize="+itoa(*tt.pageSize))
			}
			if len(queryParts) > 0 {
				url += "?" + queryParts[0]
				for _, part := range queryParts[1:] {
					url += "&" + part
				}
			}

			listResp := requestJSON(t, http.MethodGet, url, nil, withCookies(cookies))
			var listBody struct {
				Teachers []model.TeacherDetails `json:"teachers"`
				Total    int                    `json:"total"`
			}
			decodeJSON(t, listResp, &listBody)

			if tt.shouldHaveError {
				if listResp.StatusCode != http.StatusBadRequest && listResp.StatusCode != http.StatusUnprocessableEntity {
					t.Errorf("expected error status, got %d", listResp.StatusCode)
				}
				return
			}

			require.Equal(t, http.StatusOK, listResp.StatusCode, "expected status 200")
			assert.Equal(t, expectTotal, listBody.Total, "expected total teachers")
			assert.Equal(t, len(tt.expectTeachers), len(listBody.Teachers), "expected number of teachers in page")
			for i, teacher := range listBody.Teachers {
				require.NotEqual(t, 0, teacher.ID, "teacher[%d] ID should not be 0", i)
				require.Equal(t, tt.expectTeachers[i].Name, teacher.Name, "teacher[%d] name should match", i)
				require.Equal(t, tt.expectTeachers[i].Username, teacher.Username, "teacher[%d] username should match", i)
			}
		})
	}
}

func TestTeacher_GetList_Filter_ByName(t *testing.T) {
	t.Cleanup(func() { clearDB(t) })

	cookies := loginAsAdmin(t)

	// Create names with different names
	names := []string{
		"Alice Johnson",
		"Bob Smith",
		"Charlie Brown",
		"David Lee",
		"John Lee",
		"Alice Wong", // Another Alice
	}
	var newTeachers []model.NewTeacher
	for i, name := range names {
		newTeacher := validNewTeacher()
		newTeacher.Username += itoa(i)
		newTeacher.Name = name
		newTeachers = append(newTeachers, *newTeacher)
	}
	seedTeachers(t, cookies, newTeachers...)

	tests := []struct {
		name          string
		filterName    string
		expectedTotal int
		expectedNames []string
	}{
		{
			name:          "Filter by 'Alice' (case-insensitive)",
			filterName:    "Alice",
			expectedTotal: 2,
			expectedNames: []string{"Alice Johnson", "Alice Wong"},
		},
		{
			name:          "Filter by lowercase 'alice'",
			filterName:    "alice",
			expectedTotal: 2,
			expectedNames: []string{"Alice Johnson", "Alice Wong"},
		},
		{
			name:          "Filter by 'ALICE' (uppercase)",
			filterName:    "ALICE",
			expectedTotal: 2,
			expectedNames: []string{"Alice Johnson", "Alice Wong"},
		},
		{
			name:          "Filter by 'Bob'",
			filterName:    "Bob",
			expectedTotal: 1,
			expectedNames: []string{"Bob Smith"},
		},
		{
			name:          "Filter by 'Lee' (partial match)",
			filterName:    "Lee",
			expectedTotal: 2,
			expectedNames: []string{"David Lee", "John Lee"},
		},
		{
			name:          "Filter by non-existent name",
			filterName:    "NonExistent",
			expectedTotal: 0,
			expectedNames: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/teachers?name=" + tt.filterName
			listResp := requestJSON(t, http.MethodGet, url, nil, withCookies(cookies))
			var listBody struct {
				Teachers []model.TeacherDetails `json:"teachers"`
				Total    int                    `json:"total"`
			}
			decodeJSON(t, listResp, &listBody)

			require.Equal(t, http.StatusOK, listResp.StatusCode, "expected status 200")
			assert.Equal(t, tt.expectedTotal, listBody.Total, "expected total matching teachers")
			assert.Equal(t, len(tt.expectedNames), len(listBody.Teachers), "expected number of teachers in response")
			for i, teacher := range listBody.Teachers {
				assert.Equal(t, tt.expectedNames[i], teacher.Name, "teacher[%d] name should match", i)
			}
		})
	}
}

func TestTeacher_GetList_Filter_ByWorkingStatus(t *testing.T) {
	t.Cleanup(func() { clearDB(t) })

	cookies := loginAsAdmin(t)

	// Create teachers with different working statuses
	statuses := []string{"active", "inactive", "on_leave", "active", "on_leave"}
	var newTeachers []model.NewTeacher
	for i, status := range statuses {
		teacher := validNewTeacher()
		teacher.Username += itoa(i)
		teacher.WorkingStatus = status
		newTeachers = append(newTeachers, *teacher)
	}
	seedTeachers(t, cookies, newTeachers...)

	tests := []struct {
		name             string
		status           string
		expectedTeachers []model.NewTeacher
	}{
		{
			name:   "Filter by 'active'",
			status: "active",
			expectedTeachers: []model.NewTeacher{
				newTeachers[0], // active
				newTeachers[3], // active
			},
		},
		{
			name:   "Filter by 'inactive'",
			status: "inactive",
			expectedTeachers: []model.NewTeacher{
				newTeachers[1], // inactive
			},
		},
		{
			name:   "Filter by 'on_leave'",
			status: "on_leave",
			expectedTeachers: []model.NewTeacher{
				newTeachers[2], // on_leave
				newTeachers[4], // on_leave
			},
		},
		{
			name:   "Filter by non-existent status",
			status: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/teachers?workingStatus=" + tt.status
			listResp := requestJSON(t, http.MethodGet, url, nil, withCookies(cookies))
			var listBody struct {
				Teachers []model.TeacherDetails `json:"teachers"`
				Total    int                    `json:"total"`
			}
			decodeJSON(t, listResp, &listBody)

			require.Equal(t, http.StatusOK, listResp.StatusCode, "expected status 200")
			// Verify all returned teachers have the correct status
			for i, teacher := range listBody.Teachers {
				assert.Equal(t, tt.status, teacher.WorkingStatus, "teacher[%d] should have status %q", i, tt.status)
				assert.Equal(t, tt.expectedTeachers[i].Name, teacher.Name, "teacher[%d] name should match expected", i)
				assert.Equal(t, tt.expectedTeachers[i].Username, teacher.Username, "teacher[%d] username should match expected", i)
			}
		})
	}
}

func TestTeacher_GetList_EmptyList_Returns_EmptyArray(t *testing.T) {
	t.Cleanup(func() { clearDB(t) })

	cookies := loginAsAdmin(t)

	// No teachers created - list should be empty
	listResp := requestJSON(t, http.MethodGet, "/teachers", nil, withCookies(cookies))
	var listBody struct {
		Teachers []model.TeacherDetails `json:"teachers"`
		Total    int                    `json:"total"`
	}
	decodeJSON(t, listResp, &listBody)

	require.Equal(t, http.StatusOK, listResp.StatusCode, "expected status 200")
	require.NotNil(t, listBody.Teachers, "teachers should not be nil")
	assert.Equal(t, 0, listBody.Total, "expected total 0 teachers")
	assert.Equal(t, 0, len(listBody.Teachers), "expected 0 teachers in response")
}

func TestTeacher_GetList_Combined(t *testing.T) {
	t.Cleanup(func() { clearDB(t) })

	cookies := loginAsAdmin(t)
	subjectIDs := seedSubjects(t, cookies, "Math", "Physics", "Chemistry")
	var newTeachers []model.NewTeacher

	testTeachers := []struct {
		Name          string
		WorkingStatus string
	}{
		{"Alice Johnson", "active"},    // 0
		{"Bob Smith", "active"},        // 1
		{"Charlie Brown", "on_leave"},  // 2
		{"Milo Lee", "active"},         // 3
		{"John Lee", "inactive"},       // 4
		{"Alice Wong", "on_leave"},     // 5
		{"Bob Lee", "inactive"},        // 6
		{"Charlie Davis", "active"},    // 7
		{"David Kim", "on_leave"},      // 8
		{"John Kim", "active"},         // 9
		{"john doe", "active"},         // 10
		{"alicecooper", "on_leave"},    // 11
		{"bob-cooper", "active"},       // 12
		{"charLie coOper", "on_leave"}, // 13
		{"david cooper", "active"},     // 14
	}

	for i, tt := range testTeachers {
		subject := pickRandom(subjectIDs...)

		teacher := validNewTeacher()

		teacher.Name = tt.Name
		teacher.WorkingStatus = tt.WorkingStatus
		teacher.Username += itoa(i)
		teacher.SubjectID = &subject

		newTeachers = append(newTeachers, *teacher)
	}

	seedTeachers(t, cookies, newTeachers...)

	tests := []struct {
		name             string
		page             *int
		pageSize         *int
		nameFilter       *string
		statusFilter     *string
		expectedTotal    int
		expectedTeachers []model.NewTeacher
	}{
		{
			name:          "Alice, size 5",
			pageSize:      toPtr(5),
			nameFilter:    toPtr("Alice"),
			expectedTotal: 3,
			expectedTeachers: []model.NewTeacher{
				newTeachers[0],  // Alice Johnson
				newTeachers[5],  // Alice Wong
				newTeachers[11], // alicecooper
			},
		},
		{
			name:          "Alice, size 2",
			pageSize:      toPtr(2),
			nameFilter:    toPtr("Alice"),
			expectedTotal: 3,
			expectedTeachers: []model.NewTeacher{
				newTeachers[0], // Alice Johnson
				newTeachers[5], // Alice Wong
			},
		},
		{
			name:          "Alice, page 2, size 2",
			page:          toPtr(2),
			pageSize:      toPtr(2),
			nameFilter:    toPtr("Alice"),
			expectedTotal: 3,
			expectedTeachers: []model.NewTeacher{
				newTeachers[11], // alicecooper
			},
		},
		{
			name:          "alice",
			nameFilter:    toPtr("alice"),
			expectedTotal: 3,
			expectedTeachers: []model.NewTeacher{
				newTeachers[0],  // Alice Johnson
				newTeachers[5],  // Alice Wong
				newTeachers[11], // alicecooper
			},
		},
		{
			name:             "page 10, size 5 (beyond total)",
			page:             toPtr(10),
			pageSize:         toPtr(5),
			expectedTotal:    15,
			expectedTeachers: []model.NewTeacher{
				// No teachers expected in this page
			},
		},
		{
			name:             "No query parameters (default paging)",
			expectedTotal:    15,
			expectedTeachers: newTeachers, // All teachers should be returned since total < default page size
		},
		{
			name:          "active, page 2, size 5",
			page:          toPtr(2),
			pageSize:      toPtr(5),
			statusFilter:  toPtr("active"),
			expectedTotal: 8,
			expectedTeachers: []model.NewTeacher{
				newTeachers[10],
				newTeachers[12],
				newTeachers[14],
			},
		},
		{
			name:          "inactive",
			statusFilter:  toPtr("inactive"),
			expectedTotal: 2,
			expectedTeachers: []model.NewTeacher{
				newTeachers[4],
				newTeachers[6],
			},
		},
		{
			name:          "inactive, john",
			nameFilter:    toPtr("john"),
			statusFilter:  toPtr("inactive"),
			expectedTotal: 1,
			expectedTeachers: []model.NewTeacher{
				newTeachers[4], // John Lee (inactive)
			},
		},
		{
			name:          "cooper",
			nameFilter:    toPtr("cooper"),
			expectedTotal: 4,
			expectedTeachers: []model.NewTeacher{
				newTeachers[11], // alicecooper
				newTeachers[12], // bob-cooper
				newTeachers[13], // charLie coOper
				newTeachers[14], // david cooper
			},
		},
		{
			name:          "mi",
			nameFilter:    toPtr("mi"),
			expectedTotal: 2,
			expectedTeachers: []model.NewTeacher{
				newTeachers[1], // Milo Lee
				newTeachers[3], // David Kim
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := url.Values{}

			if tt.page != nil {
				q.Set("page", strconv.Itoa(*tt.page))
			}
			if tt.pageSize != nil {
				q.Set("pageSize", strconv.Itoa(*tt.pageSize))
			}
			if tt.nameFilter != nil {
				q.Set("name", *tt.nameFilter)
			}
			if tt.statusFilter != nil && *tt.statusFilter != "" {
				q.Set("workingStatus", *tt.statusFilter)
			}

			path := "/teachers"
			if len(q) > 0 {
				path += "?" + q.Encode()
			}
			listResp := requestJSON(t, http.MethodGet, path, nil, withCookies(cookies))
			var listBody struct {
				Teachers []model.TeacherDetails `json:"teachers"`
				Total    int                    `json:"total"`
			}
			decodeJSON(t, listResp, &listBody)

			require.Equal(t, http.StatusOK, listResp.StatusCode, "expected status 200")
			assert.Equal(t, tt.expectedTotal, listBody.Total, "expected total matching teachers")
			assert.Equal(t, len(tt.expectedTeachers), len(listBody.Teachers), "expected number of teachers in page")

			// Verify all returned teachers match the filters and status
			for i, teacher := range listBody.Teachers {
				require.NotZero(t, teacher.ID, "teacher[%d] ID should not be 0", i)
				if tt.expectedTeachers[i].Name != teacher.Name {
					logAsJSON(t, tt.expectedTeachers)
					logAsJSON(t, listBody)
				}
				require.Equal(t, tt.expectedTeachers[i].Name, teacher.Name, "teacher[%d] name should match expected", i)
				require.Equal(t, tt.expectedTeachers[i].Username, teacher.Username, "teacher[%d] username should match expected", i)
				if tt.nameFilter != nil {
					require.Contains(t, strings.ToLower(teacher.Name), strings.ToLower(*tt.nameFilter),
						"teacher[%d] name should contain filter", i)
				}
				// If status filter is set, verify it matches
				if tt.statusFilter != nil && *tt.statusFilter != "" {
					require.Equal(t, *tt.statusFilter, teacher.WorkingStatus, "teacher[%d] status should match", i)
				}
			}
		})
	}
}
