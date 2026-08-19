package responses

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/status"
)

type ErrorResponse struct {
	Error string `json:"error" example:"Failed to login"`
}

func ResponseGRPCError(w http.ResponseWriter, err error) {
	if st, ok := status.FromError(err); ok {
		httpStatus := runtime.HTTPStatusFromCode(st.Code())
		ResponseWithError(w, st.Message(), httpStatus)
		return
	}

	ResponseWithError(w, "Service unavailable", http.StatusBadGateway)
}

func ResponseWithError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func Response(w http.ResponseWriter, response any) error {
	if err := json.NewEncoder(w).Encode(response); err != nil {
		return errors.New("failed to decode")
	}

	return nil
}
