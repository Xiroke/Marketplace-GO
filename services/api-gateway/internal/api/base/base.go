package base

import (
	"api-gateway/internal/utils/responses"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

func HandleRequest[Req any, Res any](
	w http.ResponseWriter,
	r *http.Request,
	logger *slog.Logger,
	logic func(ctx context.Context, req *Req) (*Res, error),
) {
	defer r.Body.Close()

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	var req Req
	if r.Body != nil && r.ContentLength > 0 {
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			responses.ResponseWithError(w, "Invalid JSON data", http.StatusBadRequest)
			return
		}
	}

	data, err := logic(ctx, &req)
	if err != nil {
		responses.ResponseGRPCError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := responses.Response(w, data); err != nil {
		logger.Error(err.Error())
	}
}
