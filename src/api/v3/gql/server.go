package gql

import (
	"context"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/errcode"
	"github.com/99designs/gqlgen/graphql/executor"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type Server struct {
	exec *executor.Executor
}

func (s *Server) SetErrorPresenter(f graphql.ErrorPresenterFunc) {
	s.exec.SetErrorPresenter(f)
}

func (s *Server) SetRecoverFunc(f graphql.RecoverFunc) {
	s.exec.SetRecoverFunc(f)
}

func (s *Server) SetQueryCache(cache graphql.Cache) {
	s.exec.SetQueryCache(cache)
}

func (s *Server) Use(extension graphql.HandlerExtension) {
	s.exec.Use(extension)
}

// AroundFields is a convenience method for creating an extension that only implements field middleware
func (s *Server) AroundFields(f graphql.FieldMiddleware) {
	s.exec.AroundFields(f)
}

// AroundOperations is a convenience method for creating an extension that only implements operation middleware
func (s *Server) AroundOperations(f graphql.OperationMiddleware) {
	s.exec.AroundOperations(f)
}

// AroundResponses is a convenience method for creating an extension that only implements response middleware
func (s *Server) AroundResponses(f graphql.ResponseMiddleware) {
	s.exec.AroundResponses(f)
}

func NewDefaultServer(es graphql.ExecutableSchema) *Server {
	srv := NewServer(es)

	srv.SetQueryCache(lru.New(1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(100),
	})

	return srv
}

func NewServer(es graphql.ExecutableSchema) *Server {
	return &Server{
		exec: executor.New(es),
	}
}

func statusFor(errs gqlerror.List) int {
	switch errcode.GetErrorKind(errs) {
	case errcode.KindProtocol:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusOK
	}
}

func ProcessExecution(params *graphql.RawParams, exec graphql.GraphExecutor, baseContext context.Context) ReturnSignal {
	start := graphql.Now()
	params.ReadTime = graphql.TraceTiming{Start: start, End: graphql.Now()}

	response, listOfErrors := exec.CreateOperationContext(baseContext, params)
	if listOfErrors != nil {
		resp := exec.DispatchError(graphql.WithOperationContext(baseContext, response), listOfErrors)
		return ReturnSignal{
			Status:   statusFor(listOfErrors),
			Response: resp,
		}
	}
	responses, ctx := exec.DispatchOperation(baseContext, response)
	return ReturnSignal{
		Status:   200,
		Response: responses(ctx),
	}
}

type Response struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Reason  string `json:"reason,omitempty"`
	ReturnSignal
}

type ReturnSignal struct {
	Status   int `json:"status,omitempty"`
	Response *graphql.Response
}

func (s *Server) Process(ctx context.Context, params graphql.RawParams) (resp Response) {
	defer func() {
		if err := recover(); err != nil {
			err := s.exec.PresentRecoveredError(ctx, err)
			resp = Response{
				Message: "internal server error",
				ReturnSignal: ReturnSignal{
					Status:   500,
					Response: &graphql.Response{Errors: []*gqlerror.Error{err}},
				},
			}
			return
		}
	}()

	childContext := graphql.StartOperationTrace(ctx)
	output := ProcessExecution(&params, s.exec, childContext)
	return Response{
		ReturnSignal: output,
	}
}
