package handler_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	internalexecutor "github.com/sourcegraph/sourcegraph/internal/executor"
	metricsstore "github.com/sourcegraph/sourcegraph/internal/metrics/store"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	workerstoremocks "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store/mocks"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestHandler_Name(t *testing.T) {
	queueHandler := handler.QueueHandler[testRecord]{Name: "test"}
	h := handler.NewHandler(
		database.NewMockExecutorStore(),
		executor.NewMockJobTokenStore(),
		metricsstore.NewMockDistributedStore(),
		queueHandler,
	)
	assert.Equal(t, "test", h.Name())
}

func TestHandler_AuthMiddleware(t *testing.T) {
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{ExecutorsAccessToken: "hunter2"}})

	tests := []struct {
		name                 string
		header               http.Header
		body                 string
		mockFunc             func(executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore)
		expectedStatusCode   int
		expectedResponseBody string
		assertionFunc        func(t *testing.T, executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore)
	}{
		{
			name:   "Authorized",
			header: http.Header{"Authorization": []string{"Bearer somejobtoken"}},
			body:   `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
				jobTokenStore.GetByTokenFunc.PushReturn(executor.JobToken{JobId: 42, Queue: "test"}, nil)
				executorStore.GetByHostnameFunc.PushReturn(types.Executor{}, true, nil)
			},
			expectedStatusCode: http.StatusTeapot,
			assertionFunc: func(t *testing.T, executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
				assert.Equal(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
				require.Len(t, executorStore.GetByHostnameFunc.History(), 1)
				assert.Equal(t, executorStore.GetByHostnameFunc.History()[0].Arg1, "test-executor")
			},
		},
		{
			name:   "Authorized general access token",
			header: http.Header{"Authorization": []string{"token-executor hunter2"}},
			body:   `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
				jobTokenStore.GetByTokenFunc.PushReturn(executor.JobToken{JobId: 42, Queue: "test"}, nil)
				executorStore.GetByHostnameFunc.PushReturn(types.Executor{}, true, nil)
			},
			expectedStatusCode: http.StatusTeapot,
			assertionFunc: func(t *testing.T, executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 0)
				require.Len(t, executorStore.GetByHostnameFunc.History(), 0)
			},
		},
		{
			name:                 "No request body",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "No request body provided\n",
		},
		{
			name:                 "Malformed request body",
			body:                 `{"executorName": "test-executor"`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Failed to parse request body\n",
		},
		{
			name:                 "No worker hostname provided",
			body:                 `{"jobId": 42}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "worker hostname cannot be empty\n",
		},
		{
			name:                 "No Authorized header",
			body:                 `{"executorName": "test-executor", "jobId": 42}`,
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: "no token value in the HTTP Authorization request header\n",
		},
		{
			name:                 "Invalid Authorized header parts",
			header:               http.Header{"Authorization": []string{"token-executor"}},
			body:                 `{"executorName": "test-executor", "jobId": 42}`,
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: "HTTP Authorization request header value must be of the following form: 'Bearer \"TOKEN\"' or 'token-executor TOKEN'\n",
		},
		{
			name:                 "Invalid Authorized header prefix",
			header:               http.Header{"Authorization": []string{"Foo bar"}},
			body:                 `{"executorName": "test-executor", "jobId": 42}`,
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: "unrecognized HTTP Authorization request header scheme (supported values: \"Bearer\", \"token-executor\")\n",
		},
		{
			name:               "Invalid general access token",
			header:             http.Header{"Authorization": []string{"token-executor hunter1"}},
			body:               `{"executorName": "test-executor", "jobId": 42}`,
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:   "Failed to retrieve job token",
			header: http.Header{"Authorization": []string{"Bearer somejobtoken"}},
			body:   `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
				jobTokenStore.GetByTokenFunc.PushReturn(executor.JobToken{}, errors.New("failed to find job token"))
			},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: "invalid token\n",
			assertionFunc: func(t *testing.T, executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
				assert.Equal(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
				require.Len(t, executorStore.GetByHostnameFunc.History(), 0)
			},
		},
		{
			name:   "JobId does not match",
			header: http.Header{"Authorization": []string{"Bearer somejobtoken"}},
			body:   `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
				jobTokenStore.GetByTokenFunc.PushReturn(executor.JobToken{JobId: 7, Queue: "test"}, nil)
			},
			expectedStatusCode:   http.StatusForbidden,
			expectedResponseBody: "invalid token\n",
			assertionFunc: func(t *testing.T, executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
				assert.Equal(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
				require.Len(t, executorStore.GetByHostnameFunc.History(), 0)
			},
		},
		{
			name:   "Queue does not match",
			header: http.Header{"Authorization": []string{"Bearer somejobtoken"}},
			body:   `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
				jobTokenStore.GetByTokenFunc.PushReturn(executor.JobToken{JobId: 42, Queue: "test1"}, nil)
			},
			expectedStatusCode:   http.StatusForbidden,
			expectedResponseBody: "invalid token\n",
			assertionFunc: func(t *testing.T, executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
				assert.Equal(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
				require.Len(t, executorStore.GetByHostnameFunc.History(), 0)
			},
		},
		{
			name:   "Executor host does not exist",
			header: http.Header{"Authorization": []string{"Bearer somejobtoken"}},
			body:   `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
				jobTokenStore.GetByTokenFunc.PushReturn(executor.JobToken{JobId: 42, Queue: "test"}, nil)
				executorStore.GetByHostnameFunc.PushReturn(types.Executor{}, false, errors.New("executor does not exist"))
			},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: "invalid token\n",
			assertionFunc: func(t *testing.T, executorStore *database.MockExecutorStore, jobTokenStore *executor.MockJobTokenStore) {
				require.Len(t, jobTokenStore.GetByTokenFunc.History(), 1)
				assert.Equal(t, jobTokenStore.GetByTokenFunc.History()[0].Arg1, "somejobtoken")
				require.Len(t, executorStore.GetByHostnameFunc.History(), 1)
				assert.Equal(t, executorStore.GetByHostnameFunc.History()[0].Arg1, "test-executor")
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executorStore := database.NewMockExecutorStore()
			jobTokenStore := executor.NewMockJobTokenStore()

			h := handler.NewHandler(
				executorStore,
				jobTokenStore,
				metricsstore.NewMockDistributedStore(),
				handler.QueueHandler[testRecord]{},
			)

			router := mux.NewRouter()
			router.HandleFunc("/{queueName}", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})
			router.Use(h.AuthMiddleware)

			req, err := http.NewRequest("GET", "/test", strings.NewReader(test.body))
			require.NoError(t, err)
			req.Header = test.header

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(executorStore, jobTokenStore)
			}

			router.ServeHTTP(rw, req)

			assert.Equal(t, test.expectedStatusCode, rw.Code)

			b, err := io.ReadAll(rw.Body)
			require.NoError(t, err)
			assert.Equal(t, test.expectedResponseBody, string(b))

			if test.assertionFunc != nil {
				test.assertionFunc(t, executorStore, jobTokenStore)
			}
		})
	}
}

func TestHandler_HandleDequeue(t *testing.T) {
	tests := []struct {
		name                 string
		body                 string
		transformerFunc      handler.TransformerFunc[testRecord]
		mockFunc             func(mockStore *workerstoremocks.MockStore[testRecord], jobTokenStore *executor.MockJobTokenStore)
		expectedStatusCode   int
		expectedResponseBody string
		assertionFunc        func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], jobTokenStore *executor.MockJobTokenStore)
	}{
		{
			name: "Dequeue record",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
			transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executor.Job, error) {
				return executor.Job{ID: record.RecordID()}, nil
			},
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord], jobTokenStore *executor.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
				jobTokenStore.CreateFunc.PushReturn("sometoken", nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":1,"token":"sometoken","repositoryName":"","repositoryDirectory":"","commit":"","fetchTags":false,"shallowClone":false,"sparseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redactedValues":null}`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], jobTokenStore *executor.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				assert.Equal(t, "test-executor", mockStore.DequeueFunc.History()[0].Arg1)
				assert.Nil(t, mockStore.DequeueFunc.History()[0].Arg2)
				require.Len(t, jobTokenStore.CreateFunc.History(), 1)
				assert.Equal(t, 1, jobTokenStore.CreateFunc.History()[0].Arg1)
				assert.Equal(t, "test", jobTokenStore.CreateFunc.History()[0].Arg2)
			},
		},
		{
			name:                 "Invalid version",
			body:                 `{"executorName": "test-executor", "version":"\n1.2", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"Invalid Semantic Version"}`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], jobTokenStore *executor.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 0)
				require.Len(t, jobTokenStore.CreateFunc.History(), 0)
			},
		},
		{
			name: "Dequeue error",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord], jobTokenStore *executor.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{}, false, errors.New("failed to dequeue"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"dbworkerstore.Dequeue: failed to dequeue"}`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], jobTokenStore *executor.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CreateFunc.History(), 0)
			},
		},
		{
			name: "Nothing to dequeue",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord], jobTokenStore *executor.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{}, false, nil)
			},
			expectedStatusCode: http.StatusNoContent,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], jobTokenStore *executor.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CreateFunc.History(), 0)
			},
		},
		{
			name: "Failed to transform record",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
			transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executor.Job, error) {
				return executor.Job{}, errors.New("failed")
			},
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord], jobTokenStore *executor.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
				mockStore.MarkFailedFunc.PushReturn(true, nil)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"RecordTransformer: failed"}`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], jobTokenStore *executor.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, mockStore.MarkFailedFunc.History(), 1)
				assert.Equal(t, 1, mockStore.MarkFailedFunc.History()[0].Arg1)
				assert.Equal(t, "failed to transform record: failed", mockStore.MarkFailedFunc.History()[0].Arg2)
				assert.Equal(t, store.MarkFinalOptions{}, mockStore.MarkFailedFunc.History()[0].Arg3)
				require.Len(t, jobTokenStore.CreateFunc.History(), 0)
			},
		},
		{
			name: "Failed to mark record as failed",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
			transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executor.Job, error) {
				return executor.Job{}, errors.New("failed")
			},
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord], jobTokenStore *executor.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
				mockStore.MarkFailedFunc.PushReturn(false, errors.New("failed to mark"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"RecordTransformer: failed"}`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], jobTokenStore *executor.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, mockStore.MarkFailedFunc.History(), 1)
				assert.Equal(t, 1, mockStore.MarkFailedFunc.History()[0].Arg1)
				assert.Equal(t, "failed to transform record: failed", mockStore.MarkFailedFunc.History()[0].Arg2)
				assert.Equal(t, store.MarkFinalOptions{}, mockStore.MarkFailedFunc.History()[0].Arg3)
				require.Len(t, jobTokenStore.CreateFunc.History(), 0)
			},
		},
		{
			name: "V2 job",
			body: `{"executorName": "test-executor", "version": "dev", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
			transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executor.Job, error) {
				return executor.Job{ID: record.RecordID()}, nil
			},
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord], jobTokenStore *executor.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
				jobTokenStore.CreateFunc.PushReturn("sometoken", nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"version":2,"id":1,"token":"sometoken","repositoryName":"","repositoryDirectory":"","commit":"","fetchTags":false,"shallowClone":false,"sparseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redactedValues":null,"dockerAuthConfig":{}}`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], jobTokenStore *executor.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CreateFunc.History(), 1)
			},
		},
		{
			name: "Failed to create job token",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
			transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executor.Job, error) {
				return executor.Job{ID: record.RecordID()}, nil
			},
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord], jobTokenStore *executor.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
				jobTokenStore.CreateFunc.PushReturn("", errors.New("failed to create token"))
				jobTokenStore.ExistsFunc.PushReturn(false, nil)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"CreateToken: failed to create token"}`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], jobTokenStore *executor.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CreateFunc.History(), 1)
				require.Len(t, jobTokenStore.ExistsFunc.History(), 1)
				assert.Equal(t, 1, jobTokenStore.ExistsFunc.History()[0].Arg1)
				assert.Equal(t, "test", jobTokenStore.ExistsFunc.History()[0].Arg2)
			},
		},
		{
			name: "Failed to check of job token exists",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
			transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executor.Job, error) {
				return executor.Job{ID: record.RecordID()}, nil
			},
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord], jobTokenStore *executor.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
				jobTokenStore.CreateFunc.PushReturn("", errors.New("failed to create token"))
				jobTokenStore.ExistsFunc.PushReturn(false, errors.New("failed to check if token exists"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"2 errors occurred:\n\t* CreateToken: failed to create token\n\t* Exists: failed to check if token exists"}`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], jobTokenStore *executor.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CreateFunc.History(), 1)
				require.Len(t, jobTokenStore.ExistsFunc.History(), 1)
			},
		},
		{
			name: "Job token already exists",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
			transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executor.Job, error) {
				return executor.Job{ID: record.RecordID()}, nil
			},
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord], jobTokenStore *executor.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
				jobTokenStore.CreateFunc.PushReturn("", errors.New("failed to create token"))
				jobTokenStore.ExistsFunc.PushReturn(true, nil)
				jobTokenStore.RegenerateFunc.PushReturn("somenewtoken", nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":1,"token":"somenewtoken","repositoryName":"","repositoryDirectory":"","commit":"","fetchTags":false,"shallowClone":false,"sparseCheckout":null,"files":{},"dockerSteps":null,"cliSteps":null,"redactedValues":null}`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], jobTokenStore *executor.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CreateFunc.History(), 1)
				require.Len(t, jobTokenStore.ExistsFunc.History(), 1)
				require.Len(t, jobTokenStore.RegenerateFunc.History(), 1)
				assert.Equal(t, 1, jobTokenStore.RegenerateFunc.History()[0].Arg1)
				assert.Equal(t, "test", jobTokenStore.RegenerateFunc.History()[0].Arg2)
			},
		},
		{
			name: "Failed to regenerate token",
			body: `{"executorName": "test-executor", "numCPUs": 1, "memory": "1GB", "diskSpace": "10GB"}`,
			transformerFunc: func(ctx context.Context, version string, record testRecord, resourceMetadata handler.ResourceMetadata) (executor.Job, error) {
				return executor.Job{ID: record.RecordID()}, nil
			},
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord], jobTokenStore *executor.MockJobTokenStore) {
				mockStore.DequeueFunc.PushReturn(testRecord{id: 1}, true, nil)
				jobTokenStore.CreateFunc.PushReturn("", errors.New("failed to create token"))
				jobTokenStore.ExistsFunc.PushReturn(true, nil)
				jobTokenStore.RegenerateFunc.PushReturn("", errors.New("failed to regen token"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"Regenerate: failed to regen token"}`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], jobTokenStore *executor.MockJobTokenStore) {
				require.Len(t, mockStore.DequeueFunc.History(), 1)
				require.Len(t, jobTokenStore.CreateFunc.History(), 1)
				require.Len(t, jobTokenStore.ExistsFunc.History(), 1)
				require.Len(t, jobTokenStore.RegenerateFunc.History(), 1)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStore := workerstoremocks.NewMockStore[testRecord]()
			jobTokenStore := executor.NewMockJobTokenStore()

			h := handler.NewHandler(
				database.NewMockExecutorStore(),
				jobTokenStore,
				metricsstore.NewMockDistributedStore(),
				handler.QueueHandler[testRecord]{Store: mockStore, RecordTransformer: test.transformerFunc},
			)

			router := mux.NewRouter()
			router.HandleFunc("/{queueName}", h.HandleDequeue)

			req, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(test.body))
			require.NoError(t, err)

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(mockStore, jobTokenStore)
			}

			router.ServeHTTP(rw, req)

			assert.Equal(t, test.expectedStatusCode, rw.Code)

			b, err := io.ReadAll(rw.Body)
			require.NoError(t, err)

			if len(test.expectedResponseBody) > 0 {
				assert.JSONEq(t, test.expectedResponseBody, string(b))
			} else {
				assert.Empty(t, string(b))
			}

			if test.assertionFunc != nil {
				test.assertionFunc(t, mockStore, jobTokenStore)
			}
		})
	}
}

func TestHandler_HandleAddExecutionLogEntry(t *testing.T) {
	startTime := time.Date(2023, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name                 string
		body                 string
		mockFunc             func(mockStore *workerstoremocks.MockStore[testRecord])
		expectedStatusCode   int
		expectedResponseBody string
		assertionFunc        func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord])
	}{
		{
			name: "Add execution log entry",
			body: fmt.Sprintf(`{"executorName": "test-executor", "jobId": 42, "key": "foo", "command": ["faz", "baz"], "startTime": "%s", "exitCode": 0, "out": "done", "durationMs":100}`, startTime.Format(time.RFC3339)),
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord]) {
				mockStore.AddExecutionLogEntryFunc.PushReturn(10, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `10`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord]) {
				require.Len(t, mockStore.AddExecutionLogEntryFunc.History(), 1)
				assert.Equal(t, 42, mockStore.AddExecutionLogEntryFunc.History()[0].Arg1)
				assert.Equal(
					t,
					internalexecutor.ExecutionLogEntry{
						Key:        "foo",
						Command:    []string{"faz", "baz"},
						StartTime:  startTime,
						ExitCode:   newIntPtr(0),
						Out:        "done",
						DurationMs: newIntPtr(100),
					},
					mockStore.AddExecutionLogEntryFunc.History()[0].Arg2,
				)
				assert.Equal(
					t,
					store.ExecutionLogEntryOptions{WorkerHostname: "test-executor", State: "processing"},
					mockStore.AddExecutionLogEntryFunc.History()[0].Arg3,
				)
			},
		},
		{
			name: "Log entry not added",
			body: fmt.Sprintf(`{"executorName": "test-executor", "jobId": 42, "key": "foo", "command": ["faz", "baz"], "startTime": "%s", "exitCode": 0, "out": "done", "durationMs":100}`, startTime.Format(time.RFC3339)),
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord]) {
				mockStore.AddExecutionLogEntryFunc.PushReturn(0, errors.New("failed to add"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"dbworkerstore.AddExecutionLogEntry: failed to add"}`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord]) {
				require.Len(t, mockStore.AddExecutionLogEntryFunc.History(), 1)
			},
		},
		{
			name: "Unknown job",
			body: fmt.Sprintf(`{"executorName": "test-executor", "jobId": 42, "key": "foo", "command": ["faz", "baz"], "startTime": "%s", "exitCode": 0, "out": "done", "durationMs":100}`, startTime.Format(time.RFC3339)),
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord]) {
				mockStore.AddExecutionLogEntryFunc.PushReturn(0, store.ErrExecutionLogEntryNotUpdated)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"unknown job"}`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord]) {
				require.Len(t, mockStore.AddExecutionLogEntryFunc.History(), 1)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStore := workerstoremocks.NewMockStore[testRecord]()

			h := handler.NewHandler(
				database.NewMockExecutorStore(),
				executor.NewMockJobTokenStore(),
				metricsstore.NewMockDistributedStore(),
				handler.QueueHandler[testRecord]{Store: mockStore},
			)

			router := mux.NewRouter()
			router.HandleFunc("/{queueName}", h.HandleAddExecutionLogEntry)

			req, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(test.body))
			require.NoError(t, err)

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(mockStore)
			}

			router.ServeHTTP(rw, req)

			assert.Equal(t, test.expectedStatusCode, rw.Code)

			b, err := io.ReadAll(rw.Body)
			require.NoError(t, err)

			if len(test.expectedResponseBody) > 0 {
				assert.JSONEq(t, test.expectedResponseBody, string(b))
			} else {
				assert.Empty(t, string(b))
			}

			if test.assertionFunc != nil {
				test.assertionFunc(t, mockStore)
			}
		})
	}
}

func TestHandler_HandleUpdateExecutionLogEntry(t *testing.T) {
	startTime := time.Date(2023, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name                 string
		body                 string
		mockFunc             func(mockStore *workerstoremocks.MockStore[testRecord])
		expectedStatusCode   int
		expectedResponseBody string
		assertionFunc        func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord])
	}{
		{
			name: "Update execution log entry",
			body: fmt.Sprintf(`{"entryId": 10, "executorName": "test-executor", "jobId": 42, "key": "foo", "command": ["faz", "baz"], "startTime": "%s", "exitCode": 0, "out": "done", "durationMs":100}`, startTime.Format(time.RFC3339)),
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord]) {
				mockStore.UpdateExecutionLogEntryFunc.PushReturn(nil)
			},
			expectedStatusCode: http.StatusNoContent,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord]) {
				require.Len(t, mockStore.UpdateExecutionLogEntryFunc.History(), 1)
				assert.Equal(t, 42, mockStore.UpdateExecutionLogEntryFunc.History()[0].Arg1)
				assert.Equal(t, 10, mockStore.UpdateExecutionLogEntryFunc.History()[0].Arg2)
				assert.Equal(
					t,
					internalexecutor.ExecutionLogEntry{
						Key:        "foo",
						Command:    []string{"faz", "baz"},
						StartTime:  startTime,
						ExitCode:   newIntPtr(0),
						Out:        "done",
						DurationMs: newIntPtr(100),
					},
					mockStore.UpdateExecutionLogEntryFunc.History()[0].Arg3,
				)
				assert.Equal(
					t,
					store.ExecutionLogEntryOptions{WorkerHostname: "test-executor", State: "processing"},
					mockStore.UpdateExecutionLogEntryFunc.History()[0].Arg4,
				)
			},
		},
		{
			name: "Log entry not updated",
			body: fmt.Sprintf(`{"entryId": 10, "executorName": "test-executor", "jobId": 42, "key": "foo", "command": ["faz", "baz"], "startTime": "%s", "exitCode": 0, "out": "done", "durationMs":100}`, startTime.Format(time.RFC3339)),
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord]) {
				mockStore.UpdateExecutionLogEntryFunc.PushReturn(errors.New("failed to update"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"dbworkerstore.UpdateExecutionLogEntry: failed to update"}`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord]) {
				require.Len(t, mockStore.UpdateExecutionLogEntryFunc.History(), 1)
			},
		},
		{
			name: "Unknown job",
			body: fmt.Sprintf(`{"entryId": 10, "executorName": "test-executor", "jobId": 42, "key": "foo", "command": ["faz", "baz"], "startTime": "%s", "exitCode": 0, "out": "done", "durationMs":100}`, startTime.Format(time.RFC3339)),
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord]) {
				mockStore.UpdateExecutionLogEntryFunc.PushReturn(store.ErrExecutionLogEntryNotUpdated)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"unknown job"}`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord]) {
				require.Len(t, mockStore.UpdateExecutionLogEntryFunc.History(), 1)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStore := workerstoremocks.NewMockStore[testRecord]()

			h := handler.NewHandler(
				database.NewMockExecutorStore(),
				executor.NewMockJobTokenStore(),
				metricsstore.NewMockDistributedStore(),
				handler.QueueHandler[testRecord]{Store: mockStore},
			)

			router := mux.NewRouter()
			router.HandleFunc("/{queueName}", h.HandleUpdateExecutionLogEntry)

			req, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(test.body))
			require.NoError(t, err)

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(mockStore)
			}

			router.ServeHTTP(rw, req)

			assert.Equal(t, test.expectedStatusCode, rw.Code)

			b, err := io.ReadAll(rw.Body)
			require.NoError(t, err)

			if len(test.expectedResponseBody) > 0 {
				assert.JSONEq(t, test.expectedResponseBody, string(b))
			} else {
				assert.Empty(t, string(b))
			}

			if test.assertionFunc != nil {
				test.assertionFunc(t, mockStore)
			}
		})
	}
}

func TestHandler_HandleMarkComplete(t *testing.T) {
	tests := []struct {
		name                 string
		body                 string
		mockFunc             func(mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore)
		expectedStatusCode   int
		expectedResponseBody string
		assertionFunc        func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore)
	}{
		{
			name: "Mark complete",
			body: `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				mockStore.MarkCompleteFunc.PushReturn(true, nil)
				tokenStore.DeleteFunc.PushReturn(nil)
			},
			expectedStatusCode: http.StatusNoContent,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				require.Len(t, mockStore.MarkCompleteFunc.History(), 1)
				assert.Equal(t, 42, mockStore.MarkCompleteFunc.History()[0].Arg1)
				assert.Equal(t, store.MarkFinalOptions{WorkerHostname: "test-executor"}, mockStore.MarkCompleteFunc.History()[0].Arg2)
				require.Len(t, tokenStore.DeleteFunc.History(), 1)
				assert.Equal(t, 42, tokenStore.DeleteFunc.History()[0].Arg1)
				assert.Equal(t, "test", tokenStore.DeleteFunc.History()[0].Arg2)
			},
		},
		{
			name: "Failed to mark complete",
			body: `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				mockStore.MarkCompleteFunc.PushReturn(false, errors.New("failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"dbworkerstore.MarkComplete: failed"}`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				require.Len(t, mockStore.MarkCompleteFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 0)
			},
		},
		{
			name: "Unknown job",
			body: `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				mockStore.MarkCompleteFunc.PushReturn(false, nil)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `null`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				require.Len(t, mockStore.MarkCompleteFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 0)
			},
		},
		{
			name: "Failed to delete job token",
			body: `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				mockStore.MarkCompleteFunc.PushReturn(true, nil)
				tokenStore.DeleteFunc.PushReturn(errors.New("failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"jobTokenStore.Delete: failed"}`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				require.Len(t, mockStore.MarkCompleteFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 1)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStore := workerstoremocks.NewMockStore[testRecord]()
			tokenStore := executor.NewMockJobTokenStore()

			h := handler.NewHandler(
				database.NewMockExecutorStore(),
				tokenStore,
				metricsstore.NewMockDistributedStore(),
				handler.QueueHandler[testRecord]{Store: mockStore},
			)

			router := mux.NewRouter()
			router.HandleFunc("/{queueName}", h.HandleMarkComplete)

			req, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(test.body))
			require.NoError(t, err)

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(mockStore, tokenStore)
			}

			router.ServeHTTP(rw, req)

			assert.Equal(t, test.expectedStatusCode, rw.Code)

			b, err := io.ReadAll(rw.Body)
			require.NoError(t, err)

			if len(test.expectedResponseBody) > 0 {
				assert.JSONEq(t, test.expectedResponseBody, string(b))
			} else {
				assert.Empty(t, string(b))
			}

			if test.assertionFunc != nil {
				test.assertionFunc(t, mockStore, tokenStore)
			}
		})
	}
}

func TestHandler_HandleMarkErrored(t *testing.T) {
	tests := []struct {
		name                 string
		body                 string
		mockFunc             func(mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore)
		expectedStatusCode   int
		expectedResponseBody string
		assertionFunc        func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore)
	}{
		{
			name: "Mark errored",
			body: `{"executorName": "test-executor", "jobId": 42, "errorMessage": "it failed"}`,
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				mockStore.MarkErroredFunc.PushReturn(true, nil)
				tokenStore.DeleteFunc.PushReturn(nil)
			},
			expectedStatusCode: http.StatusNoContent,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				require.Len(t, mockStore.MarkErroredFunc.History(), 1)
				assert.Equal(t, 42, mockStore.MarkErroredFunc.History()[0].Arg1)
				assert.Equal(t, "it failed", mockStore.MarkErroredFunc.History()[0].Arg2)
				assert.Equal(t, store.MarkFinalOptions{WorkerHostname: "test-executor"}, mockStore.MarkErroredFunc.History()[0].Arg3)
				require.Len(t, tokenStore.DeleteFunc.History(), 1)
				assert.Equal(t, 42, tokenStore.DeleteFunc.History()[0].Arg1)
				assert.Equal(t, "test", tokenStore.DeleteFunc.History()[0].Arg2)
			},
		},
		{
			name: "Failed to mark errored",
			body: `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				mockStore.MarkErroredFunc.PushReturn(false, errors.New("failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"dbworkerstore.MarkErrored: failed"}`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				require.Len(t, mockStore.MarkErroredFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 0)
			},
		},
		{
			name: "Unknown job",
			body: `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				mockStore.MarkErroredFunc.PushReturn(false, nil)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `null`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				require.Len(t, mockStore.MarkErroredFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 0)
			},
		},
		{
			name: "Failed to delete job token",
			body: `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				mockStore.MarkErroredFunc.PushReturn(true, nil)
				tokenStore.DeleteFunc.PushReturn(errors.New("failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"jobTokenStore.Delete: failed"}`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				require.Len(t, mockStore.MarkErroredFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 1)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStore := workerstoremocks.NewMockStore[testRecord]()
			tokenStore := executor.NewMockJobTokenStore()

			h := handler.NewHandler(
				database.NewMockExecutorStore(),
				tokenStore,
				metricsstore.NewMockDistributedStore(),
				handler.QueueHandler[testRecord]{Store: mockStore},
			)

			router := mux.NewRouter()
			router.HandleFunc("/{queueName}", h.HandleMarkErrored)

			req, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(test.body))
			require.NoError(t, err)

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(mockStore, tokenStore)
			}

			router.ServeHTTP(rw, req)

			assert.Equal(t, test.expectedStatusCode, rw.Code)

			b, err := io.ReadAll(rw.Body)
			require.NoError(t, err)

			if len(test.expectedResponseBody) > 0 {
				assert.JSONEq(t, test.expectedResponseBody, string(b))
			} else {
				assert.Empty(t, string(b))
			}

			if test.assertionFunc != nil {
				test.assertionFunc(t, mockStore, tokenStore)
			}
		})
	}
}

func TestHandler_HandleMarkFailed(t *testing.T) {
	tests := []struct {
		name                 string
		body                 string
		mockFunc             func(mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore)
		expectedStatusCode   int
		expectedResponseBody string
		assertionFunc        func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore)
	}{
		{
			name: "Mark failed",
			body: `{"executorName": "test-executor", "jobId": 42, "errorMessage": "it failed"}`,
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				mockStore.MarkFailedFunc.PushReturn(true, nil)
				tokenStore.DeleteFunc.PushReturn(nil)
			},
			expectedStatusCode: http.StatusNoContent,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				require.Len(t, mockStore.MarkFailedFunc.History(), 1)
				assert.Equal(t, 42, mockStore.MarkFailedFunc.History()[0].Arg1)
				assert.Equal(t, "it failed", mockStore.MarkFailedFunc.History()[0].Arg2)
				assert.Equal(t, store.MarkFinalOptions{WorkerHostname: "test-executor"}, mockStore.MarkFailedFunc.History()[0].Arg3)
				require.Len(t, tokenStore.DeleteFunc.History(), 1)
				assert.Equal(t, 42, tokenStore.DeleteFunc.History()[0].Arg1)
				assert.Equal(t, "test", tokenStore.DeleteFunc.History()[0].Arg2)
			},
		},
		{
			name: "Failed to mark failed",
			body: `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				mockStore.MarkFailedFunc.PushReturn(false, errors.New("failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"dbworkerstore.MarkFailed: failed"}`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				require.Len(t, mockStore.MarkFailedFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 0)
			},
		},
		{
			name: "Unknown job",
			body: `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				mockStore.MarkErroredFunc.PushReturn(false, nil)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `null`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				require.Len(t, mockStore.MarkFailedFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 0)
			},
		},
		{
			name: "Failed to delete job token",
			body: `{"executorName": "test-executor", "jobId": 42}`,
			mockFunc: func(mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				mockStore.MarkFailedFunc.PushReturn(true, nil)
				tokenStore.DeleteFunc.PushReturn(errors.New("failed"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"error":"jobTokenStore.Delete: failed"}`,
			assertionFunc: func(t *testing.T, mockStore *workerstoremocks.MockStore[testRecord], tokenStore *executor.MockJobTokenStore) {
				require.Len(t, mockStore.MarkFailedFunc.History(), 1)
				require.Len(t, tokenStore.DeleteFunc.History(), 1)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStore := workerstoremocks.NewMockStore[testRecord]()
			tokenStore := executor.NewMockJobTokenStore()

			h := handler.NewHandler(
				database.NewMockExecutorStore(),
				tokenStore,
				metricsstore.NewMockDistributedStore(),
				handler.QueueHandler[testRecord]{Store: mockStore},
			)

			router := mux.NewRouter()
			router.HandleFunc("/{queueName}", h.HandleMarkFailed)

			req, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(test.body))
			require.NoError(t, err)

			rw := httptest.NewRecorder()

			if test.mockFunc != nil {
				test.mockFunc(mockStore, tokenStore)
			}

			router.ServeHTTP(rw, req)

			assert.Equal(t, test.expectedStatusCode, rw.Code)

			b, err := io.ReadAll(rw.Body)
			require.NoError(t, err)

			if len(test.expectedResponseBody) > 0 {
				assert.JSONEq(t, test.expectedResponseBody, string(b))
			} else {
				assert.Empty(t, string(b))
			}

			if test.assertionFunc != nil {
				test.assertionFunc(t, mockStore, tokenStore)
			}
		})
	}
}

type testRecord struct {
	id int
}

func (r testRecord) RecordID() int { return r.id }

func newIntPtr(i int) *int {
	return &i
}
